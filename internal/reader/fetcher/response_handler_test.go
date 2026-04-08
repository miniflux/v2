// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/klauspost/compress/zstd"
)

type testReadCloser struct {
	closed bool
}

func (rc *testReadCloser) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (rc *testReadCloser) Close() error {
	rc.closed = true
	return nil
}

func TestIsModified(t *testing.T) {
	var cachedEtag = "abc123"
	var cachedLastModified = "Wed, 21 Oct 2015 07:28:00 GMT"

	var testCases = map[string]struct {
		Status       int
		LastModified string
		ETag         string
		IsModified   bool
	}{
		"Unmodified 304": {
			Status:       304,
			LastModified: cachedLastModified,
			ETag:         cachedEtag,
			IsModified:   false,
		},
		"Unmodified 200": {
			Status:       200,
			LastModified: cachedLastModified,
			ETag:         cachedEtag,
			IsModified:   false,
		},
		// ETag takes precedence per RFC9110 8.8.1.
		"Last-Modified changed only": {
			Status:       200,
			LastModified: "Thu, 22 Oct 2015 07:28:00 GMT",
			ETag:         cachedEtag,
			IsModified:   false,
		},
		"ETag changed only": {
			Status:       200,
			LastModified: cachedLastModified,
			ETag:         "xyz789",
			IsModified:   true,
		},
		"ETag and Last-Modified changed": {
			Status:       200,
			LastModified: "Thu, 22 Oct 2015 07:28:00 GMT",
			ETag:         "xyz789",
			IsModified:   true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			header := http.Header{}
			header.Add("Last-Modified", tc.LastModified)
			header.Add("ETag", tc.ETag)
			rh := ResponseHandler{
				httpResponse: &http.Response{
					StatusCode: tc.Status,
					Header:     header,
				},
			}
			if tc.IsModified != rh.IsModified(cachedEtag, cachedLastModified) {
				tt.Error(name)
			}
		})
	}
}

func TestRetryDelay(t *testing.T) {
	var testCases = map[string]struct {
		RetryAfterHeader string
		ExpectedDelay    time.Duration
	}{
		"Empty header": {
			RetryAfterHeader: "",
			ExpectedDelay:    0,
		},
		"Integer value": {
			RetryAfterHeader: "42",
			ExpectedDelay:    42 * time.Second,
		},
		"HTTP-date": {
			RetryAfterHeader: time.Now().Add(42 * time.Second).Format(time.RFC1123),
			ExpectedDelay:    41 * time.Second,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			header := http.Header{}
			header.Add("Retry-After", tc.RetryAfterHeader)
			rh := ResponseHandler{
				httpResponse: &http.Response{
					Header: header,
				},
			}
			if tc.ExpectedDelay != rh.ParseRetryDelay() {
				tt.Errorf("Expected %d, got %d for scenario %q", tc.ExpectedDelay, rh.ParseRetryDelay(), name)
			}
		})
	}
}

func TestExpiresInMinutes(t *testing.T) {
	var testCases = map[string]struct {
		ExpiresHeader string
		Expected      time.Duration
	}{
		"Empty header": {
			ExpiresHeader: "",
			Expected:      0,
		},
		"Valid Expires header": {
			ExpiresHeader: time.Now().Add(10 * time.Minute).Format(time.RFC1123),
			Expected:      10 * time.Minute,
		},
		"Invalid Expires header": {
			ExpiresHeader: "invalid-date",
			Expected:      0,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			header := http.Header{}
			header.Add("Expires", tc.ExpiresHeader)
			rh := ResponseHandler{
				httpResponse: &http.Response{
					Header: header,
				},
			}
			if tc.Expected != rh.Expires() {
				t.Errorf("Expected %d, got %d for scenario %q", tc.Expected, rh.Expires(), name)
			}
		})
	}
}

func TestCacheControlMaxAgeInMinutes(t *testing.T) {
	var testCases = map[string]struct {
		CacheControlHeader string
		Expected           time.Duration
	}{
		"Empty header": {
			CacheControlHeader: "",
			Expected:           0,
		},
		"Valid max-age": {
			CacheControlHeader: "max-age=600",
			Expected:           10 * time.Minute,
		},
		"Invalid max-age": {
			CacheControlHeader: "max-age=invalid",
			Expected:           0,
		},
		"Multiple directives": {
			CacheControlHeader: "no-cache, max-age=300",
			Expected:           5 * time.Minute,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			header := http.Header{}
			header.Add("Cache-Control", tc.CacheControlHeader)
			rh := ResponseHandler{
				httpResponse: &http.Response{
					Header: header,
				},
			}
			if tc.Expected != rh.CacheControlMaxAge() {
				t.Errorf("Expected %d, got %d for scenario %q", tc.Expected, rh.CacheControlMaxAge(), name)
			}
		})
	}
}

