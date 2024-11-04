// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"strings"

	"github.com/gorilla/mux"
	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/urllib"
)

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

type EnclosureUpdateRequest struct {
	MediaProgression int64 `json:"media_progression"`
}

// Html5MimeType will modify the actual MimeType to allow direct playback from HTML5 player for some kind of MimeType
func (e Enclosure) Html5MimeType() string {
	if e.MimeType == "video/m4v" {
		return "video/x-m4v"
	}
	return e.MimeType
}

func (e *Enclosure) IsAudio() bool {
	return strings.HasPrefix(strings.ToLower(e.MimeType), "audio/")
}

func (e *Enclosure) IsVideo() bool {
	return strings.HasPrefix(strings.ToLower(e.MimeType), "video/")
}

func (e *Enclosure) IsImage() bool {
	mimeType := strings.ToLower(e.MimeType)
	mediaURL := strings.ToLower(e.URL)
	return strings.HasPrefix(mimeType, "image/") || strings.HasSuffix(mediaURL, ".jpg") || strings.HasSuffix(mediaURL, ".jpeg") || strings.HasSuffix(mediaURL, ".png") || strings.HasSuffix(mediaURL, ".gif")
}

// EnclosureList represents a list of attachments.
type EnclosureList []*Enclosure

// FindMediaPlayerEnclosure returns the first enclosure that can be played by a media player.
func (el EnclosureList) FindMediaPlayerEnclosure() *Enclosure {
	for _, enclosure := range el {
		if enclosure.URL != "" && strings.Contains(enclosure.MimeType, "audio/") || strings.Contains(enclosure.MimeType, "video/") {
			return enclosure
		}
	}

	return nil
}

func (el EnclosureList) ContainsAudioOrVideo() bool {
	for _, enclosure := range el {
		if strings.Contains(enclosure.MimeType, "audio/") || strings.Contains(enclosure.MimeType, "video/") {
			return true
		}
	}
	return false
}

func (el EnclosureList) ProxifyEnclosureURL(router *mux.Router) {
	proxyOption := config.Opts.MediaProxyMode()

	if proxyOption == "all" || proxyOption != "none" {
		for i := range el {
			if urllib.IsHTTPS(el[i].URL) {
				for _, mediaType := range config.Opts.MediaProxyResourceTypes() {
					if strings.HasPrefix(el[i].MimeType, mediaType+"/") {
						el[i].URL = mediaproxy.ProxifyAbsoluteURL(router, el[i].URL)
						break
					}
				}
			}
		}
	}
}

func (e *Enclosure) ProxifyEnclosureURL(router *mux.Router) {
	proxyOption := config.Opts.MediaProxyMode()

	if proxyOption == "all" || proxyOption != "none" && !urllib.IsHTTPS(e.URL) {
		for _, mediaType := range config.Opts.MediaProxyResourceTypes() {
			if strings.HasPrefix(e.MimeType, mediaType+"/") {
				e.URL = mediaproxy.ProxifyAbsoluteURL(router, e.URL)
				break
			}
		}
	}
}
