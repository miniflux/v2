// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package rdf

import (
	"encoding/xml"

	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/reader/processor"
	"github.com/miniflux/miniflux2/reader/sanitizer"

	"github.com/miniflux/miniflux2/model"
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
}

func (r *rdfItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.Title = sanitizer.StripTags(r.Title)
	entry.Author = sanitizer.StripTags(r.Creator)
	entry.URL = r.Link
	entry.Content = processor.ItemContentProcessor(entry.URL, r.Description)
	entry.Hash = getHash(r)
	return entry
}

func getHash(r *rdfItem) string {
	value := r.Link
	if value == "" {
		value = r.Title + r.Description
	}

	return helper.Hash(value)
}
