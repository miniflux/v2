// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"miniflux.app/v2/internal/locale"

	"github.com/gorilla/mux"
)

//go:embed templates/common/*.html
var commonTemplateFiles embed.FS

//go:embed templates/views/*.html
var viewTemplateFiles embed.FS

// Engine handles the templating system.
type Engine struct {
	templates map[string]*template.Template
	funcMap   *funcMap
}

// NewEngine returns a new template engine.
func NewEngine(router *mux.Router) *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcMap:   &funcMap{router},
	}
}

func (e *Engine) ParseTemplates() {
	funcMap := e.funcMap.Map()
	templates := map[string][]string{ // this isn't a global variable so that it can be garbage-collected.
		"about.html":               {"layout.html", "settings_menu.html"},
		"add_subscription.html":    {"feed_menu.html", "layout.html", "settings_menu.html"},
		"api_keys.html":            {"layout.html", "settings_menu.html"},
		"starred_entries.html":     {"item_meta.html", "layout.html", "pagination.html"},
		"categories.html":          {"layout.html"},
		"category_entries.html":    {"item_meta.html", "layout.html", "pagination.html"},
		"category_feeds.html":      {"feed_list.html", "layout.html"},
		"choose_subscription.html": {"feed_menu.html", "layout.html"},
		"create_api_key.html":      {"layout.html", "settings_menu.html"},
		"create_category.html":     {"layout.html"},
		"create_user.html":         {"layout.html", "settings_menu.html"},
		"edit_category.html":       {"layout.html", "settings_menu.html"},
		"edit_feed.html":           {"layout.html"},
		"edit_user.html":           {"layout.html", "settings_menu.html"},
		"entry.html":               {"layout.html"},
		"feed_entries.html":        {"item_meta.html", "layout.html", "pagination.html"},
		"feeds.html":               {"feed_list.html", "feed_menu.html", "item_meta.html", "layout.html", "pagination.html"},
		"history_entries.html":     {"item_meta.html", "layout.html", "pagination.html"},
		"import.html":              {"feed_menu.html", "layout.html"},
		"integrations.html":        {"layout.html", "settings_menu.html"},
		"login.html":               {"layout.html"},
		"offline.html":             {},
		"search.html":              {"item_meta.html", "layout.html", "pagination.html"},
		"sessions.html":            {"layout.html", "settings_menu.html"},
		"settings.html":            {"layout.html", "settings_menu.html"},
		"shared_entries.html":      {"layout.html", "pagination.html"},
		"tag_entries.html":         {"item_meta.html", "layout.html", "pagination.html"},
		"unread_entries.html":      {"item_meta.html", "layout.html", "pagination.html"},
		"users.html":               {"layout.html", "settings_menu.html"},
		"webauthn_rename.html":     {"layout.html"},
	}

	for name, dependencies := range templates {
		tpl := template.New("").Funcs(funcMap)
		for _, dependency := range dependencies {
			template.Must(tpl.ParseFS(commonTemplateFiles, "templates/common/"+dependency))
		}
		e.templates[name] = template.Must(tpl.ParseFS(viewTemplateFiles, "templates/views/"+name))
	}

	// Sanity check to ensure that all templates are correctly declared in `templates`.
	if entries, err := viewTemplateFiles.ReadDir("templates/views"); err == nil {
		for _, entry := range entries {
			if _, ok := e.templates[entry.Name()]; !ok {
				panic("Template " + entry.Name() + " isn't declared in ParseTemplates")
			}
		}
	} else {
		panic("Unable to read all embedded views templates")
	}
}

// Render process a template.
func (e *Engine) Render(name string, data map[string]any) []byte {
	tpl, ok := e.templates[name]
	if !ok {
		panic("The template " + name + " does not exists.")
	}

	printer := locale.NewPrinter(data["language"].(string))

	// Functions that need to be declared at runtime.
	tpl.Funcs(template.FuncMap{
		"elapsed": func(timezone string, t time.Time) string {
			return elapsedTime(printer, timezone, t)
		},
		"t":      printer.Printf,
		"plural": printer.Plural,
	})

	var b bytes.Buffer
	if err := tpl.ExecuteTemplate(&b, "base", data); err != nil {
		panic(err)
	}

	return b.Bytes()
}
