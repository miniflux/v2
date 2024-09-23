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

	"github.com/PuerkitoBio/goquery"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
)

var (
	youtubeRegex = regexp.MustCompile(`youtube\.com/watch\?v=(.*)$`)
	iso8601Regex = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<week>\d+)W)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
)

func shouldFetchYouTubeWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchYouTubeWatchTime() {
		return false
	}
	matches := youtubeRegex.FindStringSubmatch(entry.URL)
	urlMatchesYouTubePattern := len(matches) == 2
	return urlMatchesYouTubePattern
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
