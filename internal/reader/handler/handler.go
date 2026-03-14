// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handler // import "miniflux.app/v2/internal/reader/handler"

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/integration"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/proxyrotator"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/icon"
	"miniflux.app/v2/internal/reader/parser"
	"miniflux.app/v2/internal/reader/pinchtab"
	"miniflux.app/v2/internal/reader/processor"
	"miniflux.app/v2/internal/reader/webscraper"
	"miniflux.app/v2/internal/storage"
)

var (
	ErrCategoryNotFound = errors.New("fetcher: category not found")
	ErrFeedNotFound     = errors.New("fetcher: feed not found")
	ErrDuplicatedFeed   = errors.New("fetcher: duplicated feed")
)

func getTranslatedLocalizedError(store *storage.Storage, userID int64, originalFeed *model.Feed, localizedError *locale.LocalizedErrorWrapper) *locale.LocalizedErrorWrapper {
	user, storeErr := store.UserByID(userID)
	if storeErr != nil {
		return locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}
	originalFeed.WithTranslatedErrorMessage(localizedError.Translate(user.Language))
	store.UpdateFeedError(originalFeed)
	return localizedError
}

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
	subscription.IgnoreEntryUpdates = feedCreationRequest.IgnoreEntryUpdates
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

	// Web scraper feeds bypass RSS parsing — they scrape HTML/JSON pages directly.
	if feedCreationRequest.FeedSourceType == "web_scraper" {
		return createWebScraperFeed(store, userID, feedCreationRequest)
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
	subscription.IgnoreEntryUpdates = feedCreationRequest.IgnoreEntryUpdates
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
	if config.Opts.PollingScheduler() == model.SchedulerEntryFrequency {
		var weeklyCountErr error
		weeklyEntryCount, weeklyCountErr = store.WeeklyFeedEntryCount(userID, feedID)
		if weeklyCountErr != nil {
			return locale.NewLocalizedErrorWrapper(weeklyCountErr, "error.database_error", weeklyCountErr)
		}
	}

	originalFeed.CheckedNow()
	originalFeed.ScheduleNextCheck(weeklyEntryCount, time.Duration(0))

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

	// Web scraper feeds: re-scrape the page instead of parsing RSS.
	if originalFeed.FeedSourceType == "web_scraper" {
		scrapeConfig := &webscraper.ScrapeConfig{
			ItemSelector:        originalFeed.WebScraperItemSelector,
			TitleSelector:       originalFeed.WebScraperTitleSelector,
			LinkSelector:        originalFeed.WebScraperLinkSelector,
			DescriptionSelector: originalFeed.WebScraperDescSelector,
			NextPageSelector:    originalFeed.WebScraperNextPageSelector,
			MaxItems:            originalFeed.WebScraperMaxItems,
		}

		var results []*webscraper.ScrapeResult
		var scrapeErr error

		// When JS rendering is enabled, use pinchtab to render the listing page
		// first, then parse the rendered HTML with CSS selectors. This handles
		// pages where the item list is dynamically generated by JavaScript.
		if originalFeed.UseJSRender && config.Opts.PinchTabEnabled() {
			slog.Debug("Rendering web scraper listing page with pinchtab",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.String("feed_url", originalFeed.FeedURL),
			)
			renderedHTML, renderErr := pinchtab.RenderPageHTML(originalFeed.FeedURL, processor.ResolveProxyURLForPinchtab(originalFeed), originalFeed.ID)
			if renderErr != nil {
				slog.Warn("Unable to render web scraper listing page with pinchtab",
					slog.Int64("user_id", userID),
					slog.Int64("feed_id", feedID),
					slog.String("feed_url", originalFeed.FeedURL),
					slog.Any("error", renderErr),
				)
				// Don't fallback to HTTP fetch — it would hit the same network
				// issues (e.g. expired TLS certs) and mask the real pinchtab error.
				scrapeErr = fmt.Errorf("pinchtab JS rendering failed: %w", renderErr)
			} else if renderedHTML != "" {
				results, scrapeErr = webscraper.ScrapeRenderedHTML(renderedHTML, originalFeed.FeedURL, scrapeConfig)
			}
		}

		if results == nil && scrapeErr == nil {
			results, scrapeErr = webscraper.ScrapeWebPage(requestBuilder, originalFeed.FeedURL, scrapeConfig)
		}

		if scrapeErr != nil {
			slog.Warn("Unable to scrape web page for feed refresh",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.String("feed_url", originalFeed.FeedURL),
				slog.Any("error", scrapeErr),
			)
			localizedError := locale.NewLocalizedErrorWrapper(scrapeErr, "error.unable_to_parse_feed", scrapeErr)
			return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
		}

		entries := make(model.Entries, 0, len(results))
		for _, result := range results {
			entries = append(entries, &model.Entry{
				Title:   result.Title,
				URL:     result.Link,
				Content: result.Description,
				Hash:    crypto.HashFromBytes([]byte(result.Link)),
			})
		}

		originalFeed.Entries = entries
		processor.ProcessFeedEntries(store, originalFeed, userID, forceRefresh)

		updateExistingEntries := true
		newEntries, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, updateExistingEntries)
		if storeErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
		}

		userIntegrations, intErr := store.Integration(userID)
		if intErr != nil {
			slog.Error("Fetching integrations failed; the refresh process will go on, but no integrations will run this time",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.Any("error", intErr),
			)
		} else if userIntegrations != nil && len(newEntries) > 0 {
			// Load user language for AI summary generation in the user's preferred language.
			userLanguage := ""
			if user, userErr := store.UserByID(userID); userErr == nil && user != nil {
				userLanguage = user.Language
			}
			go func() {
				integration.PushEntries(originalFeed, newEntries, userIntegrations)
				integration.SummarizeEntries(store, newEntries, userIntegrations, userLanguage)
			}()
		}

		originalFeed.ResetErrorCounter()
		if storeErr := store.UpdateFeed(originalFeed); storeErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
		}

		return nil
	}
	responseHandler := fetcher.NewResponseHandler(requestBuilder.ExecuteRequest(originalFeed.FeedURL))
	defer responseHandler.Close()

	if responseHandler.IsRateLimited() {
		retryDelay := responseHandler.ParseRetryDelay()
		calculatedNextCheckInterval := originalFeed.ScheduleNextCheck(weeklyEntryCount, retryDelay)

		slog.Warn("Feed is rate limited",
			slog.String("feed_url", originalFeed.FeedURL),
			slog.Int("retry_delay_in_seconds", int(retryDelay.Seconds())),
			slog.Int("calculated_next_check_interval_in_minutes", int(calculatedNextCheckInterval.Minutes())),
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
		return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
	}

	if store.AnotherFeedURLExists(userID, originalFeed.ID, responseHandler.EffectiveURL()) {
		localizedError := locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
		return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
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
			return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
		}

		// Use the RSS TTL value, or the Cache-Control or Expires HTTP headers if available.
		// Otherwise, we use the default value from the configuration (min interval parameter).
		feedTTLValue := updatedFeed.TTL
		cacheControlMaxAgeValue := responseHandler.CacheControlMaxAge()
		expiresValue := responseHandler.Expires()
		refreshDelay := max(feedTTLValue, cacheControlMaxAgeValue, expiresValue)

		// Set the next check at with updated arguments.
		calculatedNextCheckInterval := originalFeed.ScheduleNextCheck(weeklyEntryCount, refreshDelay)

		slog.Debug("Updated next check date",
			slog.Int64("user_id", userID),
			slog.Int64("feed_id", feedID),
			slog.String("feed_url", originalFeed.FeedURL),
			slog.Int("feed_ttl_minutes", int(feedTTLValue.Minutes())),
			slog.Int("cache_control_max_age_in_minutes", int(cacheControlMaxAgeValue.Minutes())),
			slog.Int("expires_in_minutes", int(expiresValue.Minutes())),
			slog.Int("refresh_delay_in_minutes", int(refreshDelay.Minutes())),
			slog.Int("calculated_next_check_interval_in_minutes", int(calculatedNextCheckInterval.Minutes())),
			slog.Time("new_next_check_at", originalFeed.NextCheckAt),
		)

		originalFeed.Entries = updatedFeed.Entries
		processor.ProcessFeedEntries(store, originalFeed, userID, forceRefresh)

		// We don't update existing entries when the crawler is enabled (we crawl only inexisting entries).
		// We also skip updating existing entries if the feed has ignore_entry_updates enabled.
		// Unless it is forced to refresh.
		updateExistingEntries := forceRefresh || (!originalFeed.Crawler && !originalFeed.IgnoreEntryUpdates)
		newEntries, storeErr := store.RefreshFeedEntries(originalFeed.UserID, originalFeed.ID, originalFeed.Entries, updateExistingEntries)
		if storeErr != nil {
			localizedError := locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
			return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
		}

		userIntegrations, intErr := store.Integration(userID)
		if intErr != nil {
			slog.Error("Fetching integrations failed; the refresh process will go on, but no integrations will run this time",
				slog.Int64("user_id", userID),
				slog.Int64("feed_id", feedID),
				slog.Any("error", intErr),
			)
		} else if userIntegrations != nil && len(newEntries) > 0 {
			// Load user language for AI summary generation in the user's preferred language.
			userLanguage := ""
			if user, userErr := store.UserByID(userID); userErr == nil && user != nil {
				userLanguage = user.Language
			}
			go func() {
				integration.PushEntries(originalFeed, newEntries, userIntegrations)
				integration.SummarizeEntries(store, newEntries, userIntegrations, userLanguage)
			}()
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
		return getTranslatedLocalizedError(store, userID, originalFeed, localizedError)
	}

	return nil
}

