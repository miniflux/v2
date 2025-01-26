// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Slack Webhooks documentation: https://api.slack.com/messaging/webhooks

package slack // import "miniflux.app/v2/internal/integration/slack"

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
const slackMsgColor = "#5865F2"

type Client struct {
	webhookURL string
}

func NewClient(webhookURL string) *Client {
	return &Client{webhookURL: webhookURL}
}

func (c *Client) SendSlackMsg(feed *model.Feed, entries model.Entries) error {
	for _, entry := range entries {
		requestBody, err := json.Marshal(&slackMessage{
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
		})
		if err != nil {
			return fmt.Errorf("slack: unable to encode request body: %v", err)
		}

		request, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(requestBody))
		if err != nil {
			return fmt.Errorf("slack: unable to create request: %v", err)
		}

		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "Miniflux/"+version.Version)

		slog.Debug("Sending Slack notification",
			slog.String("webhookURL", c.webhookURL),
			slog.String("title", feed.Title),
			slog.String("entry_url", entry.URL),
		)

		httpClient := &http.Client{Timeout: defaultClientTimeout}
		response, err := httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("slack: unable to send request: %v", err)
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
