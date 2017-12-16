// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper

import (
	"errors"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/miniflux/miniflux/http"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/reader/readability"
	"github.com/miniflux/miniflux/url"
)

// Fetch download a web page a returns relevant contents.
func Fetch(websiteURL, rules string) (string, error) {
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

	// The entry URL could be a redirect somewhere else.
	websiteURL = response.EffectiveURL

	if rules == "" {
		rules = getPredefinedScraperRules(websiteURL)
	}

	var content string
	if rules != "" {
		logger.Debug(`[Scraper] Using rules "%s" for "%s"`, rules, websiteURL)
		content, err = scrapContent(page, rules)
	} else {
		logger.Debug(`[Scraper] Using readability for "%s"`, websiteURL)
		content, err = readability.ExtractContent(page)
	}

	if err != nil {
		return "", err
	}

	return content, nil
}

func scrapContent(page io.Reader, rules string) (string, error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", err
	}

	contents := ""
	document.Find(rules).Each(func(i int, s *goquery.Selection) {
		var content string

		// For some inline elements, we get the parent.
		if s.Is("img") {
			content, _ = s.Parent().Html()
		} else {
			content, _ = s.Html()
		}

		contents += content
	})

	return contents, nil
}

func getPredefinedScraperRules(websiteURL string) string {
	urlDomain := url.Domain(websiteURL)

	for domain, rules := range predefinedRules {
		if strings.Contains(urlDomain, domain) {
			return rules
		}
	}

	return ""
}
