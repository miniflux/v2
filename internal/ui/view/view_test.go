// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "miniflux.app/v2/internal/ui/view"

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/model"
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

func TestNewPublicOmitsSessionSecrets(t *testing.T) {
	configureTestOptions(t)
	session, _ := model.NewWebSession("test", "127.0.0.1")
	session.SetSuccessMessage("saved")

	r := httptest.NewRequest(http.MethodGet, "/share/code", nil)
	r = r.WithContext(context.WithValue(r.Context(), request.WebSessionContextKey, session))

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
	session, _ := model.NewWebSession("test", "127.0.0.1")

	r := httptest.NewRequest(http.MethodGet, "/unread", nil)
	r = r.WithContext(context.WithValue(r.Context(), request.WebSessionContextKey, session))

	defaultView := New(nil, r)

	if csrf := defaultView.params["csrf"]; csrf != session.CSRF() || csrf == "" {
		t.Fatalf(`Unexpected CSRF token, got %v`, csrf)
	}
}
