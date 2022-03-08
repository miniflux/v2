// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package atom // import "miniflux.app/reader/atom"

import (
	"encoding/base64"
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

// Specs: http://web.archive.org/web/20060811235523/http://www.mnot.net/drafts/draft-nottingham-atom-format-02.html
type atom03Feed struct {
	ID      string        `xml:"id"`
	Title   atom03Text    `xml:"title"`
	Author  atomPerson    `xml:"author"`
	Links   atomLinks     `xml:"link"`
	Entries []atom03Entry `xml:"entry"`
}

func (a *atom03Feed) Transform(baseURL string) *model.Feed {
	var err error

	feed := new(model.Feed)

	feedURL := a.Links.firstLinkWithRelation("self")
	feed.FeedURL, err = url.AbsoluteURL(baseURL, feedURL)
	if err != nil {
		feed.FeedURL = feedURL
	}

	siteURL := a.Links.originalLink()
	feed.SiteURL, err = url.AbsoluteURL(baseURL, siteURL)
	if err != nil {
		feed.SiteURL = siteURL
	}

	feed.Title = a.Title.String()
	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, entry := range a.Entries {
		item := entry.Transform()
		entryURL, err := url.AbsoluteURL(feed.SiteURL, item.URL)
		if err == nil {
			item.URL = entryURL
		}

		if item.Author == "" {
			item.Author = a.Author.String()
		}

		if item.Title == "" {
			item.Title = sanitizer.TruncateHTML(item.Content, 100)
		}

		if item.Title == "" {
			item.Title = item.URL
		}

		feed.Entries = append(feed.Entries, item)
	}

	return feed
}

type atom03Entry struct {
	ID       string     `xml:"id"`
	Title    atom03Text `xml:"title"`
	Modified string     `xml:"modified"`
	Issued   string     `xml:"issued"`
	Created  string     `xml:"created"`
	Links    atomLinks  `xml:"link"`
	Summary  atom03Text `xml:"summary"`
	Content  atom03Text `xml:"content"`
	Author   atomPerson `xml:"author"`
}

func (a *atom03Entry) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = a.Links.originalLink()
	entry.Date = a.entryDate()
	entry.Author = a.Author.String()
	entry.Hash = a.entryHash()
	entry.Content = a.entryContent()
	entry.Title = a.entryTitle()
	return entry
}

func (a *atom03Entry) entryTitle() string {
	return sanitizer.StripTags(a.Title.String())
}

func (a *atom03Entry) entryContent() string {
	content := a.Content.String()
	if content != "" {
		return content
	}

	summary := a.Summary.String()
	if summary != "" {
		return summary
	}

	return ""
}

func (a *atom03Entry) entryDate() time.Time {
	dateText := ""
	for _, value := range []string{a.Issued, a.Modified, a.Created} {
		if value != "" {
			dateText = value
			break
		}
	}

	if dateText != "" {
		result, err := date.Parse(dateText)
		if err != nil {
			logger.Error("atom: %v", err)
			return time.Now()
		}

		return result
	}

	return time.Now()
}

func (a *atom03Entry) entryHash() string {
	for _, value := range []string{a.ID, a.Links.originalLink()} {
		if value != "" {
			return crypto.Hash(value)
		}
	}

	return ""
}

type atom03Text struct {
	Type     string `xml:"type,attr"`
	Mode     string `xml:"mode,attr"`
	CharData string `xml:",chardata"`
	InnerXML string `xml:",innerxml"`
}

func (a *atom03Text) String() string {
	content := ""

	switch {
	case a.Mode == "xml":
		content = a.InnerXML
	case a.Mode == "escaped":
		content = a.CharData
	case a.Mode == "base64":
		b, err := base64.StdEncoding.DecodeString(a.CharData)
		if err == nil {
			content = string(b)
		}
	default:
		content = a.CharData
	}

	if a.Type != "text/html" {
		content = html.EscapeString(content)
	}

	return strings.TrimSpace(content)
}
