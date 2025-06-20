// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"testing"
)

func TestMinifyEntryContent(t *testing.T) {
	input := `<p>    Some text with a <a href="http://example.org/"> link   </a>    </p>`
	expected := `<p>Some text with a <a href="http://example.org/">link</a></p>`
	result := minifyContent(input)
	if expected != result {
		t.Errorf(`Unexpected result, got %q`, result)
	}
}
