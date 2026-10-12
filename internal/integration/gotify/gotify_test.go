// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package gotify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestSendMessageSuccess(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var receivedToken string
	var receivedContentType string
	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/message" {
			t.Errorf("expected path /message, got %s", r.URL.Path)
		}

		receivedToken = r.Header.Get("X-Gotify-Key")
		receivedContentType = r.Header.Get("Content-Type")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedBody); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1,"appid":1,"message":"test message","title":"test title","priority":5}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token-123", 5)

	extras := map[string]any{
		"client::notification": map[string]any{
			"click": map[string]any{
				"url": "https://example.com/article",
			},
		},
	}

	err := client.SendMessage("test title", "test message", 5, extras)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedToken != "test-token-123" {
		t.Errorf("expected token %q, got %q", "test-token-123", receivedToken)
	}
	if receivedContentType != "application/json" {
		t.Errorf("expected content type %q, got %q", "application/json", receivedContentType)
	}
	if receivedBody["title"] != "test title" {
		t.Errorf("expected title %q, got %v", "test title", receivedBody["title"])
	}
	if receivedBody["message"] != "test message" {
		t.Errorf("expected message %q, got %v", "test message", receivedBody["message"])
	}
	if int(receivedBody["priority"].(float64)) != 5 {
		t.Errorf("expected priority 5, got %v", receivedBody["priority"])
	}

	notificationExtra, ok := receivedBody["extras"].(map[string]any)["client::notification"].(map[string]any)
	if !ok {
		t.Fatalf("expected client::notification extra, got %v", receivedBody["extras"])
	}
	clickExtra, ok := notificationExtra["click"].(map[string]any)
	if !ok {
		t.Fatalf("expected click extra, got %v", notificationExtra)
	}
	if clickExtra["url"] != "https://example.com/article" {
		t.Errorf("expected click url %q, got %v", "https://example.com/article", clickExtra["url"])
	}
}

func TestSendMessageWithoutPriorityOrExtras(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var receivedBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedBody); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":2,"message":"plain message"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token", 0)
	err := client.SendMessage("", "plain message", 0, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["message"] != "plain message" {
		t.Errorf("expected message %q, got %v", "plain message", receivedBody["message"])
	}
	if _, exists := receivedBody["title"]; exists {
		t.Errorf("expected title to be omitted, got %v", receivedBody["title"])
	}
	if _, exists := receivedBody["priority"]; exists {
		t.Errorf("expected priority to be omitted, got %v", receivedBody["priority"])
	}
	if _, exists := receivedBody["extras"]; exists {
		t.Errorf("expected extras to be omitted, got %v", receivedBody["extras"])
	}
}

func TestSendMessageEndpointFormats(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/message" {
			t.Errorf("expected path /message, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	testURLs := []string{
		server.URL,
		server.URL + "/",
		server.URL + "/message",
		server.URL + "/message/",
	}

	for _, u := range testURLs {
		client := NewClient(u, "test-token", 0)
		if err := client.SendMessage("title", "msg", 0, nil); err != nil {
			t.Errorf("failed for URL %q: %v", u, err)
		}
	}

	if requestCount != len(testURLs) {
		t.Errorf("expected %d requests, got %d", len(testURLs), requestCount)
	}
}

func TestSendMessageValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		token       string
		message     string
		priority    int
		errContains string
	}{
		{
			name:        "missing base URL",
			baseURL:     "",
			token:       "token",
			message:     "hello",
			priority:    0,
			errContains: "gotify: missing base URL or token",
		},
		{
			name:        "missing token",
			baseURL:     "https://gotify.example.com",
			token:       "",
			message:     "hello",
			priority:    0,
			errContains: "gotify: missing base URL or token",
		},
		{
			name:        "missing message",
			baseURL:     "https://gotify.example.com",
			token:       "token",
			message:     "",
			priority:    0,
			errContains: "gotify: missing message",
		},
		{
			name:        "negative priority",
			baseURL:     "https://gotify.example.com",
			token:       "token",
			message:     "hello",
			priority:    -1,
			errContains: "gotify: priority must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.token, 0)
			err := client.SendMessage("title", tt.message, tt.priority, nil)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestSendMessageServerErrors(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	tests := []struct {
		name        string
		statusCode  int
		response    string
		errContains string
	}{
		{
			name:        "unauthorized 401",
			statusCode:  http.StatusUnauthorized,
			response:    `{"error":"Unauthorized","errorCode":401}`,
			errContains: "status=401",
		},
		{
			name:        "bad request 400",
			statusCode:  http.StatusBadRequest,
			response:    `{"error":"Bad Request","errorCode":400}`,
			errContains: "status=400",
		},
		{
			name:        "internal server error 500",
			statusCode:  http.StatusInternalServerError,
			response:    `{"error":"Internal Error","errorCode":500}`,
			errContains: "status=500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient(server.URL, "token", 0)
			err := client.SendMessage("title", "msg", 0, nil)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
			}
		})
	}
}

func TestSendMessages(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	var recordedRequests []map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		recordedRequests = append(recordedRequests, parsed)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "token", 6)
	feed := &model.Feed{Title: "Feed Title"}
	entries := model.Entries{
		{
			Title: "First Article",
			URL:   "https://example.com/first",
		},
		{
			Title: "",
			URL:   "https://example.com/second",
		},
		nil,
	}

	err := client.SendMessages(feed, entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recordedRequests) != 2 {
		t.Fatalf("expected 2 recorded requests, got %d", len(recordedRequests))
	}

	req1 := recordedRequests[0]
	if req1["title"] != "Feed Title" {
		t.Errorf("req1: expected title %q, got %v", "Feed Title", req1["title"])
	}
	if req1["message"] != "First Article" {
		t.Errorf("req1: expected message %q, got %v", "First Article", req1["message"])
	}
	if int(req1["priority"].(float64)) != 6 {
		t.Errorf("req1: expected priority 6, got %v", req1["priority"])
	}
	extra1 := req1["extras"].(map[string]any)["client::notification"].(map[string]any)["click"].(map[string]any)
	if extra1["url"] != "https://example.com/first" {
		t.Errorf("req1: expected click url %q, got %v", "https://example.com/first", extra1["url"])
	}

	req2 := recordedRequests[1]
	if req2["title"] != "Feed Title" {
		t.Errorf("req2: expected title %q, got %v", "Feed Title", req2["title"])
	}
	if req2["message"] != "https://example.com/second" {
		t.Errorf("req2: expected fallback message %q, got %v", "https://example.com/second", req2["message"])
	}
	extra2 := req2["extras"].(map[string]any)["client::notification"].(map[string]any)["click"].(map[string]any)
	if extra2["url"] != "https://example.com/second" {
		t.Errorf("req2: expected click url %q, got %v", "https://example.com/second", extra2["url"])
	}
}

