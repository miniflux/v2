// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/cookie"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/response/json"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

type WebAuthnUser struct {
	User        *model.User
	AuthnID     []byte
	Credentials []model.WebAuthnCredential
}

func (u WebAuthnUser) WebAuthnID() []byte {
	return u.AuthnID
}

func (u WebAuthnUser) WebAuthnName() string {
	return u.User.Username
}

func (u WebAuthnUser) WebAuthnDisplayName() string {
	return u.User.Username
}

func (u WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, cred := range u.Credentials {
		creds[i] = cred.Credential
	}
	return creds
}

func newWebAuthn() (*webauthn.WebAuthn, error) {
	url, err := url.Parse(config.Opts.BaseURL())
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: "Miniflux",
		RPID:          url.Hostname(),
		RPOrigins:     []string{config.Opts.RootURL()},
	})
}

func (h *handler) beginRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	uid := request.UserID(r)
	if uid == 0 {
		json.Unauthorized(w, r)
		return
	}
	user, err := h.store.UserByID(uid)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	var creds []model.WebAuthnCredential

	creds, err = h.store.WebAuthnCredentialsByUserID(user.ID)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	credsDescriptors := make([]protocol.CredentialDescriptor, len(creds))
	for i, cred := range creds {
		credsDescriptors[i] = cred.Credential.Descriptor()
	}

	options, sessionData, err := web.BeginRegistration(
		WebAuthnUser{
			user,
			crypto.GenerateRandomBytes(32),
			nil,
		},
		webauthn.WithExclusions(credsDescriptors),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		webauthn.WithExtensions(protocol.AuthenticationExtensions{"credProps": true}),
	)

	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	s := session.New(h.store, request.SessionID(r))
	s.SetWebAuthnSessionData(&model.WebAuthnSession{SessionData: sessionData})
	json.OK(w, r, options)
}

func (h *handler) finishRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	uid := request.UserID(r)
	if uid == 0 {
		json.Unauthorized(w, r)
		return
	}
	user, err := h.store.UserByID(uid)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	sessionData := request.WebAuthnSessionData(r)
	webAuthnUser := WebAuthnUser{user, sessionData.UserID, nil}
	cred, err := web.FinishRegistration(webAuthnUser, *sessionData.SessionData, r)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	err = h.store.AddWebAuthnCredential(uid, sessionData.UserID, cred)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	handleEncoded := model.WebAuthnCredential{Handle: sessionData.UserID}.HandleEncoded()
	redirect := route.Path(h.router, "webauthnRename", "credentialHandle", handleEncoded)
	json.OK(w, r, map[string]string{"redirect": redirect})
}

