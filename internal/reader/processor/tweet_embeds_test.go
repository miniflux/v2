// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

const tweetVideoJSON = `{
	"mediaDetails": [{
		"type": "video",
		"url": "https://t.co/pXdUR21Gev",
		"media_url_https": "https://pbs.twimg.com/amplify_video_thumb/1/img/poster.jpg",
		"original_info": {"width": 896, "height": 1000},
		"video_info": {"variants": [
			{"content_type": "application/x-mpegURL", "url": "https://video.twimg.com/pl/playlist.m3u8"},
			{"bitrate": 632000, "content_type": "video/mp4", "url": "https://video.twimg.com/vid/320x356/low.mp4"},
			{"bitrate": 2176000, "content_type": "video/mp4", "url": "https://video.twimg.com/vid/720x802/high.mp4"},
			{"bitrate": 950000, "content_type": "video/mp4", "url": "https://video.twimg.com/vid/480x534/mid.mp4"}
		]}
	}]
}`

const tweetLinksJSON = `{
	"__typename": "Tweet",
	"entities": {"urls": [
		{"url": "https://t.co/luC5CdcWDy", "expanded_url": "https://alzresearch.org/giving/", "display_url": "alzresearch.org/giving/"},
		{"url": "https://t.co/zvNptfyuNd", "expanded_url": "https://alzfdn.org/donate/?a=1&b=2", "display_url": "alzfdn.org/donate/?a=1&…"},
		{"url": "https://t.co/unsafe1234", "expanded_url": "javascript:alert(1)", "display_url": "unsafe"}
	]}
}`

const tweetPhotosJSON = `{
	"mediaDetails": [
		{"type": "photo", "url": "https://t.co/yV1a1lbJcr", "media_url_https": "https://pbs.twimg.com/media/one.png", "ext_alt_text": "A \"quoted\" alt", "original_info": {"width": 160, "height": 160}},
		{"type": "photo", "url": "https://t.co/yV1a1lbJcr", "media_url_https": "https://pbs.twimg.com/media/two.jpg"}
	]
}`

// mockTweetExpander returns an expander that serves the given responses and
// fails with errTweetUnavailable for other tweets, and a pointer to the
// number of fetches.
func mockTweetExpander(t *testing.T, responses map[string]string) (*tweetEmbedExpander, *int) {
	t.Helper()

	calls := 0
	expander := newTweetEmbedExpander()
	expander.fetch = func(tweetID string) (*syndicationTweet, error) {
		calls++
		body, ok := responses[tweetID]
		if !ok {
			return nil, fmt.Errorf("%w: not found", errTweetUnavailable)
		}
		var tweet syndicationTweet
		if err := json.Unmarshal([]byte(body), &tweet); err != nil {
			t.Fatalf("invalid test JSON: %v", err)
		}
		return &tweet, nil
	}
	return expander, &calls
}

func TestExpandTweetEmbedsVideo(t *testing.T) {
	expander, _ := mockTweetExpander(t, map[string]string{"2116296999297548615": tweetVideoJSON})

	input := `<p>Intro</p><blockquote class="twitter-tweet"><p lang="en" dir="ltr">Hello! <a href="https://t.co/pXdUR21Gev">pic.twitter.com/pXdUR21Gev</a></p>— Guilherme Rambo (@_inside) <a href="https://twitter.com/_inside/status/2116296999297548615?ref_src=twsrc%5Etfw">October 30, 2026</a></blockquote><script async src="https://platform.twitter.com/widgets.js"></script>`
	output := expander.expand(input)

	expectedVideo := `<figure><video src="https://video.twimg.com/vid/720x802/high.mp4" poster="https://pbs.twimg.com/amplify_video_thumb/1/img/poster.jpg"></video></figure></blockquote>`
	if !strings.Contains(output, expectedVideo) {
		t.Errorf(`Expected the highest bitrate MP4 at the end of the blockquote, got: %s`, output)
	}
	if strings.Contains(output, "pic.twitter.com") {
		t.Errorf(`Expected the media placeholder link to be removed, got: %s`, output)
	}
	if !strings.Contains(output, `October 30, 2026</a>`) {
		t.Errorf(`Expected the tweet link to be kept, got: %s`, output)
	}
}

