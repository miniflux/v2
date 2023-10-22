// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/v2/internal/reader/subscription"

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/errors"
	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/integration/rssbridge"
	"miniflux.app/v2/internal/reader/browser"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

var (
	errUnreadableDoc    = "Unable to analyze this page: %v"
	youtubeChannelRegex = regexp.MustCompile(`youtube\.com/channel/(.*)`)
	youtubeVideoRegex   = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
)

// FindSubscriptions downloads and try to find one or more subscriptions from an URL.
func FindSubscriptions(websiteURL, userAgent, cookie, username, password string, fetchViaProxy, allowSelfSignedCertificates bool, rssbridgeURL string) (Subscriptions, *errors.LocalizedError) {
	if rssbridgeURL != "" {
		bridges, err := rssbridge.DetectBridges(rssbridgeURL, websiteURL)
		if err != nil {
			return nil, errors.NewLocalizedError("RSS-Bridge: %v", err)
		}
		if len(bridges) > 0 {
			var subscriptions Subscriptions
			for _, bridge := range bridges {
				subscriptions = append(subscriptions, &Subscription{
					Title: bridge.BridgeMeta.Name,
					URL:   bridge.URL,
					Type:  "atom",
				})
			}
			return subscriptions, nil
		}
	}

	websiteURL = findYoutubeChannelFeed(websiteURL)
	websiteURL = parseYoutubeVideoPage(websiteURL)

	clt := client.NewClientWithConfig(websiteURL, config.Opts)
	clt.WithCredentials(username, password)
	clt.WithUserAgent(userAgent)
	clt.WithCookie(cookie)
	clt.AllowSelfSignedCertificates = allowSelfSignedCertificates

	if fetchViaProxy {
		clt.WithProxy()
	}

	response, err := browser.Exec(clt)
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

	subscriptions, err := parseWebPage(response.EffectiveURL, strings.NewReader(body))
	if err != nil || subscriptions != nil {
		return subscriptions, err
	}

	return tryWellKnownUrls(websiteURL, userAgent, cookie, username, password)
}

func parseWebPage(websiteURL string, data io.Reader) (Subscriptions, *errors.LocalizedError) {
	var subscriptions Subscriptions
	queries := map[string]string{
		"link[type='application/rss+xml']":   "rss",
		"link[type='application/atom+xml']":  "atom",
		"link[type='application/json']":      "json",
		"link[type='application/feed+json']": "json",
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
			}

			if feedURL, exists := s.Attr("href"); exists {
				if feedURL != "" {
					subscription.URL, _ = urllib.AbsoluteURL(websiteURL, feedURL)
				}
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

func parseYoutubeVideoPage(websiteURL string) string {
	if !youtubeVideoRegex.MatchString(websiteURL) {
		return websiteURL
	}

	clt := client.NewClientWithConfig(websiteURL, config.Opts)
	response, browserErr := browser.Exec(clt)
	if browserErr != nil {
		return websiteURL
	}

	doc, docErr := goquery.NewDocumentFromReader(response.Body)
	if docErr != nil {
		return websiteURL
	}

	if channelID, exists := doc.Find(`meta[itemprop="channelId"]`).First().Attr("content"); exists {
		return fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?channel_id=%s`, channelID)
	}

	return websiteURL
}

func tryWellKnownUrls(websiteURL, userAgent, cookie, username, password string) (Subscriptions, *errors.LocalizedError) {
	var subscriptions Subscriptions
	knownURLs := map[string]string{
		"atom.xml": "atom",
		"feed.xml": "atom",
		"feed/":    "atom",
		"rss.xml":  "rss",
		"rss/":     "rss",
	}

	websiteURLRoot := urllib.RootURL(websiteURL)
	baseURLs := []string{
		// Look for knownURLs in the root.
		websiteURLRoot,
	}
	// Look for knownURLs in current subdirectory, such as 'example.com/blog/'.
	websiteURL, _ = urllib.AbsoluteURL(websiteURL, "./")
	if websiteURL != websiteURLRoot {
		baseURLs = append(baseURLs, websiteURL)
	}

	for _, baseURL := range baseURLs {
		for knownURL, kind := range knownURLs {
			fullURL, err := urllib.AbsoluteURL(baseURL, knownURL)
			if err != nil {
				continue
			}
			clt := client.NewClientWithConfig(fullURL, config.Opts)
			clt.WithCredentials(username, password)
			clt.WithUserAgent(userAgent)
			clt.WithCookie(cookie)

			// Some websites redirects unknown URLs to the home page.
			// As result, the list of known URLs is returned to the subscription list.
			// We don't want the user to choose between invalid feed URLs.
			clt.WithoutRedirects()

			response, err := clt.Get()
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
	}

	return subscriptions, nil
}
