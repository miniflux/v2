// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"

	"miniflux.app/v2/internal/crypto"
)

func TestCreateAPIKeyStoresOnlyTheTokenHash(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	userID := createAPIKeyTestUser(t, db)

	apiKey, err := store.CreateAPIKey(userID, "Test")
	if err != nil {
		t.Fatal(err)
	}
	if len(apiKey.Token) != 64 {
		t.Fatalf(`Expected a 64-character token, got %q`, apiKey.Token)
	}

	var storedRow string
	var storedHash []byte
	if err := db.QueryRow(`SELECT row_to_json(api_keys)::text, token_hash FROM api_keys WHERE id=$1`, apiKey.ID).Scan(&storedRow, &storedHash); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(storedRow, apiKey.Token) {
		t.Errorf(`The api_keys row contains the plaintext token: %s`, storedRow)
	}
	if !bytes.Equal(storedHash, crypto.HashToken(apiKey.Token)) {
		t.Errorf(`Expected the SHA-256 digest of the token to be stored`)
	}

	user, err := store.UserByAPIKey(apiKey.Token)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil || user.ID != userID {
		t.Fatalf(`Expected the token to authenticate user %d, got %+v`, userID, user)
	}

	// The stored digest must not be usable as an API key.
	if user, err := store.UserByAPIKey(hex.EncodeToString(storedHash)); err != nil || user != nil {
		t.Errorf(`Expected no user for the stored digest, got %+v (error: %v)`, user, err)
	}

	apiKeys, err := store.APIKeys(userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(apiKeys) != 1 || apiKeys[0].Token != "" {
		t.Errorf(`Expected one API key without its token, got %+v`, apiKeys)
	}
}

func TestUserByAPIKeyFindsKeysHashedByTheMigration(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	userID := createAPIKeyTestUser(t, db)

	// The migration hashes existing tokens in SQL; the digest must match the one computed in Go.
	legacyToken := crypto.GenerateRandomStringHex(32)
	if _, err := db.Exec(
		`INSERT INTO api_keys (user_id, token_hash, description) VALUES ($1, sha256(convert_to($2, 'UTF8')), 'Legacy')`,
		userID, legacyToken,
	); err != nil {
		t.Fatal(err)
	}

	user, err := store.UserByAPIKey(legacyToken)
	if err != nil {
		t.Fatal(err)
	}
	if user == nil || user.ID != userID {
		t.Fatalf(`Expected the migrated token to authenticate user %d, got %+v`, userID, user)
	}

	if err := store.SetAPIKeyUsedTimestamp(userID, legacyToken); err != nil {
		t.Fatal(err)
	}
	var lastUsedAtIsSet bool
	if err := db.QueryRow(`SELECT last_used_at IS NOT NULL FROM api_keys WHERE user_id=$1`, userID).Scan(&lastUsedAtIsSet); err != nil {
		t.Fatal(err)
	}
	if !lastUsedAtIsSet {
		t.Errorf(`Expected SetAPIKeyUsedTimestamp to update last_used_at`)
	}
}

func createAPIKeyTestUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	var userID int64
	username := "api_key_test_user_" + crypto.GenerateRandomStringHex(10)
	if err := db.QueryRow(`INSERT INTO users (username) VALUES ($1) RETURNING id`, username).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM users WHERE id=$1`, userID); err != nil {
			t.Errorf(`Unable to remove user %d: %v`, userID, err)
		}
	})
	return userID
}
