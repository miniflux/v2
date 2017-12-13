// Copyright (c) 2017 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package helper

import (
	"math"
	"time"

	"github.com/miniflux/miniflux/locale"
)

// Texts to be translated if necessary.
var (
	NotYet     = `not yet`
	JustNow    = `just now`
	LastMinute = `1 minute ago`
	Minutes    = `%d minutes ago`
	LastHour   = `1 hour ago`
	Hours      = `%d hours ago`
	Yesterday  = `yesterday`
	Days       = `%d days ago`
	Weeks      = `%d weeks ago`
	Months     = `%d months ago`
	Years      = `%d years ago`
)

// GetElapsedTime returns in a human readable format the elapsed time
// since the given datetime.
func GetElapsedTime(translator *locale.Language, t time.Time) string {
	if t.IsZero() || time.Now().Before(t) {
		return translator.Get(NotYet)
	}
	diff := time.Since(t)
	// Duration in seconds
	s := diff.Seconds()
	// Duration in days
	d := int(s / 86400)
	switch {
	case s < 60:
		return translator.Get(JustNow)
	case s < 120:
		return translator.Get(LastMinute)
	case s < 3600:
		return translator.Get(Minutes, int(diff.Minutes()))
	case s < 7200:
		return translator.Get(LastHour)
	case s < 86400:
		return translator.Get(Hours, int(diff.Hours()))
	case d == 1:
		return translator.Get(Yesterday)
	case d < 7:
		return translator.Get(Days, d)
	case d < 31:
		return translator.Get(Weeks, int(math.Ceil(float64(d)/7)))
	case d < 365:
		return translator.Get(Months, int(math.Ceil(float64(d)/30)))
	default:
		return translator.Get(Years, int(math.Ceil(float64(d)/365)))
	}
}
