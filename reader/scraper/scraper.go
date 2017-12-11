// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper

import (
	"errors"

	"github.com/miniflux/miniflux2/http"
	"github.com/miniflux/miniflux2/reader/readability"
	"github.com/miniflux/miniflux2/reader/sanitizer"
)

// Fetch download a web page a returns relevant contents.
func Fetch(websiteURL string) (string, error) {
	client := http.NewClient(websiteURL)
	response, err := client.Get()
	if err != nil {
		return "", err
	}

	if response.HasServerFailure() {
		return "", errors.New("unable to download web page")
	}

	page, err := response.NormalizeBodyEncoding()
	if err != nil {
		return "", err
	}

	content, err := readability.ExtractContent(page)
	if err != nil {
		return "", err
	}

	return sanitizer.Sanitize(websiteURL, content), nil
}
