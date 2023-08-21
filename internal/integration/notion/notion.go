// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second

type Client struct {
	apiToken string
	pageID   string
}

func NewClient(apiToken, pageID string) *Client {
	return &Client{apiToken, pageID}
}

func (c *Client) UpdateDocument(entryURL string, entryTitle string) error {
	if c.apiToken == "" || c.pageID == "" {
		return fmt.Errorf("notion: missing API token or page ID")
	}

	apiEndpoint := "https://api.notion.com/v1/blocks/" + c.pageID + "/children"
	requestBody, err := json.Marshal(&notionDocument{
		Children: []block{
			{
				Object: "block",
				Type:   "bookmark",
				Bookmark: bookmarkObject{
					Caption: []any{},
					URL:     entryURL,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("notion: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPatch, apiEndpoint, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("notion: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)
	request.Header.Set("Notion-Version", "2022-06-28")
	request.Header.Set("Authorization", "Bearer "+c.apiToken)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("notion: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("notion: unable to update document: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

type notionDocument struct {
	Children []block `json:"children"`
}

type block struct {
	Object   string         `json:"object"`
	Type     string         `json:"type"`
	Bookmark bookmarkObject `json:"bookmark"`
}

type bookmarkObject struct {
	Caption []any  `json:"caption"`
	URL     string `json:"url"`
}
