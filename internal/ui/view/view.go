// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package view // import "influxeed-engine/v2/internal/ui/view"

import (
	"net/http"

	"influxeed-engine/v2/internal/config"
	"influxeed-engine/v2/internal/http/request"
	"influxeed-engine/v2/internal/template"
	"influxeed-engine/v2/internal/ui/session"
	"influxeed-engine/v2/internal/ui/static"
)

// View wraps template argument building.
type View struct {
	tpl    *template.Engine
	r      *http.Request
	params map[string]any
}

// Set adds a new template argument.
func (v *View) Set(param string, value any) *View {
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
	return &View{tpl, r, map[string]any{
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
