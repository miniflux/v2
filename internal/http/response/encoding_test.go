// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"testing"
)

func TestAcceptEncoding(t *testing.T) {
	t.Parallel()

	acceptable := []string{
		"br", "gzip", "deflate",
	}

	tests := []struct {
		name           string
		acceptEncoding string
		want           string
	}{
		{
			name:           "Empty input",
			acceptEncoding: "",
			want:           "identity",
		},
		{
			name:           "q=0 and identity",
			acceptEncoding: "identity;q=0",
			want:           "",
		},
		{
			name:           "q=0 and *",
			acceptEncoding: "*;q=0",
			want:           "",
		},
		{
			name:           "gzip",
			acceptEncoding: "gzip",
			want:           "gzip",
		},
		{
			name:           "gzip and br",
			acceptEncoding: "gzip,br",
			want:           "gzip",
		},
		{
			name:           "br and gzip",
			acceptEncoding: "br,gzip,deflate",
			want:           "br",
		},
		{
			name:           "unsupported encoding",
			acceptEncoding: "unknown",
			want:           "identity",
		},
		{
			name:           "empty encoding",
			acceptEncoding: ",",
			want:           "identity",
		},
		{
			name:           "multiple encodings and q=0",
			acceptEncoding: "gzip;q=0,br;q=0",
			want:           "identity",
		},
		{
			// We want br here but weights are not supported.
			name:           "multiple encodings and q values",
			acceptEncoding: "gzip;q=0.5,br;q=0.8",
			want:           "gzip",
		},
		{
			name:           "multiple encodings and wildcard",
			acceptEncoding: "*;q=0,gzip,br",
			want:           "gzip",
		},
		{
			name:           "multiple encodings and wildcard and q=0",
			acceptEncoding: "*;q=0,gzip,br;q=0",
			want:           "gzip",
		},
		{
			// We want br here but weights are not supported.
			name:           "multiple encodings and wildcard and q values",
			acceptEncoding: "*;q=0.5,gzip;q=0.8,br",
			want:           "gzip",
		},
		{
			name:           "multiple encodings and wildcard and q values and q=0",
			acceptEncoding: "*;q=0.5,gzip;q=0.8,br;q=0",
			want:           "gzip",
		},
		{
			name:           "invalid q value",
			acceptEncoding: "gzip;q=abc,deflate",
			want:           "deflate",
		},
		{
			name:           "wrong spaces placing around q value",
			acceptEncoding: "gzip;q= 0.5, deflate;q=0.8",
			want:           "deflate",
		},
		{
			name:           "correct spaces placing around q value",
			acceptEncoding: "gzip ; q=0.5, deflate;q=0.8",
			want:           "gzip",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Instantiate parser for each test to make sure it doesn't return cached values.
			parser := AcceptEncoding(acceptable...)

			got := parser.Parse(test.acceptEncoding)
			if got != test.want {
				t.Errorf("AcceptEncoding() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestAcceptEncodingCache(t *testing.T) {
	encoding := "identity;q=0,gzip,whatever"
	expected := "gzip"

	parser := AcceptEncoding("br", "gzip", "deflate")

	_, cached := parser.cache.Load(encoding)
	if cached {
		t.Errorf("empty cache can't have %q", encoding)
	}

	got := parser.Parse(encoding)
	if got != expected {
		t.Errorf("AcceptEncoding() = %q, want %q", got, expected)
	}

	cgot, cached := parser.cache.Load(encoding)
	if !cached {
		t.Errorf("parser cache doesn't have %q", encoding)
	}

	got = parser.Parse(encoding)
	if got != expected {
		t.Errorf("AcceptEncoding() = %q, want %q", got, expected)
	}

	if cgot.(string) != expected {
		t.Errorf("cached = %q, want %q", cgot, expected)
	}
}
