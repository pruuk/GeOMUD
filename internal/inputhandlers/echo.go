package inputhandlers

import (
	"github.com/GoMudEngine/GoMud/internal/connections"
	"github.com/GoMudEngine/GoMud/internal/messaging"
	"github.com/GoMudEngine/GoMud/internal/term"
)

func EchoInputHandler(clientInput *connections.ClientInput, sharedState map[string]any) (nextHandler bool) {

	// If no actual input, for now just do/change nothing
	if len(clientInput.DataIn) > 0 {
		// echo it back
		messaging.Raw(clientInput.ConnectionId, clientInput.DataIn)
	}

	// if they didn't hit enter, just keep buffering, go next.
	if !clientInput.EnterPressed {
		return false
	}

	// Echo back their Enter press
	messaging.Raw(clientInput.ConnectionId, term.CRLF)

	return true
}
