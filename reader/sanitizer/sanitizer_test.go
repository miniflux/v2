// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sanitizer // import "miniflux.app/reader/sanitizer"

import "testing"

func TestValidInput(t *testing.T) {
	input := `<p>This is a <strong>text</strong> with an image: <img src="http://example.org/" alt="Test" loading="lazy">.</p>`
	output := Sanitize("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestImgWithWidthAndHeightAttribute(t *testing.T) {
	input := `<img src="https://example.org/image.png" width="10" height="20">`
	expected := `<img src="https://example.org/image.png" width="10" height="20" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestImgWithWidthAndHeightAttributeLargerThanMinifluxLayout(t *testing.T) {
	input := `<img src="https://example.org/image.png" width="1200" height="675">`
	expected := `<img src="https://example.org/image.png" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestImgWithIncorrectWidthAndHeightAttribute(t *testing.T) {
	input := `<img src="https://example.org/image.png" width="10px" height="20px">`
	expected := `<img src="https://example.org/image.png" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestImgWithTextDataURL(t *testing.T) {
	input := `<img src="data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==" alt="Example">`
	expected := ``
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestImgWithDataURL(t *testing.T) {
	input := `<img src="data:image/gif;base64,test" alt="Example">`
	expected := `<img src="data:image/gif;base64,test" alt="Example" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestImgWithSrcset(t *testing.T) {
	input := `<img srcset="example-320w.jpg, example-480w.jpg 1.5x,   example-640w.jpg 2x, example-640w.jpg 640w" src="example-640w.jpg" alt="Example">`
	expected := `<img srcset="http://example.org/example-320w.jpg, http://example.org/example-480w.jpg 1.5x, http://example.org/example-640w.jpg 2x, http://example.org/example-640w.jpg 640w" src="http://example.org/example-640w.jpg" alt="Example" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestSourceWithSrcsetAndMedia(t *testing.T) {
	input := `<picture><source media="(min-width: 800px)" srcset="elva-800w.jpg"></picture>`
	expected := `<picture><source media="(min-width: 800px)" srcset="http://example.org/elva-800w.jpg"></picture>`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestMediumImgWithSrcset(t *testing.T) {
	input := `<img alt="Image for post" class="t u v ef aj" src="https://miro.medium.com/max/5460/1*aJ9JibWDqO81qMfNtqgqrw.jpeg" srcset="https://miro.medium.com/max/552/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 276w, https://miro.medium.com/max/1000/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 500w" sizes="500px" width="2730" height="3407">`
	expected := `<img alt="Image for post" src="https://miro.medium.com/max/5460/1*aJ9JibWDqO81qMfNtqgqrw.jpeg" srcset="https://miro.medium.com/max/552/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 276w, https://miro.medium.com/max/1000/1*aJ9JibWDqO81qMfNtqgqrw.jpeg 500w" sizes="500px" loading="lazy">`
	output := Sanitize("http://example.org/", input)

	if output != expected {
		t.Errorf(`Wrong output: %s`, output)
	}
}

func TestSelfClosingTags(t *testing.T) {
	input := `<p>This <br> is a <strong>text</strong> <br/>with an image: <img src="http://example.org/" alt="Test" loading="lazy"/>.</p>`
	output := Sanitize("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestTable(t *testing.T) {
	input := `<table><tr><th>A</th><th colspan="2">B</th></tr><tr><td>C</td><td>D</td><td>E</td></tr></table>`
	output := Sanitize("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestRelativeURL(t *testing.T) {
	input := `This <a href="/test.html">link is relative</a> and this image: <img src="../folder/image.png"/>`
	expected := `This <a href="http://example.org/test.html" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">link is relative</a> and this image: <img src="http://example.org/folder/image.png" loading="lazy"/>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestProtocolRelativeURL(t *testing.T) {
	input := `This <a href="//static.example.org/index.html">link is relative</a>.`
	expected := `This <a href="https://static.example.org/index.html" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">link is relative</a>.`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidTag(t *testing.T) {
	input := `<p>My invalid <b>tag</b>.</p>`
	expected := `<p>My invalid tag.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestVideoTag(t *testing.T) {
	input := `<p>My valid <video src="videofile.webm" autoplay poster="posterimage.jpg">fallback</video>.</p>`
	expected := `<p>My valid <video src="http://example.org/videofile.webm" poster="http://example.org/posterimage.jpg" controls>fallback</video>.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAudioAndSourceTag(t *testing.T) {
	input := `<p>My music <audio controls="controls"><source src="foo.wav" type="audio/wav"></audio>.</p>`
	expected := `<p>My music <audio controls><source src="http://example.org/foo.wav" type="audio/wav"></audio>.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestUnknownTag(t *testing.T) {
	input := `<p>My invalid <unknown>tag</unknown>.</p>`
	expected := `<p>My invalid tag.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidNestedTag(t *testing.T) {
	input := `<p>My invalid <b>tag with some <em>valid</em> tag</b>.</p>`
	expected := `<p>My invalid tag with some <em>valid</em> tag.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidIFrame(t *testing.T) {
	input := `<iframe src="http://example.org/"></iframe>`
	expected := ``
	output := Sanitize("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestIFrameWithChildElements(t *testing.T) {
	input := `<iframe src="https://www.youtube.com/"><p>test</p></iframe>`
	expected := `<iframe src="https://www.youtube.com/" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.com/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAnchorLink(t *testing.T) {
	input := `<p>This link is <a href="#some-anchor">an anchor</a></p>`
	expected := `<p>This link is <a href="#some-anchor">an anchor</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestInvalidURLScheme(t *testing.T) {
	input := `<p>This link is <a src="file:///etc/passwd">not valid</a></p>`
	expected := `<p>This link is not valid</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestAPTURIScheme(t *testing.T) {
	input := `<p>This link is <a href="apt:some-package?channel=test">valid</a></p>`
	expected := `<p>This link is <a href="apt:some-package?channel=test" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestBitcoinURIScheme(t *testing.T) {
	input := `<p>This link is <a href="bitcoin:175tWpb8K1S7NmH4Zx6rewF9WQrcZv245W">valid</a></p>`
	expected := `<p>This link is <a href="bitcoin:175tWpb8K1S7NmH4Zx6rewF9WQrcZv245W" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestCallToURIScheme(t *testing.T) {
	input := `<p>This link is <a href="callto:12345679">valid</a></p>`
	expected := `<p>This link is <a href="callto:12345679" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestFeedURIScheme(t *testing.T) {
	input := `<p>This link is <a href="feed://example.com/rss.xml">valid</a></p>`
	expected := `<p>This link is <a href="feed://example.com/rss.xml" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="feed:https://example.com/rss.xml">valid</a></p>`
	expected = `<p>This link is <a href="feed:https://example.com/rss.xml" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestGeoURIScheme(t *testing.T) {
	input := `<p>This link is <a href="geo:13.4125,103.8667">valid</a></p>`
	expected := `<p>This link is <a href="geo:13.4125,103.8667" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestItunesURIScheme(t *testing.T) {
	input := `<p>This link is <a href="itms://itunes.com/apps/my-app-name">valid</a></p>`
	expected := `<p>This link is <a href="itms://itunes.com/apps/my-app-name" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="itms-apps://itunes.com/apps/my-app-name">valid</a></p>`
	expected = `<p>This link is <a href="itms-apps://itunes.com/apps/my-app-name" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestMagnetURIScheme(t *testing.T) {
	input := `<p>This link is <a href="magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&amp;xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7">valid</a></p>`
	expected := `<p>This link is <a href="magnet:?xt.1=urn:sha1:YNCKHTQCWBTRNJIV4WNAE52SJUQCZO5C&amp;xt.2=urn:sha1:TXGCZQTH26NL6OUQAJJPFALHG2LTGBC7" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestMailtoURIScheme(t *testing.T) {
	input := `<p>This link is <a href="mailto:jsmith@example.com?subject=A%20Test&amp;body=My%20idea%20is%3A%20%0A">valid</a></p>`
	expected := `<p>This link is <a href="mailto:jsmith@example.com?subject=A%20Test&amp;body=My%20idea%20is%3A%20%0A" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestNewsURIScheme(t *testing.T) {
	input := `<p>This link is <a href="news://news.server.example/*">valid</a></p>`
	expected := `<p>This link is <a href="news://news.server.example/*" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="news:example.group.this">valid</a></p>`
	expected = `<p>This link is <a href="news:example.group.this" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="nntp://news.server.example/example.group.this">valid</a></p>`
	expected = `<p>This link is <a href="nntp://news.server.example/example.group.this" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestRTMPURIScheme(t *testing.T) {
	input := `<p>This link is <a href="rtmp://mycompany.com/vod/mp4:mycoolvideo.mov">valid</a></p>`
	expected := `<p>This link is <a href="rtmp://mycompany.com/vod/mp4:mycoolvideo.mov" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSIPURIScheme(t *testing.T) {
	input := `<p>This link is <a href="sip:+1-212-555-1212:1234@gateway.com;user=phone">valid</a></p>`
	expected := `<p>This link is <a href="sip:+1-212-555-1212:1234@gateway.com;user=phone" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="sips:alice@atlanta.com?subject=project%20x&amp;priority=urgent">valid</a></p>`
	expected = `<p>This link is <a href="sips:alice@atlanta.com?subject=project%20x&amp;priority=urgent" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSkypeURIScheme(t *testing.T) {
	input := `<p>This link is <a href="skype:echo123?call">valid</a></p>`
	expected := `<p>This link is <a href="skype:echo123?call" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSpotifyURIScheme(t *testing.T) {
	input := `<p>This link is <a href="spotify:track:2jCnn1QPQ3E8ExtLe6INsx">valid</a></p>`
	expected := `<p>This link is <a href="spotify:track:2jCnn1QPQ3E8ExtLe6INsx" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSteamURIScheme(t *testing.T) {
	input := `<p>This link is <a href="steam://settings/account">valid</a></p>`
	expected := `<p>This link is <a href="steam://settings/account" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestSubversionURIScheme(t *testing.T) {
	input := `<p>This link is <a href="svn://example.org">valid</a></p>`
	expected := `<p>This link is <a href="svn://example.org" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}

	input = `<p>This link is <a href="svn+ssh://example.org">valid</a></p>`
	expected = `<p>This link is <a href="svn+ssh://example.org" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output = Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestTelURIScheme(t *testing.T) {
	input := `<p>This link is <a href="tel:+1-201-555-0123">valid</a></p>`
	expected := `<p>This link is <a href="tel:+1-201-555-0123" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestWebcalURIScheme(t *testing.T) {
	input := `<p>This link is <a href="webcal://example.com/calendar.ics">valid</a></p>`
	expected := `<p>This link is <a href="webcal://example.com/calendar.ics" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestXMPPURIScheme(t *testing.T) {
	input := `<p>This link is <a href="xmpp:user@host?subscribe&amp;type=subscribed">valid</a></p>`
	expected := `<p>This link is <a href="xmpp:user@host?subscribe&amp;type=subscribed" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">valid</a></p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestBlacklistedLink(t *testing.T) {
	input := `<p>This image is not valid <img src="https://stats.wordpress.com/some-tracker"></p>`
	expected := `<p>This image is not valid </p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestPixelTracker(t *testing.T) {
	input := `<p><img src="https://tracker1.example.org/" height="1" width="1"> and <img src="https://tracker2.example.org/" height="1" width="1"/></p>`
	expected := `<p> and </p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestXmlEntities(t *testing.T) {
	input := `<pre>echo "test" &gt; /etc/hosts</pre>`
	expected := `<pre>echo &#34;test&#34; &gt; /etc/hosts</pre>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestEspaceAttributes(t *testing.T) {
	input := `<td rowspan="<b>test</b>">test</td>`
	expected := `<td rowspan="&lt;b&gt;test&lt;/b&gt;">test</td>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURL(t *testing.T) {
	input := `<iframe src="http://www.youtube.com/embed/test123?version=3&#038;rel=1&#038;fs=1&#038;autohide=2&#038;showsearch=0&#038;showinfo=1&#038;iv_load_policy=1&#038;wmode=transparent"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?version=3&amp;rel=1&amp;fs=1&amp;autohide=2&amp;showsearch=0&amp;showinfo=1&amp;iv_load_policy=1&amp;wmode=transparent" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURL(t *testing.T) {
	input := `<iframe src="https://www.youtube.com/embed/test123"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURLWithParameters(t *testing.T) {
	input := `<iframe src="https://www.youtube.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURLAlreadyReplaced(t *testing.T) {
	input := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceProtocolRelativeYoutubeURL(t *testing.T) {
	input := `<iframe src="//www.youtube.com/embed/Bf2W84jrGqs" width="560" height="314" allowfullscreen="allowfullscreen"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/Bf2W84jrGqs" width="560" height="314" allowfullscreen="allowfullscreen" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceIframeURL(t *testing.T) {
	input := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0"></iframe>`
	expected := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0" sandbox="allow-scripts allow-same-origin allow-popups" loading="lazy"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceNoScript(t *testing.T) {
	input := `<p>Before paragraph.</p><noscript>Inside <code>noscript</code> tag with an image: <img src="http://example.org/" alt="Test" loading="lazy"></noscript><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceScript(t *testing.T) {
	input := `<p>Before paragraph.</p><script type="text/javascript">alert("1");</script><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceStyle(t *testing.T) {
	input := `<p>Before paragraph.</p><style>body { background-color: #ff0000; }</style><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}
