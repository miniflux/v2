// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package json

import (
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/feed/date"
	"github.com/miniflux/miniflux2/reader/processor"
	"github.com/miniflux/miniflux2/reader/sanitizer"
	"log"
	"strings"
	"time"
)

type JsonFeed struct {
	Version string     `json:"version"`
	Title   string     `json:"title"`
	SiteURL string     `json:"home_page_url"`
	FeedURL string     `json:"feed_url"`
	Author  JsonAuthor `json:"author"`
	Items   []JsonItem `json:"items"`
}

type JsonAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type JsonItem struct {
	ID            string           `json:"id"`
	URL           string           `json:"url"`
	Title         string           `json:"title"`
	Summary       string           `json:"summary"`
	Text          string           `json:"content_text"`
	Html          string           `json:"content_html"`
	DatePublished string           `json:"date_published"`
	DateModified  string           `json:"date_modified"`
	Author        JsonAuthor       `json:"author"`
	Attachments   []JsonAttachment `json:"attachments"`
}

type JsonAttachment struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Title    string `json:"title"`
	Size     int    `json:"size_in_bytes"`
	Duration int    `json:"duration_in_seconds"`
}

func (j *JsonFeed) GetAuthor() string {
	return getAuthor(j.Author)
}

func (j *JsonFeed) Transform() *model.Feed {
	feed := new(model.Feed)
	feed.FeedURL = j.FeedURL
	feed.SiteURL = j.SiteURL
	feed.Title = sanitizer.StripTags(j.Title)

	if feed.Title == "" {
		feed.Title = feed.SiteURL
	}

	for _, item := range j.Items {
		entry := item.Transform()
		if entry.Author == "" {
			entry.Author = j.GetAuthor()
		}

		feed.Entries = append(feed.Entries, entry)
	}

	return feed
}

func (j *JsonItem) GetDate() time.Time {
	for _, value := range []string{j.DatePublished, j.DateModified} {
		if value != "" {
			d, err := date.Parse(value)
			if err != nil {
				log.Println(err)
				return time.Now()
			}

			return d
		}
	}

	return time.Now()
}

func (j *JsonItem) GetAuthor() string {
	return getAuthor(j.Author)
}

func (j *JsonItem) GetHash() string {
	for _, value := range []string{j.ID, j.URL, j.Text + j.Html + j.Summary} {
		if value != "" {
			return helper.Hash(value)
		}
	}

	return ""
}

func (j *JsonItem) GetTitle() string {
	for _, value := range []string{j.Title, j.Summary, j.Text, j.Html} {
		if value != "" {
			return truncate(value)
		}
	}

	return j.URL
}

func (j *JsonItem) GetContent() string {
	for _, value := range []string{j.Html, j.Text, j.Summary} {
		if value != "" {
			return value
		}
	}

	return ""
}

func (j *JsonItem) GetEnclosures() model.EnclosureList {
	enclosures := make(model.EnclosureList, 0)

	for _, attachment := range j.Attachments {
		enclosures = append(enclosures, &model.Enclosure{
			URL:      attachment.URL,
			MimeType: attachment.MimeType,
			Size:     attachment.Size,
		})
	}

	return enclosures
}

func (j *JsonItem) Transform() *model.Entry {
	entry := new(model.Entry)
	entry.URL = j.URL
	entry.Date = j.GetDate()
	entry.Author = sanitizer.StripTags(j.GetAuthor())
	entry.Hash = j.GetHash()
	entry.Content = processor.ItemContentProcessor(entry.URL, j.GetContent())
	entry.Title = sanitizer.StripTags(strings.Trim(j.GetTitle(), " \n\t"))
	entry.Enclosures = j.GetEnclosures()
	return entry
}

func getAuthor(author JsonAuthor) string {
	if author.Name != "" {
		return author.Name
	}

	return ""
}

func truncate(str string) string {
	max := 100
	if len(str) > max {
		return str[:max] + "..."
	}

	return str
}
