package gmcp

import (
	"embed"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/usercommands"
	"github.com/GoMudEngine/GoMud/internal/users"
)

var (
	//go:embed files/*
	files embed.FS
)

// MudletConfig holds the configuration for Mudlet clients
type MudletConfig struct {
	// Mapper configuration
	MapperVersion string `json:"mapper_version" yaml:"mapper_version"`
	MapperURL     string `json:"mapper_url" yaml:"mapper_url"`

	// UI configuration
	UIVersion string `json:"ui_version" yaml:"ui_version"`
	UIURL     string `json:"ui_url" yaml:"ui_url"`

	// Map data configuration
	MapVersion string `json:"map_version" yaml:"map_version"`
	MapURL     string `json:"map_url" yaml:"map_url"`
}

// GMCPMudletModule handles Mudlet-specific GMCP functionality
type GMCPMudletModule struct {
	plug   *plugins.Plugin
	config MudletConfig
}

// GMCPMudletDetected is an event fired when a Mudlet client is detected
type GMCPMudletDetected struct {
	ConnectionId uint64
	UserId       int
}

func (g GMCPMudletDetected) Type() string { return `GMCPMudletDetected` }

func init() {
	// Create module with basic structure
	g := GMCPMudletModule{
		plug: plugins.New(`gmcp.Mudlet`, `1.0`),
	}

	// Attach filesystem with proper error handling
	if err := g.plug.AttachFileSystem(files); err != nil {
		panic(err)
	}

	g.plug.Web.AdminPage("Config", "gmcp-config", "html/admin/gmcp-config.html", true, "Modules", "GMCP", "Configure GMCP protocol options and enabled namespaces.", "GMCP protocol module sending structured data to supporting MUD clients.", nil)
	g.plug.Web.AdminPage("About", "gmcp-about", "html/admin/gmcp-about.html", true, "Modules", "GMCP", "Information and version details for the GMCP module.", "", nil)

	// Register callbacks for load/save
	g.plug.Callbacks.SetOnLoad(g.load)
	g.plug.Callbacks.SetOnSave(g.save)

	// Register event listeners
	events.RegisterListener(events.PlayerSpawn{}, g.playerSpawnHandler)
	events.RegisterListener(GMCPMudletDetected{}, g.mudletDetectedHandler)

	// Register the Mudlet-specific user commands
	g.plug.AddUserCommand("mudletmap", g.sendMapCommand, true, false)
	g.plug.AddUserCommand("mudletui", g.sendUICommand, false, false)
	g.plug.AddUserCommand("checkclient", g.checkClientCommand, true, false)
}

// Helper function to load a config string from the plugin's configuration
func loadConfigString(p *plugins.Plugin, key string) string {
	if val, ok := p.Config.Get(key).(string); ok {
		return val
	}
	return ""
}

// load handles loading configuration from the plugin's storage
func (g *GMCPMudletModule) load() {
	// Load config values directly from embedded config or overrides
	g.config.MapperVersion = loadConfigString(g.plug, "mapper_version")
	g.config.MapperURL = loadConfigString(g.plug, "mapper_url")
	g.config.UIVersion = loadConfigString(g.plug, "ui_version")
	g.config.UIURL = loadConfigString(g.plug, "ui_url")
	g.config.MapVersion = loadConfigString(g.plug, "map_version")
	g.config.MapURL = loadConfigString(g.plug, "map_url")
}

// save handles saving configuration to the plugin's storage
func (g *GMCPMudletModule) save() {
	g.plug.WriteStruct(`mudlet_config`, g.config)
}

// Helper function to get user config option with default boolean value
func getUserBoolOption(user *users.UserRecord, key string, defaultValue bool) bool {
	val := user.GetConfigOption(key)
	if val == nil {
		return defaultValue
	}
	if boolVal, ok := val.(bool); ok {
		return boolVal
	}
	return defaultValue
}

// Helper to send GMCP event
func sendGMCP(userId int, module string, payload interface{}) {
	if userId < 1 {
		return
	}
	events.AddToQueue(GMCPOut{
		UserId:  userId,
		Module:  module,
		Payload: payload,
	})
}

