// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

func askCredentials() (string, string) {
	fd := int(os.Stdin.Fd())

	if !terminal.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "This is not a terminal, exiting.")
		os.Exit(1)
	}

	fmt.Print("Enter Username: ")

	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")

	state, _ := terminal.GetState(fd)
	defer terminal.Restore(fd, state)
	bytePassword, _ := terminal.ReadPassword(fd)

	fmt.Printf("\n")
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}