func TestPayloadSchemaValidation(t *testing.T) {
	t.Run("valid minimal payload", func(t *testing.T) {
		payload := &Payload{
			Message: "A minimal notification message",
		}
		if err := payload.Validate(); err != nil {
			t.Fatalf("expected valid payload, got error: %v", err)
		}

		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		if decoded["message"] != "A minimal notification message" {
			t.Errorf("expected message %q, got %v", "A minimal notification message", decoded["message"])
		}
		if _, ok := decoded["title"]; ok {
			t.Errorf("title should be omitted when empty")
		}
		if _, ok := decoded["priority"]; ok {
			t.Errorf("priority should be omitted when zero")
		}
		if _, ok := decoded["extras"]; ok {
			t.Errorf("extras should be omitted when nil")
		}
	})

	t.Run("valid complete payload with notification click extras", func(t *testing.T) {
		payload := &Payload{
			Title:    "Feed Update",
			Message:  "New post published",
			Priority: 5,
			Extras: map[string]any{
				"client::notification": map[string]any{
					"click": map[string]any{
						"url": "https://example.com/post",
					},
				},
			},
		}
		if err := payload.Validate(); err != nil {
			t.Fatalf("expected valid payload, got error: %v", err)
		}

		data, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		if decoded["title"] != "Feed Update" {
			t.Errorf("expected title %q, got %v", "Feed Update", decoded["title"])
		}
		if decoded["message"] != "New post published" {
			t.Errorf("expected message %q, got %v", "New post published", decoded["message"])
		}
		if int(decoded["priority"].(float64)) != 5 {
			t.Errorf("expected priority 5, got %v", decoded["priority"])
		}

		extras, ok := decoded["extras"].(map[string]any)
		if !ok {
			t.Fatalf("expected extras map, got %v", decoded["extras"])
		}
		notif, ok := extras["client::notification"].(map[string]any)
		if !ok {
			t.Fatalf("expected client::notification map, got %v", extras)
		}
		click, ok := notif["click"].(map[string]any)
		if !ok {
			t.Fatalf("expected click map, got %v", notif)
		}
		if click["url"] != "https://example.com/post" {
			t.Errorf("expected click url %q, got %v", "https://example.com/post", click["url"])
		}
	})

	t.Run("invalid payload with empty message", func(t *testing.T) {
		payload := &Payload{
			Title:   "Some Title",
			Message: "",
		}
		err := payload.Validate()
		if err == nil {
			t.Fatalf("expected error for empty message, got nil")
		}
		if !strings.Contains(err.Error(), "gotify: missing message") {
			t.Errorf("expected missing message error, got %v", err)
		}
	})

	t.Run("invalid payload with negative priority", func(t *testing.T) {
		payload := &Payload{
			Message:  "Valid message",
			Priority: -2,
		}
		err := payload.Validate()
		if err == nil {
			t.Fatalf("expected error for negative priority, got nil")
		}
		if !strings.Contains(err.Error(), "gotify: priority must be non-negative") {
			t.Errorf("expected non-negative priority error, got %v", err)
		}
	})
}

func TestSendMessagesMissingConfigOrEmptyEntries(t *testing.T) {
	configureIntegrationAllowPrivateNetworksOption(t)

	t.Run("missing base URL returns error", func(t *testing.T) {
		client := NewClient("", "token", 0)
		err := client.SendMessages(&model.Feed{}, model.Entries{{Title: "Test"}})
		if err == nil {
			t.Fatal("expected error for missing base URL, got nil")
		}
		if !strings.Contains(err.Error(), "gotify: missing base URL or token") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing token returns error", func(t *testing.T) {
		client := NewClient("https://gotify.example.com", "", 0)
		err := client.SendMessages(&model.Feed{}, model.Entries{{Title: "Test"}})
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if !strings.Contains(err.Error(), "gotify: missing base URL or token") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("skip entries with empty title and empty url", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":1}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, "token", 0)
		feed := &model.Feed{Title: "Feed"}
		entries := model.Entries{
			nil,
			{Title: "", URL: ""},
			{Title: "Valid Title", URL: "https://example.com/item"},
		}

		if err := client.SendMessages(feed, entries); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if requestCount != 1 {
			t.Errorf("expected 1 request for valid entry, got %d", requestCount)
		}
	})
}

func TestNewClientPriorityClamping(t *testing.T) {
	client := NewClient("https://gotify.example.com", "token", -5)
	if client.priority != 0 {
		t.Errorf("expected negative priority clamped to 0, got %d", client.priority)
	}
}

func configureIntegrationAllowPrivateNetworksOption(t *testing.T) {
	t.Helper()

	t.Setenv("INTEGRATION_ALLOW_PRIVATE_NETWORKS", "1")

	configParser := config.NewConfigParser()
	parsedOptions, err := configParser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}
