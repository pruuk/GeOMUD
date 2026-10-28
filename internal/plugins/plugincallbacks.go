package plugins

import (
	"net"

	"github.com/GoMudEngine/GoMud/internal/usercommands"
)

type PluginCallbacks struct {
	userCommands map[string]usercommands.CommandAccess

	iacHandler        func(uint64, []byte) bool
	textPrefixHandler func(uint64, []byte) bool
	onLoad            func()
	onSave            func()
	onNetConnect      func(NetConnection)
}

func newPluginCallbacks() PluginCallbacks {
	return PluginCallbacks{
		userCommands: map[string]usercommands.CommandAccess{},
	}
}

type NetConnection interface {
	IsWebSocket() bool
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close()
	RemoteAddr() net.Addr
	ConnectionId() uint64
	InputDisabled(setTo ...bool) bool
}

func (c *PluginCallbacks) SetIACHandler(f func(uint64, []byte) bool) {
	c.iacHandler = f
}

func (c *PluginCallbacks) SetTextPrefixHandler(f func(uint64, []byte) bool) {
	c.textPrefixHandler = f
}

func (c *PluginCallbacks) SetOnLoad(f func()) {
	c.onLoad = f
}

func (c *PluginCallbacks) SetOnSave(f func()) {
	c.onSave = f
}

func (c *PluginCallbacks) SetOnNetConnect(f func(NetConnection)) {
	c.onNetConnect = f
}
