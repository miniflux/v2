// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package readeck // import "miniflux.app/v2/internal/integration/readeck"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	baseURL string
	apiKey  string
	labels  string
}

func NewClient(baseURL, apiKey, labels string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, labels: labels}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("readeck: missing base URL or API key")
	}

	labelsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return fmt.Errorf(`readeck: invalid API endpoint: %v`, err)
	}

	requestBody, err := json.Marshal(&readeckBookmark{
		Url:    entryURL,
		Title:  entryTitle,
		Labels: strings.FieldsFunc(c.labels, labelsSplitFn),
	})

	if err != nil {
		return fmt.Errorf("readeck: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("readeck: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("readeck: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("readeck: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type readeckBookmark struct {
	Url    string   `json:"url"`
	Title  string   `json:"title"`
	Labels []string `json:"labels,omitempty"`
}
