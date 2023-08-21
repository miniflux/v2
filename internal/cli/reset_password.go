// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"fmt"
	"os"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
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
