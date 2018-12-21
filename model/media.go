// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import (
	"encoding/base64"
	"fmt"
	"time"
)

// Media represents a entry media cache
type Media struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	URLHash   string    `json:"url_hash"`
	MimeType  string    `json:"mime_type"`
	Content   []byte    `json:"content"`
	Size      int       `json:"size"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// DataURL returns the data URL of the media cache.
func (i *Media) DataURL() string {
	return fmt.Sprintf("%s;base64,%s", i.MimeType, base64.StdEncoding.EncodeToString(i.Content))
}

// Medias represents a list of media cache.
type Medias []*Media

// EntryMedias represents media caches of an entry.
type EntryMedias struct {
	entry  *Entry
	medias Medias
}

// EntryMedia is a jonction table between entries and media cache
type EntryMedia struct {
	EntryID  int64 `json:"entry_id"`
	MediaID  int64 `json:"media_id"`
	UseCache bool  `json:"use_cache"`
}
