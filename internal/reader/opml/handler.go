// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/validator"
)

// Handler handles the logic for OPML import/export.
type Handler struct {
	store *storage.Storage
}

// Export exports user feeds to OPML.
func (h *Handler) Export(userID int64) (string, error) {
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		return "", err
	}

	subscriptions := make([]subcription, 0, len(feeds))
	for _, feed := range feeds {
		subscriptions = append(subscriptions, subcription{
			Title:        feed.Title,
			FeedURL:      feed.FeedURL,
			SiteURL:      feed.SiteURL,
			Description:  feed.Description,
			CategoryName: feed.Category.Title,

			ScraperRules:                feed.ScraperRules,
			RewriteRules:                feed.RewriteRules,
			UrlRewriteRules:             feed.UrlRewriteRules,
			BlocklistRules:              feed.BlocklistRules,
			KeeplistRules:               feed.KeeplistRules,
			BlockFilterEntryRules:       feed.BlockFilterEntryRules,
			KeepFilterEntryRules:        feed.KeepFilterEntryRules,
			UserAgent:                   feed.UserAgent,
			ProxyURL:                    feed.ProxyURL,
			Crawler:                     feed.Crawler,
			IgnoreHTTPCache:             feed.IgnoreHTTPCache,
			FetchViaProxy:               feed.FetchViaProxy,
			Disabled:                    feed.Disabled,
			NoMediaPlayer:               feed.NoMediaPlayer,
			HideGlobally:                feed.HideGlobally,
			AllowSelfSignedCertificates: feed.AllowSelfSignedCertificates,
			DisableHTTP2:                feed.DisableHTTP2,
			IgnoreEntryUpdates:          feed.IgnoreEntryUpdates,
		})
	}

	return serialize(subscriptions), nil
}

// Import parses and create feeds from an OPML import.
func (h *Handler) Import(userID int64, data io.Reader) error {
	subscriptions, err := parse(data)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		if h.store.FeedURLExists(userID, subscription.FeedURL) {
			continue
		}

		category, err := h.resolveCategory(userID, subscription.CategoryName)
		if err != nil {
			return err
		}

		if validationErr := validateSubscription(userID, category.ID, h.store, subscription); validationErr != nil {
			return fmt.Errorf(`opml: invalid feed settings for %q: %w`, subscription.FeedURL, validationErr)
		}

		feed := &model.Feed{
			UserID:      userID,
			Title:       subscription.Title,
			FeedURL:     subscription.FeedURL,
			SiteURL:     subscription.SiteURL,
			Description: subscription.Description,
			Category:    category,
		}
		applySubscriptionSettings(feed, subscription)
		if err := h.store.CreateFeed(feed); err != nil {
			return fmt.Errorf(`opml: unable to create this feed: %q`, subscription.FeedURL)
		}
	}

	return nil
}

func (h *Handler) resolveCategory(userID int64, categoryName string) (*model.Category, error) {
	if categoryName == "" {
		category, err := h.store.FirstCategory(userID)
		if err != nil {
			return nil, fmt.Errorf("opml: unable to find first category: %w", err)
		}
		return category, nil
	}

	category, err := h.store.CategoryByTitle(userID, categoryName)
	if err != nil {
		return nil, fmt.Errorf("opml: unable to search category by title: %w", err)
	}

	if category == nil {
		category, err = h.store.CreateCategory(userID, &model.CategoryCreationRequest{Title: categoryName})
		if err != nil {
			return nil, fmt.Errorf(`opml: unable to create this category: %q`, categoryName)
		}
	}

	return category, nil
}

func applySubscriptionSettings(feed *model.Feed, s subcription) {
	feed.ScraperRules = s.ScraperRules
	feed.RewriteRules = s.RewriteRules
	feed.UrlRewriteRules = s.UrlRewriteRules
	feed.BlocklistRules = s.BlocklistRules
	feed.KeeplistRules = s.KeeplistRules
	feed.BlockFilterEntryRules = s.BlockFilterEntryRules
	feed.KeepFilterEntryRules = s.KeepFilterEntryRules
	feed.UserAgent = s.UserAgent
	feed.ProxyURL = s.ProxyURL
	feed.Crawler = s.Crawler
	feed.IgnoreHTTPCache = s.IgnoreHTTPCache
	feed.FetchViaProxy = s.FetchViaProxy
	feed.Disabled = s.Disabled
	feed.NoMediaPlayer = s.NoMediaPlayer
	feed.HideGlobally = s.HideGlobally
	feed.AllowSelfSignedCertificates = s.AllowSelfSignedCertificates
	feed.DisableHTTP2 = s.DisableHTTP2
	feed.IgnoreEntryUpdates = s.IgnoreEntryUpdates
}

func validateSubscription(userID, categoryID int64, store *storage.Storage, s subcription) error {
	feedCreationRequest := &model.FeedCreationRequest{
		FeedURL:                     s.FeedURL,
		CategoryID:                  categoryID,
		UserAgent:                   s.UserAgent,
		Crawler:                     s.Crawler,
		IgnoreEntryUpdates:          s.IgnoreEntryUpdates,
		Disabled:                    s.Disabled,
		NoMediaPlayer:               s.NoMediaPlayer,
		IgnoreHTTPCache:             s.IgnoreHTTPCache,
		AllowSelfSignedCertificates: s.AllowSelfSignedCertificates,
		FetchViaProxy:               s.FetchViaProxy,
		HideGlobally:                s.HideGlobally,
		DisableHTTP2:                s.DisableHTTP2,
		ScraperRules:                s.ScraperRules,
		RewriteRules:                s.RewriteRules,
		BlocklistRules:              s.BlocklistRules,
		KeeplistRules:               s.KeeplistRules,
		BlockFilterEntryRules:       s.BlockFilterEntryRules,
		KeepFilterEntryRules:        s.KeepFilterEntryRules,
		UrlRewriteRules:             s.UrlRewriteRules,
		ProxyURL:                    s.ProxyURL,
	}

	if validationErr := validator.ValidateFeedCreation(store, userID, feedCreationRequest); validationErr != nil {
		return validationErr.Error()
	}

	return nil
}

// NewHandler creates a new handler for OPML files.
func NewHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}
