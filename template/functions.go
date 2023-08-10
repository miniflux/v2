// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/template"

import (
	"fmt"
	"html/template"
	"math"
	"net/mail"
	"strings"
	"time"

	"miniflux.app/v2/config"
	"miniflux.app/v2/crypto"
	"miniflux.app/v2/http/route"
	"miniflux.app/v2/locale"
	"miniflux.app/v2/model"
	"miniflux.app/v2/proxy"
	"miniflux.app/v2/timezone"
	"miniflux.app/v2/url"

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
		"baseURL": func() string {
			return config.Opts.BaseURL()
		},
		"rootURL": func() string {
			return config.Opts.RootURL()
		},
		"hasOAuth2Provider": func(provider string) bool {
			return config.Opts.OAuth2Provider() == provider
		},
		"hasAuthProxy": func() bool {
			return config.Opts.AuthProxyHeader() != ""
		},
		"route": func(name string, args ...interface{}) string {
			return route.Path(f.router, name, args...)
		},
		"safeURL": func(url string) template.URL {
			return template.URL(url)
		},
		"safeCSS": func(str string) template.CSS {
			return template.CSS(str)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return proxy.ProxyRewriter(f.router, data)
		},
		"proxyURL": func(link string) string {
			proxyOption := config.Opts.ProxyOption()

			if proxyOption == "all" || (proxyOption != "none" && !url.IsHTTPS(link)) {
				return proxy.ProxifyURL(f.router, link)
			}

			return link
		},
		"mustBeProxyfied": func(mediaType string) bool {
			for _, t := range config.Opts.ProxyMediaTypes() {
				if t == mediaType {
					return true
				}
			}
			return false
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
		"replace": func(str, old, new string) string {
			return strings.Replace(str, old, new, 1)
		},
		"isodate": func(ts time.Time) string {
			return ts.Format("2006-01-02 15:04:05")
		},
		"theme_color": func(theme, colorScheme string) string {
			return model.ThemeColor(theme, colorScheme)
		},
		"icon": func(iconName string) template.HTML {
			return template.HTML(fmt.Sprintf(
				`<svg class="icon" aria-hidden="true"><use xlink:href="%s#icon-%s"/></svg>`,
				route.Path(f.router, "appIcon", "filename", "sprite.svg"),
				iconName,
			))
		},
		"nonce": func() string {
			return crypto.GenerateRandomStringHex(16)
		},
		"deRef": func(i *int) int { return *i },

		// These functions are overrode at runtime after the parsing.
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
