// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"log/slog"
	"regexp"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/reader/scraper"
	"miniflux.app/v2/internal/reader/urlcleaner"
	"miniflux.app/v2/internal/storage"
)

var customReplaceRuleRegex = regexp.MustCompile(`rewrite\("([^"]+)"\|"([^"]+)"\)`)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, userID int64, forceRefresh bool) {
	var filteredEntries model.Entries

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		slog.Error("Database error", slog.Any("error", storeErr))
		return
	}

	// Process older entries first
	for i := len(feed.Entries) - 1; i >= 0; i-- {
		entry := feed.Entries[i]

		slog.Debug("Processing entry",
			slog.Int64("user_id", user.ID),
			slog.String("entry_url", entry.URL),
			slog.String("entry_hash", entry.Hash),
			slog.String("entry_title", entry.Title),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
		)
		if isBlockedEntry(feed, entry, user) || !isAllowedEntry(feed, entry, user) || !isRecentEntry(entry) {
			continue
		}

		if cleanedURL, err := urlcleaner.RemoveTrackingParameters(entry.URL); err == nil {
			entry.URL = cleanedURL
		}

		pageBaseURL := ""
		rewrittenURL := rewriteEntryURL(feed, entry)
		entry.URL = rewrittenURL
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
				slog.String("rewritten_url", rewrittenURL),
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
				rewrittenURL,
				feed.ScraperRules,
			)

			if scrapedPageBaseURL != "" {
				pageBaseURL = scrapedPageBaseURL
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
				entry.Content = minifyEntryContent(extractedContent)
			}
		}

		rewrite.Rewriter(rewrittenURL, entry, feed.RewriteRules)

		if pageBaseURL == "" {
			pageBaseURL = rewrittenURL
		}

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered out.
		entry.Content = sanitizer.Sanitize(pageBaseURL, entry.Content)

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
	rewrittenEntryURL := rewriteEntryURL(feed, entry)

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

	pageBaseURL, extractedContent, scraperErr := scraper.ScrapeWebsite(
		requestBuilder,
		rewrittenEntryURL,
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
		entry.Content = minifyEntryContent(extractedContent)
		if user.ShowReadingTime {
			entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
		}
	}

	rewrite.Rewriter(rewrittenEntryURL, entry, entry.Feed.RewriteRules)
	entry.Content = sanitizer.Sanitize(pageBaseURL, entry.Content)

	return nil
}

func rewriteEntryURL(feed *model.Feed, entry *model.Entry) string {
	var rewrittenURL = entry.URL
	if feed.UrlRewriteRules != "" {
		parts := customReplaceRuleRegex.FindStringSubmatch(feed.UrlRewriteRules)

		if len(parts) >= 3 {
			re, err := regexp.Compile(parts[1])
			if err != nil {
				slog.Error("Failed on regexp compilation",
					slog.String("url_rewrite_rules", feed.UrlRewriteRules),
					slog.Any("error", err),
				)
				return rewrittenURL
			}
			rewrittenURL = re.ReplaceAllString(entry.URL, parts[2])
			slog.Debug("Rewriting entry URL",
				slog.String("original_entry_url", entry.URL),
				slog.String("rewritten_entry_url", rewrittenURL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
			)
		} else {
			slog.Debug("Cannot find search and replace terms for replace rule",
				slog.String("original_entry_url", entry.URL),
				slog.String("rewritten_entry_url", rewrittenURL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("url_rewrite_rules", feed.UrlRewriteRules),
			)
		}
	}

	return rewrittenURL
}

func isRecentEntry(entry *model.Entry) bool {
	if config.Opts.FilterEntryMaxAgeDays() == 0 || entry.Date.After(time.Now().AddDate(0, 0, -config.Opts.FilterEntryMaxAgeDays())) {
		return true
	}
	return false
}

func minifyEntryContent(entryContent string) string {
	m := minify.New()

	// Options required to avoid breaking the HTML content.
	m.Add("text/html", &html.Minifier{
		KeepEndTags: true,
		KeepQuotes:  true,
	})

	if minifiedHTML, err := m.String("text/html", entryContent); err == nil {
		entryContent = minifiedHTML
	}

	return entryContent
}
