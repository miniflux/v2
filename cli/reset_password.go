// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"os"

	"github.com/miniflux/miniflux/storage"
)

func resetPassword(store *storage.Storage) {
	username, password := askCredentials()
	user, err := store.UserByUsername(username)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if user == nil {
		fmt.Println("User not found!")
		os.Exit(1)
	}

	user.Password = password
	if err := user.ValidatePassword(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := store.UpdateUser(user); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Password changed!")
}
