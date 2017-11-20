// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/miniflux/miniflux2/errors"
	"github.com/miniflux/miniflux2/locale"
	"github.com/miniflux/miniflux2/server/route"
	"github.com/miniflux/miniflux2/server/template/helper"
	"github.com/miniflux/miniflux2/server/ui/filter"

	"github.com/gorilla/mux"
)

type TemplateEngine struct {
	templates     map[string]*template.Template
	router        *mux.Router
	translator    *locale.Translator
	currentLocale *locale.Language
}

func (t *TemplateEngine) ParseAll() {
	funcMap := template.FuncMap{
		"route": func(name string, args ...interface{}) string {
			return route.GetRoute(t.router, name, args...)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return filter.ImageProxyFilter(t.router, data)
		},
		"domain": func(websiteURL string) string {
			parsedURL, err := url.Parse(websiteURL)
			if err != nil {
				return websiteURL
			}

			return parsedURL.Host
		},
		"hasPrefix": func(str, prefix string) bool {
			return strings.HasPrefix(str, prefix)
		},
		"contains": func(str, substr string) bool {
			return strings.Contains(str, substr)
		},
		"isodate": func(ts time.Time) string {
			return ts.Format("2006-01-02 15:04:05")
		},
		"elapsed": func(ts time.Time) string {
			return helper.GetElapsedTime(t.currentLocale, ts)
		},
		"t": func(key interface{}, args ...interface{}) string {
			switch key.(type) {
			case string:
				return t.currentLocale.Get(key.(string), args...)
			case errors.LocalizedError:
				err := key.(errors.LocalizedError)
				return err.Localize(t.currentLocale)
			case error:
				return key.(error).Error()
			default:
				return ""
			}
		},
		"plural": func(key string, n int, args ...interface{}) string {
			return t.currentLocale.Plural(key, n, args...)
		},
	}

	commonTemplates := ""
	for _, content := range templateCommonMap {
		commonTemplates += content
	}

	for name, content := range templateViewsMap {
		log.Println("Parsing template:", name)
		t.templates[name] = template.Must(template.New("main").Funcs(funcMap).Parse(commonTemplates + content))
	}
}

func (t *TemplateEngine) SetLanguage(language string) {
	t.currentLocale = t.translator.GetLanguage(language)
}

func (t *TemplateEngine) Execute(w io.Writer, name string, data interface{}) {
	tpl, ok := t.templates[name]
	if !ok {
		log.Fatalf("The template %s does not exists.\n", name)
	}

	var b bytes.Buffer
	err := tpl.ExecuteTemplate(&b, "base", data)
	if err != nil {
		log.Fatalf("Unable to render template: %v\n", err)
	}

	b.WriteTo(w)
}

func NewTemplateEngine(router *mux.Router, translator *locale.Translator) *TemplateEngine {
	tpl := &TemplateEngine{
		templates:  make(map[string]*template.Template),
		router:     router,
		translator: translator,
	}

	tpl.ParseAll()
	return tpl
}