// createWebScraperFeed creates a feed by scraping a web page instead of parsing RSS.
func createWebScraperFeed(store *storage.Storage, userID int64, req *model.FeedCreationRequest) (*model.Feed, *locale.LocalizedErrorWrapper) {
	if store.FeedURLExists(userID, req.FeedURL) {
		return nil, locale.NewLocalizedErrorWrapper(ErrDuplicatedFeed, "error.duplicated_feed")
	}

	requestBuilder := fetcher.NewRequestBuilder()
	requestBuilder.WithUsernameAndPassword(req.Username, req.Password)
	requestBuilder.WithUserAgent(req.UserAgent, config.Opts.HTTPClientUserAgent())
	requestBuilder.WithCookie(req.Cookie)
	requestBuilder.WithTimeout(config.Opts.HTTPClientTimeout())
	requestBuilder.WithProxyRotator(proxyrotator.ProxyRotatorInstance)
	requestBuilder.WithCustomFeedProxyURL(req.ProxyURL)
	requestBuilder.WithCustomApplicationProxyURL(config.Opts.HTTPClientProxyURL())
	requestBuilder.UseCustomApplicationProxyURL(req.FetchViaProxy)
	requestBuilder.IgnoreTLSErrors(req.AllowSelfSignedCertificates)
	requestBuilder.DisableHTTP2(req.DisableHTTP2)

	scrapeConfig := &webscraper.ScrapeConfig{
		ItemSelector:        req.WebScraperItemSelector,
		TitleSelector:       req.WebScraperTitleSelector,
		LinkSelector:        req.WebScraperLinkSelector,
		DescriptionSelector: req.WebScraperDescSelector,
		NextPageSelector:    req.WebScraperNextPageSelector,
		MaxItems:            req.WebScraperMaxItems,
	}

	results, scrapeErr := webscraper.ScrapeWebPage(requestBuilder, req.FeedURL, scrapeConfig)
	if scrapeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(scrapeErr, "error.unable_to_parse_feed", scrapeErr)
	}

	entries := make(model.Entries, 0, len(results))
	for _, result := range results {
		entries = append(entries, &model.Entry{
			Title:   result.Title,
			URL:     result.Link,
			Content: result.Description,
			Hash:    crypto.HashFromBytes([]byte(result.Link)),
		})
	}

	// Derive a feed title from the URL hostname when no RSS title is available.
	feedTitle := req.FeedURL
	if parsed, err := url.Parse(req.FeedURL); err == nil && parsed.Host != "" {
		feedTitle = parsed.Host
	}

	subscription := &model.Feed{
		UserID:                      userID,
		FeedURL:                     req.FeedURL,
		SiteURL:                     req.FeedURL,
		Title:                       feedTitle,
		UserAgent:                   req.UserAgent,
		Cookie:                      req.Cookie,
		Username:                    req.Username,
		Password:                    req.Password,
		Crawler:                     req.Crawler,
		IgnoreEntryUpdates:          req.IgnoreEntryUpdates,
		Disabled:                    req.Disabled,
		NoMediaPlayer:               req.NoMediaPlayer,
		IgnoreHTTPCache:             req.IgnoreHTTPCache,
		AllowSelfSignedCertificates: req.AllowSelfSignedCertificates,
		FetchViaProxy:               req.FetchViaProxy,
		HideGlobally:                req.HideGlobally,
		DisableHTTP2:                req.DisableHTTP2,
		ScraperRules:                req.ScraperRules,
		RewriteRules:                req.RewriteRules,
		BlocklistRules:              req.BlocklistRules,
		KeeplistRules:               req.KeeplistRules,
		UrlRewriteRules:             req.UrlRewriteRules,
		BlockFilterEntryRules:       req.BlockFilterEntryRules,
		KeepFilterEntryRules:        req.KeepFilterEntryRules,
		ProxyURL:                    req.ProxyURL,
		FeedSourceType:              req.FeedSourceType,
		WebScraperItemSelector:      req.WebScraperItemSelector,
		WebScraperTitleSelector:     req.WebScraperTitleSelector,
		WebScraperLinkSelector:      req.WebScraperLinkSelector,
		WebScraperDescSelector:      req.WebScraperDescSelector,
		WebScraperNextPageSelector:  req.WebScraperNextPageSelector,
		WebScraperMaxItems:          req.WebScraperMaxItems,
		UseJSRender:                 req.UseJSRender,
		Entries:                     entries,
	}
	subscription.WithCategoryID(req.CategoryID)
	subscription.CheckedNow()

	processor.ProcessFeedEntries(store, subscription, userID, true)

	if storeErr := store.CreateFeed(subscription); storeErr != nil {
		return nil, locale.NewLocalizedErrorWrapper(storeErr, "error.database_error", storeErr)
	}

	slog.Debug("Created web scraper feed",
		slog.Int64("user_id", userID),
		slog.Int64("feed_id", subscription.ID),
		slog.String("feed_url", subscription.FeedURL),
	)

	icon.NewIconChecker(store, subscription).UpdateOrCreateFeedIcon()

	return subscription, nil
}
