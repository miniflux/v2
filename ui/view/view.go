// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package view // import "miniflux.app/ui/view"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/template"
	"miniflux.app/ui/session"
	"miniflux.app/ui/static"
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
	return v.tpl.Render(template, request.UserLanguage(v.r), v.params)
}

// New returns a new view with default parameters.
func New(tpl *template.Engine, r *http.Request, sess *session.Session) *View {
	b := &View{tpl, r, make(map[string]interface{})}
	theme := request.UserTheme(r)
	view := request.UserView(r)
	b.params["menu"] = ""
	b.params["csrf"] = request.CSRF(r)
	b.params["flashMessage"] = sess.FlashMessage(request.FlashMessage(r))
	b.params["flashErrorMessage"] = sess.FlashErrorMessage(request.FlashErrorMessage(r))
	b.params["theme"] = theme
	b.params["view"] = view
	b.params["theme_checksum"] = static.StylesheetsChecksums[theme]
	b.params["app_js_checksum"] = static.JavascriptsChecksums["app"]
	b.params["sw_js_checksum"] = static.JavascriptsChecksums["sw"]
	return b
}
