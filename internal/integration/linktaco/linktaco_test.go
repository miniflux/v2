// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linktaco

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateBookmark(t *testing.T) {
	tests := []struct {
		name           string
		apiToken       string
		orgSlug        string
		tags           string
		visibility     string
		entryURL       string
		entryTitle     string
		entryContent   string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name:         "successful bookmark creation",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			tags:         "tag1, tag2",
			visibility:   "PUBLIC",
			entryURL:     "https://example.com",
			entryTitle:   "Test Article",
			entryContent: "Test content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
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
				var req map[string]interface{}
				if err := json.Unmarshal(body, &req); err != nil {
					t.Errorf("Failed to parse request body: %v", err)
				}

				// Verify mutation exists
				if _, ok := req["query"]; !ok {
					t.Error("Missing 'query' field in request")
				}

				// Return success response
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"addLink": map[string]interface{}{
							"id":    "123",
							"url":   "https://example.com",
							"title": "Test Article",
						},
					},
				})
			},
			wantErr: false,
		},
		{
			name:         "missing API token",
			apiToken:     "",
			orgSlug:      "test-org",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Should not be called
				t.Error("Server should not be called when API token is missing")
			},
			wantErr:     true,
			errContains: "missing API token or organization slug",
		},
		{
			name:         "missing organization slug",
			apiToken:     "test-token",
			orgSlug:      "",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				// Should not be called
				t.Error("Server should not be called when org slug is missing")
			},
			wantErr:     true,
			errContains: "missing API token or organization slug",
		},
		{
			name:         "GraphQL error response",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"message": "Invalid input",
						},
					},
				})
			},
			wantErr:     true,
			errContains: "Invalid input",
		},
		{
			name:         "HTTP error status",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr:     true,
			errContains: "status=401",
		},
		{
			name:         "private visibility permission error",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			visibility:   "PRIVATE",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errors": []interface{}{
						map[string]interface{}{
							"message": "PRIVATE visibility requires a paid LinkTaco account",
						},
					},
				})
			},
			wantErr:     true,
			errContains: "PRIVATE visibility requires a paid LinkTaco account",
		},
		{
			name:         "content truncation",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: strings.Repeat("a", 600), // Content longer than 500 chars
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)

				// Check that description was truncated
				variables := req["variables"].(map[string]interface{})
				input := variables["input"].(map[string]interface{})
				description := input["description"].(string)

				if len(description) != maxDescriptionLength {
					t.Errorf("Expected description length %d, got %d", maxDescriptionLength, len(description))
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"addLink": map[string]interface{}{"id": "123"},
					},
				})
			},
			wantErr: false,
		},
		{
			name:         "tag limiting",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			tags:         "tag1,tag2,tag3,tag4,tag5,tag6,tag7,tag8,tag9,tag10,tag11,tag12",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				var req map[string]interface{}
				json.Unmarshal(body, &req)

				// Check that only 10 tags were sent
				variables := req["variables"].(map[string]interface{})
				input := variables["input"].(map[string]interface{})
				tags := input["tags"].(string)

				tagCount := len(strings.Split(tags, ","))
				if tagCount != maxTags {
					t.Errorf("Expected %d tags, got %d", maxTags, tagCount)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": map[string]interface{}{
						"addLink": map[string]interface{}{"id": "123"},
					},
				})
			},
			wantErr: false,
		},
		{
			name:         "invalid JSON response",
			apiToken:     "test-token",
			orgSlug:      "test-org",
			entryURL:     "https://example.com",
			entryTitle:   "Test",
			entryContent: "Content",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			wantErr:     true,
			errContains: "unable to decode response",
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
			client := &Client{
				graphqlURL: serverURL,
				apiToken:   tt.apiToken,
				orgSlug:    tt.orgSlug,
				tags:       tt.tags,
				visibility: tt.visibility,
			}

			// Call CreateBookmark
			err := client.CreateBookmark(tt.entryURL, tt.entryTitle, tt.entryContent)

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
		name               string
		apiToken           string
		orgSlug            string
		tags               string
		visibility         string
		expectedVisibility string
	}{
		{
			name:               "with all parameters",
			apiToken:           "token",
			orgSlug:            "org",
			tags:               "tag1,tag2",
			visibility:         "PRIVATE",
			expectedVisibility: "PRIVATE",
		},
		{
			name:               "empty visibility defaults to PUBLIC",
			apiToken:           "token",
			orgSlug:            "org",
			tags:               "tag1",
			visibility:         "",
			expectedVisibility: "PUBLIC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiToken, tt.orgSlug, tt.tags, tt.visibility)

			if client.apiToken != tt.apiToken {
				t.Errorf("Expected apiToken %s, got %s", tt.apiToken, client.apiToken)
			}
			if client.orgSlug != tt.orgSlug {
				t.Errorf("Expected orgSlug %s, got %s", tt.orgSlug, client.orgSlug)
			}
			if client.tags != tt.tags {
				t.Errorf("Expected tags %s, got %s", tt.tags, client.tags)
			}
			if client.visibility != tt.expectedVisibility {
				t.Errorf("Expected visibility %s, got %s", tt.expectedVisibility, client.visibility)
			}
			if client.graphqlURL != defaultGraphQLURL {
				t.Errorf("Expected graphqlURL %s, got %s", defaultGraphQLURL, client.graphqlURL)
			}
		})
	}
}

