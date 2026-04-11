// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
)

func (h *handler) oauth2Callback(w http.ResponseWriter, r *http.Request) {
	provider := request.RouteStringParam(r, "provider")
	if provider == "" {
		slog.Warn("Invalid or missing OAuth2 provider")
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	code := request.QueryStringParam(r, "code", "")
	if code == "" {
		slog.Warn("No code received on OAuth2 callback")
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	sess := request.WebSession(r)

	state := request.QueryStringParam(r, "state", "")
	if subtle.ConstantTimeCompare([]byte(state), []byte(sess.OAuth2State())) == 0 {
		slog.Warn("Invalid OAuth2 state value received")
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	codeVerifier := sess.OAuth2CodeVerifier()
	sess.ClearOAuth2Flow()

	authProvider, err := getOAuth2Manager(r.Context()).FindProvider(provider)
	if err != nil {
		slog.Error("Unable to initialize OAuth2 provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	profile, err := authProvider.Profile(r.Context(), code, codeVerifier)
	if err != nil {
		slog.Warn("Unable to get OAuth2 profile from provider",
			slog.String("provider", provider),
			slog.Any("error", err),
		)
		response.HTMLRedirect(w, r, h.routePath("/"))
		return
	}

	printer := locale.NewPrinter(sess.Language())

	if request.IsAuthenticated(r) {
		loggedUser, err := h.store.UserByID(request.UserID(r))
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		if h.store.AnotherUserWithFieldExists(loggedUser.ID, profile.Key, profile.ID) {
			slog.Error("Oauth2 user cannot be associated because it is already associated with another user",
				slog.Int64("user_id", loggedUser.ID),
				slog.String("oauth2_provider", provider),
				slog.String("oauth2_profile_id", profile.ID),
			)
			sess.SetErrorMessage(printer.Print("error.duplicate_linked_account"))
			response.HTMLRedirect(w, r, h.routePath("/settings"))
			return
		}

		existingProfileID := authProvider.UserProfileID(loggedUser)
		if existingProfileID != "" && existingProfileID != profile.ID {
			slog.Error("Oauth2 user cannot be associated because this user is already linked to a different identity",
				slog.Int64("user_id", loggedUser.ID),
				slog.String("oauth2_provider", provider),
				slog.String("existing_profile_id", existingProfileID),
				slog.String("new_profile_id", profile.ID),
			)
			sess.SetErrorMessage(printer.Print("error.duplicate_linked_account"))
			response.HTMLRedirect(w, r, h.routePath("/settings"))
			return
		}

		authProvider.PopulateUserWithProfileID(loggedUser, profile)
		if err := h.store.UpdateUser(loggedUser); err != nil {
			response.HTMLServerError(w, r, err)
			return
		}

		sess.SetSuccessMessage(printer.Print("alert.account_linked"))
		response.HTMLRedirect(w, r, h.routePath("/settings"))
		return
	}

	user, err := h.store.UserByField(profile.Key, profile.ID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if user == nil {
		if !config.Opts.IsOAuth2UserCreationAllowed() {
			response.HTMLForbidden(w, r)
			return
		}

		if h.store.UserExists(profile.Username) {
			response.HTMLBadRequest(w, r, errors.New(printer.Print("error.user_already_exists")))
			return
		}

		userCreationRequest := &model.UserCreationRequest{Username: profile.Username}
		authProvider.PopulateUserCreationWithProfileID(userCreationRequest, profile)

		user, err = h.store.CreateUser(userCreationRequest)
		if err != nil {
			response.HTMLServerError(w, r, err)
			return
		}
	}

	slog.Info("User authenticated successfully using OAuth2",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", request.ClientIP(r)),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username),
	)

	h.store.SetLastLogin(user.ID)
	if err := authenticateWebSession(w, r, h.store, user); err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	response.HTMLRedirect(w, r, h.basePath+"/"+user.DefaultHomePage)
}
