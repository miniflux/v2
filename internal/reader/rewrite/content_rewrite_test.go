// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/v2/internal/reader/rewrite"

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
)

func TestParseRules(t *testing.T) {
	rulesText := `add_dynamic_image,replace("article/(.*).svg"|"article/$1.png"),remove(".spam, .ads:not(.keep)")`
	expected := []rule{
		{name: "add_dynamic_image"},
		{name: "replace", args: []string{"article/(.*).svg", "article/$1.png"}},
		{name: "remove", args: []string{".spam, .ads:not(.keep)"}},
	}

	actual := parseRules(rulesText)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(`Parsed rules do not match expected rules: got %v instead of %v`, actual, expected)
	}
}

func TestReplaceTextLinks(t *testing.T) {
	scenarios := map[string]string{
		`This is a link to example.org`:                                              `This is a link to example.org`,
		`This is a link to ftp://example.org`:                                        `This is a link to ftp://example.org`,
		`This is a link to www.example.org`:                                          `This is a link to www.example.org`,
		`This is a link to http://example.org`:                                       `This is a link to <a href="http://example.org">http://example.org</a>`,
		`This is a link to http://example.org, end of sentence.`:                     `This is a link to <a href="http://example.org">http://example.org</a>, end of sentence.`,
		`This is a link to https://example.org`:                                      `This is a link to <a href="https://example.org">https://example.org</a>`,
		`This is a link to https://www.example.org/path/to?q=s`:                      `This is a link to <a href="https://www.example.org/path/to?q=s">https://www.example.org/path/to?q=s</a>`,
		`This is a link to https://example.org/index#hash-tag, http://example.org/.`: `This is a link to <a href="https://example.org/index#hash-tag">https://example.org/index#hash-tag</a>, <a href="http://example.org/">http://example.org/</a>.`,
	}

	for input, expected := range scenarios {
		actual := replaceTextLinks(input)
		if actual != expected {
			t.Errorf(`Unexpected link replacement, got "%s" instead of "%s"`, actual, expected)
		}
	}
}

