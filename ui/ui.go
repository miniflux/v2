// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/reader/feed"
	"miniflux.app/storage"
	"miniflux.app/template"
	"miniflux.app/worker"

	"github.com/gorilla/mux"
)

// Serve declares all routes for the user interface.
func Serve(router *mux.Router, store *storage.Storage, pool *worker.Pool, feedHandler *feed.Handler) {
	middleware := newMiddleware(router, store)
	handler := &handler{router, store, template.NewEngine(router), pool, feedHandler}

	uiRouter := router.NewRoute().Subrouter()
	uiRouter.Use(middleware.handleUserSession)
	uiRouter.Use(middleware.handleAppSession)

	// Static assets.
	uiRouter.HandleFunc("/stylesheets/{name}.css", handler.showStylesheet).Name("stylesheet").Methods("GET")
	uiRouter.HandleFunc("/{name}.js", handler.showJavascript).Name("javascript").Methods("GET")
	uiRouter.HandleFunc("/favicon.ico", handler.showFavicon).Name("favicon").Methods("GET")
	uiRouter.HandleFunc("/icon/{filename}", handler.showAppIcon).Name("appIcon").Methods("GET")
	uiRouter.HandleFunc("/manifest.json", handler.showWebManifest).Name("webManifest").Methods("GET")

	// New subscription pages.
	uiRouter.HandleFunc("/subscribe", handler.showAddSubscriptionPage).Name("addSubscription").Methods("GET")
	uiRouter.HandleFunc("/subscribe", handler.submitSubscription).Name("submitSubscription").Methods("POST")
	uiRouter.HandleFunc("/subscriptions", handler.showChooseSubscriptionPage).Name("chooseSubscription").Methods("POST")
	uiRouter.HandleFunc("/bookmarklet", handler.bookmarklet).Name("bookmarklet").Methods("GET")

	// Unread page.
	uiRouter.HandleFunc("/mark-all-as-read", handler.markAllAsRead).Name("markAllAsRead").Methods("POST")
	uiRouter.HandleFunc("/unread", handler.showUnreadPage).Name("unread").Methods("GET")
	uiRouter.HandleFunc("/unread/entry/{entryID}", handler.showUnreadEntryPage).Name("unreadEntry").Methods("GET")

	// History pages.
	uiRouter.HandleFunc("/history", handler.showHistoryPage).Name("history").Methods("GET")
	uiRouter.HandleFunc("/history/entry/{entryID}", handler.showReadEntryPage).Name("readEntry").Methods("GET")
	uiRouter.HandleFunc("/history/flush", handler.flushHistory).Name("flushHistory").Methods("POST")

	// Bookmark pages.
	uiRouter.HandleFunc("/starred", handler.showStarredPage).Name("starred").Methods("GET")
	uiRouter.HandleFunc("/starred/entry/{entryID}", handler.showStarredEntryPage).Name("starredEntry").Methods("GET")

	// Search pages.
	uiRouter.HandleFunc("/search", handler.showSearchEntriesPage).Name("searchEntries").Methods("GET")
	uiRouter.HandleFunc("/search/entry/{entryID}", handler.showSearchEntryPage).Name("searchEntry").Methods("GET")

	// Feed listing pages.
	uiRouter.HandleFunc("/feeds", handler.showFeedsPage).Name("feeds").Methods("GET")
	uiRouter.HandleFunc("/feeds/refresh", handler.refreshAllFeeds).Name("refreshAllFeeds").Methods("GET")

	// Individual feed pages.
	uiRouter.HandleFunc("/feed/{feedID}/refresh", handler.refreshFeed).Name("refreshFeed").Methods("GET")
	uiRouter.HandleFunc("/feed/{feedID}/edit", handler.showEditFeedPage).Name("editFeed").Methods("GET")
	uiRouter.HandleFunc("/feed/{feedID}/remove", handler.removeFeed).Name("removeFeed").Methods("POST")
	uiRouter.HandleFunc("/feed/{feedID}/update", handler.updateFeed).Name("updateFeed").Methods("POST")
	uiRouter.HandleFunc("/feed/{feedID}/entries", handler.showFeedEntriesPage).Name("feedEntries").Methods("GET")
	uiRouter.HandleFunc("/feed/{feedID}/entries/all", handler.showFeedEntriesAllPage).Name("feedEntriesAll").Methods("GET")
	uiRouter.HandleFunc("/feed/{feedID}/entry/{entryID}", handler.showFeedEntryPage).Name("feedEntry").Methods("GET")
	uiRouter.HandleFunc("/feed/icon/{iconID}", handler.showIcon).Name("icon").Methods("GET")

	// Category pages.
	uiRouter.HandleFunc("/category/{categoryID}/entry/{entryID}", handler.showCategoryEntryPage).Name("categoryEntry").Methods("GET")
	uiRouter.HandleFunc("/categories", handler.showCategoryListPage).Name("categories").Methods("GET")
	uiRouter.HandleFunc("/category/create", handler.showCreateCategoryPage).Name("createCategory").Methods("GET")
	uiRouter.HandleFunc("/category/save", handler.saveCategory).Name("saveCategory").Methods("POST")
	uiRouter.HandleFunc("/category/{categoryID}/entries", handler.showCategoryEntriesPage).Name("categoryEntries").Methods("GET")
	uiRouter.HandleFunc("/category/{categoryID}/entries/all", handler.showCategoryEntriesAllPage).Name("categoryEntriesAll").Methods("GET")
	uiRouter.HandleFunc("/category/{categoryID}/edit", handler.showEditCategoryPage).Name("editCategory").Methods("GET")
	uiRouter.HandleFunc("/category/{categoryID}/update", handler.updateCategory).Name("updateCategory").Methods("POST")
	uiRouter.HandleFunc("/category/{categoryID}/remove", handler.removeCategory).Name("removeCategory").Methods("POST")

	// Entry pages.
	uiRouter.HandleFunc("/entry/status", handler.updateEntriesStatus).Name("updateEntriesStatus").Methods("POST")
	uiRouter.HandleFunc("/entry/save/{entryID}", handler.saveEntry).Name("saveEntry").Methods("POST")
	uiRouter.HandleFunc("/entry/download/{entryID}", handler.fetchContent).Name("fetchContent").Methods("POST")
	uiRouter.HandleFunc("/proxy/{encodedURL}", handler.imageProxy).Name("proxy").Methods("GET")
	uiRouter.HandleFunc("/entry/bookmark/{entryID}", handler.toggleBookmark).Name("toggleBookmark").Methods("POST")

	// User pages.
	uiRouter.HandleFunc("/users", handler.showUsersPage).Name("users").Methods("GET")
	uiRouter.HandleFunc("/user/create", handler.showCreateUserPage).Name("createUser").Methods("GET")
	uiRouter.HandleFunc("/user/save", handler.saveUser).Name("saveUser").Methods("POST")
	uiRouter.HandleFunc("/users/{userID}/edit", handler.showEditUserPage).Name("editUser").Methods("GET")
	uiRouter.HandleFunc("/users/{userID}/update", handler.updateUser).Name("updateUser").Methods("POST")
	uiRouter.HandleFunc("/users/{userID}/remove", handler.removeUser).Name("removeUser").Methods("POST")

	// Settings pages.
	uiRouter.HandleFunc("/settings", handler.showSettingsPage).Name("settings").Methods("GET")
	uiRouter.HandleFunc("/settings", handler.updateSettings).Name("updateSettings").Methods("POST")
	uiRouter.HandleFunc("/integrations", handler.showIntegrationPage).Name("integrations").Methods("GET")
	uiRouter.HandleFunc("/integration", handler.updateIntegration).Name("updateIntegration").Methods("POST")
	uiRouter.HandleFunc("/integration/pocket/authorize", handler.pocketAuthorize).Name("pocketAuthorize").Methods("GET")
	uiRouter.HandleFunc("/integration/pocket/callback", handler.pocketCallback).Name("pocketCallback").Methods("GET")
	uiRouter.HandleFunc("/about", handler.showAboutPage).Name("about").Methods("GET")

	// Session pages.
	uiRouter.HandleFunc("/sessions", handler.showSessionsPage).Name("sessions").Methods("GET")
	uiRouter.HandleFunc("/sessions/{sessionID}/remove", handler.removeSession).Name("removeSession").Methods("POST")

	// OPML pages.
	uiRouter.HandleFunc("/export", handler.exportFeeds).Name("export").Methods("GET")
	uiRouter.HandleFunc("/import", handler.showImportPage).Name("import").Methods("GET")
	uiRouter.HandleFunc("/upload", handler.uploadOPML).Name("uploadOPML").Methods("POST")
	uiRouter.HandleFunc("/fetch", handler.fetchOPML).Name("fetchOPML").Methods("POST")

	// OAuth2 flow.
	uiRouter.HandleFunc("/oauth2/{provider}/unlink", handler.oauth2Unlink).Name("oauth2Unlink").Methods("GET")
	uiRouter.HandleFunc("/oauth2/{provider}/redirect", handler.oauth2Redirect).Name("oauth2Redirect").Methods("GET")
	uiRouter.HandleFunc("/oauth2/{provider}/callback", handler.oauth2Callback).Name("oauth2Callback").Methods("GET")

	// Authentication pages.
	uiRouter.HandleFunc("/login", handler.checkLogin).Name("checkLogin").Methods("POST")
	uiRouter.HandleFunc("/logout", handler.logout).Name("logout").Methods("GET")
	uiRouter.HandleFunc("/", handler.showLoginPage).Name("login").Methods("GET")

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	}).Name("robots")
}
