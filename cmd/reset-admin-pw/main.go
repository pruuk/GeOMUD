package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/accounts"
	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"golang.org/x/crypto/ssh/terminal"
)

var errPasswordMismatch = errors.New("passwords did not match")

func main() {
	mudlog.SetupLogger(
		nil,
		os.Getenv(`LOG_LEVEL`),
		os.Getenv(`LOG_PATH`),
		os.Getenv(`LOG_NOCOLOR`) == ``,
	)

	if err := configs.ReloadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	adminUser, err := findAdminUser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to locate admin user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Resetting password for admin user %q.\n", adminUser.Username)

	newPassword, err := promptForPassword(os.Stdin, os.Stdout, int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read password: %v\n", err)
		os.Exit(1)
	}

	if err := adminUser.SetPassword(newPassword); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set password: %v\n", err)
		os.Exit(1)
	}

	if err := accounts.SaveUser(*adminUser); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save admin user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Admin password updated for %q.\n", adminUser.Username)
}

func findAdminUser() (*accounts.Account, error) {
	var adminUser *accounts.Account

	accounts.SearchOfflineUsers(func(u *accounts.Account) bool {
		if strings.EqualFold(u.Role, accounts.RoleAdmin) {
			copyUser := *u
			adminUser = &copyUser
			return false
		}
		return true
	})

	if adminUser == nil {
		return nil, errors.New("no offline admin user record found")
	}

	return adminUser, nil
}

func promptForPassword(stdin *os.File, stdout *os.File, fd int) (string, error) {
	var reader *bufio.Reader
	if !terminal.IsTerminal(fd) {
		reader = bufio.NewReader(stdin)
	}

	fmt.Fprint(stdout, "New admin password: ")
	first, err := readPassword(stdin, fd, reader)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(stdout)
	fmt.Fprint(stdout, "Confirm new admin password: ")
	second, err := readPassword(stdin, fd, reader)
	if err != nil {
		return "", err
	}
	fmt.Fprintln(stdout)

	if first != second {
		return "", errPasswordMismatch
	}

	return first, nil
}

func readPassword(stdin *os.File, fd int, reader *bufio.Reader) (string, error) {
	if terminal.IsTerminal(fd) {
		passwordBytes, err := terminal.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(passwordBytes), nil
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
