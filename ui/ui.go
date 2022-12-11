// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/logger"
	"miniflux.app/storage"
	"miniflux.app/template"
	"miniflux.app/worker"

	"github.com/gorilla/mux"
)

// Serve declares all routes for the user interface.
func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool) {
	middleware := newMiddleware(router, store)

	templateEngine := template.NewEngine(router)
	if err := templateEngine.ParseTemplates(); err != nil {
		logger.Fatal(`Unable to parse templates: %v`, err)
	}

	handler := &handler{router, store, templateEngine, pool}

	uiRouter := router.NewRoute().Subrouter()
	uiRouter.Use(middleware.handleUserSession)
	uiRouter.Use(middleware.handleAppSession)
	uiRouter.StrictSlash(true)

	// Static assets.
	uiRouter.HandleFunc("/stylesheets/{name}.{checksum}.css", handler.showStylesheet).Name("stylesheet").Methods(http.MethodGet)
	uiRouter.HandleFunc("/{name}.{checksum}.js", handler.showJavascript).Name("javascript").Methods(http.MethodGet)
	uiRouter.HandleFunc("/favicon.ico", handler.showFavicon).Name("favicon").Methods(http.MethodGet)
	uiRouter.HandleFunc("/icon/{filename}", handler.showAppIcon).Name("appIcon").Methods(http.MethodGet)
	uiRouter.HandleFunc("/manifest.json", handler.showWebManifest).Name("webManifest").Methods(http.MethodGet)

	// New subscription pages.
	uiRouter.HandleFunc("/subscribe", handler.showAddSubscriptionPage).Name("addSubscription").Methods(http.MethodGet)
	uiRouter.HandleFunc("/subscribe", handler.submitSubscription).Name("submitSubscription").Methods(http.MethodPost)
	uiRouter.HandleFunc("/subscriptions", handler.showChooseSubscriptionPage).Name("chooseSubscription").Methods(http.MethodPost)
	uiRouter.HandleFunc("/bookmarklet", handler.bookmarklet).Name("bookmarklet").Methods(http.MethodGet)

	// Unread page.
	uiRouter.HandleFunc("/mark-all-as-read", handler.markAllAsRead).Name("markAllAsRead").Methods(http.MethodPost)
	uiRouter.HandleFunc("/unread", handler.showUnreadPage).Name("unread").Methods(http.MethodGet)
	uiRouter.HandleFunc("/unread/entry/{entryID}", handler.showUnreadEntryPage).Name("unreadEntry").Methods(http.MethodGet)

	// History pages.
	uiRouter.HandleFunc("/history", handler.showHistoryPage).Name("history").Methods(http.MethodGet)
	uiRouter.HandleFunc("/history/entry/{entryID}", handler.showReadEntryPage).Name("readEntry").Methods(http.MethodGet)
	uiRouter.HandleFunc("/history/flush", handler.flushHistory).Name("flushHistory").Methods(http.MethodPost)

	// Bookmark pages.
	uiRouter.HandleFunc("/starred", handler.showStarredPage).Name("starred").Methods(http.MethodGet)
	uiRouter.HandleFunc("/starred/entry/{entryID}", handler.showStarredEntryPage).Name("starredEntry").Methods(http.MethodGet)

	// Search pages.
	uiRouter.HandleFunc("/search", handler.showSearchEntriesPage).Name("searchEntries").Methods(http.MethodGet)
	uiRouter.HandleFunc("/search/entry/{entryID}", handler.showSearchEntryPage).Name("searchEntry").Methods(http.MethodGet)

	// Feed listing pages.
	uiRouter.HandleFunc("/feeds", handler.showFeedsPage).Name("feeds").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feeds/refresh", handler.refreshAllFeeds).Name("refreshAllFeeds").Methods(http.MethodGet)

	// Individual feed pages.
	uiRouter.HandleFunc("/feed/{feedID}/refresh", handler.refreshFeed).Name("refreshFeed").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/{feedID}/edit", handler.showEditFeedPage).Name("editFeed").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/{feedID}/remove", handler.removeFeed).Name("removeFeed").Methods(http.MethodPost)
	uiRouter.HandleFunc("/feed/{feedID}/update", handler.updateFeed).Name("updateFeed").Methods(http.MethodPost)
	uiRouter.HandleFunc("/feed/{feedID}/entries", handler.showFeedEntriesPage).Name("feedEntries").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/{feedID}/entries/all", handler.showFeedEntriesAllPage).Name("feedEntriesAll").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/{feedID}/entry/{entryID}", handler.showFeedEntryPage).Name("feedEntry").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/icon/{iconID}", handler.showIcon).Name("icon").Methods(http.MethodGet)
	uiRouter.HandleFunc("/feed/{feedID}/mark-all-as-read", handler.markFeedAsRead).Name("markFeedAsRead").Methods(http.MethodPost)

	// Category pages.
	uiRouter.HandleFunc("/category/{categoryID}/entry/{entryID}", handler.showCategoryEntryPage).Name("categoryEntry").Methods(http.MethodGet)
	uiRouter.HandleFunc("/categories", handler.showCategoryListPage).Name("categories").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/create", handler.showCreateCategoryPage).Name("createCategory").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/save", handler.saveCategory).Name("saveCategory").Methods(http.MethodPost)
	uiRouter.HandleFunc("/category/{categoryID}/feeds", handler.showCategoryFeedsPage).Name("categoryFeeds").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/feeds/refresh", handler.refreshCategoryFeedsPage).Name("refreshCategoryFeedsPage").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/entries", handler.showCategoryEntriesPage).Name("categoryEntries").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/entries/refresh", handler.refreshCategoryEntriesPage).Name("refreshCategoryEntriesPage").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/entries/all", handler.showCategoryEntriesAllPage).Name("categoryEntriesAll").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/edit", handler.showEditCategoryPage).Name("editCategory").Methods(http.MethodGet)
	uiRouter.HandleFunc("/category/{categoryID}/update", handler.updateCategory).Name("updateCategory").Methods(http.MethodPost)
	uiRouter.HandleFunc("/category/{categoryID}/remove", handler.removeCategory).Name("removeCategory").Methods(http.MethodPost)
	uiRouter.HandleFunc("/category/{categoryID}/mark-all-as-read", handler.markCategoryAsRead).Name("markCategoryAsRead").Methods(http.MethodPost)

	// Entry pages.
	uiRouter.HandleFunc("/entry/status", handler.updateEntriesStatus).Name("updateEntriesStatus").Methods(http.MethodPost)
	uiRouter.HandleFunc("/entry/save/{entryID}", handler.saveEntry).Name("saveEntry").Methods(http.MethodPost)
	uiRouter.HandleFunc("/entry/download/{entryID}", handler.fetchContent).Name("fetchContent").Methods(http.MethodPost)
	uiRouter.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", handler.imageProxy).Name("proxy").Methods(http.MethodGet)
	uiRouter.HandleFunc("/entry/bookmark/{entryID}", handler.toggleBookmark).Name("toggleBookmark").Methods(http.MethodPost)

	// Share pages.
	uiRouter.HandleFunc("/entry/share/{entryID}", handler.createSharedEntry).Name("shareEntry").Methods(http.MethodGet)
	uiRouter.HandleFunc("/entry/unshare/{entryID}", handler.unshareEntry).Name("unshareEntry").Methods(http.MethodPost)
	uiRouter.HandleFunc("/share/{shareCode}", handler.sharedEntry).Name("sharedEntry").Methods(http.MethodGet)
	uiRouter.HandleFunc("/shares", handler.sharedEntries).Name("sharedEntries").Methods(http.MethodGet)

	// User pages.
	uiRouter.HandleFunc("/users", handler.showUsersPage).Name("users").Methods(http.MethodGet)
	uiRouter.HandleFunc("/user/create", handler.showCreateUserPage).Name("createUser").Methods(http.MethodGet)
	uiRouter.HandleFunc("/user/save", handler.saveUser).Name("saveUser").Methods(http.MethodPost)
	uiRouter.HandleFunc("/users/{userID}/edit", handler.showEditUserPage).Name("editUser").Methods(http.MethodGet)
	uiRouter.HandleFunc("/users/{userID}/update", handler.updateUser).Name("updateUser").Methods(http.MethodPost)
	uiRouter.HandleFunc("/users/{userID}/remove", handler.removeUser).Name("removeUser").Methods(http.MethodPost)

	// Settings pages.
	uiRouter.HandleFunc("/settings", handler.showSettingsPage).Name("settings").Methods(http.MethodGet)
	uiRouter.HandleFunc("/settings", handler.updateSettings).Name("updateSettings").Methods(http.MethodPost)
	uiRouter.HandleFunc("/integrations", handler.showIntegrationPage).Name("integrations").Methods(http.MethodGet)
	uiRouter.HandleFunc("/integration", handler.updateIntegration).Name("updateIntegration").Methods(http.MethodPost)
	uiRouter.HandleFunc("/integration/pocket/authorize", handler.pocketAuthorize).Name("pocketAuthorize").Methods(http.MethodGet)
	uiRouter.HandleFunc("/integration/pocket/callback", handler.pocketCallback).Name("pocketCallback").Methods(http.MethodGet)
	uiRouter.HandleFunc("/about", handler.showAboutPage).Name("about").Methods(http.MethodGet)

	// Session pages.
	uiRouter.HandleFunc("/sessions", handler.showSessionsPage).Name("sessions").Methods(http.MethodGet)
	uiRouter.HandleFunc("/sessions/{sessionID}/remove", handler.removeSession).Name("removeSession").Methods(http.MethodPost)

	// API Keys pages.
	uiRouter.HandleFunc("/keys", handler.showAPIKeysPage).Name("apiKeys").Methods(http.MethodGet)
	uiRouter.HandleFunc("/keys/{keyID}/remove", handler.removeAPIKey).Name("removeAPIKey").Methods(http.MethodPost)
	uiRouter.HandleFunc("/keys/create", handler.showCreateAPIKeyPage).Name("createAPIKey").Methods(http.MethodGet)
	uiRouter.HandleFunc("/keys/save", handler.saveAPIKey).Name("saveAPIKey").Methods(http.MethodPost)

	// OPML pages.
	uiRouter.HandleFunc("/export", handler.exportFeeds).Name("export").Methods(http.MethodGet)
	uiRouter.HandleFunc("/import", handler.showImportPage).Name("import").Methods(http.MethodGet)
	uiRouter.HandleFunc("/upload", handler.uploadOPML).Name("uploadOPML").Methods(http.MethodPost)
	uiRouter.HandleFunc("/fetch", handler.fetchOPML).Name("fetchOPML").Methods(http.MethodPost)

	// OAuth2 flow.
	uiRouter.HandleFunc("/oauth2/{provider}/unlink", handler.oauth2Unlink).Name("oauth2Unlink").Methods(http.MethodGet)
	uiRouter.HandleFunc("/oauth2/{provider}/redirect", handler.oauth2Redirect).Name("oauth2Redirect").Methods(http.MethodGet)
	uiRouter.HandleFunc("/oauth2/{provider}/callback", handler.oauth2Callback).Name("oauth2Callback").Methods(http.MethodGet)

	// Offline page
	uiRouter.HandleFunc("/offline", handler.showOfflinePage).Name("offline").Methods(http.MethodGet)

	// Authentication pages.
	uiRouter.HandleFunc("/login", handler.checkLogin).Name("checkLogin").Methods(http.MethodPost)
	uiRouter.HandleFunc("/logout", handler.logout).Name("logout").Methods(http.MethodGet)
	uiRouter.Handle("/", middleware.handleAuthProxy(http.HandlerFunc(handler.showLoginPage))).Name("login").Methods(http.MethodGet)

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	}).Name("robots")
}
