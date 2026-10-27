package accounts

import (
	"time"

	"github.com/GoMudEngine/GoMud/internal/prompt"
)

// Session is the connection state of an online account. It is never saved.
// It travels with the account: when a later slice lets an account drive a
// different actor (possession), the account and its session move together.
type Session struct {
	connectionId   uint64
	unsentText     string
	connectionTime time.Time
	lastInputRound uint64
	activePrompt   *prompt.Prompt
	isLinkDead     bool // are they a link-dead connection currently?
	inputBlocked   bool // Whether input is currently intentionally turned off (for a certain category of commands)
}

// sess returns the account's session, creating an empty one for an account
// decoded from disk that has not gone online yet.
func (u *Account) sess() *Session {
	if u.session == nil {
		u.session = &Session{}
	}
	return u.session
}
