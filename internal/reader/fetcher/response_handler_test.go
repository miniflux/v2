// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"net/http"
	"testing"
	"time"
)

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
		ExpectedDelay    int
	}{
		"Empty header": {
			RetryAfterHeader: "",
			ExpectedDelay:    0,
		},
		"Integer value": {
			RetryAfterHeader: "42",
			ExpectedDelay:    42,
		},
		"HTTP-date": {
			RetryAfterHeader: time.Now().Add(42 * time.Second).Format(time.RFC1123),
			ExpectedDelay:    41,
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
