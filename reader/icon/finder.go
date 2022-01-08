// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package icon // import "miniflux.app/reader/icon"

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/http/client"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
)

// FindIcon try to find the website's icon.
func FindIcon(websiteURL, userAgent string, fetchViaProxy, allowSelfSignedCertificates bool) (*model.Icon, error) {
	rootURL := url.RootURL(websiteURL)
	logger.Debug("[FindIcon] Trying to find an icon: rootURL=%q websiteURL=%q userAgent=%q", rootURL, websiteURL, userAgent)

	clt := client.NewClientWithConfig(rootURL, config.Opts)
	clt.WithUserAgent(userAgent)
	clt.AllowSelfSignedCertificates = allowSelfSignedCertificates

	if fetchViaProxy {
		clt.WithProxy()
	}

	response, err := clt.Get()
	if err != nil {
		return nil, fmt.Errorf("icon: unable to download website index page: %v", err)
	}

	if response.HasServerFailure() {
		return nil, fmt.Errorf("icon: unable to download website index page: status=%d", response.StatusCode)
	}

	iconURL, err := parseDocument(rootURL, response.Body)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(iconURL, "data:") {
		return parseImageDataURL(iconURL)
	}

	logger.Debug("[FindIcon] Fetching icon => %s", iconURL)
	icon, err := downloadIcon(iconURL, userAgent, fetchViaProxy, allowSelfSignedCertificates)
	if err != nil {
		return nil, err
	}

	return icon, nil
}

func parseDocument(websiteURL string, data io.Reader) (string, error) {
	queries := []string{
		"link[rel='shortcut icon']",
		"link[rel='Shortcut Icon']",
		"link[rel='icon shortcut']",
		"link[rel='icon']",
	}

	doc, err := goquery.NewDocumentFromReader(data)
	if err != nil {
		return "", fmt.Errorf("icon: unable to read document: %v", err)
	}

	var iconURL string
	for _, query := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				iconURL = strings.TrimSpace(href)
			}
		})

		if iconURL != "" {
			break
		}
	}

	if iconURL == "" {
		iconURL = url.RootURL(websiteURL) + "favicon.ico"
	} else {
		iconURL, _ = url.AbsoluteURL(websiteURL, iconURL)
	}

	return iconURL, nil
}

func downloadIcon(iconURL, userAgent string, fetchViaProxy, allowSelfSignedCertificates bool) (*model.Icon, error) {
	clt := client.NewClientWithConfig(iconURL, config.Opts)
	clt.WithUserAgent(userAgent)
	clt.AllowSelfSignedCertificates = allowSelfSignedCertificates
	if fetchViaProxy {
		clt.WithProxy()
	}

	response, err := clt.Get()
	if err != nil {
		return nil, fmt.Errorf("icon: unable to download iconURL: %v", err)
	}

	if response.HasServerFailure() {
		return nil, fmt.Errorf("icon: unable to download icon: status=%d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("icon: unable to read downloaded icon: %v", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("icon: downloaded icon is empty, iconURL=%s", iconURL)
	}

	icon := &model.Icon{
		Hash:     crypto.HashFromBytes(body),
		MimeType: response.ContentType,
		Content:  body,
	}

	return icon, nil
}

func parseImageDataURL(value string) (*model.Icon, error) {
	colon := strings.Index(value, ":")
	semicolon := strings.Index(value, ";")
	comma := strings.Index(value, ",")

	if colon <= 0 || semicolon <= 0 || comma <= 0 {
		return nil, fmt.Errorf(`icon: invalid data url "%s"`, value)
	}

	mimeType := value[colon+1 : semicolon]
	encoding := value[semicolon+1 : comma]
	data := value[comma+1:]

	if encoding != "base64" {
		return nil, fmt.Errorf(`icon: unsupported data url encoding "%s"`, value)
	}

	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf(`icon: invalid mime type "%s"`, mimeType)
	}

	blob, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf(`icon: invalid data "%s" (%v)`, value, err)
	}

	if len(blob) == 0 {
		return nil, fmt.Errorf(`icon: empty data "%s"`, value)
	}

	icon := &model.Icon{
		Hash:     crypto.HashFromBytes(blob),
		Content:  blob,
		MimeType: mimeType,
	}

	return icon, nil
}
