// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
)

var (
	youtubeRegex = regexp.MustCompile(`youtube\.com/watch\?v=(.*)$`)
	iso8601Regex = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
)

func isYouTubeVideoURL(websiteURL string) bool {
	return len(youtubeRegex.FindStringSubmatch(websiteURL)) == 2
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
	slog.Debug("Fetching YouTube watch time for a single entry", slog.String("website_url", websiteURL))

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())

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

	htmlDuration, exists := doc.FindMatcher(goquery.Single(`meta[itemprop="duration"]`)).Attr("content")
	if !exists {
		return 0, errors.New("youtube: duration has not found")
	}

	parsedDuration, err := parseISO8601(htmlDuration)
	if err != nil {
		return 0, fmt.Errorf("youtube: unable to parse duration %s: %v", htmlDuration, err)
	}

	return int(parsedDuration.Minutes()), nil
}

func fetchYouTubeWatchTimeInBulk(entries []*model.Entry) {
	var videosEntriesMapping = make(map[string]*model.Entry)
	var videoIDs []string

	for _, entry := range entries {
		if !isYouTubeVideoURL(entry.URL) {
			continue
		}

		youtubeVideoID := getVideoIDFromYouTubeURL(entry.URL)
		if youtubeVideoID == "" {
			continue
		}

		videosEntriesMapping[getVideoIDFromYouTubeURL(entry.URL)] = entry
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
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(apiURL.String()))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch contentDetails from YouTube API", slog.Any("error", localizedError.Error()))
		return nil, localizedError.Error()
	}

	var videos youtubeVideoListResponse
	if err := json.NewDecoder(responseHandler.Body(config.Opts.HTTPClientMaxBodySize())).Decode(&videos); err != nil {
		return nil, fmt.Errorf("youtube: unable to decode JSON: %v", err)
	}

	watchTimeMap := make(map[string]time.Duration)
	for _, video := range videos.Items {
		duration, err := parseISO8601(video.ContentDetails.Duration)
		if err != nil {
			slog.Warn("Unable to parse ISO8601 duration", slog.Any("error", err))
			continue
		}
		watchTimeMap[video.ID] = duration
	}
	return watchTimeMap, nil
}

func parseISO8601(from string) (time.Duration, error) {
	var match []string
	var d time.Duration

	if iso8601Regex.MatchString(from) {
		match = iso8601Regex.FindStringSubmatch(from)
	} else {
		return 0, errors.New("youtube: could not parse duration string")
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
			d += time.Duration(val) * time.Hour
		case "minute":
			d += time.Duration(val) * time.Minute
		case "second":
			d += time.Duration(val) * time.Second
		default:
			return 0, fmt.Errorf("youtube: unknown field %s", name)
		}
	}

	return d, nil
}

type youtubeVideoListResponse struct {
	Items []struct {
		ID             string `json:"id"`
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
}
