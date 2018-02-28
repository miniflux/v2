// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template

import (
	"html/template"
	"net/mail"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/miniflux/miniflux/config"
	"github.com/miniflux/miniflux/errors"
	"github.com/miniflux/miniflux/filter"
	"github.com/miniflux/miniflux/http/route"
	"github.com/miniflux/miniflux/locale"
	"github.com/miniflux/miniflux/url"
)

type funcMap struct {
	cfg      *config.Config
	router   *mux.Router
	Language *locale.Language
}

func (f *funcMap) Map() template.FuncMap {
	return template.FuncMap{
		"baseURL": func() string {
			return f.cfg.BaseURL()
		},
		"rootURL": func() string {
			return f.cfg.RootURL()
		},
		"hasOAuth2Provider": func(provider string) bool {
			return f.cfg.OAuth2Provider() == provider
		},
		"hasKey": func(dict map[string]string, key string) bool {
			if value, found := dict[key]; found {
				return value != ""
			}
			return false
		},
		"route": func(name string, args ...interface{}) string {
			return route.Path(f.router, name, args...)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return filter.ImageProxyFilter(f.router, data)
		},
		"proxyURL": func(link string) string {
			if url.IsHTTPS(link) {
				return link
			}

			return filter.Proxify(f.router, link)
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
		"elapsed": func(timezone string, t time.Time) string {
			return elapsedTime(f.Language, timezone, t)
		},
		"t": func(key interface{}, args ...interface{}) string {
			switch key.(type) {
			case string:
				return f.Language.Get(key.(string), args...)
			case errors.LocalizedError:
				return key.(errors.LocalizedError).Localize(f.Language)
			case *errors.LocalizedError:
				return key.(*errors.LocalizedError).Localize(f.Language)
			case error:
				return key.(error).Error()
			default:
				return ""
			}
		},
		"plural": func(key string, n int, args ...interface{}) string {
			return f.Language.Plural(key, n, args...)
		},
		"dict": dict,
	}
}

func newFuncMap(cfg *config.Config, router *mux.Router, language *locale.Language) *funcMap {
	return &funcMap{cfg, router, language}
}
