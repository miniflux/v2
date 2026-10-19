// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package urllib // import "miniflux.app/v2/internal/urllib"

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

var (
	rfc6598SharedAddressSpacePrefix = netip.MustParsePrefix("100.64.0.0/10")
	nat64WellKnownPrefix            = netip.MustParsePrefix("64:ff9b::/96")
	nonPublicIPPrefixes             = []netip.Prefix{
		netip.MustParsePrefix("168.63.129.16/32"),
		netip.MustParsePrefix("192.0.0.192/32"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("fec0::/10"),
	}
)

// IsRelativePath reports whether the link is a relative path (no scheme, host, or scheme-relative // form).
func IsRelativePath(link string) bool {
	if link == "" {
		return false
	}

	// Reject surrounding whitespace: browsers strip leading and trailing C0
	// control characters and spaces before parsing a target, while Go's url.Parse
	// treats spaces as path characters. Without this check, a target like
	// " //evil.com" is accepted here and becomes //evil.com in the browser.
	if link != strings.TrimSpace(link) {
		return false
	}

	// Reject backslashes: Go's url.Parse treats them as ordinary path
	// characters, but browsers normalize them to forward slashes, so a target
	// like "/\evil.com" would parse as relative here yet redirect to
	// //evil.com in the browser (open redirect).
	if strings.Contains(link, "\\") {
		return false
	}

	// Reject scheme-relative URLs, including forms with extra leading slashes
	// that url.Parse treats as a relative path but browsers resolve externally.
	if strings.HasPrefix(link, "//") {
		return false
	}

	if parsedURL, err := url.Parse(link); err == nil {
		// Only allow relative paths and ensure the URL doesn't have a host
		// component.
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

// hasSOCKSPrefix reports whether the URL string begins with an SOCKS5 or SOCKS5H scheme.
func hasSOCKSPrefix(inputURL string) bool {
	return strings.HasPrefix(inputURL, "socks5://") || strings.HasPrefix(inputURL, "socks5h://")
}

// IsAbsoluteURL reports whether the link is absolute and starts with an HTTP or HTTPS scheme.
func IsAbsoluteURL(inputURL string) bool {
	if !hasHTTPPrefix(inputURL) {
		return false
	}
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return false
	}
	return parsedURL.IsAbs()
}

// IsValidProxyURL reports whether the url is absolute, has a host and starts with an HTTP, HTTPS, SOCKS5 or SOCKS5H scheme.
func IsValidProxyURL(inputURL string) bool {
	if !hasHTTPPrefix(inputURL) && !hasSOCKSPrefix(inputURL) {
		return false
	}
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return false
	}
	return parsedURL.IsAbs() && parsedURL.Host != ""
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

	finalURL, err := url.JoinPath(baseURL, path)
	if err != nil {
		return "", fmt.Errorf("unable to join base URL %s and path %s: %w", baseURL, path, err)
	}

	return finalURL, nil
}

// IsNonPublicIP returns true if the given IP is not safe to access from the network.
func IsNonPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return true
	}

	return isNonPublicAddr(addr.Unmap())
}

func isNonPublicAddr(addr netip.Addr) bool {
	if nat64WellKnownPrefix.Contains(addr) {
		addrBytes := addr.As16()
		return isNonPublicAddr(netip.AddrFrom4([4]byte(addrBytes[12:])))
	}

	if rfc6598SharedAddressSpacePrefix.Contains(addr) {
		return true
	}

	for _, prefix := range nonPublicIPPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}

	return addr.IsPrivate() ||
		addr.IsLoopback() ||
		addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() ||
		addr.IsUnspecified()
}
