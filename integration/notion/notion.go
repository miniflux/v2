// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package notion

import (
	"fmt"

	"miniflux.app/http/client"
)

// Client represents a Notion client.
type Client struct {
	token  string
	pageID string
}

// NewClient returns a new Notion client.
func NewClient(token, pageID string) *Client {
	return &Client{token, pageID}
}

func (c *Client) AddEntry(entryURL string, entryTitle string) error {
	if c.token == "" || c.pageID == "" {
		return fmt.Errorf("notion: missing credentials")
	}
	clt := client.New("https://api.notion.com/v1/blocks/" + c.pageID + "/children")
	block := &Data{
		Children: []Block{
			{
				Object: "block",
				Type:   "bookmark",
				Bookmark: Bookmark{
					Caption: []interface{}{},
					URL:     entryURL,
				},
			},
		},
	}
	clt.WithAuthorization("Bearer " + c.token)
	customHeaders := map[string]string{
		"Notion-Version": "2022-06-28",
	}
	clt.WithCustomHeaders(customHeaders)
	response, error := clt.PatchJSON(block)
	if error != nil {
		return fmt.Errorf("notion: unable to patch entry: %v", error)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("notion: request failed, status=%d", response.StatusCode)
	}
	return nil
}
