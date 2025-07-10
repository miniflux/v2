// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package icon // import "miniflux.app/v2/internal/reader/icon"

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/image/draw"
)

type iconFinder struct {
	requestBuilder *fetcher.RequestBuilder
	websiteURL     string
	feedIconURL    string
}

func newIconFinder(requestBuilder *fetcher.RequestBuilder, websiteURL, feedIconURL string) *iconFinder {
	return &iconFinder{
		requestBuilder: requestBuilder,
		websiteURL:     websiteURL,
		feedIconURL:    feedIconURL,
	}
}

func (f *iconFinder) findIcon() (*model.Icon, error) {
	slog.Debug("Begin icon discovery process",
		slog.String("website_url", f.websiteURL),
		slog.String("feed_icon_url", f.feedIconURL),
	)

	if f.feedIconURL != "" {
		if icon, err := f.fetchFeedIcon(); err != nil {
			slog.Debug("Unable to download icon from feed",
				slog.String("website_url", f.websiteURL),
				slog.String("feed_icon_url", f.feedIconURL),
				slog.Any("error", err),
			)
		} else if icon != nil {
			return icon, nil
		}
	}

	if icon, err := f.fetchIconsFromHTMLDocument(); err != nil {
		slog.Debug("Unable to fetch icons from HTML document",
			slog.String("website_url", f.websiteURL),
			slog.Any("error", err),
		)
	} else if icon != nil {
		return icon, nil
	}

	return f.fetchDefaultIcon()
}

func (f *iconFinder) fetchDefaultIcon() (*model.Icon, error) {
	slog.Debug("Fetching default icon",
		slog.String("website_url", f.websiteURL),
	)

	iconURL, err := urllib.JoinBaseURLAndPath(urllib.RootURL(f.websiteURL), "favicon.ico")
	if err != nil {
		return nil, fmt.Errorf(`icon: unable to join root URL and path: %w`, err)
	}

	icon, err := f.downloadIcon(iconURL)
	if err != nil {
		return nil, err
	}

	return icon, nil
}

func (f *iconFinder) fetchFeedIcon() (*model.Icon, error) {
	slog.Debug("Fetching feed icon",
		slog.String("website_url", f.websiteURL),
		slog.String("feed_icon_url", f.feedIconURL),
	)

	iconURL, err := urllib.AbsoluteURL(f.websiteURL, f.feedIconURL)
	if err != nil {
		return nil, fmt.Errorf(`icon: unable to convert icon URL to absolute URL: %w`, err)
	}

	return f.downloadIcon(iconURL)
}

func (f *iconFinder) fetchIconsFromHTMLDocument() (*model.Icon, error) {
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

		if icon, err := f.downloadIcon(iconURL); err != nil {
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

func (f *iconFinder) downloadIcon(iconURL string) (*model.Icon, error) {
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
	}

	icon = resizeIcon(icon)

	return icon, nil
}

func resizeIcon(icon *model.Icon) *model.Icon {
	r := bytes.NewReader(icon.Content)

	if !slices.Contains([]string{"image/jpeg", "image/png", "image/gif"}, icon.MimeType) {
		slog.Info("icon isn't a png/gif/jpeg/ico, can't resize", slog.String("mimetype", icon.MimeType))
		return icon
	}

	// Don't resize icons that we can't decode, or that already have the right size.
	config, _, err := image.DecodeConfig(r)
	if err != nil {
		slog.Warn("unable to decode the metadata of the icon", slog.Any("error", err))
		return icon
	}
	if config.Height <= 32 && config.Width <= 32 {
		slog.Debug("icon don't need to be rescaled", slog.Int("height", config.Height), slog.Int("width", config.Width))
		return icon
	}

	r.Seek(0, io.SeekStart)

	var src image.Image
	switch icon.MimeType {
	case "image/jpeg":
		src, err = jpeg.Decode(r)
	case "image/png":
		src, err = png.Decode(r)
	case "image/gif":
		src, err = gif.Decode(r)
	}
	if err != nil || src == nil {
		slog.Warn("unable to decode the icon", slog.Any("error", err))
		return icon
	}

	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	var b bytes.Buffer
	if err = png.Encode(io.Writer(&b), dst); err != nil {
		slog.Warn("unable to encode the new icon", slog.Any("error", err))
	}

	icon.Content = b.Bytes()
	icon.MimeType = "image/png"
	return icon
}

func findIconURLsFromHTMLDocument(body io.Reader, contentType string) ([]string, error) {
	htmlDocumentReader, err := encoding.NewCharsetReader(body, contentType)
	if err != nil {
		return nil, fmt.Errorf("icon: unable to create charset reader: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(htmlDocumentReader)
	if err != nil {
		return nil, fmt.Errorf("icon: unable to read document: %v", err)
	}

	queries := []string{
		"link[rel='icon' i][href]",
		"link[rel='shortcut icon' i][href]",
		"link[rel='icon shortcut' i][href]",
		"link[rel='apple-touch-icon-precomposed.png'][href]",
	}

	var iconURLs []string
	for _, query := range queries {
		slog.Debug("Searching icon URL in HTML document", slog.String("query", query))

		for _, s := range doc.Find(query).EachIter() {
			href, _ := s.Attr("href")
			if iconURL := strings.TrimSpace(href); iconURL != "" {
				iconURLs = append(iconURLs, iconURL)
				slog.Debug("Found icon URL in HTML document",
					slog.String("query", query),
					slog.String("icon_url", iconURL))
			}
		}
	}

	return iconURLs, nil
}

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs#syntax
// data:[<mediatype>][;encoding],<data>
// we consider <mediatype> to be mandatory, and it has to start with `image/`.
// we consider `base64`, `utf8` and the empty string to be the only valid encodings
func parseImageDataURL(value string) (*model.Icon, error) {
	re := regexp.MustCompile(`^data:` +
		`(?P<mediatype>image/[^;,]+)` +
		`(?:;(?P<encoding>base64|utf8))?` +
		`,(?P<data>.+)$`)

	matches := re.FindStringSubmatch(value)
	if matches == nil {
		return nil, fmt.Errorf(`icon: invalid data URL %q`, value)
	}

	mediaType := matches[re.SubexpIndex("mediatype")]
	encoding := matches[re.SubexpIndex("encoding")]
	data := matches[re.SubexpIndex("data")]

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
	}

	return &model.Icon{
		Hash:     crypto.HashFromBytes(blob),
		Content:  blob,
		MimeType: mediaType,
	}, nil
}
