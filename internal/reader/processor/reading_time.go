// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/storage"
)

func fetchWatchTime(websiteURL, query string, isoDate bool) (int, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch watch time", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return 0, localizedError.Error()
	}

	doc, docErr := goquery.NewDocumentFromReader(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
	if docErr != nil {
		return 0, docErr
	}

	duration, exists := doc.FindMatcher(goquery.Single(query)).Attr("content")
	if !exists {
		return 0, errors.New("duration not found")
	}

	ret := 0
	if isoDate {
		parsedDuration, err := parseISO8601Duration(duration)
		if err != nil {
			return 0, fmt.Errorf("unable to parse iso duration %s: %v", duration, err)
		}
		ret = int(parsedDuration.Minutes())
	} else {
		parsedDuration, err := strconv.ParseInt(duration, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to parse duration %s: %v", duration, err)
		}
		ret = int(parsedDuration / 60)
	}
	return ret, nil
}

func updateEntryReadingTime(store *storage.Storage, feed *model.Feed, entry *model.Entry, entryIsNew bool, user *model.User) {
	if !user.ShowReadingTime {
		slog.Debug("Skip reading time estimation for this user", slog.Int64("user_id", user.ID))
		return
	}

	// Define watch time fetching scenarios
	watchTimeScenarios := []struct {
		shouldFetch func(*model.Entry) bool
		fetchFunc   func(string) (int, error)
		platform    string
	}{
		{shouldFetchYouTubeWatchTimeForSingleEntry, fetchYouTubeWatchTimeForSingleEntry, "YouTube"},
		{shouldFetchNebulaWatchTime, fetchNebulaWatchTime, "Nebula"},
		{shouldFetchOdyseeWatchTime, fetchOdyseeWatchTime, "Odysee"},
		{shouldFetchBilibiliWatchTime, fetchBilibiliWatchTime, "Bilibili"},
	}

	// Iterate through scenarios and attempt to fetch watch time
	for _, scenario := range watchTimeScenarios {
		if scenario.shouldFetch(entry) {
			if entryIsNew {
				if watchTime, err := scenario.fetchFunc(entry.URL); err != nil {
					slog.Warn("Unable to fetch watch time",
						slog.String("platform", scenario.platform),
						slog.Int64("user_id", user.ID),
						slog.Int64("entry_id", entry.ID),
						slog.String("entry_url", entry.URL),
						slog.Int64("feed_id", feed.ID),
						slog.String("feed_url", feed.FeedURL),
						slog.Any("error", err),
					)
				} else {
					entry.ReadingTime = watchTime
				}
			} else {
				entry.ReadingTime = store.GetReadTime(feed.ID, entry.Hash)
			}
			break
		}
	}

	// Fallback to text-based reading time estimation
	if entry.ReadingTime == 0 && entry.Content != "" {
		entry.ReadingTime = readingtime.EstimateReadingTime(entry.Content, user.DefaultReadingSpeed, user.CJKReadingSpeed)
	}
}
