package main

import (
	"os"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/copyover"
	"github.com/GoMudEngine/GoMud/internal/plugins"
	"github.com/GoMudEngine/GoMud/internal/rooms"
)

func triggerCopyover() error {
	binaryPath, err := os.Executable()
	if err != nil {
		return err
	}

	serverAlive.Store(false)

	rooms.SaveAllRooms()
	accounts.SaveAllUsers()
	plugins.Save()

	return copyover.Execute(binaryPath, os.Args[1:])
}