func TestGraphQLMutation(t *testing.T) {
	// Test that the GraphQL mutation is properly formatted
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request: %v", err)
		}

		// Verify mutation structure
		query, ok := req["query"].(string)
		if !ok {
			t.Fatal("Missing query field")
		}

		// Check that mutation contains expected parts
		if !strings.Contains(query, "mutation AddLink") {
			t.Error("Mutation should contain 'mutation AddLink'")
		}
		if !strings.Contains(query, "$input: LinkInput!") {
			t.Error("Mutation should contain input parameter")
		}
		if !strings.Contains(query, "addLink(input: $input)") {
			t.Error("Mutation should contain addLink call")
		}

		// Verify variables structure
		variables, ok := req["variables"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing variables field")
		}

		input, ok := variables["input"].(map[string]interface{})
		if !ok {
			t.Fatal("Missing input in variables")
		}

		// Check all required fields
		requiredFields := []string{"url", "title", "description", "orgSlug", "visibility", "unread", "starred", "archive", "tags"}
		for _, field := range requiredFields {
			if _, ok := input[field]; !ok {
				t.Errorf("Missing required field: %s", field)
			}
		}

		// Return success
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"addLink": map[string]interface{}{
					"id": "123",
				},
			},
		})
	}))
	defer server.Close()

	client := &Client{
		graphqlURL: server.URL,
		apiToken:   "test-token",
		orgSlug:    "test-org",
		tags:       "test",
		visibility: "PUBLIC",
	}

	err := client.CreateBookmark("https://example.com", "Test Title", "Test Content")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func BenchmarkCreateBookmark(b *testing.B) {
	// Create a mock server that always returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"addLink": map[string]interface{}{
					"id": "123",
				},
			},
		})
	}))
	defer server.Close()

	client := &Client{
		graphqlURL: server.URL,
		apiToken:   "test-token",
		orgSlug:    "test-org",
		tags:       "tag1,tag2,tag3",
		visibility: "PUBLIC",
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.CreateBookmark("https://example.com", "Test Title", "Test Content")
	}
}

func BenchmarkTagProcessing(b *testing.B) {
	// Benchmark tag splitting and limiting
	tags := "tag1,tag2,tag3,tag4,tag5,tag6,tag7,tag8,tag9,tag10,tag11,tag12,tag13,tag14,tag15"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tagsSplitFn := func(c rune) bool {
			return c == ',' || c == ' '
		}
		splitTags := strings.FieldsFunc(tags, tagsSplitFn)
		if len(splitTags) > maxTags {
			splitTags = splitTags[:maxTags]
		}
		_ = strings.Join(splitTags, ",")
	}
}
