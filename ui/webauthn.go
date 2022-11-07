package ui // import "miniflux.app/ui"

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/http/cookie"
	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/session"
)

type WebAuthnUser struct {
	User        *model.User
	AuthnID     []byte
	Credentials []webauthn.Credential
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
	return u.Credentials
}

func newWebAuthn(h *handler) (*webauthn.WebAuthn, error) {
	url, err := url.Parse(config.Opts.BaseURL())
	if err != nil {
		return nil, err
	}
	return webauthn.New(&webauthn.Config{
		RPDisplayName: "Miniflux",
		RPID:          url.Hostname(),
		RPOrigin:      config.Opts.RootURL(),
		RPIcon:        config.Opts.BaseURL() + route.Path(h.router, "favicon"),
	})
}

func (h *handler) beginRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn(h)
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
	options, sessionData, err := web.BeginRegistration(WebAuthnUser{
		user,
		crypto.GenerateRandomBytes(32),
		nil,
	})
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	s := session.New(h.store, request.SessionID(r))
	s.SetWebAuthnSessionData(&model.WebAuthnSession{SessionData: sessionData})
	json.OK(w, r, options)
}

func (h *handler) finishRegistration(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn(h)
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

	json.NoContent(w, r)
}

func (h *handler) beginLogin(w http.ResponseWriter, r *http.Request) {
	web, err := newWebAuthn(h)
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
	web, err := newWebAuthn(h)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
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
	if user != nil {
		creds, err := h.store.WebAuthnCredentialsByUserID(user.ID)
		if err != nil {
			json.ServerError(w, r, err)
			return
		}
		sessionData.SessionData.UserID = parsedResponse.Response.UserHandle
		_, err = web.ValidateLogin(WebAuthnUser{user, parsedResponse.Response.UserHandle, creds}, *sessionData.SessionData, parsedResponse)
		if err != nil {
			json.Unauthorized(w, r)
			return
		}
	} else {
		userByHandle := func(rawID, userHandle []byte) (webauthn.User, error) {
			uid, cred, err := h.store.WebAuthnCredentialByHandle(userHandle)
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
			return WebAuthnUser{user, userHandle, []webauthn.Credential{*cred}}, nil
		}

		_, err = web.ValidateDiscoverableLogin(userByHandle, *sessionData.SessionData, parsedResponse)
		if err != nil {
			json.Unauthorized(w, r)
			return
		}
	}

	sessionToken, _, err := h.store.CreateUserSessionFromUsername(user.Username, r.UserAgent(), request.ClientIP(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}

	logger.Info("[webauthn] username=%s just logged in", user.Username)
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

func (h *handler) deleteAllCredentials(w http.ResponseWriter, r *http.Request) {
	err := h.store.DeleteAllWebAuthnCredentialsByUserID(request.UserID(r))
	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	json.NoContent(w, r)
}
