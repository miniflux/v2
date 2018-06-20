// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/miniflux/miniflux/model"
)

type feedIcon struct {
	ID       int64  `json:"id"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type entriesResponse struct {
	Total   int           `json:"total"`
	Entries model.Entries `json:"entries"`
}

type feedCreation struct {
	FeedURL    string `json:"feed_url"`
	CategoryID int64  `json:"category_id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Crawler    bool   `json:"crawler"`
}

type subscriptionDiscovery struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func decodeUserPayload(r io.ReadCloser) (*model.User, error) {
	var user model.User

	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&user); err != nil {
		return nil, fmt.Errorf("Unable to decode user JSON object: %v", err)
	}

	return &user, nil
}

func decodeURLPayload(r io.ReadCloser) (*subscriptionDiscovery, error) {
	defer r.Close()

	var s subscriptionDiscovery
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&s); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return &s, nil
}

func decodeEntryStatusPayload(r io.ReadCloser) ([]int64, string, error) {
	type payload struct {
		EntryIDs []int64 `json:"entry_ids"`
		Status   string  `json:"status"`
	}

	var p payload
	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&p); err != nil {
		return nil, "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	return p.EntryIDs, p.Status, nil
}

func decodeFeedCreationPayload(r io.ReadCloser) (*feedCreation, error) {
	defer r.Close()

	var fc feedCreation
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&fc); err != nil {
		return nil, fmt.Errorf("invalid JSON payload: %v", err)
	}

	return &fc, nil
}

func decodeFeedModificationPayload(r io.ReadCloser) (*model.Feed, error) {
	var feed model.Feed

	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("Unable to decode feed JSON object: %v", err)
	}

	return &feed, nil
}

func decodeCategoryPayload(r io.ReadCloser) (*model.Category, error) {
	var category model.Category

	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&category); err != nil {
		return nil, fmt.Errorf("Unable to decode category JSON object: %v", err)
	}

	return &category, nil
}
