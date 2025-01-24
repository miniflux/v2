// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor

import (
	"log/slog"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/readingtime"
	"miniflux.app/v2/internal/storage"
)

func updateEntryReadingTime(store *storage.Storage, feed *model.Feed, entry *model.Entry, entryIsNew bool, user *model.User) {
	if !user.ShowReadingTime {
		slog.Debug("Skip reading time estimation for this user", slog.Int64("user_id", user.ID))
		return
	}

	// Define a type for watch time fetching functions
	type watchTimeFetcher func(string) (int, error)

	// Define watch time fetching scenarios
	watchTimeScenarios := []struct {
		shouldFetch func(*model.Entry) bool
		fetchFunc   watchTimeFetcher
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
