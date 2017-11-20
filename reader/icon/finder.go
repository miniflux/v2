// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package icon

import (
	"fmt"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/http"
	"github.com/miniflux/miniflux2/reader/url"
	"io"
	"io/ioutil"
	"log"

	"github.com/PuerkitoBio/goquery"
)

// FindIcon try to find the website's icon.
func FindIcon(websiteURL string) (*model.Icon, error) {
	rootURL := url.GetRootURL(websiteURL)
	client := http.NewHttpClient(rootURL)
	response, err := client.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to download website index page: %v", err)
	}

	if response.HasServerFailure() {
		return nil, fmt.Errorf("unable to download website index page: status=%d", response.StatusCode)
	}

	iconURL, err := parseDocument(rootURL, response.Body)
	if err != nil {
		return nil, err
	}

	log.Println("[FindIcon] Fetching icon =>", iconURL)
	icon, err := downloadIcon(iconURL)
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
		return "", fmt.Errorf("unable to read document: %v", err)
	}

	var iconURL string
	for _, query := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				iconURL = href
			}
		})

		if iconURL != "" {
			break
		}
	}

	if iconURL == "" {
		iconURL = url.GetRootURL(websiteURL) + "favicon.ico"
	} else {
		iconURL, _ = url.GetAbsoluteURL(websiteURL, iconURL)
	}

	return iconURL, nil
}

func downloadIcon(iconURL string) (*model.Icon, error) {
	client := http.NewHttpClient(iconURL)
	response, err := client.Get()
	if err != nil {
		return nil, fmt.Errorf("unable to download iconURL: %v", err)
	}

	if response.HasServerFailure() {
		return nil, fmt.Errorf("unable to download icon: status=%d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read downloaded icon: %v", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("downloaded icon is empty, iconURL=%s", iconURL)
	}

	icon := &model.Icon{
		Hash:     helper.HashFromBytes(body),
		MimeType: response.ContentType,
		Content:  body,
	}

	return icon, nil
}
