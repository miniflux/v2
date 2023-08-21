// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/model"

import (
	"encoding/base64"
	"fmt"
)

// Icon represents a website icon (favicon)
type Icon struct {
	ID       int64  `json:"id"`
	Hash     string `json:"hash"`
	MimeType string `json:"mime_type"`
	Content  []byte `json:"content"`
}

// DataURL returns the data URL of the icon.
func (i *Icon) DataURL() string {
	return fmt.Sprintf("%s;base64,%s", i.MimeType, base64.StdEncoding.EncodeToString(i.Content))
}

// Icons represents a list of icons.
type Icons []*Icon

// FeedIcon is a junction table between feeds and icons.
type FeedIcon struct {
	FeedID int64 `json:"feed_id"`
	IconID int64 `json:"icon_id"`
}
