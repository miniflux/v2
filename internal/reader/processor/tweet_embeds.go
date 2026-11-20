// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"

	"github.com/PuerkitoBio/goquery"
	nethtml "golang.org/x/net/html"
)

// tweetSyndicationEndpoint is the public, unauthenticated endpoint used by
// X's own embed widget (and libraries such as react-tweet) to render tweets.
const tweetSyndicationEndpoint = "https://cdn.syndication.twimg.com/tweet-result"

// tweetSyndicationToken is sent because the endpoint returns an empty object
// without a token. X's widget derives it from the tweet ID, but the endpoint
// does not validate the value, only that it is present.
const tweetSyndicationToken = "miniflux"

const tweetFetchMaxBodySize = 2 * 1024 * 1024

// errTweetUnavailable means the endpoint answered, but not with a tweet.
var errTweetUnavailable = errors.New("tweet: unavailable")

var tweetStatusURLRegex = regexp.MustCompile(`^https?://(?:www\.|mobile\.)?(?:twitter|x)\.com/[A-Za-z0-9_]+/status(?:es)?/(\d+)`)

type tweetMedia struct {
	Type          string `json:"type"`
	URL           string `json:"url"`
	MediaURLHTTPS string `json:"media_url_https"`
	ExtAltText    string `json:"ext_alt_text"`
	OriginalInfo  struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"original_info"`
	VideoInfo struct {
		Variants []struct {
			Bitrate     int    `json:"bitrate"`
			ContentType string `json:"content_type"`
			URL         string `json:"url"`
		} `json:"variants"`
	} `json:"video_info"`
}

type syndicationTweet struct {
	TypeName     string       `json:"__typename"`
	MediaDetails []tweetMedia `json:"mediaDetails"`
	Entities     struct {
		URLs []struct {
			URL         string `json:"url"`
			ExpandedURL string `json:"expanded_url"`
			DisplayURL  string `json:"display_url"`
		} `json:"urls"`
	} `json:"entities"`
}

// tweetEmbedExpander expands the embedded tweets of the entries processed
// in one feed refresh. Each tweet is fetched at most once. After a request
// fails, for instance on a timeout, the remaining tweets are left alone so
// that a slow endpoint delays the refresh by one timeout at most.
type tweetEmbedExpander struct {
	fetch  func(tweetID string) (*syndicationTweet, error)
	tweets map[string]*syndicationTweet
	failed bool
}

func newTweetEmbedExpander() *tweetEmbedExpander {
	return &tweetEmbedExpander{
		fetch:  fetchTweetFromSyndication,
		tweets: make(map[string]*syndicationTweet),
	}
}

func (x *tweetEmbedExpander) tweet(tweetID string) *syndicationTweet {
	if tweet, ok := x.tweets[tweetID]; ok {
		return tweet
	}
	if x.failed {
		return nil
	}

	tweet, err := x.fetch(tweetID)
	if err != nil {
		slog.Debug("Unable to fetch tweet media",
			slog.String("tweet_id", tweetID),
			slog.Any("error", err),
		)
		tweet = nil
		if !errors.Is(err, errTweetUnavailable) {
			x.failed = true
		}
	}
	x.tweets[tweetID] = tweet
	return tweet
}

// hasEmbeddedTweet reports whether the content contains a tweet embed that
// expand would fetch.
func hasEmbeddedTweet(entryContent string) bool {
	if !strings.Contains(entryContent, "/status") {
		return false
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return false
	}

	found := false
	doc.Find("blockquote").EachWithBreak(func(_ int, blockquote *goquery.Selection) bool {
		found = embeddedTweetID(blockquote) != ""
		return !found
	})
	return found
}

// expand finds embedded tweets (blockquotes ending with a link to a tweet)
// and appends the tweet's photos and videos, which are otherwise lost
// because the embed script is removed by the sanitizer. It also replaces the
// tweet's t.co links with the URLs they point to.
func (x *tweetEmbedExpander) expand(entryContent string) string {
	if !strings.Contains(entryContent, "/status") {
		return entryContent
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entryContent))
	if err != nil {
		return entryContent
	}

	modified := false
	doc.Find("blockquote").Each(func(_ int, blockquote *goquery.Selection) {
		tweetID := embeddedTweetID(blockquote)
		if tweetID == "" {
			return
		}

		tweet := x.tweet(tweetID)
		if tweet == nil {
			return
		}

		// Maps each t.co URL in the tweet text to the HTML replacing it.
		shortLinks := make(map[string]string)
		for _, link := range tweet.Entities.URLs {
			if strings.HasPrefix(link.URL, "https://t.co/") && (strings.HasPrefix(link.ExpandedURL, "https://") || strings.HasPrefix(link.ExpandedURL, "http://")) {
				shortLinks[link.URL] = `<a href="` + html.EscapeString(link.ExpandedURL) + `">` + html.EscapeString(cmp.Or(link.DisplayURL, link.ExpandedURL)) + `</a>`
			}
		}

		var mediaHTML strings.Builder
		for _, media := range tweet.MediaDetails {
			if mediaElement := buildTweetMediaHTML(media); mediaElement != "" {
				mediaHTML.WriteString(mediaElement)
				// The media placeholder link is redundant once the media is shown.
				if strings.HasPrefix(media.URL, "https://t.co/") {
					shortLinks[media.URL] = ""
				}
			}
		}

		if replaceTweetShortLinks(blockquote, shortLinks) {
			modified = true
		}
		if mediaHTML.Len() > 0 {
			blockquote.AppendHtml(mediaHTML.String())
			modified = true
		}
	})

	if !modified {
		return entryContent
	}

	output, _ := doc.FindMatcher(goquery.Single("body")).Html()
	return output
}

// embeddedTweetID returns the tweet ID of a blockquote whose last link points
// to a tweet, which is how X's embed code (and most copies of it) ends. A
// link inside a nested blockquote belongs to that blockquote, so a quote
// wrapping a tweet embed doesn't get the tweet's media a second time.
func embeddedTweetID(blockquote *goquery.Selection) string {
	links := blockquote.Find("a[href]")
	if links.Length() == 0 {
		return ""
	}

	lastLink := links.Last()
	if lastLink.Closest("blockquote").Get(0) != blockquote.Get(0) {
		return ""
	}

	href, _ := lastLink.Attr("href")
	matches := tweetStatusURLRegex.FindStringSubmatch(href)
	if matches == nil {
		return ""
	}
	return matches[1]
}

func buildTweetMediaHTML(media tweetMedia) string {
	if !strings.HasPrefix(media.MediaURLHTTPS, "https://") {
		return ""
	}

	switch media.Type {
	case "photo":
		dimensions := ""
		if media.OriginalInfo.Width > 0 && media.OriginalInfo.Height > 0 {
			dimensions = fmt.Sprintf(` width="%d" height="%d"`, media.OriginalInfo.Width, media.OriginalInfo.Height)
		}
		return `<figure><img src="` + html.EscapeString(media.MediaURLHTTPS) + `" alt="` + html.EscapeString(media.ExtAltText) + `"` + dimensions + `></figure>`
	case "video", "animated_gif":
		videoURL := ""
		bestBitrate := -1
		for _, variant := range media.VideoInfo.Variants {
			if variant.ContentType == "video/mp4" && strings.HasPrefix(variant.URL, "https://") && variant.Bitrate > bestBitrate {
				videoURL = variant.URL
				bestBitrate = variant.Bitrate
			}
		}
		if videoURL == "" {
			return ""
		}
		// No width and height here: unlike images, videos don't get "height: auto"
		// from the reader stylesheets, so the attributes would keep the full height
		// when the video is scaled down. The poster gives the right aspect ratio.
		return `<figure><video src="` + html.EscapeString(videoURL) + `" poster="` + html.EscapeString(media.MediaURLHTTPS) + `"></video></figure>`
	}

	return ""
}

// replaceTweetShortLinks replaces t.co URLs in the tweet text, whether they
// are links or bare text, with the given HTML.
func replaceTweetShortLinks(blockquote *goquery.Selection, shortLinks map[string]string) bool {
	if len(shortLinks) == 0 {
		return false
	}

	modified := false
	blockquote.Find("a[href]").Each(func(_ int, link *goquery.Selection) {
		href, _ := link.Attr("href")
		if replacement, ok := shortLinks[href]; ok {
			link.ReplaceWithHtml(replacement)
			modified = true
		}
	})

	oldNew := make([]string, 0, len(shortLinks)*2)
	for shortURL, replacement := range shortLinks {
		oldNew = append(oldNew, shortURL, replacement)
	}
	replacer := strings.NewReplacer(oldNew...)

	blockquote.Find("*").AddSelection(blockquote).Contents().Each(func(_ int, node *goquery.Selection) {
		textNode := node.Get(0)
		if textNode.Type != nethtml.TextNode || !strings.Contains(textNode.Data, "https://t.co/") {
			return
		}
		// Links can't be nested.
		if node.ParentsFiltered("a").Length() > 0 {
			return
		}
		if replaced := replacer.Replace(html.EscapeString(textNode.Data)); replaced != html.EscapeString(textNode.Data) {
			node.ReplaceWithHtml(replaced)
			modified = true
		}
	})

	return modified
}

func fetchTweetFromSyndication(tweetID string) (*syndicationTweet, error) {
	query := url.Values{}
	query.Set("id", tweetID)
	query.Set("token", tweetSyndicationToken)

	requestBuilder := fetcher.NewRequestBuilder().
		WithTimeout(config.Opts.HTTPClientTimeout()).
		WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(tweetSyndicationEndpoint + "?" + query.Encode()))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		return nil, localizedError.Error()
	}

	var tweet syndicationTweet
	if err := json.NewDecoder(responseHandler.Body(tweetFetchMaxBodySize)).Decode(&tweet); err != nil {
		return nil, fmt.Errorf("%w: unable to decode JSON: %w", errTweetUnavailable, err)
	}
	// Missing, deleted, and protected tweets come back as an empty object or a
	// "TweetTombstone".
	if tweet.TypeName != "Tweet" {
		return nil, fmt.Errorf("%w: unexpected response type %q", errTweetUnavailable, tweet.TypeName)
	}
	return &tweet, nil
}
