// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handler // import "miniflux.app/v2/internal/reader/handler"

import (
	"fmt"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/errors"
	"miniflux.app/v2/internal/http/client"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/browser"
	"miniflux.app/v2/internal/reader/icon"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/timer"
)

var (
	errDuplicate        = "This feed already exists (%s)"
	errNotFound         = "Feed %d not found"
	errCategoryNotFound = "Category not found for this user"
)

// CreateFeed fetch, parse and store a new feed.
func CreateFeed(store *storage.Storage, userID int64, feedCreationRequest *model.FeedCreationRequest) (*model.Feed, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[CreateFeed] FeedURL=%s", feedCreationRequest.FeedURL))

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return nil, storeErr
	}

	if !store.CategoryIDExists(userID, feedCreationRequest.CategoryID) {
		return nil, errors.NewLocalizedError(errCategoryNotFound)
	}

	request := client.NewClientWithConfig(feedCreationRequest.FeedURL, config.Opts)
	request.WithCredentials(feedCreationRequest.Username, feedCreationRequest.Password)
	request.WithUserAgent(feedCreationRequest.UserAgent)
	request.WithCookie(feedCreationRequest.Cookie)
	request.AllowSelfSignedCertificates = feedCreationRequest.AllowSelfSignedCertificates

	if feedCreationRequest.FetchViaProxy {
		request.WithProxy()
	}

	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		return nil, requestErr
	}

	if store.FeedURLExists(userID, response.EffectiveURL) {
		return nil, errors.NewLocalizedError(errDuplicate, response.EffectiveURL)
	}

	subscription, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
	if parseErr != nil {
		return nil, parseErr
	}

	subscription.UserID = userID
	subscription.UserAgent = feedCreationRequest.UserAgent
	subscription.Cookie = feedCreationRequest.Cookie
	subscription.Username = feedCreationRequest.Username
	subscription.Password = feedCreationRequest.Password
	subscription.Crawler = feedCreationRequest.Crawler
	subscription.Disabled = feedCreationRequest.Disabled
	subscription.IgnoreHTTPCache = feedCreationRequest.IgnoreHTTPCache
	subscription.AllowSelfSignedCertificates = feedCreationRequest.AllowSelfSignedCertificates
	subscription.FetchViaProxy = feedCreationRequest.FetchViaProxy
	subscription.ScraperRules = feedCreationRequest.ScraperRules
	subscription.RewriteRules = feedCreationRequest.RewriteRules
	subscription.BlocklistRules = feedCreationRequest.BlocklistRules
	subscription.KeeplistRules = feedCreationRequest.KeeplistRules
	subscription.UrlRewriteRules = feedCreationRequest.UrlRewriteRules
	subscription.WithCategoryID(feedCreationRequest.CategoryID)
	subscription.WithClientResponse(response)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, user, true)

	if storeErr := store.CreateFeed(subscription); storeErr != nil {
		return nil, storeErr
	}

	logger.Debug("[CreateFeed] Feed saved with ID: %d", subscription.ID)

	checkFeedIcon(
		store,
		subscription.ID,
		subscription.SiteURL,
		subscription.IconURL,
		feedCreationRequest.UserAgent,
		feedCreationRequest.FetchViaProxy,
		feedCreationRequest.AllowSelfSignedCertificates,
	)
	return subscription, nil
}