// Send Mudlet map configuration
func (g *GMCPMudletModule) sendMudletMapConfig(userId int) {
	if userId < 1 {
		return
	}

	mapConfig := map[string]string{
		"url": g.config.MapURL,
	}

	sendGMCP(userId, "Client.Map", mapConfig)
	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet map config", "userId", userId)
}

// Send Mudlet UI package installation message
func (g *GMCPMudletModule) sendMudletUIInstall(userId int) {
	if userId < 1 {
		return
	}

	payload := struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}{
		Version: g.config.UIVersion,
		URL:     g.config.UIURL,
	}

	sendGMCP(userId, "Client.GUI", payload)
	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI install config", "userId", userId)
}

// Send Mudlet UI package removal message
func (g *GMCPMudletModule) sendMudletUIRemove(userId int) {
	if userId < 1 {
		return
	}

	payload := struct {
		GoMudUI string `json:"gomudui"`
	}{
		GoMudUI: "remove",
	}

	sendGMCP(userId, "Client.GUI", payload)
	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI remove command", "userId", userId)
}

// Send Mudlet UI package update message
func (g *GMCPMudletModule) sendMudletUIUpdate(userId int) {
	if userId < 1 {
		return
	}

	payload := struct {
		GoMudUI string `json:"gomudui"`
	}{
		GoMudUI: "update",
	}

	sendGMCP(userId, "Client.GUI", payload)
	mudlog.Debug("GMCP", "type", "Mudlet", "action", "Sent Mudlet UI update command", "userId", userId)
}

// Send mapper configuration to Mudlet client
func (g *GMCPMudletModule) sendMudletConfig(userId int) {
	if userId < 1 {
		return
	}

	// Send mapper info
	payload := struct {
		Version string `json:"version"`
		URL     string `json:"url"`
	}{
		Version: g.config.MapperVersion,
		URL:     g.config.MapperURL,
	}
	sendGMCP(userId, "Client.GUI", payload)

	mudlog.Info("GMCP", "type", "Mudlet", "action", "Sent Mudlet package config", "userId", userId)
}

// playerSpawnHandler handles when a player connects
func (g *GMCPMudletModule) playerSpawnHandler(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(events.PlayerSpawn)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "PlayerSpawn", "Actual Type", e.Type())
		return events.Cancel
	}

	// Check if the client is Mudlet
	if gmcpData, ok := gmcpModule.cache.Get(evt.ConnectionId); ok && gmcpData.Client.IsMudlet {
		// Send Mudlet-specific GMCP
		g.sendMudletConfig(evt.UserId)
	}

	return events.Continue
}

// mudletDetectedHandler handles when a Mudlet client is detected
func (g *GMCPMudletModule) mudletDetectedHandler(e events.Event) events.ListenerReturn {
	evt, typeOk := e.(GMCPMudletDetected)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "GMCPMudletDetected", "Actual Type", e.Type())
		return events.Cancel
	}

	if evt.UserId > 0 {
		g.sendMudletConfig(evt.UserId)
	}

	return events.Continue
}

// Helper function for handling command toggles
func (g *GMCPMudletModule) handleToggleCommand(user *users.UserRecord, settingName string, value bool, enableMsg string, disableMsg string) {
	user.SetConfigOption(settingName, value)
	if value {
		user.SendText("\n<ansi fg=\"green\">" + enableMsg + "</ansi>\n")
	} else {
		user.SendText("\n<ansi fg=\"yellow\">" + disableMsg + "</ansi>\n")
	}
}

