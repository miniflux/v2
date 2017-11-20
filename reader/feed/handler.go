// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package feed

import (
	"fmt"
	"github.com/miniflux/miniflux2/errors"
	"github.com/miniflux/miniflux2/helper"
	"github.com/miniflux/miniflux2/model"
	"github.com/miniflux/miniflux2/reader/http"
	"github.com/miniflux/miniflux2/reader/icon"
	"github.com/miniflux/miniflux2/storage"
	"log"
	"time"
)

var (
	errRequestFailed = "Unable to execute request: %v"
	errServerFailure = "Unable to fetch feed (statusCode=%d)."
	errDuplicate     = "This feed already exists (%s)."
	errNotFound      = "Feed %d not found"
)

// Handler contains all the logic to create and refresh feeds.
type Handler struct {
	store *storage.Storage
}

// CreateFeed fetch, parse and store a new feed.
func (h *Handler) CreateFeed(userID, categoryID int64, url string) (*model.Feed, error) {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Handler:CreateFeed] feedUrl=%s", url))

	client := http.NewHttpClient(url)
	response, err := client.Get()
	if err != nil {
		return nil, errors.NewLocalizedError(errRequestFailed, err)
	}

	if response.HasServerFailure() {
		return nil, errors.NewLocalizedError(errServerFailure, response.StatusCode)
	}

	if h.store.FeedURLExists(userID, response.EffectiveURL) {
		return nil, errors.NewLocalizedError(errDuplicate, response.EffectiveURL)
	}

	subscription, err := parseFeed(response.Body)
	if err != nil {
		return nil, err
	}

	subscription.Category = &model.Category{ID: categoryID}
	subscription.EtagHeader = response.ETag
	subscription.LastModifiedHeader = response.LastModified
	subscription.FeedURL = response.EffectiveURL
	subscription.UserID = userID

	err = h.store.CreateFeed(subscription)
	if err != nil {
		return nil, err
	}

	log.Println("[Handler:CreateFeed] Feed saved with ID:", subscription.ID)

	icon, err := icon.FindIcon(subscription.SiteURL)
	if err != nil {
		log.Println(err)
	} else if icon == nil {
		log.Printf("No icon found for feedID=%d\n", subscription.ID)
	} else {
		h.store.CreateFeedIcon(subscription, icon)
	}

	return subscription, nil
}

// RefreshFeed fetch and update a feed if necessary.
func (h *Handler) RefreshFeed(userID, feedID int64) error {
	defer helper.ExecutionTime(time.Now(), fmt.Sprintf("[Handler:RefreshFeed] feedID=%d", feedID))

	originalFeed, err := h.store.GetFeedById(userID, feedID)
	if err != nil {
		return err
	}

	if originalFeed == nil {
		return errors.NewLocalizedError(errNotFound, feedID)
	}

	client := http.NewHttpClientWithCacheHeaders(originalFeed.FeedURL, originalFeed.EtagHeader, originalFeed.LastModifiedHeader)
	response, err := client.Get()
	if err != nil {
		customErr := errors.NewLocalizedError(errRequestFailed, err)
		originalFeed.ParsingErrorCount++
		originalFeed.ParsingErrorMsg = customErr.Error()
		h.store.UpdateFeed(originalFeed)
		return customErr
	}

	originalFeed.CheckedAt = time.Now()

	if response.HasServerFailure() {
		err := errors.NewLocalizedError(errServerFailure, response.StatusCode)
		originalFeed.ParsingErrorCount++
		originalFeed.ParsingErrorMsg = err.Error()
		h.store.UpdateFeed(originalFeed)
		return err
	}

	if response.IsModified(originalFeed.EtagHeader, originalFeed.LastModifiedHeader) {
		log.Printf("[Handler:RefreshFeed] Feed #%d has been modified\n", feedID)

		subscription, err := parseFeed(response.Body)
		if err != nil {
			originalFeed.ParsingErrorCount++
			originalFeed.ParsingErrorMsg = err.Error()
			h.store.UpdateFeed(originalFeed)
			return err
		}

		originalFeed.EtagHeader = response.ETag
		originalFeed.LastModifiedHeader = response.LastModified

		if err := h.store.UpdateEntries(originalFeed.UserID, originalFeed.ID, subscription.Entries); err != nil {
			return err
		}

		if !h.store.HasIcon(originalFeed.ID) {
			log.Println("[Handler:RefreshFeed] Looking for feed icon")
			icon, err := icon.FindIcon(originalFeed.SiteURL)
			if err != nil {
				log.Println("[Handler:RefreshFeed]", err)
			} else {
				h.store.CreateFeedIcon(originalFeed, icon)
			}
		}
	} else {
		log.Printf("[Handler:RefreshFeed] Feed #%d not modified\n", feedID)
	}

	originalFeed.ParsingErrorCount = 0
	originalFeed.ParsingErrorMsg = ""

	return h.store.UpdateFeed(originalFeed)
}

// NewFeedHandler returns a feed handler.
func NewFeedHandler(store *storage.Storage) *Handler {
	return &Handler{store: store}
}
