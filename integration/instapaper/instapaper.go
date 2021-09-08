// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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
