// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package daemon

import (
	"net/http"

	"github.com/miniflux/miniflux/api"
	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/fever"
	"github.com/miniflux/miniflux/http/handler"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/middleware"
	"github.com/miniflux/miniflux/reader/feed"
	"github.com/miniflux/miniflux/scheduler"
	"github.com/miniflux/miniflux/storage"
	"github.com/miniflux/miniflux/template"
	"github.com/miniflux/miniflux/ui"

	"github.com/gorilla/mux"
)

func routes(cfg *config.Config, store *storage.Storage, feedHandler *feed.Handler, pool *scheduler.WorkerPool, translator *locale.Translator) *mux.Router {
	router := mux.NewRouter()
	templateEngine := template.NewEngine(cfg, router, translator)

	apiController := api.NewController(store, feedHandler)
	feverController := fever.NewController(store)
	uiController := ui.NewController(cfg, store, pool, feedHandler)

	apiHandler := handler.NewHandler(cfg, store, router, templateEngine, translator)
	feverHandler := handler.NewHandler(cfg, store, router, templateEngine, translator)
	uiHandler := handler.NewHandler(cfg, store, router, templateEngine, translator)
	middleware := middleware.New(cfg, store, router)

	if cfg.BasePath() != "" {
		router = router.PathPrefix(cfg.BasePath()).Subrouter()
	}

	router.Use(middleware.HeaderConfig)
	router.Use(middleware.Logging)

	router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	})

	feverRouter := router.PathPrefix("/fever").Subrouter()
	feverRouter.Use(middleware.FeverAuth)
	feverRouter.Handle("/", feverHandler.Use(feverController.Handler)).Name("feverEndpoint")

	apiRouter := router.PathPrefix("/v1").Subrouter()
	apiRouter.Use(middleware.BasicAuth)
	apiRouter.Handle("/users", apiHandler.Use(apiController.CreateUser)).Methods("POST")
	apiRouter.Handle("/users", apiHandler.Use(apiController.Users)).Methods("GET")
	apiRouter.Handle("/users/{userID:[0-9]+}", apiHandler.Use(apiController.UserByID)).Methods("GET")
	apiRouter.Handle("/users/{userID:[0-9]+}", apiHandler.Use(apiController.UpdateUser)).Methods("PUT")
	apiRouter.Handle("/users/{userID:[0-9]+}", apiHandler.Use(apiController.RemoveUser)).Methods("DELETE")
	apiRouter.Handle("/users/{username}", apiHandler.Use(apiController.UserByUsername)).Methods("GET")
	apiRouter.Handle("/categories", apiHandler.Use(apiController.CreateCategory)).Methods("POST")
	apiRouter.Handle("/categories", apiHandler.Use(apiController.GetCategories)).Methods("GET")
	apiRouter.Handle("/categories/{categoryID}", apiHandler.Use(apiController.UpdateCategory)).Methods("PUT")
	apiRouter.Handle("/categories/{categoryID}", apiHandler.Use(apiController.RemoveCategory)).Methods("DELETE")
	apiRouter.Handle("/discover", apiHandler.Use(apiController.GetSubscriptions)).Methods("POST")
	apiRouter.Handle("/feeds", apiHandler.Use(apiController.CreateFeed)).Methods("POST")
	apiRouter.Handle("/feeds", apiHandler.Use(apiController.GetFeeds)).Methods("Get")
	apiRouter.Handle("/feeds/{feedID}/refresh", apiHandler.Use(apiController.RefreshFeed)).Methods("PUT")
	apiRouter.Handle("/feeds/{feedID}", apiHandler.Use(apiController.GetFeed)).Methods("GET")
	apiRouter.Handle("/feeds/{feedID}", apiHandler.Use(apiController.UpdateFeed)).Methods("PUT")
	apiRouter.Handle("/feeds/{feedID}", apiHandler.Use(apiController.RemoveFeed)).Methods("DELETE")
	apiRouter.Handle("/feeds/{feedID}/icon", apiHandler.Use(apiController.FeedIcon)).Methods("GET")
	apiRouter.Handle("/export", apiHandler.Use(apiController.Export)).Methods("GET")
	apiRouter.Handle("/feeds/{feedID}/entries", apiHandler.Use(apiController.GetFeedEntries)).Methods("GET")
	apiRouter.Handle("/feeds/{feedID}/entries/{entryID}", apiHandler.Use(apiController.GetFeedEntry)).Methods("GET")
	apiRouter.Handle("/entries", apiHandler.Use(apiController.GetEntries)).Methods("GET")
	apiRouter.Handle("/entries", apiHandler.Use(apiController.SetEntryStatus)).Methods("PUT")
	apiRouter.Handle("/entries/{entryID}", apiHandler.Use(apiController.GetEntry)).Methods("GET")
	apiRouter.Handle("/entries/{entryID}/bookmark", apiHandler.Use(apiController.ToggleBookmark)).Methods("PUT")

	uiRouter := router.NewRoute().Subrouter()
	uiRouter.Use(middleware.AppSession)
	uiRouter.Use(middleware.UserSession)
	uiRouter.Handle("/stylesheets/{name}.css", uiHandler.Use(uiController.Stylesheet)).Name("stylesheet").Methods("GET")
	uiRouter.Handle("/js", uiHandler.Use(uiController.Javascript)).Name("javascript").Methods("GET")
	uiRouter.Handle("/favicon.ico", uiHandler.Use(uiController.Favicon)).Name("favicon").Methods("GET")
	uiRouter.Handle("/icon/{filename}", uiHandler.Use(uiController.AppIcon)).Name("appIcon").Methods("GET")
	uiRouter.Handle("/manifest.json", uiHandler.Use(uiController.WebManifest)).Name("webManifest").Methods("GET")
	uiRouter.Handle("/subscribe", uiHandler.Use(uiController.AddSubscription)).Name("addSubscription").Methods("GET")
	uiRouter.Handle("/subscribe", uiHandler.Use(uiController.SubmitSubscription)).Name("submitSubscription").Methods("POST")
	uiRouter.Handle("/subscriptions", uiHandler.Use(uiController.ChooseSubscription)).Name("chooseSubscription").Methods("POST")
	uiRouter.Handle("/mark-all-as-read", uiHandler.Use(uiController.MarkAllAsRead)).Name("markAllAsRead").Methods("GET")
	uiRouter.Handle("/unread", uiHandler.Use(uiController.ShowUnreadPage)).Name("unread").Methods("GET")
	uiRouter.Handle("/history", uiHandler.Use(uiController.ShowHistoryPage)).Name("history").Methods("GET")
	uiRouter.Handle("/starred", uiHandler.Use(uiController.ShowStarredPage)).Name("starred").Methods("GET")
	uiRouter.Handle("/feed/{feedID}/refresh", uiHandler.Use(uiController.RefreshFeed)).Name("refreshFeed").Methods("GET")
	uiRouter.Handle("/feed/{feedID}/edit", uiHandler.Use(uiController.EditFeed)).Name("editFeed").Methods("GET")
	uiRouter.Handle("/feed/{feedID}/remove", uiHandler.Use(uiController.RemoveFeed)).Name("removeFeed").Methods("POST")
	uiRouter.Handle("/feed/{feedID}/update", uiHandler.Use(uiController.UpdateFeed)).Name("updateFeed").Methods("POST")
	uiRouter.Handle("/feed/{feedID}/entries", uiHandler.Use(uiController.ShowFeedEntries)).Name("feedEntries").Methods("GET")
	uiRouter.Handle("/feeds", uiHandler.Use(uiController.ShowFeedsPage)).Name("feeds").Methods("GET")
	uiRouter.Handle("/feeds/refresh", uiHandler.Use(uiController.RefreshAllFeeds)).Name("refreshAllFeeds").Methods("GET")
	uiRouter.Handle("/unread/entry/{entryID}", uiHandler.Use(uiController.ShowUnreadEntry)).Name("unreadEntry").Methods("GET")
	uiRouter.Handle("/history/entry/{entryID}", uiHandler.Use(uiController.ShowReadEntry)).Name("readEntry").Methods("GET")
	uiRouter.Handle("/history/flush", uiHandler.Use(uiController.FlushHistory)).Name("flushHistory").Methods("GET")
	uiRouter.Handle("/feed/{feedID}/entry/{entryID}", uiHandler.Use(uiController.ShowFeedEntry)).Name("feedEntry").Methods("GET")
	uiRouter.Handle("/category/{categoryID}/entry/{entryID}", uiHandler.Use(uiController.ShowCategoryEntry)).Name("categoryEntry").Methods("GET")
	uiRouter.Handle("/starred/entry/{entryID}", uiHandler.Use(uiController.ShowStarredEntry)).Name("starredEntry").Methods("GET")
	uiRouter.Handle("/entry/status", uiHandler.Use(uiController.UpdateEntriesStatus)).Name("updateEntriesStatus").Methods("POST")
	uiRouter.Handle("/entry/save/{entryID}", uiHandler.Use(uiController.SaveEntry)).Name("saveEntry").Methods("POST")
	uiRouter.Handle("/entry/download/{entryID}", uiHandler.Use(uiController.FetchContent)).Name("fetchContent").Methods("POST")
	uiRouter.Handle("/entry/bookmark/{entryID}", uiHandler.Use(uiController.ToggleBookmark)).Name("toggleBookmark").Methods("POST")
	uiRouter.Handle("/categories", uiHandler.Use(uiController.ShowCategories)).Name("categories").Methods("GET")
	uiRouter.Handle("/category/create", uiHandler.Use(uiController.CreateCategory)).Name("createCategory").Methods("GET")
	uiRouter.Handle("/category/save", uiHandler.Use(uiController.SaveCategory)).Name("saveCategory").Methods("POST")
	uiRouter.Handle("/category/{categoryID}/entries", uiHandler.Use(uiController.ShowCategoryEntries)).Name("categoryEntries").Methods("GET")
	uiRouter.Handle("/category/{categoryID}/edit", uiHandler.Use(uiController.EditCategory)).Name("editCategory").Methods("GET")
	uiRouter.Handle("/category/{categoryID}/update", uiHandler.Use(uiController.UpdateCategory)).Name("updateCategory").Methods("POST")
	uiRouter.Handle("/category/{categoryID}/remove", uiHandler.Use(uiController.RemoveCategory)).Name("removeCategory").Methods("POST")
	uiRouter.Handle("/feed/icon/{iconID}", uiHandler.Use(uiController.ShowIcon)).Name("icon").Methods("GET")
	uiRouter.Handle("/proxy/{encodedURL}", uiHandler.Use(uiController.ImageProxy)).Name("proxy").Methods("GET")
	uiRouter.Handle("/users", uiHandler.Use(uiController.ShowUsers)).Name("users").Methods("GET")
	uiRouter.Handle("/user/create", uiHandler.Use(uiController.CreateUser)).Name("createUser").Methods("GET")
	uiRouter.Handle("/user/save", uiHandler.Use(uiController.SaveUser)).Name("saveUser").Methods("POST")
	uiRouter.Handle("/users/{userID}/edit", uiHandler.Use(uiController.EditUser)).Name("editUser").Methods("GET")
	uiRouter.Handle("/users/{userID}/update", uiHandler.Use(uiController.UpdateUser)).Name("updateUser").Methods("POST")
	uiRouter.Handle("/users/{userID}/remove", uiHandler.Use(uiController.RemoveUser)).Name("removeUser").Methods("POST")
	uiRouter.Handle("/about", uiHandler.Use(uiController.AboutPage)).Name("about").Methods("GET")
	uiRouter.Handle("/settings", uiHandler.Use(uiController.ShowSettings)).Name("settings").Methods("GET")
	uiRouter.Handle("/settings", uiHandler.Use(uiController.UpdateSettings)).Name("updateSettings").Methods("POST")
	uiRouter.Handle("/bookmarklet", uiHandler.Use(uiController.Bookmarklet)).Name("bookmarklet").Methods("GET")
	uiRouter.Handle("/integrations", uiHandler.Use(uiController.ShowIntegrations)).Name("integrations").Methods("GET")
	uiRouter.Handle("/integration", uiHandler.Use(uiController.UpdateIntegration)).Name("updateIntegration").Methods("POST")
	uiRouter.Handle("/sessions", uiHandler.Use(uiController.ShowSessions)).Name("sessions").Methods("GET")
	uiRouter.Handle("/sessions/{sessionID}/remove", uiHandler.Use(uiController.RemoveSession)).Name("removeSession").Methods("POST")
	uiRouter.Handle("/export", uiHandler.Use(uiController.Export)).Name("export").Methods("GET")
	uiRouter.Handle("/import", uiHandler.Use(uiController.Import)).Name("import").Methods("GET")
	uiRouter.Handle("/upload", uiHandler.Use(uiController.UploadOPML)).Name("uploadOPML").Methods("POST")
	uiRouter.Handle("/oauth2/{provider}/unlink", uiHandler.Use(uiController.OAuth2Unlink)).Name("oauth2Unlink").Methods("GET")
	uiRouter.Handle("/oauth2/{provider}/redirect", uiHandler.Use(uiController.OAuth2Redirect)).Name("oauth2Redirect").Methods("GET")
	uiRouter.Handle("/oauth2/{provider}/callback", uiHandler.Use(uiController.OAuth2Callback)).Name("oauth2Callback").Methods("GET")
	uiRouter.Handle("/login", uiHandler.Use(uiController.CheckLogin)).Name("checkLogin").Methods("POST")
	uiRouter.Handle("/logout", uiHandler.Use(uiController.Logout)).Name("logout").Methods("GET")
	uiRouter.Handle("/", uiHandler.Use(uiController.ShowLoginPage)).Name("login").Methods("GET")

	return router
}
