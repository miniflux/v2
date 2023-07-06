// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"reflect"
	"strings"
	"testing"

	"miniflux.app/model"
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
		Title:   `A title`,
		Content: `Some text.`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `Some text.`,
	}
	Rewriter("https://example.org/article", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithYoutubeLink(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/1234" allowfullscreen></iframe><br>Video Description`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `Video Description`,
	}
	Rewriter("https://www.youtube.com/watch?v=1234", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithInexistingCustomRule(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `Video Description`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `Video Description`,
	}
	Rewriter("https://www.youtube.com/watch?v=1234", testEntry, `some rule`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLink(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you."/><figcaption><p>Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you.</p></figcaption></figure>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	Rewriter("https://xkcd.com/1912/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkHtmlInjection(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="&lt;foo&gt;"/><figcaption><p>&lt;foo&gt;</p></figcaption></figure>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="<foo>" alt="<foo>" />`,
	}
	Rewriter("https://xkcd.com/1912/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkAndImageNoTitle(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`,
	}
	Rewriter("https://xkcd.com/1912/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdLinkAndNoImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `test`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `test`,
	}
	Rewriter("https://xkcd.com/1912/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithXkcdAndNoImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `test`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `test`,
	}
	Rewriter("https://xkcd.com/1912/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteMailtoLink(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact [blah blah]</a>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact</a>`,
	}
	Rewriter("https://www.qwantz.com/", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithPDFLink(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<a href="https://example.org/document.pdf">PDF</a><br>test`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `test`,
	}
	Rewriter("https://example.org/document.pdf", testEntry, ``)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithNoLazyImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"><noscript><p>Some text</p></noscript>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"><noscript><p>Some text</p></noscript>`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" data-url="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="" data-url="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazyDivImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div data-url="https://example.org/image.jpg" alt="Image"></div><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithUnknownLazyNoScriptImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"/><img src="https://example.org/fallback.jpg" alt="Fallback"/>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithLazySrcset(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteWithImageAndLazySrcset(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="meow" srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="meow" srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`,
	}
	Rewriter("https://example.org/article", testEntry, "add_dynamic_image")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestNewLineRewriteRule(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `A<br>B<br>C`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: "A\nB\nC",
	}
	Rewriter("https://example.org/article", testEntry, "nl2br")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestConvertTextLinkRewriteRule(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `Test: <a href="http://example.org/a/b">http://example.org/a/b</a>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `Test: http://example.org/a/b`,
	}
	Rewriter("https://example.org/article", testEntry, "convert_text_link")

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestMediumImage(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img alt="Image for post" class="t u v if aj" src="https://miro.medium.com/max/2560/1*ephLSqSzQYLvb7faDwzRbw.jpeg" width="1280" height="720" srcset="https://miro.medium.com/max/552/1*ephLSqSzQYLvb7faDwzRbw.jpeg 276w, https://miro.medium.com/max/1104/1*ephLSqSzQYLvb7faDwzRbw.jpeg 552w, https://miro.medium.com/max/1280/1*ephLSqSzQYLvb7faDwzRbw.jpeg 640w, https://miro.medium.com/max/1400/1*ephLSqSzQYLvb7faDwzRbw.jpeg 700w" sizes="700px"/>`,
	}
	testEntry := &model.Entry{
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
	Rewriter("https://example.org/article", testEntry, "fix_medium_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteNoScriptImageWithoutNoScriptTag(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."/><figcaption>MDN Logo</figcaption></figure>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><figcaption>MDN Logo</figcaption></figure>`,
	}
	Rewriter("https://example.org/article", testEntry, "use_noscript_figure_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteNoScriptImageWithNoScriptTag(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="http://example.org/logo.svg"/><figcaption>MDN Logo</figcaption></figure>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><noscript><img src="http://example.org/logo.svg"></noscript><figcaption>MDN Logo</figcaption></figure>`,
	}
	Rewriter("https://example.org/article", testEntry, "use_noscript_figure_images")
	testEntry.Content = strings.TrimSpace(testEntry.Content)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteReplaceCustom(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.png">`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.svg">`,
	}
	Rewriter("https://example.org/article", testEntry, `replace("article/(.*).svg"|"article/$1.png")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteRemoveCustom(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum <span class="ads keep">Super important info</span></div>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum <span class="spam">I dont want to see this</span><span class="ads keep">Super important info</span></div>`,
	}
	Rewriter("https://example.org/article", testEntry, `remove(".spam, .ads:not(.keep)")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteAddCastopodEpisode(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<iframe width="650" frameborder="0" src="https://podcast.demo/@demo/episodes/test/embed/light"></iframe><br>Episode Description`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `Episode Description`,
	}
	Rewriter("https://podcast.demo/@demo/episodes/test", testEntry, `add_castopod_episode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64Decode(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `This is some base64 encoded content`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=`,
	}
	Rewriter("https://example.org/article", testEntry, `base64_decode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64DecodeInHTML(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum not valid base64<span class="base64">This is some base64 encoded content</span></div>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum not valid base64<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`,
	}
	Rewriter("https://example.org/article", testEntry, `base64_decode`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteBase64DecodeArgs(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum<span class="base64">This is some base64 encoded content</span></div>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<div>Lorem Ipsum<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`,
	}
	Rewriter("https://example.org/article", testEntry, `base64_decode(".base64")`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRewriteRemoveTables(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `A title`,
		Content: `<p>Test</p><p>Hello World!</p><p>Test</p>`,
	}
	testEntry := &model.Entry{
		Title:   `A title`,
		Content: `<table class="container"><tbody><tr><td><p>Test</p><table class="row"><tbody><tr><td><p>Hello World!</p></td><td><p>Test</p></td></tr></tbody></table></td></tr></tbody></table>`,
	}
	Rewriter("https://example.org/article", testEntry, `remove_tables`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}

func TestRemoveClickbait(t *testing.T) {
	controlEntry := &model.Entry{
		Title:   `This Is Amazing`,
		Content: `Some description`,
	}
	testEntry := &model.Entry{
		Title:   `THIS IS AMAZING`,
		Content: `Some description`,
	}
	Rewriter("https://example.org/article", testEntry, `remove_clickbait`)

	if !reflect.DeepEqual(testEntry, controlEntry) {
		t.Errorf(`Not expected output: got "%+v" instead of "%+v"`, testEntry, controlEntry)
	}
}
