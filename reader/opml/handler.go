// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/storage"
)

// Handler handles the logic for OPML import/export.
type Handler struct {
	store *storage.Storage
}

// Export exports user feeds to OPML.
func (h *Handler) Export(userID int64) (string, error) {
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		log.Println(err)
		return "", errors.New("unable to fetch feeds")
	}

	var subscriptions SubcriptionList
	for _, feed := range feeds {
		subscriptions = append(subscriptions, &Subcription{
			Title:        feed.Title,
			FeedURL:      feed.FeedURL,
			SiteURL:      feed.SiteURL,
			CategoryName: feed.Category.Title,
		})
	}

	return Serialize(subscriptions), nil
}

// Import parses and create feeds from an OPML import.
func (h *Handler) Import(userID int64, data io.Reader) (err error) {
	subscriptions, err := Parse(data)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		if !h.store.FeedURLExists(userID, subscription.FeedURL) {
			var category *model.Category

			if subscription.CategoryName == "" {
				category, err = h.store.FirstCategory(userID)
				if err != nil {
					log.Println(err)
					return errors.New("unable to find first category")
				}
			} else {
				category, err = h.store.CategoryByTitle(userID, subscription.CategoryName)
				if err != nil {
					log.Println(err)
					return errors.New("unable to search category by title")
				}

				if category == nil {
					category = &model.Category{
						UserID: userID,
						Title:  subscription.CategoryName,
					}

					err := h.store.CreateCategory(category)
					if err != nil {
						log.Println(err)
						return fmt.Errorf(`unable to create this category: "%s"`, subscription.CategoryName)
					}
				}
			}

			feed := &model.Feed{
				UserID:   userID,
				Title:    subscription.Title,
				FeedURL:  subscription.FeedURL,
				SiteURL:  subscription.SiteURL,
				Category: category,
			}

			h.store.CreateFeed(feed)
		}
	}

	return nil
}

// NewHandler creates a new handler for OPML files.
func NewHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}
