// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"encoding/json"
	"errors"
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

const youTubeVideosAPIURL = "https://www.googleapis.com/youtube/v3/videos"

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
	return config.Opts.FetchYouTubeWatchTime() && config.Opts.YouTubeAPIKey() == "" && isYouTubeVideoURL(entry.URL)
}

func shouldFetchYouTubeWatchTimeInBulk() bool {
	return config.Opts.FetchYouTubeWatchTime() && config.Opts.YouTubeAPIKey() != ""
}

func fetchYouTubeWatchTimeForSingleEntry(websiteURL string) (int, error) {
	return fetchWatchTime(websiteURL, `meta[itemprop="duration"]`, true)
}

func fetchYouTubeWatchTimeInBulk(entries []*model.Entry) {
	videosEntriesMapping := make(map[string]*model.Entry, len(entries))
	videoIDs := make([]string, 0, len(entries))

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

	watchTimeMap, err := fetchYouTubeWatchTimeFromApiInBulk(youTubeVideosAPIURL, videoIDs)
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

func fetchYouTubeWatchTimeFromApiInBulk(apiURL string, videoIDs []string) (map[string]time.Duration, error) {
	slog.Debug("Fetching YouTube watch time in bulk", slog.Any("video_ids", videoIDs))

	apiQuery := url.Values{}
	apiQuery.Set("id", strings.Join(videoIDs, ","))
	apiQuery.Set("part", "contentDetails")

	// The key goes in a header to keep it out of logged URLs, and redirects are not
	// followed because the HTTP client forwards custom headers to the redirect target.
	requestBuilder := fetcher.NewRequestBuilder().
		WithHeader("X-Goog-Api-Key", config.Opts.YouTubeAPIKey()).
		WithoutRedirects().
		WithTimeout(config.Opts.HTTPClientTimeout()).
		WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(apiURL + "?" + apiQuery.Encode()))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch contentDetails from YouTube API", slog.Any("error", localizedError.Error()))
		return nil, localizedError.Error()
	}

	if responseHandler.IsRedirect() {
		return nil, errors.New("youtube: unexpected redirect from the API")
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