func TestExpandTweetEmbedsPhotosWithBareMediaURL(t *testing.T) {
	expander, _ := mockTweetExpander(t, map[string]string{"2116296960600928345": tweetPhotosJSON})

	input := `<blockquote><p>HomeView Software contains this image https://t.co/yV1a1lbJcr</p>— Aaron (@aaronp613) <a href="https://x.com/aaronp613/status/2116296960600928345">October 30, 2026</a></blockquote>`
	output := expander.expand(input)

	expected := `<blockquote><p>HomeView Software contains this image </p>— Aaron (@aaronp613) <a href="https://x.com/aaronp613/status/2116296960600928345">October 30, 2026</a>` +
		`<figure><img src="https://pbs.twimg.com/media/one.png" alt="A &#34;quoted&#34; alt" width="160" height="160"/></figure>` +
		`<figure><img src="https://pbs.twimg.com/media/two.jpg" alt=""/></figure></blockquote>`
	if output != expected {
		t.Errorf(`Unexpected output:\ngot:  %s\nwant: %s`, output, expected)
	}
}

func TestExpandTweetEmbedsIgnoresOtherBlockquotes(t *testing.T) {
	expander, calls := mockTweetExpander(t, nil)

	inputs := []string{
		`<blockquote><p>A regular quote.</p></blockquote>`,
		`<blockquote><p>Quote with a link to <a href="https://example.org/status/123">a status page</a>.</p></blockquote>`,
		`<blockquote><p>See <a href="https://x.com/user/status/123">this tweet</a> and <a href="https://example.org/">this site</a>.</p></blockquote>`,
		`<p>A link to a tweet outside a blockquote: <a href="https://x.com/user/status/123">tweet</a></p>`,
	}

	for _, input := range inputs {
		if output := expander.expand(input); output != input {
			t.Errorf(`Expected content to be unchanged:\ngot:  %s\nwant: %s`, output, input)
		}
	}
	if *calls != 0 {
		t.Errorf(`Expected no fetches, got %d`, *calls)
	}
}

func TestExpandTweetEmbedsKeepsContentWhenFetchFails(t *testing.T) {
	expander, _ := mockTweetExpander(t, nil)

	input := `<blockquote><p>Text https://t.co/abc</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`
	if output := expander.expand(input); output != input {
		t.Errorf(`Expected content to be unchanged:\ngot:  %s\nwant: %s`, output, input)
	}
}

func TestExpandTweetEmbedsKeepsContentWithoutMedia(t *testing.T) {
	expander, _ := mockTweetExpander(t, map[string]string{"42": `{"mediaDetails": []}`})

	input := `<blockquote><p>Just text</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`
	if output := expander.expand(input); output != input {
		t.Errorf(`Expected content to be unchanged:\ngot:  %s\nwant: %s`, output, input)
	}
}

func TestExpandTweetEmbedsSkipsUnsafeMediaURLs(t *testing.T) {
	expander, _ := mockTweetExpander(t, map[string]string{"42": `{"mediaDetails": [
		{"type": "photo", "media_url_https": "javascript:alert(1)"},
		{"type": "video", "media_url_https": "https://pbs.twimg.com/poster.jpg", "video_info": {"variants": [{"bitrate": 1, "content_type": "video/mp4", "url": "http://video.twimg.com/insecure.mp4"}]}}
	]}`})

	input := `<blockquote><p>Text</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`
	if output := expander.expand(input); output != input {
		t.Errorf(`Expected content to be unchanged:\ngot:  %s\nwant: %s`, output, input)
	}
}

func TestExpandTweetEmbedsExpandsShortLinks(t *testing.T) {
	expander, _ := mockTweetExpander(t, map[string]string{"42": tweetLinksJSON})

	input := `<blockquote><p>Give to <a href="https://t.co/luC5CdcWDy">https://t.co/luC5CdcWDy</a> &amp; https://t.co/zvNptfyuNd or <a href="https://example.org/">https://t.co/luC5CdcWDy</a> https://t.co/unsafe1234</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`
	expected := `<blockquote><p>Give to <a href="https://alzresearch.org/giving/">alzresearch.org/giving/</a> &amp; <a href="https://alzfdn.org/donate/?a=1&amp;b=2">alzfdn.org/donate/?a=1&amp;…</a> or <a href="https://example.org/">https://t.co/luC5CdcWDy</a> https://t.co/unsafe1234</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`

	if output := expander.expand(input); output != expected {
		t.Errorf(`Unexpected output:\ngot:  %s\nwant: %s`, output, expected)
	}
}

