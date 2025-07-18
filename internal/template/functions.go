// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package template // import "miniflux.app/v2/internal/template"

import (
	"fmt"
	"html/template"
	"math"
	"net/mail"
	"net/url"
	"slices"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/locale"
	"miniflux.app/v2/internal/mediaproxy"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/timezone"
	"miniflux.app/v2/internal/urllib"

	"github.com/gorilla/mux"
)

type funcMap struct {
	router *mux.Router
}

// Map returns a map of template functions that are compiled during template parsing.
func (f *funcMap) Map() template.FuncMap {
	return template.FuncMap{
		"formatFileSize":   formatFileSize,
		"dict":             dict,
		"truncate":         truncate,
		"isEmail":          isEmail,
		"baseURL":          config.Opts.BaseURL,
		"rootURL":          config.Opts.RootURL,
		"disableLocalAuth": config.Opts.DisableLocalAuth,
		"oidcProviderName": config.Opts.OIDCProviderName,
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
		"safeJS": func(str string) template.JS {
			return template.JS(str)
		},
		"noescape": func(str string) template.HTML {
			return template.HTML(str)
		},
		"proxyFilter": func(data string) string {
			return mediaproxy.RewriteDocumentWithRelativeProxyURL(f.router, data)
		},
		"proxyURL": func(link string) string {
			mediaProxyMode := config.Opts.MediaProxyMode()

			if mediaProxyMode == "all" || (mediaProxyMode != "none" && !urllib.IsHTTPS(link)) {
				return mediaproxy.ProxifyRelativeURL(f.router, link)
			}

			return link
		},
		"mustBeProxyfied": func(mediaType string) bool {
			return slices.Contains(config.Opts.MediaProxyResourceTypes(), mediaType)
		},
		"domain": urllib.Domain,
		"replace": func(str, old, new string) string {
			return strings.Replace(str, old, new, 1)
		},
		"isodate": func(ts time.Time) string {
			return ts.Format("2006-01-02 15:04:05")
		},
		"theme_color": model.ThemeColor,
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
		"deRef":     func(i *int) int { return *i },
		"duration":  duration,
		"urlEncode": url.PathEscape,
		"subtract": func(a, b int) int {
			return a - b
		},

		// These functions are overridden at runtime after parsing.
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

func truncate(str string, max int) string {
	if runes := []rune(str); len(runes) > max {
		return string(runes[:max]) + "…"
	}
	return str
}

func isEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

// Returns the duration in human readable format (hours and minutes).
func duration(t time.Time) string {
	return durationImpl(t, time.Now())
}

// Accepts now argument for easy testing
func durationImpl(t time.Time, now time.Time) string {
	if t.IsZero() {
		return ""
	}

	if diff := t.Sub(now); diff >= 0 {
		// Round to nearest second to get e.g. "14m56s" rather than "14m56.245483933s"
		return diff.Round(time.Second).String()
	}
	return ""
}

func elapsedTime(printer *locale.Printer, tz string, t time.Time) string {
	if t.IsZero() {
		return printer.Print("time_elapsed.not_yet")
	}

	now := timezone.Now(tz)
	t = timezone.Convert(tz, t)
	if now.Before(t) {
		return printer.Print("time_elapsed.not_yet")
	}

	diff := now.Sub(t)
	// Duration in seconds
	s := diff.Seconds()
	// Duration in days
	d := int(s / 86400)
	switch {
	case s < 60:
		return printer.Print("time_elapsed.now")
	case s < 3600:
		minutes := int(diff.Minutes())
		return printer.Plural("time_elapsed.minutes", minutes, minutes)
	case s < 86400:
		hours := int(diff.Hours())
		return printer.Plural("time_elapsed.hours", hours, hours)
	case d == 1:
		return printer.Print("time_elapsed.yesterday")
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
	base := math.Log(float64(b)) / math.Log(unit)
	number := math.Pow(unit, base-math.Floor(base))
	return fmt.Sprintf("%.1f %ciB", number, "KMGTPE"[int64(base)-1])
}
