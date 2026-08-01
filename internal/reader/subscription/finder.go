// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package subscription // import "miniflux.app/v2/internal/reader/subscription"

import (
	"bytes"
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

	// It's not a feed, so we have to process its HTML.
	doc, err := parseHTMLDocument(responseHandler.ContentType(), responseBody)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.unable_to_parse_html_document", err)
	}
	baseURL := getBaseURL(websiteURL, doc)

	// Step 2) Find the canonical URL of the website.
	slog.Debug("Try to find the canonical URL of the website", slog.String("website_url", websiteURL))
	websiteURL = f.findCanonicalURL(websiteURL, baseURL, doc)

	// Step 3) Check if the website URL is a YouTube channel.
	slog.Debug("Try to detect feeds for a YouTube page", slog.String("website_url", websiteURL))
	if subscriptions, localizedError := f.findSubscriptionsFromYouTube(websiteURL); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from YouTube page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 4) Check if the website URL is a GitHub page.
	slog.Debug("Try to detect feeds for a GitHub page", slog.String("website_url", websiteURL))
	if subscriptions, localizedError := f.findSubscriptionsFromGitHub(websiteURL); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from GitHub page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 5) Parse web page to find feeds from HTML meta tags.
	slog.Debug("Try to detect feeds from HTML meta tags",
		slog.String("website_url", websiteURL),
		slog.String("content_type", responseHandler.ContentType()),
	)

	if subscriptions, localizedError := f.findSubscriptionsFromWebPage(baseURL, doc); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found from web page", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	// Step 6) Check if the website URL can use RSS-Bridge.
	if rssBridgeURL != "" {
		slog.Debug("Try to detect feeds with RSS-Bridge", slog.String("website_url", websiteURL))
		if subscriptions, localizedError := f.findSubscriptionsFromRSSBridge(websiteURL, rssBridgeURL, rssBridgeToken); localizedError != nil {
			return nil, localizedError
		} else if len(subscriptions) > 0 {
			slog.Debug("Subscriptions found from RSS-Bridge", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
			return subscriptions, nil
		}
	}

	// Step 7) Check if the website has a known feed URL.
	slog.Debug("Try to detect feeds from well-known URLs", slog.String("website_url", websiteURL))
	if subscriptions, localizedError := f.findSubscriptionsFromWellKnownURLs(websiteURL); localizedError != nil {
		return nil, localizedError
	} else if len(subscriptions) > 0 {
		slog.Debug("Subscriptions found with well-known URLs", slog.String("website_url", websiteURL), slog.Any("subscriptions", subscriptions))
		return subscriptions, nil
	}

	return nil, nil
}

func (f *subscriptionFinder) findSubscriptionsFromWebPage(websiteURL string, doc *goquery.Document) (Subscriptions, *locale.LocalizedErrorWrapper) {
	var subscriptions Subscriptions
	// There are 4 possible feed formats
	subscriptionURLs := make(map[string]bool, 4)

	// Single DOM walk over every <link> with a type attribute, then dispatch on
	// the MIME type. This is better than doing a separate goquery.Find pass per
	// type.
	doc.Find("link[type]").Each(func(_ int, s *goquery.Selection) {
		typeAttr, _ := s.Attr("type")
		var feedFormat string
		switch typeAttr {
		case "application/rss+xml":
			feedFormat = parser.FormatRSS
		case "application/atom+xml":
			feedFormat = parser.FormatAtom
		case "application/feed+json":
			feedFormat = parser.FormatJSON
		case "application/json":
			// Ignore JSON feed URLs that contain "/wp-json/" to avoid confusion
			// with WordPress REST API endpoints.
			if href, _ := s.Attr("href"); strings.Contains(href, "/wp-json/") {
				return
			}
			feedFormat = parser.FormatJSON
		default:
			return
		}

		feedURL, _ := s.Attr("href")
		if feedURL == "" {
			return // without an url, there can be no subscription.
		}

		absoluteURL, err := urllib.ResolveToAbsoluteURL(websiteURL, feedURL)
		if err != nil {
			return
		}

		if subscriptionURLs[absoluteURL] {
			return
		}
		subscriptionURLs[absoluteURL] = true

		title, _ := s.Attr("title")
		if title == "" {
			title = absoluteURL
		}

		subscriptions = append(subscriptions, &subscription{
			Type:  feedFormat,
			Title: title,
			URL:   absoluteURL,
		})
	})

	return subscriptions, nil
}

