# usercommands

The `usercommands` package defines a function type and contains all of the commands any actor, player or NPC, can run.

```
type Command func(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error)
```

All commands follow that definition, where
* `rest` contains everything except the initial command that was entered (for `glarble some text`, `rest` would contain `some text`)
* `actor` is the actor running the command. It has an account when a player drives it; `actor.Controller().Account()` is nil for an NPC
* `room` is the room object for the room the actor is in while executing the command

```

func Glarble(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

    room.SendText(`This glarble text goes out to all players in room`)

    room.SendText(`This glarble text goes out to all players in room except the actor`, ExcludeId(actor))

    Tell(actor, `This glarble text goes out to the actor (nowhere for an NPC)`)

    // we handled this command in this function, so return true
    return true, nil
}
```

Then the command must be entered into the `userCommands` map in [usercommands.go](/internal/usercommands/usercommands.go)

Each entry into `userCommands` defines:
* `Command` - the function for the command
* `AllowedWhenDowned` - can users execute this command when downed?
* `AdminOnly` - is this a command only admins can use?

Must return two values:
* bool - whether this was handled or should be allowed to continue down processing chain (and indicate a failed command if never gets handled)
* error - An error with error message (for logging purposes)
  
TODO: More info.

