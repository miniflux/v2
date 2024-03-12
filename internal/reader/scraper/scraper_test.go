// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGetPredefinedRules(t *testing.T) {
	if getPredefinedScraperRules("http://www.phoronix.com/") == "" {
		t.Error("Unable to find rule for phoronix.com")
	}

	if getPredefinedScraperRules("https://www.linux.com/") == "" {
		t.Error("Unable to find rule for linux.com")
	}

	if getPredefinedScraperRules("https://linux.com/") == "" {
		t.Error("Unable to find rule for linux.com")
	}

	if getPredefinedScraperRules("https://example.org/") != "" {
		t.Error("A rule not defined should not return anything")
	}
}

func TestWhitelistedContentTypes(t *testing.T) {
	scenarios := map[string]bool{
		"text/html":                            true,
		"TeXt/hTmL":                            true,
		"application/xhtml+xml":                true,
		"text/html; charset=utf-8":             true,
		"application/xhtml+xml; charset=utf-8": true,
		"text/css":                             false,
		"application/javascript":               false,
		"image/png":                            false,
		"application/pdf":                      false,
	}

	for inputValue, expectedResult := range scenarios {
		actualResult := isAllowedContentType(inputValue)
		if actualResult != expectedResult {
			t.Errorf(`Unexpected result for content type whitelist, got "%v" instead of "%v"`, actualResult, expectedResult)
		}
	}
}

func TestSelectorRules(t *testing.T) {
	var ruleTestCases = map[string]string{
		"img.html":    "article > img",
		"iframe.html": "article > iframe",
		"p.html":      "article > p",
	}

	for filename, rule := range ruleTestCases {
		html, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			t.Fatalf(`Unable to read file %q: %v`, filename, err)
		}

		actualResult, err := findContentUsingCustomRules(bytes.NewReader(html), rule)
		if err != nil {
			t.Fatalf(`Scraping error for %q - %q: %v`, filename, rule, err)
		}

		expectedResult, err := os.ReadFile("testdata/" + filename + "-result")
		if err != nil {
			t.Fatalf(`Unable to read file %q: %v`, filename, err)
		}

		if actualResult != strings.TrimSpace(string(expectedResult)) {
			t.Errorf(`Unexpected result for %q, got "%s" instead of "%s"`, rule, actualResult, expectedResult)
		}
	}
}
