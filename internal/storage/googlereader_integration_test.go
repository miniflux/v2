// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"database/sql"
	"encoding/hex"
	"strings"
	"testing"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
)

const testGoogleReaderPassword = "googlereader password"

func TestGoogleReaderUserLoginStoresOnlyTheTokenSaltAndHash(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	userID, username := createGoogleReaderTestUser(t, db, store, func(integration *model.Integration) {
		integration.SetGoogleReaderPassword(testGoogleReaderPassword)
	})

	if _, _, err := store.GoogleReaderUserLogin(username, "wrong password"); err == nil {
		t.Fatal(`Expected an error for a wrong password`)
	}

	loginUserID, secret, err := store.GoogleReaderUserLogin(username, testGoogleReaderPassword)
	if err != nil {
		t.Fatal(err)
	}
	if loginUserID != userID {
		t.Fatalf(`Expected user %d, got %d`, userID, loginUserID)
	}

	var storedRow string
	if err := db.QueryRow(`SELECT row_to_json(integrations)::text FROM integrations WHERE user_id=$1`, userID).Scan(&storedRow); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(storedRow, secret) {
		t.Errorf(`The integrations row contains the token secret: %s`, storedRow)
	}

	if tokenUserID, err := store.GoogleReaderUserIDByToken(username, secret); err != nil || tokenUserID != userID {
		t.Fatalf(`Expected the token to authenticate user %d, got %d (error: %v)`, userID, tokenUserID, err)
	}

	// The stored digest must not be usable as a token secret.
	integration, err := store.Integration(userID)
	if err != nil {
		t.Fatal(err)
	}
	if tokenUserID, err := store.GoogleReaderUserIDByToken(username, hex.EncodeToString(integration.GoogleReaderTokenHash)); err != nil || tokenUserID != 0 {
		t.Errorf(`Expected no user for the stored digest, got %d (error: %v)`, tokenUserID, err)
	}
}

func TestGoogleReaderUserLoginAfterAConcurrentUpgrade(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	legacyHash, err := crypto.HashPassword(testGoogleReaderPassword)
	if err != nil {
		t.Fatal(err)
	}
	userID, username := createGoogleReaderTestUser(t, db, store, func(integration *model.Integration) {
		integration.GoogleReaderPassword = legacyHash
	})

	// Two clients log in at the same time: both read the legacy hash before either upgrades it.
	staleIntegration, err := store.Integration(userID)
	if err != nil {
		t.Fatal(err)
	}
	_, firstSecret, err := store.GoogleReaderUserLogin(username, testGoogleReaderPassword)
	if err != nil {
		t.Fatal(err)
	}
	_, secondSecret, err := store.upgradeGoogleReaderPassword(staleIntegration, username, testGoogleReaderPassword)
	if err != nil {
		t.Fatal(err)
	}

	if secondSecret != firstSecret {
		t.Errorf(`Expected both logins to return the same token secret, got %q and %q`, firstSecret, secondSecret)
	}
	if tokenUserID, err := store.GoogleReaderUserIDByToken(username, firstSecret); err != nil || tokenUserID != userID {
		t.Errorf(`Expected the first token to remain valid, got %d (error: %v)`, tokenUserID, err)
	}
}

func TestGoogleReaderUserIDByTokenChecksTheUsername(t *testing.T) {
	db := openTestDatabase(t)
	store := NewStorage(db)
	userID, username := createGoogleReaderTestUser(t, db, store, func(integration *model.Integration) {
		integration.SetGoogleReaderPassword(testGoogleReaderPassword)
	})

	_, secret, err := store.GoogleReaderUserLogin(username, testGoogleReaderPassword)
	if err != nil {
		t.Fatal(err)
	}

	// The token secret does not depend on the username, so changing only the username keeps the password working.
	newUsername := username + "_renamed"
	if _, err := db.Exec(`UPDATE integrations SET googlereader_username=$1 WHERE user_id=$2`, newUsername, userID); err != nil {
		t.Fatal(err)
	}

	if tokenUserID, err := store.GoogleReaderUserIDByToken(username, secret); err != nil || tokenUserID != 0 {
		t.Errorf(`Expected no user for the previous username, got %d (error: %v)`, tokenUserID, err)
	}
	if tokenUserID, err := store.GoogleReaderUserIDByToken(newUsername, secret); err != nil || tokenUserID != userID {
		t.Errorf(`Expected the token with the new username to authenticate user %d, got %d (error: %v)`, userID, tokenUserID, err)
	}
	if _, renamedSecret, err := store.GoogleReaderUserLogin(newUsername, testGoogleReaderPassword); err != nil || renamedSecret != secret {
		t.Errorf(`Expected ClientLogin with the new username to return the same secret, got %q (error: %v)`, renamedSecret, err)
	}
}

// createGoogleReaderTestUser creates a user with the Google Reader integration enabled and returns its ID and Google Reader username.
func createGoogleReaderTestUser(t *testing.T, db *sql.DB, store *Storage, setPassword func(*model.Integration)) (int64, string) {
	t.Helper()

	userID := createTestUser(t, db)
	if _, err := db.Exec(`INSERT INTO integrations (user_id) VALUES ($1)`, userID); err != nil {
		t.Fatal(err)
	}

	integration, err := store.Integration(userID)
	if err != nil {
		t.Fatal(err)
	}

	username := "googlereader_" + crypto.GenerateRandomStringHex(10)
	integration.GoogleReaderEnabled = true
	integration.GoogleReaderUsername = username
	setPassword(integration)
	if err := store.UpdateIntegration(integration); err != nil {
		t.Fatal(err)
	}
	return userID, username
}
