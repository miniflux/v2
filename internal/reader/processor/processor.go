// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/metric"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/reader/rewrite"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/reader/scraper"
	"miniflux.app/v2/internal/storage"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

var (
	youtubeRegex           = regexp.MustCompile(`youtube\.com/watch\?v=(.*)$`)
	nebulaRegex            = regexp.MustCompile(`^https://nebula\.tv`)
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
			slog.String("entry_url", entry.URL),
			slog.String("entry_hash", entry.Hash),
			slog.String("entry_title", entry.Title),
			slog.Int64("feed_id", feed.ID),
			slog.String("feed_url", feed.FeedURL),
		)
		if isBlockedEntry(feed, entry, user) || !isAllowedEntry(feed, entry, user) || !isRecentEntry(entry) {
			continue
		}

		websiteURL := getUrlFromEntry(feed, entry)
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
				slog.String("website_url", websiteURL),
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

			content, scraperErr := scraper.ScrapeWebsite(
				requestBuilder,
				websiteURL,
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
				slog.Warn("Unable to scrape entry",
					slog.Int64("user_id", user.ID),
					slog.String("entry_url", entry.URL),
					slog.Int64("feed_id", feed.ID),
					slog.String("feed_url", feed.FeedURL),
					slog.Any("error", scraperErr),
				)
			} else if content != "" {
				// We replace the entry content only if the scraper doesn't return any error.
				entry.Content = minifyEntryContent(content)
			}
		}

		rewrite.Rewriter(websiteURL, entry, feed.RewriteRules)

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
		entry.Content = sanitizer.Sanitize(websiteURL, entry.Content)

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
	websiteURL := getUrlFromEntry(feed, entry)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUserAgent(feed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(feed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feed.DisableHTTP2)

	content, scraperErr := scraper.ScrapeWebsite(
		requestBuilder,
		websiteURL,
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

	if content != "" {
		entry.Content = minifyEntryContent(content)
		if user.ShowReadingTime {
			entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
		}
	}

	rewrite.Rewriter(websiteURL, entry, entry.Feed.RewriteRules)
	entry.Content = sanitizer.Sanitize(websiteURL, entry.Content)

	return nil
}

func getUrlFromEntry(feed *model.Feed, entry *model.Entry) string {
	var url = entry.URL
	if feed.UrlRewriteRules != "" {
		parts := customReplaceRuleRegex.FindStringSubmatch(feed.UrlRewriteRules)

		if len(parts) >= 3 {
			re, err := regexp.Compile(parts[1])
			if err != nil {
				slog.Error("Failed on regexp compilation",
					slog.String("url_rewrite_rules", feed.UrlRewriteRules),
					slog.Any("error", err),
				)
				return url
			}
			url = re.ReplaceAllString(entry.URL, parts[2])
			slog.Debug("Rewriting entry URL",
				slog.String("original_entry_url", entry.URL),
				slog.String("rewritten_entry_url", url),
				slog.Int64("feed_id", feed.ID),
				slog.String("feed_url", feed.FeedURL),
			)
		} else {
			slog.Debug("Cannot find search and replace terms for replace rule",
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

func shouldFetchNebulaWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchNebulaWatchTime() {
		return false
	}
	matches := nebulaRegex.FindStringSubmatch(entry.URL)
	return matches != nil
}

func shouldFetchOdyseeWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchOdyseeWatchTime() {
		return false
	}
	matches := odyseeRegex.FindStringSubmatch(entry.URL)
	return matches != nil
}

func fetchYouTubeWatchTime(websiteURL string) (int, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch YouTube page", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return 0, localizedError.Error()
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
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

func fetchNebulaWatchTime(websiteURL string) (int, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch Nebula watch time", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return 0, localizedError.Error()
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
	if docErr != nil {
		return 0, docErr
	}

	durs, exists := doc.Find(`meta[property="video:duration"]`).First().Attr("content")
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

func fetchOdyseeWatchTime(websiteURL string) (int, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch Odysee watch time", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return 0, localizedError.Error()
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
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
			d += (time.Duration(val) * time.Hour)
		case "minute":
			d += (time.Duration(val) * time.Minute)
		case "second":
			d += (time.Duration(val) * time.Second)
		default:
			return 0, fmt.Errorf("unknown field %s", name)
		}
	}

	return d, nil
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
