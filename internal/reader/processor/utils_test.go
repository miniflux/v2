// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
	"time"
)

func TestISO8601DurationParsing(t *testing.T) {
	var scenarios = []struct {
		duration string
		expected time.Duration
	}{
		// Live streams and radio.
		{"PT0M0S", 0},
		// https://www.youtube.com/watch?v=HLrqNhgdiC0
		{"PT6M20S", (6 * time.Minute) + (20 * time.Second)},
		// https://www.youtube.com/watch?v=LZa5KKfqHtA
		{"PT5M41S", (5 * time.Minute) + (41 * time.Second)},
		// https://www.youtube.com/watch?v=yIxEEgEuhT4
		{"PT51M52S", (51 * time.Minute) + (52 * time.Second)},
		// https://www.youtube.com/watch?v=bpHf1XcoiFs
		{"PT80M42S", (1 * time.Hour) + (20 * time.Minute) + (42 * time.Second)},
		// Hours only
		{"PT2H", 2 * time.Hour},
		// Seconds only
		{"PT30S", 30 * time.Second},
		// Hours and minutes
		{"PT1H30M", (1 * time.Hour) + (30 * time.Minute)},
		// Hours and seconds
		{"PT2H45S", (2 * time.Hour) + (45 * time.Second)},
		// Empty duration
		{"PT", 0},
	}

	for _, tc := range scenarios {
		result, err := parseISO8601Duration(tc.duration)
		if err != nil {
			t.Errorf("Got an error when parsing %q: %v", tc.duration, err)
		}

		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for duration %q`, result, tc.duration)
		}
	}
}

func TestISO8601DurationParsingErrors(t *testing.T) {
	var errorScenarios = []struct {
		duration    string
		expectedErr string
	}{
		// Missing PT prefix
		{"6M20S", "the period doesn't start with PT"},
		// Unsupported Year specifier
		{"PT1Y", "the 'Y' specifier isn't supported"},
		// Unsupported Week specifier
		{"PT2W", "the 'W' specifier isn't supported"},
		// Unsupported Day specifier
		{"PT3D", "the 'D' specifier isn't supported"},
		// Invalid number for hours (letter at start of number)
		{"PTaH", "invalid character in the period"},
		// Invalid number for minutes (letter at start of number)
		{"PTbM", "invalid character in the period"},
		// Invalid number for seconds (letter at start of number)
		{"PTcS", "invalid character in the period"},
		// Invalid character in the middle of a number
		{"PT1a2H", "invalid character in the period"},
		{"PT3b4M", "invalid character in the period"},
		{"PT5c6S", "invalid character in the period"},
		// Test cases for actual ParseFloat errors (empty number before specifier)
		{"PTH", "strconv.ParseFloat: parsing \"\": invalid syntax"},
		{"PTM", "strconv.ParseFloat: parsing \"\": invalid syntax"},
		{"PTS", "strconv.ParseFloat: parsing \"\": invalid syntax"},
		// Invalid character
		{"PT1X", "invalid character in the period"},
		// Invalid character mixed
		{"PT1H@M", "invalid character in the period"},
	}

	for _, tc := range errorScenarios {
		_, err := parseISO8601Duration(tc.duration)
		if err == nil {
			t.Errorf("Expected an error when parsing %q, but got none", tc.duration)
		} else if err.Error() != tc.expectedErr {
			t.Errorf("Expected error %q when parsing %q, but got %q", tc.expectedErr, tc.duration, err.Error())
		}
	}
}

func TestMinifyEntryContentWithWhitespace(t *testing.T) {
	input := `<p>    Some text with a <a href="http://example.org/"> link   </a>    </p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a></p>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}

func TestMinifyContentWithDefaultAttributes(t *testing.T) {
	input := `<script type="application/javascript">console.log("Hello, World!");</script>`
	expected := `<script>console.log("Hello, World!");</script>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}

func TestMinifyContentWithComments(t *testing.T) {
	input := `<p>Some text<!-- This is a comment --> with a <a href="http://example.org/">link</a>.</p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a>.</p>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}

func TestMinifyContentWithSpecialComments(t *testing.T) {
	input := `<p>Some text <!--[if IE 6]><p>IE6</p><![endif]--> with a <a href="http://example.org/">link</a>.</p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a>.</p>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}
