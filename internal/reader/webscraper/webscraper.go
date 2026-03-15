// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package webscraper // import "miniflux.app/v2/internal/reader/webscraper"

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

const defaultMaxItems = 25

// ScrapeConfig defines CSS selectors (for HTML) or gjson paths (for JSON) to extract feed items.
type ScrapeConfig struct {
	ItemSelector        string
	TitleSelector       string
	LinkSelector        string
	DescriptionSelector string
	NextPageSelector    string
	MaxItems            int
}

// ScrapeResult holds a single extracted feed item.
type ScrapeResult struct {
	Title       string
	Link        string
	Description string
}

// ScrapeWebPage fetches a web page and extracts structured feed items using CSS selectors (HTML)
// or gjson paths (JSON) based on the response Content-Type header.
func ScrapeWebPage(requestBuilder *fetcher.RequestBuilder, pageURL string, scrapeConfig *ScrapeConfig) ([]*ScrapeResult, error) {
	return scrapeWithAccumulator(requestBuilder, pageURL, scrapeConfig, nil)
}

// scrapeWithAccumulator is the recursive implementation that carries accumulated items across pages.
func scrapeWithAccumulator(requestBuilder *fetcher.RequestBuilder, pageURL string, scrapeConfig *ScrapeConfig, accumulated []*ScrapeResult) ([]*ScrapeResult, error) {
	maxItems := scrapeConfig.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}

	// Stop pagination if we already have enough items.
	if len(accumulated) >= maxItems {
		return accumulated[:maxItems], nil
	}

	slog.Debug("Scraping web page",
		slog.String("url", pageURL),
		slog.Int("accumulated_items", len(accumulated)),
	)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(pageURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to scrape web page",
			slog.String("url", pageURL),
			slog.Any("error", localizedError.Error()),
		)
		return accumulated, localizedError.Error()
	}

	contentType := responseHandler.ContentType()

	// Route to JSON or HTML handler based on Content-Type.
	if isJSONContentType(contentType) {
		return processJSONResponse(requestBuilder, pageURL, scrapeConfig, accumulated, maxItems, responseHandler)
	}

	if isHTMLContentType(contentType) {
		return processHTMLResponse(requestBuilder, pageURL, scrapeConfig, accumulated, maxItems, responseHandler)
	}

	return accumulated, fmt.Errorf("webscraper: unsupported content type %q for %s", contentType, pageURL)
}

func processJSONResponse(
	requestBuilder *fetcher.RequestBuilder,
	pageURL string,
	scrapeConfig *ScrapeConfig,
	accumulated []*ScrapeResult,
	maxItems int,
	responseHandler *fetcher.ResponseHandler,
) ([]*ScrapeResult, error) {
	bodyBytes, localizedErr := responseHandler.ReadBody(config.Opts.HTTPClientMaxBodySize())
	if localizedErr != nil {
		return accumulated, localizedErr.Error()
	}
	body := string(bodyBytes)

	gjson.Get(body, scrapeConfig.ItemSelector).ForEach(func(key, value gjson.Result) bool {
		var result ScrapeResult

		if scrapeConfig.TitleSelector != "" {
			result.Title = strings.TrimSpace(value.Get(scrapeConfig.TitleSelector).String())
		}
		if scrapeConfig.LinkSelector != "" {
			result.Link = strings.TrimSpace(value.Get(scrapeConfig.LinkSelector).String())
		}
		if scrapeConfig.DescriptionSelector != "" {
			result.Description = strings.TrimSpace(value.Get(scrapeConfig.DescriptionSelector).String())
		}

		// Skip items where both Title and Link are empty.
		if result.Title == "" && result.Link == "" {
			return true
		}

		// Resolve relative URLs to absolute using the page URL as base.
		result.Link = mergeURL(pageURL, result.Link)

		accumulated = append(accumulated, &result)
		return len(accumulated) < maxItems
	})

	// JSON pagination: use gjson path to get next page URL.
	if len(accumulated) < maxItems && scrapeConfig.NextPageSelector != "" {
		nextPageResult := gjson.Get(body, scrapeConfig.NextPageSelector)
		if nextPageResult.Exists() && nextPageResult.String() != "" {
			nextURL := mergeURL(pageURL, nextPageResult.String())
			return scrapeWithAccumulator(requestBuilder, nextURL, scrapeConfig, accumulated)
		}
	}

	return accumulated, nil
}

