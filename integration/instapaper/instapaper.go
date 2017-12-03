// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package instapaper

import (
	"fmt"
	"net/url"

	"github.com/miniflux/miniflux2/http"
)

// Client represents an Instapaper client.
type Client struct {
	username string
	password string
}

// AddURL sends a link to Instapaper.
func (c *Client) AddURL(link, title string) error {
	values := url.Values{}
	values.Add("url", link)
	values.Add("title", title)

	apiURL := "https://www.instapaper.com/api/add?" + values.Encode()
	client := http.NewClientWithCredentials(apiURL, c.username, c.password)
	response, err := client.Get()
	if response.HasServerFailure() {
		return fmt.Errorf("unable to send bookmark to instapaper, status=%d", response.StatusCode)
	}

	return err
}

// NewClient returns a new Instapaper client.
func NewClient(username, password string) *Client {
	return &Client{username: username, password: password}
}
