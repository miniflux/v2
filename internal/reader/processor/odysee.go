// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
)

var odyseeRegex = regexp.MustCompile(`^https://odysee\.com`)

func shouldFetchOdyseeWatchTime(entry *model.Entry) bool {
	if !config.Opts.FetchOdyseeWatchTime() {
		return false
	}
	matches := odyseeRegex.FindStringSubmatch(entry.URL)
	return matches != nil
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
