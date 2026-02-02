// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package sanitizer // import "miniflux.app/v2/internal/reader/sanitizer"

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"miniflux.app/v2/internal/config"
)

func sanitizeHTMLWithDefaultOptions(baseURL, rawHTML string) string {
	return SanitizeHTML(baseURL, rawHTML, &SanitizerOptions{
		OpenLinksInNewTab: true,
	})
}

func BenchmarkSanitize(b *testing.B) {
	var testCases = map[string][]string{
		"miniflux_github.html":    {"https://github.com/miniflux/v2", ""},
		"miniflux_wikipedia.html": {"https://fr.wikipedia.org/wiki/Miniflux", ""},
	}
	for filename := range testCases {
		data, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			b.Fatalf(`Unable to read file %q: %v`, filename, err)
		}
		testCases[filename][1] = string(data)
	}
	for b.Loop() {
		for _, v := range testCases {
			sanitizeHTMLWithDefaultOptions(v[0], v[1])
		}
	}
}

func FuzzSanitizer(f *testing.F) {
	f.Fuzz(func(t *testing.T, orig string) {
		tok := html.NewTokenizer(strings.NewReader(orig))
		i := 0
		for tok.Next() != html.ErrorToken {
			i++
		}

		out := sanitizeHTMLWithDefaultOptions("", orig)

		tok = html.NewTokenizer(strings.NewReader(out))
		j := 0
		for tok.Next() != html.ErrorToken {
			j++
		}

		if j > i {
			t.Errorf("Got more html tokens in the sanitized html.")
		}
	})
}

