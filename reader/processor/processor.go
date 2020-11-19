// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/model"
	"miniflux.app/reader/rewrite"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/reader/scraper"
	"miniflux.app/storage"

	"github.com/rylans/getlang"
)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed) {
	var filteredEntries model.Entries

	for _, entry := range feed.Entries {
		logger.Debug("[Processor] Processing entry %q from feed %q", entry.URL, feed.FeedURL)

		if isBlockedEntry(feed, entry) || !isAllowedEntry(feed, entry) {
			continue
		}

		if feed.Crawler {
			if !store.EntryURLExists(feed.ID, entry.URL) {
				logger.Debug("[Processor] Crawling entry %q from feed %q", entry.URL, feed.FeedURL)

				startTime := time.Now()
				content, scraperErr := scraper.Fetch(entry.URL, feed.ScraperRules, feed.UserAgent)

				if config.Opts.HasMetricsCollector() {
					status := "success"
					if scraperErr != nil {
						status = "error"
					}
					metric.ScraperRequestDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
				}

				if scraperErr != nil {
					logger.Error(`[Processor] Unable to crawl this entry: %q => %v`, entry.URL, scraperErr)
				} else if content != "" {
					// We replace the entry content only if the scraper doesn't return any error.
					entry.Content = content
				}
			}
		}

		entry.Content = rewrite.Rewriter(entry.URL, entry.Content, feed.RewriteRules)

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
		entry.Content = sanitizer.Sanitize(entry.URL, entry.Content)

		entry.ReadingTime = calculateReadingTime(entry.Content)
		filteredEntries = append(filteredEntries, entry)
	}

	feed.Entries = filteredEntries
}

func isBlockedEntry(feed *model.Feed, entry *model.Entry) bool {
	if feed.BlocklistRules != "" {
		match, _ := regexp.MatchString(feed.BlocklistRules, entry.Title)
		if match {
			logger.Debug("[Processor] Blocking entry %q from feed %q based on rule %q", entry.Title, feed.FeedURL, feed.BlocklistRules)
			return true
		}
	}
	return false
}

func isAllowedEntry(feed *model.Feed, entry *model.Entry) bool {
	if feed.KeeplistRules != "" {
		match, _ := regexp.MatchString(feed.KeeplistRules, entry.Title)
		if match {
			logger.Debug("[Processor] Allow entry %q from feed %q based on rule %q", entry.Title, feed.FeedURL, feed.KeeplistRules)
			return true
		}
		return false
	}
	return true
}

// ProcessEntryWebPage downloads the entry web page and apply rewrite rules.
func ProcessEntryWebPage(entry *model.Entry) error {
	startTime := time.Now()
	content, scraperErr := scraper.Fetch(entry.URL, entry.Feed.ScraperRules, entry.Feed.UserAgent)
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

	content = rewrite.Rewriter(entry.URL, content, entry.Feed.RewriteRules)
	content = sanitizer.Sanitize(entry.URL, content)

	if content != "" {
		entry.Content = content
		entry.ReadingTime = calculateReadingTime(content)
	}

	return nil
}

func calculateReadingTime(content string) int {
	sanitizedContent := sanitizer.StripTags(content)
	languageInfo := getlang.FromString(sanitizedContent)

	var timeToReadInt int
	if languageInfo.LanguageCode() == "ko" || languageInfo.LanguageCode() == "zh" || languageInfo.LanguageCode() == "jp" {
		timeToReadInt = int(math.Ceil(float64(utf8.RuneCountInString(sanitizedContent)) / 500))
	} else {
		nbOfWords := len(strings.Fields(sanitizedContent))
		timeToReadInt = int(math.Ceil(float64(nbOfWords) / 265))
	}

	return timeToReadInt
}
