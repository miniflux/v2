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
	httpRequest := httptest.NewRequest(http.MethodGet, path, nil)
	return httpRequest.WithContext(context.WithValue(httpRequest.Context(), request.WebSessionContextKey, session)), session
}

func TestNewPublicOmitsSessionSecrets(t *testing.T) {
	configureTestOptions(t)
	session, _ := model.NewWebSession("test", "127.0.0.1")
	session.SetSuccessMessage("saved")

	httpRequest, _ := webSessionRequest("/share/code")

	publicView := NewPublic(nil, httpRequest)

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

	httpRequest, session := webSessionRequest("/unread")

	defaultView := New(nil, httpRequest)

	if csrf := defaultView.params["csrf"]; csrf != session.CSRF() || csrf == "" {
		t.Fatalf(`Unexpected CSRF token, got %v`, csrf)
	}
}

// Every HTML view must carry a per-request CSP nonce and use it in its Content-Security-Policy.
func TestViewSetsCSPNonceAndPolicy(t *testing.T) {
	configureTestOptions(t)

	engine := template.NewEngine("")
	httpRequest, _ := webSessionRequest("/")

	defaultView := New(engine, httpRequest)

	nonce, ok := defaultView.params["cspNonce"].(string)
	if !ok || nonce == "" {
		t.Fatalf("view params must contain a non-empty cspNonce, got %v", defaultView.params["cspNonce"])
	}

	if policy := defaultView.CSP(); !strings.Contains(policy, "'nonce-"+nonce+"'") {
		t.Fatalf("view CSP() must contain the request nonce, nonce=%q policy=%q", nonce, policy)
	}
}

func TestViewHTMLSetsCSPHeaderWithMatchingNonce(t *testing.T) {
	configureTestOptions(t)

	engine := template.NewEngine("")
	engine.ParseTemplates()
	httpRequest, _ := webSessionRequest("/")

	recorder := httptest.NewRecorder()

	defaultView := New(engine, httpRequest)
	defaultView.HTML(recorder, httpRequest, "offline")

	if policy := recorder.Result().Header.Get("Content-Security-Policy"); policy != defaultView.CSP() {
		t.Fatalf("view HTML() must send the policy of the view, got %q, want %q", policy, defaultView.CSP())
	}

	body := recorder.Body.String()
	if strings.Contains(strings.ToLower(body), `http-equiv="content-security-policy"`) {
		t.Fatal("rendered HTML must not contain a meta Content-Security-Policy tag")
	}
}
