// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package gotify // import "miniflux.app/v2/internal/integration/gotify"

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

type Client struct {
	baseURL  string
	token    string
	priority int
}

func NewClient(baseURL, token string, priority int) *Client {
	if priority < 0 {
		priority = 0
	}

	return &Client{
		baseURL:  baseURL,
		token:    token,
		priority: priority,
	}
}

// Payload represents a Gotify push notification message payload.
type Payload struct {
	Title    string         `json:"title,omitempty"`
	Message  string         `json:"message"`
	Priority int            `json:"priority,omitempty"`
	Extras   map[string]any `json:"extras,omitempty"`
}

// Validate validates the Gotify message payload according to the Gotify schema.
func (p *Payload) Validate() error {
	if p.Message == "" {
		return errors.New("gotify: missing message")
	}

	if p.Priority < 0 {
		return errors.New("gotify: priority must be non-negative")
	}

	return nil
}

func (c *Client) SendMessage(title, message string, priority int, extras map[string]any) error {
	if c.baseURL == "" || c.token == "" {
		return errors.New("gotify: missing base URL or token")
	}

	msg := &Payload{
		Title:    title,
		Message:  message,
		Priority: priority,
		Extras:   extras,
	}

	if err := msg.Validate(); err != nil {
		return err
	}

	apiEndpoint := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(apiEndpoint, "/message") {
		var err error
		apiEndpoint, err = urllib.JoinBaseURLAndPath(c.baseURL, "/message")
		if err != nil {
			return fmt.Errorf("gotify: invalid API endpoint: %w", err)
		}
	}

	response, err := client.NewRequestBuilder(apiEndpoint).
		WithMethod(http.MethodPost).
		WithHeader("X-Gotify-Key", c.token).
		WithJSON(msg).
		Do()
	if err != nil {
		return fmt.Errorf("gotify: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("gotify: unable to send message: url=%s status=%d", apiEndpoint, response.StatusCode)
	}

	return nil
}

func (c *Client) SendMessages(feed *model.Feed, entries model.Entries) error {
	if c.baseURL == "" || c.token == "" {
		return errors.New("gotify: missing base URL or token")
	}

	var title string
	if feed != nil {
		title = feed.Title
	}

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		message := entry.Title
		if message == "" {
			message = entry.URL
		}

		if message == "" {
			continue
		}

		slog.Debug("Sending Gotify notification",
			slog.String("url", c.baseURL),
			slog.String("title", title),
			slog.String("entry_title", entry.Title),
			slog.String("entry_url", entry.URL),
		)

		extras := make(map[string]any)
		if entry.URL != "" {
			extras["client::notification"] = map[string]any{
				"click": map[string]any{
					"url": entry.URL,
				},
			}
		}

		if err := c.SendMessage(title, message, c.priority, extras); err != nil {
			return err
		}
	}

	return nil
}
