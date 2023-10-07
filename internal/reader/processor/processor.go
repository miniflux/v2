// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/browser"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/reader/scraper"
	"miniflux.app/v2/internal/storage"

	"github.com/PuerkitoBio/goquery"
)

var (
	youtubeRegex           = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
	odyseeRegex            = regexp.MustCompile(`^https://odysee\.com`)
	iso8601Regex           = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	customReplaceRuleRegex = regexp.MustCompile(`rewrite\("(.*)"\|"(.*)"\)`)
)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, user *model.User, forceRefresh bool) {
	var filteredEntries model.Entries

	// Process older entries first
	for i := len(feed.Entries) - 1; i >= 0; i-- {
		entry := feed.Entries[i]

		slog.Debug("Processing entry",
			slog.Int64("user_id", user.ID),
			slog.Int64("entry_id", entry.ID),
			slog.String("entry_url", entry.URL),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
		)

		if isBlockedEntry(feed, entry) || !isAllowedEntry(feed, entry) {
			continue
		}

		url := getUrlFromEntry(feed, entry)
		entryIsNew := !store.EntryURLExists(feed.ID, entry.URL)
		if feed.Crawler && (entryIsNew || forceRefresh) {
			slog.Debug("Scraping entry",
				slog.Int64("user_id", user.ID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
			)

			startTime := time.Now()
			content, scraperErr := scraper.Fetch(
				url,
				feed.ScraperRules,
				feed.UserAgent,
				feed.Cookie,
				feed.AllowSelfSignedCertificates,
				feed.FetchViaProxy,
			)

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
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", scraperErr),
				)
			} else if content != "" {
				// We replace the entry content only if the scraper doesn't return any error.
				entry.Content = content
			}
		}

		rewrite.Rewriter(url, entry, feed.RewriteRules)

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
		entry.Content = sanitizer.Sanitize(url, entry.Content)

		updateEntryReadingTime(store, feed, entry, entryIsNew, user)
		filteredEntries = append(filteredEntries, entry)
	}

	feed.Entries = filteredEntries
}

func isBlockedEntry(feed *model.Feed, entry *model.Entry) bool {
	if feed.BlocklistRules != "" {
		match, _ := regexp.MatchString(feed.BlocklistRules, entry.Title)
		if match {
			slog.Debug("Blocking entry based on rule",
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("rule", feed.BlocklistRules),
			)
			return true
		}
	}
	return false
}

func isAllowedEntry(feed *model.Feed, entry *model.Entry) bool {
	if feed.KeeplistRules != "" {
		match, _ := regexp.MatchString(feed.KeeplistRules, entry.Title)
		if match {
			slog.Debug("Allow entry based on rule",
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("rule", feed.KeeplistRules),
			)
			return true
		}
		return false
	}
	return true
}

// ProcessEntryWebPage downloads the entry web page and apply rewrite rules.
func ProcessEntryWebPage(feed *model.Feed, entry *model.Entry, user *model.User) error {
	startTime := time.Now()
	url := getUrlFromEntry(feed, entry)

	content, scraperErr := scraper.Fetch(
		url,
		entry.Feed.ScraperRules,
		entry.Feed.UserAgent,
		entry.Feed.Cookie,
		feed.AllowSelfSignedCertificates,
		feed.FetchViaProxy,
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

	if content != "" {
		entry.Content = content
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}

	rewrite.Rewriter(url, entry, entry.Feed.RewriteRules)
	entry.Content = sanitizer.Sanitize(url, entry.Content)

	return nil
}

func getUrlFromEntry(feed *model.Feed, entry *model.Entry) string {
	var url = entry.URL
	if feed.UrlRewriteRules != "" {
		parts := customReplaceRuleRegex.FindStringSubmatch(feed.UrlRewriteRules)

		if len(parts) >= 3 {
			re := regexp.MustCompile(parts[1])
			url = re.ReplaceAllString(entry.URL, parts[2])
			slog.Debug("Rewriting entry URL",
				slog.Int64("entry_id", entry.ID),
				slog.String("original_entry_url", entry.URL),
				slog.String("rewritten_entry_url", url),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
			)
		} else {
			slog.Debug("Cannot find search and replace terms for replace rule",
				slog.Int64("entry_id", entry.ID),
				slog.String("original_entry_url", entry.URL),
				slog.String("rewritten_entry_url", url),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
				slog.String("url_rewrite_rules", feed.UrlRewriteRules),
			)
		}
	}
	return url
}

func updateEntryReadingTime(store *storage.Storage, feed *model.Feed, entry *model.Entry, entryIsNew bool, user *model.User) {
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
			entry.ReadingTime = store.GetReadTime(entry, feed)
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
			entry.ReadingTime = store.GetReadTime(entry, feed)
		}
	}
	// Handle YT error case and non-YT entries.
	if entry.ReadingTime == 0 {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}
}

func shouldFetchYouTubeWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchYouTubeWatchTime() {
		return false
	}
	matches := youtubeRegex.FindStringSubmatch(entry.URL)
	urlMatchesYouTubePattern := len(matches) == 2
	return urlMatchesYouTubePattern
}

func shouldFetchOdyseeWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchOdyseeWatchTime() {
		return false
	}
	matches := odyseeRegex.FindStringSubmatch(entry.URL)
	return matches != nil
}

func fetchYouTubeWatchTime(url string) (int, error) {
	clt := client.NewClientWithConfig(url, config.Opts)
	response, browserErr := browser.Exec(clt)
	if browserErr != nil {
		return 0, browserErr
	}

	doc, docErr := goquery.NewDocumentFromReader(response.Body)
	if docErr != nil {
		return 0, docErr
	}

	durs, exists := doc.Find(`meta[itemprop="duration"]`).First().Attr("content")
	if !exists {
		return 0, errors.New("duration has not found")
	}

	dur, err := parseISO8601(durs)
	if err != nil {
		return 0, fmt.Errorf("unable to parse duration %s: %v", durs, err)
	}

	return int(dur.Minutes()), nil
}

func fetchOdyseeWatchTime(url string) (int, error) {
	clt := client.NewClientWithConfig(url, config.Opts)
	response, browserErr := browser.Exec(clt)
	if browserErr != nil {
		return 0, browserErr
	}

	doc, docErr := goquery.NewDocumentFromReader(response.Body)
	if docErr != nil {
		return 0, docErr
	}

	durs, exists := doc.Find(`meta[property="og:video:duration"]`).First().Attr("content")
	// durs contains video watch time in seconds
	if !exists {
		return 0, errors.New("duration has not found")
	}

	dur, err := strconv.ParseInt(durs, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse duration %s: %v", durs, err)
	}

	return int(dur / 60), nil
}

// parseISO8601 parses an ISO 8601 duration string.
func parseISO8601(from string) (time.Duration, error) {
	var match []string
	var d time.Duration

	if iso8601Regex.MatchString(from) {
		match = iso8601Regex.FindStringSubmatch(from)
	} else {
		return 0, errors.New("could not parse duration string")
	}

	for i, name := range iso8601Regex.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, err
		}

		switch name {
		case "hour":
			d = d + (time.Duration(val) * time.Hour)
		case "minute":
			d = d + (time.Duration(val) * time.Minute)
		case "second":
			d = d + (time.Duration(val) * time.Second)
		default:
			return 0, fmt.Errorf("unknown field %s", name)
		}
	}

	return d, nil
}
