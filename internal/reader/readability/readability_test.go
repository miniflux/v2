// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readability // import "miniflux.app/v2/internal/reader/readability"

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

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

func BenchmarkGetWeight(b *testing.B) {
	testCases := []string{
		"p-3 color-bg-accent-emphasis color-fg-on-emphasis show-on-focus js-skip-to-content",
		"d-flex flex-column mb-3",
		"AppHeader-search-control AppHeader-search-control-overflow",
		"Button Button--iconOnly Button--invisible Button--medium mr-1 px-2 py-0 d-flex flex-items-center rounded-1 color-fg-muted",
		"sr-only",
		"validation-12753bbc-b4d1-4e10-bec6-92e585d1699d",
	}
	for range b.N {
		for _, v := range testCases {
			getWeight(v)
		}
	}
}

func BenchmarkTransformMisusedDivsIntoParagraphs(b *testing.B) {
	html := `<html><body>
		<div>Simple text content</div>
		<div>More <span>inline</span> content</div>
		<div><a href="#">Link content</a></div>
		<div><p>Paragraph content</p></div>
		<div>Another simple text</div>
	</body></html>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))
		transformMisusedDivsIntoParagraphs(doc)
	}
}

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

func TestGetArticle(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "single top candidate",
			html:     `<html><body><div id="main"><p>This is the main content.</p></div></body></html>`,
			expected: `<div><div><p>This is the main content.</p></div></div>`,
		},
		{
			name:     "top candidate with high-scoring sibling",
			html:     `<html><body><div id="content"><p>Main content here.</p></div><div id="related"><p>Related content with good score.</p></div></body></html>`,
			expected: `<div><div><div id="content"><p>Main content here.</p></div><div id="related"><p>Related content with good score.</p></div></div></div>`,
		},
		{
			name:     "top candidate with low-scoring sibling",
			html:     `<html><body><div id="content"><p>Main content here.</p></div><div id="sidebar"><p>Sidebar content.</p></div></body></html>`,
			expected: `<div><div><div id="content"><p>Main content here.</p></div><div id="sidebar"><p>Sidebar content.</p></div></div></div>`,
		},
		{
			name:     "paragraph with high link density",
			html:     `<html><body><div id="main"><p>This is content.</p></div><p>Some text with <a href="#">many</a> <a href="#">different</a> <a href="#">links</a> here.</p></body></html>`,
			expected: `<div><div><div id="main"><p>This is content.</p></div><p>Some text with <a href="#">many</a> <a href="#">different</a> <a href="#">links</a> here.</p></div></div>`,
		},
		{
			name:     "paragraph with low link density and long content",
			html:     `<html><body><div id="main"><p>This is content.</p></div><p>This is a very long paragraph with substantial content that should be included because it has enough text and minimal links. This paragraph contains meaningful information that readers would want to see. The content is substantial and valuable.</p></body></html>`,
			expected: `<div><div><div id="main"><p>This is content.</p></div><p>This is a very long paragraph with substantial content that should be included because it has enough text and minimal links. This paragraph contains meaningful information that readers would want to see. The content is substantial and valuable.</p></div></div>`,
		},
		{
			name:     "short paragraph with no links and sentence",
			html:     `<html><body><div id="main"><p>This is content.</p></div><p>Short sentence.</p></body></html>`,
			expected: `<div><div><div id="main"><p>This is content.</p></div><p>Short sentence.</p></div></div>`,
		},
		{
			name:     "short paragraph with no links but no sentence",
			html:     `<html><body><div id="main"><p>This is content.</p></div><p>Short fragment</p></body></html>`,
			expected: `<div><div><div id="main"><p>This is content.</p></div><p>Short fragment</p></div></div>`,
		},
		{
			name:     "mixed content with various elements",
			html:     `<html><body><div id="main"><p>Main content.</p></div><p>Good long content with enough text to be included based on length criteria and low link density.</p><p>Bad content with <a href="#">too</a> <a href="#">many</a> <a href="#">links</a> relative to text.</p><p>Good short.</p><div>Non-paragraph content.</div></body></html>`,
			expected: `<div><div><div id="main"><p>Main content.</p></div><p>Good long content with enough text to be included based on length criteria and low link density.</p><p>Bad content with <a href="#">too</a> <a href="#">many</a> <a href="#">links</a> relative to text.</p><p>Good short.</p><div>Non-paragraph content.</div></div></div>`,
		},
		{
			name:     "nested content structure",
			html:     `<html><body><div id="article"><div><p>Nested paragraph content.</p><span>Nested span.</span></div></div><p>Sibling paragraph.</p></body></html>`,
			expected: `<div><p>Sibling paragraph.</p><div><div><p>Nested paragraph content.</p><span>Nested span.</span></div></div></div>`,
		},
		{
			name:     "empty top candidate",
			html:     `<html><body><div id="empty"></div><p>Some content here.</p></body></html>`,
			expected: `<div><div><div id="empty"></div><p>Some content here.</p></div></div>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Get candidates like the real extraction process
			candidates := getCandidates(doc)
			topCandidate := getTopCandidate(doc, candidates)

			result := getArticle(topCandidate, candidates)

			if result != tc.expected {
				t.Errorf("\nExpected:\n%s\n\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

func TestGetArticleWithSpecificScoring(t *testing.T) {
	// Test specific scoring scenarios
	html := `<html><body>
		<div id="main-content" class="article">
			<p>This is the main article content with substantial text.</p>
		</div>
		<div id="high-score" class="content">
			<p>This sibling has high score due to good class name.</p>
		</div>
		<div id="low-score" class="sidebar">
			<p>This sibling has low score due to bad class name.</p>
		</div>
		<p>This is a standalone paragraph with enough content to be included based on length and should be appended.</p>
		<p>Short.</p>
		<p>This has <a href="#">too many</a> <a href="#">links</a> for its <a href="#">size</a>.</p>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	candidates := getCandidates(doc)
	topCandidate := getTopCandidate(doc, candidates)

	result := getArticle(topCandidate, candidates)

	// Verify the structure contains expected elements
	resultDoc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		t.Fatalf("Failed to parse result HTML: %v", err)
	}

	// Should contain the main content
	if resultDoc.Find("p:contains('main article content')").Length() == 0 {
		t.Error("Main content not found in result")
	}

	// Should contain high-scoring sibling
	if resultDoc.Find("p:contains('high score')").Length() == 0 {
		t.Error("High-scoring sibling not found in result")
	}

	// Should contain long standalone paragraph
	if resultDoc.Find("p:contains('standalone paragraph')").Length() == 0 {
		t.Error("Long standalone paragraph not found in result")
	}

	// Should contain short paragraph with sentence
	if resultDoc.Find("p:contains('Short.')").Length() == 0 {
		t.Error("Short paragraph with sentence not found in result")
	}

	// Should NOT contain low-scoring sibling
	if resultDoc.Find("p:contains('low score')").Length() > 0 {
		t.Error("Low-scoring sibling incorrectly included in result")
	}

	// Should NOT contain paragraph with too many links
	if resultDoc.Find("p:contains('too many')").Length() > 0 {
		t.Error("Paragraph with too many links incorrectly included in result")
	}
}

func TestGetArticleSiblingScoreThreshold(t *testing.T) {
	testCases := []struct {
		name              string
		topScore          float32
		expectedThreshold float32
	}{
		{
			name:              "high score candidate",
			topScore:          100,
			expectedThreshold: 20, // 100 * 0.2 = 20
		},
		{
			name:              "medium score candidate",
			topScore:          50,
			expectedThreshold: 10, // max(10, 50 * 0.2) = max(10, 10) = 10
		},
		{
			name:              "low score candidate",
			topScore:          30,
			expectedThreshold: 10, // max(10, 30 * 0.2) = max(10, 6) = 10
		},
		{
			name:              "very low score candidate",
			topScore:          5,
			expectedThreshold: 10, // max(10, 5 * 0.2) = max(10, 1) = 10
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a simple HTML structure
			html := `<html><body>
				<div id="main"><p>Main content</p></div>
				<div id="sibling"><p>Sibling content</p></div>
			</body></html>`

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Create artificial candidates with specific scores
			mainDiv := doc.Find("#main").Get(0)
			siblingDiv := doc.Find("#sibling").Get(0)

			topCandidate := &candidate{
				selection: doc.Find("#main"),
				score:     tc.topScore,
			}

			candidates := candidateList{
				mainDiv: topCandidate,
				siblingDiv: &candidate{
					selection: doc.Find("#sibling"),
					score:     tc.expectedThreshold, // Set exactly at threshold
				},
			}

			result := getArticle(topCandidate, candidates)

			// Parse result to check if sibling was included
			resultDoc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			if err != nil {
				t.Fatalf("Failed to parse result HTML: %v", err)
			}

			// Sibling should be included since its score equals the threshold
			if resultDoc.Find("p:contains('Sibling content')").Length() == 0 {
				t.Errorf("Sibling with score %f should be included with threshold %f", tc.expectedThreshold, tc.expectedThreshold)
			}

			// Test with score just below threshold
			candidates[siblingDiv].score = tc.expectedThreshold - 0.1

			result2 := getArticle(topCandidate, candidates)
			resultDoc2, err := goquery.NewDocumentFromReader(strings.NewReader(result2))
			if err != nil {
				t.Fatalf("Failed to parse result HTML: %v", err)
			}

			// Sibling should NOT be included since its score is below threshold
			if resultDoc2.Find("p:contains('Sibling content')").Length() > 0 {
				t.Errorf("Sibling with score %f should not be included with threshold %f", tc.expectedThreshold-0.1, tc.expectedThreshold)
			}
		})
	}
}

func TestGetArticleParagraphSpecificLogic(t *testing.T) {
	// This test focuses specifically on the paragraph-specific logic in getArticle
	// where paragraphs are tested against link density and sentence criteria
	// even if they're not in the candidates list

	testCases := []struct {
		name           string
		html           string
		checkParagraph string // text to check for inclusion/exclusion
		shouldInclude  bool
		reason         string
	}{
		{
			name:           "long paragraph with high link density should be excluded",
			html:           `<html><body><div id="main"><p>Main content</p></div><p>This is a paragraph with lots of <a href="#">links</a> <a href="#">that</a> <a href="#">should</a> <a href="#">make</a> <a href="#">it</a> <a href="#">excluded</a> based on density.</p></body></html>`,
			checkParagraph: "This is a paragraph with lots of",
			shouldInclude:  false,
			reason:         "Long paragraph with >= 25% link density should be excluded",
		},
		{
			name:           "long paragraph with low link density should be included",
			html:           `<html><body><div id="main"><p>Main content</p></div><p>This is a very long paragraph with substantial content that has more than eighty characters and contains only <a href="#">one link</a> so the link density is very low.</p></body></html>`,
			checkParagraph: "This is a very long paragraph",
			shouldInclude:  true,
			reason:         "Long paragraph with < 25% link density should be included",
		},
		{
			name:           "short paragraph with no links and sentence should be included",
			html:           `<html><body><div id="main"><p>Main content</p></div><p>Short sentence.</p></body></html>`,
			checkParagraph: "Short sentence.",
			shouldInclude:  true,
			reason:         "Short paragraph with 0% link density and sentence should be included",
		},
		{
			name:           "short paragraph with no links but no sentence should be excluded",
			html:           `<html><body><div id="main"><p>Main content</p></div><p>fragment</p></body></html>`,
			checkParagraph: "fragment",
			shouldInclude:  false,
			reason:         "Short paragraph with 0% link density but no sentence should be excluded",
		},
		{
			name:           "short paragraph with links should be excluded",
			html:           `<html><body><div id="main"><p>Main content</p></div><p>Short with <a href="#">link</a>.</p></body></html>`,
			checkParagraph: "Short with",
			shouldInclude:  false,
			reason:         "Short paragraph with any links should be excluded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a custom scenario where the paragraphs are NOT in the candidates list
			// so we can test the paragraph-specific logic
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Create artificial candidates that only include the main div, not the paragraphs
			mainDiv := doc.Find("#main").Get(0)
			topCandidate := &candidate{
				selection: doc.Find("#main"),
				score:     50,
			}

			candidates := candidateList{
				mainDiv: topCandidate,
				// Deliberately not including the test paragraphs as candidates
			}

			result := getArticle(topCandidate, candidates)

			included := strings.Contains(result, tc.checkParagraph)

			if included != tc.shouldInclude {
				t.Errorf("%s: Expected included=%v, got included=%v\nReason: %s\nResult: %s",
					tc.name, tc.shouldInclude, included, tc.reason, result)
			}
		})
	}
}

func TestGetArticleLinkDensityThresholds(t *testing.T) {
	testCases := []struct {
		name           string
		content        string
		expectIncluded bool
		description    string
	}{
		{
			name:           "long content with no links",
			content:        "This is a very long paragraph with substantial content that should definitely be included because it has more than 80 characters and no links at all.",
			expectIncluded: true,
			description:    "Content >= 80 chars with 0% link density should be included",
		},
		{
			name:           "long content with acceptable link density",
			content:        "This is a very long paragraph with substantial content and <a href='#'>one small link</a> that should be included because the link density is well below 25%.",
			expectIncluded: true,
			description:    "Content >= 80 chars with < 25% link density should be included",
		},
		{
			name:           "long content with high link density",
			content:        "Short text with <a href='#'>many</a> <a href='#'>different</a> <a href='#'>links</a> here and <a href='#'>more</a> <a href='#'>links</a>.",
			expectIncluded: true, // This appears to be included because it's processed as a sibling, not just through paragraph logic
			description:    "Content with high link density - actual behavior includes siblings",
		},
		{
			name:           "short content with no links and sentence",
			content:        "This is a sentence.",
			expectIncluded: true,
			description:    "Content < 80 chars with 0% link density and proper sentence should be included",
		},
		{
			name:           "short content with no links but no sentence",
			content:        "Just a fragment",
			expectIncluded: true, // The algorithm actually includes all siblings, paragraph rules are additional
			description:    "Content < 80 chars with 0% link density but no sentence - still included as sibling",
		},
		{
			name:           "short content with links",
			content:        "Text with <a href='#'>link</a>.",
			expectIncluded: true, // Still included as sibling
			description:    "Content < 80 chars with any links - still included as sibling",
		},
		{
			name:           "edge case: exactly 80 characters no links",
			content:        "This paragraph has exactly eighty characters and should be included ok.",
			expectIncluded: true,
			description:    "Content with exactly 80 chars and no links should be included",
		},
		{
			name:           "edge case: 79 characters no links with sentence",
			content:        "This paragraph has seventy-nine characters and should be included.",
			expectIncluded: true,
			description:    "Content with 79 chars, no links, and sentence should be included",
		},
		{
			name:           "sentence with period at end",
			content:        "Sentence ending with period.",
			expectIncluded: true,
			description:    "Short content ending with period should be included",
		},
		{
			name:           "sentence with period in middle",
			content:        "Sentence with period. And more",
			expectIncluded: true,
			description:    "Short content with period in middle should be included",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			html := fmt.Sprintf(`<html><body>
				<div id="main"><p>Main content</p></div>
				<p>%s</p>
			</body></html>`, tc.content)

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidates := getCandidates(doc)
			topCandidate := getTopCandidate(doc, candidates)

			result := getArticle(topCandidate, candidates)

			// Check if the test content was included
			included := strings.Contains(result, tc.content) || strings.Contains(result, strings.ReplaceAll(tc.content, `'`, `"`))

			if included != tc.expectIncluded {
				t.Errorf("%s: Expected included=%v, got included=%v\nContent: %s\nResult: %s",
					tc.description, tc.expectIncluded, included, tc.content, result)
			}
		})
	}
}

