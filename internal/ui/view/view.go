// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "miniflux.app/v2/internal/ui/view"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/template"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/static"
)

// View wraps template argument building.
type View struct {
	tpl    *template.Engine
	r      *http.Request
	params map[string]interface{}
}

// Set adds a new template argument.
func (v *View) Set(param string, value interface{}) *View {
	v.params[param] = value
	return v
}

// Render executes the template with arguments.
func (v *View) Render(template string) []byte {
	return v.tpl.Render(template+".html", v.params)
}

// New returns a new view with default parameters.
func New(tpl *template.Engine, r *http.Request, sess *session.Session) *View {
	theme := request.UserTheme(r)
	return &View{tpl, r, map[string]interface{}{
		"menu":                 "",
		"csrf":                 request.CSRF(r),
		"flashMessage":         sess.FlashMessage(request.FlashMessage(r)),
		"flashErrorMessage":    sess.FlashErrorMessage(request.FlashErrorMessage(r)),
		"theme":                theme,
		"language":             request.UserLanguage(r),
		"theme_checksum":       static.StylesheetBundleChecksums[theme],
		"app_js_checksum":      static.JavascriptBundleChecksums["app"],
		"sw_js_checksum":       static.JavascriptBundleChecksums["service-worker"],
		"webauthn_js_checksum": static.JavascriptBundleChecksums["webauthn"],
		"webAuthnEnabled":      config.Opts.WebAuthn(),
	}}
}
