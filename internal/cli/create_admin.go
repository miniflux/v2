// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/v2/internal/cli"

import (
	"log/slog"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
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
		slog.Info("Skipping admin user creation because it already exists",
			slog.String("username", userCreationRequest.Username),
		)
		return
	}

	if validationErr := validator.ValidateUserCreationWithPassword(store, userCreationRequest); validationErr != nil {
		printErrorAndExit(validationErr.Error())
	}

	if _, err := store.CreateUser(userCreationRequest); err != nil {
		printErrorAndExit(err)
	}
}
