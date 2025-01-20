// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Discord Webhooks documentation: https://discord.com/developers/docs/resources/webhook

package discord // import "miniflux.app/v2/internal/integration/discord"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
	"miniflux.app/v2/internal/version"
)

const defaultClientTimeout = 10 * time.Second
const discordMsgColor = 5793266

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

func (c *Client) SendDiscordMsg(feed *model.Feed, entries model.Entries) error {
	for _, entry := range entries {
		requestBody, err := json.Marshal(&discordMessage{
			Embeds: []discordEmbed{
				{
					Title: "RSS feed update from Miniflux",
					Color: discordMsgColor,
					Fields: []discordFields{
						{
							Name:  "Updated feed",
							Value: feed.Title,
						},
						{
							Name:  "Article link",
							Value: "[" + entry.Title + "]" + "(" + entry.URL + ")",
						},
						{
							Name:   "Author",
							Value:  entry.Author,
							Inline: true,
						},
						{
							Name:   "Source website",
							Value:  urllib.RootURL(feed.SiteURL),
							Inline: true,
						},
					},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("discord: unable to encode request body: %v", err)
		}

		request, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("discord: unable to create request: %v", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "Miniflux/"+version.Version)

		slog.Debug("Sending Discord notification",
			slog.String("webhookURL", c.webhookURL),
			slog.String("title", feed.Title),
			slog.String("entry_url", entry.URL),
		)

		httpClient := &http.Client{Timeout: defaultClientTimeout}
		response, err := httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("discord: unable to send request: %v", err)
		}
		response.Body.Close()

		if response.StatusCode >= 400 {
			return fmt.Errorf("discord: unable to send a notification: url=%s status=%d", c.webhookURL, response.StatusCode)
		}
	}

	return nil
}

type discordFields struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordEmbed struct {
	Title  string          `json:"title"`
	Color  int             `json:"color"`
	Fields []discordFields `json:"fields"`
}

type discordMessage struct {
	Embeds []discordEmbed `json:"embeds"`
}
