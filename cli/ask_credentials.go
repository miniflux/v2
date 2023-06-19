// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func askCredentials() (string, string) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		fmt.Fprintf(os.Stderr, "This is not a terminal, exiting.\n")
		os.Exit(1)
	}

	fmt.Print("Enter Username: ")

	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")

	state, _ := term.GetState(fd)
	defer term.Restore(fd, state)
	bytePassword, _ := term.ReadPassword(fd)

	fmt.Printf("\n")
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}
