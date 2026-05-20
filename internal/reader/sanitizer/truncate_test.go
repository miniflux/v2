// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import (
	"os"
	"strconv"
	"testing"
)

func TestTruncateHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "text lower than limit",
			input:    "This is a <strong>bug 🐛</strong>.",
			maxLen:   50,
			expected: "This is a bug 🐛.",
		},
		{
			name:     "text above limit",
			input:    "This is <strong>HTML</strong>.",
			maxLen:   4,
			expected: "This…",
		},
		{
			name:     "unicode text above limit",
			input:    "This is a <strong>bike 🚲</strong>.",
			maxLen:   4,
			expected: "This…",
		},
		{
			name:     "multiline text above limit",
			input:    "\n\t\tThis is a <strong>bike\n\t\t🚲</strong>.\n\n\t",
			maxLen:   15,
			expected: "This is a bike…",
		},
		{
			name:     "multiline text lower than limit",
			input:    "\n\t\tThis is a <strong>bike\n 🚲</strong>.\n\n\t",
			maxLen:   20,
			expected: "This is a bike 🚲.",
		},
		{
			name:     "multiple spaces",
			input:    "hello    world   test",
			maxLen:   20,
			expected: "hello world test",
		},
		{
			name:     "tabs and newlines",
			input:    "hello\t\tworld\n\ntest",
			maxLen:   20,
			expected: "hello world test",
		},
		{
			name:     "truncation with unicode",
			input:    "hello world 你好",
			maxLen:   11,
			expected: "hello world…",
		},
		{
			name:     "html stripping",
			input:    "<p>hello    <b>world</b>   test</p>",
			maxLen:   20,
			expected: "hello world test",
		},
		{
			name:     "no truncation needed",
			input:    "hello world",
			maxLen:   20,
			expected: "hello world",
		},
		{
			name:     "just enough characters",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "just enough unicode characters",
			input:    "Привет",
			maxLen:   6,
			expected: "Привет",
		},
		{
			name:     "spaces around tag",
			input:    "hello <br/> world",
			maxLen:   20,
			expected: "hello world",
		},
		{
			name:     "leading spaces",
			input:    "  hello world",
			maxLen:   5,
			expected: "hello…",
		},
		{
			name:     "text above limit with space at the end",
			input:    "hello world",
			maxLen:   6,
			expected: "hello…",
		},
		{
			name:     "leading space before tag",
			input:    " <a>hello</a>",
			maxLen:   15,
			expected: "hello",
		},
		{
			name:     "space-only tokens in between tags",
			input:    "hello <br/>\t<a> </a>world",
			maxLen:   15,
			expected: "hello world",
		},
		{
			name:     "truncate mid-word",
			input:    "hello world",
			maxLen:   8,
			expected: "hello wo…",
		},
		{
			name:     "truncate mid-word with unicode",
			input:    "Съешь ещё этих мягких французских булок, да выпей же чаю",
			maxLen:   25,
			expected: "Съешь ещё этих мягких фра…",
		},
		{
			name:     "negative limit",
			input:    "whatever",
			maxLen:   -10,
			expected: "…",
		},
		{
			name:     "zero limit",
			input:    "whatever",
			maxLen:   0,
			expected: "…",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateHTML(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateHTML(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func BenchmarkTruncateHTML(b *testing.B) {
	benches := []struct {
		filename string
		limit    int
	}{
		{
			filename: "miniflux_github.html",
			limit:    100,
		},
		{
			filename: "miniflux_github.html",
			limit:    10_000,
		},
		{
			filename: "miniflux_wikipedia.html",
			limit:    100,
		},
		{
			filename: "miniflux_wikipedia.html",
			limit:    100_000,
		},
	}

	for _, f := range benches {
		data, err := os.ReadFile("testdata/" + f.filename)
		if err != nil {
			b.Fatalf(`Unable to read file %q: %v`, f.filename, err)
		}

		b.Run(f.filename+"_"+strconv.Itoa(f.limit), func(b *testing.B) {
			var junk string

			str := string(data)
			for b.Loop() {
				junk = TruncateHTML(str, 100)
			}

			_ = junk
		})
	}
}
