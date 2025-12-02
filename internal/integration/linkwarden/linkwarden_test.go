// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkwarden

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestCreateBookmark(t *testing.T) {
	tests := []struct {
		name           string
		baseURL        string
		apiKey         string
		collectionID   *int64
		entryURL       string
		entryTitle     string
		serverResponse func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionID *int64)
		wantErr        bool
		errContains    string
	}{
		{
			name:         "successful bookmark creation without collection",
			baseURL:      "",
			apiKey:       "test-api-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test Article",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-api-key" {
					t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", auth)
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

				// Verify URL
				if reqURL := req["url"]; reqURL != "https://example.com" {
					t.Errorf("Expected URL 'https://example.com', got %v", reqURL)
				}

				// Verify title/name
				if reqName := req["name"]; reqName != "Test Article" {
					t.Errorf("Expected name 'Test Article', got %v", reqName)
				}

				// Verify collection is not present when nil
				if _, ok := req["collection"]; ok {
					t.Error("Expected collection field to be omitted when collectionId is nil")
				}

				// Return success response
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":   "123",
					"url":  "https://example.com",
					"name": "Test Article",
				})
			},
			wantErr: false,
		},
		{
			name:         "successful bookmark creation with collection",
			baseURL:      "",
			apiKey:       "test-api-key",
			collectionID: model.OptionalNumber(int64(42)),
			entryURL:     "https://example.com/article",
			entryTitle:   "Test Article With Collection",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionID *int64) {
				// Verify authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-api-key" {
					t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", auth)
				}

				// Parse and verify request
				body, _ := io.ReadAll(r.Body)
				var req map[string]any
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("Failed to parse request body: %v", err)
				}

				// Verify URL
				if reqURL := req["url"]; reqURL != "https://example.com/article" {
					t.Errorf("Expected URL 'https://example.com/article', got %v", reqURL)
				}

				// Verify title/name
				if reqName := req["name"]; reqName != "Test Article With Collection" {
					t.Errorf("Expected name 'Test Article With Collection', got %v", reqName)
				}

				// Verify collection is present and correct
				if collection, ok := req["collection"]; ok {
					collectionMap, ok := collection.(map[string]any)
					if !ok {
						t.Error("Expected collection to be a map")
					}
					if collectionID, ok := collectionMap["id"]; ok {
						// JSON numbers are float64
						if collectionIDFloat, ok := collectionID.(float64); !ok || int64(collectionIDFloat) != 42 {
							t.Errorf("Expected collection id 42, got %v", collectionID)
						}
					} else {
						t.Error("Expected collection to have 'id' field")
					}
				} else {
					t.Error("Expected collection field to be present when collectionId is set")
				}

				// Return success response
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{
					"id":   "124",
					"url":  "https://example.com/article",
					"name": "Test Article With Collection",
				})
			},
			wantErr: false,
		},
		{
			name:         "missing API key",
			baseURL:      "",
			apiKey:       "",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				// Should not be called
				t.Error("Server should not be called when API key is missing")
			},
			wantErr:     true,
			errContains: "missing base URL or API key",
		},
		{
			name:         "server error",
			baseURL:      "",
			apiKey:       "test-api-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			},
			wantErr:     true,
			errContains: "unable to create link: status=500",
		},
		{
			name:         "bad request with null collection id error",
			baseURL:      "",
			apiKey:       "test-api-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"response":"Error: Expected number, received null [collection, id]"}`))
			},
			wantErr:     true,
			errContains: "unable to create link: status=400",
		},
		{
			name:         "unauthorized",
			baseURL:      "",
			apiKey:       "invalid-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized"}`))
			},
			wantErr:     true,
			errContains: "unable to create link: status=401",
		},
		{
			name:         "invalid base URL",
			baseURL:      ":",
			apiKey:       "test-api-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				// Should not be called
				t.Error("Server should not be called when base URL is invalid")
			},
			wantErr:     true,
			errContains: "invalid API endpoint",
		},
		{
			name:         "missing base URL",
			baseURL:      "",
			apiKey:       "",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				// Should not be called
				t.Error("Server should not be called when base URL is missing")
			},
			wantErr:     true,
			errContains: "missing base URL or API key",
		},
		{
			name:         "network connection error",
			baseURL:      "http://localhost:1", // Invalid port that should fail to connect
			apiKey:       "test-api-key",
			collectionID: nil,
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			serverResponse: func(w http.ResponseWriter, r *http.Request, t *testing.T, collectionId *int64) {
				// Should not be called due to connection failure
				t.Error("Server should not be called when connection fails")
			},
			wantErr:     true,
			errContains: "unable to send request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server only if we have a valid apiKey and don't have a custom baseURL for error testing
			var server *httptest.Server
			if tt.apiKey != "" && tt.baseURL != ":" && tt.baseURL != "http://localhost:1" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					tt.serverResponse(w, r, t, tt.collectionID)
				}))
				defer server.Close()
			}

			// Use test server URL if baseURL is empty and we have a server
			baseURL := tt.baseURL
			if baseURL == "" && server != nil {
				baseURL = server.URL
			}

			// Create client
			client := NewClient(baseURL, tt.apiKey, tt.collectionID)

			// Call CreateBookmark
			err := client.CreateBookmark(tt.entryURL, tt.entryTitle)

			// Check error
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		apiKey       string
		collectionID *int64
	}{
		{
			name:         "client without collection",
			baseURL:      "https://linkwarden.example.com",
			apiKey:       "test-key",
			collectionID: nil,
		},
		{
			name:         "client with collection",
			baseURL:      "https://linkwarden.example.com",
			apiKey:       "test-key",
			collectionID: model.OptionalNumber(int64(123)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.apiKey, tt.collectionID)

			if client.baseURL != tt.baseURL {
				t.Errorf("Expected baseURL %s, got %s", tt.baseURL, client.baseURL)
			}

			if client.apiKey != tt.apiKey {
				t.Errorf("Expected apiKey %s, got %s", tt.apiKey, client.apiKey)
			}

			if tt.collectionID == nil {
				if client.collectionID != nil {
					t.Errorf("Expected collectionId to be nil, got %v", *client.collectionID)
				}
			} else {
				if client.collectionID == nil {
					t.Error("Expected collectionId to be set, got nil")
				} else if *client.collectionID != *tt.collectionID {
					t.Errorf("Expected collectionId %d, got %d", *tt.collectionID, *client.collectionID)
				}
			}
		})
	}
}
