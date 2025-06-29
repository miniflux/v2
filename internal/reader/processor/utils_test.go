// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
)

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
