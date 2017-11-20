// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sanitizer

import "testing"

func TestStripTags(t *testing.T) {
	input := `This <a href="/test.html">link is relative</a> and <strong>this</strong> image: <img src="../folder/image.png"/>`
	expected := `This link is relative and this image: `
	output := StripTags(input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}
