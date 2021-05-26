// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"fmt"
	"os"

	"miniflux.app/model"
	"miniflux.app/storage"
	"miniflux.app/validator"
)

func resetPassword(store *storage.Storage) {
	username, password := askCredentials()
	user, err := store.UserByUsername(username)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if user == nil {
		fmt.Fprintf(os.Stderr, "User not found!\n")
		os.Exit(1)
	}

	userModificationRequest := &model.UserModificationRequest{
		Password: &password,
	}
	if validationErr := validator.ValidateUserModification(store, user.ID, userModificationRequest); validationErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", validationErr)
		os.Exit(1)
	}

	user.Password = password
	if err := store.UpdateUser(user); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Password changed!")
}
