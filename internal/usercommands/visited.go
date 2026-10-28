package usercommands

import (
	"fmt"
	"math"
	"sort"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Visited(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {

	type ZoneRecord struct {
		Name       string
		Completion string
		BarFull    string
		BarEmpty   string
	}

	type VisitedInfo struct {
		ZonesTotal   int
		ZonesVisited int
		Records      []ZoneRecord
	}

	allZoneNames := rooms.GetAllZoneNames()
	sort.Strings(allZoneNames)

	records := []ZoneRecord{}

	for _, zoneName := range allZoneNames {
		zCfg := rooms.GetZoneConfig(zoneName)
		if zCfg == nil {
			continue
		}

		visited, total := actor.Character.ZoneVisitProgress(zoneName, zCfg.RoomIds)
		if visited == 0 {
			continue
		}

		pct := int(math.Floor(float64(visited) / float64(total) * 100))
		barFull, barEmpty := util.ProgressBar(float64(visited)/float64(total), 35)

		records = append(records, ZoneRecord{
			Name:       zoneName,
			Completion: fmt.Sprintf(`%d%%`, pct),
			BarFull:    barFull,
			BarEmpty:   barEmpty,
		})
	}

	info := VisitedInfo{
		ZonesTotal:   len(allZoneNames),
		ZonesVisited: len(records),
		Records:      records,
	}

	tplTxt, _ := templates.Process("character/visited", info, ExcludeId(actor))
	Tell(actor, tplTxt)

	return true, nil
}
