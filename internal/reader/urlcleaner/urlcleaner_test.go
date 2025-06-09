// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package urlcleaner // import "miniflux.app/v2/internal/reader/urlcleaner"

import (
	"net/url"
	"reflect"
	"testing"
)

func TestRemoveTrackingParams(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expected         string
		baseUrl          string
		feedUrl          string
		strictComparison bool
	}{
		{
			name:     "URL with tracking parameters",
			input:    "https://example.com/page?id=123&utm_source=newsletter&utm_medium=email&fbclid=abc123",
			expected: "https://example.com/page?id=123",
		},
		{
			name:     "URL with only tracking parameters",
			input:    "https://example.com/page?utm_source=newsletter&utm_medium=email",
			expected: "https://example.com/page",
		},
		{
			name:     "URL with no tracking parameters",
			input:    "https://example.com/page?id=123&foo=bar",
			expected: "https://example.com/page?id=123&foo=bar",
		},
		{
			name:             "URL with no parameters",
			input:            "https://example.com/page",
			expected:         "https://example.com/page",
			strictComparison: true,
		},
		{
			name:     "URL with mixed case tracking parameters",
			input:    "https://example.com/page?UTM_SOURCE=newsletter&utm_MEDIUM=email",
			expected: "https://example.com/page",
		},
		{
			name:     "URL with tracking parameters and fragments",
			input:    "https://example.com/page?id=123&utm_source=newsletter#section1",
			expected: "https://example.com/page?id=123#section1",
		},
		{
			name:     "URL with only tracking parameters and fragments",
			input:    "https://example.com/page?utm_source=newsletter#section1",
			expected: "https://example.com/page#section1",
		},
		{
			name:     "URL with only one tracking parameter",
			input:    "https://example.com/page?utm_source=newsletter",
			expected: "https://example.com/page",
		},
		{
			name:     "URL with encoded characters",
			input:    "https://example.com/page?name=John%20Doe&utm_source=newsletter",
			expected: "https://example.com/page?name=John+Doe",
		},
		{
			name:     "ref parameter for another url",
			input:    "https://example.com/page?ref=test.com",
			baseUrl:  "https://example.com/page",
			expected: "https://example.com/page?ref=test.com",
		},
		{
			name:     "ref parameter for feed url",
			input:    "https://example.com/page?ref=feed.com",
			baseUrl:  "https://example.com/page",
			expected: "https://example.com/page",
			feedUrl:  "http://feed.com",
		},
		{
			name:     "ref parameter for site url",
			input:    "https://example.com/page?ref=example.com",
			baseUrl:  "https://example.com/page",
			expected: "https://example.com/page",
		},
		{
			name:     "ref parameter for base url",
			input:    "https://example.com/page?ref=example.com",
			expected: "https://example.com/page",
			baseUrl:  "https://example.com",
			feedUrl:  "https://feedburned.com/example",
		},
		{
			name:     "ref parameter for base url on subdomain",
			input:    "https://blog.exploits.club/some-path?ref=blog.exploits.club",
			expected: "https://blog.exploits.club/some-path",
			baseUrl:  "https://blog.exploits.club/some-path",
			feedUrl:  "https://feedburned.com/exploit.club",
		},
		{
			name:             "Non-standard URL parameter with no tracker",
			input:            "https://example.com/foo.jpg?crop/1420x708/format/webp",
			expected:         "https://example.com/foo.jpg?crop/1420x708/format/webp",
			baseUrl:          "https://example.com/page",
			strictComparison: true,
		},
		{
			name:     "Invalid URL",
			input:    "https://example|org/",
			baseUrl:  "https://example.com/page",
			expected: "",
		},
		{
			name:             "Non-HTTP URL",
			input:            "mailto:user@example.org",
			expected:         "mailto:user@example.org",
			baseUrl:          "https://example.com/page",
			strictComparison: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedBaseUrl, _ := url.Parse(tt.baseUrl)
			parsedFeedUrl, _ := url.Parse(tt.feedUrl)
			parsedInputUrl, _ := url.Parse(tt.input)
			result, err := RemoveTrackingParameters(parsedBaseUrl, parsedFeedUrl, parsedInputUrl)
			if tt.expected == "" {
				if err == nil {
					t.Errorf("Expected an error for invalid URL, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.strictComparison && result != tt.expected {
					t.Errorf("removeTrackingParams(%q) = %q, want %q", tt.input, result, tt.expected)
				}
				if !urlsEqual(result, tt.expected) {
					t.Errorf("removeTrackingParams(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// urlsEqual compares two URLs for equality, ignoring the order of query parameters
func urlsEqual(url1, url2 string) bool {
	u1, err1 := url.Parse(url1)
	u2, err2 := url.Parse(url2)

	if err1 != nil || err2 != nil {
		return false
	}

	if u1.Scheme != u2.Scheme || u1.Host != u2.Host || u1.Path != u2.Path || u1.Fragment != u2.Fragment {
		return false
	}

	return reflect.DeepEqual(u1.Query(), u2.Query())
}
