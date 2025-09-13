// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/v2/internal/reader/subscription"

import (
	"bytes"
	"io"
	"log/slog"
	"net/url"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/integration/rssbridge"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

type subscriptionFinder struct {
	requestBuilder   *fetcher.RequestBuilder
	feedDownloaded   bool
	feedResponseInfo *model.FeedCreationRequestFromSubscriptionDiscovery
}

func NewSubscriptionFinder(requestBuilder *fetcher.RequestBuilder) *subscriptionFinder {
	return &subscriptionFinder{
		requestBuilder: requestBuilder,
	}
}

func (f *subscriptionFinder) IsFeedAlreadyDownloaded() bool {
	return f.feedDownloaded
}

func (f *subscriptionFinder) FeedResponseInfo() *model.FeedCreationRequestFromSubscriptionDiscovery {
	return f.feedResponseInfo
}

func (f *subscriptionFinder) FindSubscriptions(websiteURL, rssBridgeURL string, rssBridgeToken string) (Subscriptions, *locale.LocalizedErrorWrapper) {
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

	// Step 1) Check if the website URL is already a feed.
	if feedFormat, _ := parser.DetectFeedFormat(f.feedResponseInfo.Content); feedFormat != parser.FormatUnknown {
		f.feedDownloaded = true
		return Subscriptions{NewSubscription(responseHandler.EffectiveURL(), responseHandler.EffectiveURL(), feedFormat)}, nil
	}

	// Step 2) Check if the website URL is a YouTube channel.
	slog.Debug("Try to detect feeds for a YouTube page", slog.String("website_url", websiteURL))
	if subscriptions, localizedError := f.findSubscriptionsFromYouTube(websiteURL); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from YouTube page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 3) Parse web page to find feeds from HTML meta tags.
	slog.Debug("Try to detect feeds from HTML meta tags",
		slog.String("website_url", websiteURL),
		slog.String("content_type", responseHandler.ContentType()),
	)
	if subscriptions, localizedError := f.findSubscriptionsFromWebPage(websiteURL, responseHandler.ContentType(), bytes.NewReader(responseBody)); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from web page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 4) Check if the website URL can use RSS-Bridge.
	if rssBridgeURL != "" {
		slog.Debug("Try to detect feeds with RSS-Bridge", slog.String("website_url", websiteURL))
		if subscriptions, localizedError := f.findSubscriptionsFromRSSBridge(websiteURL, rssBridgeURL, rssBridgeToken); localizedError != nil {
			return nil, localizedError
		} else if len(subscriptions) > 0 {
			slog.Debug("Subscriptions found from RSS-Bridge", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
			return subscriptions, nil
		}
	}

	// Step 5) Check if the website has a known feed URL.
	slog.Debug("Try to detect feeds from well-known URLs", slog.String("website_url", websiteURL))
	if subscriptions, localizedError := f.findSubscriptionsFromWellKnownURLs(websiteURL); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found with well-known URLs", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	return nil, nil
}

func (f *subscriptionFinder) findSubscriptionsFromWebPage(websiteURL, contentType string, body io.Reader) (Subscriptions, *locale.LocalizedErrorWrapper) {
	queries := map[string]string{
		"link[type='application/rss+xml']":                                  parser.FormatRSS,
		"link[type='application/atom+xml']":                                 parser.FormatAtom,
		"link[type='application/json'], link[type='application/feed+json']": parser.FormatJSON,
	}

	htmlDocumentReader, err := encoding.NewCharsetReader(body, contentType)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
	}

	doc, err := goquery.NewDocumentFromReader(htmlDocumentReader)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
	}

	if hrefValue, exists := doc.FindMatcher(goquery.Single("head base")).Attr("href"); exists {
		hrefValue = strings.TrimSpace(hrefValue)
		if urllib.IsAbsoluteURL(hrefValue) {
			websiteURL = hrefValue
		}
	}

	var subscriptions Subscriptions
	subscriptionURLs := make(map[string]bool)
	for query, kind := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			subscription := new(subscription)
			subscription.Type = kind

			if feedURL, exists := s.Attr("href"); exists && feedURL != "" {
				subscription.URL, err = urllib.AbsoluteURL(websiteURL, feedURL)
				if err != nil {
					return
				}
			} else {
				return // without an url, there can be no subscription.
			}

			if title, exists := s.Attr("title"); exists {
				subscription.Title = title
			}

			if subscription.Title == "" {
				subscription.Title = subscription.URL
			}

			if !subscriptionURLs[subscription.URL] {
				subscriptionURLs[subscription.URL] = true
				subscriptions = append(subscriptions, subscription)
			}
		})
	}

	return subscriptions, nil
}

