// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor // import "miniflux.app/v2/internal/reader/processor"

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
)

var (
	bilibiliVideoIdRegex = regexp.MustCompile(`/video/(?:av(\d+)|BV([a-zA-Z0-9]+))`)
)

func shouldFetchBilibiliWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchBilibiliWatchTime() {
		return false
	}
	return strings.Contains(entry.URL, "bilibili.com/video/")
}

func extractBilibiliVideoID(websiteURL string) (string, string, error) {
	matches := bilibiliVideoIdRegex.FindStringSubmatch(websiteURL)
	if matches == nil {
		return "", "", fmt.Errorf("no video ID found in URL: %s", websiteURL)
	}
	if matches[1] != "" {
		return "aid", matches[1], nil
	}
	if matches[2] != "" {
		return "bvid", matches[2], nil
	}
	return "", "", fmt.Errorf("unexpected regex match result for URL: %s", websiteURL)
}

func fetchBilibiliWatchTime(websiteURL string) (int, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)

	idType, videoID, extractErr := extractBilibiliVideoID(websiteURL)
	if extractErr != nil {
		return 0, extractErr
	}
	bilibiliApiURL := "https://api.bilibili.com/x/web-interface/view?" + idType + "=" + videoID

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(bilibiliApiURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch Bilibili API",
			slog.String("website_url", websiteURL),
			slog.String("api_url", bilibiliApiURL),
			slog.Any("error", localizedError.Error()))
		return 0, localizedError.Error()
	}

	var result map[string]any
	doc := json.NewDecoder(responseHandler.Body(config.Opts.HTTPClientMaxBodySize()))
	if docErr := doc.Decode(&result); docErr != nil {
		return 0, fmt.Errorf("failed to decode API response: %v", docErr)
	}

	if code, ok := result["code"].(float64); !ok || code != 0 {
		return 0, fmt.Errorf("API returned error code: %v", result["code"])
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("data field not found or not an object")
	}

	duration, ok := data["duration"].(float64)
	if !ok {
		return 0, fmt.Errorf("duration not found or not a number")
	}
	intDuration := int(duration)
	durationMin := intDuration / 60
	if intDuration%60 != 0 {
		durationMin++
	}
	return durationMin, nil
}
