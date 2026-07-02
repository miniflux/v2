// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkding // import "miniflux.app/v2/internal/integration/linkding"

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/urllib"
)

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
		return errors.New("linkding: missing base URL or API key")
	}

	tagsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return fmt.Errorf(`linkding: invalid API endpoint: %v`, err)
	}

	response, err := client.NewRequestBuilder(apiEndpoint).
		WithMethod(http.MethodPost).
		WithJSON(&linkdingBookmark{
			URL:      entryURL,
			Title:    entryTitle,
			TagNames: strings.FieldsFunc(c.tags, tagsSplitFn),
			Unread:   c.unread,
		}).
		WithHeader("Authorization", "Token "+c.apiKey).
		Do()
	if err != nil {
		return fmt.Errorf("linkding: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("linkding: unable to create bookmark: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type linkdingBookmark struct {
	URL      string   `json:"url,omitempty"`
	Title    string   `json:"title,omitempty"`
	TagNames []string `json:"tag_names,omitempty"`
	Unread   bool     `json:"unread,omitempty"`
}
