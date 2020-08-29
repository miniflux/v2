// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"fmt"
	"os"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
)

func createAdmin(store *storage.Storage) {
	user := model.NewUser()
	user.Username = config.Opts.AdminUsername()
	user.Password = config.Opts.AdminPassword()
	user.IsAdmin = true

	if user.Username == "" || user.Password == "" {
		user.Username, user.Password = askCredentials()
	}

	if err := user.ValidateUserCreation(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if store.UserExists(user.Username) {
		logger.Info(`User %q already exists, skipping creation`, user.Username)
		return
	}

	if err := store.CreateUser(user); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
