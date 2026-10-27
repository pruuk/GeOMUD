package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/events"
)

//
// Update movement points for each player
// TODO: Optimize this to avoid re-loops through users
//

func ActionPoints(e events.Event) events.ListenerReturn {

	/*
		evt, typeOk := e.(events.NewTurn)
		if !typeOk {
			mudlog.Error("Event", "Expected Type", "NewTurn", "Actual Type", e.Type())
			return events.Cancel
		}
	*/

	for _, user := range accounts.GetAllActiveUsers() {
		user.Character.ActionPoints += 1
		if user.Character.ActionPoints > user.Character.ActionPointsMax.Value {
			user.Character.ActionPoints = user.Character.ActionPointsMax.Value
		}
	}

	return events.Continue
}
