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

func TestRemoveUnlikelyCandidates(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "removes elements with popupbody class",
			html:     `<html><body><div class="popupbody">popup content</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "removes elements with -ad in class",
			html:     `<html><body><div class="super-ad">ad content</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "removes elements with g-plus in class",
			html:     `<html><body><div class="g-plus-share">social content</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "removes elements with unlikely candidates in class",
			html:     `<html><body><div class="banner">banner</div><div class="sidebar">sidebar</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "preserves elements with unlikely candidates but also good candidates in class",
			html:     `<html><body><div class="banner article">mixed content</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="banner article">mixed content</div><div class="content">good content</div></body></html>`,
		},
		{
			name:     "removes elements with unlikely candidates in id",
			html:     `<html><body><div id="banner">banner</div><div id="main-content">good content</div></body></html>`,
			expected: `<html><head></head><body><div id="main-content">good content</div></body></html>`,
		},
		{
			name:     "preserves elements with unlikely candidates but also good candidates in id",
			html:     `<html><body><div id="comment-article">mixed content</div><div id="main">good content</div></body></html>`,
			expected: `<html><head></head><body><div id="comment-article">mixed content</div><div id="main">good content</div></body></html>`,
		},
		{
			name:     "preserves html and body tags",
			html:     `<html class="banner"><body class="sidebar"><div class="banner">content</div></body></html>`,
			expected: `<html class="banner"><head></head><body class="sidebar"></body></html>`,
		},
		{
			name:     "preserves elements within code blocks",
			html:     `<html><body><pre><code><span class="banner">code content</span></code></pre><div class="banner">remove this</div></body></html>`,
			expected: `<html><head></head><body><pre><code><span class="banner">code content</span></code></pre></body></html>`,
		},
		{
			name:     "preserves elements within pre tags",
			html:     `<html><body><pre><div class="sidebar">preformatted content</div></pre><div class="sidebar">remove this</div></body></html>`,
			expected: `<html><head></head><body><pre><div class="sidebar">preformatted content</div></pre></body></html>`,
		},
		{
			name:     "case insensitive matching",
			html:     `<html><body><div class="BANNER">uppercase banner</div><div class="Banner">mixed case banner</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "multiple unlikely patterns in single class",
			html:     `<html><body><div class="banner sidebar footer">multiple bad</div><div class="content">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="content">good content</div></body></html>`,
		},
		{
			name:     "elements without class or id are preserved",
			html:     `<html><body><div>no attributes</div><p>paragraph</p></body></html>`,
			expected: `<html><head></head><body><div>no attributes</div><p>paragraph</p></body></html>`,
		},
		{
			name:     "removes nested unlikely elements",
			html:     `<html><body><div class="main"><div class="banner">nested banner</div><p>good content</p></div></body></html>`,
			expected: `<html><head></head><body><div class="main"><p>good content</p></div></body></html>`,
		},
		{
			name:     "comprehensive unlikely candidates test",
			html:     `<html><body><div class="breadcrumbs">breadcrumbs</div><div class="combx">combx</div><div class="comment">comment</div><div class="community">community</div><div class="cover-wrap">cover-wrap</div><div class="disqus">disqus</div><div class="extra">extra</div><div class="foot">foot</div><div class="header">header</div><div class="legends">legends</div><div class="menu">menu</div><div class="modal">modal</div><div class="related">related</div><div class="remark">remark</div><div class="replies">replies</div><div class="rss">rss</div><div class="shoutbox">shoutbox</div><div class="skyscraper">skyscraper</div><div class="social">social</div><div class="sponsor">sponsor</div><div class="supplemental">supplemental</div><div class="ad-break">ad-break</div><div class="agegate">agegate</div><div class="pagination">pagination</div><div class="pager">pager</div><div class="popup">popup</div><div class="yom-remote">yom-remote</div><div class="article">good content</div></body></html>`,
			expected: `<html><head></head><body><div class="article">good content</div></body></html>`,
		},
		{
			name:     "preserves good candidates that contain unlikely words",
			html:     `<html><body><div class="banner article">should be preserved</div><div class="comment main">should be preserved</div><div class="sidebar body">should be preserved</div><div class="footer column">should be preserved</div><div class="header and">should be preserved</div><div class="menu shadow">should be preserved</div><div class="pure-banner">should be removed</div></body></html>`,
			expected: `<html><head></head><body><div class="banner article">should be preserved</div><div class="comment main">should be preserved</div><div class="sidebar body">should be preserved</div><div class="footer column">should be preserved</div><div class="header and">should be preserved</div><div class="menu shadow">should be preserved</div></body></html>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			removeUnlikelyCandidates(doc)

			result, err := doc.Html()
			if err != nil {
				t.Fatalf("Failed to get HTML: %v", err)
			}

			// Normalize whitespace for comparison
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("\nExpected:\n%s\n\nGot:\n%s", expected, result)
			}
		})
	}
}

