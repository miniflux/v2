package googlereader // import "miniflux.app/googlereader"

import (
	"net/http"

	"github.com/gorilla/mux"
	"miniflux.app/logger"
	"miniflux.app/storage"
)

type handler struct {
	store *storage.Storage
}

// Serve handles Google Reader API calls.
func Serve(router *mux.Router, store *storage.Storage) {
	handler := &handler{store}
	sr := router.PathPrefix("/google-reader-api").Subrouter()
	sr.Use(newMiddleware(store).serve)
	sr.HandleFunc("/", handler.serve).Name("googleReaderEndpoint")
}

func (h *handler) serve(w http.ResponseWriter, r *http.Request) {
	logger.Info("Call to Google Reader API not implemented yet!!")
}
