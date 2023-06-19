// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package instapaper // import "miniflux.app/integration/instapaper"

import (
	"fmt"
	"net/url"

	"miniflux.app/http/client"
)

// Client represents an Instapaper client.
type Client struct {
	username string
	password string
}

// NewClient returns a new Instapaper client.
func NewClient(username, password string) *Client {
	return &Client{username: username, password: password}
}

// AddURL sends a link to Instapaper.
func (c *Client) AddURL(link, title string) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("instapaper: missing credentials")
	}

	values := url.Values{}
	values.Add("url", link)
	values.Add("title", title)

	apiURL := "https://www.instapaper.com/api/add?" + values.Encode()
	clt := client.New(apiURL)
	clt.WithCredentials(c.username, c.password)
	response, err := clt.Get()
	if err != nil {
		return fmt.Errorf("instapaper: unable to send url: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("instapaper: unable to send url, status=%d", response.StatusCode)
	}

	return nil
}
