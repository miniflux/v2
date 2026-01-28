// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer

import "testing"

func assertCandidates(t *testing.T, input string, expectedCount int, expectedString string) {
	t.Helper()

	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != expectedCount {
		t.Fatalf("Incorrect number of candidates for input %q: got %d, want %d", input, len(candidates), expectedCount)
	}

	if output := candidates.String(); output != expectedString {
		t.Fatalf("Unexpected string output for input %q. Got %q, want %q", input, output, expectedString)
	}
}

func TestParseSrcSetAttributeValidCandidates(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedCount  int
		expectedString string
	}{
		{
			name:           "relative urls",
			input:          `example-320w.jpg, example-480w.jpg 1.5x,   example-640,w.jpg 2x, example-640w.jpg 640w`,
			expectedCount:  4,
			expectedString: `example-320w.jpg, example-480w.jpg 1.5x, example-640,w.jpg 2x, example-640w.jpg 640w`,
		},
		{
			name:           "relative urls no space after comma",
			input:          `a.png 1x,b.png 2x`,
			expectedCount:  2,
			expectedString: `a.png 1x, b.png 2x`,
		},
		{
			name:           "relative urls extra spaces",
			input:          `  a.png   1x  ,   b.png    2x   `,
			expectedCount:  2,
			expectedString: `a.png 1x, b.png 2x`,
		},
		{
			name:           "absolute urls",
			input:          `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x`,
			expectedCount:  2,
			expectedString: `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x`,
		},
		{
			name:           "absolute urls no space after comma",
			input:          `http://example.org/example-320w.jpg 320w,http://example.org/example-480w.jpg 1.5x`,
			expectedCount:  2,
			expectedString: `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x`,
		},
		{
			name:           "absolute urls trailing comma",
			input:          `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x, `,
			expectedCount:  2,
			expectedString: `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x`,
		},
		{
			name:           "one candidate",
			input:          `http://example.org/example-320w.jpg`,
			expectedCount:  1,
			expectedString: `http://example.org/example-320w.jpg`,
		},
		{
			name:           "comma inside url",
			input:          `http://example.org/example,a:b/d.jpg , example-480w.jpg 1.5x`,
			expectedCount:  2,
			expectedString: `http://example.org/example,a:b/d.jpg, example-480w.jpg 1.5x`,
		},
		{
			name:           "width and height descriptor",
			input:          `http://example.org/example-320w.jpg 320w 240h`,
			expectedCount:  1,
			expectedString: `http://example.org/example-320w.jpg 320w`,
		},
		{
			name:           "data url",
			input:          `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA 1x, image@2x.png 2x`,
			expectedCount:  2,
			expectedString: `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUA 1x, image@2x.png 2x`,
		},
		{
			name:           "url with parentheses",
			input:          `http://example.org/example(1).jpg 1x`,
			expectedCount:  1,
			expectedString: `http://example.org/example(1).jpg 1x`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCandidates(t, testCase.input, testCase.expectedCount, testCase.expectedString)
		})
	}
}

func TestParseSrcSetAttributeInvalidCandidates(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input",
			input: ``,
		},
		{
			name:  "incorrect descriptor",
			input: `http://example.org/example-320w.jpg test`,
		},
		{
			name:  "too many descriptors",
			input: `http://example.org/example-320w.jpg 10w 1x`,
		},
		{
			name:  "height descriptor only",
			input: `http://example.org/example-320w.jpg 20h`,
		},
		{
			name:  "invalid density descriptor +",
			input: `http://example.org/example-320w.jpg +1x`,
		},
		{
			name:  "invalid density descriptor dot",
			input: `http://example.org/example-320w.jpg 1.x`,
		},
		{
			name:  "invalid density descriptor -",
			input: `http://example.org/example-320w.jpg -1x`,
		},
		{
			name:  "invalid width descriptor zero",
			input: `http://example.org/example-320w.jpg 0w`,
		},
		{
			name:  "invalid width descriptor negative",
			input: `http://example.org/example-320w.jpg -10w`,
		},
		{
			name:  "invalid width descriptor float",
			input: `http://example.org/example-320w.jpg 10.5w`,
		},
		{
			name:  "descriptor with parentheses",
			input: `http://example.org/example-320w.jpg (1x)`,
		},
		{
			name:  "descriptor with parentheses and comma",
			input: `http://example.org/example-320w.jpg calc(1x, 2x)`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assertCandidates(t, testCase.input, 0, "")
		})
	}
}
