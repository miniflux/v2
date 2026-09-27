// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "miniflux.app/v2/internal/ui/view"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/template"
	"miniflux.app/v2/internal/ui/static"
)

// view wraps template argument building.
type view struct {
	tpl    *template.Engine
	r      *http.Request
	params map[string]any
}

// Set adds a new template argument.
func (v *view) Set(param string, value any) *view {
	v.params[param] = value
	return v
}

// Render executes the template with arguments.
func (v *view) Render(template string) []byte {
	return v.tpl.Render(template+".html", v.params)
}

// CSP returns the Content-Security-Policy header value for this view.
// The policy is delivered as an HTTP header (never as a meta tag, see
// deadhead rule meta/http-equiv-content-security-policy): meta policies
// drop frame-ancestors/report-uri/sandbox, stop the preload scanner, and
// apply late.
func (v *view) CSP() string {
	var user *model.User
	if u, ok := v.params["user"].(*model.User); ok {
		user = u
	}
	nonce, _ := v.params["cspNonce"].(string)
	return template.CSPPolicy(user, nonce)
}

// HTML renders the given template and writes it as an HTML response with
// the Content-Security-Policy header set.
func (v *view) HTML(w http.ResponseWriter, r *http.Request, template string) {
	response.NewBuilder(w, r).
		WithHeader("Content-Type", "text/html; charset=utf-8").
		WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store").
		WithHeader("Content-Security-Policy", v.CSP()).
		WithBodyAsBytes(v.Render(template)).
		Write()
}

// New returns a new view with default parameters.
func New(tpl *template.Engine, r *http.Request) *view {
	webSession := request.WebSession(r)
	theme := webSession.Theme()
	flashSuccessMessage, flashErrorMessage := webSession.ConsumeMessages()
	return &view{tpl, r, map[string]any{
		"menu":                "",
		"csrf":                webSession.CSRF(),
		"flashSuccessMessage": flashSuccessMessage,
		"flashErrorMessage":   flashErrorMessage,
		"theme":               theme,
		"language":            webSession.Language(),
		"theme_checksum":      static.StylesheetBundles[theme+".css"].Checksum,
		"app_js_checksum":     static.JavascriptBundles["app.js"].Checksum,
		"sw_js_checksum":      static.JavascriptBundles["service-worker.js"].Checksum,
		"webAuthnEnabled":     config.Opts.WebAuthn(),
		"cspNonce":            crypto.GenerateRandomStringHex(16),
	}}
}
