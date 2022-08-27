// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rewrite // import "miniflux.app/reader/rewrite"

import (
	"reflect"
	"strings"
	"testing"
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
	output := Rewriter("https://example.org/article", `Some text.`, ``)
	expected := `Some text.`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithYoutubeLink(t *testing.T) {
	output := Rewriter("https://www.youtube.com/watch?v=1234", "Video Description", ``)
	expected := `<iframe width="650" height="350" frameborder="0" src="https://www.youtube-nocookie.com/embed/1234" allowfullscreen></iframe><br>Video Description`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithInexistingCustomRule(t *testing.T) {
	output := Rewriter("https://www.youtube.com/watch?v=1234", `Video Description`, `some rule`)
	expected := `Video Description`
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithXkcdLink(t *testing.T) {
	description := `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`
	output := Rewriter("https://xkcd.com/1912/", description, ``)
	expected := `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you."/><figcaption><p>Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you.</p></figcaption></figure>`
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithXkcdLinkHtmlInjection(t *testing.T) {
	description := `<img src="https://imgs.xkcd.com/comics/thermostat.png" title="<foo>" alt="<foo>" />`
	output := Rewriter("https://xkcd.com/1912/", description, ``)
	expected := `<figure><img src="https://imgs.xkcd.com/comics/thermostat.png" alt="&lt;foo&gt;"/><figcaption><p>&lt;foo&gt;</p></figcaption></figure>`
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithXkcdLinkAndImageNoTitle(t *testing.T) {
	description := `<img src="https://imgs.xkcd.com/comics/thermostat.png" alt="Your problem is so terrible, I worry that, if I help you, I risk drawing the attention of whatever god of technology inflicted it on you." />`
	output := Rewriter("https://xkcd.com/1912/", description, ``)
	expected := description
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithXkcdLinkAndNoImage(t *testing.T) {
	description := "test"
	output := Rewriter("https://xkcd.com/1912/", description, ``)
	expected := description
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithXkcdAndNoImage(t *testing.T) {
	description := "test"
	output := Rewriter("https://xkcd.com/1912/", description, ``)
	expected := description

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteMailtoLink(t *testing.T) {
	description := `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact</a>`
	output := Rewriter("https://www.qwantz.com/", description, ``)
	expected := `<a href="mailto:ryan@qwantz.com?subject=blah%20blah">contact [blah blah]</a>`
	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithPDFLink(t *testing.T) {
	description := "test"
	output := Rewriter("https://example.org/document.pdf", description, ``)
	expected := `<a href="https://example.org/document.pdf">PDF</a><br>test`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithNoLazyImage(t *testing.T) {
	description := `<img src="https://example.org/image.jpg" alt="Image"><noscript><p>Some text</p></noscript>`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := description

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithLazyImage(t *testing.T) {
	description := `<img src="" data-url="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := `<img src="https://example.org/image.jpg" data-url="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithLazyDivImage(t *testing.T) {
	description := `<div data-url="https://example.org/image.jpg" alt="Image"></div><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := `<img src="https://example.org/image.jpg" alt="Image"/><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithUnknownLazyNoScriptImage(t *testing.T) {
	description := `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"><noscript><img src="https://example.org/fallback.jpg" alt="Fallback"></noscript>`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := `<img src="" data-non-candidate="https://example.org/image.jpg" alt="Image"/><img src="https://example.org/fallback.jpg" alt="Fallback"/>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithLazySrcset(t *testing.T) {
	description := `<img srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := `<img srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteWithImageAndLazySrcset(t *testing.T) {
	description := `<img src="meow" srcset="" data-srcset="https://example.org/image.jpg" alt="Image">`
	output := Rewriter("https://example.org/article", description, "add_dynamic_image")
	expected := `<img src="meow" srcset="https://example.org/image.jpg" data-srcset="https://example.org/image.jpg" alt="Image"/>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestNewLineRewriteRule(t *testing.T) {
	description := "A\nB\nC"
	output := Rewriter("https://example.org/article", description, "nl2br")
	expected := `A<br>B<br>C`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestConvertTextLinkRewriteRule(t *testing.T) {
	description := "Test: http://example.org/a/b"
	output := Rewriter("https://example.org/article", description, "convert_text_link")
	expected := `Test: <a href="http://example.org/a/b">http://example.org/a/b</a>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestMediumImage(t *testing.T) {
	content := `
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
	`
	expected := `<img alt="Image for post" class="t u v if aj" src="https://miro.medium.com/max/2560/1*ephLSqSzQYLvb7faDwzRbw.jpeg" width="1280" height="720" srcset="https://miro.medium.com/max/552/1*ephLSqSzQYLvb7faDwzRbw.jpeg 276w, https://miro.medium.com/max/1104/1*ephLSqSzQYLvb7faDwzRbw.jpeg 552w, https://miro.medium.com/max/1280/1*ephLSqSzQYLvb7faDwzRbw.jpeg 640w, https://miro.medium.com/max/1400/1*ephLSqSzQYLvb7faDwzRbw.jpeg 700w" sizes="700px"/>`
	output := Rewriter("https://example.org/article", content, "fix_medium_images")
	output = strings.TrimSpace(output)

	if expected != output {
		t.Errorf(`Not expected output: %s`, output)
	}
}

func TestRewriteNoScriptImageWithoutNoScriptTag(t *testing.T) {
	content := `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><figcaption>MDN Logo</figcaption></figure>`
	expected := `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."/><figcaption>MDN Logo</figcaption></figure>`
	output := Rewriter("https://example.org/article", content, "use_noscript_figure_images")
	output = strings.TrimSpace(output)

	if expected != output {
		t.Errorf(`Not expected output: %s`, output)
	}
}

func TestRewriteNoScriptImageWithNoScriptTag(t *testing.T) {
	content := `<figure><img src="https://developer.mozilla.org/static/img/favicon144.png" alt="The beautiful MDN logo."><noscript><img src="http://example.org/logo.svg"></noscript><figcaption>MDN Logo</figcaption></figure>`
	expected := `<figure><img src="http://example.org/logo.svg"/><figcaption>MDN Logo</figcaption></figure>`
	output := Rewriter("https://example.org/article", content, "use_noscript_figure_images")
	output = strings.TrimSpace(output)

	if expected != output {
		t.Errorf(`Not expected output: %s`, output)
	}
}

func TestRewriteReplaceCustom(t *testing.T) {
	content := `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.svg">`
	expected := `<img src="http://example.org/logo.svg"><img src="https://example.org/article/picture.png">`
	output := Rewriter("https://example.org/article", content, `replace("article/(.*).svg"|"article/$1.png")`)

	if expected != output {
		t.Errorf(`Not expected output: %s`, output)
	}
}

func TestRewriteRemoveCustom(t *testing.T) {
	content := `<div>Lorem Ipsum <span class="spam">I dont want to see this</span><span class="ads keep">Super important info</span></div>`
	expected := `<div>Lorem Ipsum <span class="ads keep">Super important info</span></div>`
	output := Rewriter("https://example.org/article", content, `remove(".spam, .ads:not(.keep)")`)

	if expected != output {
		t.Errorf(`Not expected output: %s`, output)
	}
}

func TestRewriteAddCastopodEpisode(t *testing.T) {
	output := Rewriter("https://podcast.demo/@demo/episodes/test", "Episode Description", `add_castopod_episode`)
	expected := `<iframe width="650" frameborder="0" src="https://podcast.demo/@demo/episodes/test/embed/light"></iframe><br>Episode Description`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteBase64Decode(t *testing.T) {
	content := `VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=`
	expected := `This is some base64 encoded content`
	output := Rewriter("https://example.org/article", content, `base64_decode`)

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteBase64DecodeInHTML(t *testing.T) {
	content := `<div>Lorem Ipsum not valid base64<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`
	expected := `<div>Lorem Ipsum not valid base64<span class="base64">This is some base64 encoded content</span></div>`
	output := Rewriter("https://example.org/article", content, `base64_decode`)

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestRewriteBase64DecodeArgs(t *testing.T) {
	content := `<div>Lorem Ipsum<span class="base64">VGhpcyBpcyBzb21lIGJhc2U2NCBlbmNvZGVkIGNvbnRlbnQ=</span></div>`
	expected := `<div>Lorem Ipsum<span class="base64">This is some base64 encoded content</span></div>`
	output := Rewriter("https://example.org/article", content, `base64_decode(".base64")`)

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}
