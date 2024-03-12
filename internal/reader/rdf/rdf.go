// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/v2/internal/reader/rdf"

import (
	"encoding/xml"
	"html"
	"log/slog"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/date"
	"miniflux.app/v2/internal/reader/dublincore"
	"miniflux.app/v2/internal/reader/sanitizer"
	"miniflux.app/v2/internal/urllib"
)

type rdfFeed struct {
	XMLName xml.Name  `xml:"RDF"`
	Title   string    `xml:"channel>title"`
	Link    string    `xml:"channel>link"`
	Items   []rdfItem `xml:"item"`
	dublincore.DublinCoreFeedElement
}

func (r *rdfFeed) Transform(baseURL string) *model.Feed {
	var err error
	feed := new(model.Feed)
	feed.Title = sanitizer.StripTags(r.Title)
	feed.FeedURL = baseURL
	feed.SiteURL, err = urllib.AbsoluteURL(baseURL, r.Link)
	if err != nil {
		feed.SiteURL = r.Link
	}

	for _, item := range r.Items {
		entry := item.Transform()
		if entry.Author == "" && r.DublinCoreCreator != "" {
			entry.Author = r.GetSanitizedCreator()
		}

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := urllib.AbsoluteURL(feed.SiteURL, entry.URL)
			if err == nil {
				entry.URL = entryURL
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

type rdfItem struct {
	Title       string `xml:"http://purl.org/rss/1.0/ title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	dublincore.DublinCoreItemElement
}

func (r *rdfItem) Transform() *model.Entry {
	entry := model.NewEntry()
	entry.Title = r.entryTitle()
	entry.Author = r.entryAuthor()
	entry.URL = r.entryURL()
	entry.Content = r.entryContent()
	entry.Hash = r.entryHash()
	entry.Date = r.entryDate()

	if entry.Title == "" {
		entry.Title = entry.URL
	}
	return entry
}

func (r *rdfItem) entryTitle() string {
	for _, title := range []string{r.Title, r.DublinCoreTitle} {
		title = strings.TrimSpace(title)
		if title != "" {
			return html.UnescapeString(title)
		}
	}
	return ""
}

func (r *rdfItem) entryContent() string {
	switch {
	case r.DublinCoreContent != "":
		return r.DublinCoreContent
	default:
		return r.Description
	}
}

func (r *rdfItem) entryAuthor() string {
	return r.GetSanitizedCreator()
}

func (r *rdfItem) entryURL() string {
	return strings.TrimSpace(r.Link)
}

func (r *rdfItem) entryDate() time.Time {
	if r.DublinCoreDate != "" {
		result, err := date.Parse(r.DublinCoreDate)
		if err != nil {
			slog.Debug("Unable to parse date from RDF feed",
				slog.String("date", r.DublinCoreDate),
				slog.String("link", r.Link),
				slog.Any("error", err),
			)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (r *rdfItem) entryHash() string {
	value := r.Link
	if value == "" {
		value = r.Title + r.Description
	}

	return crypto.Hash(value)
}
