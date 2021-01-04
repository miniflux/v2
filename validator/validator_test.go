// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package validator // import "miniflux.app/validator"

import "testing"

func TestIsValidURL(t *testing.T) {
	scenarios := map[string]bool{
		"https://www.example.org": true,
		"http://www.example.org/": true,
		"www.example.org":         false,
	}

	for link, expected := range scenarios {
		result := isValidURL(link)
		if result != expected {
			t.Errorf(`Unexpected result, got %v instead of %v`, result, expected)
		}
	}
}
