package googlereader // import "miniflux.app/googlereader"

import (
	"net/http"

	"miniflux.app/storage"
)

type middleware struct {
	store *storage.Storage
}

func newMiddleware(s *storage.Storage) *middleware {
	return &middleware{s}
}

func (m *middleware) serve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Not implemented yet !!
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