func TestExpandTweetEmbedsNestedBlockquote(t *testing.T) {
	expander, calls := mockTweetExpander(t, map[string]string{"2116296960600928345": tweetPhotosJSON})

	input := `<blockquote><p>Someone wrote:</p><blockquote><p>Image</p>— Aaron <a href="https://x.com/aaronp613/status/2116296960600928345">Date</a></blockquote></blockquote>`
	output := expander.expand(input)

	if count := strings.Count(output, `<img src="https://pbs.twimg.com/media/one.png"`); count != 1 {
		t.Errorf(`Expected the photo to be added once, got %d times: %s`, count, output)
	}
	if !strings.Contains(output, `two.jpg" alt=""/></figure></blockquote></blockquote>`) {
		t.Errorf(`Expected the media to be added to the inner blockquote, got: %s`, output)
	}
	if *calls != 1 {
		t.Errorf(`Expected 1 fetch, got %d`, *calls)
	}
}

func TestExpandTweetEmbedsFetchesEachTweetOnce(t *testing.T) {
	expander, calls := mockTweetExpander(t, map[string]string{"2116296999297548615": tweetVideoJSON})

	input := `<blockquote><p>Hello!</p><a href="https://x.com/_inside/status/2116296999297548615">Date</a></blockquote>`
	expander.expand(input)
	expander.expand(input)
	expander.expand(`<blockquote><p>Missing</p><a href="https://x.com/user/status/1">Date</a></blockquote>`)
	expander.expand(`<blockquote><p>Missing</p><a href="https://x.com/user/status/1">Date</a></blockquote>`)

	if *calls != 2 {
		t.Errorf(`Expected 2 fetches (one success, one failure), got %d`, *calls)
	}
}

func TestExpandTweetEmbedsStopsFetchingAfterRequestFailure(t *testing.T) {
	calls := 0
	expander := newTweetEmbedExpander()
	expander.fetch = func(tweetID string) (*syndicationTweet, error) {
		calls++
		return nil, errors.New("timeout")
	}

	expander.expand(`<blockquote><p>One</p><a href="https://x.com/user/status/1">Date</a></blockquote>`)
	expander.expand(`<blockquote><p>Two</p><a href="https://x.com/user/status/2">Date</a></blockquote>`)

	if calls != 1 {
		t.Errorf(`Expected 1 fetch, got %d`, calls)
	}
}

func TestExpandTweetEmbedsKeepsFetchingAfterUnavailableTweet(t *testing.T) {
	expander, calls := mockTweetExpander(t, map[string]string{"2": tweetVideoJSON})

	expander.expand(`<blockquote><p>Deleted</p><a href="https://x.com/user/status/1">Date</a></blockquote>`)
	output := expander.expand(`<blockquote><p>Two</p><a href="https://x.com/user/status/2">Date</a></blockquote>`)

	if *calls != 2 {
		t.Errorf(`Expected 2 fetches, got %d`, *calls)
	}
	if !strings.Contains(output, "<video") {
		t.Errorf(`Expected the second tweet to be expanded, got: %s`, output)
	}
}

func TestHasEmbeddedTweet(t *testing.T) {
	scenarios := map[string]bool{
		`<blockquote><p>Text</p>— User <a href="https://twitter.com/user/status/42">Date</a></blockquote>`:                   true,
		`<blockquote><p>Quote</p><blockquote><a href="https://x.com/user/status/42">Date</a></blockquote></blockquote>`:      true,
		`<blockquote><p>A regular quote.</p></blockquote>`:                                                                   false,
		`<p>A link to a tweet outside a blockquote: <a href="https://x.com/user/status/123">tweet</a></p>`:                   false,
		`<blockquote><a href="https://x.com/user/status/123">tweet</a> <a href="https://example.org/">site</a></blockquote>`: false,
	}

	for input, expected := range scenarios {
		if result := hasEmbeddedTweet(input); result != expected {
			t.Errorf(`hasEmbeddedTweet(%q) = %v, want %v`, input, result, expected)
		}
	}
}
