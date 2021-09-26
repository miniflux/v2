// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package url // import "miniflux.app/url"

import "testing"

func TestIsAbsoluteURL(t *testing.T) {
	scenarios := map[string]bool{
		"https://example.org/file.pdf": true,
		"magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7": true,
		"invalid url": false,
	}

	for input, expected := range scenarios {
		actual := IsAbsoluteURL(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v for %q`, actual, expected, input)
		}
	}
}

func TestAbsoluteURL(t *testing.T) {
	scenarios := [][]string{
		{"https://example.org/path/file.ext", "https://example.org/folder/", "/path/file.ext"},
		{"https://example.org/folder/path/file.ext", "https://example.org/folder/", "path/file.ext"},
		{"https://example.org/path/file.ext", "https://example.org/folder", "path/file.ext"},
		{"https://example.org/path/file.ext", "https://example.org/folder/", "https://example.org/path/file.ext"},
		{"https://static.example.org/path/file.ext", "https://www.example.org/", "//static.example.org/path/file.ext"},
		{"magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a", "https://www.example.org/", "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a"},
		{"magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7", "https://www.example.org/", "magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7"},
	}

	for _, scenario := range scenarios {
		actual, err := AbsoluteURL(scenario[1], scenario[2])

		if err != nil {
			t.Errorf(`Got error for (%q, %q): %v`, scenario[1], scenario[2], err)
		}

		if actual != scenario[0] {
			t.Errorf(`Unexpected result, got %q instead of %q for (%q, %q)`, actual, scenario[0], scenario[1], scenario[2])
		}
	}
}

func TestRootURL(t *testing.T) {
	scenarios := map[string]string{
		"https://example.org/path/file.ext":  "https://example.org/",
		"//static.example.org/path/file.ext": "https://static.example.org/",
		"https://example|org/path/file.ext":  "https://example|org/path/file.ext",
	}

	for input, expected := range scenarios {
		actual := RootURL(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, actual, expected)
		}
	}
}

func TestIsHTTPS(t *testing.T) {
	scenarios := map[string]bool{
		"https://example.org/": true,
		"http://example.org/":  false,
		"https://example|org/": false,
	}

	for input, expected := range scenarios {
		actual := IsHTTPS(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %v instead of %v`, actual, expected)
		}
	}
}

func TestDomain(t *testing.T) {
	scenarios := map[string]string{
		"https://static.example.org/": "static.example.org",
		"https://example|org/":        "https://example|org/",
	}

	for input, expected := range scenarios {
		actual := Domain(input)
		if actual != expected {
			t.Errorf(`Unexpected result, got %q instead of %q`, actual, expected)
		}
	}
}
