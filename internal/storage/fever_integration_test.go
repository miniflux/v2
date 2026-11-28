// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"miniflux.app/v2/internal/crypto"
)

func TestUserByFeverTokenFindsTokensHashedByTheMigration(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	userID := createTestUser(t, db)

	// The migration hashes existing tokens in SQL after lowercasing them; the digest must match the one computed in Go.
	feverToken := crypto.GenerateRandomStringHex(16)
	if _, err := db.Exec(
		`INSERT INTO integrations (user_id, fever_enabled, fever_username, fever_token_hash) VALUES ($1, 't', $2, sha256(convert_to(lower($3), 'UTF8')))`,
		userID, "fever_"+feverToken, strings.ToUpper(feverToken),
	); err != nil {
		t.Fatal(err)
	}

	for _, candidate := range []string{feverToken, strings.ToUpper(feverToken)} {
		user, err := store.UserByFeverToken(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if user == nil || user.ID != userID {
			t.Fatalf(`Expected the Fever token %q to authenticate user %d, got %+v`, candidate, userID, user)
		}
	}

	// The stored digest must not be usable as a Fever API key.
	integration, err := store.Integration(userID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(integration.FeverTokenHash, crypto.HashToken(feverToken)) {
		t.Fatalf(`Expected the SHA-256 digest of the lowercase token to be stored`)
	}
	if user, err := store.UserByFeverToken(hex.EncodeToString(integration.FeverTokenHash)); err != nil || user != nil {
		t.Errorf(`Expected no user for the stored digest, got %+v (error: %v)`, user, err)
	}

	// Saving the integration keeps the digest.
	if err := store.UpdateIntegration(integration); err != nil {
		t.Fatal(err)
	}
	if user, err := store.UserByFeverToken(feverToken); err != nil || user == nil {
		t.Fatalf(`Expected the Fever token to still authenticate the user after saving the integration, got %+v (error: %v)`, user, err)
	}

	// Disabling Fever clears the digest, as the integrations page does.
	integration.FeverEnabled = false
	integration.FeverTokenHash = nil
	if err := store.UpdateIntegration(integration); err != nil {
		t.Fatal(err)
	}
	if user, err := store.UserByFeverToken(feverToken); err != nil || user != nil {
		t.Errorf(`Expected no user once Fever is disabled, got %+v (error: %v)`, user, err)
	}
}
