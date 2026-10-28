package characters

import "github.com/GoMudEngine/GoMud/internal/ids"

const (
	CharmPermanent     = -1 // Never expires due to time
	CharmExpiredRest   = `emote bows and bids you farewell, disappearing into the scenery;rest charmed mob expired`
	CharmExpiredRevert = `emote reverts to its old ways`
)

var ()

type CharmInfo struct {
	By              ids.ID // The charmer's body
	RoundsRemaining int    // If -1, never expires
	ExpiredCommand  string // Any valid mob commands such as `emote bows and waves farewell;say Farewell`
}

func NewCharm(by ids.ID, rounds int, expireCommand string) *CharmInfo {
	return &CharmInfo{By: by, RoundsRemaining: rounds, ExpiredCommand: expireCommand}
}

func (ch *CharmInfo) Expire() {
	ch.RoundsRemaining = 0
}
