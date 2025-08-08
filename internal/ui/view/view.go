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

// New returns a new view with default parameters.
func New(tpl *template.Engine, r *http.Request, sess *session.Session) *view {
	theme := request.UserTheme(r)
	return &view{tpl, r, map[string]any{
		"menu":              "",
		"csrf":              request.CSRF(r),
		"flashMessage":      sess.FlashMessage(request.FlashMessage(r)),
		"flashErrorMessage": sess.FlashErrorMessage(request.FlashErrorMessage(r)),
		"theme":             theme,
		"language":          request.UserLanguage(r),
		"theme_checksum":    static.StylesheetBundles[theme].Checksum,
		"app_js_checksum":   static.JavascriptBundles["app"].Checksum,
		"sw_js_checksum":    static.JavascriptBundles["service-worker"].Checksum,
		"webAuthnEnabled":   config.Opts.WebAuthn(),
	}}
}
