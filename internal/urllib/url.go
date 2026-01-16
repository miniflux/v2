// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package urllib // import "miniflux.app/v2/internal/urllib"

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
)

// IsRelativePath reports whether the link is a relative path (no scheme, host, or scheme-relative // form).
func IsRelativePath(link string) bool {
	if link == "" {
		return false
	}
	if parsedURL, err := url.Parse(link); err == nil {
		// Only allow relative paths (not scheme-relative URLs like //example.org)
		// and ensure the URL doesn't have a host component
		if !parsedURL.IsAbs() && parsedURL.Host == "" && parsedURL.Scheme == "" {
			return true
		}
	}
	return false
}

// hasHTTPPrefix reports whether the URL string begins with an HTTP or HTTPS scheme.
func hasHTTPPrefix(inputURL string) bool {
	return strings.HasPrefix(inputURL, "https://") || strings.HasPrefix(inputURL, "http://")
}

// IsAbsoluteURL reports whether the link is absolute.
func IsAbsoluteURL(inputURL string) bool {
	if hasHTTPPrefix(inputURL) {
		return true
	}
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return false
	}
	return parsedURL.IsAbs()
}

// resolveToAbsoluteURL resolves a relative URL using a base URL, parsing the base only if needed.
func resolveToAbsoluteURL(parsedBaseURL *url.URL, baseURL, relativeURL string) (string, error) {
	// Avoid parsing the relative URL if it's already absolute
	if strings.HasPrefix(relativeURL, "//") {
		return "https:" + relativeURL, nil
	}
	if hasHTTPPrefix(relativeURL) {
		return relativeURL, nil
	}

	// Parse the relative URL and check if it's already absolute
	parsedRelativeURL, err := url.Parse(relativeURL)
	if err != nil {
		return "", fmt.Errorf("unable to parse relative URL: %w", err)
	}
	if parsedRelativeURL.IsAbs() {
		return relativeURL, nil
	}

	// Parse the base URL if not already parsed
	if parsedBaseURL == nil {
		parsedBaseURL, err = url.Parse(baseURL)
		if err != nil {
			return "", fmt.Errorf("unable to parse base URL: %w", err)
		}
	}

	return parsedBaseURL.ResolveReference(parsedRelativeURL).String(), nil
}

// ResolveToAbsoluteURL resolves a relative URL against a base URL and returns the absolute URL.
func ResolveToAbsoluteURL(baseURL, relativeURL string) (string, error) {
	return resolveToAbsoluteURL(nil, baseURL, relativeURL)
}

// ResolveToAbsoluteURLWithParsedBaseURL resolves a relative URL using a pre-parsed base URL and returns the absolute URL.
func ResolveToAbsoluteURLWithParsedBaseURL(parsedBaseURL *url.URL, relativeURL string) (string, error) {
	return resolveToAbsoluteURL(parsedBaseURL, "", relativeURL)
}

// RootURL returns the scheme and host of the given URL with a trailing slash.
func RootURL(websiteURL string) string {
	if websiteURL == "" {
		return ""
	}

	if strings.HasPrefix(websiteURL, "//") {
		websiteURL = "https://" + websiteURL[2:]
	}

	u, err := url.Parse(websiteURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return websiteURL
	}

	u.Fragment = ""
	u.RawQuery = ""
	u.Path = "/"
	u.RawPath = ""

	return u.Scheme + "://" + u.Host + "/"
}

// IsHTTPS reports whether the URL uses HTTPS.
func IsHTTPS(websiteURL string) bool {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return false
	}

	return strings.EqualFold(parsedURL.Scheme, "https")
}

// Domain returns the host component of the given URL.
func Domain(websiteURL string) string {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return websiteURL
	}

	return parsedURL.Host
}

// DomainWithoutWWW returns the host component without a leading "www." prefix when present.
func DomainWithoutWWW(websiteURL string) string {
	return strings.TrimPrefix(Domain(websiteURL), "www.")
}

// JoinBaseURLAndPath joins a base URL and a path segment into a single URL string.
func JoinBaseURLAndPath(baseURL, path string) (string, error) {
	if baseURL == "" {
		return "", errors.New("empty base URL")
	}

	if path == "" {
		return "", errors.New("empty path")
	}

	_, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	finalURL, err := url.JoinPath(baseURL, path)
	if err != nil {
		return "", fmt.Errorf("unable to join base URL %s and path %s: %w", baseURL, path, err)
	}

	return finalURL, nil
}

// ResolvesToPrivateIP resolves a hostname and reports whether any resolved IP address is non-public.
func ResolvesToPrivateIP(host string) (bool, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, err
	}

	if slices.ContainsFunc(ips, isNonPublicIP) {
		return true, nil
	}

	return false, nil
}

// isNonPublicIP returns true if the given IP is private, loopback,
// link-local, multicast, or unspecified.
func isNonPublicIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	return ip.IsPrivate() ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified()
}
