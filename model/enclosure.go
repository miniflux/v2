// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"
import "strings"

// Enclosure represents an attachment.
type Enclosure struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	EntryID          int64  `json:"entry_id"`
	URL              string `json:"url"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
	MediaProgression int64  `json:"media_progression"`
}

// Html5MimeType will modify the actual MimeType to allow direct playback from HTML5 player for some kind of MimeType
func (e Enclosure) Html5MimeType() string {
	if strings.HasPrefix(e.MimeType, "video") {
		switch e.MimeType {
		// Solution from this stackoverflow discussion:
		// https://stackoverflow.com/questions/15277147/m4v-mimetype-video-mp4-or-video-m4v/66945470#66945470
		// tested at the time of this commit (06/2023) on latest Firefox & Vivaldi on this feed
		// https://www.florenceporcel.com/podcast/lfhdu.xml
		case "video/m4v":
			return "video/x-m4v"
		}
	}
	return e.MimeType
}

// EnclosureList represents a list of attachments.
type EnclosureList []*Enclosure
