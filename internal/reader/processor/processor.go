// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
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
			requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
			requestBuilder.UseProxy(feed.FetchViaProxy)
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

	feed.Entries = filteredEntries
}

func isBlockedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if user.BlockFilterEntryRules != "" {
		rules := strings.Split(user.BlockFilterEntryRules, "\n")
		for _, rule := range rules {
			parts := strings.SplitN(rule, "=", 2)

			var match bool
			switch parts[0] {
			case "EntryDate":
				datePattern := parts[1]
				match = isDateMatchingPattern(entry.Date, datePattern)
			case "EntryTitle":
				match, _ = regexp.MatchString(parts[1], entry.Title)
			case "EntryURL":
				match, _ = regexp.MatchString(parts[1], entry.URL)
			case "EntryCommentsURL":
				match, _ = regexp.MatchString(parts[1], entry.CommentsURL)
			case "EntryContent":
				match, _ = regexp.MatchString(parts[1], entry.Content)
			case "EntryAuthor":
				match, _ = regexp.MatchString(parts[1], entry.Author)
			case "EntryTag":
				containsTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
					match, _ = regexp.MatchString(parts[1], tag)
					return match
				})
				if containsTag {
					match = true
				}
			}

			if match {
				slog.Debug("Blocking entry based on rule",
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.String("rule", rule),
				)
				return true
			}
		}
	}

	if feed.BlocklistRules == "" {
		return false
	}

	compiledBlocklist, err := regexp.Compile(feed.BlocklistRules)
	if err != nil {
		slog.Debug("Failed on regexp compilation",
			slog.String("pattern", feed.BlocklistRules),
			slog.Any("error", err),
		)
		return false
	}

	containsBlockedTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledBlocklist.MatchString(tag)
	})

	if compiledBlocklist.MatchString(entry.URL) || compiledBlocklist.MatchString(entry.Title) || compiledBlocklist.MatchString(entry.Author) || containsBlockedTag {
		slog.Debug("Blocking entry based on rule",
			slog.String("entry_url", entry.URL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
			slog.String("rule", feed.BlocklistRules),
		)
		return true
	}

	return false
}

func isAllowedEntry(feed *model.Feed, entry *model.Entry, user *model.User) bool {
	if user.KeepFilterEntryRules != "" {
		rules := strings.Split(user.KeepFilterEntryRules, "\n")
		for _, rule := range rules {
			parts := strings.SplitN(rule, "=", 2)

			var match bool
			switch parts[0] {
			case "EntryDate":
				datePattern := parts[1]
				match = isDateMatchingPattern(entry.Date, datePattern)
			case "EntryTitle":
				match, _ = regexp.MatchString(parts[1], entry.Title)
			case "EntryURL":
				match, _ = regexp.MatchString(parts[1], entry.URL)
			case "EntryCommentsURL":
				match, _ = regexp.MatchString(parts[1], entry.CommentsURL)
			case "EntryContent":
				match, _ = regexp.MatchString(parts[1], entry.Content)
			case "EntryAuthor":
				match, _ = regexp.MatchString(parts[1], entry.Author)
			case "EntryTag":
				containsTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
					match, _ = regexp.MatchString(parts[1], tag)
					return match
				})
				if containsTag {
					match = true
				}
			}

			if match {
				slog.Debug("Allowing entry based on rule",
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.String("rule", rule),
				)
				return true
			}
		}
		return false
	}

	if feed.KeeplistRules == "" {
		return true
	}

	compiledKeeplist, err := regexp.Compile(feed.KeeplistRules)
	if err != nil {
		slog.Debug("Failed on regexp compilation",
			slog.String("pattern", feed.KeeplistRules),
			slog.Any("error", err),
		)
		return false
	}
	containsAllowedTag := slices.ContainsFunc(entry.Tags, func(tag string) bool {
		return compiledKeeplist.MatchString(tag)
	})

	if compiledKeeplist.MatchString(entry.URL) || compiledKeeplist.MatchString(entry.Title) || compiledKeeplist.MatchString(entry.Author) || containsAllowedTag {
		slog.Debug("Allow entry based on rule",
			slog.String("entry_url", entry.URL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
			slog.String("rule", feed.KeeplistRules),
		)
		return true
	}
	return false
}

// ProcessEntryWebPage downloads the entry web page and apply rewrite rules.
func ProcessEntryWebPage(feed *model.Feed, entry *model.Entry, user *model.User) error {
	startTime := time.Now()
	rewrittenEntryURL := rewriteEntryURL(feed, entry)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUserAgent(feed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(feed.FetchViaProxy)
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

func updateEntryReadingTime(store *storage.Storage, feed *model.Feed, entry *model.Entry, entryIsNew bool, user *model.User) {
	if !user.ShowReadingTime {
		slog.Debug("Skip reading time estimation for this user", slog.Int64("user_id", user.ID))
		return
	}

	if shouldFetchYouTubeWatchTime(entry) {
		if entryIsNew {
			watchTime, err := fetchYouTubeWatchTime(entry.URL)
			if err != nil {
				slog.Warn("Unable to fetch YouTube watch time",
					slog.Int64("user_id", user.ID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", err),
				)
			}
			entry.ReadingTime = watchTime
		} else {
			entry.ReadingTime = store.GetReadTime(feed.ID, entry.Hash)
		}
	}

	if shouldFetchNebulaWatchTime(entry) {
		if entryIsNew {
			watchTime, err := fetchNebulaWatchTime(entry.URL)
			if err != nil {
				slog.Warn("Unable to fetch Nebula watch time",
					slog.Int64("user_id", user.ID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", err),
				)
			}
			entry.ReadingTime = watchTime
		} else {
			entry.ReadingTime = store.GetReadTime(feed.ID, entry.Hash)
		}
	}

	if shouldFetchOdyseeWatchTime(entry) {
		if entryIsNew {
			watchTime, err := fetchOdyseeWatchTime(entry.URL)
			if err != nil {
				slog.Warn("Unable to fetch Odysee watch time",
					slog.Int64("user_id", user.ID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", err),
				)
			}
			entry.ReadingTime = watchTime
		} else {
			entry.ReadingTime = store.GetReadTime(feed.ID, entry.Hash)
		}
	}

	if shouldFetchBilibiliWatchTime(entry) {
		if entryIsNew {
			watchTime, err := fetchBilibiliWatchTime(entry.URL)
			if err != nil {
				slog.Warn("Unable to fetch Bilibili watch time",
					slog.Int64("user_id", user.ID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", err),
				)
			}
			entry.ReadingTime = watchTime
		} else {
			entry.ReadingTime = store.GetReadTime(feed.ID, entry.Hash)
		}
	}

	// Handle YT error case and non-YT entries.
	if entry.ReadingTime == 0 {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}
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

func isDateMatchingPattern(entryDate time.Time, pattern string) bool {
	if pattern == "future" {
		return entryDate.After(time.Now())
	}

	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) != 2 {
		return false
	}

	operator := parts[0]
	dateStr := parts[1]

	switch operator {
	case "before":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.Before(targetDate)
	case "after":
		targetDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return false
		}
		return entryDate.After(targetDate)
	case "between":
		dates := strings.Split(dateStr, ",")
		if len(dates) != 2 {
			return false
		}
		startDate, err1 := time.Parse("2006-01-02", dates[0])
		endDate, err2 := time.Parse("2006-01-02", dates[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return entryDate.After(startDate) && entryDate.Before(endDate)
	}
	return false
}
