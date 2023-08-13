// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkding // import "miniflux.app/v2/internal/integration/linkding"

import (
	"fmt"
	"strings"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/url"
)

// Document structure of a Linkding document
type Document struct {
	Url      string   `json:"url,omitempty"`
	Title    string   `json:"title,omitempty"`
	TagNames []string `json:"tag_names,omitempty"`
	Unread   bool     `json:"unread,omitempty"`
}

// Client represents an Linkding client.
type Client struct {
	baseURL string
	apiKey  string
	tags    string
	unread  bool
}

// NewClient returns a new Linkding client.
func NewClient(baseURL, apiKey, tags string, unread bool) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey, tags: tags, unread: unread}
}

// AddEntry sends an entry to Linkding.
func (c *Client) AddEntry(title, entryURL string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("linkding: missing credentials")
	}

	tagsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	doc := &Document{
		Url:      entryURL,
		Title:    title,
		TagNames: strings.FieldsFunc(c.tags, tagsSplitFn),
		Unread:   c.unread,
	}

	apiEndpoint, err := url.JoinBaseURLAndPath(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return fmt.Errorf(`linkding: invalid API endpoint: %v`, err)
	}

	clt := client.New(apiEndpoint)
	clt.WithAuthorization("Token " + c.apiKey)
	response, err := clt.PostJSON(doc)
	if err != nil {
		return fmt.Errorf("linkding: unable to send entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("linkding: unable to send entry, status=%d", response.StatusCode)
	}

	return nil
}
