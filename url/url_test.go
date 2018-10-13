// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package url // import "miniflux.app/url"

import "testing"

func TestAbsoluteURL(t *testing.T) {
	scenarios := [][]string{
		[]string{"https://example.org/path/file.ext", "https://example.org/folder/", "/path/file.ext"},
		[]string{"https://example.org/folder/path/file.ext", "https://example.org/folder/", "path/file.ext"},
		[]string{"https://example.org/path/file.ext", "https://example.org/folder", "path/file.ext"},
		[]string{"https://example.org/path/file.ext", "https://example.org/folder/", "https://example.org/path/file.ext"},
		[]string{"https://static.example.org/path/file.ext", "https://www.example.org/", "//static.example.org/path/file.ext"},
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
