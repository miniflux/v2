// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package url

import "net/url"
import "fmt"
import "strings"

// GetAbsoluteURL converts the input URL as absolute URL if necessary.
func GetAbsoluteURL(baseURL, input string) (string, error) {
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

// GetRootURL returns absolute URL without the path.
func GetRootURL(websiteURL string) string {
	if strings.HasPrefix(websiteURL, "//") {
		websiteURL = "https://" + websiteURL[2:]
	}

	absoluteURL, err := GetAbsoluteURL(websiteURL, "")
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
