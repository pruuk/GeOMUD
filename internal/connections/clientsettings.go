package connections

const (
	// DefaultScreenWidth is the default width of the screen
	DefaultScreenWidth = 80
	// DefaultScreenHeight is the default height of the screen
	DefaultScreenHeight = 24
)

type ClientSettings struct {
	Display           DisplaySettings
	SendTelnetGoAhead bool // Defaults false, should we send a IAC GA after prompts?
	// IsMudlet is true when the client identified itself as Mudlet (via the MNES
	// NEW-ENVIRON CLIENT_NAME variable). Mudlet echoes input locally and treats
	// the telnet ECHO option purely as a password-masking hint, so it must not
	// receive server-side echo.
	IsMudlet bool
	// DetectionComplete is set once the connect-time client-type probe has
	// resolved (either a NEW-ENVIRON reply arrived or the probe timed out).
	DetectionComplete bool
}

type DisplaySettings struct {
	ScreenWidth  uint32
	ScreenHeight uint32
}

func (c DisplaySettings) GetScreenWidth() int {
	if c.ScreenWidth == 0 {
		return DefaultScreenWidth
	}

	return int(c.ScreenWidth)
}

func (c DisplaySettings) GetScreenHeight() int {
	if c.ScreenHeight == 0 {
		return DefaultScreenHeight
	}
	return int(c.ScreenHeight)
}
