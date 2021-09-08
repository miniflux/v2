// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package pinboard // import "miniflux.app/integration/pinboard"

import (
	"fmt"
	"net/url"

	"miniflux.app/http/client"
)

// Client represents a Pinboard client.
type Client struct {
	authToken string
}

// NewClient returns a new Pinboard client.
func NewClient(authToken string) *Client {
	return &Client{authToken: authToken}
}

// AddBookmark sends a link to Pinboard.
func (c *Client) AddBookmark(link, title, tags string, markAsUnread bool) error {
	if c.authToken == "" {
		return fmt.Errorf("pinboard: missing credentials")
	}

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

	clt := client.New("https://api.pinboard.in/v1/posts/add?" + values.Encode())
	response, err := clt.Get()
	if err != nil {
		return fmt.Errorf("pinboard: unable to send bookmark: %v", err)
	}

	if response.HasServerFailure() {
		return fmt.Errorf("pinboard: unable to send bookmark, status=%d", response.StatusCode)
	}

	return nil
}
