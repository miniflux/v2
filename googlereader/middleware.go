// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package googlereader // import "miniflux.app/googlereader"

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"

	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
)

type middleware struct {
	store *storage.Storage
}

func newMiddleware(s *storage.Storage) *middleware {
	return &middleware{s}
}

func (m *middleware) clientLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	var username, password, output string
	var integration *model.Integration
	err := r.ParseForm()
	if err != nil {
		logger.Error("[GoogleReader][Login] [ClientIP=%s] Could not parse form", clientIP)
		json.Unauthorized(w, r)
		return
	}
	username = r.Form.Get("Email")
	password = r.Form.Get("Passwd")
	output = r.Form.Get("output")

	if username == "" || password == "" {
		logger.Error("[GoogleReader][Login] [ClientIP=%s] Empty username or password", clientIP)
		json.Unauthorized(w, r)
		return
	}

	if err = m.store.GoogleReaderUserCheckPassword(username, password); err != nil {
		logger.Error("[GoogleReader][Login] [ClientIP=%s] Invalid username or password: %s", clientIP, username)
		json.Unauthorized(w, r)
		return
	}

	logger.Info("[GoogleReader][Login] [ClientIP=%s] User authenticated: %s", clientIP, username)

	if integration, err = m.store.GoogleReaderUserGetIntegration(username); err != nil {
		logger.Error("[GoogleReader][Login] [ClientIP=%s] Could not load integration: %s", clientIP, username)
		json.Unauthorized(w, r)
		return
	}

	m.store.SetLastLogin(integration.UserID)

	token := getAuthToken(integration.GoogleReaderUsername, integration.GoogleReaderPassword)
	logger.Info("[GoogleReader][Login] [ClientIP=%s] Created token: %s", clientIP, token)
	result := login{SID: token, LSID: token, Auth: token}
	if output == "json" {
		json.OK(w, r, result)
		return
	}
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/plain; charset=UTF-8")
	builder.WithBody(result.String())
	builder.Write()
}

func (m *middleware) token(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)

	if !request.IsAuthenticated(r) {
		logger.Error("[GoogleReader][Token] [ClientIP=%s] User is not authenticated", clientIP)
		json.Unauthorized(w, r)
		return
	}
	token := request.GoolgeReaderToken(r)
	if token == "" {
		logger.Error("[GoogleReader][Token] [ClientIP=%s] User does not have token: %s", clientIP, request.UserID(r))
		json.Unauthorized(w, r)
		return
	}
	logger.Info("[GoogleReader][Token] [ClientIP=%s] token: %s", clientIP, token)
	w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (m *middleware) handleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) apiKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)

		var token string
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				logger.Error("[GoogleReader][Login] [ClientIP=%s] Could not parse form", clientIP)
				Unauthorized(w, r)
				return
			}
			token = r.Form.Get("T")
			if token == "" {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] Post-Form T field is empty", clientIP)
				Unauthorized(w, r)
				return
			}
		} else {
			authorization := r.Header.Get("Authorization")

			if authorization == "" {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] No token provided", clientIP)
				Unauthorized(w, r)
				return
			}
			fields := strings.Fields(authorization)
			if len(fields) != 2 {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
				Unauthorized(w, r)
				return
			}
			if fields[0] != "GoogleLogin" {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] Authorization header does not begin with GoogleLogin - '%s'", clientIP, authorization)
				Unauthorized(w, r)
				return
			}
			auths := strings.Split(fields[1], "=")
			if len(auths) != 2 {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
				Unauthorized(w, r)
				return
			}
			if auths[0] != "auth" {
				logger.Error("[GoogleReader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
				Unauthorized(w, r)
				return
			}
			token = auths[1]
		}

		parts := strings.Split(token, "/")
		if len(parts) != 2 {
			logger.Error("[GoogleReader][Auth] [ClientIP=%s] Auth token does not have the expected structure username/hash - '%s'", clientIP, token)
			Unauthorized(w, r)
			return
		}
		var integration *model.Integration
		var user *model.User
		var err error
		if integration, err = m.store.GoogleReaderUserGetIntegration(parts[0]); err != nil {
			logger.Error("[GoogleReader][Auth] [ClientIP=%s] token: %s", clientIP, token)
			logger.Error("[GoogleReader][Auth] [ClientIP=%s] No user found with the given google reader username: %s", clientIP, parts[0])
			Unauthorized(w, r)
			return
		}
		expectedToken := getAuthToken(integration.GoogleReaderUsername, integration.GoogleReaderPassword)
		if expectedToken != token {
			logger.Error("[GoogleReader][Auth] [ClientIP=%s] Token does not match: %s", clientIP, token)
			Unauthorized(w, r)
			return
		}
		if user, err = m.store.UserByID(integration.UserID); err != nil {
			logger.Error("[GoogleReader][Auth] [ClientIP=%s] No user found with the userID: %d", clientIP, integration.UserID)
			Unauthorized(w, r)
			return
		}

		m.store.SetLastLogin(integration.UserID)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, request.GoogleReaderToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getAuthToken(username, password string) string {
	token := hex.EncodeToString(hmac.New(sha1.New, []byte(username+password)).Sum(nil))
	token = username + "/" + token
	return token
}
