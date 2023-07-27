// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Readwise Reader API documentation: https://readwise.io/reader_api

package readwise // import "miniflux.app/integration/readwise"

import (
	"fmt"
	"net/url"

	"miniflux.app/http/client"
)

// Document structure of a Readwise Reader document
// This initial version accepts only the one required field, the URL
type Document struct {
	Url string `json:"url"`
}

// Client represents a Readwise Reader client.
type Client struct {
	apiKey string
}

// NewClient returns a new Readwise Reader client.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

// AddEntry sends an entry to Readwise Reader.
func (c *Client) AddEntry(link string) error {
	if c.apiKey == "" {
		return fmt.Errorf("readwise: missing credentials")
	}

	doc := &Document{
		Url: link,
	}

	apiURL, err := getAPIEndpoint("https://readwise.io/api/v3/save/")
	if err != nil {
		return err
	}

	clt := client.New(apiURL)
	clt.WithAuthorization("Token " + c.apiKey)
	response, err := clt.PostJSON(doc)
	if err != nil {
		return fmt.Errorf("readwise: unable to send entry: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("readwise: unable to send entry, status=%d", response.StatusCode)
	}

	return nil
}

func getAPIEndpoint(pathURL string) (string, error) {
	u, err := url.Parse(pathURL)
	if err != nil {
		return "", fmt.Errorf("readwise: invalid API endpoint: %v", err)
	}
	return u.String(), nil
}
