// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Readwise Reader API documentation: https://readwise.io/reader_api

package readwise // import "miniflux.app/v2/internal/integration/readwise"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/version"
)

const (
	readwiseApiEndpoint  = "https://readwise.io/api/v3/save/"
	defaultClientTimeout = 10 * time.Second
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) CreateDocument(entryURL string) error {
	if c.apiKey == "" {
		return fmt.Errorf("readwise: missing API key")
	}

	requestBody, err := json.Marshal(&readwiseDocument{
		URL: entryURL,
	})

	if err != nil {
		return fmt.Errorf("readwise: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, readwiseApiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("readwise: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Token "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("readwise: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("readwise: unable to create document: url=%s status=%d", readwiseApiEndpoint, response.StatusCode)
	}

	return nil
}

type readwiseDocument struct {
	URL string `json:"url"`
}
