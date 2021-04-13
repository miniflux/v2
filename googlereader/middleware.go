// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package googlereader // import "miniflux.app/googlereader"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muesli/cache2go"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/storage"
)

type middleware struct {
	store *storage.Storage
	cache *cache2go.CacheTable
}

func newMiddleware(s *storage.Storage) *middleware {
	c := cache2go.Cache("GoogleReader")
	return &middleware{s, c}
}

func (m *middleware) clientLogin(w http.ResponseWriter, r *http.Request) {
	clientIP := request.ClientIP(r)
	var username, password, output string
	dump, _ := httputil.DumpRequest(r, false)
	logger.Info("[Reader][Login] [ClientIP=%s] URL: %s", clientIP, dump)
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			logger.Error("[Reader][Login] [ClientIP=%s] Could not parse form", clientIP)
			json.Unauthorized(w, r)
			return
		}
		username = r.Form.Get("Email")
		password = r.Form.Get("Passwd")
		output = r.Form.Get("output")
	} else {
		username = request.QueryStringParam(r, "Email", "")
		password = request.QueryStringParam(r, "Passwd", "")
		output = request.QueryStringParam(r, "output", "")
	}

	if username == "" || password == "" {
		logger.Error("[Reader][Login] [ClientIP=%s] Empty username or password", clientIP)
		json.Unauthorized(w, r)
		return
	}

	if err := m.store.CheckPassword(username, password); err != nil {
		logger.Error("[Reader][Login] [ClientIP=%s] Invalid username or password: %s", clientIP, username)
		json.Unauthorized(w, r)
		return
	}

	user, err := m.store.UserByUsername(username)
	if err != nil {
		logger.Error("[Reader][Login] %v", err)
		json.ServerError(w, r, err)
		return
	}

	if user == nil {
		logger.Error("[Reader][Login] [ClientIP=%s] User not found: %s", clientIP, username)
		json.Unauthorized(w, r)
		return
	}

	logger.Info("[Reader][Login] [ClientIP=%s] User authenticated: %s", clientIP, username)

	m.store.SetLastLogin(user.ID)

	token := username + "/" + uuid.New().String()
	// token := strings.ReplaceAll(uuid.New().String(), "-", "")
	// token := username + "/" + strings.ReplaceAll(uuid.New().String(), "-", "")
	token = fmt.Sprintf("%-57s", token)
	token = strings.ReplaceAll(token, " ", "x")
	logger.Info("[Reader][Login] [ClientIP=%s] Created token: %s", clientIP, token)
	result := login{SID: token, LSID: token, Auth: token}
	m.cache.Add(token, 7*24*time.Hour, user)
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
		logger.Error("[Reader][Token] [ClientIP=%s] User is not authenticated", clientIP)
		json.Unauthorized(w, r)
		return
	}
	token := request.GoolgeReaderToken(r)
	if token == "" {
		logger.Error("[Reader][Token] [ClientIP=%s] User does not have token: %s", clientIP, request.UserID(r))
		json.Unauthorized(w, r)
		return
	}
	logger.Info("[Reader][Token] [ClientIP=%s] token: %s", clientIP, token)
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
		authorization := r.Header.Get("Authorization")

		if authorization == "" {
			logger.Error("[Reader][Auth] [ClientIP=%s] No token provided", clientIP)
			json.Unauthorized(w, r)
			return
		}
		fields := strings.Fields(authorization)
		if len(fields) != 2 {
			logger.Error("[Reader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
			json.Unauthorized(w, r)
			return
		}
		if fields[0] != "GoogleLogin" {
			logger.Error("[Reader][Auth] [ClientIP=%s] Authorization header does not begin with GoogleLogin - '%s'", clientIP, authorization)
			json.Unauthorized(w, r)
			return
		}
		auths := strings.Split(fields[1], "=")
		if len(auths) != 2 {
			logger.Error("[Reader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
			json.Unauthorized(w, r)
			return
		}
		if auths[0] != "auth" {
			logger.Error("[Reader][Auth] [ClientIP=%s] Authorization header does not have the expected structure GoogleLogin auth=xxxxxx - '%s'", clientIP, authorization)
			json.Unauthorized(w, r)
			return
		}

		token := auths[1]
		res, err := m.cache.Value(token)
		if err != nil {
			logger.Error("[Reader][Auth] [ClientIP=%s] No user found with the given API key", clientIP)
			json.Unauthorized(w, r)
			return
		}
		user, ok := (res.Data()).(*model.User)
		if !ok {
			err = fmt.Errorf("could not cast to user")
			logger.Error("[API][BasicAuth] could not cast to user!")
			json.ServerError(w, r, err)
			return
		}
		logger.Info("[Reader][Auth] [ClientIP=%s] User authenticated: %s", clientIP, user.Username)
		m.store.SetLastLogin(user.ID)
		m.store.SetAPIKeyUsedTimestamp(user.ID, authorization)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserTimezoneContextKey, user.Timezone)
		ctx = context.WithValue(ctx, request.IsAdminUserContextKey, user.IsAdmin)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, request.GoogleReaderToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
