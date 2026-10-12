// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"net/http"
	"strings"
)

// MergeCookies merges Set-Cookie response headers into an existing cookie string.
// The existing cookie string uses the request Cookie header format: "name=value; name2=value2".
// Each Set-Cookie header is parsed for its name=value pair; existing values are updated,
// new names are appended. Returns the merged cookie string.
func MergeCookies(existing string, setCookieHeaders []string) string {
	// Parse existing cookies into an ordered map (preserve insertion order via slice).
	type cookiePair struct {
		name  string
		value string
	}
	var pairs []cookiePair
	index := make(map[string]int)

	for _, part := range strings.Split(existing, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		eqIdx := strings.IndexByte(part, '=')
		if eqIdx <= 0 {
			continue
		}
		name := strings.TrimSpace(part[:eqIdx])
		value := strings.TrimSpace(part[eqIdx+1:])
		if name == "" {
			continue
		}
		if _, ok := index[name]; !ok {
			index[name] = len(pairs)
			pairs = append(pairs, cookiePair{name, value})
		}
	}

	// Parse each Set-Cookie header and update or append.
	for _, header := range setCookieHeaders {
		// http.ParseSetCookie is available in Go 1.23+; fall back to manual parsing for safety.
		parsed := parseSetCookieNameValue(header)
		if parsed == nil {
			continue
		}
		if i, ok := index[parsed.Name]; ok {
			pairs[i].value = parsed.Value
		} else {
			index[parsed.Name] = len(pairs)
			pairs = append(pairs, cookiePair{parsed.Name, parsed.Value})
		}
	}

	// Reconstruct cookie string.
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, p.name+"="+p.value)
	}
	return strings.Join(parts, "; ")
}

// parseSetCookieNameValue extracts the name and value from a Set-Cookie header value.
// It returns nil if the header is malformed.
func parseSetCookieNameValue(header string) *http.Cookie {
	// A Set-Cookie value looks like "name=value; Path=/; HttpOnly; ..."
	// We only need the first token.
	first := header
	if idx := strings.IndexByte(header, ';'); idx >= 0 {
		first = header[:idx]
	}
	first = strings.TrimSpace(first)
	eqIdx := strings.IndexByte(first, '=')
	if eqIdx <= 0 {
		return nil
	}
	name := strings.TrimSpace(first[:eqIdx])
	value := strings.TrimSpace(first[eqIdx+1:])
	if name == "" {
		return nil
	}
	return &http.Cookie{Name: name, Value: value}
}
