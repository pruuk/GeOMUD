package inputhandlers

import (
	"strings"

	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/term"
)

var (
	iacHandlers = []IACHandler{}
)

type IACHandler interface {
	HandleIAC(uint64, []byte) bool
}

func AddIACHandler(h IACHandler) {
	iacHandlers = append(iacHandlers, h)
}

func TelnetIACHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// Check for Telnet IAC commands
	// If not, pass it on to next handler
	if !term.IsTelnetCommand(clientInput.DataIn) {
		return true
	}

	// Multiple Telnet IAC's can be stacked into one send, so useful to split them out
	iacCmds := [][]byte{}

	var lastIAC int = 0
	for i, b := range clientInput.DataIn {
		if i != 0 && b == term.TELNET_IAC {
			if i < len(clientInput.DataIn)-1 && clientInput.DataIn[i+1] != term.TELNET_SE {
				iacCmds = append(iacCmds, clientInput.DataIn[lastIAC:i])
				lastIAC = i
			}
		}
	}

	//mudlog.Debug("Received", "type", "IAC (TEST)", "data", term.BytesString(clientInput.DataIn))

	if lastIAC < len(clientInput.DataIn) {
		iacCmds = append(iacCmds, clientInput.DataIn[lastIAC:])
	}

	for _, iacCmd := range iacCmds {
		// Check incoming Telnet IAC commands for anything useful...

		if len(iacHandlers) > 0 {

			handlerFound := false
			for _, h := range iacHandlers {
				if h.HandleIAC(clientInput.ConnectionId, iacCmd) {
					handlerFound = true
					break
				}
			}

			if handlerFound {
				continue
			}

		}

		if ok, payload := term.Matches(iacCmd, term.TelnetAcceptedChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetAcceptedChangeCharset)", "data", term.BytesString(payload))
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetRejectedChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetRejectedChangeCharset)")
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetAgreeChangeCharset); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetAgreeChangeCharset)")
			messaging.Raw(
				clientInput.ConnectionId,
				term.TelnetCharset.BytesWithPayload([]byte(" UTF-8")),
			)
			continue
		}

		if ok, _ := term.Matches(iacCmd, term.TelnetDontSuppressGoAhead); ok {
			mudlog.Debug("Received", "type", "IAC (TelnetDontSuppressGoAhead)")

			cs := connections.GetClientSettings(clientInput.ConnectionId)
			cs.SendTelnetGoAhead = true
			connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

			continue
		}

		// Is it a screen size report?
		if ok, payload := term.Matches(iacCmd, term.TelnetScreenSizeResponse); ok {

			w, h, err := term.TelnetParseScreenSizePayload(payload)
			if err != nil {
				mudlog.Debug("Received", "type", "IAC (Screensize)", "data", term.BytesString(payload), "error", err)
			} else {
				mudlog.Debug("Received", "type", "IAC (Screensize)", "width", w, "height", h)

				if err == nil {

					cs := connections.GetClientSettings(clientInput.ConnectionId)
					cs.Display.ScreenWidth = uint32(w)
					cs.Display.ScreenHeight = uint32(h)
					connections.OverwriteClientSettings(clientInput.ConnectionId, cs)
					connections.NotifyWindowChange(clientInput.ConnectionId, uint32(w), uint32(h))

				}

			}

			continue
		}

		//
		// NEW-ENVIRON / MNES client detection (see handleTelnetConnection in main.go).
		//

		// Client agreed to NEW-ENVIRON: ask it to send all of its variables.
		// (Requesting all is more broadly compatible than naming specific vars;
		// Mudlet replies with CLIENT_NAME among them.)
		if ok, _ := term.Matches(iacCmd, term.TelnetWillNewEnviron); ok {
			mudlog.Debug("Received", "type", "IAC (WILL NEW-ENVIRON)")
			messaging.Raw(clientInput.ConnectionId, term.TelnetNewEnvironSendRequest.BytesWithPayload(nil))
			continue
		}

		// Client refused NEW-ENVIRON: it won't identify itself this way, so
		// treat detection as complete (and not Mudlet).
		if ok, _ := term.Matches(iacCmd, term.TelnetWontNewEnviron); ok {
			mudlog.Debug("Received", "type", "IAC (WONT NEW-ENVIRON)")

			cs := connections.GetClientSettings(clientInput.ConnectionId)
			cs.DetectionComplete = true
			connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

			continue
		}

		// Client sent its NEW-ENVIRON variables. Scan for CLIENT_NAME=MUDLET.
		if ok, payload := term.Matches(iacCmd, term.TelnetNewEnvironResponse); ok {
			isMudlet := newEnvironIsMudlet(payload)
			mudlog.Debug("Received", "type", "IAC (NEW-ENVIRON IS)", "isMudlet", isMudlet, "data", term.BytesString(payload))

			cs := connections.GetClientSettings(clientInput.ConnectionId)
			cs.IsMudlet = isMudlet
			cs.DetectionComplete = true
			connections.OverwriteClientSettings(clientInput.ConnectionId, cs)

			continue
		}

		// Unhanlded IAC command, log it
		mudlog.Debug("Received", "type", "IAC (Unhandled)", "size", len(clientInput.DataIn), "data", term.TelnetCommandToString(iacCmd))

	}

	// We handled it, so don't pass it on
	return false
}

// newEnvironIsMudlet walks a NEW-ENVIRON IS payload looking for the MNES
// CLIENT_NAME variable. It returns true when CLIENT_NAME equals "MUDLET"
// (case-insensitive). The payload is a sequence of VAR/USERVAR name segments,
// each optionally followed by a VALUE segment; segments are delimited by the
// VAR(0)/VALUE(1)/USERVAR(3) control bytes. IAC (255) is also treated as a
// terminator so a trailing `IAC SE` left in the payload by the lenient matcher
// never bleeds into the final variable's value.
func newEnvironIsMudlet(payload []byte) bool {
	isControl := func(b byte) bool {
		return b == term.TELNET_NEWENV_VAR || b == term.TELNET_NEWENV_VALUE ||
			b == term.TELNET_NEWENV_USERVAR || b == term.TELNET_IAC
	}

	i := 0
	for i < len(payload) {
		code := payload[i]
		i++

		// Only VAR / USERVAR start a named variable.
		if code != term.TELNET_NEWENV_VAR && code != term.TELNET_NEWENV_USERVAR {
			continue
		}

		// Read the variable name up to the next control byte.
		nameStart := i
		for i < len(payload) && !isControl(payload[i]) {
			i++
		}
		name := string(payload[nameStart:i])

		// An optional VALUE segment follows.
		value := ""
		if i < len(payload) && payload[i] == term.TELNET_NEWENV_VALUE {
			i++
			valueStart := i
			for i < len(payload) && !isControl(payload[i]) {
				i++
			}
			value = string(payload[valueStart:i])
		}

		if name == "CLIENT_NAME" && strings.EqualFold(value, "MUDLET") {
			return true
		}
	}

	return false
}
