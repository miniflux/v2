// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/v2/internal/reader/subscription"

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"regexp"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/integration/rssbridge"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

var (
	youtubeChannelRegex = regexp.MustCompile(`youtube\.com/channel/(.*)`)
	youtubeVideoRegex   = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
)

func FindSubscriptions(websiteURL, userAgent, cookie, username, password string, fetchViaProxy, allowSelfSignedCertificates bool, rssbridgeURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	websiteURL = findYoutubeChannelFeed(websiteURL)
	websiteURL = parseYoutubeVideoPage(websiteURL)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(username, password)
	requestBuilder.WithUserAgent(userAgent)
	requestBuilder.WithCookie(cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(fetchViaProxy)
	requestBuilder.IgnoreTLSErrors(allowSelfSignedCertificates)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to find subscriptions", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return nil, localizedError
	}

	responseBody, localizedError := responseHandler.ReadBody(config.Opts.HTTPClientMaxBodySize())
	if localizedError != nil {
		slog.Warn("Unable to find subscriptions", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return nil, localizedError
	}

	if format := parser.DetectFeedFormat(string(responseBody)); format != parser.FormatUnknown {
		var subscriptions Subscriptions
		subscriptions = append(subscriptions, &Subscription{
			Title: responseHandler.EffectiveURL(),
			URL:   responseHandler.EffectiveURL(),
			Type:  format,
		})

		return subscriptions, nil
	}

	subscriptions, localizedError := parseWebPage(responseHandler.EffectiveURL(), bytes.NewReader(responseBody))
	if localizedError != nil || subscriptions != nil {
		return subscriptions, localizedError
	}

	if rssbridgeURL != "" {
		slog.Debug("Trying to detect feeds using RSS-Bridge",
			slog.String("website_url", websiteURL),
			slog.String("rssbridge_url", rssbridgeURL),
		)

		bridges, err := rssbridge.DetectBridges(rssbridgeURL, websiteURL)
		if err != nil {
			return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_detect_rssbridge", err)
		}

		slog.Debug("RSS-Bridge results",
			slog.String("website_url", websiteURL),
			slog.String("rssbridge_url", rssbridgeURL),
			slog.Int("nb_bridges", len(bridges)),
		)

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

	return tryWellKnownUrls(websiteURL, userAgent, cookie, username, password, fetchViaProxy, allowSelfSignedCertificates)
}

func parseWebPage(websiteURL string, data io.Reader) (Subscriptions, *locale.LocalizedErrorWrapper) {
	var subscriptions Subscriptions
	queries := map[string]string{
		"link[type='application/rss+xml']":   "rss",
		"link[type='application/atom+xml']":  "atom",
		"link[type='application/json']":      "json",
		"link[type='application/feed+json']": "json",
	}

	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
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

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to find subscriptions", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return websiteURL
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
	if docErr != nil {
		return websiteURL
	}

	if channelID, exists := doc.Find(`meta[itemprop="channelId"]`).First().Attr("content"); exists {
		return fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?channel_id=%s`, channelID)
	}

	return websiteURL
}

func tryWellKnownUrls(websiteURL, userAgent, cookie, username, password string, fetchViaProxy, allowSelfSignedCertificates bool) (Subscriptions, *locale.LocalizedErrorWrapper) {
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

			requestBuilder := fetcher.NewRequestBuilder()
			requestBuilder.WithUsernameAndPassword(username, password)
			requestBuilder.WithUserAgent(userAgent)
			requestBuilder.WithCookie(cookie)
			requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
			requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
			requestBuilder.UseProxy(fetchViaProxy)
			requestBuilder.IgnoreTLSErrors(allowSelfSignedCertificates)

			// Some websites redirects unknown URLs to the home page.
			// As result, the list of known URLs is returned to the subscription list.
			// We don't want the user to choose between invalid feed URLs.
			requestBuilder.WithoutRedirects()

			responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(fullURL))
			defer responseHandler.Close()

			if localizedError := responseHandler.LocalizedError(); localizedError != nil {
				continue
			}

			subscription := new(Subscription)
			subscription.Type = kind
			subscription.Title = fullURL
			subscription.URL = fullURL
			subscriptions = append(subscriptions, subscription)
		}
	}

	return subscriptions, nil
}
