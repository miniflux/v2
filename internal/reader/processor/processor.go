// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"log/slog"
	"net/url"
	"slices"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/filter"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/reader/scraper"
	"miniflux.app/v2/internal/reader/urlcleaner"
	"miniflux.app/v2/internal/storage"
)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, userID int64, forceRefresh bool) {
	var filteredEntries model.Entries

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		slog.Error("Database error", slog.Any("error", storeErr))
		return
	}

	// The errors are handled in RemoveTrackingParameters.
	parsedFeedURL, _ := url.Parse(feed.FeedURL)
	parsedSiteURL, _ := url.Parse(feed.SiteURL)

	// Process older entries first
	for _, entry := range slices.Backward(feed.Entries) {
		slog.Debug("Processing entry",
			slog.Int64("user_id", user.ID),
			slog.String("entry_url", entry.URL),
			slog.String("entry_hash", entry.Hash),
			slog.String("entry_title", entry.Title),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
		)

		if filter.IsBlockedEntry(feed, entry, user) || !filter.IsAllowedEntry(feed, entry, user) {
			continue
		}

		parsedInputUrl, _ := url.Parse(entry.URL)
		if cleanedURL, err := urlcleaner.RemoveTrackingParameters(parsedFeedURL, parsedSiteURL, parsedInputUrl); err == nil {
			entry.URL = cleanedURL
		}

		webpageBaseURL := ""
		entry.URL = rewrite.RewriteEntryURL(feed, entry)
		entryIsNew := store.IsNewEntry(feed.ID, entry.Hash)
		if feed.Crawler && (entryIsNew || forceRefresh) {
			slog.Debug("Scraping entry",
				slog.Int64("user_id", user.ID),
				slog.String("entry_url", entry.URL),
				slog.String("entry_hash", entry.Hash),
				slog.String("entry_title", entry.Title),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.Bool("entry_is_new", entryIsNew),
				slog.Bool("force_refresh", forceRefresh),
			)

			startTime := time.Now()

			requestBuilder := fetcher.NewRequestBuilder()
			requestBuilder.WithUserAgent(feed.UserAgent, config.Opts.HTTPClientUserAgent())
			requestBuilder.WithCookie(feed.Cookie)
			requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
			requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
			requestBuilder.WithCustomFeedProxyURL(feed.ProxyURL)
			requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
			requestBuilder.UseCustomApplicationProxyURL(feed.FetchViaProxy)
			requestBuilder.IgnoreTLSErrors(feed.AllowSelfSignedCertificates)
			requestBuilder.DisableHTTP2(feed.DisableHTTP2)

			scrapedPageBaseURL, extractedContent, scraperErr := scraper.ScrapeWebsite(
				requestBuilder,
				entry.URL,
				feed.ScraperRules,
			)

			if scrapedPageBaseURL != "" {
				webpageBaseURL = scrapedPageBaseURL
			}

			if config.Opts.HasMetricsCollector() {
				status := "success"
				if scraperErr != nil {
					status = "error"
				}
				metric.ScraperRequestDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
			}

			if scraperErr != nil {
				slog.Warn("Unable to scrape entry",
					slog.Int64("user_id", user.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", scraperErr),
				)
			} else if extractedContent != "" {
				// We replace the entry content only if the scraper doesn't return any error.
				entry.Content = minifyContent(extractedContent)
			}
		}

		rewrite.ApplyContentRewriteRules(entry, feed.RewriteRules)

		if webpageBaseURL == "" {
			webpageBaseURL = entry.URL
		}

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered out.
		entry.Content = sanitizer.SanitizeHTML(webpageBaseURL, entry.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})

		updateEntryReadingTime(store, feed, entry, entryIsNew, user)

		filteredEntries = append(filteredEntries, entry)
	}

	if user.ShowReadingTime && shouldFetchYouTubeWatchTimeInBulk() {
		fetchYouTubeWatchTimeInBulk(filteredEntries)
	}

	feed.Entries = filteredEntries
}

// ProcessEntryWebPage downloads the entry web page and apply rewrite rules.
func ProcessEntryWebPage(feed *model.Feed, entry *model.Entry, user *model.User) error {
	startTime := time.Now()
	entry.URL = rewrite.RewriteEntryURL(feed, entry)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUserAgent(feed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(feed.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(feed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feed.DisableHTTP2)

	webpageBaseURL, extractedContent, scraperErr := scraper.ScrapeWebsite(
		requestBuilder,
		entry.URL,
		feed.ScraperRules,
	)

	if config.Opts.HasMetricsCollector() {
		status := "success"
		if scraperErr != nil {
			status = "error"
		}
		metric.ScraperRequestDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
	}

	if scraperErr != nil {
		return scraperErr
	}

	if extractedContent != "" {
		entry.Content = minifyContent(extractedContent)
		if user.ShowReadingTime {
			entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
		}
	}

	rewrite.ApplyContentRewriteRules(entry, entry.Feed.RewriteRules)
	entry.Content = sanitizer.SanitizeHTML(webpageBaseURL, entry.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})

	return nil
}
