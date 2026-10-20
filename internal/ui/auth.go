// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"context"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/oauth2"
	"miniflux.app/v2/internal/storage"
)

const (
	sessionCookieName         = "MinifluxSessionID"
	preferencesCookieName     = "MinifluxPreferences"
	preferencesCookieLifetime = 365 * 24 * time.Hour
)

// authenticateWebSession binds the current browser session to the given user,
// rotates its identifier and secret, and refreshes the client cookie.
func authenticateWebSession(w http.ResponseWriter, r *http.Request, store *storage.Storage, user *model.User) error {
	session := request.WebSession(r)
	session.SetUser(user)

	oldID, secret := session.Rotate()
	if err := store.RotateWebSession(oldID, session); err != nil {
		return err
	}

	setSessionCookie(w, session, secret)
	return nil
}

// setSessionCookie writes the session cookie to the response with the
// security attributes used by miniflux (HttpOnly, SameSite=Lax, Secure
// when HTTPS).
func setSessionCookie(w http.ResponseWriter, session *model.WebSession, secret string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID + "." + secret,
		Path:     cookiePath(),
		Secure:   config.Opts.HTTPS(),
		HttpOnly: true,
		Expires:  time.Now().Add(config.Opts.CleanupRemoveSessionsInterval()),
		SameSite: http.SameSiteLaxMode,
	})
}

// setPreferencesCookie remembers the language and theme of a user who logs
// out, so the login page keeps using them without keeping a session around.
func setPreferencesCookie(w http.ResponseWriter, language, theme string) {
	http.SetCookie(w, &http.Cookie{
		Name:     preferencesCookieName,
		Value:    language + "." + theme,
		Path:     cookiePath(),
		Secure:   config.Opts.HTTPS(),
		HttpOnly: true,
		Expires:  time.Now().Add(preferencesCookieLifetime),
		SameSite: http.SameSiteLaxMode,
	})
}

// preferencesFromCookie returns the language and theme stored by
// setPreferencesCookie. It reports false when the cookie is missing or holds
// an unknown language or theme.
func preferencesFromCookie(r *http.Request) (language, theme string, ok bool) {
	cookie, err := r.Cookie(preferencesCookieName)
	if err != nil {
		return "", "", false
	}

	language, theme, ok = strings.Cut(cookie.Value, ".")
	if !ok {
		return "", "", false
	}

	if _, found := locale.AvailableLanguages[language]; !found {
		return "", "", false
	}

	if _, found := model.Themes()[theme]; !found {
		return "", "", false
	}

	return language, theme, true
}

func cookiePath() string {
	if path := config.Opts.BasePath(); path != "" {
		return path
	}
	return "/"
}

func getOAuth2Manager(ctx context.Context) *oauth2.Manager {
	return oauth2.NewManager(
		ctx,
		config.Opts.OAuth2Provider(),
		config.Opts.OAuth2ClientID(),
		config.Opts.OAuth2ClientSecret(),
		config.Opts.OAuth2RedirectURL(),
		config.Opts.OAuth2OIDCDiscoveryEndpoint(),
	)
}
