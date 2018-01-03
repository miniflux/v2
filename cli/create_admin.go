// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"
	"os"

	"github.com/miniflux/miniflux/model"
	"github.com/miniflux/miniflux/storage"
)

func createAdmin(store *storage.Storage) {
	user := model.NewUser()
	user.Username = os.Getenv("ADMIN_USERNAME")
	user.Password = os.Getenv("ADMIN_PASSWORD")
	user.IsAdmin = true

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
