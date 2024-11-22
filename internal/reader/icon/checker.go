// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package icon // import "miniflux.app/v2/internal/reader/icon"

import (
	"log/slog"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/storage"
)

type IconChecker struct {
	store *storage.Storage
	feed  *model.Feed
}

func NewIconChecker(store *storage.Storage, feed *model.Feed) *IconChecker {
	return &IconChecker{
		store: store,
		feed:  feed,
	}
}

func (c *IconChecker) fetchAndStoreIcon() {
	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUserAgent(c.feed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(c.feed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(c.feed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(c.feed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(c.feed.DisableHTTP2)

	iconFinder := NewIconFinder(requestBuilder, c.feed.SiteURL, c.feed.IconURL)
	if icon, err := iconFinder.FindIcon(); err != nil {
		slog.Debug("Unable to find feed icon",
			slog.Int64("feed_id", c.feed.ID),
			slog.String("website_url", c.feed.SiteURL),
			slog.String("feed_icon_url", c.feed.IconURL),
			slog.Any("error", err),
		)
	} else if icon == nil {
		slog.Debug("No icon found",
			slog.Int64("feed_id", c.feed.ID),
			slog.String("website_url", c.feed.SiteURL),
			slog.String("feed_icon_url", c.feed.IconURL),
		)
	} else {
		if err := c.store.StoreFeedIcon(c.feed.ID, icon); err != nil {
			slog.Error("Unable to store feed icon",
				slog.Int64("feed_id", c.feed.ID),
				slog.String("website_url", c.feed.SiteURL),
				slog.String("feed_icon_url", c.feed.IconURL),
				slog.Any("error", err),
			)
		} else {
			slog.Debug("Feed icon stored",
				slog.Int64("feed_id", c.feed.ID),
				slog.String("website_url", c.feed.SiteURL),
				slog.String("feed_icon_url", c.feed.IconURL),
				slog.Int64("icon_id", icon.ID),
				slog.String("icon_hash", icon.Hash),
			)
		}
	}
}

func (c *IconChecker) CreateFeedIconIfMissing() {
	if c.store.HasFeedIcon(c.feed.ID) {
		slog.Debug("Feed icon already exists",
			slog.Int64("feed_id", c.feed.ID),
		)
		return
	}

	c.fetchAndStoreIcon()
}

func (c *IconChecker) UpdateOrCreateFeedIcon() {
	c.fetchAndStoreIcon()
}
