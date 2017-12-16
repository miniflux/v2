// Copyright 2017 Frédéric Guilloe. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"html/template"
	"io"
	"net/mail"
	"strings"
	"time"

	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/logger"
	"github.com/miniflux/miniflux/server/route"
	"github.com/miniflux/miniflux/server/template/helper"
	"github.com/miniflux/miniflux/server/ui/filter"
	"github.com/miniflux/miniflux/url"

	"github.com/gorilla/mux"
)

// Engine handles the templating system.
type Engine struct {
	templates     map[string]*template.Template
	router        *mux.Router
	translator    *locale.Translator
	currentLocale *locale.Language
	cfg           *config.Config
}

func (e *Engine) parseAll() {
	funcMap := template.FuncMap{
		"baseURL": func() string {
			return e.cfg.Get("BASE_URL", config.DefaultBaseURL)
		},
		"hasOAuth2Provider": func(provider string) bool {
			return e.cfg.Get("OAUTH2_PROVIDER", "") == provider
		},
		"hasKey": func(dict map[string]string, key string) bool {
			if value, found := dict[key]; found {
				return value != ""
			}
			return false
		},
		"route": func(name string, args ...interface{}) string {
			return route.Path(e.router, name, args...)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return filter.ImageProxyFilter(e.router, data)
		},
		"proxyURL": func(link string) string {
			if url.IsHTTPS(link) {
				return link
			}

			return filter.Proxify(e.router, link)
		},
		"domain": func(websiteURL string) string {
			return url.Domain(websiteURL)
		},
		"isEmail": func(str string) bool {
			_, err := mail.ParseAddress(str)
			if err != nil {
				return false
			}
			return true
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
			return helper.GetElapsedTime(e.currentLocale, ts)
		},
		"t": func(key interface{}, args ...interface{}) string {
			switch key.(type) {
			case string:
				return e.currentLocale.Get(key.(string), args...)
			case errors.LocalizedError:
				err := key.(errors.LocalizedError)
				return err.Localize(e.currentLocale)
			case error:
				return key.(error).Error()
			default:
				return ""
			}
		},
		"plural": func(key string, n int, args ...interface{}) string {
			return e.currentLocale.Plural(key, n, args...)
		},
	}

	commonTemplates := ""
	for _, content := range templateCommonMap {
		commonTemplates += content
	}

	for name, content := range templateViewsMap {
		logger.Debug("[Template] Parsing: %s", name)
		e.templates[name] = template.Must(template.New("main").Funcs(funcMap).Parse(commonTemplates + content))
	}
}

// SetLanguage change the language for template processing.
func (e *Engine) SetLanguage(language string) {
	e.currentLocale = e.translator.GetLanguage(language)
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

// NewEngine returns a new template Engine.
func NewEngine(cfg *config.Config, router *mux.Router, translator *locale.Translator) *Engine {
	tpl := &Engine{
		templates:  make(map[string]*template.Template),
		router:     router,
		translator: translator,
		cfg:        cfg,
	}

	tpl.parseAll()
	return tpl
}