// sendUICommand handles UI-related commands
func (g *GMCPMudletModule) sendUICommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Only proceed if client is Mudlet
	connId := user.ConnectionId()
	if gmcpData, ok := gmcpModule.cache.Get(connId); !ok || !gmcpData.Client.IsMudlet {
		user.SendText("\n<ansi fg=\"red\">This command is only available for Mudlet clients.</ansi> You are currently using: " + gmcpData.Client.Name + "\n")
		return true, nil
	}

	// Process arguments
	args := strings.Fields(rest)
	if len(args) == 0 {
		// No arguments - show status and help
		mudName := configs.GetServerConfig().MudName.String()

		// Check prompt display status
		var promptStatus string
		if getUserBoolOption(user, "mudlet_ui_prompt_disabled", false) {
			promptStatus = "<ansi fg=\"red\">HIDDEN</ansi>"
		} else {
			promptStatus = "<ansi fg=\"green\">ENABLED</ansi>"
		}

		user.SendText("\n<ansi fg=\"cyan-bold\">" + mudName + " Mudlet UI Management</ansi>\n")
		user.SendText("<ansi fg=\"yellow-bold\">Status:</ansi>\n")
		user.SendText("  Login message display: " + promptStatus + "\n")
		user.SendText("<ansi fg=\"yellow-bold\">Available Commands:</ansi>\n")
		user.SendText("  <ansi fg=\"command\">mudletui install</ansi> - Install the Mudlet UI package\n")
		user.SendText("  <ansi fg=\"command\">mudletui remove</ansi>  - Remove the Mudlet UI package\n")
		user.SendText("  <ansi fg=\"command\">mudletui update</ansi>  - Manually check for updates to the Mudlet UI package\n")
		user.SendText("  <ansi fg=\"command\">mudletui hide</ansi>    - Hide login messages\n")
		user.SendText("  <ansi fg=\"command\">mudletui show</ansi>    - Enable login messages\n\n")
		user.SendText("For more information, type <ansi fg=\"command\">help mudletui</ansi>\n")
		return true, nil
	}

	// Handle specific commands
	switch args[0] {
	case "install":
		g.sendMudletUIInstall(user.UserId)
		user.SetConfigOption("mudlet_ui_prompt_disabled", true)
		user.SendText("\n<ansi fg=\"green\">UI installation package sent to your Mudlet client.</ansi> If it doesn't install automatically, you may need to accept the installation prompt in Mudlet.\n")

	case "remove":
		g.sendMudletUIRemove(user.UserId)
		user.SendText("\n<ansi fg=\"yellow\">UI removal command sent to your Mudlet client.</ansi>\n")

	case "update":
		g.sendMudletUIUpdate(user.UserId)
		user.SendText("\n<ansi fg=\"cyan\">Manual UI update check sent to your Mudlet client.</ansi>\n")

	case "hide":
		g.handleToggleCommand(user, "mudlet_ui_prompt_disabled", true,
			"The Mudlet UI prompt has been hidden.",
			"")
		user.SendText("You can use <ansi fg=\"command\">mudletui show</ansi> in the future if you want to see the prompts again.\n")

	case "show":
		g.handleToggleCommand(user, "mudlet_ui_prompt_disabled", false,
			"The Mudlet UI prompt has been re-enabled.",
			"")
		user.SendText("You can use <ansi fg=\"command\">mudletui hide</ansi> in the future if you want to hide the prompts again.\n")

	default:
		user.SendText("\nUsage: mudletui install|remove|update|hide|show\n\nType '<ansi fg=\"command\">help mudletui</ansi>' for more information.\n")
	}

	return true, nil
}

// sendMapCommand sends map configuration
func (g *GMCPMudletModule) sendMapCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Only send if the client is Mudlet
	connId := user.ConnectionId()
	if gmcpData, ok := gmcpModule.cache.Get(connId); ok && gmcpData.Client.IsMudlet {
		g.sendMudletMapConfig(user.UserId)
		return true, nil
	}
	return false, nil
}

// checkClientCommand checks if client is Mudlet and shows info
func (g *GMCPMudletModule) checkClientCommand(rest string, user *users.UserRecord, room *rooms.Room, flags events.EventFlag) (bool, error) {
	// Check if client is Mudlet
	connId := user.ConnectionId()
	if gmcpData, ok := gmcpModule.cache.Get(connId); ok && gmcpData.Client.IsMudlet {
		// Skip if prompt is disabled
		if getUserBoolOption(user, "mudlet_ui_prompt_disabled", false) {
			return true, nil
		}

		// Show Mudlet help
		user.SendText("\n\n<ansi fg=\"cyan-bold\">We have detected you are using Mudlet as a client.</ansi>\n")
		usercommands.Help("mudletui", user, room, flags)
	}
	return true, nil
}
