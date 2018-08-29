// +build !solaris

package password

import (
	"syscall"

	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh/terminal"
)

// This file contains all the calls needed to properly
// handle password input from stdin/terminal on all
// operating systems that aren't solaris

// IsTerminal checks whether we are running in a terminal.
func IsTerminal() bool {
	return terminal.IsTerminal(int(syscall.Stdin)) // nolint: unconvert
}

// GetPass prompts the user for a password and returns it as a string.
func GetPass() string {
	pass, _ := gopass.GetPasswd()
	return string(pass)
}
