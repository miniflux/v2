// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"time"

	"miniflux.app/v2/internal/locale"

	"github.com/gorilla/mux"
)

//go:embed templates/common/*.html
var commonTemplateFiles embed.FS

//go:embed templates/views/*.html
var viewTemplateFiles embed.FS

//go:embed templates/standalone/*.html
var standaloneTemplateFiles embed.FS

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

// ParseTemplates parses template files embed into the application.
func (e *Engine) ParseTemplates() error {
	funcMap := e.funcMap.Map()
	commonTemplates := template.Must(template.New("").Funcs(funcMap).ParseFS(commonTemplateFiles, "templates/common/*.html"))

	dirEntries, err := viewTemplateFiles.ReadDir("templates/views")
	if err != nil {
		return err
	}
	for _, dirEntry := range dirEntries {
		fullName := "templates/views/" + dirEntry.Name()
		slog.Debug("Parsing template", slog.String("template_name", fullName))
		commonTemplatesClone, err := commonTemplates.Clone()
		if err != nil {
			panic("Unable to clone the common template")
		}
		e.templates[dirEntry.Name()] = template.Must(commonTemplatesClone.ParseFS(viewTemplateFiles, fullName))
	}

	dirEntries, err = standaloneTemplateFiles.ReadDir("templates/standalone")
	if err != nil {
		return err
	}
	for _, dirEntry := range dirEntries {
		fullName := "templates/standalone/" + dirEntry.Name()
		slog.Debug("Parsing template", slog.String("template_name", fullName))
		e.templates[dirEntry.Name()] = template.Must(template.New(dirEntry.Name()).Funcs(funcMap).ParseFS(standaloneTemplateFiles, fullName))
	}

	return nil
}

// Render process a template.
func (e *Engine) Render(name string, data map[string]any) []byte {
	tpl, ok := e.templates[name]
	if !ok {
		panic("This template does not exists: " + name)
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
	err := tpl.ExecuteTemplate(&b, "base", data)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}
