// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package espial // import "miniflux.app/integration/espial"

import (
	"fmt"
	"net/url"
	"path"

	"miniflux.app/http/client"
)

// Document structure of an Espial document
type Document struct {
	Title  string `json:"title,omitempty"`
	Url    string `json:"url,omitempty"`
	ToRead bool   `json:"toread,omitempty"`
	Tags   string `json:"tags,omitempty"`
}

// Client represents an Espial client.
type Client struct {
	baseURL string
	apiKey  string
}

// NewClient returns a new Espial client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

// AddEntry sends an entry to Espial.
func (c *Client) AddEntry(link, title, content, tags string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("espial: missing credentials")
	}

	doc := &Document{
		Title:  title,
		Url:    link,
		ToRead: true,
		Tags:   tags,
	}

	apiURL, err := getAPIEndpoint(c.baseURL, "/api/add")
	if err != nil {
		return err
	}

	clt := client.New(apiURL)
	clt.WithAuthorization("ApiKey " + c.apiKey)
	response, err := clt.PostJSON(doc)
	if err != nil {
		return fmt.Errorf("espial: unable to send entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("espial: unable to send entry, status=%d", response.StatusCode)
	}

	return nil
}

func getAPIEndpoint(baseURL, pathURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("espial: invalid API endpoint: %v", err)
	}
	u.Path = path.Join(u.Path, pathURL)
	return u.String(), nil
}
