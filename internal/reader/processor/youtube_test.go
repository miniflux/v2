// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"miniflux.app/v2/internal/config"
)

func TestGetYouTubeVideoIDFromURL(t *testing.T) {
	scenarios := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0", "HLrqNhgdiC0"},
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0&feature=youtu.be", "HLrqNhgdiC0"},
		{"https://example.org/test", ""},
	}
	for _, tc := range scenarios {
		result := getVideoIDFromYouTubeURL(tc.url)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %q for url %q`, result, tc.url)
		}
	}
}

func TestIsYouTubeVideoURL(t *testing.T) {
	scenarios := []struct {
		url      string
		expected bool
	}{
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0", true},
		{"https://www.youtube.com/watch?v=HLrqNhgdiC0&feature=youtu.be", true},
		{"https://example.org/test", false},
	}
	for _, tc := range scenarios {
		result := isYouTubeVideoURL(tc.url)
		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for url %q`, result, tc.url)
		}
	}
}

func configureYouTubeAPIOptions(t *testing.T, apiKey string) {
	t.Helper()

	t.Setenv("YOUTUBE_API_KEY", apiKey)
	t.Setenv("FETCHER_ALLOW_PRIVATE_NETWORKS", "1")

	parsedOptions, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf("Unable to configure test options: %v", err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}

func captureDebugLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })
	return &logs
}

func TestFetchYouTubeWatchTimeFromAPIInBulkSendsKeyInHeader(t *testing.T) {
	const apiKey = "AIzaSyTestKey"
	configureYouTubeAPIOptions(t, apiKey)
	logs := captureDebugLogs(t)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("X-Goog-Api-Key"); got != apiKey {
			t.Errorf("Expected the X-Goog-Api-Key header to be %q, got %q", apiKey, got)
		}
		if request.URL.Query().Has("key") {
			t.Errorf("Expected no key query parameter, got %q", request.URL.RawQuery)
		}
		if got := request.URL.Query().Get("id"); got != "video1,video2" {
			t.Errorf("Expected the id query parameter to be %q, got %q", "video1,video2", got)
		}
		if got := request.URL.Query().Get("part"); got != "contentDetails" {
			t.Errorf("Expected the part query parameter to be %q, got %q", "contentDetails", got)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write([]byte(`{"items":[{"id":"video1","contentDetails":{"duration":"PT1M30S"}},{"id":"video2","contentDetails":{"duration":"PT1H"}}]}`))
	}))
	defer server.Close()

	watchTimeMap, err := fetchYouTubeWatchTimeFromApiInBulk(server.URL, []string{"video1", "video2"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedWatchTimes := map[string]time.Duration{
		"video1": 90 * time.Second,
		"video2": time.Hour,
	}
	for videoID, expectedWatchTime := range expectedWatchTimes {
		if watchTimeMap[videoID] != expectedWatchTime {
			t.Errorf("Expected watch time %v for %q, got %v", expectedWatchTime, videoID, watchTimeMap[videoID])
		}
	}

	if strings.Contains(logs.String(), apiKey) {
		t.Errorf("Logs contain the API key:\n%s", logs.String())
	}
}

func TestFetchYouTubeWatchTimeFromAPIInBulkKeepsKeyOutOfErrors(t *testing.T) {
	const apiKey = "AIzaSyTestKey"
	configureYouTubeAPIOptions(t, apiKey)
	logs := captureDebugLogs(t)

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	serverURL := server.URL
	server.Close()

	_, err := fetchYouTubeWatchTimeFromApiInBulk(serverURL, []string{"video1"})
	if err == nil {
		t.Fatal("Expected an error when the API is unreachable")
	}

	if strings.Contains(err.Error(), apiKey) {
		t.Errorf("Error contains the API key: %v", err)
	}
	if strings.Contains(logs.String(), apiKey) {
		t.Errorf("Logs contain the API key:\n%s", logs.String())
	}
}

func TestFetchYouTubeWatchTimeFromAPIInBulkDoesNotFollowRedirects(t *testing.T) {
	const apiKey = "AIzaSyTestKey"
	configureYouTubeAPIOptions(t, apiKey)

	redirectTarget := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Errorf("Expected the redirect not to be followed, got a request with X-Goog-Api-Key %q", request.Header.Get("X-Goog-Api-Key"))
	}))
	defer redirectTarget.Close()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, redirectTarget.URL, http.StatusFound)
	}))
	defer server.Close()

	if _, err := fetchYouTubeWatchTimeFromApiInBulk(server.URL, []string{"video1"}); err == nil {
		t.Fatal("Expected an error when the API responds with a redirect")
	}
}
