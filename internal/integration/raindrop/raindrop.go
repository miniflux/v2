// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package raindrop // import "miniflux.app/v2/internal/integration/raindrop"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	token        string
	collectionID string
	tags         []string
}

func NewClient(token, collectionID, tags string) *Client {
	return &Client{token: token, collectionID: collectionID, tags: strings.Split(tags, ",")}
}

// https://developer.raindrop.io/v1/raindrops/single#create-raindrop
func (c *Client) CreateRaindrop(entryURL, entryTitle string) error {
	if c.token == "" {
		return fmt.Errorf("raindrop: missing token")
	}

	var request *http.Request
	requestBodyJson, err := json.Marshal(&raindrop{
		Link:       entryURL,
		Title:      entryTitle,
		Collection: collection{Id: c.collectionID},
		Tags:       c.tags,
	})
	if err != nil {
		return fmt.Errorf("raindrop: unable to encode request body: %v", err)
	}

	request, err = http.NewRequest(http.MethodPost, "https://api.raindrop.io/rest/v1/raindrop", bytes.NewReader(requestBodyJson))
	if err != nil {
		return fmt.Errorf("raindrop: unable to create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.token)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("raindrop: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("raindrop: unable to create bookmark: status=%d", response.StatusCode)
	}

	return nil
}

type raindrop struct {
	Link       string     `json:"link"`
	Title      string     `json:"title"`
	Collection collection `json:"collection,omitempty"`
	Tags       []string   `json:"tags"`
}

type collection struct {
	Id string `json:"$id"`
}
