// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/storage"

	"golang.org/x/crypto/ssh/terminal"
)

func askCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(0)

	fmt.Printf("\n")
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}

func createAdmin(store *storage.Storage) {
	user := &model.User{
		Username: os.Getenv("ADMIN_USERNAME"),
		Password: os.Getenv("ADMIN_PASSWORD"),
		IsAdmin:  true,
	}

	if user.Username == "" || user.Password == "" {
		user.Username, user.Password = askCredentials()
	}

	if err := user.ValidateUserCreation(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := store.CreateUser(user); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