func TestRemoveUnlikelyCandidatesShouldRemoveFunction(t *testing.T) {
	// Test the internal shouldRemove function behavior through the public interface
	testCases := []struct {
		name     string
		attr     string
		attrType string // "class" or "id"
		expected bool   // true if should be removed
	}{
		// Special hardcoded cases
		{"popupbody in class", "popupbody", "class", true},
		{"contains popupbody in class", "main-popupbody-content", "class", true},
		{"ad suffix in class", "super-ad", "class", true},
		{"ad in middle of class", "pre-ad-post", "class", true},
		{"g-plus in class", "g-plus-share", "class", true},
		{"contains g-plus in class", "social-g-plus-button", "class", true},

		// Unlikely candidates regexp
		{"banner class", "banner", "class", true},
		{"breadcrumbs class", "breadcrumbs", "class", true},
		{"comment class", "comment", "class", true},
		{"sidebar class", "sidebar", "class", true},
		{"footer class", "footer", "class", true},

		// Unlikely candidates with good candidates (should not be removed)
		{"banner with article", "banner article", "class", false},
		{"comment with main", "comment main", "class", false},
		{"sidebar with body", "sidebar body", "class", false},
		{"footer with column", "footer column", "class", false},
		{"menu with shadow", "menu shadow", "class", false},

		// Case insensitive
		{"uppercase banner", "BANNER", "class", true},
		{"mixed case comment", "Comment", "class", true},
		{"uppercase with good", "BANNER ARTICLE", "class", false},

		// ID attributes
		{"banner id", "banner", "id", true},
		{"comment id", "comment", "id", true},
		{"banner with article id", "banner article", "id", false},

		// Good candidates only
		{"article class", "article", "class", false},
		{"main class", "main", "class", false},
		{"content class", "content", "class", false},
		{"body class", "body", "class", false},

		// No matches
		{"random class", "random-class", "class", false},
		{"normal content", "normal-content", "class", false},
		{"empty string", "", "class", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var html string
			if tc.attrType == "class" {
				html = `<html><body><div class="` + tc.attr + `">content</div></body></html>`
			} else {
				html = `<html><body><div id="` + tc.attr + `">content</div></body></html>`
			}

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Count elements before removal
			beforeCount := doc.Find("div").Length()

			removeUnlikelyCandidates(doc)

			// Count elements after removal
			afterCount := doc.Find("div").Length()

			wasRemoved := beforeCount > afterCount

			if wasRemoved != tc.expected {
				t.Errorf("Expected element to be removed: %v, but was removed: %v", tc.expected, wasRemoved)
			}
		})
	}
}

func TestRemoveUnlikelyCandidatesPreservation(t *testing.T) {
	testCases := []struct {
		name        string
		html        string
		description string
	}{
		{
			name:        "preserves html tag",
			html:        `<html class="banner sidebar footer"><body><div>content</div></body></html>`,
			description: "HTML tag should never be removed regardless of class",
		},
		{
			name:        "preserves body tag",
			html:        `<html><body class="banner sidebar footer"><div>content</div></body></html>`,
			description: "Body tag should never be removed regardless of class",
		},
		{
			name:        "preserves elements in pre tags",
			html:        `<html><body><pre><span class="banner">code</span></pre></body></html>`,
			description: "Elements within pre tags should be preserved",
		},
		{
			name:        "preserves elements in code tags",
			html:        `<html><body><code><span class="sidebar">code</span></code></body></html>`,
			description: "Elements within code tags should be preserved",
		},
		{
			name:        "preserves nested elements in code blocks",
			html:        `<html><body><pre><code><div class="comment"><span class="banner">nested</span></div></code></pre></body></html>`,
			description: "Deeply nested elements in code blocks should be preserved",
		},
		{
			name:        "preserves elements in mixed code scenarios",
			html:        `<html><body><div class="main"><pre><span class="sidebar">code</span></pre><code><div class="banner">more code</div></code></div></body></html>`,
			description: "Multiple code block scenarios should work correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Count specific elements before removal
			beforeHtml := doc.Find("html").Length()
			beforeBody := doc.Find("body").Length()
			beforePre := doc.Find("pre").Length()
			beforeCode := doc.Find("code").Length()

			removeUnlikelyCandidates(doc)

			// Count specific elements after removal
			afterHtml := doc.Find("html").Length()
			afterBody := doc.Find("body").Length()
			afterPre := doc.Find("pre").Length()
			afterCode := doc.Find("code").Length()

			// These elements should always be preserved
			if beforeHtml != afterHtml {
				t.Errorf("HTML elements were removed: before=%d, after=%d", beforeHtml, afterHtml)
			}
			if beforeBody != afterBody {
				t.Errorf("Body elements were removed: before=%d, after=%d", beforeBody, afterBody)
			}
			if beforePre != afterPre {
				t.Errorf("Pre elements were removed: before=%d, after=%d", beforePre, afterPre)
			}
			if beforeCode != afterCode {
				t.Errorf("Code elements were removed: before=%d, after=%d", beforeCode, afterCode)
			}

			// Verify that elements within code blocks are preserved
			if tc.name == "preserves elements in pre tags" || tc.name == "preserves elements in code tags" || tc.name == "preserves nested elements in code blocks" {
				spanInCode := doc.Find("pre span, code span, pre div, code div").Length()
				if spanInCode == 0 {
					t.Error("Elements within code blocks were incorrectly removed")
				}
			}
		})
	}
}
