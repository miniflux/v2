// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package url // import "miniflux.app/url"

import "testing"

func TestGetAbsoluteURLWithAbsolutePath(t *testing.T) {
	expected := `https://example.org/path/file.ext`
	input := `/path/file.ext`
	output, err := AbsoluteURL("https://example.org/folder/", input)

	if err != nil {
		t.Error(err)
	}

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestGetAbsoluteURLWithRelativePath(t *testing.T) {
	expected := `https://example.org/folder/path/file.ext`
	input := `path/file.ext`
	output, err := AbsoluteURL("https://example.org/folder/", input)

	if err != nil {
		t.Error(err)
	}

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestGetAbsoluteURLWithRelativePaths(t *testing.T) {
	expected := `https://example.org/path/file.ext`
	input := `path/file.ext`
	output, err := AbsoluteURL("https://example.org/folder", input)

	if err != nil {
		t.Error(err)
	}

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestWhenInputIsAlreadyAbsolute(t *testing.T) {
	expected := `https://example.org/path/file.ext`
	input := `https://example.org/path/file.ext`
	output, err := AbsoluteURL("https://example.org/folder/", input)

	if err != nil {
		t.Error(err)
	}

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestGetAbsoluteURLWithProtocolRelative(t *testing.T) {
	expected := `https://static.example.org/path/file.ext`
	input := `//static.example.org/path/file.ext`
	output, err := AbsoluteURL("https://www.example.org/", input)

	if err != nil {
		t.Error(err)
	}

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestGetRootURL(t *testing.T) {
	expected := `https://example.org/`
	input := `https://example.org/path/file.ext`
	output := RootURL(input)

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestGetRootURLWithProtocolRelativePath(t *testing.T) {
	expected := `https://static.example.org/`
	input := `//static.example.org/path/file.ext`
	output := RootURL(input)

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}

func TestIsHTTPS(t *testing.T) {
	if !IsHTTPS("https://example.org/") {
		t.Error("Unable to recognize HTTPS URL")
	}

	if IsHTTPS("http://example.org/") {
		t.Error("Unable to recognize HTTP URL")
	}

	if IsHTTPS("") {
		t.Error("Unable to recognize malformed URL")
	}
}

func TestGetDomain(t *testing.T) {
	expected := `static.example.org`
	input := `http://static.example.org/`
	output := Domain(input)

	if expected != output {
		t.Errorf(`Unexpected output, got "%s" instead of "%s"`, output, expected)
	}
}
