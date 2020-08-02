// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package subscription // import "miniflux.app/reader/subscription"

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"miniflux.app/errors"
	"miniflux.app/http/client"
	"miniflux.app/reader/browser"
	"miniflux.app/reader/parser"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
)

var (
	errUnreadableDoc    = "Unable to analyze this page: %v"
	youtubeChannelRegex = regexp.MustCompile(`youtube\.com/channel/(.*)`)
)

// FindSubscriptions downloads and try to find one or more subscriptions from an URL.
func FindSubscriptions(websiteURL, userAgent, username, password string) (Subscriptions, *errors.LocalizedError) {
	websiteURL = findYoutubeChannelFeed(websiteURL)

	request := client.New(websiteURL)
	request.WithCredentials(username, password)
	request.WithUserAgent(userAgent)
	response, err := browser.Exec(request)
	if err != nil {
		return nil, err
	}

	body := response.BodyAsString()
	if format := parser.DetectFeedFormat(body); format != parser.FormatUnknown {
		var subscriptions Subscriptions
		subscriptions = append(subscriptions, &Subscription{
			Title: response.EffectiveURL,
			URL:   response.EffectiveURL,
			Type:  format,
		})

		return subscriptions, nil
	}

	subscriptions, err := parseDocument(response.EffectiveURL, strings.NewReader(body))
	if err != nil || subscriptions != nil {
		return subscriptions, err
	}
	return tryWellKnownUrls(websiteURL, userAgent, username, password)
}

func parseDocument(websiteURL string, data io.Reader) (Subscriptions, *errors.LocalizedError) {
	var subscriptions Subscriptions
	queries := map[string]string{
		"link[type='application/rss+xml']":  "rss",
		"link[type='application/atom+xml']": "atom",
		"link[type='application/json']":     "json",
	}

	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return nil, errors.NewLocalizedError(errUnreadableDoc, err)
	}

	for query, kind := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			subscription := new(Subscription)
			subscription.Type = kind

			if title, exists := s.Attr("title"); exists {
				subscription.Title = title
			} else {
				subscription.Title = "Feed"
			}

			if feedURL, exists := s.Attr("href"); exists {
				subscription.URL, _ = url.AbsoluteURL(websiteURL, feedURL)
			}

			if subscription.Title == "" {
				subscription.Title = subscription.URL
			}

			if subscription.URL != "" {
				subscriptions = append(subscriptions, subscription)
			}
		})
	}

	return subscriptions, nil
}

func findYoutubeChannelFeed(websiteURL string) string {
	matches := youtubeChannelRegex.FindStringSubmatch(websiteURL)

	if len(matches) == 2 {
		return fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?channel_id=%s`, matches[1])
	}
	return websiteURL
}

func tryWellKnownUrls(websiteURL, userAgent, username, password string) (Subscriptions, *errors.LocalizedError) {
	var subscriptions Subscriptions
	knownURLs := map[string]string{
		"/atom.xml": "atom",
		"/feed.xml": "atom",
		"/feed/":    "atom",
		"/rss.xml":  "rss",
	}

	lastCharacter := websiteURL[len(websiteURL)-1:]
	if lastCharacter == "/" {
		websiteURL = websiteURL[:len(websiteURL)-1]
	}

	for knownURL, kind := range knownURLs {
		fullURL, err := url.AbsoluteURL(websiteURL, knownURL)
		if err != nil {
			continue
		}
		request := client.New(fullURL)
		request.WithCredentials(username, password)
		request.WithUserAgent(userAgent)
		response, err := request.Get()
		if err != nil {
			continue
		}

		if response != nil && response.StatusCode == 200 {
			subscription := new(Subscription)
			subscription.Type = kind
			subscription.Title = fullURL
			subscription.URL = fullURL
			if subscription.URL != "" {
				subscriptions = append(subscriptions, subscription)
			}
		}
	}

	return subscriptions, nil
}
