// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"fmt"
	"os"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
	"miniflux.app/validator"
)

func createAdmin(store *storage.Storage) {
	userCreationRequest := &model.UserCreationRequest{
		Username: config.Opts.AdminUsername(),
		Password: config.Opts.AdminPassword(),
		IsAdmin:  true,
	}

	if userCreationRequest.Username == "" || userCreationRequest.Password == "" {
		userCreationRequest.Username, userCreationRequest.Password = askCredentials()
	}

	if store.UserExists(userCreationRequest.Username) {
		logger.Info(`User %q already exists, skipping creation`, userCreationRequest.Username)
		return
	}

	if validationErr := validator.ValidateUserCreationWithPassword(store, userCreationRequest); validationErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", validationErr)
		os.Exit(1)
	}

	if _, err := store.CreateUser(userCreationRequest); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
