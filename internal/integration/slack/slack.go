// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Slack Webhooks documentation: https://api.slack.com/messaging/webhooks

package slack // import "miniflux.app/v2/internal/integration/slack"

import (
	"fmt"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/urllib"
)

const slackMsgColor = "#5865F2"

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

func (c *Client) SendSlackMsg(feed *model.Feed, entries model.Entries) error {
	for _, entry := range entries {
		slog.Debug("Sending Slack notification",
			slog.String("webhookURL", c.webhookURL),
			slog.String("title", feed.Title),
			slog.String("entry_url", entry.URL),
		)

		response, err := client.NewRequestBuilder(c.webhookURL).
			WithMethod(http.MethodPost).
			WithJSON(&slackMessage{
				Attachments: []slackAttachments{
					{
						Title: "RSS feed update from Miniflux",
						Color: slackMsgColor,
						Fields: []slackFields{
							{
								Title: "Updated feed",
								Value: feed.Title,
							},
							{
								Title: "Article title",
								Value: entry.Title,
							},
							{
								Title: "Article link",
								Value: entry.URL,
							},
							{
								Title: "Author",
								Value: entry.Author,
								Short: true,
							},
							{
								Title: "Source website",
								Value: urllib.RootURL(feed.SiteURL),
								Short: true,
							},
						},
					},
				},
			}).
			Do()
		if err != nil {
			return fmt.Errorf("slack: %w", err)
		}
		response.Body.Close()

		if response.StatusCode >= 400 {
			return fmt.Errorf("slack: unable to send a notification: url=%s status=%d", c.webhookURL, response.StatusCode)
		}
	}

	return nil
}

type slackFields struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

type slackAttachments struct {
	Title  string        `json:"title"`
	Color  string        `json:"color"`
	Fields []slackFields `json:"fields"`
}

type slackMessage struct {
	Attachments []slackAttachments `json:"attachments"`
}
