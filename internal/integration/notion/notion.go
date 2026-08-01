// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package notion

import (
	"errors"
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/http/client"
)

type Client struct {
	apiToken string
	pageID   string
}

func NewClient(apiToken, pageID string) *Client {
	return &Client{apiToken, pageID}
}

func (c *Client) UpdateDocument(entryURL string, entryTitle string) error {
	if c.apiToken == "" || c.pageID == "" {
		return errors.New("notion: missing API token or page ID")
	}

	apiEndpoint := "https://api.notion.com/v1/blocks/" + c.pageID + "/children"
	response, err := client.NewRequestBuilder(apiEndpoint).
		WithMethod(http.MethodPatch).
		WithJSON(&notionDocument{
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
		}).
		WithHeader("Notion-Version", "2022-06-28").
		WithHeader("Authorization", "Bearer "+c.apiToken).
		Do()
	if err != nil {
		return fmt.Errorf("notion: %w", err)
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
