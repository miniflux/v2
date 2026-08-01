// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Readwise Reader API documentation: https://readwise.io/reader_api

package readwise // import "miniflux.app/v2/internal/integration/readwise"

import (
	"errors"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/http/client"
)

const readwiseApiEndpoint = "https://readwise.io/api/v3/save/"

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) CreateDocument(entryURL string) error {
	if c.apiKey == "" {
		return errors.New("readwise: missing API key")
	}

	response, err := client.NewRequestBuilder(readwiseApiEndpoint).
		WithMethod(http.MethodPost).
		WithJSON(&readwiseDocument{
			URL: entryURL,
		}).
		WithHeader("Authorization", "Token "+c.apiKey).
		Do()
	if err != nil {
		return fmt.Errorf("readwise: %w", err)
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
