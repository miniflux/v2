// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewRequestBuilder(t *testing.T) {
	builder := NewRequestBuilder()
	if builder == nil {
		t.Fatal("NewRequestBuilder should not return nil")
	}
	if builder.clientTimeout != defaultHTTPClientTimeout {
		t.Errorf("Expected default timeout %d, got %d", defaultHTTPClientTimeout, builder.clientTimeout)
	}
	if builder.headers == nil {
		t.Fatal("Headers should be initialized")
	}
}

func TestRequestBuilder_WithHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Custom-Header") != "custom-value" {
			t.Errorf("Expected Custom-Header to be 'custom-value', got '%s'", r.Header.Get("Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.WithHeader("Custom-Header", "custom-value").ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_WithETag(t *testing.T) {
	tests := []struct {
		name     string
		etag     string
		expected string
	}{
		{"with etag", "test-etag", "test-etag"},
		{"empty etag", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("If-None-Match") != tt.expected {
					t.Errorf("Expected If-None-Match to be '%s', got '%s'", tt.expected, r.Header.Get("If-None-Match"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			builder := NewRequestBuilder()
			resp, err := builder.WithETag(tt.etag).ExecuteRequest(server.URL)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestRequestBuilder_WithLastModified(t *testing.T) {
	tests := []struct {
		name         string
		lastModified string
		expected     string
	}{
		{"with last modified", "Mon, 02 Jan 2006 15:04:05 GMT", "Mon, 02 Jan 2006 15:04:05 GMT"},
		{"empty last modified", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("If-Modified-Since") != tt.expected {
					t.Errorf("Expected If-Modified-Since to be '%s', got '%s'", tt.expected, r.Header.Get("If-Modified-Since"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			builder := NewRequestBuilder()
			resp, err := builder.WithLastModified(tt.lastModified).ExecuteRequest(server.URL)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestRequestBuilder_WithUserAgent(t *testing.T) {
	tests := []struct {
		name           string
		userAgent      string
		defaultAgent   string
		expectedHeader string
	}{
		{"custom user agent", "CustomAgent/1.0", "DefaultAgent/1.0", "CustomAgent/1.0"},
		{"default user agent", "", "DefaultAgent/1.0", "DefaultAgent/1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != tt.expectedHeader {
					t.Errorf("Expected User-Agent to be '%s', got '%s'", tt.expectedHeader, r.Header.Get("User-Agent"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			builder := NewRequestBuilder()
			resp, err := builder.WithUserAgent(tt.userAgent, tt.defaultAgent).ExecuteRequest(server.URL)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestRequestBuilder_WithCookie(t *testing.T) {
	tests := []struct {
		name     string
		cookie   string
		expected string
	}{
		{"with cookie", "session=abc123; lang=en", "session=abc123; lang=en"},
		{"empty cookie", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Cookie") != tt.expected {
					t.Errorf("Expected Cookie to be '%s', got '%s'", tt.expected, r.Header.Get("Cookie"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			builder := NewRequestBuilder()
			resp, err := builder.WithCookie(tt.cookie).ExecuteRequest(server.URL)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestRequestBuilder_WithUsernameAndPassword(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{"with credentials", "test", "password", "Basic dGVzdDpwYXNzd29yZA=="}, // base64 of "test:password"
		{"empty username", "", "password", ""},
		{"empty password", "test", "", ""},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != tt.expected {
					t.Errorf("Expected Authorization to be '%s', got '%s'", tt.expected, r.Header.Get("Authorization"))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			builder := NewRequestBuilder()
			resp, err := builder.WithUsernameAndPassword(tt.username, tt.password).ExecuteRequest(server.URL)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

func TestRequestBuilder_DefaultAcceptHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != defaultAcceptHeader {
			t.Errorf("Expected Accept to be '%s', got '%s'", defaultAcceptHeader, r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_CustomAcceptHeaderNotOverridden(t *testing.T) {
	customAccept := "application/json"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != customAccept {
			t.Errorf("Expected Accept to be '%s', got '%s'", customAccept, r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.WithHeader("Accept", customAccept).ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_WithTimeout(t *testing.T) {
	builder := NewRequestBuilder()
	builder = builder.WithTimeout(30 * time.Second)

	if builder.clientTimeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30, got %d", builder.clientTimeout)
	}
}

func TestRequestBuilder_WithoutRedirects(t *testing.T) {
	// Create a redirect server
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectServer.URL, http.StatusFound)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.WithoutRedirects().ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("Expected status code %d, got %d", http.StatusFound, resp.StatusCode)
	}
}

func TestRequestBuilder_DisableHTTP2(t *testing.T) {
	builder := NewRequestBuilder()
	builder = builder.DisableHTTP2(true)

	if !builder.disableHTTP2 {
		t.Error("Expected disableHTTP2 to be true")
	}
}

func TestRequestBuilder_IgnoreTLSErrors(t *testing.T) {
	builder := NewRequestBuilder()
	builder = builder.IgnoreTLSErrors(true)

	if !builder.ignoreTLSErrors {
		t.Error("Expected ignoreTLSErrors to be true")
	}
}

func TestRequestBuilder_WithoutCompression(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") != "identity" {
			t.Errorf("Expected Accept-Encoding to be 'identity', got '%s'", r.Header.Get("Accept-Encoding"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.WithoutCompression().ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_WithCompression(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Encoding") != "br, gzip" {
			t.Errorf("Expected Accept-Encoding to be 'br, gzip', got '%s'", r.Header.Get("Accept-Encoding"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_ConnectionCloseHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Connection") != "close" {
			t.Errorf("Expected Connection to be 'close', got '%s'", r.Header.Get("Connection"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_WithCustomApplicationProxyURL(t *testing.T) {
	proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	builder := NewRequestBuilder()
	builder = builder.WithCustomApplicationProxyURL(proxyURL)

	if builder.clientProxyURL != proxyURL {
		t.Error("Expected clientProxyURL to be set")
	}
}

func TestRequestBuilder_UseCustomApplicationProxyURL(t *testing.T) {
	builder := NewRequestBuilder()
	builder = builder.UseCustomApplicationProxyURL(true)

	if !builder.useClientProxy {
		t.Error("Expected useClientProxy to be true")
	}
}

func TestRequestBuilder_WithCustomFeedProxyURL(t *testing.T) {
	proxyURL := "http://feed-proxy.example.com:8080"
	builder := NewRequestBuilder()
	builder = builder.WithCustomFeedProxyURL(proxyURL)

	if builder.feedProxyURL != proxyURL {
		t.Errorf("Expected feedProxyURL to be '%s', got '%s'", proxyURL, builder.feedProxyURL)
	}
}

func TestRequestBuilder_ChainedMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check multiple headers
		if r.Header.Get("User-Agent") != "TestAgent/1.0" {
			t.Errorf("Expected User-Agent to be 'TestAgent/1.0', got '%s'", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Cookie") != "test=value" {
			t.Errorf("Expected Cookie to be 'test=value', got '%s'", r.Header.Get("Cookie"))
		}
		if r.Header.Get("If-None-Match") != "etag123" {
			t.Errorf("Expected If-None-Match to be 'etag123', got '%s'", r.Header.Get("If-None-Match"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	resp, err := builder.
		WithUserAgent("TestAgent/1.0", "DefaultAgent/1.0").
		WithCookie("test=value").
		WithETag("etag123").
		WithTimeout(10 * time.Second).
		ExecuteRequest(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()
}

func TestRequestBuilder_InvalidURL(t *testing.T) {
	builder := NewRequestBuilder()
	_, err := builder.ExecuteRequest("invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestRequestBuilder_TimeoutConfiguration(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	builder := NewRequestBuilder()
	start := time.Now()
	_, err := builder.WithTimeout(1 * time.Second).ExecuteRequest(server.URL)
	duration := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	// Should timeout around 1 second, allow some margin
	if duration > 1500*time.Millisecond {
		t.Errorf("Expected timeout around 1s, took %v", duration)
	}
}
