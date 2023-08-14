// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package espial // import "miniflux.app/v2/internal/integration/espial"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL string
	apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

func (c *Client) CreateLink(entryURL, entryTitle, espialTags string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("espial: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/add")
	if err != nil {
		return fmt.Errorf("espial: invalid API endpoint: %v", err)
	}

	requestBody, err := json.Marshal(&espialDocument{
		Title:  entryTitle,
		Url:    entryURL,
		ToRead: true,
		Tags:   espialTags,
	})

	if err != nil {
		return fmt.Errorf("espial: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("espial: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "ApiKey "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("espial: unable to send request: %v", err)
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
	Url    string `json:"url,omitempty"`
	ToRead bool   `json:"toread,omitempty"`
	Tags   string `json:"tags,omitempty"`
}
