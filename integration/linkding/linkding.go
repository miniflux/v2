// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package linkding // import "miniflux.app/integration/linkding"

import (
	"fmt"
	"net/url"

	"miniflux.app/http/client"
)

// Document structure of a Linkding document
type Document struct {
	Url   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

// Client represents an Linkding client.
type Client struct {
	baseURL string
	apiKey  string
}

// NewClient returns a new Linkding client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{baseURL: baseURL, apiKey: apiKey}
}

// AddEntry sends an entry to Linkding.
func (c *Client) AddEntry(title, url string) error {
	if c.baseURL == "" || c.apiKey == "" {
		return fmt.Errorf("linkding: missing credentials")
	}

	doc := &Document{
		Url:   url,
		Title: title,
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
