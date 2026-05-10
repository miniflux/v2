// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
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
	baseURL, err := url.Parse(config.Opts.BaseURL())
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: "Miniflux",
		RPID:          baseURL.Hostname(),
		RPOrigins:     []string{config.Opts.RootURL()},
	})
}

func (h *handler) beginRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	credentials, err := h.store.WebAuthnCredentialsByUserID(user.ID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	credentialDescriptors := make([]protocol.CredentialDescriptor, len(credentials))
	for i, credential := range credentials {
		credentialDescriptors[i] = credential.Credential.Descriptor()
	}

	options, sessionData, err := web.BeginRegistration(
		WebAuthnUser{
			User:    user,
			AuthnID: crypto.GenerateRandomBytes(32),
		},
		webauthn.WithExclusions(credentialDescriptors),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementPreferred),
		webauthn.WithExtensions(protocol.AuthenticationExtensions{"credProps": true}),
	)

	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	request.WebSession(r).SetWebAuthn(sessionData)
	response.JSON(w, r, options)
}

func (h *handler) finishRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	userID := request.UserID(r)
	user, err := h.store.UserByID(userID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	sessionData := request.WebSession(r).ConsumeWebAuthnSession()
	if sessionData == nil {
		response.JSONBadRequest(w, r, errors.New("missing webauthn session data"))
		return
	}
	webAuthnUser := WebAuthnUser{User: user, AuthnID: sessionData.UserID}
	credential, err := web.FinishRegistration(webAuthnUser, *sessionData, r)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	err = h.store.AddWebAuthnCredential(userID, sessionData.UserID, credential)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	handleEncoded := model.WebAuthnCredential{Handle: sessionData.UserID}.HandleEncoded()
	redirect := h.routePath("/webauthn/%s/rename", handleEncoded)
	response.JSON(w, r, map[string]string{"redirect": redirect})
}

