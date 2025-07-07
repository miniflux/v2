// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
)

func isYouTubeVideoURL(websiteURL string) bool {
	return strings.Contains(websiteURL, "youtube.com/watch?v=")
}

func getVideoIDFromYouTubeURL(websiteURL string) string {
	parsedWebsiteURL, err := url.Parse(websiteURL)
	if err != nil {
		return ""
	}

	return parsedWebsiteURL.Query().Get("v")
}

func shouldFetchYouTubeWatchTimeForSingleEntry(entry *model.Entry) bool {
	return config.Opts.FetchYouTubeWatchTime() && config.Opts.YouTubeApiKey() == "" && isYouTubeVideoURL(entry.URL)
}

func shouldFetchYouTubeWatchTimeInBulk() bool {
	return config.Opts.FetchYouTubeWatchTime() && config.Opts.YouTubeApiKey() != ""
}

func fetchYouTubeWatchTimeForSingleEntry(websiteURL string) (int, error) {
	return fetchWatchTime(websiteURL, `meta[itemprop="duration"]`, true)
}

func fetchYouTubeWatchTimeInBulk(entries []*model.Entry) {
	var videosEntriesMapping = make(map[string]*model.Entry, len(entries))
	var videoIDs []string

	for _, entry := range entries {
		if !isYouTubeVideoURL(entry.URL) {
			continue
		}

		youtubeVideoID := getVideoIDFromYouTubeURL(entry.URL)
		if youtubeVideoID == "" {
			continue
		}

		videosEntriesMapping[youtubeVideoID] = entry
		videoIDs = append(videoIDs, youtubeVideoID)
	}

	if len(videoIDs) == 0 {
		return
	}

	watchTimeMap, err := fetchYouTubeWatchTimeFromApiInBulk(videoIDs)
	if err != nil {
		slog.Warn("Unable to fetch YouTube watch time in bulk", slog.Any("error", err))
		return
	}

	for videoID, watchTime := range watchTimeMap {
		if entry, ok := videosEntriesMapping[videoID]; ok {
			entry.ReadingTime = int(watchTime.Minutes())
		}
	}
}

func fetchYouTubeWatchTimeFromApiInBulk(videoIDs []string) (map[string]time.Duration, error) {
	slog.Debug("Fetching YouTube watch time in bulk", slog.Any("video_ids", videoIDs))

	apiQuery := url.Values{}
	apiQuery.Set("id", strings.Join(videoIDs, ","))
	apiQuery.Set("key", config.Opts.YouTubeApiKey())
	apiQuery.Set("part", "contentDetails")

	apiURL := url.URL{
		Scheme:   "https",
		Host:     "www.googleapis.com",
		Path:     "youtube/v3/videos",
		RawQuery: apiQuery.Encode(),
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(apiURL.String()))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch contentDetails from YouTube API", slog.Any("error", localizedError.Error()))
		return nil, localizedError.Error()
	}

	videos := struct {
		Items []struct {
			ID             string `json:"id"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}{}
	if err := json.NewDecoder(responseHandler.Body(config.Opts.HTTPClientMaxBodySize())).Decode(&videos); err != nil {
		return nil, fmt.Errorf("youtube: unable to decode JSON: %v", err)
	}

	watchTimeMap := make(map[string]time.Duration, len(videos.Items))
	for _, video := range videos.Items {
		duration, err := parseISO8601Duration(video.ContentDetails.Duration)
		if err != nil {
			slog.Warn("Unable to parse ISO8601 duration", slog.Any("error", err))
			continue
		}
		watchTimeMap[video.ID] = duration
	}
	return watchTimeMap, nil
}
