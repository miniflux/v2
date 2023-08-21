// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nunuxkeeper // import "miniflux.app/v2/internal/integration/nunuxkeeper"

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

func (c *Client) AddEntry(entryURL, entryTitle, entryContent string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("nunux-keeper: missing base URL or API key")
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/v2/documents")
	if err != nil {
		return fmt.Errorf(`nunux-keeper: invalid API endpoint: %v`, err)
	}

	requestBody, err := json.Marshal(&nunuxKeeperDocument{
		Title:       entryTitle,
		Origin:      entryURL,
		Content:     entryContent,
		ContentType: "text/html",
	})
	if err != nil {
		return fmt.Errorf("notion: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("nunux-keeper: unable to create request: %v", err)
	}

	request.SetBasicAuth("api", c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("nunux-keeper: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("nunux-keeper: unable to create document: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type nunuxKeeperDocument struct {
	Title       string `json:"title,omitempty"`
	Origin      string `json:"origin,omitempty"`
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}
