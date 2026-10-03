// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"miniflux.app/v2/internal/storage"
)

func readAPIKeyFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to read API key file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	decoded, err := hex.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return "", errors.New("API key file must contain 64 hexadecimal characters (32 random bytes)")
	}
	return token, nil
}

func provisionAPIKey(store *storage.Storage, username, description, filename string) error {
	if strings.TrimSpace(description) == "" || filename == "" {
		return errors.New("create-api-key requires api-key-description and api-key-file")
	}
	token, err := readAPIKeyFile(filename)
	if err != nil {
		return err
	}
	user, err := store.UserByUsername(username)
	if err != nil {
		return fmt.Errorf("unable to find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user %q not found", username)
	}
	return store.ProvisionAPIKey(user.ID, description, token)
}
