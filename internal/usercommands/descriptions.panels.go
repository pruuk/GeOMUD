package usercommands

import (
	"fmt"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/templates"
	"github.com/GoMudEngine/GoMud/internal/term"
	"github.com/GoMudEngine/ansitags"
)

const (
	alertBorderDashes = 69
	alertContentWidth = alertBorderDashes - 2 // one space padding on each side inside the border
)

var (
	alertBorderTop     = `    <ansi fg="red">┌` + strings.Repeat(`─`, alertBorderDashes) + `┐</ansi>`
	alertBorderBottom  = `    <ansi fg="red">└` + strings.Repeat(`─`, alertBorderDashes) + `┘</ansi>`
	alertContentPrefix = `    <ansi fg="red">│</ansi> `
)

// wrapAlertLine splits a single alert line into visual chunks that each fit
// within alertContentWidth visible characters, respecting ANSI tags.
// Blank input produces a single empty string so blank lines are preserved.
func wrapAlertLine(line string) []string {
	chunks := ansitags.SplitStringOnSpaces(line, alertContentWidth, true)
	var result []string
	for _, chunk := range chunks {
		if len(strings.TrimSpace(ansitags.Parse(chunk, ansitags.StripTags))) > 0 {
			result = append(result, chunk)
		}
	}
	if len(result) == 0 {
		return []string{``}
	}
	return result
}

func buildRoomDescPanel(details rooms.RoomTemplateDetails) string {
	var out strings.Builder

	descColor := `room-description`
	if details.IsNight || details.IsDark {
		descColor = `room-description-dark`
	}
	out.WriteString(fmt.Sprintf(`<ansi fg="%s">%s</ansi>`, descColor, details.Description))

	for _, group := range details.RoomAlerts {
		out.WriteString(term.CRLFStr)
		out.WriteString(term.CRLFStr)
		out.WriteString(alertBorderTop)
		for _, line := range group {
			for _, wline := range wrapAlertLine(line) {
				out.WriteString(term.CRLFStr)
				out.WriteString(alertContentPrefix + wline)
			}
		}
		out.WriteString(term.CRLFStr)
		out.WriteString(alertBorderBottom)
	}

	if details.TrackingString != `` {
		out.WriteString(term.CRLFStr)
		out.WriteString(fmt.Sprintf(`<ansi fg="182">%s</ansi>`, details.TrackingString))
		out.WriteString(term.CRLFStr)
	}

	return out.String()
}

func buildInsideContainerPanel(itemNames []string, itemNamesFormatted []string) string {
	layout, err := templates.LoadPanelLayout("room/container-inside")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 0)
		layout.AddPanelsToSlot(layout.AddSlot(), "inside")
		layout.Panel("inside").SetWidth(78)
	}

	p := layout.Panel("inside")
	if len(itemNames) == 0 {
		p.Add(`<ansi fg="white">Inside:</ansi>`, `<ansi fg="white">Inside:</ansi>`, `Nothing`)
	} else {
		// Wrap item names into lines of ~66 chars
		var line strings.Builder
		lineLen := 0
		first := true
		label := `<ansi fg="white">Inside:</ansi>`
		for i, name := range itemNames {
			proposed := lineLen + len(name) + 2
			if !first && proposed > 66 {
				p.Add(label, label, line.String())
				label = ``
				line.Reset()
				lineLen = 0
			}
			if line.Len() > 0 {
				line.WriteString(`, `)
				lineLen += 2
			}
			line.WriteString(itemNamesFormatted[i])
			lineLen += len(name)
			first = false
		}
		if line.Len() > 0 {
			p.Add(label, label, line.String())
		}
	}

	return layout.Render() + term.CRLFStr
}

func buildBiomePanel(biome *rooms.BiomeInfo) string {
	layout, err := templates.LoadPanelLayout("room/biome")
	if err != nil {
		layout = templates.NewPanelLayout("open", "single", 1, 0)
		layout.AddPanelsToSlot(layout.AddSlot(), "biome")
		layout.Panel("biome").SetTitle(` <ansi fg="black-bold">.:</ansi> Biome Info `).SetWidth(78)
	}
	layout.Panel("biome").SetLabelWidth(13)

	var lighting string
	if biome.IsDark() {
		lighting = `It's always dark.`
	} else if biome.IsLit() {
		lighting = `It is kept well lit at night.`
	} else {
		lighting = `Visibility is affected by the day/night cycle.`
	}

	p := layout.Panel("biome")
	p.Add(`<ansi fg="yellow">Name:</ansi>`, `<ansi fg="yellow">Name:</ansi>`, biome.Name)
	p.Add(`<ansi fg="yellow">Symbol:</ansi>`, `<ansi fg="yellow">Sym:</ansi>`, biome.SymbolString())
	p.Add(`<ansi fg="yellow">Lighting:</ansi>`, `<ansi fg="yellow">Light:</ansi>`, lighting)
	p.Add(`<ansi fg="yellow">Description:</ansi>`, `<ansi fg="yellow">Description:</ansi>`, biome.Description)

	return layout.Render() + term.CRLFStr
}
