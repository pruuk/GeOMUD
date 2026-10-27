package web

import (
	"sync"

	"github.com/GoMudEngine/GoMud/internal/accounts"
)

type Stats struct {
	OnlineUsers          []accounts.OnlineInfo
	TelnetPorts          []int
	WebSocketPort        int
	SSHPort              int
	TelnetConnections    int
	WebSocketConnections int
	SSHConnections       int
}

var (
	statsLock   = sync.RWMutex{}
	serverStats = Stats{
		WebSocketPort: 0,
		SSHPort:       0,
		OnlineUsers:   []accounts.OnlineInfo{},
		TelnetPorts:   []int{},
	}
)

// Returns a copy of the server stats
func GetStats() Stats {
	statsLock.RLock()
	defer statsLock.RUnlock()

	return serverStats
}

// Returns a copy of the server stats
func UpdateStats(s Stats) {
	statsLock.RLock()
	defer statsLock.RUnlock()

	serverStats = s
}

func (s *Stats) Reset() {
	s.WebSocketPort = 0
	s.SSHPort = 0
	s.OnlineUsers = []accounts.OnlineInfo{}
	s.TelnetPorts = []int{}
	s.TelnetConnections = 0
	s.WebSocketConnections = 0
	s.SSHConnections = 0
}
