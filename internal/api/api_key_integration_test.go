// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"errors"
	"testing"

	miniflux "miniflux.app/v2/client"
	"miniflux.app/v2/internal/http/server/servertest"
)

func TestAPIKeysEndpoint(t *testing.T) {
	t.Parallel()

	regularUserClient, regularTestUser := newRegularUser(t)
	apiKeys, err := regularUserClient.APIKeys()
	if err != nil {
		t.Fatal(err)
	}

	if len(apiKeys) != 0 {
		t.Fatalf(`Expected no API keys, got %d`, len(apiKeys))
	}

	// Create an API key for the user.
	apiKey, err := regularUserClient.CreateAPIKey("Test API Key")
	if err != nil {
		t.Fatal(err)
	}
	if apiKey.ID == 0 {
		t.Fatalf(`Invalid API key ID, got "%v"`, apiKey.ID)
	}
	if apiKey.UserID != regularTestUser.ID {
		t.Fatalf(`Invalid user ID for API key, got "%v" instead of "%v"`, apiKey.UserID, regularTestUser.ID)
	}
	if apiKey.Token == "" {
		t.Fatalf(`Invalid API key token, got "%v"`, apiKey.Token)
	}
	if apiKey.Description != "Test API Key" {
		t.Fatalf(`Invalid API key description, got "%v" instead of "Test API Key"`, apiKey.Description)
	}

	// Create a duplicate API key with the same description.
	if _, err := regularUserClient.CreateAPIKey("Test API Key"); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating a duplicate API key with the same description should raise an error, got %v`, err)
	}

	// Fetch the API keys again.
	apiKeys, err = regularUserClient.APIKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(apiKeys) != 1 {
		t.Fatalf(`Expected 1 API key, got %d`, len(apiKeys))
	}
	if apiKeys[0].ID != apiKey.ID {
		t.Fatalf(`Invalid API key ID, got "%v" instead of "%v"`, apiKeys[0].ID, apiKey.ID)
	}
	if apiKeys[0].UserID != regularTestUser.ID {
		t.Fatalf(`Invalid user ID for API key, got "%v" instead of "%v"`, apiKeys[0].UserID, regularTestUser.ID)
	}
	if apiKeys[0].Token != apiKey.Token {
		t.Fatalf(`Invalid API key token, got "%v" instead of "%v"`, apiKeys[0].Token, apiKey.Token)
	}
	if apiKeys[0].Description != "Test API Key" {
		t.Fatalf(`Invalid API key description, got "%v" instead of "Test API Key"`, apiKeys[0].Description)
	}

	// Create a new client using the API key.
	apiKeyClient := miniflux.NewClient(servertest.Require(t).URL, apiKey.Token)

	// Fetch the user using the API key client.
	user, err := apiKeyClient.Me()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the user matches the regular test user.
	if user.ID != regularTestUser.ID {
		t.Fatalf(`Expected user ID %d, got %d`, regularTestUser.ID, user.ID)
	}

	// Delete the API key.
	if err := regularUserClient.DeleteAPIKey(apiKey.ID); err != nil {
		t.Fatal(err)
	}

	// Verify the API key is deleted.
	apiKeys, err = regularUserClient.APIKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(apiKeys) != 0 {
		t.Fatalf(`Expected no API keys after deletion, got %d`, len(apiKeys))
	}

	// Try to delete the API key again, it should return an error.
	err = regularUserClient.DeleteAPIKey(apiKey.ID)
	if !errors.Is(err, miniflux.ErrNotFound) {
		t.Fatalf(`Deleting a non-existent API key should raise a "not found" error, got %v`, err)
	}

	// Try to create an API key with an empty description.
	if _, err := regularUserClient.CreateAPIKey(""); !errors.Is(err, miniflux.ErrBadRequest) {
		t.Fatalf(`Creating an API key with an empty description should raise an error, got %v`, err)
	}
}
