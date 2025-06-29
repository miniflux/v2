// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/v2/internal/reader/readability"

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

func TestNestedSpanInCodeBlock(t *testing.T) {
	html := `
		<html>
			<head>
				<title>Test</title>
			</head>
			<body>
				<article><p>Some content</p><pre><code class="hljs-built_in">Code block with <span class="hljs-built_in">nested span</span> <span class="hljs-comment"># exit 1</span></code></pre></article>
			</body>
		</html>`
	want := `<div><div><p>Some content</p><pre><code class="hljs-built_in">Code block with <span class="hljs-built_in">nested span</span> <span class="hljs-comment"># exit 1</span></code></pre></div></div>`

	_, result, err := ExtractContent(strings.NewReader(html))
	if err != nil {
		t.Fatal(err)
	}

	if result != want {
		t.Errorf(`Invalid content, got %s instead of %s`, result, want)
	}
}

func BenchmarkExtractContent(b *testing.B) {
	var testCases = map[string][]byte{
		"miniflux_github.html":    {},
		"miniflux_wikipedia.html": {},
	}
	for filename := range testCases {
		data, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			b.Fatalf(`Unable to read file %q: %v`, filename, err)
		}
		testCases[filename] = data
	}
	for range b.N {
		for _, v := range testCases {
			ExtractContent(bytes.NewReader(v))
		}
	}
}

func TestGetClassWeight(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected float32
	}{
		{
			name:     "no class or id",
			html:     `<div>content</div>`,
			expected: 0,
		},
		{
			name:     "positive class only",
			html:     `<div class="article">content</div>`,
			expected: 25,
		},
		{
			name:     "negative class only",
			html:     `<div class="comment">content</div>`,
			expected: -25,
		},
		{
			name:     "positive id only",
			html:     `<div id="main">content</div>`,
			expected: 25,
		},
		{
			name:     "negative id only",
			html:     `<div id="sidebar">content</div>`,
			expected: -25,
		},
		{
			name:     "positive class and positive id",
			html:     `<div class="content" id="main">content</div>`,
			expected: 50,
		},
		{
			name:     "negative class and negative id",
			html:     `<div class="comment" id="sidebar">content</div>`,
			expected: -50,
		},
		{
			name:     "positive class and negative id",
			html:     `<div class="article" id="comment">content</div>`,
			expected: 0,
		},
		{
			name:     "negative class and positive id",
			html:     `<div class="banner" id="content">content</div>`,
			expected: 0,
		},
		{
			name:     "multiple positive classes",
			html:     `<div class="article content">content</div>`,
			expected: 25,
		},
		{
			name:     "multiple negative classes",
			html:     `<div class="comment sidebar">content</div>`,
			expected: -25,
		},
		{
			name:     "mixed positive and negative classes",
			html:     `<div class="article comment">content</div>`,
			expected: -25, // negative takes precedence since it's checked first
		},
		{
			name:     "case insensitive class",
			html:     `<div class="ARTICLE">content</div>`,
			expected: 25,
		},
		{
			name:     "case insensitive id",
			html:     `<div id="MAIN">content</div>`,
			expected: 25,
		},
		{
			name:     "non-matching class and id",
			html:     `<div class="random" id="unknown">content</div>`,
			expected: 0,
		},
		{
			name:     "empty class and id",
			html:     `<div class="" id="">content</div>`,
			expected: 0,
		},
		{
			name:     "class with special characters",
			html:     `<div class="com-section">content</div>`,
			expected: -25, // matches com- in negative regex
		},
		{
			name:     "id with special characters",
			html:     `<div id="h-entry-123">content</div>`,
			expected: 25, // matches h-entry in positive regex
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			selection := doc.Find("div").First()
			if selection.Length() == 0 {
				t.Fatal("No div element found in HTML")
			}

			result := getClassWeight(selection)
			if result != tc.expected {
				t.Errorf("Expected weight %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestGetClassWeightRegexPatterns(t *testing.T) {
	// Test specific regex patterns used in getClassWeight
	positiveWords := []string{"article", "body", "content", "entry", "hentry", "h-entry", "main", "page", "pagination", "post", "text", "blog", "story"}
	negativeWords := []string{"hid", "banner", "combx", "comment", "com-", "contact", "foot", "masthead", "media", "meta", "modal", "outbrain", "promo", "related", "scroll", "share", "shoutbox", "sidebar", "skyscraper", "sponsor", "shopping", "tags", "tool", "widget", "byline", "author", "dateline", "writtenby"}

	for _, word := range positiveWords {
		t.Run("positive_"+word, func(t *testing.T) {
			html := `<div class="` + word + `">content</div>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			selection := doc.Find("div").First()
			result := getClassWeight(selection)
			if result != 25 {
				t.Errorf("Expected positive weight 25 for word '%s', got %f", word, result)
			}
		})
	}

	for _, word := range negativeWords {
		t.Run("negative_"+word, func(t *testing.T) {
			html := `<div class="` + word + `">content</div>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			selection := doc.Find("div").First()
			result := getClassWeight(selection)
			if result != -25 {
				t.Errorf("Expected negative weight -25 for word '%s', got %f", word, result)
			}
		})
	}
}
