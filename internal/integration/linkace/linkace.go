// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkace // import "miniflux.app/v2/internal/integration/linkace"

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/urllib"
)

type Client struct {
	baseURL       string
	apiKey        string
	tags          string
	private       bool
	checkDisabled bool
}

func NewClient(baseURL, apiKey, tags string, private bool, checkDisabled bool) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, tags: tags, private: private, checkDisabled: checkDisabled}
}

func (c *Client) AddURL(entryURL, entryTitle string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return errors.New("linkace: missing base URL or API key")
	}

	tagsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	apiEndpoint, err := urllib.JoinBaseURLAndPath(c.baseURL, "/api/v2/links")
	if err != nil {
		return fmt.Errorf("linkace: invalid API endpoint: %v", err)
	}
	response, err := client.NewRequestBuilder(apiEndpoint).
		WithMethod(http.MethodPost).
		WithJSON(&createItemRequest{
			URL:           entryURL,
			Title:         entryTitle,
			Tags:          strings.FieldsFunc(c.tags, tagsSplitFn),
			Private:       c.private,
			CheckDisabled: c.checkDisabled,
		}).
		WithHeader("Accept", "application/json").
		WithHeader("Authorization", "Bearer "+c.apiKey).
		Do()
	if err != nil {
		return fmt.Errorf("linkace: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("linkace: unable to create item: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type createItemRequest struct {
	Title         string   `json:"title,omitempty"`
	URL           string   `json:"url"`
	Tags          []string `json:"tags,omitempty"`
	Private       bool     `json:"is_private,omitempty"`
	CheckDisabled bool     `json:"check_disabled,omitempty"`
}
