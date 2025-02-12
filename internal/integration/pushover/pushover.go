// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pushover // import "miniflux.app/v2/internal/integration/pushover"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/version"
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultPushoverURL   = "https://api.pushover.net"
)

type Client struct {
	prefix string

	token  string
	user   string
	device string

	priority int
}

type Message struct {
	Token string `json:"token"`
	User  string `json:"user"`

	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`

	URL      string `json:"url"`
	URLTitle string `json:"url_title"`
	Device   string `json:"device,omitempty"`
}

type ErrorResponse struct {
	User    string   `json:"user"`
	Errors  []string `json:"errors"`
	Status  int      `json:"status"`
	Request string   `json:"request"`
}

func New(user, token string, priority int, device, urlPrefix string) *Client {
	if urlPrefix == "" {
		urlPrefix = defaultPushoverURL
	}
	if priority < -2 {
		priority = -2
	}
	if priority > 2 {
		priority = 2
	}

	return &Client{
		user:     user,
		token:    token,
		device:   device,
		prefix:   urlPrefix,
		priority: priority,
	}
}

func (c *Client) SendMessages(feed *model.Feed, entries model.Entries) error {
	if c.token == "" || c.user == "" {
		return fmt.Errorf("pushover token and user are required")
	}
	for _, entry := range entries {
		msg := &Message{
			User:   c.user,
			Token:  c.token,
			Device: c.device,

			Message:  entry.Title,
			Title:    feed.Title,
			Priority: c.priority,
			URL:      entry.URL,
		}

		slog.Debug("Sending Pushover message",
			slog.Int("priority", msg.Priority),
			slog.String("message", msg.Message),
			slog.String("entry_url", msg.URL),
		)

		if err := c.makeRequest(msg); err != nil {
			return fmt.Errorf("c.makeRequest: %w", err)
		}
	}

	return nil
}

func (c *Client) makeRequest(payload *Message) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	url := c.prefix + "/1/messages.json"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http.NewRequest: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Miniflux/"+version.Version)

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		errorMessage := resp.Status

		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			if len(errResp.Errors) > 0 {
				errorMessage = strings.Join(errResp.Errors, ",")
			}
		}

		return fmt.Errorf("pushover API error (%d): %s", resp.StatusCode, errorMessage)
	}

	return nil
}
