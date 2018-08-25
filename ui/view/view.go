// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package view // import "miniflux.app/ui/view"

import (
	"miniflux.app/http/context"
	"miniflux.app/template"
	"miniflux.app/ui/session"
	"miniflux.app/ui/static"
)

// View wraps template argument building.
type View struct {
	tpl    *template.Engine
	ctx    *context.Context
	params map[string]interface{}
}

// Set adds a new template argument.
func (v *View) Set(param string, value interface{}) *View {
	v.params[param] = value
	return v
}

// Render executes the template with arguments.
func (v *View) Render(template string) []byte {
	return v.tpl.Render(template, v.ctx.UserLanguage(), v.params)
}

// New returns a new view with default parameters.
func New(tpl *template.Engine, ctx *context.Context, sess *session.Session) *View {
	b := &View{tpl, ctx, make(map[string]interface{})}
	theme := ctx.UserTheme()
	b.params["menu"] = ""
	b.params["csrf"] = ctx.CSRF()
	b.params["flashMessage"] = sess.FlashMessage()
	b.params["flashErrorMessage"] = sess.FlashErrorMessage()
	b.params["theme"] = theme
	b.params["theme_checksum"] = static.StylesheetsChecksums[theme]
	b.params["app_js_checksum"] = static.JavascriptsChecksums["app"]
	b.params["sw_js_checksum"] = static.JavascriptsChecksums["sw"]
	return b
}