func (h *handler) beginLogin(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	var user *model.User
	username := request.QueryStringParam(r, "username", "")
	if username != "" {
		user, err = h.store.UserByUsername(username)
		if err != nil {
			json.Unauthorized(w, r)
			return
		}
	}

	var assertion *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	if user != nil {
		creds, err := h.store.WebAuthnCredentialsByUserID(user.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		assertion, sessionData, err = web.BeginLogin(WebAuthnUser{user, nil, creds})
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	} else {
		assertion, sessionData, err = web.BeginDiscoverableLogin()
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
	}

	s := session.New(h.store, request.SessionID(r))
	s.SetWebAuthnSessionData(&model.WebAuthnSession{SessionData: sessionData})
	json.OK(w, r, assertion)
}

func (h *handler) finishLogin(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	slog.Debug("WebAuthn: parsed response flags",
		slog.Bool("user_present", parsedResponse.Response.AuthenticatorData.Flags.HasUserPresent()),
		slog.Bool("user_verified", parsedResponse.Response.AuthenticatorData.Flags.HasUserVerified()),
		slog.Bool("has_attested_credential_data", parsedResponse.Response.AuthenticatorData.Flags.HasAttestedCredentialData()),
		slog.Bool("has_backup_eligible", parsedResponse.Response.AuthenticatorData.Flags.HasBackupEligible()),
		slog.Bool("has_backup_state", parsedResponse.Response.AuthenticatorData.Flags.HasBackupState()),
	)

	sessionData := request.WebAuthnSessionData(r)

	var user *model.User
	username := request.QueryStringParam(r, "username", "")
	if username != "" {
		user, err = h.store.UserByUsername(username)
		if err != nil {
			json.Unauthorized(w, r)
			return
		}
	}

	var matchingCredential *model.WebAuthnCredential
	if user != nil {
		storedCredentials, err := h.store.WebAuthnCredentialsByUserID(user.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}

		sessionData.UserID = parsedResponse.Response.UserHandle
		webAuthUser := WebAuthnUser{user, parsedResponse.Response.UserHandle, storedCredentials}

		// Since go-webauthn v0.11.0, the backup eligibility flag is strictly validated, but Miniflux does not store this flag.
		// This workaround set the flag based on the parsed response, and avoid "BackupEligible flag inconsistency detected during login validation" error.
		// See https://github.com/go-webauthn/webauthn/pull/240
		for index := range webAuthUser.Credentials {
			webAuthUser.Credentials[index].Credential.Flags.BackupEligible = parsedResponse.Response.AuthenticatorData.Flags.HasBackupEligible()
		}

		for _, webAuthCredential := range webAuthUser.WebAuthnCredentials() {
			slog.Debug("WebAuthn: stored credential flags",
				slog.Bool("user_present", webAuthCredential.Flags.UserPresent),
				slog.Bool("user_verified", webAuthCredential.Flags.UserVerified),
				slog.Bool("backup_eligible", webAuthCredential.Flags.BackupEligible),
				slog.Bool("backup_state", webAuthCredential.Flags.BackupState),
			)
		}

		credCredential, err := web.ValidateLogin(webAuthUser, *sessionData.SessionData, parsedResponse)
		if err != nil {
			slog.Warn("WebAuthn: ValidateLogin failed", slog.Any("error", err))
			json.Unauthorized(w, r)
			return
		}

		for _, storedCredential := range storedCredentials {
			if bytes.Equal(credCredential.ID, storedCredential.Credential.ID) {
				matchingCredential = &storedCredential
			}
		}

		if matchingCredential == nil {
			json.ServerError(w, r, fmt.Errorf("no matching credential for %v", credCredential))
			return
		}
	} else {
		userByHandle := func(rawID, userHandle []byte) (webauthn.User, error) {
			var uid int64
			uid, matchingCredential, err = h.store.WebAuthnCredentialByHandle(userHandle)
			if err != nil {
				return nil, err
			}
			if uid == 0 {
				return nil, fmt.Errorf("no user found for handle %x", userHandle)
			}
			user, err = h.store.UserByID(uid)
			if err != nil {
				return nil, err
			}
			if user == nil {
				return nil, fmt.Errorf("no user found for handle %x", userHandle)
			}

			// Since go-webauthn v0.11.0, the backup eligibility flag is strictly validated, but Miniflux does not store this flag.
			// This workaround set the flag based on the parsed response, and avoid "BackupEligible flag inconsistency detected during login validation" error.
			// See https://github.com/go-webauthn/webauthn/pull/240
			matchingCredential.Credential.Flags.BackupEligible = parsedResponse.Response.AuthenticatorData.Flags.HasBackupEligible()

			return WebAuthnUser{user, userHandle, []model.WebAuthnCredential{*matchingCredential}}, nil
		}

		_, err = web.ValidateDiscoverableLogin(userByHandle, *sessionData.SessionData, parsedResponse)
		if err != nil {
			slog.Warn("WebAuthn: ValidateDiscoverableLogin failed", slog.Any("error", err))
			json.Unauthorized(w, r)
			return
		}
	}

	sessionToken, _, err := h.store.CreateUserSessionFromUsername(user.Username, r.UserAgent(), request.ClientIP(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	h.store.WebAuthnSaveLogin(matchingCredential.Handle)

	slog.Info("User authenticated successfully with webauthn",
		slog.Bool("authentication_successful", true),
		slog.String("client_ip", request.ClientIP(r)),
		slog.String("user_agent", r.UserAgent()),
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username),
	)
	h.store.SetLastLogin(user.ID)

	sess := session.New(h.store, request.SessionID(r))
	sess.SetLanguage(user.Language)
	sess.SetTheme(user.Theme)

	http.SetCookie(w, cookie.New(
		cookie.CookieUserSessionID,
		sessionToken,
		config.Opts.HTTPS,
		config.Opts.BasePath(),
	))

	json.NoContent(w, r)
}

func (h *handler) renameCredential(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	cred_uid, cred, err := h.store.WebAuthnCredentialByHandle(credentialHandle)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if cred_uid != user.ID {
		html.Forbidden(w, r)
		return
	}

	webauthnForm := form.WebauthnForm{Name: cred.Name}

	view.Set("form", webauthnForm)
	view.Set("cred", cred)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID))

	html.OK(w, r, view.Render("webauthn_rename"))
}

func (h *handler) saveCredential(w http.ResponseWriter, r *http.Request) {
	_, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	newName := r.FormValue("name")
	err = h.store.WebAuthnUpdateName(credentialHandle, newName)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "settings"))
}

func (h *handler) deleteCredential(w http.ResponseWriter, r *http.Request) {
	uid := request.UserID(r)
	if uid == 0 {
		json.Unauthorized(w, r)
		return
	}

	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	err = h.store.DeleteCredentialByHandle(uid, credentialHandle)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.NoContent(w, r)
}

func (h *handler) deleteAllCredentials(w http.ResponseWriter, r *http.Request) {
	err := h.store.DeleteAllWebAuthnCredentialsByUserID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	json.NoContent(w, r)
}
