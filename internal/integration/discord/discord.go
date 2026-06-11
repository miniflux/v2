// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Discord Webhooks documentation: https://discord.com/developers/docs/resources/webhook

package discord // import "miniflux.app/v2/internal/integration/discord"

import (
	"fmt"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

const discordMsgColor = 5793266

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

func (c *Client) SendDiscordMsg(feed *model.Feed, entries model.Entries) error {
	for _, entry := range entries {
		slog.Debug("Sending Discord notification",
			slog.String("webhookURL", c.webhookURL),
			slog.String("title", feed.Title),
			slog.String("entry_url", entry.URL),
		)

		response, err := client.NewRequestBuilder(c.webhookURL).
			WithMethod(http.MethodPost).
			WithJSON(&discordMessage{
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
			}).
			Do()
		if err != nil {
			return fmt.Errorf("discord: %w", err)
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