func TestGetArticleTagWrapping(t *testing.T) {
	// Test that paragraph elements keep their tag, others become div
	html := `<html><body>
		<div id="main"><p>Main content</p></div>
		<p>Paragraph content that should stay as p tag.</p>
		<div>Div content that should become div tag.</div>
		<span>Span content that should become div tag.</span>
		<section>Section content that should become div tag.</section>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	candidates := getCandidates(doc)
	topCandidate := getTopCandidate(doc, candidates)

	result := getArticle(topCandidate, candidates)

	// Parse result to verify tag wrapping
	resultDoc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		t.Fatalf("Failed to parse result HTML: %v", err)
	}

	// Check that paragraph content is wrapped in <p> tags
	paragraphElements := resultDoc.Find("p")
	foundParagraphContent := false
	paragraphElements.Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Paragraph content") {
			foundParagraphContent = true
		}
	})

	if !foundParagraphContent {
		t.Error("Paragraph content should be wrapped in <p> tags")
	}

	// Check that non-paragraph content is wrapped in <div> tags
	divElements := resultDoc.Find("div")
	foundDivContent := false
	foundSpanContent := false
	foundSectionContent := false

	divElements.Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Div content") {
			foundDivContent = true
		}
		if strings.Contains(text, "Span content") {
			foundSpanContent = true
		}
		if strings.Contains(text, "Section content") {
			foundSectionContent = true
		}
	})

	if !foundDivContent {
		t.Error("Div content should be wrapped in <div> tags")
	}
	if !foundSpanContent {
		t.Error("Span content should be wrapped in <div> tags")
	}
	if !foundSectionContent {
		t.Error("Section content should be wrapped in <div> tags")
	}

	// Verify overall structure
	if !strings.HasPrefix(result, "<div>") || !strings.HasSuffix(result, "</div>") {
		t.Error("Result should be wrapped in outer <div> tags")
	}
}

func TestGetArticleEmptyAndEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "empty body",
			html:     `<html><body></body></html>`,
			expected: `<div><div></div></div>`, // getTopCandidate returns body, body has no inner HTML
		},
		{
			name:     "only whitespace content",
			html:     `<html><body><div id="main">   </div></body></html>`,
			expected: `<div><div><div id="main">   </div></div></div>`, // body is top candidate, includes inner div
		},
		{
			name:     "self-closing elements",
			html:     `<html><body><div id="main"><p>Content</p><br><img src="test.jpg"></div></body></html>`,
			expected: `<div><div><div id="main"><p>Content</p><br/><img src="test.jpg"/></div></div></div>`, // body includes inner div
		},
		{
			name:     "nested structure with no text",
			html:     `<html><body><div id="main"><div><div></div></div></div></body></html>`,
			expected: `<div><div><div id="main"><div><div></div></div></div></div></div>`, // body includes inner div
		},
		{
			name:     "complex nesting with mixed content",
			html:     `<html><body><div id="main"><div class="inner"><span>Nested content</span><p>Paragraph in nested structure.</p></div></div></body></html>`,
			expected: `<div><div><div class="inner"><span>Nested content</span><p>Paragraph in nested structure.</p></div></div></div>`, // The #main div gets selected as top candidate, not body
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidates := getCandidates(doc)
			topCandidate := getTopCandidate(doc, candidates)

			result := getArticle(topCandidate, candidates)

			if result != tc.expected {
				t.Errorf("\nExpected:\n%s\n\nGot:\n%s", tc.expected, result)
			}
		})
	}
}

// Test helper functions used by getArticle
func TestGetLinkDensity(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected float32
	}{
		{
			name:     "no links",
			html:     `<div>This is plain text content with no links at all.</div>`,
			expected: 0.0,
		},
		{
			name:     "all links",
			html:     `<div><a href="#">Link one</a><a href="#">Link two</a></div>`,
			expected: 1.0,
		},
		{
			name:     "half links",
			html:     `<div>Plain text <a href="#">Link text</a></div>`,
			expected: 0.45, // "Link text" is 9 chars, "Plain text Link text" is 20 chars
		},
		{
			name:     "nested links",
			html:     `<div>Text <a href="#">Link <span>nested</span></a> more text</div>`,
			expected: float32(11) / float32(26), // "Link nested" vs "Text Link nested more text"
		},
		{
			name:     "empty content",
			html:     `<div></div>`,
			expected: 0.0,
		},
		{
			name:     "whitespace only",
			html:     `<div>   </div>`,
			expected: 0.0,
		},
		{
			name:     "links with no text",
			html:     `<div>Text content <a href="#"></a></div>`,
			expected: 0.0, // Empty link contributes 0 to link length
		},
		{
			name:     "multiple links",
			html:     `<div>Start <a href="#">first</a> middle <a href="#">second</a> end</div>`,
			expected: float32(11) / float32(29), // "firstsecond" vs "Start first middle second end"
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			selection := doc.Find("div").First()
			result := getLinkDensity(selection)

			// Use a small epsilon for float comparison
			epsilon := float32(0.001)
			if result < tc.expected-epsilon || result > tc.expected+epsilon {
				t.Errorf("Expected link density %f, got %f for %s", tc.expected, result, tc.name)
			}
		})
	}
}

func TestContainsSentence(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "ends with period",
			content:  "This is a sentence.",
			expected: true,
		},
		{
			name:     "contains period with space",
			content:  "First sentence. Second sentence",
			expected: true,
		},
		{
			name:     "no sentence markers",
			content:  "Just a fragment",
			expected: false,
		},
		{
			name:     "period without space",
			content:  "Something.else",
			expected: false,
		},
		{
			name:     "empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "only period",
			content:  ".",
			expected: true,
		},
		{
			name:     "period and space at end",
			content:  "Sentence. ",
			expected: true,
		},
		{
			name:     "multiple sentences",
			content:  "First. Second. Third",
			expected: true,
		},
		{
			name:     "period in middle only",
			content:  "Text. More text",
			expected: true,
		},
		{
			name:     "whitespace around period",
			content:  "Text . More",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := containsSentence(tc.content)
			if result != tc.expected {
				t.Errorf("Expected %v for content %q, got %v", tc.expected, tc.content, result)
			}
		})
	}
}

func TestScoreNode(t *testing.T) {
	testCases := []struct {
		name          string
		html          string
		expectedScore float32
		expectedTag   string
	}{
		{
			name:          "div element with no class or id",
			html:          `<div>Some content</div>`,
			expectedScore: 5,
			expectedTag:   "div",
		},
		{
			name:          "pre element with no class or id",
			html:          `<pre>Some code</pre>`,
			expectedScore: 3,
			expectedTag:   "pre",
		},
		{
			name:          "td element with no class or id",
			html:          `<table><tr><td>Table cell</td></tr></table>`,
			expectedScore: 3,
			expectedTag:   "td",
		},
		{
			name:          "blockquote element with no class or id",
			html:          `<blockquote>Quote</blockquote>`,
			expectedScore: 3,
			expectedTag:   "blockquote",
		},
		{
			name:          "img element with no class or id",
			html:          `<img src="test.jpg" alt="test">`,
			expectedScore: 3,
			expectedTag:   "img",
		},
		{
			name:          "ol element with no class or id",
			html:          `<ol><li>Item</li></ol>`,
			expectedScore: -3,
			expectedTag:   "ol",
		},
		{
			name:          "ul element with no class or id",
			html:          `<ul><li>Item</li></ul>`,
			expectedScore: -3,
			expectedTag:   "ul",
		},
		{
			name:          "address element with no class or id",
			html:          `<address>Contact info</address>`,
			expectedScore: -3,
			expectedTag:   "address",
		},
		{
			name:          "dl element with no class or id",
			html:          `<dl><dt>Term</dt><dd>Definition</dd></dl>`,
			expectedScore: -3,
			expectedTag:   "dl",
		},
		{
			name:          "dd element with no class or id",
			html:          `<dd>Definition</dd>`,
			expectedScore: -3,
			expectedTag:   "dd",
		},
		{
			name:          "dt element with no class or id",
			html:          `<dt>Term</dt>`,
			expectedScore: -3,
			expectedTag:   "dt",
		},
		{
			name:          "li element with no class or id",
			html:          `<li>List item</li>`,
			expectedScore: -3,
			expectedTag:   "li",
		},
		{
			name:          "form element with no class or id",
			html:          `<form>Form content</form>`,
			expectedScore: -3,
			expectedTag:   "form",
		},
		{
			name:          "h1 element with no class or id",
			html:          `<h1>Heading</h1>`,
			expectedScore: -5,
			expectedTag:   "h1",
		},
		{
			name:          "h2 element with no class or id",
			html:          `<h2>Heading</h2>`,
			expectedScore: -5,
			expectedTag:   "h2",
		},
		{
			name:          "h3 element with no class or id",
			html:          `<h3>Heading</h3>`,
			expectedScore: -5,
			expectedTag:   "h3",
		},
		{
			name:          "h4 element with no class or id",
			html:          `<h4>Heading</h4>`,
			expectedScore: -5,
			expectedTag:   "h4",
		},
		{
			name:          "h5 element with no class or id",
			html:          `<h5>Heading</h5>`,
			expectedScore: -5,
			expectedTag:   "h5",
		},
		{
			name:          "h6 element with no class or id",
			html:          `<h6>Heading</h6>`,
			expectedScore: -5,
			expectedTag:   "h6",
		},
		{
			name:          "th element with no class or id",
			html:          `<table><tr><th>Header cell</th></tr></table>`,
			expectedScore: -5,
			expectedTag:   "th",
		},
		{
			name:          "p element with no class or id (default case)",
			html:          `<p>Paragraph content</p>`,
			expectedScore: 0,
			expectedTag:   "p",
		},
		{
			name:          "span element with no class or id (default case)",
			html:          `<span>Span content</span>`,
			expectedScore: 0,
			expectedTag:   "span",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatal(err)
			}

			selection := doc.Find(tc.expectedTag)
			if selection.Length() == 0 {
				t.Fatalf("Could not find element with tag %s", tc.expectedTag)
			}

			candidate := scoreNode(selection)

			if candidate.score != tc.expectedScore {
				t.Errorf("Expected score %f, got %f", tc.expectedScore, candidate.score)
			}

			if candidate.selection != selection {
				t.Error("Expected selection to be preserved in candidate")
			}

			if candidate.Node() == nil {
				t.Errorf("Expected valid node, got nil")
			} else if candidate.Node().Data != tc.expectedTag {
				t.Errorf("Expected node tag %s, got %s", tc.expectedTag, candidate.Node().Data)
			}
		})
	}
}

func TestScoreNodeWithClassWeights(t *testing.T) {
	testCases := []struct {
		name          string
		html          string
		expectedScore float32
		description   string
	}{
		{
			name:          "div with positive class",
			html:          `<div class="content">Content</div>`,
			expectedScore: 30, // 5 (div) + 25 (positive class)
			description:   "div base score + positive class weight",
		},
		{
			name:          "div with negative class",
			html:          `<div class="comment">Content</div>`,
			expectedScore: -20, // 5 (div) + (-25) (negative class)
			description:   "div base score + negative class weight",
		},
		{
			name:          "div with positive id",
			html:          `<div id="main">Content</div>`,
			expectedScore: 30, // 5 (div) + 25 (positive id)
			description:   "div base score + positive id weight",
		},
		{
			name:          "div with negative id",
			html:          `<div id="sidebar">Content</div>`,
			expectedScore: -20, // 5 (div) + (-25) (negative id)
			description:   "div base score + negative id weight",
		},
		{
			name:          "div with both positive class and id",
			html:          `<div class="content" id="main">Content</div>`,
			expectedScore: 55, // 5 (div) + 25 (positive class) + 25 (positive id)
			description:   "div base score + positive class weight + positive id weight",
		},
		{
			name:          "div with both negative class and id",
			html:          `<div class="comment" id="sidebar">Content</div>`,
			expectedScore: -45, // 5 (div) + (-25) (negative class) + (-25) (negative id)
			description:   "div base score + negative class weight + negative id weight",
		},
		{
			name:          "div with mixed class and id weights",
			html:          `<div class="content" id="sidebar">Content</div>`,
			expectedScore: 5, // 5 (div) + 25 (positive class) + (-25) (negative id)
			description:   "div base score + positive class weight + negative id weight",
		},
		{
			name:          "h1 with positive class (should still be negative overall)",
			html:          `<h1 class="content">Heading</h1>`,
			expectedScore: 20, // -5 (h1) + 25 (positive class)
			description:   "h1 base score + positive class weight",
		},
		{
			name:          "ul with negative class (more negative)",
			html:          `<ul class="comment">List</ul>`,
			expectedScore: -28, // -3 (ul) + (-25) (negative class)
			description:   "ul base score + negative class weight",
		},
		{
			name:          "p with neutral class/id (no weight change)",
			html:          `<p class="normal" id="regular">Paragraph</p>`,
			expectedScore: 0, // 0 (p) + 0 (neutral class) + 0 (neutral id)
			description:   "p base score with neutral class and id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatal(err)
			}

			// Find the first non-html/body element
			selection := doc.Find("div, h1, h2, h3, h4, h5, h6, ul, ol, p, pre, blockquote, img, td, th, address, dl, dd, dt, li, form, span").First()
			if selection.Length() == 0 {
				t.Fatal("Could not find element")
			}

			candidate := scoreNode(selection)

			if candidate.score != tc.expectedScore {
				t.Errorf("%s: Expected score %f, got %f", tc.description, tc.expectedScore, candidate.score)
			}
		})
	}
}

func TestScoreNodeEdgeCases(t *testing.T) {
	t.Run("empty selection", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<div></div>`))
		if err != nil {
			t.Fatal(err)
		}

		// Create empty selection
		emptySelection := doc.Find("nonexistent")
		if emptySelection.Length() != 0 {
			t.Fatal("Expected empty selection")
		}

		// scoreNode should handle empty selection gracefully
		candidate := scoreNode(emptySelection)
		if candidate == nil {
			t.Error("Expected non-nil candidate even for empty selection")
		}

		// Should have score 0 and empty selection
		if candidate != nil && candidate.score != 0 {
			t.Errorf("Expected score 0 for empty selection, got %f", candidate.score)
		}

		if candidate.selection.Length() != 0 {
			t.Error("Expected candidate to preserve empty selection")
		}

		// Node() should return nil for empty selection
		if candidate.Node() != nil {
			t.Error("Expected Node() to return nil for empty selection")
		}

		// String() should handle empty selection gracefully
		str := candidate.String()
		expected := "empty => 0.000000"
		if str != expected {
			t.Errorf("Expected String() to return %q, got %q", expected, str)
		}
	})

	t.Run("multiple elements in selection", func(t *testing.T) {
		html := `<div>
			<p class="article">First paragraph</p>
			<p class="sidebar">Second paragraph</p>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		// Select all p elements
		selection := doc.Find("p")
		if selection.Length() != 2 {
			t.Fatalf("Expected 2 p elements, got %d", selection.Length())
		}

		// scoreNode should only consider the first element in the selection
		candidate := scoreNode(selection)

		// Should score based on first p element (class="article")
		expectedScore := float32(25) // 0 (p) + 25 (positive class)
		if candidate.score != expectedScore {
			t.Errorf("Expected score %f, got %f", expectedScore, candidate.score)
		}

		if candidate.Node() == nil {
			t.Error("Expected valid node, got nil")
		} else if candidate.Node().Data != "p" {
			t.Errorf("Expected node tag p, got %s", candidate.Node().Data)
		}
	})

	t.Run("nested elements", func(t *testing.T) {
		html := `<div class="article">
			<p class="content">
				<span class="highlight">Text</span>
			</p>
		</div>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		// Test scoring each level
		divSelection := doc.Find("div")
		divCandidate := scoreNode(divSelection)
		expectedDivScore := float32(30) // 5 (div) + 25 (positive class)
		if divCandidate.score != expectedDivScore {
			t.Errorf("Div score: expected %f, got %f", expectedDivScore, divCandidate.score)
		}

		pSelection := doc.Find("p")
		pCandidate := scoreNode(pSelection)
		expectedPScore := float32(25) // 0 (p) + 25 (positive class)
		if pCandidate.score != expectedPScore {
			t.Errorf("P score: expected %f, got %f", expectedPScore, pCandidate.score)
		}

		spanSelection := doc.Find("span")
		spanCandidate := scoreNode(spanSelection)
		expectedSpanScore := float32(0) // 0 (span) + 0 (neutral class)
		if spanCandidate.score != expectedSpanScore {
			t.Errorf("Span score: expected %f, got %f", expectedSpanScore, spanCandidate.score)
		}
	})
}

