package usercommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func Biome(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	biome, ok := rooms.GetBiome(room.Biome)

	if !ok {
		Tell(actor, `No biome information found about this area.`)
		return false, fmt.Errorf(`biome %s not found`, room.Biome)
	}

	Tell(actor, buildBiomePanel(biome))

	return true, nil
}