func TestValidInput(t *testing.T) {
	input := `<p>This is a <strong>text</strong> with an image: <img src="http://example.org/" alt="Test" loading="lazy">.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestImgSanitization(t *testing.T) {
	baseURL := "http://example.org/"
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "width-and-height-attributes",
			input:    `<img src="https://example.org/image.png" width="10" height="20">`,
			expected: `<img src="https://example.org/image.png" width="10" height="20" loading="lazy">`,
		},
		{
			name:     "invalid-width-and-height-attributes",
			input:    `<img src="https://example.org/image.png" width="10px" height="20px">`,
			expected: `<img src="https://example.org/image.png" loading="lazy">`,
		},
		{
			name:     "invalid-width-attribute",
			input:    `<img src="https://example.org/image.png" width="10px" height="20">`,
			expected: `<img src="https://example.org/image.png" height="20" loading="lazy">`,
		},
		{
			name:     "empty-width-and-height-attributes",
			input:    `<img src="https://example.org/image.png" width="" height="">`,
			expected: `<img src="https://example.org/image.png" loading="lazy">`,
		},
		{
			name:     "invalid-height-attribute",
			input:    `<img src="https://example.org/image.png" width="10" height="20px">`,
			expected: `<img src="https://example.org/image.png" width="10" loading="lazy">`,
		},
		{
			name:     "negative-width-attribute",
			input:    `<img src="https://example.org/image.png" width="-10" height="20">`,
			expected: `<img src="https://example.org/image.png" height="20" loading="lazy">`,
		},
		{
			name:     "negative-height-attribute",
			input:    `<img src="https://example.org/image.png" width="10" height="-20">`,
			expected: `<img src="https://example.org/image.png" width="10" loading="lazy">`,
		},
		{
			name:     "text-data-url",
			input:    `<img src="data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==" alt="Example">`,
			expected: ``,
		},
		{
			name:     "image-data-url",
			input:    `<img src="data:image/gif;base64,test" alt="Example">`,
			expected: `<img src="data:image/gif;base64,test" alt="Example" loading="lazy">`,
		},
		{
			name:     "srcset-attribute",
			input:    `<img srcset="example-320w.jpg, example-480w.jpg 1.5x,   example-640w.jpg 2x, example-640w.jpg 640w" src="example-640w.jpg" alt="Example">`,
			expected: `<img srcset="http://example.org/example-320w.jpg, http://example.org/example-480w.jpg 1.5x, http://example.org/example-640w.jpg 2x, http://example.org/example-640w.jpg 640w" src="http://example.org/example-640w.jpg" alt="Example" loading="lazy">`,
		},
		{
			name:     "srcset-attribute-without-src",
			input:    `<img srcset="example-320w.jpg, example-480w.jpg 1.5x,   example-640w.jpg 2x, example-640w.jpg 640w" alt="Example">`,
			expected: `<img srcset="http://example.org/example-320w.jpg, http://example.org/example-480w.jpg 1.5x, http://example.org/example-640w.jpg 2x, http://example.org/example-640w.jpg 640w" alt="Example" loading="lazy">`,
		},
		{
			name:     "srcset-attribute-with-blocked-candidate",
			input:    `<img srcset="https://stats.wordpress.com/tracker.png 1x, /example-640w.jpg 2x" src="/example-640w.jpg" alt="Example">`,
			expected: `<img srcset="http://example.org/example-640w.jpg 2x" src="http://example.org/example-640w.jpg" alt="Example" loading="lazy">`,
		},
		{
			name:     "srcset-attribute-all-candidates-invalid",
			input:    `<img srcset="javascript:alert(1) 1x, data:text/plain;base64,SGVsbG8= 2x" alt="Example">`,
			expected: ``,
		},
		{
			name:     "fetchpriority-high",
			input:    `<img src="https://example.org/image.png" fetchpriority="high">`,
			expected: `<img src="https://example.org/image.png" fetchpriority="high" loading="lazy">`,
		},
		{
			name:     "fetchpriority-low",
			input:    `<img src="https://example.org/image.png" fetchpriority="low">`,
			expected: `<img src="https://example.org/image.png" fetchpriority="low" loading="lazy">`,
		},
		{
			name:     "fetchpriority-auto",
			input:    `<img src="https://example.org/image.png" fetchpriority="auto">`,
			expected: `<img src="https://example.org/image.png" fetchpriority="auto" loading="lazy">`,
		},
		{
			name:     "fetchpriority-invalid",
			input:    `<img src="https://example.org/image.png" fetchpriority="invalid">`,
			expected: `<img src="https://example.org/image.png" loading="lazy">`,
		},
		{
			name:     "decoding-sync",
			input:    `<img src="https://example.org/image.png" decoding="sync">`,
			expected: `<img src="https://example.org/image.png" decoding="sync" loading="lazy">`,
		},
		{
			name:     "decoding-async",
			input:    `<img src="https://example.org/image.png" decoding="async">`,
			expected: `<img src="https://example.org/image.png" decoding="async" loading="lazy">`,
		},
		{
			name:     "decoding-auto",
			input:    `<img src="https://example.org/image.png" decoding="auto">`,
			expected: `<img src="https://example.org/image.png" decoding="auto" loading="lazy">`,
		},
		{
			name:     "decoding-invalid",
			input:    `<img src="https://example.org/image.png" decoding="invalid">`,
			expected: `<img src="https://example.org/image.png" loading="lazy">`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := sanitizeHTMLWithDefaultOptions(baseURL, tc.input)
			if output != tc.expected {
				t.Errorf(`Wrong output for input %q: expected %q, got %q`, tc.input, tc.expected, output)
			}
		})
	}
}

func TestNonImgWithFetchPriorityAttribute(t *testing.T) {
	input := `<p fetchpriority="high">Text</p>`
	expected := `<p>Text</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: expected %q, got %q`, expected, output)
	}
}

func TestNonImgWithDecodingAttribute(t *testing.T) {
	input := `<p decoding="async">Text</p>`
	expected := `<p>Text</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: expected %q, got %q`, expected, output)
	}
}

func TestMediumImgWithSrcset(t *testing.T) {
	input := `<img alt="Image for post" class="t u v ef aj" src="https://miro.medium.com/max/5460/1*aJ9JibWDqO81qMfNtqgqrw.jpeg" srcset="https://miro.medium.com/max/552/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 276w, https://miro.medium.com/max/1000/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 500w" sizes="500px" width="2730" height="3407">`
	expected := `<img alt="Image for post" src="https://miro.medium.com/max/5460/1*aJ9JibWDqO81qMfNtqgqrw.jpeg" srcset="https://miro.medium.com/max/552/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 276w, https://miro.medium.com/max/1000/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 500w" sizes="500px" width="2730" height="3407" loading="lazy">`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestSelfClosingTags(t *testing.T) {
	baseURL := "http://example.org/"
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "br",
			input:    `<p>Line<br>Break</p>`,
			expected: `<p>Line<br>Break</p>`,
		},
		{
			name:     "hr",
			input:    `<p>Before</p><hr><p>After</p>`,
			expected: `<p>Before</p><hr><p>After</p>`,
		},
		{
			name:     "img",
			input:    `<p>Image <img src="http://example.org/image.png" alt="Test"></p>`,
			expected: `<p>Image <img src="http://example.org/image.png" alt="Test" loading="lazy"></p>`,
		},
		{
			name:     "source",
			input:    `<picture><source src="http://example.org/video.mp4" type="video/mp4"></picture>`,
			expected: `<picture><source src="http://example.org/video.mp4" type="video/mp4"></picture>`,
		},
		{
			name:     "wbr",
			input:    `<p>soft<wbr>break</p>`,
			expected: `<p>soft<wbr>break</p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := sanitizeHTMLWithDefaultOptions(baseURL, tc.input)
			if output != tc.expected {
				t.Errorf(`Wrong output for input %q: expected %q, got %q`, tc.input, tc.expected, output)
			}
		})
	}
}

