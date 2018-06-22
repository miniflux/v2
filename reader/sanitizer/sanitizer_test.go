// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package sanitizer

import "testing"

func TestValidInput(t *testing.T) {
	input := `<p>This is a <strong>text</strong> with an image: <img src="http://example.org/" alt="Test">.</p>`
	output := Sanitize("http://example.org/", input)

	if input != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, input, output)
	}
}

func TestSelfClosingTags(t *testing.T) {
	input := `<p>This <br> is a <strong>text</strong> <br/>with an image: <img src="http://example.org/" alt="Test"/>.</p>`
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
	expected := `This <a href="http://example.org/test.html" rel="noopener noreferrer" target="_blank" referrerpolicy="no-referrer">link is relative</a> and this image: <img src="http://example.org/folder/image.png"/>`
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
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?version=3&amp;rel=1&amp;fs=1&amp;autohide=2&amp;showsearch=0&amp;showinfo=1&amp;iv_load_policy=1&amp;wmode=transparent"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURL(t *testing.T) {
	input := `<iframe src="https://www.youtube.com/embed/test123"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceSecureYoutubeURLWithParameters(t *testing.T) {
	input := `<iframe src="https://www.youtube.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceYoutubeURLAlreadyReplaced(t *testing.T) {
	input := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	expected := `<iframe src="https://www.youtube-nocookie.com/embed/test123?rel=0&amp;controls=0"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceIframeURL(t *testing.T) {
	input := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0"></iframe>`
	expected := `<iframe src="https://player.vimeo.com/video/123456?title=0&amp;byline=0"></iframe>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}

func TestReplaceNoScript(t *testing.T) {
	input := `<p>Before paragraph.</p><noscript>Inside <code>noscript</code> tag with an image: <img src="http://example.org/" alt="Test"></noscript><p>After paragraph.</p>`
	expected := `<p>Before paragraph.</p><p>After paragraph.</p>`
	output := Sanitize("http://example.org/", input)

	if expected != output {
		t.Errorf(`Wrong output: "%s" != "%s"`, expected, output)
	}
}
