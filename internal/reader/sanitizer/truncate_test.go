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