func processHTMLResponse(
	requestBuilder *fetcher.RequestBuilder,
	pageURL string,
	scrapeConfig *ScrapeConfig,
	accumulated []*ScrapeResult,
	maxItems int,
	responseHandler *fetcher.ResponseHandler,
) ([]*ScrapeResult, error) {
	// Use charset-aware reader for proper encoding handling (same pattern as scraper.go).
	htmlReader, err := encoding.NewCharsetReader(
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
		responseHandler.ContentType(),
	)
	if err != nil {
		return accumulated, fmt.Errorf("webscraper: unable to read HTML with charset reader: %w", err)
	}

	// Read all HTML for both goquery parsing and regex-based pagination.
	htmlBytes, err := io.ReadAll(htmlReader)
	if err != nil {
		return accumulated, fmt.Errorf("webscraper: unable to read HTML body: %w", err)
	}
	htmlBody := string(htmlBytes)

	extracted, err := extractItemsFromHTML(htmlBody, pageURL, scrapeConfig, maxItems-len(accumulated))
	if err != nil {
		return accumulated, err
	}
	accumulated = append(accumulated, extracted...)

	// HTML pagination: use regex to match next page URL from raw HTML.
	if len(accumulated) < maxItems && scrapeConfig.NextPageSelector != "" {
		matcher := regexp.MustCompile(scrapeConfig.NextPageSelector)
		matches := matcher.FindStringSubmatch(htmlBody)
		if len(matches) > 1 && matches[1] != "" {
			nextURL := mergeURL(pageURL, matches[1])
			return scrapeWithAccumulator(requestBuilder, nextURL, scrapeConfig, accumulated)
		}
	}

	return accumulated, nil
}

func extractItemsFromHTML(htmlBody, pageURL string, scrapeConfig *ScrapeConfig, limit int) ([]*ScrapeResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlBody))
	if err != nil {
		return nil, fmt.Errorf("webscraper: unable to parse HTML: %w", err)
	}

	var results []*ScrapeResult
	doc.Find(scrapeConfig.ItemSelector).EachWithBreak(func(i int, s *goquery.Selection) bool {
		var result ScrapeResult

		if scrapeConfig.TitleSelector != "" {
			result.Title = strings.TrimSpace(s.Find(scrapeConfig.TitleSelector).First().Text())
		}
		if scrapeConfig.LinkSelector != "" {
			result.Link = strings.TrimSpace(s.Find(scrapeConfig.LinkSelector).First().AttrOr("href", ""))
		}
		if scrapeConfig.DescriptionSelector != "" {
			result.Description = strings.TrimSpace(s.Find(scrapeConfig.DescriptionSelector).First().Text())
		}

		if result.Title == "" && result.Link == "" {
			return true
		}

		result.Link = mergeURL(pageURL, result.Link)

		results = append(results, &result)
		return len(results) < limit
	})

	return results, nil
}

// ScrapeRenderedHTML extracts items from pre-rendered HTML (e.g. from headless
// browser JS rendering) without making HTTP requests. Pagination is not
// supported since the HTML is already fully rendered.
func ScrapeRenderedHTML(htmlBody, pageURL string, scrapeConfig *ScrapeConfig) ([]*ScrapeResult, error) {
	maxItems := scrapeConfig.MaxItems
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	return extractItemsFromHTML(htmlBody, pageURL, scrapeConfig, maxItems)
}

// mergeURL resolves a potentially relative target URL against a base URL.
func mergeURL(base, target string) string {
	if target == "" {
		return target
	}

	// Already absolute.
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return target
	}

	// Protocol-relative URL (e.g. "//example.com/path").
	if strings.HasPrefix(target, "//") {
		return baseURL.Scheme + ":" + target
	}

	// Absolute path (e.g. "/path/to/page").
	if strings.HasPrefix(target, "/") {
		return baseURL.Scheme + "://" + baseURL.Host + target
	}

	// Relative path — append to base directory.
	if strings.HasSuffix(base, "/") {
		return base + target
	}
	return base + "/" + target
}

func isJSONContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/json")
}

func isHTMLContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "text/html") || strings.HasPrefix(ct, "application/xhtml+xml")
}
