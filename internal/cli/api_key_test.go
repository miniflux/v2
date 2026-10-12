// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/database"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

func TestReadAPIKeyFile(t *testing.T) {
	token := strings.Repeat("ab", 32)
	for _, tc := range []struct {
		name, input string
		valid       bool
	}{
		{"hex", token, true}, {"newline", token + "\n", true},
		{"empty", "", false}, {"short", "abc", false},
		{"not_hex", strings.Repeat("z", 64), false},
		{"too_long", token + "ab", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "secret")
			if err := os.WriteFile(filename, []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := readAPIKeyFile(filename)
			if tc.valid {
				if err != nil || got != token {
					t.Fatalf("valid key was not read: %v", err)
				}
			} else if err == nil || got != "" {
				t.Fatal("invalid key was accepted")
			}
		})
	}
	if _, err := readAPIKeyFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestProvisionAPIKeyRequiredOptions(t *testing.T) {
	for _, options := range [][2]string{{"", "secret"}, {"   ", "secret"}, {"description", ""}} {
		if err := provisionAPIKey(nil, "admin", options[0], options[1]); err == nil {
			t.Fatal("missing options accepted")
		}
	}
}

func TestProvisionAPIKeyIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_MINIFLUX_DATABASE_URL")
	if dsn == "" {
		t.Skip("Set TEST_MINIFLUX_DATABASE_URL to run CLI integration tests")
	}
	db, err := database.NewConnectionPool(dsn, 1, 10, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := storage.NewStorage(db)
	user, err := store.CreateUser(&model.UserCreationRequest{Username: "provision-" + crypto.GenerateRandomStringHex(8), Password: "test-password"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.RemoveUser(user.ID)
	token := crypto.GenerateRandomStringHex(32)
	filename := filepath.Join(t.TempDir(), "api-key")
	if err := os.WriteFile(filename, []byte(token+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := provisionAPIKey(store, user.Username, "Homepage", filename); err != nil {
		t.Fatal(err)
	}
	keys, err := store.APIKeys(user.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("expected one key: %v", err)
	}
	original := keys[0]
	if original.Token != token {
		t.Fatal("stored token does not match secret file")
	}
	if err := store.SetAPIKeyUsedTimestamp(user.ID, token); err != nil {
		t.Fatal(err)
	}

	results := make(chan error, 8)
	for range 8 {
		go func() { results <- provisionAPIKey(store, user.Username, "homepage", filename) }()
	}
	for range 8 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	keys, err = store.APIKeys(user.ID)
	if err != nil || len(keys) != 1 {
		t.Fatalf("repeat created extra keys: %v", err)
	}
	if keys[0].ID != original.ID || !keys[0].CreatedAt.Equal(original.CreatedAt) || keys[0].LastUsedAt == nil {
		t.Fatal("repeat replaced existing key")
	}

	for _, tc := range []struct{ name, description, token string }{
		{"conflicting_description", "homepage", crypto.GenerateRandomStringHex(32)},
		{"reused_token", "another-key", token},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := store.ProvisionAPIKey(user.ID, tc.description, tc.token)
			if err == nil {
				t.Fatal("conflicting key accepted")
			}
			if strings.Contains(err.Error(), tc.token) {
				t.Fatal("error exposed token")
			}
		})
	}
	otherUser, err := store.CreateUser(&model.UserCreationRequest{Username: "other-" + user.Username, Password: "test-password"})
	if err != nil {
		t.Fatal(err)
	}
	defer store.RemoveUser(otherUser.ID)
	if err := store.ProvisionAPIKey(otherUser.ID, "Homepage", token); err == nil || strings.Contains(err.Error(), token) {
		t.Fatal("cross-user token collision was not safely rejected")
	}
	if err := provisionAPIKey(store, "missing-"+user.Username, "test", filename); err == nil {
		t.Fatal("unknown user accepted")
	}
	keys, err = store.APIKeys(user.ID)
	if err != nil || len(keys) != 1 || keys[0].Token != token {
		t.Fatal("failed provisioning changed keys")
	}
}