func TestIsCloudflareChallenge(t *testing.T) {
	makeResp := func(status int, headers map[string]string) *http.Response {
		h := http.Header{}
		for k, v := range headers {
			h.Set(k, v)
		}
		return &http.Response{StatusCode: status, Header: h}
	}

	cases := map[string]struct {
		response *http.Response
		expected bool
	}{
		"403 with cf-mitigated challenge and html": {
			response: makeResp(http.StatusForbidden, map[string]string{
				"Cf-Mitigated": "challenge",
				"Content-Type": "text/html; charset=UTF-8",
			}),
			expected: true,
		},
		"cf-mitigated challenge header on 200": {
			response: makeResp(http.StatusOK, map[string]string{
				"Cf-Mitigated": "challenge",
				"Content-Type": "text/html",
			}),
			expected: false,
		},
		"403 cf-mitigated challenge without html": {
			response: makeResp(http.StatusForbidden, map[string]string{
				"Cf-Mitigated": "challenge",
				"Content-Type": "application/json",
			}),
			expected: false,
		},
		"403 from cloudflare with html but no challenge signal": {
			response: makeResp(http.StatusForbidden, map[string]string{
				"Server":       "cloudflare",
				"Cf-Ray":       "8abc123def456-IAD",
				"Content-Type": "text/html; charset=UTF-8",
			}),
			expected: false,
		},
		"503 from cloudflare with html but no challenge signal": {
			response: makeResp(http.StatusServiceUnavailable, map[string]string{
				"Server":       "cloudflare",
				"Cf-Ray":       "8abc123def456-IAD",
				"Content-Type": "text/html",
			}),
			expected: false,
		},
		"403 from non-cloudflare server": {
			response: makeResp(http.StatusForbidden, map[string]string{
				"Server":       "nginx",
				"Content-Type": "text/html",
			}),
			expected: false,
		},
		"500 from cloudflare with html": {
			response: makeResp(http.StatusInternalServerError, map[string]string{
				"Server":       "cloudflare",
				"Cf-Ray":       "8abc123def456-IAD",
				"Content-Type": "text/html",
			}),
			expected: false,
		},
		"200 OK from cloudflare": {
			response: makeResp(http.StatusOK, map[string]string{
				"Server":       "cloudflare",
				"Cf-Ray":       "8abc123def456-IAD",
				"Content-Type": "application/rss+xml",
			}),
			expected: false,
		},
		"nil response": {
			response: nil,
			expected: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rh := &ResponseHandler{httpResponse: tc.response}
			if got := rh.isCloudflareChallenge(); got != tc.expected {
				t.Errorf("isCloudflareChallenge() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestResponseHandlerCloseClosesBodyOnClientError(t *testing.T) {
	body := &testReadCloser{}
	rh := ResponseHandler{
		httpResponse: &http.Response{Body: body},
		clientErr:    errors.New("boom"),
	}

	rh.Close()

	if !body.closed {
		t.Error("Expected response body to be closed")
	}
}

func TestGetReaderWithZstdContentEncoding(t *testing.T) {
	original := []byte("The quick brown fox jumps over the lazy dog")

	var compressed bytes.Buffer
	encoder, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatalf("Failed to create zstd encoder: %v", err)
	}
	if _, err := encoder.Write(original); err != nil {
		t.Fatalf("Failed to write zstd data: %v", err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatalf("Failed to close zstd encoder: %v", err)
	}

	header := http.Header{}
	header.Set("Content-Encoding", "zstd")
	rh := ResponseHandler{
		httpResponse: &http.Response{
			Header: header,
			Body:   io.NopCloser(bytes.NewReader(compressed.Bytes())),
			Request: &http.Request{
				URL: &url.URL{Scheme: "https", Host: "example.com", Path: "/feed"},
			},
		},
	}

	reader := rh.getReader(1024 * 1024)
	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read from getReader: %v", err)
	}

	if !bytes.Equal(result, original) {
		t.Errorf("Expected %q, got %q", original, result)
	}
}
