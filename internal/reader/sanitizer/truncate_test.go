// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import "testing"

func TestTruncateHTMWithTextLowerThanLimitL(t *testing.T) {
	input := `This is a <strong>bug 🐛</strong>.`
	expected := `This is a bug 🐛.`
	output := TruncateHTML(input, 50)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestTruncateHTMLWithTextAboveLimit(t *testing.T) {
	input := `This is <strong>HTML</strong>.`
	expected := `This…`
	output := TruncateHTML(input, 4)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestTruncateHTMLWithUnicodeTextAboveLimit(t *testing.T) {
	input := `This is a <strong>bike 🚲</strong>.`
	expected := `This…`
	output := TruncateHTML(input, 4)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestTruncateHTMLWithMultilineTextAboveLimit(t *testing.T) {
	input := `
		This is a <strong>bike
		🚲</strong>.

	`
	expected := `This is a bike…`
	output := TruncateHTML(input, 15)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestTruncateHTMLWithMultilineTextLowerThanLimit(t *testing.T) {
	input := `
		This is a <strong>bike
 🚲</strong>.

	`
	expected := `This is a bike 🚲.`
	output := TruncateHTML(input, 20)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestTruncateHTMLWithMultipleSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
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