func TestTable(t *testing.T) {
	input := `<table><tr><th>A</th><th colspan="2">B</th></tr><tr><td>C</td><td>D</td><td>E</td></tr></table>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestRelativeURL(t *testing.T) {
	input := `This <a href="/test.html">link is relative</a> and this image: <img src="../folder/image.png">`
	expected := `This <a href="http://example.org/test.html" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">link is relative</a> and this image: <img src="http://example.org/folder/image.png" loading="lazy">`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestProtocolRelativeURL(t *testing.T) {
	input := `This <a href="//static.example.org/index.html">link is relative</a>.`
	expected := `This <a href="https://static.example.org/index.html" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">link is relative</a>.`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidTag(t *testing.T) {
	input := `<p>My invalid <z>tag</z>.</p>`
	expected := `<p>My invalid tag.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSourceSanitization(t *testing.T) {
	baseURL := "http://example.org/"
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "srcset-and-media",
			input:    `<picture><source media="(min-width: 800px)" srcset="elva-800w.jpg"></picture>`,
			expected: `<picture><source media="(min-width: 800px)" srcset="http://example.org/elva-800w.jpg"></picture>`,
		},
		{
			name:     "src-attribute",
			input:    `<picture><source src="video.mp4" type="video/mp4"></picture>`,
			expected: `<picture><source src="http://example.org/video.mp4" type="video/mp4"></picture>`,
		},
		{
			name:     "srcset-with-blocked-candidate",
			input:    `<picture><source srcset="https://stats.wordpress.com/tracker.png 1x, /elva-800w.jpg 2x"></picture>`,
			expected: `<picture><source srcset="http://example.org/elva-800w.jpg 2x"></picture>`,
		},
		{
			name:     "srcset-all-invalid",
			input:    `<picture><source srcset="javascript:alert(1) 1x, data:text/plain;base64,SGVsbG8= 2x"></picture>`,
			expected: `<picture></picture>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := sanitizeHTMLWithDefaultOptions(baseURL, tc.input)
			if output != tc.expected {
				t.Errorf(`Wrong output for input %q: expected %q, got %q`, tc.input, tc.expected, output)
			}
		})
	}
}

