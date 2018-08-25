// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// Enclosure represents an attachment.
type Enclosure struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	EntryID  int64  `json:"entry_id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// EnclosureList represents a list of attachments.
type EnclosureList []*Enclosure
