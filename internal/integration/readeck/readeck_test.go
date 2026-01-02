// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readeck

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateBookmark(t *testing.T) {
	entryURL := "https://example.com/article"
	entryTitle := "Example Title"
	entryContent := "<p>Some HTML content</p>"
	labels := "tag1,tag2"

	tests := []struct {
		name           string
		onlyURL        bool
		baseURL        string
		apiKey         string
		labels         string
		entryURL       string
		entryTitle     string
		entryContent   string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		errContains    string
	}{
		{
			name:         "successful bookmark creation with only URL",
			onlyURL:      true,
			labels:       labels,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/bookmarks/" {
					t.Errorf("expected path /api/bookmarks/, got %s", r.URL.Path)
				}
				if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
					t.Errorf("expected Authorization Bearer header, got %q", got)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}

				body, _ := io.ReadAll(r.Body)
				var payload map[string]any
				if err := json.Unmarshal(body, &payload); err != nil {
					t.Fatalf("failed to parse JSON body: %v", err)
				}
				if u := payload["url"]; u != entryURL {
					t.Errorf("expected url %s, got %v", entryURL, u)
				}
				if title := payload["title"]; title != entryTitle {
					t.Errorf("expected title %s, got %v", entryTitle, title)
				}
				// Labels should be split into an array
				if raw := payload["labels"]; raw == nil {
					t.Errorf("expected labels to be set")
				} else if arr, ok := raw.([]any); ok {
					if len(arr) != 2 || arr[0] != "tag1" || arr[1] != "tag2" {
						t.Errorf("unexpected labels: %#v", arr)
					}
				} else {
					t.Errorf("labels should be an array, got %T", raw)
				}
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name:         "successful bookmark creation with content (multipart)",
			onlyURL:      false,
			labels:       labels,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/bookmarks/" {
					t.Errorf("expected path /api/bookmarks/, got %s", r.URL.Path)
				}
				if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
					t.Errorf("expected Authorization Bearer header, got %q", got)
				}
				ct := r.Header.Get("Content-Type")
				if !strings.HasPrefix(ct, "multipart/form-data;") {
					t.Errorf("expected multipart/form-data, got %s", ct)
				}
				boundaryIdx := strings.Index(ct, "boundary=")
				if boundaryIdx == -1 {
					t.Fatalf("missing multipart boundary in Content-Type: %s", ct)
				}
				boundary := ct[boundaryIdx+len("boundary="):]
				mr := multipart.NewReader(r.Body, boundary)

				seenLabels := []string{}
				var seenURL, seenTitle, seenFeature string
				var resourceHeader map[string]any
				var resourceBody string

				for {
					part, err := mr.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatalf("reading multipart: %v", err)
					}
					name := part.FormName()
					data, _ := io.ReadAll(part)
					switch name {
					case "url":
						seenURL = string(data)
					case "title":
						seenTitle = string(data)
					case "feature_find_main":
						seenFeature = string(data)
					case "labels":
						seenLabels = append(seenLabels, string(data))
					case "resource":
						// First line is JSON header, then newline, then content
						all := string(data)
						idx := strings.IndexByte(all, '\n')
						if idx == -1 {
							t.Fatalf("resource content missing header separator")
						}
						headerJSON := all[:idx]
						resourceBody = all[idx+1:]
						if err := json.Unmarshal([]byte(headerJSON), &resourceHeader); err != nil {
							t.Fatalf("invalid resource header JSON: %v", err)
						}
					}
				}

				if seenURL != entryURL {
					t.Errorf("expected url %s, got %s", entryURL, seenURL)
				}
				if seenTitle != entryTitle {
					t.Errorf("expected title %s, got %s", entryTitle, seenTitle)
				}
				if seenFeature != "false" {
					t.Errorf("expected feature_find_main to be 'false', got %s", seenFeature)
				}
				if len(seenLabels) != 2 || seenLabels[0] != "tag1" || seenLabels[1] != "tag2" {
					t.Errorf("unexpected labels: %#v", seenLabels)
				}
				if resourceHeader == nil {
					t.Fatalf("missing resource header")
				}
				if hURL, _ := resourceHeader["url"].(string); hURL != entryURL {
					t.Errorf("expected resource header url %s, got %v", entryURL, hURL)
				}
				if headers, ok := resourceHeader["headers"].(map[string]any); ok {
					if ct, _ := headers["content-type"].(string); ct != "text/html; charset=utf-8" {
						t.Errorf("expected resource header content-type text/html; charset=utf-8, got %v", ct)
					}
				} else {
					t.Errorf("missing resource header 'headers' field")
				}
				if resourceBody != entryContent {
					t.Errorf("expected resource body %q, got %q", entryContent, resourceBody)
				}

				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name:         "error when server returns 400",
			onlyURL:      true,
			labels:       labels,
			entryURL:     entryURL,
			entryTitle:   entryTitle,
			entryContent: entryContent,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr:     true,
			errContains: "unable to create bookmark",
		},
		{
			name:           "error when missing baseURL or apiKey",
			onlyURL:        true,
			baseURL:        "",
			apiKey:         "",
			labels:         labels,
			entryURL:       entryURL,
			entryTitle:     entryTitle,
			entryContent:   entryContent,
			serverResponse: nil,
			wantErr:        true,
			errContains:    "missing base URL or API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			if tt.serverResponse != nil {
				srv := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
				defer srv.Close()
				serverURL = srv.URL
			}
			baseURL := tt.baseURL
			if baseURL == "" {
				baseURL = serverURL
			}
			apiKey := tt.apiKey
			if apiKey == "" {
				apiKey = "test-api-key"
			}

			client := NewClient(baseURL, apiKey, tt.labels, tt.onlyURL)
			err := client.CreateBookmark(tt.entryURL, tt.entryTitle, tt.entryContent)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	baseURL := "https://readeck.example.com"
	apiKey := "key"
	labels := "tag1,tag2"
	onlyURL := true

	c := NewClient(baseURL, apiKey, labels, onlyURL)
	if c.baseURL != baseURL {
		t.Errorf("expected baseURL %s, got %s", baseURL, c.baseURL)
	}
	if c.apiKey != apiKey {
		t.Errorf("expected apiKey %s, got %s", apiKey, c.apiKey)
	}
	if c.labels != labels {
		t.Errorf("expected labels %s, got %s", labels, c.labels)
	}
	if c.onlyURL != onlyURL {
		t.Errorf("expected onlyURL %v, got %v", onlyURL, c.onlyURL)
	}
}
