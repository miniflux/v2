// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"html/template"
	"io"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"

	"github.com/gorilla/mux"
)

// Engine handles the templating system.
type Engine struct {
	templates  map[string]*template.Template
	translator *locale.Translator
	funcMap    *funcMap
}

func (e *Engine) parseAll() {
	commonTemplates := ""
	for _, content := range templateCommonMap {
		commonTemplates += content
	}

	for name, content := range templateViewsMap {
		logger.Debug("[Template] Parsing: %s", name)
		e.templates[name] = template.Must(template.New("main").Funcs(e.funcMap.Map()).Parse(commonTemplates + content))
	}
}

// SetLanguage change the language for template processing.
func (e *Engine) SetLanguage(language string) {
	e.funcMap.Language = e.translator.GetLanguage(language)
}

// Execute process a template.
func (e *Engine) Execute(w io.Writer, name string, data interface{}) {
	tpl, ok := e.templates[name]
	if !ok {
		logger.Fatal("[Template] The template %s does not exists", name)
	}

	var b bytes.Buffer
	err := tpl.ExecuteTemplate(&b, "base", data)
	if err != nil {
		logger.Fatal("[Template] Unable to render template: %v", err)
	}

	b.WriteTo(w)
}

// NewEngine returns a new template engine.
func NewEngine(cfg *config.Config, router *mux.Router, translator *locale.Translator) *Engine {
	tpl := &Engine{
		templates:  make(map[string]*template.Template),
		translator: translator,
		funcMap:    newFuncMap(cfg, router, translator.GetLanguage("en_US")),
	}

	tpl.parseAll()
	return tpl
}
