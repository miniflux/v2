// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package pocket

import (
	"fmt"

	"github.com/miniflux/miniflux/http/client"
)

// Client represents a Pocket client.
type Client struct {
	accessToken string
	consumerKey string
}

// Parameters for a Pocket add call.
type Parameters struct {
	AccessToken string `json:"access_token"`
	ConsumerKey string `json:"consumer_key"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
}

// AddURL sends a single link to Pocket.
func (c *Client) AddURL(link, title string) error {
	if c.consumerKey == "" || c.accessToken == "" {
		return fmt.Errorf("pocket: missing credentials")
	}

	parameters := &Parameters{
		AccessToken: c.accessToken,
		ConsumerKey: c.consumerKey,
		Title:       title,
		URL:         link,
	}

	clt := client.New("https://getpocket.com/v3/add")
	response, err := clt.PostJSON(parameters)
	if response.HasServerFailure() {
		return fmt.Errorf("pocket: unable to send url, status=%d", response.StatusCode)
	}

	return err
}

// NewClient returns a new Pocket client.
func NewClient(accessToken, consumerKey string) *Client {
	return &Client{accessToken: accessToken, consumerKey: consumerKey}
}
