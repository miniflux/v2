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
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
)

var (
	youtubeChannelRegex  = regexp.MustCompile(`youtube\.com/channel/(.*)$`)
	youtubeVideoRegex    = regexp.MustCompile(`youtube\.com/watch\?v=(.*)$`)
	youtubePlaylistRegex = regexp.MustCompile(`youtube\.com/playlist\?list=(.*)$`)
)

type SubscriptionFinder struct {
	requestBuilder   *fetcher.RequestBuilder
	feedDownloaded   bool
	feedResponseInfo *model.FeedCreationRequestFromSubscriptionDiscovery
}

func NewSubscriptionFinder(requestBuilder *fetcher.RequestBuilder) *SubscriptionFinder {
	return &SubscriptionFinder{
		requestBuilder: requestBuilder,
	}
}

func (f *SubscriptionFinder) IsFeedAlreadyDownloaded() bool {
	return f.feedDownloaded
}

func (f *SubscriptionFinder) FeedResponseInfo() *model.FeedCreationRequestFromSubscriptionDiscovery {
	return f.feedResponseInfo
}

func (f *SubscriptionFinder) FindSubscriptions(websiteURL, rssBridgeURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	responseHandler := fetcher.NewResponseHandler(f.requestBuilder.ExecuteRequest(websiteURL))
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

	f.feedResponseInfo = &model.FeedCreationRequestFromSubscriptionDiscovery{
		Content:      bytes.NewReader(responseBody),
		ETag:         responseHandler.ETag(),
		LastModified: responseHandler.LastModified(),
	}

	// Step 1) Check if the website URL is a feed.
	if feedFormat, _ := parser.DetectFeedFormat(f.feedResponseInfo.Content); feedFormat != parser.FormatUnknown {
		f.feedDownloaded = true
		return Subscriptions{NewSubscription(responseHandler.EffectiveURL(), responseHandler.EffectiveURL(), feedFormat)}, nil
	}

	// Step 2) Check if the website URL is a YouTube channel.
	slog.Debug("Try to detect feeds from YouTube channel page", slog.String("website_url", websiteURL))
	subscriptions, localizedError := f.FindSubscriptionsFromYouTubeChannelPage(websiteURL)
	if localizedError != nil {
		return nil, localizedError
	}

	if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from YouTube channel page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 3) Check if the website URL is a YouTube video.
	slog.Debug("Try to detect feeds from YouTube video page", slog.String("website_url", websiteURL))
	subscriptions, localizedError = f.FindSubscriptionsFromYouTubeVideoPage(websiteURL)
	if localizedError != nil {
		return nil, localizedError
	}

	if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from YouTube video page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 4) Check if the website URL is a YouTube playlist.
	slog.Debug("Try to detect feeds from YouTube playlist page", slog.String("website_url", websiteURL))
	subscriptions, localizedError = f.FindSubscriptionsFromYouTubePlaylistPage(websiteURL)
	if localizedError != nil {
		return nil, localizedError
	}

	if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from YouTube playlist page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 5) Parse web page to find feeds from HTML meta tags.
	slog.Debug("Try to detect feeds from HTML meta tags",
		slog.String("website_url", websiteURL),
		slog.String("content_type", responseHandler.ContentType()),
	)
	subscriptions, localizedError = f.FindSubscriptionsFromWebPage(websiteURL, responseHandler.ContentType(), bytes.NewReader(responseBody))
	if localizedError != nil {
		return nil, localizedError
	}

	if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from web page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 6) Check if the website URL can use RSS-Bridge.
	if rssBridgeURL != "" {
		slog.Debug("Try to detect feeds with RSS-Bridge", slog.String("website_url", websiteURL))
		subscriptions, localizedError := f.FindSubscriptionsFromRSSBridge(websiteURL, rssBridgeURL)
		if localizedError != nil {
			return nil, localizedError
		}

		if len(subscriptions) > 0 {
			slog.Debug("Subscriptions found from RSS-Bridge", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
			return subscriptions, nil
		}
	}

	// Step 7) Check if the website has a known feed URL.
	slog.Debug("Try to detect feeds from well-known URLs", slog.String("website_url", websiteURL))
	subscriptions, localizedError = f.FindSubscriptionsFromWellKnownURLs(websiteURL)
	if localizedError != nil {
		return nil, localizedError
	}

	if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found with well-known URLs", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	return nil, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromWebPage(websiteURL, contentType string, body io.Reader) (Subscriptions, *locale.LocalizedErrorWrapper) {
	queries := map[string]string{
		"link[type='application/rss+xml']":   parser.FormatRSS,
		"link[type='application/atom+xml']":  parser.FormatAtom,
		"link[type='application/json']":      parser.FormatJSON,
		"link[type='application/feed+json']": parser.FormatJSON,
	}

	htmlDocumentReader, err := charset.NewReader(body, contentType)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
	}

	doc, err := goquery.NewDocumentFromReader(htmlDocumentReader)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
	}

	var subscriptions Subscriptions
	subscriptionURLs := make(map[string]bool)
	for query, kind := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			subscription := new(Subscription)
			subscription.Type = kind

			if title, exists := s.Attr("title"); exists {
				subscription.Title = title
			}

			if feedURL, exists := s.Attr("href"); exists {
				if feedURL != "" {
					subscription.URL, err = urllib.AbsoluteURL(websiteURL, feedURL)
					if err != nil {
						return
					}
				}
			}

			if subscription.Title == "" {
				subscription.Title = subscription.URL
			}

			if subscription.URL != "" && !subscriptionURLs[subscription.URL] {
				subscriptionURLs[subscription.URL] = true
				subscriptions = append(subscriptions, subscription)
			}
		})
	}

	return subscriptions, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromWellKnownURLs(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	knownURLs := map[string]string{
		"atom.xml":  parser.FormatAtom,
		"feed.xml":  parser.FormatAtom,
		"feed/":     parser.FormatAtom,
		"rss.xml":   parser.FormatRSS,
		"rss/":      parser.FormatRSS,
		"index.rss": parser.FormatRSS,
		"index.xml": parser.FormatRSS,
		"feed.atom": parser.FormatAtom,
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

	var subscriptions Subscriptions
	for _, baseURL := range baseURLs {
		for knownURL, kind := range knownURLs {
			fullURL, err := urllib.AbsoluteURL(baseURL, knownURL)
			if err != nil {
				continue
			}

			// Some websites redirects unknown URLs to the home page.
			// As result, the list of known URLs is returned to the subscription list.
			// We don't want the user to choose between invalid feed URLs.
			f.requestBuilder.WithoutRedirects()

			responseHandler := fetcher.NewResponseHandler(f.requestBuilder.ExecuteRequest(fullURL))
			localizedError := responseHandler.LocalizedError()
			responseHandler.Close()

			if localizedError != nil {
				slog.Debug("Unable to subscribe", slog.String("fullURL", fullURL), slog.Any("error", localizedError.Error()))
				continue
			}

			subscriptions = append(subscriptions, &Subscription{
				Type:  kind,
				Title: fullURL,
				URL:   fullURL,
			})
		}
	}

	return subscriptions, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromRSSBridge(websiteURL, rssBridgeURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	slog.Debug("Trying to detect feeds using RSS-Bridge",
		slog.String("website_url", websiteURL),
		slog.String("rssbridge_url", rssBridgeURL),
	)

	bridges, err := rssbridge.DetectBridges(rssBridgeURL, websiteURL)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_detect_rssbridge", err)
	}

	slog.Debug("RSS-Bridge results",
		slog.String("website_url", websiteURL),
		slog.String("rssbridge_url", rssBridgeURL),
		slog.Int("nb_bridges", len(bridges)),
	)

	if len(bridges) == 0 {
		return nil, nil
	}

	subscriptions := make(Subscriptions, 0, len(bridges))
	for _, bridge := range bridges {
		subscriptions = append(subscriptions, &Subscription{
			Title: bridge.BridgeMeta.Name,
			URL:   bridge.URL,
			Type:  parser.FormatAtom,
		})
	}

	return subscriptions, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromYouTubeChannelPage(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	matches := youtubeChannelRegex.FindStringSubmatch(websiteURL)

	if len(matches) == 2 {
		feedURL := fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?channel_id=%s`, matches[1])
		return Subscriptions{NewSubscription(websiteURL, feedURL, parser.FormatAtom)}, nil
	}

	slog.Debug("This website is not a YouTube channel page, the regex doesn't match", slog.String("website_url", websiteURL))

	return nil, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromYouTubeVideoPage(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	if !youtubeVideoRegex.MatchString(websiteURL) {
		slog.Debug("This website is not a YouTube video page, the regex doesn't match", slog.String("website_url", websiteURL))
		return nil, nil
	}

	responseHandler := fetcher.NewResponseHandler(f.requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		return nil, localizedError
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
	if docErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(docErr, "error.unable_to_parse_html_document", docErr)
	}

	if channelID, exists := doc.Find(`meta[itemprop="channelId"]`).First().Attr("content"); exists {
		feedURL := fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?channel_id=%s`, channelID)
		return Subscriptions{NewSubscription(websiteURL, feedURL, parser.FormatAtom)}, nil
	}

	return nil, nil
}

func (f *SubscriptionFinder) FindSubscriptionsFromYouTubePlaylistPage(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	matches := youtubePlaylistRegex.FindStringSubmatch(websiteURL)

	if len(matches) == 2 {
		feedURL := fmt.Sprintf(`https://www.youtube.com/feeds/videos.xml?playlist_id=%s`, matches[1])
		return Subscriptions{NewSubscription(websiteURL, feedURL, parser.FormatAtom)}, nil
	}

	slog.Debug("This website is not a YouTube playlist page, the regex doesn't match", slog.String("website_url", websiteURL))

	return nil, nil
}
