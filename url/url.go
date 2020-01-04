// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package url // import "miniflux.app/url"

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// IsAbsoluteURL returns true if the link is absolute.
func IsAbsoluteURL(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// AbsoluteURL converts the input URL as absolute URL if necessary.
func AbsoluteURL(baseURL, input string) (string, error) {
	if strings.HasPrefix(input, "//") {
		input = "https://" + input[2:]
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("unable to parse input URL: %v", err)
	}

	if u.IsAbs() {
		return u.String(), nil
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse base URL: %v", err)
	}

	return base.ResolveReference(u).String(), nil
}

// RootURL returns absolute URL without the path.
func RootURL(websiteURL string) string {
	if strings.HasPrefix(websiteURL, "//") {
		websiteURL = "https://" + websiteURL[2:]
	}

	absoluteURL, err := AbsoluteURL(websiteURL, "")
	if err != nil {
		return websiteURL
	}

	u, err := url.Parse(absoluteURL)
	if err != nil {
		return absoluteURL
	}

	return u.Scheme + "://" + u.Host + "/"
}

// IsHTTPS returns true if the URL is using HTTPS.
func IsHTTPS(websiteURL string) bool {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return false
	}

	return strings.ToLower(parsedURL.Scheme) == "https"
}

// Domain returns only the domain part of the given URL.
func Domain(websiteURL string) string {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return websiteURL
	}

	return parsedURL.Host
}

// RequestURI returns the encoded URI to be used in HTTP requests.
func RequestURI(websiteURL string) string {
	u, err := url.Parse(websiteURL)
	if err != nil {
		return websiteURL
	}

	queryValues := u.Query()
	u.RawQuery = "" // Clear RawQuery to make sure it's encoded properly.
	u.Fragment = "" // Clear fragment because Web browsers strip #fragment before sending the URL to a web server.

	var buf strings.Builder
	buf.WriteString(u.String())

	if len(queryValues) > 0 {
		buf.WriteByte('?')

		// Sort keys.
		keys := make([]string, 0, len(queryValues))
		for k := range queryValues {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		i := 0
		for _, key := range keys {
			keyEscaped := url.QueryEscape(key)
			values := queryValues[key]
			for _, value := range values {
				if i > 0 {
					buf.WriteByte('&')
				}
				buf.WriteString(keyEscaped)

				// As opposed to the standard library, we append the = only if the value is not empty.
				if value != "" {
					buf.WriteByte('=')
					buf.WriteString(url.QueryEscape(value))
				}

				i++
			}
		}
	}

	return buf.String()
}
