// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package icon // import "miniflux.app/v2/internal/reader/icon"

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

type IconFinder struct {
	requestBuilder *fetcher.RequestBuilder
	websiteURL     string
	feedIconURL    string
}

func NewIconFinder(requestBuilder *fetcher.RequestBuilder, websiteURL, feedIconURL string) *IconFinder {
	return &IconFinder{
		requestBuilder: requestBuilder,
		websiteURL:     websiteURL,
		feedIconURL:    feedIconURL,
	}
}

func (f *IconFinder) FindIcon() (*model.Icon, error) {
	slog.Debug("Begin icon discovery process",
		slog.String("website_url", f.websiteURL),
		slog.String("feed_icon_url", f.feedIconURL),
	)

	if f.feedIconURL != "" {
		if icon, err := f.FetchFeedIcon(); err != nil {
			slog.Debug("Unable to download icon from feed",
				slog.String("website_url", f.websiteURL),
				slog.String("feed_icon_url", f.feedIconURL),
				slog.Any("error", err),
			)
		} else if icon != nil {
			return icon, nil
		}
	}

	if icon, err := f.FetchIconsFromHTMLDocument(); err != nil {
		slog.Debug("Unable to fetch icons from HTML document",
			slog.String("website_url", f.websiteURL),
			slog.Any("error", err),
		)
	} else if icon != nil {
		return icon, nil
	}

	return f.FetchDefaultIcon()
}

func (f *IconFinder) FetchDefaultIcon() (*model.Icon, error) {
	slog.Debug("Fetching default icon",
		slog.String("website_url", f.websiteURL),
	)

	iconURL, err := urllib.JoinBaseURLAndPath(urllib.RootURL(f.websiteURL), "favicon.ico")
	if err != nil {
		return nil, fmt.Errorf(`icon: unable to join root URL and path: %w`, err)
	}

	icon, err := f.DownloadIcon(iconURL)
	if err != nil {
		return nil, err
	}

	return icon, nil
}

func (f *IconFinder) FetchFeedIcon() (*model.Icon, error) {
	slog.Debug("Fetching feed icon",
		slog.String("website_url", f.websiteURL),
		slog.String("feed_icon_url", f.feedIconURL),
	)

	iconURL, err := urllib.AbsoluteURL(f.websiteURL, f.feedIconURL)
	if err != nil {
		return nil, fmt.Errorf(`icon: unable to convert icon URL to absolute URL: %w`, err)
	}

	return f.DownloadIcon(iconURL)
}

