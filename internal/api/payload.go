// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api // import "miniflux.app/v2/internal/api"

import (
	"miniflux.app/v2/internal/model"
)

type feedIconResponse struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type entriesResponse struct {
	Total   int           `json:"total"`
	Entries model.Entries `json:"entries"`
}

type integrationsStatusResponse struct {
	HasIntegrations bool `json:"has_integrations"`
}

type entryIDResponse struct {
	ID int64 `json:"id"`
}

type entryContentResponse struct {
	Content     string `json:"content"`
	ReadingTime int    `json:"reading_time"`
}

type entryImportRequest struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Author      string   `json:"author"`
	CommentsURL string   `json:"comments_url"`
	PublishedAt int64    `json:"published_at"`
	Status      string   `json:"status"`
	Starred     bool     `json:"starred"`
	Tags        []string `json:"tags"`
	ExternalID  string   `json:"external_id"`
}

type feedCreationResponse struct {
	FeedID int64 `json:"feed_id"`
}

type importFeedsResponse struct {
	Message string `json:"message"`
}

type versionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Arch      string `json:"arch"`
	OS        string `json:"os"`
}
