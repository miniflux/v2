// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package feed // import "miniflux.app/reader/feed"

import (
	"fmt"
	"time"

	"miniflux.app/config"
	"miniflux.app/errors"
	"miniflux.app/http/client"
	"miniflux.app/locale"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/browser"
	"miniflux.app/reader/icon"
	"miniflux.app/reader/parser"
	"miniflux.app/reader/processor"
	"miniflux.app/storage"
	"miniflux.app/timer"
)

var (
	errDuplicate        = "This feed already exists (%s)"
	errNotFound         = "Feed %d not found"
	errCategoryNotFound = "Category not found for this user"
)

// Handler contains all the logic to create and refresh feeds.
type Handler struct {
	store *storage.Storage
}

// CreateFeed fetch, parse and store a new feed.
func (h *Handler) CreateFeed(userID, categoryID int64, url string, crawler bool, userAgent, username, password, scraperRules, rewriteRules, blocklistRules, keeplistRules string, fetchViaProxy bool) (*model.Feed, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Handler:CreateFeed] feedUrl=%s", url))

	if !h.store.CategoryExists(userID, categoryID) {
		return nil, errors.NewLocalizedError(errCategoryNotFound)
	}

	request := client.NewClientWithConfig(url, config.Opts)
	request.WithCredentials(username, password)
	request.WithUserAgent(userAgent)

	if fetchViaProxy {
		request.WithProxy()
	}

	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		return nil, requestErr
	}

	if h.store.FeedURLExists(userID, response.EffectiveURL) {
		return nil, errors.NewLocalizedError(errDuplicate, response.EffectiveURL)
	}

	subscription, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
	if parseErr != nil {
		return nil, parseErr
	}

	subscription.UserID = userID
	subscription.WithCategoryID(categoryID)
	subscription.WithBrowsingParameters(crawler, userAgent, username, password, scraperRules, rewriteRules, blocklistRules, keeplistRules, fetchViaProxy)
	subscription.WithClientResponse(response)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(h.store, subscription)

	if storeErr := h.store.CreateFeed(subscription); storeErr != nil {
		return nil, storeErr
	}

	logger.Debug("[Handler:CreateFeed] Feed saved with ID: %d", subscription.ID)

	checkFeedIcon(h.store, subscription.ID, subscription.SiteURL, fetchViaProxy)
	return subscription, nil
}

// RefreshFeed refreshes a feed.
func (h *Handler) RefreshFeed(userID, feedID int64) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Handler:RefreshFeed] feedID=%d", feedID))
	userLanguage := h.store.UserLanguage(userID)
	printer := locale.NewPrinter(userLanguage)

	originalFeed, storeErr := h.store.FeedByID(userID, feedID)
	if storeErr != nil {
		return storeErr
	}

	if originalFeed == nil {
		return errors.NewLocalizedError(errNotFound, feedID)
	}

	weeklyEntryCount := 0
	if config.Opts.PollingScheduler() == model.SchedulerEntryFrequency {
		var weeklyCountErr error
		weeklyEntryCount, weeklyCountErr = h.store.WeeklyFeedEntryCount(userID, feedID)
		if weeklyCountErr != nil {
			return weeklyCountErr
		}
	}

	originalFeed.CheckedNow()
	originalFeed.ScheduleNextCheck(weeklyEntryCount)

	request := client.NewClientWithConfig(originalFeed.FeedURL, config.Opts)
	request.WithCredentials(originalFeed.Username, originalFeed.Password)
	request.WithUserAgent(originalFeed.UserAgent)

	if !originalFeed.IgnoreHTTPCache {
		request.WithCacheHeaders(originalFeed.EtagHeader, originalFeed.LastModifiedHeader)
	}

	if originalFeed.FetchViaProxy {
		request.WithProxy()
	}

	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		originalFeed.WithError(requestErr.Localize(printer))
		h.store.UpdateFeedError(originalFeed)
		return requestErr
	}

	if h.store.AnotherFeedURLExists(userID, originalFeed.ID, response.EffectiveURL) {
		storeErr := errors.NewLocalizedError(errDuplicate, response.EffectiveURL)
		originalFeed.WithError(storeErr.Error())
		h.store.UpdateFeedError(originalFeed)
		return storeErr
	}

	if originalFeed.IgnoreHTTPCache || response.IsModified(originalFeed.EtagHeader, originalFeed.LastModifiedHeader) {
		logger.Debug("[Handler:RefreshFeed] Feed #%d has been modified", feedID)

		updatedFeed, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
		if parseErr != nil {
			originalFeed.WithError(parseErr.Localize(printer))
			h.store.UpdateFeedError(originalFeed)
			return parseErr
		}

		originalFeed.Entries = updatedFeed.Entries
		processor.ProcessFeedEntries(h.store, originalFeed)

		// We don't update existing entries when the crawler is enabled (we crawl only inexisting entries).
		if storeErr := h.store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, !originalFeed.Crawler); storeErr != nil {
			originalFeed.WithError(storeErr.Error())
			h.store.UpdateFeedError(originalFeed)
			return storeErr
		}

		// We update caching headers only if the feed has been modified,
		// because some websites don't return the same headers when replying with a 304.
		originalFeed.WithClientResponse(response)
		checkFeedIcon(h.store, originalFeed.ID, originalFeed.SiteURL, originalFeed.FetchViaProxy)
	} else {
		logger.Debug("[Handler:RefreshFeed] Feed #%d not modified", feedID)
	}

	originalFeed.ResetErrorCounter()

	if storeErr := h.store.UpdateFeed(originalFeed); storeErr != nil {
		originalFeed.WithError(storeErr.Error())
		h.store.UpdateFeedError(originalFeed)
		return storeErr
	}

	return nil
}

// NewFeedHandler returns a feed handler.
func NewFeedHandler(store *storage.Storage) *Handler {
	return &Handler{store}
}

func checkFeedIcon(store *storage.Storage, feedID int64, websiteURL string, fetchViaProxy bool) {
	if !store.HasIcon(feedID) {
		icon, err := icon.FindIcon(websiteURL, fetchViaProxy)
		if err != nil {
			logger.Debug("CheckFeedIcon: %v (feedID=%d websiteURL=%s)", err, feedID, websiteURL)
		} else if icon == nil {
			logger.Debug("CheckFeedIcon: No icon found (feedID=%d websiteURL=%s)", feedID, websiteURL)
		} else {
			if err := store.CreateFeedIcon(feedID, icon); err != nil {
				logger.Debug("CheckFeedIcon: %v (feedID=%d websiteURL=%s)", err, feedID, websiteURL)
			}
		}
	}
}
