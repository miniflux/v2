// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/v2/internal/reader/rdf"

import (
	"html"
	"log/slog"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type RDFAdapter struct {
	rdf *RDF
}

func NewRDFAdapter(rdf *RDF) *RDFAdapter {
	return &RDFAdapter{rdf}
}

func (r *RDFAdapter) BuildFeed(feedURL string) *model.Feed {
	feed := &model.Feed{
		Title:   stripTags(r.rdf.Channel.Title),
		FeedURL: feedURL,
	}

	if feed.Title == "" {
		feed.Title = feedURL
	}

	if siteURL, err := urllib.AbsoluteURL(feedURL, r.rdf.Channel.Link); err != nil {
		feed.SiteURL = r.rdf.Channel.Link
	} else {
		feed.SiteURL = siteURL
	}

	for _, item := range r.rdf.Items {
		entry := model.NewEntry()
		itemLink := strings.TrimSpace(item.Link)

		// Populate the entry URL.
		if itemLink == "" {
			entry.URL = feed.SiteURL // Fallback to the feed URL if the entry URL is empty.
		} else if entryURL, err := urllib.AbsoluteURL(feed.SiteURL, itemLink); err == nil {
			entry.URL = entryURL
		} else {
			entry.URL = itemLink
		}

		// Populate the entry title.
		for _, title := range []string{item.Title, item.DublinCoreTitle} {
			title = strings.TrimSpace(title)
			if title != "" {
				entry.Title = html.UnescapeString(title)
				break
			}
		}

		// If the entry title is empty, we use the entry URL as a fallback.
		if entry.Title == "" {
			entry.Title = entry.URL
		}

		// Populate the entry content.
		if item.DublinCoreContent != "" {
			entry.Content = item.DublinCoreContent
		} else {
			entry.Content = item.Description
		}

		// Generate the entry hash.
		hashValue := itemLink
		if hashValue == "" {
			hashValue = item.Title + item.Description // Fallback to the title and description if the link is empty.
		}

		entry.Hash = crypto.Hash(hashValue)

		// Populate the entry date.
		entry.Date = time.Now()
		if item.DublinCoreDate != "" {
			if itemDate, err := date.Parse(item.DublinCoreDate); err != nil {
				slog.Debug("Unable to parse date from RDF feed",
					slog.String("date", item.DublinCoreDate),
					slog.String("link", itemLink),
					slog.Any("error", err),
				)
			} else {
				entry.Date = itemDate
			}
		}

		// Populate the entry author.
		switch {
		case item.DublinCoreCreator != "":
			entry.Author = stripTags(item.DublinCoreCreator)
		case r.rdf.Channel.DublinCoreCreator != "":
			entry.Author = stripTags(r.rdf.Channel.DublinCoreCreator)
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func stripTags(value string) string {
	return strings.TrimSpace(sanitizer.StripTags(value))
}
