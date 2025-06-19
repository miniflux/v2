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
	"miniflux.app/v2/internal/proxyrotator"
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
		slog.String("proxy_url", feedCreationRequest.ProxyURL),
	)

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
	subscription.BlockFilterEntryRules = feedCreationRequest.BlockFilterEntryRules
	subscription.KeepFilterEntryRules = feedCreationRequest.KeepFilterEntryRules
	subscription.EtagHeader = feedCreationRequest.ETag
	subscription.LastModifiedHeader = feedCreationRequest.LastModified
	subscription.FeedURL = feedCreationRequest.FeedURL
	subscription.DisableHTTP2 = feedCreationRequest.DisableHTTP2
	subscription.WithCategoryID(feedCreationRequest.CategoryID)
	subscription.ProxyURL = feedCreationRequest.ProxyURL
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, userID, true)

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
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(feedCreationRequest.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(feedCreationRequest.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(feedCreationRequest.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(feedCreationRequest.DisableHTTP2)

	icon.NewIconChecker(store, subscription).UpdateOrCreateFeedIcon()

	return subscription, nil
}

// CreateFeed fetch, parse and store a new feed.
func CreateFeed(store *storage.Storage, userID int64, feedCreationRequest *model.FeedCreationRequest) (*model.Feed, *locale.LocalizedErrorWrapper) {
	slog.Debug("Begin feed creation process",
		slog.Int64("user_id", userID),
		slog.String("feed_url", feedCreationRequest.FeedURL),
		slog.String("proxy_url", feedCreationRequest.ProxyURL),
	)

	if !store.CategoryIDExists(userID, feedCreationRequest.CategoryID) {
		return nil, locale.NewLocalizedErrorWrapper(ErrCategoryNotFound, "error.category_not_found")
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(feedCreationRequest.Username, feedCreationRequest.Password)
	requestBuilder.WithUserAgent(feedCreationRequest.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(feedCreationRequest.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(feedCreationRequest.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(feedCreationRequest.FetchViaProxy)
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
	subscription.UrlRewriteRules = feedCreationRequest.UrlRewriteRules
	subscription.BlocklistRules = feedCreationRequest.BlocklistRules
	subscription.KeeplistRules = feedCreationRequest.KeeplistRules
	subscription.BlockFilterEntryRules = feedCreationRequest.BlockFilterEntryRules
	subscription.KeepFilterEntryRules = feedCreationRequest.KeepFilterEntryRules
	subscription.HideGlobally = feedCreationRequest.HideGlobally
	subscription.EtagHeader = responseHandler.ETag()
	subscription.LastModifiedHeader = responseHandler.LastModified()
	subscription.FeedURL = responseHandler.EffectiveURL()
	subscription.ProxyURL = feedCreationRequest.ProxyURL
	subscription.WithCategoryID(feedCreationRequest.CategoryID)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, userID, true)

	if storeErr := store.CreateFeed(subscription); storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	slog.Debug("Created feed",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", subscription.ID),
		slog.String("feed_url", subscription.FeedURL),
	)

	icon.NewIconChecker(store, subscription).UpdateOrCreateFeedIcon()

	return subscription, nil
}

// RefreshFeed refreshes a feed.
func RefreshFeed(store *storage.Storage, userID, feedID int64, forceRefresh bool) *locale.LocalizedErrorWrapper {
	slog.Debug("Begin feed refresh process",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", feedID),
		slog.Bool("force_refresh", forceRefresh),
	)

	originalFeed, storeErr := store.FeedByID(userID, feedID)
	if storeErr != nil {
		return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	if originalFeed == nil {
		return locale.NewLocalizedErrorWrapper(ErrFeedNotFound, "error.feed_not_found")
	}

	weeklyEntryCount := 0
	refreshDelayInMinutes := 0
	if config.Opts.PollingScheduler() == model.SchedulerEntryFrequency {
		var weeklyCountErr error
		weeklyEntryCount, weeklyCountErr = store.WeeklyFeedEntryCount(userID, feedID)
		if weeklyCountErr != nil {
			return locale.NewLocalizedErrorWrapper(weeklyCountErr, "error.database_error", weeklyCountErr)
		}
	}

	originalFeed.CheckedNow()
	originalFeed.ScheduleNextCheck(weeklyEntryCount, refreshDelayInMinutes)

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(originalFeed.Username, originalFeed.Password)
	requestBuilder.WithUserAgent(originalFeed.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(originalFeed.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(originalFeed.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(originalFeed.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(originalFeed.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(originalFeed.DisableHTTP2)

	ignoreHTTPCache := originalFeed.IgnoreHTTPCache || forceRefresh
	if !ignoreHTTPCache {
		requestBuilder.WithETag(originalFeed.EtagHeader)
		requestBuilder.WithLastModified(originalFeed.LastModifiedHeader)
	}

	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(originalFeed.FeedURL))
	defer responseHandler.Close()

	if responseHandler.IsRateLimited() {
		retryDelayInSeconds := responseHandler.ParseRetryDelay()
		refreshDelayInMinutes = retryDelayInSeconds / 60
		calculatedNextCheckIntervalInMinutes := originalFeed.ScheduleNextCheck(weeklyEntryCount, refreshDelayInMinutes)

		slog.Warn("Feed is rate limited",
			slog.String("feed_url", originalFeed.FeedURL),
			slog.Int("retry_delay_in_seconds", retryDelayInSeconds),
			slog.Int("refresh_delay_in_minutes", refreshDelayInMinutes),
			slog.Int("calculated_next_check_interval_in_minutes", calculatedNextCheckIntervalInMinutes),
			slog.Time("new_next_check_at", originalFeed.NextCheckAt),
		)
	}

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to fetch feed",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.String("feed_url", originalFeed.FeedURL),
			slog.Any("error", localizedError.Error()),
		)
		user, storeErr := store.UserByID(userID)
		if storeErr != nil {
			return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
		}
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	if store.AnotherFeedURLExists(userID, originalFeed.ID, responseHandler.EffectiveURL()) {
		localizedError := locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
		user, storeErr := store.UserByID(userID)
		if storeErr != nil {
			return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
		}
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	if ignoreHTTPCache || responseHandler.IsModified(originalFeed.EtagHeader, originalFeed.LastModifiedHeader) {
		slog.Debug("Feed modified",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.String("etag_header", originalFeed.EtagHeader),
			slog.String("last_modified_header", originalFeed.LastModifiedHeader),
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
			user, storeErr := store.UserByID(userID)
			if storeErr != nil {
				return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			}

			originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
			store.UpdateFeedError(originalFeed)
			return localizedError
		}

		// Use the RSS TTL value, or the Cache-Control or Expires HTTP headers if available.
		// Otherwise, we use the default value from the configuration (min interval parameter).
		feedTTLValue := updatedFeed.TTL
		cacheControlMaxAgeValue := responseHandler.CacheControlMaxAgeInMinutes()
		expiresValue := responseHandler.ExpiresInMinutes()
		refreshDelayInMinutes = max(feedTTLValue, cacheControlMaxAgeValue, expiresValue)

		// Set the next check at with updated arguments.
		calculatedNextCheckIntervalInMinutes := originalFeed.ScheduleNextCheck(weeklyEntryCount, refreshDelayInMinutes)

		slog.Debug("Updated next check date",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.String("feed_url", originalFeed.FeedURL),
			slog.Int("feed_ttl_minutes", feedTTLValue),
			slog.Int("cache_control_max_age_in_minutes", cacheControlMaxAgeValue),
			slog.Int("expires_in_minutes", expiresValue),
			slog.Int("refresh_delay_in_minutes", refreshDelayInMinutes),
			slog.Int("calculated_next_check_interval_in_minutes", calculatedNextCheckIntervalInMinutes),
			slog.Time("new_next_check_at", originalFeed.NextCheckAt),
		)

		originalFeed.Entries = updatedFeed.Entries
		processor.ProcessFeedEntries(store, originalFeed, userID, forceRefresh)

		// We don't update existing entries when the crawler is enabled (we crawl only inexisting entries). Unless it is forced to refresh
		updateExistingEntries := forceRefresh || !originalFeed.Crawler
		newEntries, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, updateExistingEntries)
		if storeErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			user, storeErr := store.UserByID(userID)
			if storeErr != nil {
				return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			}
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

		originalFeed.EtagHeader = responseHandler.ETag()
		originalFeed.LastModifiedHeader = responseHandler.LastModified()

		originalFeed.IconURL = updatedFeed.IconURL
		iconChecker := icon.NewIconChecker(store, originalFeed)
		if forceRefresh {
			iconChecker.UpdateOrCreateFeedIcon()
		} else {
			iconChecker.CreateFeedIconIfMissing()
		}
	} else {
		slog.Debug("Feed not modified",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
		)

		// Last-Modified may be updated even if ETag is not. In this case, per
		// RFC9111 sections 3.2 and 4.3.4, the stored response must be updated.
		if responseHandler.LastModified() != "" {
			originalFeed.LastModifiedHeader = responseHandler.LastModified()
		}
	}

	originalFeed.ResetErrorCounter()

	if storeErr := store.UpdateFeed(originalFeed); storeErr != nil {
		localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
		user, storeErr := store.UserByID(userID)
		if storeErr != nil {
			return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
		}
		originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
		store.UpdateFeedError(originalFeed)
		return localizedError
	}

	return nil
}
