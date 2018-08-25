// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rdf // import "miniflux.app/reader/rdf"

import (
	"encoding/xml"
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
	Creator string    `xml:"channel>creator"`
	Items   []rdfItem `xml:"item"`
}

func (r *rdfFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.Title = sanitizer.StripTags(r.Title)
	feed.SiteURL = r.Link

	for _, item := range r.Items {
		entry := item.Transform()
		if entry.Author == "" && r.Creator != "" {
			entry.Author = sanitizer.StripTags(r.Creator)
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
	Creator     string `xml:"creator"`
	Date        string `xml:"date"`
}

func (r *rdfItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.Title = strings.TrimSpace(r.Title)
	entry.Author = strings.TrimSpace(r.Creator)
	entry.URL = r.Link
	entry.Content = r.Description
	entry.Hash = getHash(r)
	entry.Date = getDate(r)
	return entry
}

func getDate(r *rdfItem) time.Time {
	if r.Date != "" {
		result, err := date.Parse(r.Date)
		if err != nil {
			logger.Error("rdf: %v", err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func getHash(r *rdfItem) string {
	value := r.Link
	if value == "" {
		value = r.Title + r.Description
	}

	return crypto.Hash(value)
}
