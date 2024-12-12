// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/v2/internal/reader/readability"

import (
	"strings"
	"testing"
)

func TestBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="https://example.org/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "https://example.org/" {
		t.Errorf(`Unexpected base URL, got %q instead of "https://example.org/"`, baseURL)
	}
}

func TestMultipleBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="https://example.org/ ">
				<base href="https://example.com/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "https://example.org/" {
		t.Errorf(`Unexpected base URL, got %q instead of "https://example.org/"`, baseURL)
	}
}

func TestRelativeBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<base href="/test/ ">
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "" {
		t.Errorf(`Unexpected base URL, got %q`, baseURL)
	}
}

func TestWithoutBaseURL(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article>
					Some content
				</article>
			</body>
		</html>`

	baseURL, _, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if baseURL != "" {
		t.Errorf(`Unexpected base URL, got %q instead of ""`, baseURL)
	}
}

func TestRemoveStyleScript(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
				    <script src="tololo.js"></script>
			</head>
			<body>
				<script src="tololo.js"></script>
				<style>
			  		h1 {color:red;}
			  		p {color:blue;}
				</style>
				<article>Some content</article>
			</body>
		</html>`
	want := `<div><div><article>Somecontent</article></div></div>`

	_, content, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\t", "")

	if content != want {
		t.Errorf(`Invalid content, got %s instead of %s`, content, want)
	}
}

func TestRemoveBlacklist(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article class="super-ad">Some content</article>
				<article class="g-plus-crap">Some other thing</article>
				<article class="stuff popupbody">And more</article>
				<article class="legit">Valid!</article>
			</body>
		</html>`
	want := `<div><div><articleclass="legit">Valid!</article></div></div>`

	_, content, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\t", "")

	if content != want {
		t.Errorf(`Invalid content, got %s instead of %s`, content, want)
	}
}