func (f *IconFinder) FetchIconsFromHTMLDocument() (*model.Icon, error) {
	slog.Debug("Searching icons from HTML document",
		slog.String("website_url", f.websiteURL),
	)

	rootURL := urllib.RootURL(f.websiteURL)

	responseHandler := fetcher.NewResponseHandler(f.requestBuilder.ExecuteRequest(rootURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		return nil, fmt.Errorf("icon: unable to download website index page: %w", localizedError.Error())
	}

	iconURLs, err := findIconURLsFromHTMLDocument(
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
		responseHandler.ContentType(),
	)
	if err != nil {
		return nil, err
	}

	slog.Debug("Searched icon from HTML document",
		slog.String("website_url", f.websiteURL),
		slog.String("icon_urls", strings.Join(iconURLs, ",")),
	)

	for _, iconURL := range iconURLs {
		if strings.HasPrefix(iconURL, "data:") {
			slog.Debug("Found icon with data URL",
				slog.String("website_url", f.websiteURL),
			)
			return parseImageDataURL(iconURL)
		}

		iconURL, err = urllib.AbsoluteURL(f.websiteURL, iconURL)
		if err != nil {
			return nil, fmt.Errorf(`icon: unable to convert icon URL to absolute URL: %w`, err)
		}

		if icon, err := f.DownloadIcon(iconURL); err != nil {
			slog.Debug("Unable to download icon from HTML document",
				slog.String("website_url", f.websiteURL),
				slog.String("icon_url", iconURL),
				slog.Any("error", err),
			)
		} else if icon != nil {
			slog.Debug("Found icon from HTML document",
				slog.String("website_url", f.websiteURL),
				slog.String("icon_url", iconURL),
			)
			return icon, nil
		}
	}

	return nil, nil
}

func (f *IconFinder) DownloadIcon(iconURL string) (*model.Icon, error) {
	slog.Debug("Downloading icon",
		slog.String("website_url", f.websiteURL),
		slog.String("icon_url", iconURL),
	)

	responseHandler := fetcher.NewResponseHandler(f.requestBuilder.ExecuteRequest(iconURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		return nil, fmt.Errorf("icon: unable to download website icon: %w", localizedError.Error())
	}

	responseBody, localizedError := responseHandler.ReadBody(config.Opts.HTTPClientMaxBodySize())
	if localizedError != nil {
		return nil, fmt.Errorf("icon: unable to read response body: %w", localizedError.Error())
	}

	icon := &model.Icon{
		Hash:     crypto.HashFromBytes(responseBody),
		MimeType: responseHandler.ContentType(),
		Content:  responseBody,
		URL:      iconURL,
	}

	return icon, nil
}

func findIconURLsFromHTMLDocument(body io.Reader, contentType string) ([]string, error) {
	queries := []string{
		"link[rel='shortcut icon']",
		"link[rel='Shortcut Icon']",
		"link[rel='icon shortcut']",
		"link[rel='icon']",
	}

	htmlDocumentReader, err := encoding.CharsetReaderFromContentType(contentType, body)
	if err != nil {
		return nil, fmt.Errorf("icon: unable to create charset reader: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(htmlDocumentReader)
	if err != nil {
		return nil, fmt.Errorf("icon: unable to read document: %v", err)
	}

	var iconURLs []string
	for _, query := range queries {
		slog.Debug("Searching icon URL in HTML document", slog.String("query", query))

		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			var iconURL string

			if href, exists := s.Attr("href"); exists {
				iconURL = strings.TrimSpace(href)
			}

			if iconURL != "" {
				iconURLs = append(iconURLs, iconURL)

				slog.Debug("Found icon URL in HTML document",
					slog.String("query", query),
					slog.String("icon_url", iconURL))
			}
		})
	}

	return iconURLs, nil
}

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs#syntax
// data:[<mediatype>][;base64],<data>
func parseImageDataURL(value string) (*model.Icon, error) {
	var mediaType string
	var encoding string

	if !strings.HasPrefix(value, "data:") {
		return nil, fmt.Errorf(`icon: invalid data URL (missing data:) %q`, value)
	}

	value = value[5:]

	comma := strings.Index(value, ",")
	if comma < 0 {
		return nil, fmt.Errorf(`icon: invalid data URL (no comma) %q`, value)
	}

	data := value[comma+1:]
	semicolon := strings.Index(value[0:comma], ";")

	if semicolon > 0 {
		mediaType = value[0:semicolon]
		encoding = value[semicolon+1 : comma]
	} else {
		mediaType = value[0:comma]
	}

	if !strings.HasPrefix(mediaType, "image/") {
		return nil, fmt.Errorf(`icon: invalid media type %q`, mediaType)
	}

	var blob []byte
	switch encoding {
	case "base64":
		var err error
		blob, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, fmt.Errorf(`icon: invalid data %q (%v)`, value, err)
		}
	case "":
		decodedData, err := url.QueryUnescape(data)
		if err != nil {
			return nil, fmt.Errorf(`icon: unable to decode data URL %q`, value)
		}
		blob = []byte(decodedData)
	case "utf8":
		blob = []byte(data)
	default:
		return nil, fmt.Errorf(`icon: unsupported data URL encoding %q`, value)
	}

	if len(blob) == 0 {
		return nil, fmt.Errorf(`icon: empty data URL %q`, value)
	}

	icon := &model.Icon{
		Hash:     crypto.HashFromBytes(blob),
		Content:  blob,
		MimeType: mediaType,
	}

	return icon, nil
}

func URLHasChanged(original, updated, siteURL string) bool {
	if updated == "" {
		return false
	}
	updated, _ = urllib.AbsoluteURL(siteURL, updated)
	return original == updated
}
