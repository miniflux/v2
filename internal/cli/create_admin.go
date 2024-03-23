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

func createAdminUserFromEnvironmentVariables(store *storage.Storage) {
	createAdminUser(store, config.Opts.AdminUsername(), config.Opts.AdminPassword())
}

func createAdminUserFromInteractiveTerminal(store *storage.Storage) {
	username, password := askCredentials()
	createAdminUser(store, username, password)
}

func createAdminUser(store *storage.Storage, username, password string) {
	userCreationRequest := &model.UserCreationRequest{
		Username: username,
		Password: password,
		IsAdmin:  true,
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

	if user, err := store.CreateUser(userCreationRequest); err != nil {
		printErrorAndExit(err)
	} else {
		slog.Info("Created new admin user",
			slog.String("username", user.Username),
			slog.Int64("user_id", user.ID),
		)
	}
}
