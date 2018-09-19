// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"bytes"
	"html/template"
	"net/mail"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"miniflux.app/config"
	"miniflux.app/filter"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/url"
)

type funcMap struct {
	cfg    *config.Config
	router *mux.Router
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
			return filter.ImageProxyFilter(f.router, f.cfg, data)
		},
		"proxyURL": func(link string) string {
			proxyImages := f.cfg.ProxyImages()

			if proxyImages == "all" || (proxyImages != "none" && !url.IsHTTPS(link)) {
				return filter.Proxify(f.router, link)
			}

			return link
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
		"dict": dict,
		"truncate": func(s string, maxRunes int) string {
			var numRunes = 0
			for i := range s {
				numRunes++
				if numRunes > maxRunes {
					return s[:i] + "…"
				}
			}
			return s
		},
		"theme_color": func(theme string) string {
			return model.ThemeColor(theme)
		},

		// These functions are overrided at runtime after the parsing.
		"elapsed": func(timezone string, t time.Time) string {
			return ""
		},
		"t": func(key interface{}, args ...interface{}) string {
			return ""
		},
		"plural": func(key string, n int, args ...interface{}) string {
			return ""
		},
	}
}

func newFuncMap(cfg *config.Config, router *mux.Router) *funcMap {
	return &funcMap{cfg, router}
}