func TestVideoTag(t *testing.T) {
	input := `<p>My valid <video src="videofile.webm" autoplay poster="posterimage.jpg">fallback</video>.</p>`
	expected := `<p>My valid <video src="http://example.org/videofile.webm" poster="http://example.org/posterimage.jpg" controls>fallback</video>.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAudioAndSourceTag(t *testing.T) {
	input := `<p>My music <audio controls="controls"><source src="foo.wav" type="audio/wav"></audio>.</p>`
	expected := `<p>My music <audio controls><source src="http://example.org/foo.wav" type="audio/wav"></audio>.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestUnknownTag(t *testing.T) {
	input := `<p>My invalid <unknown>tag</unknown>.</p>`
	expected := `<p>My invalid tag.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidNestedTag(t *testing.T) {
	input := `<p>My invalid <z>tag with some <em>valid</em> tag</z>.</p>`
	expected := `<p>My invalid tag with some <em>valid</em> tag.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidIFrame(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	input := `<iframe src="http://example.org/"></iframe>`
	expected := ``
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSameDomainIFrame(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	input := `<iframe src="http://example.com/test"></iframe>`
	expected := ``
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestInvidiousIFrame(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	input := `<iframe src="https://yewtu.be/watch?v=video_id"></iframe>`
	expected := `<iframe src="https://yewtu.be/watch?v=video_id" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestCustomYoutubeEmbedURL(t *testing.T) {
	os.Setenv("YOUTUBE_EMBED_URL_OVERRIDE", "https://www.invidious.custom/embed/")

	defer os.Clearenv()
	var err error
	if config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables(); err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="https://www.invidious.custom/embed/1234"></iframe>`
	expected := `<iframe src="https://www.invidious.custom/embed/1234" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestIFrameWithChildElements(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	input := `<iframe src="https://www.youtube.com/"><p>test</p></iframe>`
	expected := `<iframe src="https://www.youtube.com/" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestIFrameWithReferrerPolicy(t *testing.T) {
	config.Opts = config.NewConfigOptions()

	input := `<iframe src="https://www.youtube.com/embed/test123" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestLinkWithTarget(t *testing.T) {
	input := `<p>This link is <a href="http://example.org/index.html">an anchor</a></p>`
	expected := `<p>This link is <a href="http://example.org/index.html" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">an anchor</a></p>`
	output := SanitizeHTML("http://example.org/", input, &SanitizerOptions{OpenLinksInNewTab: true})

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestLinkWithNoTarget(t *testing.T) {
	input := `<p>This link is <a href="http://example.org/index.html">an anchor</a></p>`
	expected := `<p>This link is <a href="http://example.org/index.html" rel="noopener noreferrer" referrerpolicy="no-referrer">an anchor</a></p>`
	output := SanitizeHTML("http://example.org/", input, &SanitizerOptions{OpenLinksInNewTab: false})

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAnchorLink(t *testing.T) {
	input := `<p>This link is <a href="#some-anchor">an anchor</a></p>`
	expected := `<p>This link is <a href="#some-anchor">an anchor</a></p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidURLScheme(t *testing.T) {
	input := `<p>This link is <a src="file:///etc/passwd">not valid</a></p>`
	expected := `<p>This link is not valid</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestURISchemes(t *testing.T) {
	baseURL := "http://example.org/"
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "apt",
			input:    `<p>This link is <a href="apt:some-package?channel=test">valid</a></p>`,
			expected: `<p>This link is <a href="apt:some-package?channel=test" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "bitcoin",
			input:    `<p>This link is <a href="bitcoin:175tWpb8K1S7NmH4Zx6rewF9WQrcZv245W">valid</a></p>`,
			expected: `<p>This link is <a href="bitcoin:175tWpb8K1S7NmH4Zx6rewF9WQrcZv245W" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "callto",
			input:    `<p>This link is <a href="callto:12345679">valid</a></p>`,
			expected: `<p>This link is <a href="callto:12345679" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "feed-double-slash",
			input:    `<p>This link is <a href="feed://example.com/rss.xml">valid</a></p>`,
			expected: `<p>This link is <a href="feed://example.com/rss.xml" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "feed-https",
			input:    `<p>This link is <a href="feed:https://example.com/rss.xml">valid</a></p>`,
			expected: `<p>This link is <a href="feed:https://example.com/rss.xml" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "geo",
			input:    `<p>This link is <a href="geo:13.4125,103.8667">valid</a></p>`,
			expected: `<p>This link is <a href="geo:13.4125,103.8667" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "itms",
			input:    `<p>This link is <a href="itms://itunes.com/apps/my-app-name">valid</a></p>`,
			expected: `<p>This link is <a href="itms://itunes.com/apps/my-app-name" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "itms-apps",
			input:    `<p>This link is <a href="itms-apps://itunes.com/apps/my-app-name">valid</a></p>`,
			expected: `<p>This link is <a href="itms-apps://itunes.com/apps/my-app-name" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "magnet",
			input:    `<p>This link is <a href="magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&amp;xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7">valid</a></p>`,
			expected: `<p>This link is <a href="magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&amp;xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "mailto",
			input:    `<p>This link is <a href="mailto:jsmith@example.com?subject=A%20Test&amp;body=My%20idea%20is%3A%20%0A">valid</a></p>`,
			expected: `<p>This link is <a href="mailto:jsmith@example.com?subject=A%20Test&amp;body=My%20idea%20is%3A%20%0A" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "news-double-slash",
			input:    `<p>This link is <a href="news://news.server.example/*">valid</a></p>`,
			expected: `<p>This link is <a href="news://news.server.example/*" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "news-single-colon",
			input:    `<p>This link is <a href="news:example.group.this">valid</a></p>`,
			expected: `<p>This link is <a href="news:example.group.this" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "nntp",
			input:    `<p>This link is <a href="nntp://news.server.example/example.group.this">valid</a></p>`,
			expected: `<p>This link is <a href="nntp://news.server.example/example.group.this" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "rtmp",
			input:    `<p>This link is <a href="rtmp://mycompany.com/vod/mp4:mycoolvideo.mov">valid</a></p>`,
			expected: `<p>This link is <a href="rtmp://mycompany.com/vod/mp4:mycoolvideo.mov" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "sip",
			input:    `<p>This link is <a href="sip:+1-212-555-1212:1234@gateway.com;user=phone">valid</a></p>`,
			expected: `<p>This link is <a href="sip:+1-212-555-1212:1234@gateway.com;user=phone" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "sips",
			input:    `<p>This link is <a href="sips:alice@atlanta.com?subject=project%20x&amp;priority=urgent">valid</a></p>`,
			expected: `<p>This link is <a href="sips:alice@atlanta.com?subject=project%20x&amp;priority=urgent" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "skype",
			input:    `<p>This link is <a href="skype:echo123?call">valid</a></p>`,
			expected: `<p>This link is <a href="skype:echo123?call" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "spotify",
			input:    `<p>This link is <a href="spotify:track:2jCnn1QPQ3E8ExtLe6INsx">valid</a></p>`,
			expected: `<p>This link is <a href="spotify:track:2jCnn1QPQ3E8ExtLe6INsx" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "steam",
			input:    `<p>This link is <a href="steam://settings/account">valid</a></p>`,
			expected: `<p>This link is <a href="steam://settings/account" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "svn",
			input:    `<p>This link is <a href="svn://example.org">valid</a></p>`,
			expected: `<p>This link is <a href="svn://example.org" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "svn-ssh",
			input:    `<p>This link is <a href="svn+ssh://example.org">valid</a></p>`,
			expected: `<p>This link is <a href="svn+ssh://example.org" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "tel",
			input:    `<p>This link is <a href="tel:+1-201-555-0123">valid</a></p>`,
			expected: `<p>This link is <a href="tel:+1-201-555-0123" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "webcal",
			input:    `<p>This link is <a href="webcal://example.com/calendar.ics">valid</a></p>`,
			expected: `<p>This link is <a href="webcal://example.com/calendar.ics" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
		{
			name:     "xmpp",
			input:    `<p>This link is <a href="xmpp:user@host?subscribe&amp;type=subscribed">valid</a></p>`,
			expected: `<p>This link is <a href="xmpp:user@host?subscribe&amp;type=subscribed" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">valid</a></p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := sanitizeHTMLWithDefaultOptions(baseURL, tc.input)
			if tc.expected != output {
				t.Errorf(`Wrong output for input %q: expected %q, got %q`, tc.input, tc.expected, output)
			}
		})
	}
}

func TestBlacklistedLink(t *testing.T) {
	input := `<p>This image is not valid <img src="https://stats.wordpress.com/some-tracker"></p>`
	expected := `<p>This image is not valid </p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestLinkWithTrackers(t *testing.T) {
	input := `<p>This link has trackers <a href="https://example.com/page?utm_source=newsletter">Test</a></p>`
	expected := `<p>This link has trackers <a href="https://example.com/page" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">Test</a></p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestImageSrcWithTrackers(t *testing.T) {
	input := `<p>This image has trackers <img src="https://example.org/?id=123&utm_source=newsletter&utm_medium=email&fbclid=abc123"></p>`
	expected := `<p>This image has trackers <img src="https://example.org/?id=123" loading="lazy"></p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func Test1x1PixelTracker(t *testing.T) {
	input := `<p><img src="https://tracker1.example.org/" height="1" width="1"> and <img src="https://tracker2.example.org/" height="1" width="1"/></p>`
	expected := `<p> and </p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func Test0x0PixelTracker(t *testing.T) {
	input := `<p><img src="https://tracker1.example.org/" height="0" width="0"> and <img src="https://tracker2.example.org/" height="0" width="0"/></p>`
	expected := `<p> and </p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestXmlEntities(t *testing.T) {
	input := `<pre>echo "test" &gt; /etc/hosts</pre>`
	expected := `<pre>echo &#34;test&#34; &gt; /etc/hosts</pre>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestEspaceAttributes(t *testing.T) {
	input := `<td rowspan="<b>injection</b>">text</td>`
	expected := `text`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURL(t *testing.T) {
	os.Clearenv()

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="http://www.youtube.com/embed/test123?version=3&#038;rel=1&#038;fs=1&#038;autohide=2&#038;showsearch=0&#038;showinfo=1&#038;iv_load_policy=1&#038;wmode=transparent"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?version=3&amp;rel=1&amp;fs=1&amp;autohide=2&amp;showsearch=0&amp;showinfo=1&amp;iv_load_policy=1&amp;wmode=transparent" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURL(t *testing.T) {
	os.Clearenv()

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="https://www.youtube.com/embed/test123"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURLWithParameters(t *testing.T) {
	os.Clearenv()

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="https://www.youtube.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURLAlreadyReplaced(t *testing.T) {
	os.Clearenv()

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceProtocolRelativeYoutubeURL(t *testing.T) {
	os.Clearenv()

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="//www.youtube.com/embed/Bf2W84jrGqs" width="560" height="314" allowfullscreen="allowfullscreen"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/Bf2W84jrGqs" width="560" height="314" allowfullscreen="allowfullscreen" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURLWithCustomURL(t *testing.T) {
	defer os.Clearenv()
	os.Setenv("YOUTUBE_EMBED_URL_OVERRIDE", "https://invidious.custom/embed/")

	var err error
	config.Opts, err = config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<iframe src="https://www.youtube.com/embed/test123?version=3&#038;rel=1&#038;fs=1&#038;autohide=2&#038;showsearch=0&#038;showinfo=1&#038;iv_load_policy=1&#038;wmode=transparent"></iframe>`
	expected := `<iframe src="https://invidious.custom/embed/test123?version=3&amp;rel=1&amp;fs=1&amp;autohide=2&amp;showsearch=0&amp;showinfo=1&amp;iv_load_policy=1&amp;wmode=transparent" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy" referrerpolicy="strict-origin-when-cross-origin"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestVimeoIframeRewriteWithQueryString(t *testing.T) {
	input := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0"></iframe>`
	expected := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0&amp;dnt=1" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestVimeoIframeRewriteWithoutQueryString(t *testing.T) {
	input := `<iframe src="https://player.vimeo.com/video/123456"></iframe>`
	expected := `<iframe src="https://player.vimeo.com/video/123456?dnt=1" sandbox="allow-scripts allow-same-origin allow-popups allow-popups-to-escape-sandbox" loading="lazy"></iframe>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: %q != %q`, expected, output)
	}
}

func TestReplaceNoScript(t *testing.T) {
	input := `<p>Before paragraph.</p><noscript>Inside <code>noscript</code> tag with an image: <img src="http://example.org/" alt="Test" loading="lazy"></noscript><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceScript(t *testing.T) {
	input := `<p>Before paragraph.</p><script type="text/javascript">alert("1");</script><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceStyle(t *testing.T) {
	input := `<p>Before paragraph.</p><style>body { background-color: #ff0000; }</style><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestHiddenParagraph(t *testing.T) {
	input := `<p>Before paragraph.</p><p hidden>This should <em>not</em> appear in the <strong>output</strong></p><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAttributesAreStripped(t *testing.T) {
	input := `<p style="color: red;">Some text.<hr style="color: blue"/>Test.</p>`
	expected := `<p>Some text.</p><hr>Test.<p></p>`

	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)
	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestMathML(t *testing.T) {
	input := `<math xmlns="http://www.w3.org/1998/Math/MathML"><msup><mi>x</mi><mn>2</mn></msup></math>`
	expected := `<math xmlns="http://www.w3.org/1998/Math/MathML"><msup><mi>x</mi><mn>2</mn></msup></math>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidMathMLXMLNamespace(t *testing.T) {
	input := `<math xmlns="http://example.org"><msup><mi>x</mi><mn>2</mn></msup></math>`
	expected := `<math xmlns="http://www.w3.org/1998/Math/MathML"><msup><mi>x</mi><mn>2</mn></msup></math>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestBlockedResourcesSubstrings(t *testing.T) {
	input := `<p>Before paragraph.</p><img src="http://stats.wordpress.com/something.php" alt="Blocked Resource"><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>Before paragraph.</p><img src="http://twitter.com/share?text=This+is+google+a+search+engine&url=https%3A%2F%2Fwww.google.com" alt="Blocked Resource"><p>After paragraph.</p>`
	expected = `<p>Before paragraph.</p><p>After paragraph.</p>`
	output = sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>Before paragraph.</p><img src="http://www.facebook.com/sharer.php?u=https%3A%2F%2Fwww.google.com%[title]=This+Is%2C+Google+a+search+engine" alt="Blocked Resource"><p>After paragraph.</p>`
	expected = `<p>Before paragraph.</p><p>After paragraph.</p>`
	output = sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAttrLowerCase(t *testing.T) {
	baseURL := "http://example.org/"
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "href-and-hidden-mixed-case",
			input:    `<a HrEF="http://example.com" HIddEN>test</a>`,
			expected: ``,
		},
		{
			name:     "href-mixed-case",
			input:    `<a HrEF="http://example.com">test</a>`,
			expected: `<a href="http://example.com" rel="noopener noreferrer" referrerpolicy="no-referrer" target="_blank">test</a>`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			output := sanitizeHTMLWithDefaultOptions(baseURL, tc.input)
			if tc.expected != output {
				t.Errorf(`Wrong output for input %q: expected %q, got %q`, tc.input, tc.expected, output)
			}
		})
	}
}

func TestDeeplyNestedpage(t *testing.T) {
	input := "test"
	// -3 instead of -1 because <html><body> is automatically added.
	for range maxDepth - 3 {
		input = "<div>" + input + "</div>"
	}
	output := sanitizeHTMLWithDefaultOptions("http://example.org/", input)
	want := "test"

	if output != want {
		t.Errorf(`Wrong output: "%s" != "%s"`, want, output)
	}

	input = "test"
	for range maxDepth - 2 {
		input = "<div>" + input + "</div>"
	}
	output = sanitizeHTMLWithDefaultOptions("http://example.org/", input)

	if output != "" {
		t.Errorf(`Wrong output: "%s" != "%s"`, "", output)
	}
}
