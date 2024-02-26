// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/readability"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

func ScrapeWebsite(requestBuilder *fetcher.RequestBuilder, websiteURL, rules string) (string, error) {
	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(websiteURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to scrape website", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return "", localizedError.Error()
	}

	if !isAllowedContentType(responseHandler.ContentType()) {
		return "", fmt.Errorf("scraper: this resource is not a HTML document (%s)", responseHandler.ContentType())
	}

	// The entry URL could redirect somewhere else.
	sameSite := urllib.Domain(websiteURL) == urllib.Domain(responseHandler.EffectiveURL())
	websiteURL = responseHandler.EffectiveURL()

	if rules == "" {
		rules = getPredefinedScraperRules(websiteURL)
	}

	var content string
	var err error

	htmlDocumentReader, err := encoding.CharsetReaderFromContentType(
		responseHandler.ContentType(),
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
	)
	if err != nil {
		return "", fmt.Errorf("scraper: unable to read HTML document: %v", err)
	}

	if sameSite && rules != "" {
		slog.Debug("Extracting content with custom rules",
			"url", websiteURL,
			"rules", rules,
		)
		content, err = findContentUsingCustomRules(htmlDocumentReader, rules)
	} else {
		slog.Debug("Extracting content with readability",
			"url", websiteURL,
		)
		content, err = readability.ExtractContent(htmlDocumentReader)
	}

	if err != nil {
		return "", err
	}

	return content, nil
}

func findContentUsingCustomRules(page io.Reader, rules string) (string, error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", err
	}

	contents := ""
	document.Find(rules).Each(func(i int, s *goquery.Selection) {
		if content, err := goquery.OuterHtml(s); err == nil {
			contents += content
		}
	})

	return contents, nil
}

func getPredefinedScraperRules(websiteURL string) string {
	urlDomain := urllib.Domain(websiteURL)
	urlDomain = strings.TrimPrefix(urlDomain, "www.")

	if rules, ok := predefinedRules[urlDomain]; ok {
		return rules
	}
	return ""
}

func isAllowedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/html") ||
		strings.HasPrefix(contentType, "application/xhtml+xml")
}
