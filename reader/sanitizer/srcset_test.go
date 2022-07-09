// Copyright 2022 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sanitizer

import "testing"

func TestParseSrcSetAttributeWithRelativeURLs(t *testing.T) {
	input := `example-320w.jpg, example-480w.jpg 1.5x,   example-640,w.jpg 2x, example-640w.jpg 640w`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 4 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `example-320w.jpg, example-480w.jpg 1.5x, example-640,w.jpg 2x, example-640w.jpg 640w` {
		t.Errorf(`Unexpected string output`)
	}
}

func TestParseSrcSetAttributeWithAbsoluteURLs(t *testing.T) {
	input := `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 2 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `http://example.org/example-320w.jpg 320w, http://example.org/example-480w.jpg 1.5x` {
		t.Errorf(`Unexpected string output`)
	}
}

func TestParseSrcSetAttributeWithOneCandidate(t *testing.T) {
	input := `http://example.org/example-320w.jpg`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 1 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `http://example.org/example-320w.jpg` {
		t.Errorf(`Unexpected string output`)
	}
}

func TestParseSrcSetAttributeWithCommaURL(t *testing.T) {
	input := `http://example.org/example,a:b/d.jpg , example-480w.jpg 1.5x`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 2 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `http://example.org/example,a:b/d.jpg, example-480w.jpg 1.5x` {
		t.Errorf(`Unexpected string output`)
	}
}

func TestParseSrcSetAttributeWithIncorrectDescriptor(t *testing.T) {
	input := `http://example.org/example-320w.jpg test`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 0 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `` {
		t.Errorf(`Unexpected string output`)
	}
}

func TestParseSrcSetAttributeWithTooManyDescriptors(t *testing.T) {
	input := `http://example.org/example-320w.jpg 10w 1x`
	candidates := ParseSrcSetAttribute(input)

	if len(candidates) != 0 {
		t.Error(`Incorrect number of candidates`)
	}

	if candidates.String() != `` {
		t.Errorf(`Unexpected string output`)
	}
}
