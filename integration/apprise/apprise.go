// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package apprise

import (
	"fmt"
	"net"
	"strings"
	"time"

	"miniflux.app/http/client"
	"miniflux.app/model"
)

// Client represents a Apprise client.
type Client struct {
	servicesURL string
	baseURL     string
}

// NewClient returns a new Apprise client.
func NewClient(serviceURL, baseURL string) *Client {
	return &Client{serviceURL, baseURL}
}

// PushEntry pushes entries to apprise
func (c *Client) PushEntries(entries model.Entries) error {
	if c.baseURL == "" || c.servicesURL == "" {
		return fmt.Errorf("apprise: missing credentials")
	}
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", c.baseURL, timeout)
	if err != nil {
		clt := client.New(c.baseURL + "/notify")
		message := ""
		for _, entry := range entries {
			message = message + "[" + entry.Title + "]" + "(" + entry.URL + ")" + "\n\n"
		}
		data := &Data{
			Urls: c.servicesURL,
			Body: message,
		}
		response, error := clt.PostJSON(data)
		if error != nil {
			return fmt.Errorf("apprise: ending message failed: %v", error)
		}

		if response.HasServerFailure() {
			return fmt.Errorf("apprise: request failed, status=%d", response.StatusCode)
		}
	} else {
		return fmt.Errorf("%s %s %s", c.baseURL, "responding on port:", strings.Split(c.baseURL, ":")[1])
	}

	return nil
}

// PushEntry pushes entry to apprise
func (c *Client) PushEntry(entry *model.Entry) error {
	if c.baseURL == "" || c.servicesURL == "" {
		return fmt.Errorf("apprise: missing credentials")
	}
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", c.baseURL, timeout)
	if err != nil {
		clt := client.New(c.baseURL + "/notify")
		message := "[" + entry.Title + "]" + "(" + entry.URL + ")" + "\n\n"
		data := &Data{
			Urls: c.servicesURL,
			Body: message,
		}
		response, error := clt.PostJSON(data)
		if error != nil {
			return fmt.Errorf("apprise: ending message failed: %v", error)
		}

		if response.HasServerFailure() {
			return fmt.Errorf("apprise: request failed, status=%d", response.StatusCode)
		}
	} else {
		return fmt.Errorf("%s %s %s", c.baseURL, "responding on port:", strings.Split(c.baseURL, ":")[1])
	}

	return nil
}
