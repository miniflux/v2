// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom

import (
	"encoding/xml"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed/date"
	"github.com/miniflux/miniflux2/reader/processor"
	"github.com/miniflux/miniflux2/reader/sanitizer"
	"log"
	"strconv"
	"strings"
	"time"
)

type AtomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Author  Author      `xml:"author"`
	Links   []Link      `xml:"link"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomEntry struct {
	ID         string     `xml:"id"`
	Title      string     `xml:"title"`
	Updated    string     `xml:"updated"`
	Links      []Link     `xml:"link"`
	Summary    string     `xml:"summary"`
	Content    Content    `xml:"content"`
	MediaGroup MediaGroup `xml:"http://search.yahoo.com/mrss/ group"`
	Author     Author     `xml:"author"`
}

type Author struct {
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type Link struct {
	Url    string `xml:"href,attr"`
	Type   string `xml:"type,attr"`
	Rel    string `xml:"rel,attr"`
	Length string `xml:"length,attr"`
}

type Content struct {
	Type string `xml:"type,attr"`
	Data string `xml:",chardata"`
	Xml  string `xml:",innerxml"`
}

type MediaGroup struct {
	Description string `xml:"http://search.yahoo.com/mrss/ description"`
}

func (a *AtomFeed) getSiteURL() string {
	for _, link := range a.Links {
		if strings.ToLower(link.Rel) == "alternate" {
			return link.Url
		}

		if link.Rel == "" && link.Type == "" {
			return link.Url
		}
	}

	return ""
}

func (a *AtomFeed) getFeedURL() string {
	for _, link := range a.Links {
		if strings.ToLower(link.Rel) == "self" {
			return link.Url
		}
	}

	return ""
}

func (a *AtomFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.FeedURL = a.getFeedURL()
	feed.SiteURL = a.getSiteURL()
	feed.Title = sanitizer.StripTags(a.Title)

	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, entry := range a.Entries {
		item := entry.Transform()
		if item.Author == "" {
			item.Author = a.GetAuthor()
		}

		feed.Entries = append(feed.Entries, item)
	}

	return feed
}

func (a *AtomFeed) GetAuthor() string {
	return getAuthor(a.Author)
}

func (e *AtomEntry) GetDate() time.Time {
	if e.Updated != "" {
		result, err := date.Parse(e.Updated)
		if err != nil {
			log.Println(err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (e *AtomEntry) GetURL() string {
	for _, link := range e.Links {
		if strings.ToLower(link.Rel) == "alternate" {
			return link.Url
		}

		if link.Rel == "" && link.Type == "" {
			return link.Url
		}
	}

	return ""
}

func (e *AtomEntry) GetAuthor() string {
	return getAuthor(e.Author)
}

func (e *AtomEntry) GetHash() string {
	for _, value := range []string{e.ID, e.GetURL()} {
		if value != "" {
			return helper.Hash(value)
		}
	}

	return ""
}

func (e *AtomEntry) GetContent() string {
	if e.Content.Type == "html" || e.Content.Type == "text" {
		return e.Content.Data
	}

	if e.Content.Type == "xhtml" {
		return e.Content.Xml
	}

	if e.Summary != "" {
		return e.Summary
	}

	if e.MediaGroup.Description != "" {
		return e.MediaGroup.Description
	}

	return ""
}

func (e *AtomEntry) GetEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)

	for _, link := range e.Links {
		if strings.ToLower(link.Rel) == "enclosure" {
			length, _ := strconv.Atoi(link.Length)
			enclosures = append(enclosures, &model.Enclosure{URL: link.Url, MimeType: link.Type, Size: length})
		}
	}

	return enclosures
}

func (e *AtomEntry) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = e.GetURL()
	entry.Date = e.GetDate()
	entry.Author = sanitizer.StripTags(e.GetAuthor())
	entry.Hash = e.GetHash()
	entry.Content = processor.ItemContentProcessor(entry.URL, e.GetContent())
	entry.Title = sanitizer.StripTags(strings.Trim(e.Title, " \n\t"))
	entry.Enclosures = e.GetEnclosures()

	if entry.Title == "" {
		entry.Title = entry.URL
	}

	return entry
}

func getAuthor(author Author) string {
	if author.Name != "" {
		return author.Name
	}

	if author.Email != "" {
		return author.Email
	}

	return ""
}