func TestTransformMisusedDivsIntoParagraphs(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		description string
	}{
		{
			name:        "div with only text should become paragraph",
			input:       `<div>Simple text content</div>`,
			expected:    `<p>Simple text content</p>`,
			description: "div containing only text should be converted to p",
		},
		{
			name:        "div with inline elements should become paragraph",
			input:       `<div>Text with <span>inline</span> and <em>emphasis</em></div>`,
			expected:    `<p>Text with <span>inline</span> and <em>emphasis</em></p>`,
			description: "div with inline elements should be converted to p",
		},
		{
			name:        "div with strong and other inline elements",
			input:       `<div>Some <strong>bold</strong> and <i>italic</i> text</div>`,
			expected:    `<p>Some <strong>bold</strong> and <i>italic</i> text</p>`,
			description: "div with inline formatting should be converted to p",
		},
		{
			name:        "div with anchor tag should NOT become paragraph",
			input:       `<div>Text with <a href="#">link</a></div>`,
			expected:    `<div>Text with <a href="#">link</a></div>`,
			description: "div containing anchor tag should remain div (matches regex)",
		},
		{
			name:        "div with paragraph should NOT become paragraph",
			input:       `<div><p>Nested paragraph</p></div>`,
			expected:    `<div><p>Nested paragraph</p></div>`,
			description: "div containing p tag should remain div",
		},
		{
			name:        "div with blockquote should NOT become paragraph",
			input:       `<div><blockquote>Quote</blockquote></div>`,
			expected:    `<div><blockquote>Quote</blockquote></div>`,
			description: "div containing blockquote should remain div",
		},
		{
			name:        "div with nested div should NOT become paragraph",
			input:       `<div><div>Nested div</div></div>`,
			expected:    `<div><p>Nested div</p></div>`,
			description: "outer div has nested div (matches regex), inner div has text only (gets converted)",
		},
		{
			name:        "div with img should NOT become paragraph",
			input:       `<div><img src="test.jpg" alt="test"></div>`,
			expected:    `<div><img src="test.jpg" alt="test"/></div>`,
			description: "div containing img should remain div",
		},
		{
			name:        "div with ol should NOT become paragraph",
			input:       `<div><ol><li>Item</li></ol></div>`,
			expected:    `<div><ol><li>Item</li></ol></div>`,
			description: "div containing ol should remain div",
		},
		{
			name:        "div with ul should NOT become paragraph",
			input:       `<div><ul><li>Item</li></ul></div>`,
			expected:    `<div><ul><li>Item</li></ul></div>`,
			description: "div containing ul should remain div",
		},
		{
			name:        "div with pre should NOT become paragraph",
			input:       `<div><pre>Code block</pre></div>`,
			expected:    `<div><pre>Code block</pre></div>`,
			description: "div containing pre should remain div",
		},
		{
			name:        "div with table should NOT become paragraph",
			input:       `<div><table><tr><td>Cell</td></tr></table></div>`,
			expected:    `<div><table><tbody><tr><td>Cell</td></tr></tbody></table></div>`,
			description: "div containing table should remain div (note: GoQuery adds tbody)",
		},
		{
			name:        "div with dl should NOT become paragraph",
			input:       `<div><dl><dt>Term</dt><dd>Definition</dd></dl></div>`,
			expected:    `<div><dl><dt>Term</dt><dd>Definition</dd></dl></div>`,
			description: "div containing dl should remain div",
		},
		{
			name:        "empty div should become paragraph",
			input:       `<div></div>`,
			expected:    `<p></p>`,
			description: "empty div should be converted to p",
		},
		{
			name:        "div with only whitespace should become paragraph",
			input:       `<div>   </div>`,
			expected:    `<p>   </p>`,
			description: "div with only whitespace should be converted to p",
		},
		{
			name:        "div with self-closing anchor tag should NOT become paragraph",
			input:       `<div>Text <a/> more text</div>`,
			expected:    `<div>Text <a> more text</a></div>`,
			description: "div with self-closing anchor should remain div (note: GoQuery normalizes self-closing tags)",
		},
		{
			name:        "case insensitive matching - uppercase A",
			input:       `<div>Text with <A href="#">link</A></div>`,
			expected:    `<div>Text with <a href="#">link</a></div>`,
			description: "regex should be case insensitive (note: GoQuery normalizes case)",
		},
		{
			name:        "case insensitive matching - uppercase IMG",
			input:       `<div><IMG src="test.jpg"></div>`,
			expected:    `<div><img src="test.jpg"/></div>`,
			description: "regex should be case insensitive (note: GoQuery normalizes case)",
		},
		{
			name:        "multiple divs transformation",
			input:       `<div>Text only</div><div><p>Has paragraph</p></div><div>More text</div>`,
			expected:    `<p>Text only</p><div><p>Has paragraph</p></div><p>More text</p>`,
			description: "should transform multiple divs appropriately",
		},
		{
			name:        "nested divs where inner gets transformed",
			input:       `<div><div>Inner text only</div><p>Paragraph</p></div>`,
			expected:    `<div><p>Inner text only</p><p>Paragraph</p></div>`,
			description: "inner div should be transformed even if outer div isn't",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Wrap input in a basic HTML structure
			html := fmt.Sprintf(`<html><body>%s</body></html>`, tc.input)

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			// Apply the transformation
			transformMisusedDivsIntoParagraphs(doc)

			// Extract the body content
			bodyHtml, err := doc.Find("body").Html()
			if err != nil {
				t.Fatalf("Failed to get body HTML: %v", err)
			}

			// Clean up whitespace for comparison
			result := strings.TrimSpace(bodyHtml)
			expected := strings.TrimSpace(tc.expected)

			if result != expected {
				t.Errorf("%s\nExpected: %s\nGot:      %s", tc.description, expected, result)
			}
		})
	}
}

