// Copyright 2022 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
