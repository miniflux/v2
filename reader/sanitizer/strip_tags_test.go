// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/reader/sanitizer"

import "testing"

func TestStripTags(t *testing.T) {
	input := `This <a href="/test.html">link is relative</a> and <strong>this</strong> image: <img src="../folder/image.png"/>`
	expected := `This link is relative and this image: `
	output := StripTags(input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}
