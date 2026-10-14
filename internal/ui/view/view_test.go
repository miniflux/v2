// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/template"
)

// Every HTML view must carry a per-request CSP nonce and be able to
// render its Content-Security-Policy header value (no meta tag).
func TestViewSetsCSPNonceAndPolicy(t *testing.T) {
	previousOptions := config.Opts
	config.Opts = config.NewConfigOptions()
	defer func() { config.Opts = previousOptions }()

	engine := template.NewEngine("")
	session, _ := model.NewWebSession("test-agent", "127.0.0.1")
	r := httptest.NewRequest("GET", "/", nil).WithContext(
		context.WithValue(context.Background(), request.WebSessionContextKey, session),
	)

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
	previousOptions := config.Opts
	config.Opts = config.NewConfigOptions()
	defer func() { config.Opts = previousOptions }()

	engine := template.NewEngine("")
	engine.ParseTemplates()

	session, _ := model.NewWebSession("test-agent", "127.0.0.1")
	r := httptest.NewRequest("GET", "/", nil).WithContext(
		context.WithValue(context.Background(), request.WebSessionContextKey, session),
	)
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
