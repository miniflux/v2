// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package payload

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/miniflux/miniflux/model"
)

// EntriesResponse represents the response sent when fetching entries.
type EntriesResponse struct {
	Total   int           `json:"total"`
	Entries model.Entries `json:"entries"`
}

// DecodeUserPayload unserialize JSON user object.
func DecodeUserPayload(data io.Reader) (*model.User, error) {
	var user model.User

	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("Unable to decode user JSON object: %v", err)
	}

	return &user, nil
}

// DecodeURLPayload unserialize JSON subscription object.
func DecodeURLPayload(data io.Reader) (string, error) {
	type payload struct {
		URL string `json:"url"`
	}

	var p payload
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&p); err != nil {
		return "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	return p.URL, nil
}

// DecodeEntryStatusPayload unserialize JSON entry statuses object.
func DecodeEntryStatusPayload(data io.Reader) ([]int64, string, error) {
	type payload struct {
		EntryIDs []int64 `json:"entry_ids"`
		Status   string  `json:"status"`
	}

	var p payload
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&p); err != nil {
		return nil, "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	return p.EntryIDs, p.Status, nil
}

// DecodeFeedCreationPayload unserialize JSON feed creation object.
func DecodeFeedCreationPayload(data io.Reader) (string, int64, error) {
	type payload struct {
		FeedURL    string `json:"feed_url"`
		CategoryID int64  `json:"category_id"`
	}

	var p payload
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&p); err != nil {
		return "", 0, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return p.FeedURL, p.CategoryID, nil
}

// DecodeFeedModificationPayload unserialize JSON feed object.
func DecodeFeedModificationPayload(data io.Reader) (*model.Feed, error) {
	var feed model.Feed

	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("Unable to decode feed JSON object: %v", err)
	}

	return &feed, nil
}

// DecodeCategoryPayload unserialize JSON category object.
func DecodeCategoryPayload(data io.Reader) (*model.Category, error) {
	var category model.Category

	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&category); err != nil {
		return nil, fmt.Errorf("Unable to decode category JSON object: %v", err)
	}

	return &category, nil
}
