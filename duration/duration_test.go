// Copyright (c) 2017 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package duration

import (
	"fmt"
	"testing"
	"time"

	"github.com/miniflux/miniflux/locale"
)

func TestElapsedTime(t *testing.T) {
	var dt = []struct {
		in  time.Time
		out string
	}{
		{time.Time{}, NotYet},
		{time.Now().Add(time.Hour), NotYet},
		{time.Now(), JustNow},
		{time.Now().Add(-time.Minute), LastMinute},
		{time.Now().Add(-time.Minute * 40), fmt.Sprintf(Minutes, 40)},
		{time.Now().Add(-time.Hour), LastHour},
		{time.Now().Add(-time.Hour * 3), fmt.Sprintf(Hours, 3)},
		{time.Now().Add(-time.Hour * 32), Yesterday},
		{time.Now().Add(-time.Hour * 24 * 3), fmt.Sprintf(Days, 3)},
		{time.Now().Add(-time.Hour * 24 * 14), fmt.Sprintf(Weeks, 2)},
		{time.Now().Add(-time.Hour * 24 * 60), fmt.Sprintf(Months, 2)},
		{time.Now().Add(-time.Hour * 24 * 365 * 3), fmt.Sprintf(Years, 3)},
	}
	for i, tt := range dt {
		if out := ElapsedTime(&locale.Language{}, tt.in); out != tt.out {
			t.Errorf("%d. content mismatch for %v:exp=%q got=%q", i, tt.in, tt.out, out)
		}
	}
}
