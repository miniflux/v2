// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rss // import "miniflux.app/v2/internal/reader/rss"

import (
	"bytes"
	"fmt"
	"io"
	"regexp"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/xml"
)

// rssDefaultNamespaceRe matches a default (unprefixed) xmlns declaration on
// the root <rss> element, e.g. <rss xmlns="http://backend.userland.com/rss2">.
// It intentionally does not match prefixed namespaces like xmlns:atom="...".
var rssDefaultNamespaceRe = regexp.MustCompile(`(<rss\b[^>]*?)\s+xmlns="[^"]*"`)

// stripRSSDefaultNamespace removes a default namespace declared on the root
// <rss> element. Some feeds bind all elements to a default namespace (often
// the legacy Userland RSS2 namespace), which prevents them from matching the
// unprefixed "rss ..." struct tags used throughout this package, since
// DefaultSpace only applies to elements that don't already carry a namespace.
// Stripping the declaration lets those elements fall back to DefaultSpace as
// usual, while prefixed extension namespaces are left untouched.
func stripRSSDefaultNamespace(data []byte) []byte {
	return rssDefaultNamespaceRe.ReplaceAll(data, []byte("$1"))
}

// Parse returns a normalized feed struct from a RSS feed.
func Parse(baseURL string, data io.ReadSeeker) (*model.Feed, error) {
	body, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("rss: unable to read feed: %w", err)
	}
	body = stripRSSDefaultNamespace(body)

	rssFeed := new(rss)
	decoder := xml.NewXMLDecoder(bytes.NewReader(body))
	decoder.DefaultSpace = "rss"
	if err := decoder.Decode(rssFeed); err != nil {
		return nil, fmt.Errorf("rss: unable to parse feed: %w", err)
	}
	adapter := &rssAdapter{rssFeed}
	return adapter.buildFeed(baseURL), nil
}
