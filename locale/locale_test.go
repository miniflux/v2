// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/locale"

import "testing"

func TestAvailableLanguages(t *testing.T) {
	results := AvailableLanguages()
	for k, v := range results {
		if k == "" {
			t.Errorf(`Empty language key detected`)
		}

		if v == "" {
			t.Errorf(`Empty language value detected`)
		}
	}

	if _, found := results["en_US"]; !found {
		t.Errorf(`We must have at least the default language (en_US)`)
	}
}
