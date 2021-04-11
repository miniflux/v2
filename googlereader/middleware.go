package googlereader // import "miniflux.app/googlereader"

import (
	"context"
	"net/http"
	"strings"

	"miniflux.app/http/request"
	"miniflux.app/logger"
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
		clientIP := request.ClientIP(r)
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			logger.Info("[Google Reader API] [ClientIP=%s] No API key provided", clientIP)
			//TODO Failure Response
			return
		}
		authSplits := strings.Split(authHeader, "=")

		if len(authSplits) == 1 {
			logger.Info("[Google Reader API] [ClientIP=%s] Authorisation Header is invalid", clientIP)
			//TODO Response
			return
		}
		auth := authSplits[1]
		user, err := m.store.UserByGoogleReaderToken(auth)
		if err != nil {
			logger.Error("[Google Reader API] %v", err)
			//TODO Failure Response
			return
		}

		if user == nil {
			logger.Info("[Google Reader API] [ClientIP=%s] No user found with this API key", clientIP)
			//TODO Failure Response
			return
		}

		logger.Info("[Google Reader API] [ClientIP=%s] User #%d is authenticated with user agent %q", clientIP, user.ID, r.UserAgent())
		m.store.SetLastLogin(user.ID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
