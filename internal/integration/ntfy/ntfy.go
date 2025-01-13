// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ntfy // import "miniflux.app/v2/internal/integration/ntfy"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/version"
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultNtfyURL       = "https://ntfy.sh"
)

type Client struct {
	ntfyURL, ntfyTopic, ntfyApiToken, ntfyUsername, ntfyPassword, ntfyIconURL string
	ntfyInternalLinks                                                         bool
	ntfyPriority                                                              int
}

func NewClient(ntfyURL, ntfyTopic, ntfyApiToken, ntfyUsername, ntfyPassword, ntfyIconURL string, ntfyInternalLinks bool, ntfyPriority int) *Client {
	if ntfyURL == "" {
		ntfyURL = defaultNtfyURL
	}
	return &Client{ntfyURL, ntfyTopic, ntfyApiToken, ntfyUsername, ntfyPassword, ntfyIconURL, ntfyInternalLinks, ntfyPriority}
}

func (c *Client) SendMessages(feed *model.Feed, entries model.Entries) error {
	for _, entry := range entries {
		ntfyMessage := &ntfyMessage{
			Topic:    c.ntfyTopic,
			Message:  entry.Title,
			Title:    feed.Title,
			Priority: c.ntfyPriority,
			Click:    entry.URL,
		}

		if c.ntfyIconURL != "" {
			ntfyMessage.Icon = c.ntfyIconURL
		}

		if c.ntfyInternalLinks {
			url, err := url.Parse(config.Opts.BaseURL())
			if err != nil {
				slog.Error("Unable to parse base URL", slog.Any("error", err))
			} else {
				ntfyMessage.Click = fmt.Sprintf("%s%s%d", url, "/unread/entry/", entry.ID)
			}
		}

		slog.Debug("Sending Ntfy message",
			slog.String("url", c.ntfyURL),
			slog.String("topic", c.ntfyTopic),
			slog.Int("priority", ntfyMessage.Priority),
			slog.String("message", ntfyMessage.Message),
			slog.String("entry_url", ntfyMessage.Click),
		)

		if err := c.makeRequest(ntfyMessage); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) makeRequest(payload any) error {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ntfy: unable to encode request body: %v", err)
	}

	request, err := http.NewRequest(http.MethodPost, c.ntfyURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("ntfy: unable to create request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Miniflux/"+version.Version)

	// See https://docs.ntfy.sh/publish/#access-tokens
	if c.ntfyApiToken != "" {
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.ntfyApiToken))
	}

	// See https://docs.ntfy.sh/publish/#username-password
	if c.ntfyUsername != "" && c.ntfyPassword != "" {
		request.SetBasicAuth(c.ntfyUsername, c.ntfyPassword)
	}

	httpClient := &http.Client{Timeout: defaultClientTimeout}
	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("ntfy: unable to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return fmt.Errorf("ntfy: incorrect response status code %d for url %s", response.StatusCode, c.ntfyURL)
	}

	return nil
}

// See https://docs.ntfy.sh/publish/#publish-as-json
type ntfyMessage struct {
	Topic    string       `json:"topic"`
	Message  string       `json:"message"`
	Title    string       `json:"title"`
	Tags     []string     `json:"tags,omitempty"`
	Priority int          `json:"priority,omitempty"`
	Icon     string       `json:"icon,omitempty"` // https://docs.ntfy.sh/publish/#icons
	Click    string       `json:"click,omitempty"`
	Actions  []ntfyAction `json:"actions,omitempty"`
}

// See https://docs.ntfy.sh/publish/#action-buttons
type ntfyAction struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
}