// RefreshFeed refreshes a feed.
func RefreshFeed(store *storage.Storage, userID, feedID int64, forceRefresh bool) error {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[RefreshFeed] feedID=%d", feedID))
	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return storeErr
	}

	printer := locale.NewPrinter(user.Language)

	originalFeed, storeErr := store.FeedByID(userID, feedID)
	if storeErr != nil {
		return storeErr
	}

	if originalFeed == nil {
		return errors.NewLocalizedError(errNotFound, feedID)
	}

	weeklyEntryCount := 0
	if config.Opts.PollingScheduler() == model.SchedulerEntryFrequency {
		var weeklyCountErr error
		weeklyEntryCount, weeklyCountErr = store.WeeklyFeedEntryCount(userID, feedID)
		if weeklyCountErr != nil {
			return weeklyCountErr
		}
	}

	originalFeed.CheckedNow()
	originalFeed.ScheduleNextCheck(weeklyEntryCount)

	request := client.NewClientWithConfig(originalFeed.FeedURL, config.Opts)
	request.WithCredentials(originalFeed.Username, originalFeed.Password)
	request.WithUserAgent(originalFeed.UserAgent)
	request.WithCookie(originalFeed.Cookie)
	request.AllowSelfSignedCertificates = originalFeed.AllowSelfSignedCertificates

	if !originalFeed.IgnoreHTTPCache {
		request.WithCacheHeaders(originalFeed.EtagHeader, originalFeed.LastModifiedHeader)
	}

	if originalFeed.FetchViaProxy {
		request.WithProxy()
	}

	response, requestErr := browser.Exec(request)
	if requestErr != nil {
		originalFeed.WithError(requestErr.Localize(printer))
		store.UpdateFeedError(originalFeed)
		return requestErr
	}

	if store.AnotherFeedURLExists(userID, originalFeed.ID, response.EffectiveURL) {
		storeErr := errors.NewLocalizedError(errDuplicate, response.EffectiveURL)
		originalFeed.WithError(storeErr.Error())
		store.UpdateFeedError(originalFeed)
		return storeErr
	}

	if originalFeed.IgnoreHTTPCache || response.IsModified(originalFeed.EtagHeader, originalFeed.LastModifiedHeader) {
		logger.Debug("[RefreshFeed] Feed #%d has been modified", feedID)

		updatedFeed, parseErr := parser.ParseFeed(response.EffectiveURL, response.BodyAsString())
		if parseErr != nil {
			originalFeed.WithError(parseErr.Localize(printer))
			store.UpdateFeedError(originalFeed)
			return parseErr
		}

		originalFeed.Entries = updatedFeed.Entries
		processor.ProcessFeedEntries(store, originalFeed, user, forceRefresh)

		// We don't update existing entries when the crawler is enabled (we crawl only inexisting entries). Unless it is forced to refresh
		updateExistingEntries := forceRefresh || !originalFeed.Crawler
		newEntries, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, updateExistingEntries)
		if storeErr != nil {
			originalFeed.WithError(storeErr.Error())
			store.UpdateFeedError(originalFeed)
			return storeErr
		}

		userIntegrations, intErr := store.Integration(userID)
		if intErr != nil {
			logger.Error("[RefreshFeed] Fetching integrations for user %d failed: %v; the refresh process will go on, but no integrations will run this time.", userID, intErr)
		} else if userIntegrations != nil && len(newEntries) > 0 {
			go integration.PushEntries(originalFeed, newEntries, userIntegrations)
		}

		// We update caching headers only if the feed has been modified,
		// because some websites don't return the same headers when replying with a 304.
		originalFeed.WithClientResponse(response)

		checkFeedIcon(
			store,
			originalFeed.ID,
			originalFeed.SiteURL,
			updatedFeed.IconURL,
			originalFeed.UserAgent,
			originalFeed.FetchViaProxy,
			originalFeed.AllowSelfSignedCertificates,
		)
	} else {
		logger.Debug("[RefreshFeed] Feed #%d not modified", feedID)
	}

	originalFeed.ResetErrorCounter()

	if storeErr := store.UpdateFeed(originalFeed); storeErr != nil {
		originalFeed.WithError(storeErr.Error())
		store.UpdateFeedError(originalFeed)
		return storeErr
	}

	return nil
}

func checkFeedIcon(store *storage.Storage, feedID int64, websiteURL, iconURL, userAgent string, fetchViaProxy, allowSelfSignedCertificates bool) {
	if !store.HasIcon(feedID) {
		icon, err := icon.FindIcon(websiteURL, iconURL, userAgent, fetchViaProxy, allowSelfSignedCertificates)
		if err != nil {
			logger.Debug(`[CheckFeedIcon] %v (feedID=%d websiteURL=%s)`, err, feedID, websiteURL)
		} else if icon == nil {
			logger.Debug(`[CheckFeedIcon] No icon found (feedID=%d websiteURL=%s)`, feedID, websiteURL)
		} else {
			if err := store.CreateFeedIcon(feedID, icon); err != nil {
				logger.Debug(`[CheckFeedIcon] %v (feedID=%d websiteURL=%s)`, err, feedID, websiteURL)
			}
		}
	}
}