func (h *handler) beginLogin(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	var user *model.User
	username := request.QueryStringParam(r, "username", "")
	if username != "" {
		user, err = h.store.UserByUsername(username)
		if err != nil {
			response.JSONUnauthorized(w, r)
			return
		}
	}

	var assertion *protocol.CredentialAssertion
	var sessionData *webauthn.SessionData
	if user != nil {
		credentials, err := h.store.WebAuthnCredentialsByUserID(user.ID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		assertion, sessionData, err = web.BeginLogin(WebAuthnUser{User: user, Credentials: credentials})
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	} else {
		assertion, sessionData, err = web.BeginDiscoverableLogin()
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
	}

	request.WebSession(r).SetWebAuthn(sessionData)
	response.JSON(w, r, assertion)
}

func (h *handler) finishLogin(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn()
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	slog.Debug("WebAuthn: parsed response flags",
		slog.Bool("user_present", parsedResponse.Response.AuthenticatorData.Flags.HasUserPresent()),
		slog.Bool("user_verified", parsedResponse.Response.AuthenticatorData.Flags.HasUserVerified()),
		slog.Bool("has_attested_credential_data", parsedResponse.Response.AuthenticatorData.Flags.HasAttestedCredentialData()),
		slog.Bool("has_backup_eligible", parsedResponse.Response.AuthenticatorData.Flags.HasBackupEligible()),
		slog.Bool("has_backup_state", parsedResponse.Response.AuthenticatorData.Flags.HasBackupState()),
	)

	sessionData := request.WebSession(r).ConsumeWebAuthnSession()
	if sessionData == nil {
		response.JSONBadRequest(w, r, errors.New("missing webauthn session data"))
		return
	}

	var user *model.User
	username := request.QueryStringParam(r, "username", "")
	if username != "" {
		user, err = h.store.UserByUsername(username)
		if err != nil {
			response.JSONUnauthorized(w, r)
			return
		}
	}

	var matchingCredential *model.WebAuthnCredential
	if user != nil {
		storedCredentials, err := h.store.WebAuthnCredentialsByUserID(user.ID)
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}

		sessionData.UserID = parsedResponse.Response.UserHandle
		webAuthnUser := WebAuthnUser{
			User:        user,
			AuthnID:     parsedResponse.Response.UserHandle,
			Credentials: storedCredentials,
		}

		// Since go-webauthn v0.11.0, the backup eligibility flag is strictly validated, but Miniflux does not store this flag.
		// This workaround set the flag based on the parsed response, and avoid "BackupEligible flag inconsistency detected during login validation" error.
		// See https://github.com/go-webauthn/webauthn/pull/240
		for index := range webAuthnUser.Credentials {
			webAuthnUser.Credentials[index].Credential.Flags.BackupEligible = parsedResponse.Response.AuthenticatorData.Flags.HasBackupEligible()
		}

		for _, cred := range webAuthnUser.WebAuthnCredentials() {
			slog.Debug("WebAuthn: stored credential flags",
				slog.Bool("user_present", cred.Flags.UserPresent),
				slog.Bool("user_verified", cred.Flags.UserVerified),
				slog.Bool("backup_eligible", cred.Flags.BackupEligible),
				slog.Bool("backup_state", cred.Flags.BackupState),
			)
		}

		validatedCredential, err := web.ValidateLogin(webAuthnUser, *sessionData, parsedResponse)
		if err != nil {
			slog.Warn("WebAuthn: ValidateLogin failed", slog.Any("error", err))
			response.JSONUnauthorized(w, r)
			return
		}

		for _, storedCredential := range storedCredentials {
			if bytes.Equal(validatedCredential.ID, storedCredential.Credential.ID) {
				matchingCredential = &storedCredential
			}
		}

		if matchingCredential == nil {
			response.JSONServerError(w, r, fmt.Errorf("no matching credential for %v", validatedCredential))
			return
		}
	} else {
		userByHandle := func(rawID, userHandle []byte) (webauthn.User, error) {
			var userID int64
			userID, matchingCredential, err = h.store.WebAuthnCredentialByHandle(userHandle)
			if err != nil {
				return nil, err
			}
			if userID == 0 {
				return nil, fmt.Errorf("no user found for handle %x", userHandle)
			}
			user, err = h.store.UserByID(userID)
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

			return WebAuthnUser{
				User:        user,
				AuthnID:     userHandle,
				Credentials: []model.WebAuthnCredential{*matchingCredential},
			}, nil
		}

		_, err = web.ValidateDiscoverableLogin(userByHandle, *sessionData, parsedResponse)
		if err != nil {
			slog.Warn("WebAuthn: ValidateDiscoverableLogin failed", slog.Any("error", err))
			response.JSONUnauthorized(w, r)
			return
		}
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

	if err := authenticateWebSession(w, r, h.store, user); err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) renameCredential(w http.ResponseWriter, r *http.Request) {
	view := view.New(h.tpl, r)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}
	credUserID, credential, err := h.store.WebAuthnCredentialByHandle(credentialHandle)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	if credUserID != user.ID {
		response.HTMLForbidden(w, r)
		return
	}

	webauthnForm := form.WebauthnForm{Name: credential.Name}

	view.Set("form", webauthnForm)
	view.Set("cred", credential)
	view.Set("menu", "settings")
	view.Set("user", user)
	navMetadata, _ := h.store.GetNavMetadata(user.ID)
	view.Set("countUnread", navMetadata.CountUnread)
	view.Set("countErrorFeeds", navMetadata.CountErrorFeeds)

	response.HTML(w, r, view.Render("webauthn_rename"))
}

func (h *handler) saveCredential(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	_, err := h.store.UserByID(userID)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}

	newName := r.FormValue("name")
	changed, err := h.store.WebAuthnUpdateName(userID, credentialHandle, newName)
	if err != nil {
		response.HTMLServerError(w, r, err)
		return
	}
	if changed == 0 {
		response.HTMLNotFound(w, r)
		return
	}

	response.HTMLRedirect(w, r, h.routePath("/settings"))
}

func (h *handler) deleteCredential(w http.ResponseWriter, r *http.Request) {
	credentialHandleEncoded := request.RouteStringParam(r, "credentialHandle")
	credentialHandle, err := hex.DecodeString(credentialHandleEncoded)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	err = h.store.DeleteCredentialByHandle(request.UserID(r), credentialHandle)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.NoContent(w, r)
}

func (h *handler) deleteAllCredentials(w http.ResponseWriter, r *http.Request) {
	err := h.store.DeleteAllWebAuthnCredentialsByUserID(request.UserID(r))
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}
	response.NoContent(w, r)
}
