package googlereader // import "miniflux.app/googlereader"

import (
	"fmt"
	"net/http"

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

func (h *handler) serve(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	logger.Error("Call to Google Reader API not implemented yet!!")
	json.OK(w, r, []string{})
	logger.Info("[Reader][Login] [ClientIP=%s] Sending", clientIP)
}

func (h *handler) subscriptionList(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	logger.Error("Call to Google Reader API not implemented yet!!")
	json.OK(w, r, []string{})
	logger.Info("[Reader][SubscriptionList] [ClientIP=%s] Sending", clientIP)
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
	userInfo := userInfo{UserId: fmt.Sprint(user.ID), UserName: user.Username, UserProfileId: fmt.Sprint(user.ID), UserEmail: user.Username}
	json.OK(w, r, userInfo)
}
