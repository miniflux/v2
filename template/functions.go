// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"fmt"
	"html/template"
	"math"
	"net/mail"
	"strings"
	"time"

	netUrl "net/url"

	"miniflux.app/config"
	"miniflux.app/http/route"
	"miniflux.app/locale"
	"miniflux.app/model"
	"miniflux.app/proxy"
	"miniflux.app/timezone"
	"miniflux.app/url"

	"github.com/gorilla/mux"
)

type funcMap struct {
	router *mux.Router
}

// Map returns a map of template functions that are compiled during template parsing.
func (f *funcMap) Map() template.FuncMap {
	return template.FuncMap{
		"formatFileSize": formatFileSize,
		"dict":           dict,
		"hasKey":         hasKey,
		"truncate":       truncate,
		"isEmail":        isEmail,
		"buildQuery":     buildQuery,
		"baseURL": func() string {
			return config.Opts.BaseURL()
		},
		"rootURL": func() string {
			return config.Opts.RootURL()
		},
		"hasOAuth2Provider": func(provider string) bool {
			return config.Opts.OAuth2Provider() == provider
		},
		"route": func(name string, args ...interface{}) string {
			return route.Path(f.router, name, args...)
		},
		"safeURL": func(url string) template.URL {
			return template.URL(url)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return proxy.ImageProxyRewriter(f.router, data)
		},
		"proxyURL": func(link string) string {
			proxyImages := config.Opts.ProxyImages()

			if proxyImages == "all" || (proxyImages != "none" && !url.IsHTTPS(link)) {
				return proxy.ProxifyURL(f.router, link)
			}

			return link
		},
		"domain": func(websiteURL string) string {
			return url.Domain(websiteURL)
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
		"theme_color": func(theme string) string {
			return model.ThemeColor(theme)
		},
		"icon": func(iconName string) template.HTML {
			return template.HTML(fmt.Sprintf(
				`<svg class="icon" aria-hidden="true"><use xlink:href="%s#icon-%s"></svg>`,
				route.Path(f.router, "appIcon", "filename", "sprite.svg"),
				iconName,
			))
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

func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict expects an even number of arguments")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func hasKey(dict map[string]string, key string) bool {
	if value, found := dict[key]; found {
		return value != ""
	}
	return false
}

func truncate(str string, max int) string {
	runes := 0
	for i := range str {
		runes++
		if runes > max {
			return str[:i] + "…"
		}
	}
	return str
}

func isEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

func elapsedTime(printer *locale.Printer, tz string, t time.Time) string {
	if t.IsZero() {
		return printer.Printf("time_elapsed.not_yet")
	}

	now := timezone.Now(tz)
	t = timezone.Convert(tz, t)
	if now.Before(t) {
		return printer.Printf("time_elapsed.not_yet")
	}

	diff := now.Sub(t)
	// Duration in seconds
	s := diff.Seconds()
	// Duration in days
	d := int(s / 86400)
	switch {
	case s < 60:
		return printer.Printf("time_elapsed.now")
	case s < 3600:
		minutes := int(diff.Minutes())
		return printer.Plural("time_elapsed.minutes", minutes, minutes)
	case s < 86400:
		hours := int(diff.Hours())
		return printer.Plural("time_elapsed.hours", hours, hours)
	case d == 1:
		return printer.Printf("time_elapsed.yesterday")
	case d < 21:
		return printer.Plural("time_elapsed.days", d, d)
	case d < 31:
		weeks := int(math.Round(float64(d) / 7))
		return printer.Plural("time_elapsed.weeks", weeks, weeks)
	case d < 365:
		months := int(math.Round(float64(d) / 30))
		return printer.Plural("time_elapsed.months", months, months)
	default:
		years := int(math.Round(float64(d) / 365))
		return printer.Plural("time_elapsed.years", years, years)
	}
}

func formatFileSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func buildQuery(args ...interface{}) string {
	vals := netUrl.Values{}
	for i := 0; i < len(args)-1; i += 2 {
		key := args[i].(string)
		switch v := args[i+1].(type) {
		case int, int64:
			if v != 0 {
				vals.Set(key, fmt.Sprintf("%v", v))
			}
		case string:
			if v != "" {
				vals.Set(key, v)
			}
		case bool:
			if v {
				vals.Set(key, "t")
			}
		}
	}

	qs := vals.Encode()
	if qs == "" {
		return ""
	}
	return "?" + qs
}
