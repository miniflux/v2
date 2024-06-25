// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinboard // import "miniflux.app/v2/internal/integration/pinboard"

import (
	"encoding/xml"
	"net/url"
	"strings"
	"time"
)

// Post a Pinboard bookmark.  "inspiration" from https://github.com/drags/pinboard/blob/master/posts.go#L32-L42
type Post struct {
	XMLName     xml.Name  `xml:"post"`
	Url         string    `xml:"href,attr"`
	Description string    `xml:"description,attr"`
	Tags        string    `xml:"tag,attr"`
	Extended    string    `xml:"extended,attr"`
	Date        time.Time `xml:"time,attr"`
	Shared      string    `xml:"shared,attr"`
	Toread      string    `xml:"toread,attr"`
}

// Posts A result of a Pinboard API call
type posts struct {
	XMLName xml.Name `xml:"posts"`
	Posts   []Post   `xml:"post"`
}

func NewPost(url string, description string) *Post {
	return &Post{
		Url:         url,
		Description: description,
		Date:        time.Now(),
		Toread:      "no",
	}
}

func (p *Post) addTag(tag string) {
	if !strings.Contains(p.Tags, tag) {
		p.Tags += " " + tag
	}
}

func (p *Post) SetToread() {
	p.Toread = "yes"
}

func (p *Post) AddValues(values url.Values) {
	values.Add("url", p.Url)
	values.Add("description", p.Description)
	values.Add("tags", p.Tags)
	if p.Toread != "" {
		values.Add("toread", p.Toread)
	}
	if p.Shared != "" {
		values.Add("shared", p.Shared)
	}
	values.Add("dt", p.Date.Format(time.RFC3339))
	values.Add("extended", p.Extended)
}
