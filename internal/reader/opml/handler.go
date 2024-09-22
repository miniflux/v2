// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package opml // import "miniflux.app/v2/internal/reader/opml"

import (
	"fmt"
	"io"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
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

	subscriptions := make(SubcriptionList, 0, len(feeds))
	for _, feed := range feeds {
		var categoryNames CategoryNameList
		for _, category := range feed.Categories {
			categoryNames = append(categoryNames, category.Title)
		}
		subscriptions = append(subscriptions, &Subcription{
			Title:         feed.Title,
			FeedURL:       feed.FeedURL,
			SiteURL:       feed.SiteURL,
			Description:   feed.Description,
			CategoryNames: categoryNames,
		})
	}

	return Serialize(subscriptions), nil
}

// Import parses and create feeds from an OPML import.
func (h *Handler) Import(userID int64, data io.Reader) error {
	subscriptions, err := Parse(data)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		var categories []*model.Category
		for _, categoryName := range subscription.CategoryNames {
			var category *model.Category
			var err error
			category, err = h.store.CategoryByTitle(userID, categoryName)
			if err != nil {
				return fmt.Errorf("opml: unable to search category by title: %w", err)
			}

			if category == nil {
				category, err = h.store.CreateCategory(userID, &model.CategoryRequest{Title: categoryName})
				if err != nil {
					return fmt.Errorf(`opml: unable to create this category: %q`, categoryName)
				}
			}
			categories = append(categories, category)
		}

		feed := &model.Feed{
			UserID:      userID,
			Title:       subscription.Title,
			FeedURL:     subscription.FeedURL,
			SiteURL:     subscription.SiteURL,
			Description: subscription.Description,
			Categories:  categories,
		}
		if !h.store.FeedURLExists(userID, subscription.FeedURL) {
			h.store.CreateFeed(feed)
		} else {
			h.store.UpdateFeed(feed) // TODO maybe only update categories?
		}
	}

	return nil
}

// NewHandler creates a new handler for OPML files.
func NewHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}
