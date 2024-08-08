// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handler // import "miniflux.app/v2/internal/reader/handler"

import (
	"bytes"
	"errors"
	"log/slog"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/icon"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/storage"
)

var (
	ErrCategoryNotFound = errors.New("fetcher: category not found")
	ErrFeedNotFound     = errors.New("fetcher: feed not found")
	ErrDuplicatedFeed   = errors.New("fetcher: duplicated feed")
)

func CreateFeedFromSubscriptionDiscovery(store *storage.Storage, userID int64, feedCreationRequest *model.FeedCreationRequestFromSubscriptionDiscovery) (*model.Feed, *locale.LocalizedErrorWrapper) {
	slog.Debug("Begin feed creation process from subscription discovery",
		slog.Int64("user_id", userID),
		slog.String("feed_url", feedCreationRequest.FeedURL),
	)

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	if !store.CategoryIDExists(userID, feedCreationRequest.CategoryID) {
		return nil, locale.NewLocalizedErrorWrapper(ErrCategoryNotFound, "error.category_not_found")
	}

	if store.FeedURLExists(userID, feedCreationRequest.FeedURL) {
		return nil, locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
	}

	subscription, parseErr := parser.ParseFeed(feedCreationRequest.FeedURL, feedCreationRequest.Content)
	if parseErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(parseErr, "error.unable_to_parse_feed", parseErr)
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
	subscription.EtagHeader = feedCreationRequest.ETag
	subscription.LastModifiedHeader = feedCreationRequest.LastModified
	subscription.FeedURL = feedCreationRequest.FeedURL
	subscription.DisableHTTP2 = feedCreationRequest.DisableHTTP2
	subscription.WithCategoryID(feedCreationRequest.CategoryID)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, user, true)

	if storeErr := store.CreateFeed(subscription); storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	slog.Debug("Created feed",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", subscription.ID),
		slog.String("feed_url", subscription.FeedURL),
	)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(feedCreationRequest.Username, feedCreationRequest.Password)
	requestBuilder.WithUserAgent(feedCreationRequest.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feedCreationRequest.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(feedCreationRequest.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feedCreationRequest.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feedCreationRequest.DisableHTTP2)

	checkFeedIcon(
		store,
		requestBuilder,
		subscription.ID,
		subscription.SiteURL,
		subscription.IconURL,
	)

	return subscription, nil
}

