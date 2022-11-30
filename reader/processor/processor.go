// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"miniflux.app/integration"

	"miniflux.app/config"
	"miniflux.app/http/client"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/model"
	"miniflux.app/reader/browser"
	"miniflux.app/reader/rewrite"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/reader/scraper"
	"miniflux.app/storage"

	"github.com/PuerkitoBio/goquery"
	"github.com/rylans/getlang"
)

var (
	youtubeRegex           = regexp.MustCompile(`youtube\.com/watch\?v=(.*)`)
	iso8601Regex           = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	customReplaceRuleRegex = regexp.MustCompile(`rewrite\("(.*)"\|"(.*)"\)`)
)

// ProcessFeedEntries downloads original web page for entries and apply filters.
func ProcessFeedEntries(store *storage.Storage, feed *model.Feed, user *model.User) {
	var filteredEntries model.Entries

	// array used for bulk push
	entriesToPush := model.Entries{}

	for _, entry := range feed.Entries {
		logger.Debug("[Processor] Processing entry %q from feed %q", entry.URL, feed.FeedURL)

		if isBlockedEntry(feed, entry) || !isAllowedEntry(feed, entry) {
			continue
		}

		url := getUrlFromEntry(feed, entry)
		entryIsNew := !store.EntryURLExists(feed.ID, entry.URL)
		if feed.Crawler && entryIsNew {
			logger.Debug("[Processor] Crawling entry %q from feed %q", url, feed.FeedURL)

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
				logger.Error(`[Processor] Unable to crawl this entry: %q => %v`, entry.URL, scraperErr)
			} else if content != "" {
				// We replace the entry content only if the scraper doesn't return any error.
				entry.Content = content
			}
		}

		entry.Content = rewrite.Rewriter(url, entry.Content, feed.RewriteRules)

		// The sanitizer should always run at the end of the process to make sure unsafe HTML is filtered.
		entry.Content = sanitizer.Sanitize(url, entry.Content)

		if entryIsNew {
			intg, err := store.Integration(feed.UserID)
			if err != nil {
				logger.Error("[Processor] Get integrations for user %d failed: %v; the refresh process will go on, but no integrations will run this time.", feed.UserID, err)
			} else if intg != nil {
				localEntry := entry
				go func() {
					integration.PushEntry(localEntry, intg)
				}()
				entriesToPush = append(entriesToPush, localEntry)
			}
		}

		updateEntryReadingTime(store, feed, entry, entryIsNew, user)
		filteredEntries = append(filteredEntries, entry)
	}

	intg, err := store.Integration(feed.UserID)
	if err != nil {
		logger.Error("[Processor] Get integrations for user %d failed: %v; the refresh process will go on, but no integrations will run this time.", feed.UserID, err)
	} else if intg != nil && len(entriesToPush) > 0 {
		go func() {
			integration.PushEntries(entriesToPush, intg)
		}()
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

	content = rewrite.Rewriter(url, content, entry.Feed.RewriteRules)
	content = sanitizer.Sanitize(url, content)

	if content != "" {
		entry.Content = content
		entry.ReadingTime = calculateReadingTime(content, user)
	}

	return nil
}

func getUrlFromEntry(feed *model.Feed, entry *model.Entry) string {
	var url = entry.URL
	if feed.UrlRewriteRules != "" {
		parts := customReplaceRuleRegex.FindStringSubmatch(feed.UrlRewriteRules)

		if len(parts) >= 3 {
			re := regexp.MustCompile(parts[1])
			url = re.ReplaceAllString(entry.URL, parts[2])
			logger.Debug(`[Processor] Rewriting entry URL %s to %s`, entry.URL, url)
		} else {
			logger.Debug("[Processor] Cannot find search and replace terms for replace rule %s", feed.UrlRewriteRules)
		}
	}
	return url
}

func updateEntryReadingTime(store *storage.Storage, feed *model.Feed, entry *model.Entry, entryIsNew bool, user *model.User) {
	if shouldFetchYouTubeWatchTime(entry) {
		if entryIsNew {
			watchTime, err := fetchYouTubeWatchTime(entry.URL)
			if err != nil {
				logger.Error("[Processor] Unable to fetch YouTube watch time: %q => %v", entry.URL, err)
			}
			entry.ReadingTime = watchTime
		} else {
			entry.ReadingTime = store.GetReadTime(entry, feed)
		}
	}

	// Handle YT error case and non-YT entries.
	if entry.ReadingTime == 0 {
		entry.ReadingTime = calculateReadingTime(entry.Content, user)
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

func calculateReadingTime(content string, user *model.User) int {
	sanitizedContent := sanitizer.StripTags(content)
	languageInfo := getlang.FromString(sanitizedContent)

	var timeToReadInt int
	if languageInfo.LanguageCode() == "ko" || languageInfo.LanguageCode() == "zh" || languageInfo.LanguageCode() == "jp" {
		timeToReadInt = int(math.Ceil(float64(utf8.RuneCountInString(sanitizedContent)) / float64(user.CJKReadingSpeed)))
	} else {
		nbOfWords := len(strings.Fields(sanitizedContent))
		timeToReadInt = int(math.Ceil(float64(nbOfWords) / float64(user.DefaultReadingSpeed)))
	}

	return timeToReadInt
}
