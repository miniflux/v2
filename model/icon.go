// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

// Icon represents a website icon (favicon)
type Icon struct {
	ID       int64  `json:"id"`
	Hash     string `json:"hash"`
	MimeType string `json:"mime_type"`
	Content  []byte `json:"content"`
}

// FeedIcon is a jonction table between feeds and icons
type FeedIcon struct {
	FeedID int64 `json:"feed_id"`
	IconID int64 `json:"icon_id"`
}
