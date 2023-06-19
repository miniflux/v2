// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rdf // import "miniflux.app/reader/rdf"

import (
	"encoding/xml"
	"html"
	"strings"
	"time"

	"miniflux.app/crypto"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/date"
	"miniflux.app/reader/sanitizer"
	"miniflux.app/url"
)

type rdfFeed struct {
	XMLName xml.Name  `xml:"RDF"`
	Title   string    `xml:"channel>title"`
	Link    string    `xml:"channel>link"`
	Items   []rdfItem `xml:"item"`
	DublinCoreFeedElement
}

func (r *rdfFeed) Transform(baseURL string) *model.Feed {
	var err error
	feed := new(model.Feed)
	feed.Title = sanitizer.StripTags(r.Title)
	feed.FeedURL = baseURL
	feed.SiteURL, err = url.AbsoluteURL(baseURL, r.Link)
	if err != nil {
		feed.SiteURL = r.Link
	}

	for _, item := range r.Items {
		entry := item.Transform()
		if entry.Author == "" && r.DublinCoreCreator != "" {
			entry.Author = strings.TrimSpace(r.DublinCoreCreator)
		}

		if entry.URL == "" {
			entry.URL = feed.SiteURL
		} else {
			entryURL, err := url.AbsoluteURL(feed.SiteURL, entry.URL)
			if err == nil {
				entry.URL = entryURL
			}
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

type rdfItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	DublinCoreEntryElement
}

func (r *rdfItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.Title = r.entryTitle()
	entry.Author = r.entryAuthor()
	entry.URL = r.entryURL()
	entry.Content = r.entryContent()
	entry.Hash = r.entryHash()
	entry.Date = r.entryDate()
	return entry
}

func (r *rdfItem) entryTitle() string {
	return html.UnescapeString(strings.TrimSpace(r.Title))
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
	return strings.TrimSpace(r.DublinCoreCreator)
}

func (r *rdfItem) entryURL() string {
	return strings.TrimSpace(r.Link)
}

func (r *rdfItem) entryDate() time.Time {
	if r.DublinCoreDate != "" {
		result, err := date.Parse(r.DublinCoreDate)
		if err != nil {
			logger.Error("rdf: %v (entry link = %s)", err, r.Link)
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