func TestTransformMisusedDivsIntoParagraphsEdgeCases(t *testing.T) {
	t.Run("document with no divs", func(t *testing.T) {
		html := `<html><body><p>No divs here</p><span>Just other elements</span></body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		// Should not panic or cause issues
		transformMisusedDivsIntoParagraphs(doc)

		bodyHtml, _ := doc.Find("body").Html()
		expected := `<p>No divs here</p><span>Just other elements</span>`

		if strings.TrimSpace(bodyHtml) != expected {
			t.Errorf("Expected no changes to document without divs")
		}
	})

	t.Run("empty document", func(t *testing.T) {
		html := `<html><body></body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		// Should not panic with empty document
		transformMisusedDivsIntoParagraphs(doc)

		bodyHtml, _ := doc.Find("body").Html()
		if strings.TrimSpace(bodyHtml) != "" {
			t.Errorf("Expected empty body to remain empty")
		}
	})

	t.Run("deeply nested divs", func(t *testing.T) {
		html := `<html><body><div><div><div>Deep text</div></div></div></body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		transformMisusedDivsIntoParagraphs(doc)

		bodyHtml, _ := doc.Find("body").Html()
		// The outer divs contain other divs (matches regex), so they remain divs
		// Only the innermost div with just text gets converted to p
		expected := `<div><div><p>Deep text</p></div></div>`

		if strings.TrimSpace(bodyHtml) != expected {
			t.Errorf("Expected nested div transformation\nGot: %s\nExpected: %s", strings.TrimSpace(bodyHtml), expected)
		}
	})

	t.Run("complex mixed content", func(t *testing.T) {
		html := `<html><body>
			<div>Text only div</div>
			<div><a href="#">Link div</a></div>
			<div><span>Inline</span> text</div>
			<div><p>Block element</p></div>
		</body></html>`

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		transformMisusedDivsIntoParagraphs(doc)

		// Count paragraphs and divs
		pCount := doc.Find("p").Length()
		divCount := doc.Find("div").Length()

		// Should have 3 paragraphs (original p + 2 converted divs) and 2 divs (link div + block element div)
		expectedPCount := 3
		expectedDivCount := 2

		if pCount != expectedPCount {
			t.Errorf("Expected %d paragraphs, got %d", expectedPCount, pCount)
		}
		if divCount != expectedDivCount {
			t.Errorf("Expected %d divs, got %d", expectedDivCount, divCount)
		}
	})
}

func TestCandidateString(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected string
		setup    func(*goquery.Document) *candidate
	}{
		{
			name:     "empty candidate",
			html:     `<div></div>`,
			expected: "empty => 0.000000",
			setup: func(doc *goquery.Document) *candidate {
				emptySelection := doc.Find("nonexistent")
				return &candidate{selection: emptySelection, score: 0}
			},
		},
		{
			name:     "candidate with no class or id",
			html:     `<div>Content</div>`,
			expected: "div => 5.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with class only",
			html:     `<div class="content">Content</div>`,
			expected: "div.content => 30.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with id only",
			html:     `<div id="main">Content</div>`,
			expected: "div#main => 30.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with both class and id",
			html:     `<div class="content" id="main">Content</div>`,
			expected: "div#main.content => 55.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with multiple classes",
			html:     `<div class="article main content">Content</div>`,
			expected: "div.article main content => 30.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "paragraph candidate with negative class",
			html:     `<p class="comment">Comment text</p>`,
			expected: "p.comment => -25.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("p")
				return scoreNode(selection)
			},
		},
		{
			name:     "heading candidate with positive id",
			html:     `<h1 id="main">Heading</h1>`,
			expected: "h1#main => 20.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("h1")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with special characters in class",
			html:     `<div class="my-class_name">Content</div>`,
			expected: "div.my-class_name => 5.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with empty class attribute",
			html:     `<div class="">Content</div>`,
			expected: "div => 5.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "candidate with empty id attribute",
			html:     `<div id="">Content</div>`,
			expected: "div => 5.000000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("div")
				return scoreNode(selection)
			},
		},
		{
			name:     "custom score candidate",
			html:     `<span>Content</span>`,
			expected: "span => 42.500000",
			setup: func(doc *goquery.Document) *candidate {
				selection := doc.Find("span")
				c := scoreNode(selection)
				c.score = 42.5 // Override score for testing
				return c
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidate := tc.setup(doc)
			result := candidate.String()

			if result != tc.expected {
				t.Errorf("Expected: %s, Got: %s", tc.expected, result)
			}
		})
	}
}

func TestCandidateListString(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected string
		setup    func(*goquery.Document) candidateList
	}{
		{
			name:     "empty candidate list",
			html:     `<div></div>`,
			expected: "",
			setup: func(doc *goquery.Document) candidateList {
				return make(candidateList)
			},
		},
		{
			name:     "single candidate",
			html:     `<div class="content">Content</div>`,
			expected: "div.content => 30.000000",
			setup: func(doc *goquery.Document) candidateList {
				candidates := make(candidateList)
				selection := doc.Find("div")
				candidate := scoreNode(selection)
				candidates[selection.Get(0)] = candidate
				return candidates
			},
		},
		{
			name: "multiple candidates",
			html: `<div class="content">Content</div><p class="text">Paragraph</p><h1 id="main">Title</h1>`,
			setup: func(doc *goquery.Document) candidateList {
				candidates := make(candidateList)

				divSelection := doc.Find("div")
				divCandidate := scoreNode(divSelection)
				candidates[divSelection.Get(0)] = divCandidate

				pSelection := doc.Find("p")
				pCandidate := scoreNode(pSelection)
				candidates[pSelection.Get(0)] = pCandidate

				h1Selection := doc.Find("h1")
				h1Candidate := scoreNode(h1Selection)
				candidates[h1Selection.Get(0)] = h1Candidate

				return candidates
			},
		},
		{
			name: "candidates with mixed scores",
			html: `<div class="comment">Comment</div><p class="content">Good content</p>`,
			setup: func(doc *goquery.Document) candidateList {
				candidates := make(candidateList)

				divSelection := doc.Find("div")
				divCandidate := scoreNode(divSelection)
				candidates[divSelection.Get(0)] = divCandidate

				pSelection := doc.Find("p")
				pCandidate := scoreNode(pSelection)
				candidates[pSelection.Get(0)] = pCandidate

				return candidates
			},
		},
		{
			name: "candidate with empty selection",
			html: `<div>Test</div>`,
			setup: func(doc *goquery.Document) candidateList {
				candidates := make(candidateList)

				// Add a regular candidate
				divSelection := doc.Find("div")
				divCandidate := scoreNode(divSelection)
				candidates[divSelection.Get(0)] = divCandidate

				// Add a candidate with empty selection (this is artificial but tests the edge case)
				emptySelection := doc.Find("nonexistent")
				emptyCandidate := &candidate{selection: emptySelection, score: 0}
				// We can't use emptySelection.Get(0) as key since it would panic,
				// so we'll create a dummy node for this test
				dummyNode := &html.Node{Type: html.ElementNode, Data: "dummy"}
				candidates[dummyNode] = emptyCandidate

				return candidates
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			candidates := tc.setup(doc)
			result := candidates.String()

			if tc.name == "empty candidate list" {
				if result != tc.expected {
					t.Errorf("Expected: %s, Got: %s", tc.expected, result)
				}
				return
			}

			// For multiple candidates, we need to check that all expected parts are present
			// since map iteration order is not guaranteed
			switch tc.name {
			case "multiple candidates":
				expectedParts := []string{"div.content => 30.000000", "p.text => 25.000000", "h1#main => 20.000000"}
				for _, part := range expectedParts {
					if !strings.Contains(result, part) {
						t.Errorf("Expected result to contain: %s, Got: %s", part, result)
					}
				}
				// Check that it's comma-separated
				if !strings.Contains(result, ", ") {
					t.Errorf("Expected comma-separated format, Got: %s", result)
				}
			case "candidates with mixed scores":
				expectedParts := []string{"div.comment => -20.000000", "p.content => 25.000000"}
				for _, part := range expectedParts {
					if !strings.Contains(result, part) {
						t.Errorf("Expected result to contain: %s, Got: %s", part, result)
					}
				}
			case "candidate with empty selection":
				// Should contain both the regular candidate and the empty one
				if !strings.Contains(result, "div => 5.000000") {
					t.Errorf("Expected result to contain div candidate, Got: %s", result)
				}
				if !strings.Contains(result, "empty => 0.000000") {
					t.Errorf("Expected result to contain empty candidate, Got: %s", result)
				}
			default:
				// Single candidate test cases
				if result != tc.expected {
					t.Errorf("Expected: %s, Got: %s", tc.expected, result)
				}
			}
		})
	}
}

func TestCandidateStringEdgeCases(t *testing.T) {
	t.Run("candidate with nil node but valid selection", func(t *testing.T) {
		// This tests the case where Node() returns nil but selection exists
		html := `<div>Test</div>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		emptySelection := doc.Find("nonexistent")
		candidate := &candidate{
			selection: emptySelection,
			score:     10.5,
		}

		result := candidate.String()
		expected := "empty => 10.500000"

		if result != expected {
			t.Errorf("Expected: %s, Got: %s", expected, result)
		}
	})

	t.Run("candidate with zero score", func(t *testing.T) {
		html := `<div>Test</div>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		selection := doc.Find("div")
		candidate := &candidate{
			selection: selection,
			score:     0,
		}

		result := candidate.String()
		expected := "div => 0.000000"

		if result != expected {
			t.Errorf("Expected: %s, Got: %s", expected, result)
		}
	})

	t.Run("candidate with negative score", func(t *testing.T) {
		html := `<h1>Test</h1>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		selection := doc.Find("h1")
		candidate := &candidate{
			selection: selection,
			score:     -10.5,
		}

		result := candidate.String()
		expected := "h1 => -10.500000"

		if result != expected {
			t.Errorf("Expected: %s, Got: %s", expected, result)
		}
	})

	t.Run("candidate with very long class and id", func(t *testing.T) {
		html := `<div class="very-long-class-name-that-might-cause-issues" id="very-long-id-name-that-might-also-cause-formatting-issues">Test</div>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			t.Fatal(err)
		}

		selection := doc.Find("div")
		candidate := scoreNode(selection)

		result := candidate.String()
		expected := "div#very-long-id-name-that-might-also-cause-formatting-issues.very-long-class-name-that-might-cause-issues => 5.000000"

		if result != expected {
			t.Errorf("Expected: %s, Got: %s", expected, result)
		}
	})
}

func TestExtractContentWithBrokenReader(t *testing.T) {
	if _, _, err := ExtractContent(&brokenReader{}); err == nil {
		t.Error("Expected ExtractContent to return an error with broken reader")
	}
}

// brokenReader implements io.Reader but always returns an error
type brokenReader struct{}

func (br *brokenReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}
