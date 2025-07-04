// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package archiveorg

import (
	"log/slog"
	"net/http"
	"net/url"
)

// See https://docs.google.com/document/d/1Nsv52MvSjbLb2PCpHlat0gkzw0EvtSgpKHu4mk0MnrA/edit?tab=t.0
const options = "delay_wb_availability=1&if_not_archived_within=5d"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) SendURL(entryURL, title string) error {
	// We're using a goroutine here as submissions to archive.org might take a long time,
	// and thus trigger a timeout on miniflux' side.
	go func(entryURL string) {
		slog.Debug("Sending entry to archive.org",
			slog.String("title", title))

		res, err := http.Get("https://web.archive.org/save/" + url.QueryEscape(entryURL) + "?" + options)
		if err != nil {
			slog.Error("archiveorg: unable to send request: %v", slog.Any("err", err))
		}
		defer res.Body.Close()
		if res.StatusCode > 299 {
			slog.Error("archiveorg: failed with status code", slog.Int("code", res.StatusCode))
		}
	}(entryURL)

	return nil
}
