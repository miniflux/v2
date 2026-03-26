// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func askCredentials() (string, string) {
	fd := int(os.Stdin.Fd())

	if !term.IsTerminal(fd) {
		printErrorAndExit(errors.New("this is not an interactive terminal, exiting"))
	}

	fmt.Print("Enter Username: ")

	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	if err != nil {
		printErrorAndExit(fmt.Errorf("unable to read username: %w", err))
	}

	fmt.Print("Enter Password: ")

	state, err := term.GetState(fd)
	if err != nil {
		printErrorAndExit(fmt.Errorf("unable to get terminal state: %w", err))
	}
	defer func() {
		if restoreErr := term.Restore(fd, state); restoreErr != nil {
			printErrorAndExit(fmt.Errorf("unable to restore terminal state: %w", restoreErr))
		}
	}()

	bytePassword, err := term.ReadPassword(fd)
	if err != nil {
		printErrorAndExit(fmt.Errorf("unable to read password: %w", err))
	}

	fmt.Print("\n")
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}
