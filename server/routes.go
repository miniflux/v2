// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package server

import (
	"net/http"

	"github.com/miniflux/miniflux2/locale"
	"github.com/miniflux/miniflux2/reader/feed"
	"github.com/miniflux/miniflux2/reader/opml"
	api_controller "github.com/miniflux/miniflux2/server/api/controller"
	"github.com/miniflux/miniflux2/server/core"
	"github.com/miniflux/miniflux2/server/middleware"
	"github.com/miniflux/miniflux2/server/template"
	ui_controller "github.com/miniflux/miniflux2/server/ui/controller"
	"github.com/miniflux/miniflux2/storage"

	"github.com/gorilla/mux"
)

func getRoutes(store *storage.Storage, feedHandler *feed.Handler) *mux.Router {
	router := mux.NewRouter()
	translator := locale.Load()
	templateEngine := template.NewTemplateEngine(router, translator)

	apiController := api_controller.NewController(store, feedHandler)
	uiController := ui_controller.NewController(store, feedHandler, opml.NewHandler(store))

	apiHandler := core.NewHandler(store, router, templateEngine, translator, middleware.NewMiddlewareChain(
		middleware.NewBasicAuthMiddleware(store).Handler,
	))

	uiHandler := core.NewHandler(store, router, templateEngine, translator, middleware.NewMiddlewareChain(
		middleware.NewSessionMiddleware(store, router).Handler,
		middleware.Csrf,
	))

	router.Handle("/v1/users", apiHandler.Use(apiController.CreateUser)).Methods("POST")
	router.Handle("/v1/users", apiHandler.Use(apiController.GetUsers)).Methods("GET")
	router.Handle("/v1/users/{userID}", apiHandler.Use(apiController.GetUser)).Methods("GET")
	router.Handle("/v1/users/{userID}", apiHandler.Use(apiController.UpdateUser)).Methods("PUT")
	router.Handle("/v1/users/{userID}", apiHandler.Use(apiController.RemoveUser)).Methods("DELETE")

	router.Handle("/v1/categories", apiHandler.Use(apiController.CreateCategory)).Methods("POST")
	router.Handle("/v1/categories", apiHandler.Use(apiController.GetCategories)).Methods("GET")
	router.Handle("/v1/categories/{categoryID}", apiHandler.Use(apiController.UpdateCategory)).Methods("PUT")
	router.Handle("/v1/categories/{categoryID}", apiHandler.Use(apiController.RemoveCategory)).Methods("DELETE")

	router.Handle("/v1/discover", apiHandler.Use(apiController.GetSubscriptions)).Methods("POST")

	router.Handle("/v1/feeds", apiHandler.Use(apiController.CreateFeed)).Methods("POST")
	router.Handle("/v1/feeds", apiHandler.Use(apiController.GetFeeds)).Methods("Get")
	router.Handle("/v1/feeds/{feedID}/refresh", apiHandler.Use(apiController.RefreshFeed)).Methods("PUT")
	router.Handle("/v1/feeds/{feedID}", apiHandler.Use(apiController.GetFeed)).Methods("GET")
	router.Handle("/v1/feeds/{feedID}", apiHandler.Use(apiController.UpdateFeed)).Methods("PUT")
	router.Handle("/v1/feeds/{feedID}", apiHandler.Use(apiController.RemoveFeed)).Methods("DELETE")

	router.Handle("/v1/feeds/{feedID}/entries", apiHandler.Use(apiController.GetFeedEntries)).Methods("GET")
	router.Handle("/v1/feeds/{feedID}/entries/{entryID}", apiHandler.Use(apiController.GetEntry)).Methods("GET")
	router.Handle("/v1/feeds/{feedID}/entries/{entryID}", apiHandler.Use(apiController.SetEntryStatus)).Methods("PUT")

	router.Handle("/stylesheets/{name}.css", uiHandler.Use(uiController.Stylesheet)).Name("stylesheet").Methods("GET")
	router.Handle("/js", uiHandler.Use(uiController.Javascript)).Name("javascript").Methods("GET")
	router.Handle("/favicon.ico", uiHandler.Use(uiController.Favicon)).Name("favicon").Methods("GET")

	router.Handle("/subscribe", uiHandler.Use(uiController.AddSubscription)).Name("addSubscription").Methods("GET")
	router.Handle("/subscribe", uiHandler.Use(uiController.SubmitSubscription)).Name("submitSubscription").Methods("POST")
	router.Handle("/subscriptions", uiHandler.Use(uiController.ChooseSubscription)).Name("chooseSubscription").Methods("POST")

	router.Handle("/unread", uiHandler.Use(uiController.ShowUnreadPage)).Name("unread").Methods("GET")
	router.Handle("/history", uiHandler.Use(uiController.ShowHistoryPage)).Name("history").Methods("GET")

	router.Handle("/feed/{feedID}/refresh", uiHandler.Use(uiController.RefreshFeed)).Name("refreshFeed").Methods("GET")
	router.Handle("/feed/{feedID}/edit", uiHandler.Use(uiController.EditFeed)).Name("editFeed").Methods("GET")
	router.Handle("/feed/{feedID}/remove", uiHandler.Use(uiController.RemoveFeed)).Name("removeFeed").Methods("POST")
	router.Handle("/feed/{feedID}/update", uiHandler.Use(uiController.UpdateFeed)).Name("updateFeed").Methods("POST")
	router.Handle("/feed/{feedID}/entries", uiHandler.Use(uiController.ShowFeedEntries)).Name("feedEntries").Methods("GET")
	router.Handle("/feeds", uiHandler.Use(uiController.ShowFeedsPage)).Name("feeds").Methods("GET")

	router.Handle("/unread/entry/{entryID}", uiHandler.Use(uiController.ShowUnreadEntry)).Name("unreadEntry").Methods("GET")
	router.Handle("/history/entry/{entryID}", uiHandler.Use(uiController.ShowReadEntry)).Name("readEntry").Methods("GET")
	router.Handle("/feed/{feedID}/entry/{entryID}", uiHandler.Use(uiController.ShowFeedEntry)).Name("feedEntry").Methods("GET")
	router.Handle("/category/{categoryID}/entry/{entryID}", uiHandler.Use(uiController.ShowCategoryEntry)).Name("categoryEntry").Methods("GET")

	router.Handle("/entry/status", uiHandler.Use(uiController.UpdateEntriesStatus)).Name("updateEntriesStatus").Methods("POST")

	router.Handle("/categories", uiHandler.Use(uiController.ShowCategories)).Name("categories").Methods("GET")
	router.Handle("/category/create", uiHandler.Use(uiController.CreateCategory)).Name("createCategory").Methods("GET")
	router.Handle("/category/save", uiHandler.Use(uiController.SaveCategory)).Name("saveCategory").Methods("POST")
	router.Handle("/category/{categoryID}/entries", uiHandler.Use(uiController.ShowCategoryEntries)).Name("categoryEntries").Methods("GET")
	router.Handle("/category/{categoryID}/edit", uiHandler.Use(uiController.EditCategory)).Name("editCategory").Methods("GET")
	router.Handle("/category/{categoryID}/update", uiHandler.Use(uiController.UpdateCategory)).Name("updateCategory").Methods("POST")
	router.Handle("/category/{categoryID}/remove", uiHandler.Use(uiController.RemoveCategory)).Name("removeCategory").Methods("POST")

	router.Handle("/icon/{iconID}", uiHandler.Use(uiController.ShowIcon)).Name("icon").Methods("GET")
	router.Handle("/proxy/{encodedURL}", uiHandler.Use(uiController.ImageProxy)).Name("proxy").Methods("GET")

	router.Handle("/users", uiHandler.Use(uiController.ShowUsers)).Name("users").Methods("GET")
	router.Handle("/user/create", uiHandler.Use(uiController.CreateUser)).Name("createUser").Methods("GET")
	router.Handle("/user/save", uiHandler.Use(uiController.SaveUser)).Name("saveUser").Methods("POST")
	router.Handle("/users/{userID}/edit", uiHandler.Use(uiController.EditUser)).Name("editUser").Methods("GET")
	router.Handle("/users/{userID}/update", uiHandler.Use(uiController.UpdateUser)).Name("updateUser").Methods("POST")
	router.Handle("/users/{userID}/remove", uiHandler.Use(uiController.RemoveUser)).Name("removeUser").Methods("POST")

	router.Handle("/about", uiHandler.Use(uiController.AboutPage)).Name("about").Methods("GET")

	router.Handle("/settings", uiHandler.Use(uiController.ShowSettings)).Name("settings").Methods("GET")
	router.Handle("/settings", uiHandler.Use(uiController.UpdateSettings)).Name("updateSettings").Methods("POST")

	router.Handle("/sessions", uiHandler.Use(uiController.ShowSessions)).Name("sessions").Methods("GET")
	router.Handle("/sessions/{sessionID}/remove", uiHandler.Use(uiController.RemoveSession)).Name("removeSession").Methods("POST")

	router.Handle("/export", uiHandler.Use(uiController.Export)).Name("export").Methods("GET")
	router.Handle("/import", uiHandler.Use(uiController.Import)).Name("import").Methods("GET")
	router.Handle("/upload", uiHandler.Use(uiController.UploadOPML)).Name("uploadOPML").Methods("POST")

	router.Handle("/login", uiHandler.Use(uiController.CheckLogin)).Name("checkLogin").Methods("POST")
	router.Handle("/logout", uiHandler.Use(uiController.Logout)).Name("logout").Methods("GET")
	router.Handle("/", uiHandler.Use(uiController.ShowLoginPage)).Name("login").Methods("GET")

	router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	router.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("User-agent: *\nDisallow: /"))
	})

	return router
}