// CreateFeed fetch, parse and store a new feed.
func CreateFeed(store *storage.Storage, userID int64, feedCreationRequest *model.FeedCreationRequest) (*model.Feed, *locale.LocalizedErrorWrapper) {
	slog.Debug("Begin feed creation process",
		slog.Int64("user_id", userID),
		slog.String("feed_url", feedCreationRequest.FeedURL),
	)

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	if !store.CategoryIDExists(userID, feedCreationRequest.CategoryID) {
		return nil, locale.NewLocalizedErrorWrapper(ErrCategoryNotFound, "error.category_not_found")
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(feedCreationRequest.Username, feedCreationRequest.Password)
	requestBuilder.WithUserAgent(feedCreationRequest.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feedCreationRequest.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(feedCreationRequest.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feedCreationRequest.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feedCreationRequest.DisableHTTP2)

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(feedCreationRequest.FeedURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch feed", slog.String("feed_url", feedCreationRequest.FeedURL), slog.Any("error", localizedError.Error()))
		return nil, localizedError
	}

	responseBody, localizedError := responseHandler.ReadBody(config.Opts.HTTPClientMaxBodySize())
	if localizedError != nil {
		slog.Warn("Unable to fetch feed", slog.String("feed_url", feedCreationRequest.FeedURL), slog.Any("error", localizedError.Error()))
		return nil, localizedError
	}

	if store.FeedURLExists(userID, responseHandler.EffectiveURL()) {
		return nil, locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
	}

	subscription, parseErr := parser.ParseFeed(responseHandler.EffectiveURL(), bytes.NewReader(responseBody))
	if parseErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(parseErr, "error.unable_to_parse_feed", parseErr)
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
	subscription.DisableHTTP2 = feedCreationRequest.DisableHTTP2
	subscription.FetchViaProxy = feedCreationRequest.FetchViaProxy
	subscription.ScraperRules = feedCreationRequest.ScraperRules
	subscription.RewriteRules = feedCreationRequest.RewriteRules
	subscription.BlocklistRules = feedCreationRequest.BlocklistRules
	subscription.KeeplistRules = feedCreationRequest.KeeplistRules
	subscription.UrlRewriteRules = feedCreationRequest.UrlRewriteRules
	subscription.HideGlobally = feedCreationRequest.HideGlobally
	subscription.EtagHeader = responseHandler.ETag()
	subscription.LastModifiedHeader = responseHandler.LastModified()
	subscription.FeedURL = responseHandler.EffectiveURL()
	subscription.WithCategoryID(feedCreationRequest.CategoryID)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, user, true)

	if storeErr := store.CreateFeed(subscription); storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	slog.Debug("Created feed",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", subscription.ID),
		slog.String("feed_url", subscription.FeedURL),
	)

	checkFeedIcon(
		store,
		requestBuilder,
		subscription.ID,
		subscription.SiteURL,
		subscription.IconURL,
	)
	return subscription, nil
}

// RefreshFeed refreshes a feed.
func RefreshFeed(store *storage.Storage, userID, feedID int64, forceRefresh bool) *locale.LocalizedErrorWrapper {
	slog.Debug("Begin feed refresh process",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", feedID),
		slog.Bool("force_refresh", forceRefresh),
	)

	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	originalFeed, storeErr := store.FeedByID(userID, feedID)
	if storeErr != nil {
		return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	if originalFeed == nil {
		return locale.NewLocalizedErrorWrapper(ErrFeedNotFound, "error.feed_not_found")
	}

	weeklyEntryCount := 0
	newTTL := 0
	if config.Opts.PollingScheduler() == model.SchedulerEntryFrequency {
		var weeklyCountErr error
		weeklyEntryCount, weeklyCountErr = store.WeeklyFeedEntryCount(userID, feedID)
		if weeklyCountErr != nil {
			return locale.NewLocalizedErrorWrapper(weeklyCountErr, "error.database_error", weeklyCountErr)
		}
	}

	originalFeed.CheckedNow()
	originalFeed.ScheduleNextCheck(weeklyEntryCount, newTTL)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(originalFeed.Username, originalFeed.Password)
	requestBuilder.WithUserAgent(originalFeed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(originalFeed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxy(config.Opts.HTTPClientProxy())
	requestBuilder.UseProxy(originalFeed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(originalFeed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(originalFeed.DisableHTTP2)

	ignoreHTTPCache := originalFeed.IgnoreHTTPCache || forceRefresh
	if !ignoreHTTPCache {
		requestBuilder.WithETag(originalFeed.EtagHeader)
		requestBuilder.WithLastModified(originalFeed.LastModifiedHeader)
	}

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(originalFeed.FeedURL))
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch feed", slog.String("feed_url", originalFeed.FeedURL), slog.Any("error", localizedError.Error()))
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	if store.AnotherFeedURLExists(userID, originalFeed.ID, responseHandler.EffectiveURL()) {
		localizedError := locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	if ignoreHTTPCache || responseHandler.IsModified(originalFeed.EtagHeader, originalFeed.LastModifiedHeader) {
		slog.Debug("Feed modified",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
		)

		responseBody, localizedError := responseHandler.ReadBody(config.Opts.HTTPClientMaxBodySize())
		if localizedError != nil {
			slog.Warn("Unable to fetch feed", slog.String("feed_url", originalFeed.FeedURL), slog.Any("error", localizedError.Error()))
			return localizedError
		}

		updatedFeed, parseErr := parser.ParseFeed(responseHandler.EffectiveURL(), bytes.NewReader(responseBody))
		if parseErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(parseErr, "error.unable_to_parse_feed", parseErr)

			if errors.Is(parseErr, parser.ErrFeedFormatNotDetected) {
				localizedError = locale.NewLocalizedErrorWrapper(parseErr, "error.feed_format_not_detected", parseErr)
			}

			originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
			store.UpdateFeedError(originalFeed)
			return localizedError
		}

		// If the feed has a TTL defined, we use it to make sure we don't check it too often.
		newTTL = updatedFeed.TTL
		// Set the next check at with updated arguments.
		originalFeed.ScheduleNextCheck(weeklyEntryCount, newTTL)
		slog.Debug("Updated next check date",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.Int("ttl", newTTL),
			slog.Time("new_next_check_at", originalFeed.NextCheckAt),
		)

		originalFeed.Entries = updatedFeed.Entries
		processor.ProcessFeedEntries(store, originalFeed, user, forceRefresh)

		// We don't update existing entries when the crawler is enabled (we crawl only inexisting entries). Unless it is forced to refresh
		updateExistingEntries := forceRefresh || !originalFeed.Crawler
		newEntries, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, updateExistingEntries)
		if storeErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
			store.UpdateFeedError(originalFeed)
			return localizedError
		}

		userIntegrations, intErr := store.Integration(userID)
		if intErr != nil {
			slog.Error("Fetching integrations failed; the refresh process will go on, but no integrations will run this time",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.Any("error", intErr),
			)
		} else if userIntegrations != nil && len(newEntries) > 0 {
			go integration.PushEntries(originalFeed, newEntries, userIntegrations)
		}

		// We update caching headers only if the feed has been modified,
		// because some websites don't return the same headers when replying with a 304.
		originalFeed.EtagHeader = responseHandler.ETag()
		originalFeed.LastModifiedHeader = responseHandler.LastModified()

		checkFeedIcon(
			store,
			requestBuilder,
			originalFeed.ID,
			originalFeed.SiteURL,
			updatedFeed.IconURL,
		)
	} else {
		slog.Debug("Feed not modified",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
		)
	}

	originalFeed.ResetErrorCounter()

	if storeErr := store.UpdateFeed(originalFeed); storeErr != nil {
		localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	return nil
}

func checkFeedIcon(store *storage.Storage, requestBuilder *fetcher.RequestBuilder, feedID int64, websiteURL, feedIconURL string) {
	if !store.HasIcon(feedID) {
		iconFinder := icon.NewIconFinder(requestBuilder, websiteURL, feedIconURL)
		if icon, err := iconFinder.FindIcon(); err != nil {
			slog.Debug("Unable to find feed icon",
				slog.Int64("feed_id", feedID),
				slog.String("website_url", websiteURL),
				slog.String("feed_icon_url", feedIconURL),
				slog.Any("error", err),
			)
		} else if icon == nil {
			slog.Debug("No icon found",
				slog.Int64("feed_id", feedID),
				slog.String("website_url", websiteURL),
				slog.String("feed_icon_url", feedIconURL),
			)
		} else {
			if err := store.CreateFeedIcon(feedID, icon); err != nil {
				slog.Error("Unable to store feed icon",
					slog.Int64("feed_id", feedID),
					slog.String("website_url", websiteURL),
					slog.String("feed_icon_url", feedIconURL),
					slog.Any("error", err),
				)
			}
		}
	}
}
