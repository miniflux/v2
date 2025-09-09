// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package wallabag

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateEntry(t *testing.T) {
	entryURL := "https://example.com"
	entryTitle := "title"
	entryContent := "content"
	tags := "tag1,tag2,tag3"

	tests := []struct {
		name           string
		username       string
		password       string
		clientID       string
		clientSecret   string
		tags           string
		onlyURL        bool
		entryURL       string
		entryTitle     string
		entryContent   string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name:         "successful entry creation with url only",
			wantErr:      false,
			onlyURL:      true,
			username:     "username",
			password:     "password",
			clientID:     "clientId",
			clientSecret: "clientSecret",
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/v2/token") {
					// Return success response
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]any{
						"access_token":  "test-token",
						"expires_in":    3600,
						"refresh_token": "token",
						"scope":         "scope",
						"token_type":    "token_type",
					})
					return
				}
				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
				}
				// Verify content type
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
				}
				// Parse and verify request
				body, _ := io.ReadAll(r.Body)
				var req map[string]any
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("Failed to parse request body: %v", err)
				}
				if requstEntryURL := req["url"]; requstEntryURL != entryURL {
					t.Errorf("Expected entryURL %s, got %s", entryURL, requstEntryURL)
				}
				if requestEntryTitle := req["title"]; requestEntryTitle != entryTitle {
					t.Errorf("Expected entryTitle %s, got %s", entryTitle, requestEntryTitle)
				}
				if _, ok := req["content"]; ok {
					t.Errorf("Expected entryContent to be empty, got value")
				}
				if requestTags := req["tags"]; requestTags != tags {
					t.Errorf("Expected tags %s, got %s", tags, requestTags)
				} // Return success response
				w.WriteHeader(http.StatusOK)
			},
			errContains: "",
		},
		{
			name:         "successful entry creation with content",
			wantErr:      false,
			onlyURL:      false,
			username:     "username",
			password:     "password",
			clientID:     "clientId",
			clientSecret: "clientSecret",
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/v2/token") {
					// Return success response
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]any{
						"access_token":  "test-token",
						"expires_in":    3600,
						"refresh_token": "token",
						"scope":         "scope",
						"token_type":    "token_type",
					})
					return
				}
				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-token" {
					t.Errorf("Expected Authorization header 'Bearer test-token', got %s", auth)
				}
				// Verify content type
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
				}
				// Parse and verify request
				body, _ := io.ReadAll(r.Body)
				var req map[string]any
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("Failed to parse request body: %v", err)
				}
				if requstEntryURL := req["url"]; requstEntryURL != entryURL {
					t.Errorf("Expected entryURL %s, got %s", entryURL, requstEntryURL)
				}
				if requestEntryTitle := req["title"]; requestEntryTitle != entryTitle {
					t.Errorf("Expected entryTitle %s, got %s", entryTitle, requestEntryTitle)
				}
				if requestEntryContent := req["content"]; requestEntryContent != entryContent {
					t.Errorf("Expected entryContent %s, got %s", entryContent, requestEntryContent)
				}
				if requestTags := req["tags"]; requestTags != tags {
					t.Errorf("Expected tags %s, got %s", tags, requestTags)
				} // Return success response
				w.WriteHeader(http.StatusOK)
			},
			errContains: "",
		},
		{
			name:         "failed when unable to decode accessToken response",
			wantErr:      true,
			onlyURL:      true,
			username:     "username",
			password:     "password",
			clientID:     "clientId",
			clientSecret: "clientSecret",
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/v2/token") {
					// Return success response
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
					return
				}
				t.Error("Server should not be called when failed to get accessToken")
			},
			errContains: "unable to decode token response",
		},
		{
			name:         "failed when saving entry",
			wantErr:      true,
			onlyURL:      true,
			username:     "username",
			password:     "password",
			clientID:     "clientId",
			clientSecret: "clientSecret",
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/v2/token") {
					// Return success response
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]any{
						"access_token":  "test-token",
						"expires_in":    3600,
						"refresh_token": "token",
						"scope":         "scope",
						"token_type":    "token_type",
					})
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
			},
			errContains: "unable to get save entry",
		},
		{
			name:         "failure due to no accessToken",
			wantErr:      true,
			onlyURL:      false,
			username:     "username",
			password:     "password",
			clientID:     "clientId",
			clientSecret: "clientSecret",
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/oauth/v2/token") {
					// Return error response
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				t.Error("Server should not be called when failed to get accessToken")
			},
			errContains: "unable to get access token",
		},
		{
			name:         "failure due to missing client parameters",
			wantErr:      true,
			onlyURL:      false,
			tags:         tags,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Server should not be called when failed to get accessToken")
			},
			errContains: "wallabag: missing base URL, client ID, client secret, username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server if we have a server response function
			var serverURL string
			if tt.serverResponse != nil {
				server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer server.Close()
				serverURL = server.URL
			}

			// Create client with test server URL
			client := NewClient(serverURL, tt.clientID, tt.clientSecret, tt.username, tt.password, tt.tags, tt.onlyURL)

			// Call CreateBookmark
			err := client.CreateEntry(tt.entryURL, tt.entryTitle, tt.entryContent)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		clientID     string
		clientSecret string
		username     string
		password     string
		tags         string
		onlyURL      bool
	}{
		{
			name:         "with all parameters",
			baseURL:      "https://wallabag.example.com",
			clientID:     "clientID",
			clientSecret: "clientSecret",
			username:     "wallabag",
			password:     "wallabag",
			tags:         "",
			onlyURL:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.clientID, tt.clientSecret, tt.username, tt.password, tt.tags, tt.onlyURL)

			if client.baseURL != tt.baseURL {
				t.Errorf("Expected.baseURL %s, got %s", tt.baseURL, client.baseURL)
			}
			if client.username != tt.username {
				t.Errorf("Expected username %s, got %s", tt.username, client.username)
			}
			if client.password != tt.password {
				t.Errorf("Expected password %s, got %s", tt.password, client.password)
			}
			if client.clientID != tt.clientID {
				t.Errorf("Expected clientID %s, got %s", tt.clientID, client.clientID)
			}
			if client.clientSecret != tt.clientSecret {
				t.Errorf("Expected clientSecret %s, got %s", tt.clientSecret, client.clientSecret)
			}
			if client.tags != tt.tags {
				t.Errorf("Expected tags %s, got %s", tt.tags, client.tags)
			}
			if client.onlyURL != tt.onlyURL {
				t.Errorf("Expected onlyURL %v, got %v", tt.onlyURL, client.onlyURL)
			}
		})
	}
}