func TestRewriteWithNoMatchingRule(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `Some text.`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `Some text.`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteYoutubeVideoLink(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/1234" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `Video Description`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteYoutubeShortLink(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/shorts/1LUWKWZkPjo",
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/1LUWKWZkPjo" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/shorts/1LUWKWZkPjo",
		Title:   `A title`,
		Content: `Video Description`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteIncorrectYoutubeLink(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/some-page",
		Title:   `A title`,
		Content: `Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/some-page",
		Title:   `A title`,
		Content: `Video Description`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteYoutubeLinkAndCustomEmbedURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("YOUTUBE_EMBED_URL_OVERRIDE", "https://invidious.custom/embed/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://invidious.custom/embed/1234" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `Video Description`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteYoutubeVideoLinkUsingInvidious(t *testing.T) {
	config.Opts = config.NewConfigOptions()
	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://yewtu.be/embed/1234" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `Video Description`,
	}

	ApplyContentRewriteRules(testEntry, `add_youtube_video_using_invidious_player`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteYoutubeShortLinkUsingInvidious(t *testing.T) {
	config.Opts = config.NewConfigOptions()
	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/shorts/1LUWKWZkPjo",
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://yewtu.be/embed/1LUWKWZkPjo" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/shorts/1LUWKWZkPjo",
		Title:   `A title`,
		Content: `Video Description`,
	}

	ApplyContentRewriteRules(testEntry, `add_youtube_video_using_invidious_player`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestAddYoutubeVideoFromId(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	scenarios := map[string]string{
		// Test with single YouTube ID
		`Some content with youtube ID <script type="text/javascript" data-reactid="6">window.__APOLLO_STATE__ = {youtube_id: "9uASADiYe_8"}</script>`: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/9uASADiYe_8" allowfullscreen></iframe><br>Some content with youtube ID <script type="text/javascript" data-reactid="6">window.__APOLLO_STATE__ = {youtube_id: "9uASADiYe_8"}</script>`,

		// Test with multiple YouTube IDs
		`Content with youtube_id: "dQw4w9WgXcQ" and youtube_id: "jNQXAC9IVRw"`: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br><iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/jNQXAC9IVRw" allowfullscreen></iframe><br>Content with youtube_id: "dQw4w9WgXcQ" and youtube_id: "jNQXAC9IVRw"`,

		// Test with YouTube ID using equals sign
		`Some content with youtube_id = "dQw4w9WgXcQ"`: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>Some content with youtube_id = "dQw4w9WgXcQ"`,

		// Test with spaces around delimiters
		`Some content with youtube_id : "dQw4w9WgXcQ"`: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>Some content with youtube_id : "dQw4w9WgXcQ"`,

		// Test with YouTube ID without quotes (regex requires quotes)
		`Some content with youtube_id: dQw4w9WgXcQ and more`: `Some content with youtube_id: dQw4w9WgXcQ and more`,

		// Test with no YouTube ID
		`Some regular content without any video ID`: `Some regular content without any video ID`,

		// Test with invalid YouTube ID (wrong length)
		`Some content with youtube_id: "invalid"`: `Some content with youtube_id: "invalid"`,

		// Test with empty content
		``: ``,
	}

	for input, expected := range scenarios {
		actual := addYoutubeVideoFromId(input)
		if actual != expected {
			t.Errorf(`addYoutubeVideoFromId test failed for input "%s"`, input)
			t.Errorf(`Expected: "%s"`, expected)
			t.Errorf(`Actual: "%s"`, actual)
		}
	}
}

func TestAddYoutubeVideoFromIdWithCustomEmbedURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("YOUTUBE_EMBED_URL_OVERRIDE", "https://invidious.custom/embed/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `Some content with youtube_id: "dQw4w9WgXcQ"`
	expected := `<iframe width="650" height="350" frameborder="0" src="https://invidious.custom/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>Some content with youtube_id: "dQw4w9WgXcQ"`

	actual := addYoutubeVideoFromId(input)
	if actual != expected {
		t.Errorf(`addYoutubeVideoFromId with custom embed URL failed`)
		t.Errorf(`Expected: "%s"`, expected)
		t.Errorf(`Actual: "%s"`, actual)
	}
}

func TestAddInvidiousVideo(t *testing.T) {
	scenarios := map[string][]string{
		// Test with various Invidious instances
		"https://invidious.io/watch?v=dQw4w9WgXcQ": {
			"Some video content",
			`<iframe width="650" height="350" frameborder="0" src="https://invidious.io/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>Some video content`,
		},
		"https://yewtu.be/watch?v=jNQXAC9IVRw": {
			"Another video description",
			`<iframe width="650" height="350" frameborder="0" src="https://yewtu.be/embed/jNQXAC9IVRw" allowfullscreen></iframe><br>Another video description`,
		},
		"http://invidious.snopyta.org/watch?v=dQw4w9WgXcQ": {
			"HTTP instance test",
			`<iframe width="650" height="350" frameborder="0" src="https://invidious.snopyta.org/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>HTTP instance test`,
		},
		"https://youtube.com/watch?v=dQw4w9WgXcQ": {
			"YouTube URL (also matches regex)",
			`<iframe width="650" height="350" frameborder="0" src="https://youtube.com/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>YouTube URL (also matches regex)`,
		},
		"https://example.org/watch?v=dQw4w9WgXcQ": {
			"Any domain with watch pattern",
			`<iframe width="650" height="350" frameborder="0" src="https://example.org/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>Any domain with watch pattern`,
		},

		// Test with query parameters
		"https://invidious.io/watch?v=dQw4w9WgXcQ&t=30s": {
			"Video with timestamp",
			`<iframe width="650" height="350" frameborder="0" src="https://invidious.io/embed/dQw4w9WgXcQ?t=30s" allowfullscreen></iframe><br>Video with timestamp`,
		},

		// Test with more complex query parameters
		"https://invidious.io/watch?v=dQw4w9WgXcQ&t=30s&autoplay=1": {
			"Video with multiple parameters",
			`<iframe width="650" height="350" frameborder="0" src="https://invidious.io/embed/dQw4w9WgXcQ?autoplay=1&t=30s" allowfullscreen></iframe><br>Video with multiple parameters`,
		},

		// Test with non-matching URLs (should return content unchanged)
		"https://invidious.io/": {
			"Invidious homepage",
			"Invidious homepage",
		},
		"https://invidious.io/some-other-page": {
			"Other page",
			"Other page",
		},
		"https://invidious.io/search?q=test": {
			"Search page",
			"Search page",
		},

		// Test with empty content
		"https://empty.invidious.io/watch?v=dQw4w9WgXcQ": {
			"",
			`<iframe width="650" height="350" frameborder="0" src="https://empty.invidious.io/embed/dQw4w9WgXcQ" allowfullscreen></iframe><br>`,
		},
	}

	for entryURL, testData := range scenarios {
		entryContent := testData[0]
		expected := testData[1]

		actual := addInvidiousVideo(entryURL, entryContent)
		if actual != expected {
			t.Errorf(`addInvidiousVideo test failed for URL "%s" and content "%s"`, entryURL, entryContent)
			t.Errorf(`Expected: "%s"`, expected)
			t.Errorf(`Actual: "%s"`, actual)
		}
	}
}

func TestRewriteWithInexistingCustomRule(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `Video Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.youtube.com/watch?v=1234",
		Title:   `A title`,
		Content: `Video Description`,
	}
	ApplyContentRewriteRules(testEntry, `some rule`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLink(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you."/><figcaption><p>Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you.</p></figcaption></figure>`,
	}
	testEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkHtmlInjection(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="&lt;foo&gt;"/><figcaption><p>&lt;foo&gt;</p></figcaption></figure>`,
	}
	testEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="<foo>" alt="<foo>" />`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkAndImageNoTitle(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	testEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkAndNoImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `test`,
	}
	testEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `test`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdAndNoImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `test`,
	}
	testEntry := &model.Entry{
		URL:     "https://xkcd.com/1912/",
		Title:   `A title`,
		Content: `test`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteMailtoLink(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://www.qwantz.com/",
		Title:   `A title`,
		Content: `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact [blah blah]</a>`,
	}
	testEntry := &model.Entry{
		URL:     "https://www.qwantz.com/",
		Title:   `A title`,
		Content: `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact</a>`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithPDFLink(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/document.pdf",
		Title:   `A title`,
		Content: `<a href="https://example.org/document.pdf">PDF</a><br>test`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/document.pdf",
		Title:   `A title`,
		Content: `test`,
	}
	ApplyContentRewriteRules(testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithNoLazyImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"><noscript><p>Some text</p></noscript>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"><noscript><p>Some text</p></noscript>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" data-url="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"/></noscript>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="" data-url="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyDivImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"/></noscript>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div data-url="https://example.org/image.jpg" alt="Image"></div><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithUnknownLazyNoScriptImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"/><img src="https://example.org/fallback.jpg" alt="Fallback"/>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazySrcset(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithImageAndLazySrcset(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="meow" srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="meow" srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithNoLazyIframe(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe src="https://example.org/embed" allowfullscreen></iframe>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe src="https://example.org/embed" allowfullscreen></iframe>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_iframe")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyIframe(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe data-src="https://example.org/embed" allowfullscreen="" src="https://example.org/embed"></iframe>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe data-src="https://example.org/embed" allowfullscreen></iframe>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_iframe")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyIframeAndSrc(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe src="https://example.org/embed" data-src="https://example.org/embed" allowfullscreen=""></iframe>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<iframe src="about:blank" data-src="https://example.org/embed" allowfullscreen></iframe>`,
	}
	ApplyContentRewriteRules(testEntry, "add_dynamic_iframe")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestNewLineRewriteRule(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `A<br>B<br>C`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: "A\nB\nC",
	}
	ApplyContentRewriteRules(testEntry, "nl2br")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestConvertTextLinkRewriteRule(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `Test: <a href="http://example.org/a/b">http://example.org/a/b</a>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `Test: http://example.org/a/b`,
	}
	ApplyContentRewriteRules(testEntry, "convert_text_link")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestMediumImage(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img alt="Image for post" class="t u v if aj" src="https://miro.medium.com/max/2560/1*ephLSqSzQYLvb7faDwzRbw.jpeg" width="1280" height="720" srcset="https://miro.medium.com/max/552/1*ephLSqSzQYLvb7faDwzRbw.jpeg 276w, https://miro.medium.com/max/1104/1*ephLSqSzQYLvb7faDwzRbw.jpeg 552w, https://miro.medium.com/max/1280/1*ephLSqSzQYLvb7faDwzRbw.jpeg 640w, https://miro.medium.com/max/1400/1*ephLSqSzQYLvb7faDwzRbw.jpeg 700w" sizes="700px"/>`,
	}
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `
		<figure class="ht hu hv hw hx hy cy cz paragraph-image">
			<div class="hz ia ib ic aj">
				<div class="cy cz hs">
					<div class="ii s ib ij">
						<div class="ik il s">
							<div class="id ie t u v if aj bk ig ih">
								<img alt="Image for post" class="t u v if aj im in io" src="https://miro.medium.com/max/60/1*ephLSqSzQYLvb7faDwzRbw.jpeg?q=20" width="1280" height="720"/>
							</div>
							<img alt="Image for post" class="id ie t u v if aj c" width="1280" height="720"/>
							<noscript>
								<img alt="Image for post" class="t u v if aj" src="https://miro.medium.com/max/2560/1*ephLSqSzQYLvb7faDwzRbw.jpeg" width="1280" height="720" srcSet="https://miro.medium.com/max/552/1*ephLSqSzQYLvb7faDwzRbw.jpeg 276w, https://miro.medium.com/max/1104/1*ephLSqSzQYLvb7faDwzRbw.jpeg 552w, https://miro.medium.com/max/1280/1*ephLSqSzQYLvb7faDwzRbw.jpeg 640w, https://miro.medium.com/max/1400/1*ephLSqSzQYLvb7faDwzRbw.jpeg 700w" sizes="700px"/>
							</noscript>
						</div>
					</div>
				</div>
			</div>
		</figure>
		`,
	}
	ApplyContentRewriteRules(testEntry, "fix_medium_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteNoScriptImageWithoutNoScriptTag(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."/><figcaption>MDN Logo</figcaption></figure>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><figcaption>MDN Logo</figcaption></figure>`,
	}
	ApplyContentRewriteRules(testEntry, "use_noscript_figure_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteNoScriptImageWithNoScriptTag(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<figure><img src="http://example.org/logo.svg"/><figcaption>MDN Logo</figcaption></figure>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><noscript><img src="http://example.org/logo.svg"></noscript><figcaption>MDN Logo</figcaption></figure>`,
	}
	ApplyContentRewriteRules(testEntry, "use_noscript_figure_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteReplaceCustom(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.png">`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.svg">`,
	}
	ApplyContentRewriteRules(testEntry, `replace("article/(.*).svg"|"article/$1.png")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteReplaceTitleCustom(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `Ouch, a thistle`,
		Content: `The replace_title rewrite rule should not modify the content.`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `The replace_title rewrite rule should not modify the content.`,
	}
	ApplyContentRewriteRules(testEntry, `replace_title("(?i)^a\\s*ti"|"Ouch, a this")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteRemoveCustom(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum <span class="ads keep">Super important info</span></div>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum <span class="spam">I dont want to see this</span><span class="ads keep">Super important info</span></div>`,
	}
	ApplyContentRewriteRules(testEntry, `remove(".spam, .ads:not(.keep)")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteRemoveQuotedSelector(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum</div>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum<img alt="LINUX KERNEL" src="/assets/categories/linuxkernel.webp" width="100" height="100"></div>`,
	}
	ApplyContentRewriteRules(testEntry, `remove("img[src^='/assets/categories/']")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteAddCastopodEpisode(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://podcast.demo/@demo/episodes/test",
		Title:   `A title`,
		Content: `<iframe width="650" frameborder="0" src="https://podcast.demo/@demo/episodes/test/embed/light"></iframe><br>Episode Description`,
	}
	testEntry := &model.Entry{
		URL:     "https://podcast.demo/@demo/episodes/test",
		Title:   `A title`,
		Content: `Episode Description`,
	}
	ApplyContentRewriteRules(testEntry, `add_castopod_episode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64Decode(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `This is some base64 encoded content`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=`,
	}
	ApplyContentRewriteRules(testEntry, `base64_decode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64DecodeInHTML(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum not valid base64<span class="base64">This is some base64 encoded content</span></div>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum not valid base64<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`,
	}
	ApplyContentRewriteRules(testEntry, `base64_decode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64DecodeArgs(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum<span class="base64">This is some base64 encoded content</span></div>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<div>Lorem Ipsum<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`,
	}
	ApplyContentRewriteRules(testEntry, `base64_decode(".base64")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteRemoveTables(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<p>Test</p><p>Hello World!</p><p>Test</p>`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<table class="container"><tbody><tr><td><p>Test</p><table class="row"><tbody><tr><td><p>Hello World!</p></td><td><p>Test</p></td></tr></tbody></table></td></tr></tbody></table>`,
	}
	ApplyContentRewriteRules(testEntry, `remove_tables`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRemoveClickbait(t *testing.T) {
	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `This Is Amazing`,
		Content: `Some description`,
	}
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `THIS IS AMAZING`,
		Content: `Some description`,
	}
	ApplyContentRewriteRules(testEntry, `remove_clickbait`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestAddHackerNewsLinksUsingHack(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<p>Article URL: <a href="https://example.org/url">https://example.org/article</a></p>
		<p>Comments URL: <a href="https://news.ycombinator.com/item?id=37620043">https://news.ycombinator.com/item?id=37620043</a></p>
		<p>Points: 23</p>
		<p># Comments: 38</p>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<p>Article URL: <a href="https://example.org/url">https://example.org/article</a></p>
		<p>Comments URL: <a href="https://news.ycombinator.com/item?id=37620043">https://news.ycombinator.com/item?id=37620043</a> <a href="hack://item?id=37620043">Open with HACK</a></p>
		<p>Points: 23</p>
		<p># Comments: 38</p>`,
	}
	ApplyContentRewriteRules(testEntry, `add_hn_links_using_hack`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestAddHackerNewsLinksUsingOpener(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<p>Article URL: <a href="https://example.org/url">https://example.org/article</a></p>
		<p>Comments URL: <a href="https://news.ycombinator.com/item?id=37620043">https://news.ycombinator.com/item?id=37620043</a></p>
		<p>Points: 23</p>
		<p># Comments: 38</p>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<p>Article URL: <a href="https://example.org/url">https://example.org/article</a></p>
		<p>Comments URL: <a href="https://news.ycombinator.com/item?id=37620043">https://news.ycombinator.com/item?id=37620043</a> <a href="opener://x-callback-url/show-options?url=https%3A%2F%2Fnews.ycombinator.com%2Fitem%3Fid%3D37620043">Open with Opener</a></p>
		<p>Points: 23</p>
		<p># Comments: 38</p>`,
	}
	ApplyContentRewriteRules(testEntry, `add_hn_links_using_opener`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestAddImageTitle(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `
		<img src="pif" title="pouf">
		<img src="pif" title="pouf" alt='"onerror=alert(1) a="'>
		<img src="pif" title="pouf" alt='&quot;onerror=alert(1) a=&quot'>
		<img src="pif" title="pouf" alt=';&amp;quot;onerror=alert(1) a=;&amp;quot;'>
		<img src="pif" alt="pouf" title='"onerror=alert(1) a="'>
		<img src="pif" alt="pouf" title='&quot;onerror=alert(1) a=&quot'>
		<img src="pif" alt="pouf" title=';&amp;quot;onerror=alert(1) a=;&amp;quot;'>
		`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure><img src="pif" alt=""/><figcaption><p>pouf</p></figcaption></figure>
		<figure><img src="pif" alt="" onerror="alert(1)" a=""/><figcaption><p>pouf</p></figcaption></figure>
		<figure><img src="pif" alt="" onerror="alert(1)" a=""/><figcaption><p>pouf</p></figcaption></figure>
		<figure><img src="pif" alt=";&#34;onerror=alert(1) a=;&#34;"/><figcaption><p>pouf</p></figcaption></figure>
		<figure><img src="pif" alt="pouf"/><figcaption><p>&#34;onerror=alert(1) a=&#34;</p></figcaption></figure>
		<figure><img src="pif" alt="pouf"/><figcaption><p>&#34;onerror=alert(1) a=&#34;</p></figcaption></figure>
		<figure><img src="pif" alt="pouf"/><figcaption><p>;&amp;quot;onerror=alert(1) a=;&amp;quot;</p></figcaption></figure>
		`,
	}
	ApplyContentRewriteRules(testEntry, `add_image_title`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCard(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<a href="https://example.org/article">Example Article - Example</a>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardNoCard(t *testing.T) {
	testEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<a href="https://example.org/article">Example Article - Example</a>`,
	}

	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<a href="https://example.org/article">Example Article - Example</a>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardInvalidCard(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a href="https://example.org/article">This card does not have the required fields</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a href="https://example.org/article">This card does not have the required fields</a>
		</figure>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardMissingAuthor(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<a href="https://example.org/article">Example Article</a>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardDuplicatedAuthor(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article - Example</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<a href="https://example.org/article">Example Article - Example</a>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardMultiple(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article1">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article 1 - Example</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>
		<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article2">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article 2 - Example</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:     "https://example.org/article",
		Title:   `A title`,
		Content: `<ul><li><a href="https://example.org/article1">Example Article 1 - Example</a></li><li><a href="https://example.org/article2">Example Article 2 - Example</a></li></ul>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestFixGhostCardMultipleSplit(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article1">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article 1 - Example</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>
		<p>This separates the two cards</p>
		<figure class="kg-card kg-bookmark-card">
			<a class="kg-bookmark-container" href="https://example.org/article2">
				<div class="kg-bookmark-content">
					<div class="kg-bookmark-title">Example Article 2 - Example</div>
					<div class="kg-bookmark-description">Lorem ipsum odor amet, consectetuer adipiscing elit. Pretium magnis luctus ligula conubia quam, donec orci vehicula efficitur...</div>
					<div class="kg-bookmark-metadata">
						<img class="kg-bookmark-icon" src="https://example.org/favicon.ico" alt="">
						<span class="kg-bookmark-author">Example</span>
						<span class="kg-bookmark-publisher">Test Author</span>
					</div>
				</div>
				<div class="kg-bookmark-thumbnail">
					<img src="https://example.org/article-image.jpg" alt="" onerror="this.style.display = 'none'">
				</div>
			</a>
		</figure>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `A title`,
		Content: `<a href="https://example.org/article1">Example Article 1 - Example</a>
		<p>This separates the two cards</p>
		<a href="https://example.org/article2">Example Article 2 - Example</a>`,
	}
	ApplyContentRewriteRules(testEntry, `fix_ghost_cards`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestStripImageQueryParams(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `News Article Title`,
		Content: `
		<article>
			<p>Article content with images having query parameters:</p>
			<img src="https://example.org/images/image1.jpg?width=200&height=113&q=80&blur=90" alt="Image with params">
			<img src="https://example.org/images/image2.jpg?width=800&height=600&q=85" alt="Another image with params">

			<p>More images with various query parameters:</p>
			<img src="https://example.org/image123.jpg?blur=50&size=small&format=webp" alt="Complex query params">
			<img src="https://example.org/image123.jpg?size=large&quality=95&cache=123" alt="Different params">

			<p>Image without query parameters:</p>
			<img src="https://example.org/single-image.jpg" alt="Clean image">

			<p>Images with various other params:</p>
			<img src="https://example.org/normal1.jpg?width=300&format=jpg" alt="Normal 1">
			<img src="https://example.org/normal1.jpg?width=600&quality=high" alt="Normal 2">
		</article>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `News Article Title`,
		Content: `<article>
			<p>Article content with images having query parameters:</p>
			<img src="https://example.org/images/image1.jpg" alt="Image with params"/>
			<img src="https://example.org/images/image2.jpg?width=800&amp;height=600&amp;q=85" alt="Another image with params"/>

			<p>More images with various query parameters:</p>
			<img src="https://example.org/image123.jpg" alt="Complex query params"/>
			<img src="https://example.org/image123.jpg?size=large&amp;quality=95&amp;cache=123" alt="Different params"/>

			<p>Image without query parameters:</p>
			<img src="https://example.org/single-image.jpg" alt="Clean image"/>

			<p>Images with various other params:</p>
			<img src="https://example.org/normal1.jpg?width=300&amp;format=jpg" alt="Normal 1"/>
			<img src="https://example.org/normal1.jpg?width=600&amp;quality=high" alt="Normal 2"/>
		</article>`,
	}
	ApplyContentRewriteRules(testEntry, `remove_img_blur_params`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestStripImageQueryParamsNoChanges(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Article Without Images`,
		Content: `<p>No images here:</p>
		<div>Just some text content</div>
		<a href="https://example.org">A link</a>`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Article Without Images`,
		Content: `<p>No images here:</p>
		<div>Just some text content</div>
		<a href="https://example.org">A link</a>`,
	}
	ApplyContentRewriteRules(testEntry, `remove_img_blur_params`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestStripImageQueryParamsEdgeCases(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Edge Cases`,
		Content: `
		<p>Edge cases for image query parameter stripping:</p>

		<!-- Various query parameters -->
		<img src="https://example.org/image1.jpg?blur=80&width=300" alt="Multiple params">

		<!-- Complex query parameters -->
		<img src="https://example.org/image2.jpg?BLUR=60&format=webp&cache=123" alt="Complex params">
		<img src="https://example.org/image3.jpg?quality=high&version=2" alt="Other params">

		<!-- Query params in middle of string -->
		<img src="https://example.org/image4.jpg?size=large&blur=30&format=webp&quality=90" alt="Middle params">

		<!-- Image without query params -->
		<img src="https://example.org/clean.jpg" alt="Clean image">
		`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Edge Cases`,
		Content: `<p>Edge cases for image query parameter stripping:</p>

		<!-- Various query parameters -->
		<img src="https://example.org/image1.jpg" alt="Multiple params"/>

		<!-- Complex query parameters -->
		<img src="https://example.org/image2.jpg?BLUR=60&amp;format=webp&amp;cache=123" alt="Complex params"/>
		<img src="https://example.org/image3.jpg?quality=high&amp;version=2" alt="Other params"/>

		<!-- Query params in middle of string -->
		<img src="https://example.org/image4.jpg" alt="Middle params"/>

		<!-- Image without query params -->
		<img src="https://example.org/clean.jpg" alt="Clean image"/>
		`,
	}
	ApplyContentRewriteRules(testEntry, `remove_img_blur_params`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestStripImageQueryParamsSimple(t *testing.T) {
	testEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Simple Test`,
		Content: `
		<p>Testing query parameter stripping:</p>

		<!-- Images with various query parameters -->
		<img src="https://example.org/test1.jpg?blur=0&width=300" alt="With blur zero">
		<img src="https://example.org/test2.jpg?blur=50&width=300&format=webp" alt="With blur fifty">
		<img src="https://example.org/test3.jpg?width=800&quality=high" alt="No blur param">
		<img src="https://example.org/test4.jpg" alt="No params at all">
		`,
	}

	controlEntry := &model.Entry{
		URL:   "https://example.org/article",
		Title: `Simple Test`,
		Content: `<p>Testing query parameter stripping:</p>

		<!-- Images with various query parameters -->
		<img src="https://example.org/test1.jpg?blur=0&amp;width=300" alt="With blur zero"/>
		<img src="https://example.org/test2.jpg" alt="With blur fifty"/>
		<img src="https://example.org/test3.jpg?width=800&amp;quality=high" alt="No blur param"/>
		<img src="https://example.org/test4.jpg" alt="No params at all"/>
		`,
	}
	ApplyContentRewriteRules(testEntry, `remove_img_blur_params`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}