func (f *subscriptionFinder) findSubscriptionsFromWellKnownURLs(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	type pair struct{ path, format string }
	knownURLs := []pair{
		{"atom.xml", parser.FormatAtom},
		{"feed.atom", parser.FormatAtom},
		{"feed.xml", parser.FormatAtom},
		{"feed/", parser.FormatAtom},
		{"index.rss", parser.FormatRSS},
		{"index.xml", parser.FormatRSS},
		{"rss.xml", parser.FormatRSS},
		{"rss/", parser.FormatRSS},
		{"rss/feed.xml", parser.FormatRSS},
	}

	websiteURLRoot := urllib.RootURL(websiteURL)
	baseURLs := []string{
		// Look for knownURLs in the root.
		websiteURLRoot,
	}

	// Look for knownURLs in current subdirectory, such as 'example.com/blog/'.
	websiteURL, _ = urllib.ResolveToAbsoluteURL(websiteURL, "./")
	if websiteURL != websiteURLRoot {
		baseURLs = append(baseURLs, websiteURL)
	}

	var subscriptions Subscriptions
	for _, baseURL := range baseURLs {
		for _, known := range knownURLs {
			fullURL, err := urllib.ResolveToAbsoluteURL(baseURL, known.path)
			if err != nil {
				continue
			}

			// Some websites redirects unknown URLs to the home page.
			// As result, the list of known URLs is returned to the subscription list.
			// We don't want the user to choose between invalid feed URLs.
			//
			// Probe each known URL on its own builder so disabling redirects
			// here doesn't leak into the finder's other requests.
			requestBuilder := f.requestBuilder.Clone().WithoutRedirects()

			responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(fullURL))
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
				Type:  known.format,
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
	playlistPrefixes := []struct {
		prefix string
		title  string
	}{
		{"UULF", "Videos"},
		{"UUSH", "Short videos"},
		{"UULV", "Live streams"},
	}

	decodedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.invalid_site_url", err)
	}

	if !strings.HasSuffix(decodedURL.Host, "youtube.com") {
		slog.Debug("YouTube feed discovery skipped: not a YouTube domain", slog.String("website_url", websiteURL))
		return nil, nil
	}

	if _, baseID, found := strings.Cut(decodedURL.Path, "channel/UC"); found {
		var subscriptions Subscriptions

		channelFeedURL := "https://www.youtube.com/feeds/videos.xml?channel_id=UC" + baseID
		subscriptions = append(subscriptions, NewSubscription("Channel", channelFeedURL, parser.FormatAtom))

		for _, playlist := range playlistPrefixes {
			playlistFeedURL := "https://www.youtube.com/feeds/videos.xml?playlist_id=" + playlist.prefix + baseID
			subscriptions = append(subscriptions, NewSubscription(playlist.title, playlistFeedURL, parser.FormatAtom))
		}

		return subscriptions, nil
	}

	if strings.HasPrefix(decodedURL.Path, "/watch") || strings.HasPrefix(decodedURL.Path, "/playlist") {
		if playlistID := decodedURL.Query().Get("list"); playlistID != "" {
			feedURL := "https://www.youtube.com/feeds/videos.xml?playlist_id=" + playlistID
			return Subscriptions{NewSubscription(decodedURL.String(), feedURL, parser.FormatAtom)}, nil
		}
	}

	return nil, nil
}

// findSubscriptionsFromGitHub builds the Atom feed URLs that GitHub exposes for
// user/organization profiles and repositories. These feeds are not advertised
// in the HTML of the pages, so they cannot be discovered through the usual
// mechanisms.
func (f *subscriptionFinder) findSubscriptionsFromGitHub(websiteURL string) (Subscriptions, *locale.LocalizedErrorWrapper) {
	decodedURL, err := url.Parse(websiteURL)
	if err != nil {
		return nil, locale.NewLocalizedErrorWrapper(err, "error.invalid_site_url", err)
	}

	if decodedURL.Host != "github.com" && decodedURL.Host != "www.github.com" {
		slog.Debug("GitHub feed discovery skipped: not a GitHub domain", slog.String("website_url", websiteURL))
		return nil, nil
	}

	// Split the path into at most three segments to determine the kind of page.
	// We only care about the first two, so there is no need to keep splitting.
	path := strings.Trim(decodedURL.Path, "/")
	if path == "" {
		// The root page (https://github.com/) has no feed.
		return nil, nil
	}
	segments := strings.SplitN(path, "/", 3)

	switch len(segments) {
	case 1:
		// User or organization profile: https://github.com/<user>
		// The activity feed lives at https://github.com/<user>.atom
		user := segments[0]
		feedURL := "https://github.com/" + user + ".atom"
		return Subscriptions{NewSubscription(user, feedURL, parser.FormatAtom)}, nil
	case 2:
		// Repository: https://github.com/<owner>/<repo>
		// GitHub exposes commits, releases and tags Atom feeds for repositories.
		repoPath := "https://github.com/" + segments[0] + "/" + segments[1] + "/"
		return Subscriptions{
			NewSubscription("Commits", repoPath+"commits.atom", parser.FormatAtom),
			NewSubscription("Releases", repoPath+"releases.atom", parser.FormatAtom),
			NewSubscription("Tags", repoPath+"tags.atom", parser.FormatAtom),
		}, nil
	default:
		// Deeper paths that don't map to a user profile or a repository (e.g.
		// https://github.com/owner/repo/tree/branch/dir) have no dedicated feed.
		return nil, nil
	}
}

// findCanonicalURL extracts the canonical URL from the HTML <link rel="canonical"> tag.
// Returns the canonical URL if found, otherwise returns the effective URL.
func (f *subscriptionFinder) findCanonicalURL(effectiveURL, baseURL string, doc *goquery.Document) string {
	canonicalHref, exists := doc.FindMatcher(goquery.Single("head link[rel='canonical' i]")).Attr("href")
	if !exists {
		return effectiveURL
	}
	canonicalHref = strings.TrimSpace(canonicalHref)
	if canonicalHref == "" {
		return effectiveURL
	}

	canonicalURL, err := urllib.ResolveToAbsoluteURL(baseURL, canonicalHref)
	if err != nil {
		return effectiveURL
	}

	return canonicalURL
}

// getBaseURL returns the url specified in the <base> tag, and `websiteURL` otherwise.
func getBaseURL(websiteURL string, doc *goquery.Document) string {
	baseURL := websiteURL
	if hrefValue, exists := doc.FindMatcher(goquery.Single("head base")).Attr("href"); exists {
		hrefValue = strings.TrimSpace(hrefValue)
		if urllib.IsAbsoluteURL(hrefValue) {
			baseURL = hrefValue
		}
	}
	return baseURL
}

func parseHTMLDocument(contentType string, body []byte) (*goquery.Document, error) {
	htmlDocumentReader, err := encoding.NewCharsetReaderFromBytes(body, contentType)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(htmlDocumentReader)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
