package googlereader // import "miniflux.app/googlereader"

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
	"miniflux.app/storage"
)

type handler struct {
	store *storage.Storage
}

// Serve handles Google Reader API calls.
func Serve(router *mux.Router, store *storage.Storage) {
	handler := &handler{store}
	middleware := newMiddleware(store)
	router.HandleFunc("/accounts/ClientLogin", middleware.clientLogin).Methods(http.MethodPost).Name("ClientLogin")
	sr := router.PathPrefix("/reader/api/0").Subrouter()
	sr.Use(middleware.handleCORS)
	sr.Use(middleware.apiKeyAuth)
	sr.Methods(http.MethodOptions)
	sr.HandleFunc("/token", middleware.token).Methods(http.MethodGet).Name("Token")

	sr.HandleFunc("/user-info", handler.userInfo).Methods(http.MethodGet).Name("UserInfo")
	sr.HandleFunc("/subscription/list", handler.subscriptionList).Methods(http.MethodGet).Name("SubscriptonList")
	sr.PathPrefix("/").HandlerFunc(handler.serve).Methods(http.MethodPost, http.MethodGet).Name("GoogleReaderApiEndpoint")
}
func (h *handler) subscriptionList(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	clientIP := request.ClientIP(r)

	logger.Info("[Reader][subscription/list][ClientIP=%s] Incoming Request for userID  #%d", clientIP, userID)

	var result subscriptionsResponse
	feeds, err := h.store.Feeds(userID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	result.Subscriptions = make([]subscription, 0)
	for _, feed := range feeds {
		category, err := h.store.Category(userID, feed.Category.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		result.Subscriptions = append(result.Subscriptions, subscription{
			ID:         fmt.Sprint(feed.ID),
			Title:      feed.Title,
			URL:        feed.FeedURL,
			Categories: []subscriptionCategory{{fmt.Sprint(category.ID), category.Title}},
			HTMLURL:    feed.SiteURL,
			IconURL:    "", //TODO Icons are only base64 encode in DB yet
		})
	}
	json.OK(w, r, result)
}

func (h *handler) serve(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	dump, _ := httputil.DumpRequest(r, false)
	logger.Info("[Reader][UNKNOWN] [ClientIP=%s] URL: %s", clientIP, dump)
	logger.Error("Call to Google Reader API not implemented yet!!")
	json.OK(w, r, []string{})
}

func (h *handler) userInfo(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	logger.Info("[Reader][UserInfo] [ClientIP=%s] Sending", clientIP)
	output := request.QueryStringParam(r, "output", "")
	if output != "json" {
		err := fmt.Errorf("output only as json supported")
		logger.Error("[Reader][Login] %v", err)
		json.ServerError(w, r, err)
		return
	}
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	userInfo := userInfo{UserID: fmt.Sprint(user.ID), UserName: user.Username, UserProfileID: fmt.Sprint(user.ID), UserEmail: user.Username}
	json.OK(w, r, userInfo)
}
