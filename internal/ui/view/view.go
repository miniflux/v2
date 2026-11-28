// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "miniflux.app/v2/internal/ui/view"

import (
	"net/http"
	"strings"

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
func (v *view) CSP() string {
	user, _ := v.params["user"].(*model.User)
	nonce, _ := v.params["cspNonce"].(string)
	return csp(user, nonce)
}

// csp returns the Content-Security-Policy of templated pages, with the directives in a fixed order.
func csp(user *model.User, nonce string) string {
	var fontSrc string
	styleSrc := "'nonce-" + nonce + "'"
	if user != nil && user.ExternalFontHosts != "" {
		fontSrc = user.ExternalFontHosts
		if user.Stylesheet != "" {
			styleSrc += " " + user.ExternalFontHosts
		}
	}

	directives := []struct{ name, value string }{
		{"default-src", "'none'"},
		{"connect-src", "'self'"},
		{"font-src", fontSrc},
		{"frame-src", "*"},
		{"img-src", "* data:"},
		{"manifest-src", "'self'"},
		{"media-src", "*"},
		{"require-trusted-types-for", "'script'"},
		{"script-src", "'nonce-" + nonce + "' 'strict-dynamic'"},
		{"style-src", styleSrc},
		{"trusted-types", "html url"},
	}

	policy := make([]string, 0, len(directives))
	for _, directive := range directives {
		// font-src is only set when the user has external font hosts.
		if directive.value != "" {
			policy = append(policy, directive.name+" "+directive.value)
		}
	}
	return strings.Join(policy, "; ")
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
	flashSuccessMessage, flashErrorMessage := webSession.ConsumeMessages()
	v := newView(tpl, r)
	v.params["csrf"] = webSession.CSRF()
	v.params["flashSuccessMessage"] = flashSuccessMessage
	v.params["flashErrorMessage"] = flashErrorMessage
	return v
}

// NewPublic returns a view for public pages, without the CSRF token or flash messages.
func NewPublic(tpl *template.Engine, r *http.Request) *view {
	return newView(tpl, r)
}

func newView(tpl *template.Engine, r *http.Request) *view {
	webSession := request.WebSession(r)
	theme := webSession.Theme()
	return &view{tpl, r, map[string]any{
		"menu":            "",
		"theme":           theme,
		"language":        webSession.Language(),
		"theme_checksum":  static.StylesheetBundles[theme+".css"].Checksum,
		"app_js_checksum": static.JavascriptBundles["app.js"].Checksum,
		"sw_js_checksum":  static.JavascriptBundles["service-worker.js"].Checksum,
		"webAuthnEnabled": config.Opts.WebAuthn(),
		"cspNonce":        crypto.GenerateRandomStringHex(16),
	}}
}
