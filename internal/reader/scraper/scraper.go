// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"bytes"
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

// ScrapeResult holds the artifacts collected while scraping an entry URL.
//
// Metadata contains OpenGraph and Twitter Card values collected from the
// page's <head>. It can be empty when the page does not expose any preview
// metadata or when scraping was not performed (for example when the feed has
// the crawler disabled).
type ScrapeResult struct {
	BaseURL          string
	ExtractedContent string
	Metadata         map[string]string
}

func ScrapeWebsite(requestBuilder *fetcher.RequestBuilder, pageURL, rules string) (ScrapeResult, error) {
	result := ScrapeResult{Metadata: map[string]string{}}

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(pageURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to scrape website", slog.String("website_url", pageURL), slog.Any("error", localizedError.Error()))
		return result, localizedError.Error()
	}

	if !isAllowedContentType(responseHandler.ContentType()) {
		return result, fmt.Errorf("scraper: this resource is not a HTML document (%s)", responseHandler.ContentType())
	}

	// The entry URL could redirect somewhere else.
	sameSite := urllib.Domain(pageURL) == urllib.Domain(responseHandler.EffectiveURL())
	pageURL = responseHandler.EffectiveURL()

	if rules == "" {
		rules = getPredefinedScraperRules(pageURL)
	}

	htmlDocumentReader, err := encoding.NewCharsetReader(
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
		responseHandler.ContentType(),
	)

	if err != nil {
		return result, fmt.Errorf("scraper: unable to read HTML document with charset reader: %v", err)
	}

	// Buffer the document so we can both run the chosen extraction strategy
	// and parse <head> meta tags without re-fetching the page.
	htmlBytes, err := io.ReadAll(htmlDocumentReader)
	if err != nil {
		return result, fmt.Errorf("scraper: unable to read HTML document body: %v", err)
	}

	if sameSite && rules != "" {
		slog.Debug("Extracting content with custom rules",
			"url", pageURL,
			"rules", rules,
		)
		result.BaseURL, result.ExtractedContent, err = findContentUsingCustomRules(bytes.NewReader(htmlBytes), rules)
	} else {
		slog.Debug("Extracting content with readability",
			"url", pageURL,
		)
		result.BaseURL, result.ExtractedContent, err = readability.ExtractContent(bytes.NewReader(htmlBytes))
	}

	if err != nil {
		slog.Debug("Content extraction returned an error",
			"url", pageURL,
			"error", err,
		)
	}

	if metadataDoc, metadataErr := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes)); metadataErr == nil {
		result.Metadata = ExtractHeadMetadata(metadataDoc)
	}

	if result.BaseURL == "" {
		result.BaseURL = pageURL
	} else {
		slog.Debug("Using base URL from HTML document", "base_url", result.BaseURL)
	}

	return result, nil
}

func findContentUsingCustomRules(page io.Reader, rules string) (baseURL string, extractedContent string, err error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", "", err
	}

	if hrefValue, exists := document.FindMatcher(goquery.Single("head base")).Attr("href"); exists {
		hrefValue = strings.TrimSpace(hrefValue)
		if urllib.IsAbsoluteURL(hrefValue) {
			baseURL = hrefValue
		}
	}

	var buf strings.Builder
	document.Find(rules).Each(func(i int, s *goquery.Selection) {
		if content, err := goquery.OuterHtml(s); err == nil {
			buf.WriteString(content)
		}
	})

	return baseURL, buf.String(), nil
}

func getPredefinedScraperRules(websiteURL string) string {
	urlDomain := urllib.DomainWithoutWWW(websiteURL)

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
