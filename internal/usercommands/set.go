package usercommands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/actors"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func Set(rest string, actor *actors.Actor, room *rooms.Room, flags events.EventFlag) (bool, error) {
	user := actor.Controller().Account()
	if user == nil {
		return false, nil
	}

	args := util.SplitButRespectQuotes(strings.ToLower(rest))
	c := configs.GetTextFormatsConfig()

	if len(args) == 0 {

		Tell(actor,`<ansi fg="yellow-bold">description:</ansi>`)
		Tell(actor,`<ansi fg="yellow">` + util.SplitStringNL(user.Character.Description, 80) + `</ansi>`)
		Tell(actor,``)

		screenReaderTxt := `<ansi fg="red">OFF</ansi>`
		if user.ScreenReader {
			screenReaderTxt = `<ansi fg="green">ON</ansi>`
		}
		Tell(actor,`<ansi fg="yellow-bold">ScreenReader:</ansi> ` + screenReaderTxt)
		Tell(actor,``)

		on := user.GetConfigOption(`auction`)
		onTxt := `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		Tell(actor,`<ansi fg="yellow-bold">auction:</ansi> ` + onTxt)
		Tell(actor,``)

		on = user.GetConfigOption(`shortadjectives`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		Tell(actor,`<ansi fg="yellow-bold">shortadjectives:</ansi> ` + onTxt)
		Tell(actor,``)

		on = user.GetConfigOption(`tinymap`)
		onTxt = `<ansi fg="red">OFF</ansi>`
		if on == nil || on.(bool) {
			onTxt = `<ansi fg="green">ON</ansi>`
		}
		Tell(actor,`<ansi fg="yellow-bold">tinymap:</ansi> ` + onTxt)
		Tell(actor,``)

		currentPrompt := user.GetConfigOption(`prompt`)
		if currentPrompt == nil {
			currentPrompt = c.Prompt.String()
		}
		Tell(actor,`<ansi fg="yellow-bold">prompt:</ansi> ` + currentPrompt.(string))
		Tell(actor,``)

		currentPrompt = user.GetConfigOption(`fprompt`)
		if currentPrompt == nil {
			currentPrompt = c.Prompt.String()
		}
		Tell(actor,`<ansi fg="yellow-bold">fprompt:</ansi> ` + currentPrompt.(string))
		Tell(actor,``)

		currentWimpy := user.GetConfigOption(`wimpy`)
		if currentWimpy == nil {
			currentWimpy = 0
		}
		Tell(actor,fmt.Sprintf(`<ansi fg="yellow-bold">wimpy:</ansi> %d%%`, currentWimpy.(int)))
		Tell(actor,``)

		currentStyle := user.GetConfigOption(`shopstyle`)
		if currentStyle == nil {
			currentStyle = "default"
		}
		Tell(actor,`<ansi fg="yellow-bold">shopstyle:</ansi> ` + currentStyle.(string))
		Tell(actor,``)

		Tell(actor,`See: <ansi fg="command">help set</ansi>`)

		return true, nil
	}

	setTarget := args[0]
	args = args[1:]

	if setTarget == `description` {

		rest = strings.TrimSpace(rest[len(setTarget):])
		if len(rest) > 1024 {
			rest = rest[:1024]
		}
		user.Character.Description = rest

		Tell(actor,"Description set. Look at yourself to confirm.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `description`,
		})

		return true, nil
	}

	if setTarget == `auction` {
		on := user.GetConfigOption(`auction`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			Tell(actor,`Auctions toggled <ansi fg="green">ON</ansi>.`)
		} else {
			on = false
			Tell(actor,`Auctions toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`auction`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `auction`,
		})

		return true, nil

	}

	if setTarget == `shortadjectives` {
		on := user.GetConfigOption(`shortadjectives`)
		if on == nil {
			on = false
		}
		if !on.(bool) {
			on = true
			Tell(actor,`Short Adjectives toggled <ansi fg="green">ON</ansi>.`)
		} else {
			on = false
			Tell(actor,`Short Adjectives toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`shortadjectives`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `shortadjectives`,
		})

		return true, nil

	}

	if setTarget == `tinymap` {
		on := user.GetConfigOption(`tinymap`)
		if on == nil {
			on = true
		}
		if !on.(bool) {
			on = true
			Tell(actor,`Tinymap toggled <ansi fg="green">ON</ansi>.`)
		} else {
			on = false
			Tell(actor,`Tinymap toggled <ansi fg="red">OFF</ansi>.`)
		}

		user.SetConfigOption(`tinymap`, on)

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `tinymap`,
		})

		return true, nil

	}

	if setTarget == `prompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`prompt`)
			if currentPrompt == nil {
				currentPrompt = c.Prompt.String()
			}
			Tell(actor,"Your current prompt:\n")
			Tell(actor,currentPrompt.(string))
			Tell(actor,"\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`prompt`, c.Prompt.String())
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(c.Prompt.String()))
		} else if promptStr == `none` {
			user.SetConfigOption(`prompt`, ``)
			user.SetConfigOption(`prompt-compiled`, ``)
		} else {
			user.SetConfigOption(`prompt`, promptStr)
			user.SetConfigOption(`prompt-compiled`, util.ConvertColorShortTags(promptStr))
		}

		Tell(actor,"Prompt set.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `prompt`,
		})

		return true, nil

	}

	if setTarget == `fprompt` {

		if len(args) < 1 {
			currentPrompt := user.GetConfigOption(`fprompt`)
			if currentPrompt == nil {
				currentPrompt = c.Prompt.String()
			}
			Tell(actor,"Your current fprompt:\n")
			Tell(actor,currentPrompt.(string))
			Tell(actor,"\n" + `Type <ansi fg="command">help set-prompt</ansi> for more info on customizing prompts.` + "\n")
			return true, nil
		}

		promptStr := rest[len(setTarget)+1:]

		if promptStr == `default` {
			user.SetConfigOption(`fprompt`, nil)
			user.SetConfigOption(`fprompt-compiled`, nil)
		} else if promptStr == `none` {
			user.SetConfigOption(`fprompt`, ``)
			user.SetConfigOption(`fprompt-compiled`, ``)
		} else {
			user.SetConfigOption(`fprompt`, promptStr)
			user.SetConfigOption(`fprompt-compiled`, util.ConvertColorShortTags(promptStr))
		}

		Tell(actor,"fprompt set.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `fprompt`,
		})

		return true, nil

	}

	if setTarget == `wimpy` {

		if len(args) < 1 {
			currentWimpy := user.GetConfigOption(`wimpy`)
			if currentWimpy == nil {
				currentWimpy = 0
			}
			Tell(actor,"Your current wimpy:\n")
			Tell(actor,fmt.Sprintf(`%d%%`, currentWimpy.(int)))
			Tell(actor,"\n" + `Type <ansi fg="command">help wimpy</ansi> to learn about the wimpy setting.` + "\n")
			return true, nil
		}

		wimpyStr := args[0]
		wimpyInt, _ := strconv.Atoi(wimpyStr)

		if wimpyInt <= 0 {
			user.SetConfigOption(`wimpy`, nil)
		} else {
			if wimpyInt > 99 {
				wimpyInt = 99
			}
			user.SetConfigOption(`wimpy`, wimpyInt)
		}

		Tell(actor,"wimpy set.")

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `wimpy`,
		})

		return true, nil

	}

	if setTarget == `shopstyle` {

		validStyles := map[string]bool{"default": true, "fancy": true, "minimal": true, "grid": true, "arcane": true, "rugged": true}

		if len(args) < 1 {
			currentStyle := user.GetConfigOption(`shopstyle`)
			if currentStyle == nil {
				currentStyle = "default"
			}
			Tell(actor,fmt.Sprintf("Current shop style: %s", currentStyle))
			Tell(actor,`Available styles: <ansi fg="command">default</ansi>, <ansi fg="command">fancy</ansi>, <ansi fg="command">minimal</ansi>, <ansi fg="command">grid</ansi>, <ansi fg="command">arcane</ansi>, <ansi fg="command">rugged</ansi>`)
			return true, nil
		}

		style := args[0]
		if !validStyles[style] {
			Tell(actor,`Invalid style. Available: <ansi fg="command">default</ansi>, <ansi fg="command">fancy</ansi>, <ansi fg="command">minimal</ansi>, <ansi fg="command">grid</ansi>, <ansi fg="command">arcane</ansi>, <ansi fg="command">rugged</ansi>`)
			return true, nil
		}

		if style == "default" {
			user.SetConfigOption(`shopstyle`, nil)
		} else {
			user.SetConfigOption(`shopstyle`, style)
		}

		Tell(actor,fmt.Sprintf(`Shop list style set to: <ansi fg="command">%s</ansi>`, style))

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `shopstyle`,
		})

		return true, nil
	}

	if setTarget == `screenreader` {
		if user.ScreenReader {
			Tell(actor,`ScreenReader mode toggled <ansi fg="red">OFF</ansi>.`)
		} else {
			Tell(actor,`ScreenReader mode toggled <ansi fg="green">ON</ansi>.`)
		}
		user.ScreenReader = !user.ScreenReader

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `screenreader`,
		})

		return true, nil
	}

	// Are they setting a macro? // setTarget should be "=1" etc
	if len(setTarget) == 2 && setTarget[0] == '=' {

		setVal := strings.Join(args, ` `)

		macroNum, _ := strconv.Atoi(string(setTarget[1]))
		if macroNum == 0 {
			Tell(actor,"Invalid macro number supplied.")
			return true, nil
		}

		if user.Macros == nil {
			user.Macros = make(map[string]string)
		}

		// Keep macros small enough.
		if len(setVal) > 128 {
			setVal = setVal[:128]
		}

		if len(setVal) == 0 {
			delete(user.Macros, setTarget)
			Tell(actor,fmt.Sprintf(`Macro <ansi fg="command">=%d</ansi> deleted.`, macroNum))
		} else {

			allComands := strings.Split(setVal, ";")
			if len(allComands) > 10 {
				Tell(actor,`Macros are limited to 10 commands.`)
				return true, nil
			}

			finalMacroCommands := []string{}

			for _, cmd := range allComands {

				cmd = strings.TrimSpace(cmd)
				if len(cmd) > 0 {
					if cmd[0] == '=' {
						Tell(actor,`You cannot reference macros inside of a macro`)
						return true, nil
					}
					finalMacroCommands = append(finalMacroCommands, cmd)
				}

			}

			if len(finalMacroCommands) < 1 {
				Tell(actor,`There was a problem setting your macro.`)
				return true, nil
			}

			user.Macros[setTarget] = strings.Join(finalMacroCommands, `;`)

			Tell(actor,fmt.Sprintf(`Macro set. Type <ansi fg="command">=%d</ansi> or (if your terminal supports it) press <ansi fg="command">F%d</ansi> to use it.`, macroNum, macroNum))
		}

		events.AddToQueue(events.UserSettingChanged{
			UserId: user.UserId,
			Name:   `macro`,
		})

		return true, nil
	}

	return true, nil
}
