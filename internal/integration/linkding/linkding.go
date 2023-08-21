// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkding // import "miniflux.app/v2/internal/integration/linkding"

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
	tags    string
	unread  bool
}

func NewClient(baseURL, apiKey, tags string, unread bool) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, tags: tags, unread: unread}
}

func (c *Client) CreateBookmark(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("linkding: missing base URL or API key")
	}

	tagsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return fmt.Errorf(`linkding: invalid API endpoint: %v`, err)
	}

	requestBody, err := json.Marshal(&linkdingBookmark{
		Url:      entryURL,
		Title:    entryTitle,
		TagNames: strings.FieldsFunc(c.tags, tagsSplitFn),
		Unread:   c.unread,
	})

	if err != nil {
		return fmt.Errorf("linkding: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("linkding: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Authorization", "Token "+c.apiKey)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("linkding: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("linkding: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type linkdingBookmark struct {
	Url      string   `json:"url,omitempty"`
	Title    string   `json:"title,omitempty"`
	TagNames []string `json:"tag_names,omitempty"`
	Unread   bool     `json:"unread,omitempty"`
}
