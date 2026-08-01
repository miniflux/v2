// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package espial // import "miniflux.app/v2/internal/integration/espial"

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/urllib"
)

type Client struct {
	baseURL string
	apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) CreateLink(entryURL, entryTitle, espialTags string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return errors.New("espial: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/add")
	if err != nil {
		return fmt.Errorf("espial: invalid API endpoint: %v", err)
	}

	response, err := client.NewRequestBuilder(apiEndpoint).
		WithMethod(http.MethodPost).
		WithJSON(&espialDocument{
			Title:  entryTitle,
			URL:    entryURL,
			ToRead: true,
			Tags:   espialTags,
		}).
		WithHeader("Authorization", "ApiKey "+c.apiKey).
		Do()
	if err != nil {
		return fmt.Errorf("espial: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		responseBody := new(bytes.Buffer)
		responseBody.ReadFrom(response.Body)

		return fmt.Errorf("espial: unable to create link: url=%s status=%d body=%s", apiEndpoint, response.StatusCode, responseBody.String())
	}

	return nil
}

type espialDocument struct {
	Title  string `json:"title,omitempty"`
	URL    string `json:"url,omitempty"`
	ToRead bool   `json:"toread,omitempty"`
	Tags   string `json:"tags,omitempty"`
}
