// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/storage"
	"miniflux.app/v2/internal/template"
	"miniflux.app/v2/internal/worker"
)

// Serve returns an http.Handler that serves the user interface.
// The returned handler expects the base path to be stripped from the request URL.
func Serve(store *storage.Storage, pool *worker.Pool) http.Handler {
	basePath := config.Opts.BasePath()
	webSessionMiddleware := newWebSessionMiddleware(basePath, store)
	csrfMiddleware := newCSRFMiddleware(basePath)
	authProxyMiddleware := newAuthProxyMiddleware(basePath, store)

	templateEngine := template.NewEngine(basePath)
	templateEngine.ParseTemplates()

	handler := &handler{basePath, store, templateEngine, pool}

	mux := http.NewServeMux()

	// Static assets.
	mux.HandleFunc("GET /stylesheets/{checksum}/{filename}", handler.showStylesheet)
	mux.HandleFunc("GET /js/{checksum}/{filename}", handler.showJavascript)
	mux.HandleFunc("GET /icon/{checksum}/{filename}", handler.showAppIcon)
	mux.HandleFunc("GET /favicon.ico", handler.showFavicon)
	mux.HandleFunc("GET /manifest.json", handler.showWebManifest)

	// New subscription pages.
	mux.HandleFunc("GET /subscribe", handler.showAddSubscriptionPage)
	mux.HandleFunc("POST /subscribe", handler.submitSubscription)
	mux.HandleFunc("POST /subscriptions", handler.showChooseSubscriptionPage)
	mux.HandleFunc("GET /bookmarklet", handler.bookmarklet)

	// Unread page.
	mux.HandleFunc("POST /mark-all-as-read", handler.markAllAsRead)
	mux.HandleFunc("GET /unread", handler.showUnreadPage)
	mux.HandleFunc("GET /unread/entry/{entryID}", handler.showUnreadEntryPage)

	// History pages.
	mux.HandleFunc("GET /history", handler.showHistoryPage)
	mux.HandleFunc("GET /history/entry/{entryID}", handler.showReadEntryPage)
	mux.HandleFunc("POST /history/flush", handler.flushHistory)

	// Starred pages.
	mux.HandleFunc("GET /starred", handler.showStarredPage)
	mux.HandleFunc("GET /starred/entry/{entryID}", handler.showStarredEntryPage)

	// Search pages.
	mux.HandleFunc("GET /search", handler.showSearchPage)
	mux.HandleFunc("GET /search/entry/{entryID}", handler.showSearchEntryPage)

	// Feed listing pages.
	mux.HandleFunc("GET /feeds", handler.showFeedsPage)
	mux.HandleFunc("GET /feeds/refresh", handler.refreshAllFeeds)

	// Individual feed pages.
	mux.HandleFunc("GET /feed/{feedID}/refresh", handler.refreshFeed)
	mux.HandleFunc("POST /feed/{feedID}/refresh", handler.refreshFeed)
	mux.HandleFunc("GET /feed/{feedID}/edit", handler.showEditFeedPage)
	mux.HandleFunc("POST /feed/{feedID}/remove", handler.removeFeed)
	mux.HandleFunc("POST /feed/{feedID}/update", handler.updateFeed)
	mux.HandleFunc("GET /feed/{feedID}/entries", handler.showFeedEntriesPage)
	mux.HandleFunc("GET /feed/{feedID}/entries/all", handler.showFeedEntriesAllPage)
	mux.HandleFunc("GET /feed/{feedID}/entry/{entryID}", handler.showFeedEntryPage)
	mux.HandleFunc("GET /unread/feed/{feedID}/entry/{entryID}", handler.showUnreadFeedEntryPage)
	mux.HandleFunc("POST /feed/{feedID}/mark-all-as-read", handler.markFeedAsRead)
	mux.HandleFunc("GET /feed-icon/{externalIconID}", handler.showFeedIcon)

	// Category pages.
	mux.HandleFunc("GET /category/{categoryID}/entry/{entryID}", handler.showCategoryEntryPage)
	mux.HandleFunc("GET /unread/category/{categoryID}/entry/{entryID}", handler.showUnreadCategoryEntryPage)
	mux.HandleFunc("GET /starred/category/{categoryID}/entry/{entryID}", handler.showStarredCategoryEntryPage)
	mux.HandleFunc("GET /categories", handler.showCategoryListPage)
	mux.HandleFunc("GET /category/create", handler.showCreateCategoryPage)
	mux.HandleFunc("POST /category/save", handler.saveCategory)
	mux.HandleFunc("GET /category/{categoryID}/feeds", handler.showCategoryFeedsPage)
	mux.HandleFunc("POST /category/{categoryID}/feed/{feedID}/remove", handler.removeCategoryFeed)
	mux.HandleFunc("POST /category/{categoryID}/feed/{feedID}/mark-all-as-read", handler.markCategoryFeedAsRead)
	mux.HandleFunc("GET /category/{categoryID}/feeds/refresh", handler.refreshCategoryFeedsPage)
	mux.HandleFunc("GET /category/{categoryID}/entries", handler.showCategoryEntriesPage)
	mux.HandleFunc("GET /category/{categoryID}/entries/refresh", handler.refreshCategoryEntriesPage)
	mux.HandleFunc("GET /category/{categoryID}/entries/all", handler.showCategoryEntriesAllPage)
	mux.HandleFunc("GET /category/{categoryID}/entries/starred", handler.showCategoryEntriesStarredPage)
	mux.HandleFunc("GET /category/{categoryID}/edit", handler.showEditCategoryPage)
	mux.HandleFunc("POST /category/{categoryID}/update", handler.updateCategory)
	mux.HandleFunc("POST /category/{categoryID}/remove", handler.removeCategory)
	mux.HandleFunc("POST /category/{categoryID}/mark-all-as-read", handler.markCategoryAsRead)

	// Tag pages.
	mux.HandleFunc("GET /tags/{tagName}/entries/all", handler.showTagEntriesAllPage)
	mux.HandleFunc("GET /tags/{tagName}/entry/{entryID}", handler.showTagEntryPage)

	// Entry pages.
	mux.HandleFunc("POST /entry/status", handler.updateEntriesStatus)
	mux.HandleFunc("POST /entry/save/{entryID}", handler.saveEntry)
	mux.HandleFunc("POST /entry/enclosure/{enclosureID}/save-progression", handler.saveEnclosureProgression)
	mux.HandleFunc("POST /entry/download/{entryID}", handler.fetchContent)
	mux.HandleFunc("POST /entry/star/{entryID}", handler.toggleStarred)

	// Media proxy.
	mux.HandleFunc("GET /proxy/{encodedDigest}/{encodedURL}", handler.mediaProxy)

	// Share pages.
	mux.HandleFunc("POST /entry/share/{entryID}", handler.createSharedEntry)
	mux.HandleFunc("POST /entry/unshare/{entryID}", handler.unshareEntry)
	mux.HandleFunc("GET /share/{shareCode}", handler.sharedEntry)
	mux.HandleFunc("GET /shares", handler.sharedEntries)

	// User pages.
	mux.HandleFunc("GET /users", handler.showUsersPage)
	mux.HandleFunc("GET /user/create", handler.showCreateUserPage)
	mux.HandleFunc("POST /user/save", handler.saveUser)
	mux.HandleFunc("GET /users/{userID}/edit", handler.showEditUserPage)
	mux.HandleFunc("POST /users/{userID}/update", handler.updateUser)
	mux.HandleFunc("POST /users/{userID}/remove", handler.removeUser)

	// Settings pages.
	mux.HandleFunc("GET /settings", handler.showSettingsPage)
	mux.HandleFunc("POST /settings", handler.updateSettings)
	mux.HandleFunc("GET /integrations", handler.showIntegrationPage)
	mux.HandleFunc("POST /integration", handler.updateIntegration)
	mux.HandleFunc("GET /about", handler.showAboutPage)

	// Session pages.
	mux.HandleFunc("GET /sessions", handler.showSessionsPage)
	mux.HandleFunc("POST /sessions/{sessionID}/remove", handler.removeSession)

	// API Keys pages.
	if config.Opts.HasAPI() {
		mux.HandleFunc("GET /keys", handler.showAPIKeysPage)
		mux.HandleFunc("POST /keys/{keyID}/delete", handler.deleteAPIKey)
		mux.HandleFunc("GET /keys/create", handler.showCreateAPIKeyPage)
		mux.HandleFunc("POST /keys/save", handler.saveAPIKey)
	}

	// OPML pages.
	mux.HandleFunc("GET /export", handler.exportFeeds)
	mux.HandleFunc("GET /import", handler.showImportPage)
	mux.HandleFunc("POST /upload", handler.uploadOPML)
	mux.HandleFunc("POST /fetch", handler.fetchOPML)

	// OAuth2 flow.
	if config.Opts.OAuth2Provider() != "" {
		mux.HandleFunc("GET /oauth2/{provider}/unlink", handler.oauth2Unlink)
		mux.HandleFunc("GET /oauth2/{provider}/redirect", handler.oauth2Redirect)
		mux.HandleFunc("GET /oauth2/{provider}/callback", handler.oauth2Callback)
	}

	// Offline page.
	mux.HandleFunc("GET /offline", handler.showOfflinePage)

	// Authentication pages.
	mux.HandleFunc("POST /login", handler.checkLogin)
	mux.HandleFunc("GET /logout", handler.logout)
	mux.Handle("GET /{$}", authProxyMiddleware.handle(http.HandlerFunc(handler.showLoginPage)))

	// WebAuthn flow.
	if config.Opts.WebAuthn() {
		mux.HandleFunc("GET /webauthn/register/begin", handler.beginRegistration)
		mux.HandleFunc("POST /webauthn/register/finish", handler.finishRegistration)
		mux.HandleFunc("GET /webauthn/login/begin", handler.beginLogin)
		mux.HandleFunc("POST /webauthn/login/finish", handler.finishLogin)
		mux.HandleFunc("POST /webauthn/deleteall", handler.deleteAllCredentials)
		mux.HandleFunc("POST /webauthn/{credentialHandle}/delete", handler.deleteCredential)
		mux.HandleFunc("GET /webauthn/{credentialHandle}/rename", handler.renameCredential)
		mux.HandleFunc("POST /webauthn/{credentialHandle}/save", handler.saveCredential)
	}

	// robots.txt
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	})

	// Apply middleware chain: cross-origin protection -> web session -> CSRF validation -> handlers.
	return http.NewCrossOriginProtection().Handler(webSessionMiddleware.handle(csrfMiddleware.handle(mux)))
}