func (f *subscriptionFinder) findSubscriptionsFromWellKnownURLs(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	knownURLs := map[string]string{
		"atom.xml":     parser.FormatAtom,
		"feed.atom":    parser.FormatAtom,
		"feed.xml":     parser.FormatAtom,
		"feed/":        parser.FormatAtom,
		"index.rss":    parser.FormatRSS,
		"index.xml":    parser.FormatRSS,
		"rss.xml":      parser.FormatRSS,
		"rss/":         parser.FormatRSS,
		"rss/feed.xml": parser.FormatRSS,
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

			// Do not add redirections to the possible list of subscriptions to avoid confusion.
			if responseHandler.IsRedirect() {
				slog.Debug("Ignore URL redirection during feed discovery", slog.String("fullURL", fullURL))
				continue
			}

			if localizedError != nil {
				slog.Debug("Ignore invalid feed URL during feed discovery",
					slog.String("fullURL", fullURL),
					slog.Any("error", localizedError.Error()),
				)
				continue
			}

			subscriptions = append(subscriptions, &subscription{
				Type:  kind,
				Title: fullURL,
				URL:   fullURL,
			})
		}
	}

	return subscriptions, nil
}

func (f *subscriptionFinder) findSubscriptionsFromRSSBridge(websiteURL, rssBridgeURL string, rssBridgeToken string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	slog.Debug("Trying to detect feeds using RSS-Bridge",
		slog.String("website_url", websiteURL),
		slog.String("rssbridge_url", rssBridgeURL),
		slog.String("rssbridge_token", rssBridgeToken),
	)

	bridges, err := rssbridge.DetectBridges(rssBridgeURL, rssBridgeToken, websiteURL)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_detect_rssbridge", err)
	}

	slog.Debug("RSS-Bridge results",
		slog.String("website_url", websiteURL),
		slog.String("rssbridge_url", rssBridgeURL),
		slog.String("rssbridge_token", rssBridgeToken),
		slog.Int("nb_bridges", len(bridges)),
	)

	if len(bridges) == 0 {
		return nil, nil
	}

	subscriptions := make(Subscriptions, 0, len(bridges))
	for _, bridge := range bridges {
		subscriptions = append(subscriptions, &subscription{
			Title: bridge.BridgeMeta.Name,
			URL:   bridge.URL,
			Type:  parser.FormatAtom,
		})
	}

	return subscriptions, nil
}

func (f *subscriptionFinder) findSubscriptionsFromYouTube(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	decodedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.invalid_site_url", err)
	}

	if !strings.HasSuffix(decodedURL.Host, "youtube.com") {
		slog.Debug("YouTube feed discovery skipped: not a YouTube domain", slog.String("website_url", websiteURL))
		return nil, nil
	}
	if _, channelID, found := strings.Cut(decodedURL.Path, "channel/"); found {
		feedURL := "https://www.youtube.com/feeds/videos.xml?channel_id=" + channelID
		return Subscriptions{NewSubscription(decodedURL.String(), feedURL, parser.FormatAtom)}, nil
	}

	if strings.HasPrefix(decodedURL.Path, "/watch") || strings.HasPrefix(decodedURL.Path, "/playlist") {
		if playlistID := decodedURL.Query().Get("list"); playlistID != "" {
			feedURL := "https://www.youtube.com/feeds/videos.xml?playlist_id=" + playlistID
			return Subscriptions{NewSubscription(decodedURL.String(), feedURL, parser.FormatAtom)}, nil
		}
	}

	return nil, nil
}
