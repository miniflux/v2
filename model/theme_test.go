// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import "testing"

func TestValidateTheme(t *testing.T) {
	for _, status := range []string{"default", "black", "sansserif"} {
		if err := ValidateTheme(status); err != nil {
			t.Error(`A valid theme should not generate any error`)
		}
	}

	if err := ValidateTheme("invalid"); err == nil {
		t.Error(`An invalid theme should generate a error`)
	}
}
