// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nunuxkeeper // import "miniflux.app/integration/nunuxkeeper"

import (
	"fmt"
	"net/url"
	"path"

	"miniflux.app/http/client"
)

// Document structure of a Nununx Keeper document
type Document struct {
	Title       string `json:"title,omitempty"`
	Origin      string `json:"origin,omitempty"`
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

// Client represents an Nunux Keeper client.
type Client struct {
	baseURL string
	apiKey  string
}

// NewClient returns a new Nunux Keeepr client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

// AddEntry sends an entry to Nunux Keeper.
func (c *Client) AddEntry(link, title, content string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("nunux-keeper: missing credentials")
	}

	doc := &Document{
		Title:       title,
		Origin:      link,
		Content:     content,
		ContentType: "text/html",
	}

	apiURL, err := getAPIEndpoint(c.baseURL, "/v2/documents")
	if err != nil {
		return err
	}

	clt := client.New(apiURL)
	clt.WithCredentials("api", c.apiKey)
	response, err := clt.PostJSON(doc)
	if err != nil {
		return fmt.Errorf("nunux-keeper: unable to send entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("nunux-keeper: unable to send entry, status=%d", response.StatusCode)
	}

	return nil
}

func getAPIEndpoint(baseURL, pathURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("nunux-keeper: invalid API endpoint: %v", err)
	}
	u.Path = path.Join(u.Path, pathURL)
	return u.String(), nil
}
