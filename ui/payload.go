// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/json"
	"fmt"
	"io"

	"miniflux.app/model"
)

func decodeEntryStatusPayload(r io.ReadCloser) (entryIDs []int64, status string, err error) {
	type payload struct {
		EntryIDs []int64 `json:"entry_ids"`
		Status   string  `json:"status"`
	}

	var p payload
	decoder := json.NewDecoder(r)
	defer r.Close()
	if err = decoder.Decode(&p); err != nil {
		return nil, "", fmt.Errorf("invalid JSON payload: %v", err)
	}

	if err := model.ValidateEntryStatus(p.Status); err != nil {
		return nil, "", err
	}

	return p.EntryIDs, p.Status, nil
}

func decodeCreateWebpushSubscriptionPayload(r io.ReadCloser) (*model.WebpushSubscription, error) {
	var webpushSubscription model.WebpushSubscription

	decoder := json.NewDecoder(r)
	defer r.Close()
	if err := decoder.Decode(&webpushSubscription); err != nil {
		return nil, fmt.Errorf("Unable to decode webpush creation JSON object: %v", err)
	}

	return &webpushSubscription, nil
}
