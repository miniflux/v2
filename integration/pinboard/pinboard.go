// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package pinboard

import (
	"fmt"
	"net/url"

	"github.com/miniflux/miniflux2/http"
)

// Client represents a Pinboard client.
type Client struct {
	authToken string
}

// AddBookmark sends a link to Pinboard.
func (c *Client) AddBookmark(link, title, tags string, markAsUnread bool) error {
	toRead := "no"
	if markAsUnread {
		toRead = "yes"
	}

	values := url.Values{}
	values.Add("auth_token", c.authToken)
	values.Add("url", link)
	values.Add("description", title)
	values.Add("tags", tags)
	values.Add("toread", toRead)

	client := http.NewClient("https://api.pinboard.in/v1/posts/add?" + values.Encode())
	response, err := client.Get()
	if response.HasServerFailure() {
		return fmt.Errorf("unable to send bookmark to pinboard, status=%d", response.StatusCode)
	}

	return err
}

// NewClient returns a new Pinboard client.
func NewClient(authToken string) *Client {
	return &Client{authToken: authToken}
}
