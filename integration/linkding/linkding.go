// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package linkding // import "miniflux.app/integration/linkding"

import (
	"fmt"
	"net/url"
	"strings"

	"miniflux.app/http/client"
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
func (c *Client) AddEntry(title, url string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("linkding: missing credentials")
	}

	tagsSplitFn := func(c rune) bool {
		return c == ',' || c == ' '
	}

	doc := &Document{
		Url:      url,
		Title:    title,
		TagNames: strings.FieldsFunc(c.tags, tagsSplitFn),
		Unread:   c.unread,
	}

	apiURL, err := getAPIEndpoint(c.baseURL, "/api/bookmarks/")
	if err != nil {
		return err
	}

	clt := client.New(apiURL)
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

func getAPIEndpoint(baseURL, pathURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("linkding: invalid API endpoint: %v", err)
	}

	relative, err := url.Parse(pathURL)
	if err != nil {
		return "", fmt.Errorf("linkding: invalid API endpoint: %v", err)
	}

	u = u.ResolveReference(relative)
	return u.String(), nil
}
