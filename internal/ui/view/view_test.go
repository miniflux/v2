// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "miniflux.app/v2/internal/ui/view"

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/template"
)

func configureTestOptions(t *testing.T) {
	t.Helper()

	parsedOptions, err := config.NewConfigParser().ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Unable to configure test options: %v`, err)
	}

	previousOptions := config.Opts
	config.Opts = parsedOptions
	t.Cleanup(func() {
		config.Opts = previousOptions
	})
}

func webSessionRequest(path string) (*http.Request, *model.WebSession) {
	session, _ := model.NewWebSession("test", "127.0.0.1")
	r := httptest.NewRequest(http.MethodGet, path, nil)
	return r.WithContext(context.WithValue(r.Context(), request.WebSessionContextKey, session)), session
}

func TestNewPublicOmitsSessionSecrets(t *testing.T) {
	configureTestOptions(t)
	session, _ := model.NewWebSession("test", "127.0.0.1")
	session.SetSuccessMessage("saved")

	r, _ := webSessionRequest("/share/code")

	publicView := NewPublic(nil, r)

	for _, key := range []string{"csrf", "flashSuccessMessage", "flashErrorMessage"} {
		if value, ok := publicView.params[key]; ok {
			t.Fatalf(`Public view should not set %q, got %v`, key, value)
		}
	}

	successMessage, _ := session.ConsumeMessages()
	if successMessage != "saved" {
		t.Fatalf(`Flash message should be kept for the next page, got %q`, successMessage)
	}
}

func TestNewIncludesCSRFToken(t *testing.T) {
	configureTestOptions(t)

	r, session := webSessionRequest("/unread")

	defaultView := New(nil, r)

	if csrf := defaultView.params["csrf"]; csrf != session.CSRF() || csrf == "" {
		t.Fatalf(`Unexpected CSRF token, got %v`, csrf)
	}
}

// Every HTML view must carry a per-request CSP nonce and be able to
// render its Content-Security-Policy header value (no meta tag).
func TestViewSetsCSPNonceAndPolicy(t *testing.T) {
	configureTestOptions(t)

	engine := template.NewEngine("")
	r, _ := webSessionRequest("/")

	v := New(engine, r)

	nonce, ok := v.params["cspNonce"].(string)
	if !ok || nonce == "" {
		t.Fatalf("view params must contain a non-empty cspNonce, got %v", v.params["cspNonce"])
	}

	policy := v.CSP()
	if strings.Contains(strings.ToLower(policy), "<meta") {
		t.Fatalf("view CSP() must not return a <meta> tag, got %q", policy)
	}
	if !strings.Contains(policy, "'nonce-"+nonce+"'") {
		t.Fatalf("view CSP() must contain the request nonce, nonce=%q policy=%q", nonce, policy)
	}
}

func TestViewHTMLSetsCSPHeaderWithMatchingNonce(t *testing.T) {
	configureTestOptions(t)

	engine := template.NewEngine("")
	engine.ParseTemplates()
	r, _ := webSessionRequest("/")

	w := httptest.NewRecorder()

	v := New(engine, r)
	v.HTML(w, r, "offline")

	resp := w.Result()
	policy := resp.Header.Get("Content-Security-Policy")
	if policy == "" {
		t.Fatal("view HTML() must set a Content-Security-Policy header")
	}
	if strings.Contains(strings.ToLower(policy), "<meta") {
		t.Fatalf("CSP header must be a header value, not a meta tag, got %q", policy)
	}

	body := w.Body.String()
	if strings.Contains(strings.ToLower(body), `http-equiv="content-security-policy"`) {
		t.Fatal("rendered HTML must not contain a meta Content-Security-Policy tag")
	}
}
