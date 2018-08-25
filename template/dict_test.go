// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"testing"
)

func TestDict(t *testing.T) {
	d, err := dict("k1", "v1", "k2", "v2")
	if err != nil {
		t.Fatalf(`The dict should be valid: %v`, err)
	}

	if value, found := d["k1"]; found {
		if value != "v1" {
			t.Fatalf(`Incorrect value for k1: %q`, value)
		}
	}

	if value, found := d["k2"]; found {
		if value != "v2" {
			t.Fatalf(`Incorrect value for k2: %q`, value)
		}
	}
}

func TestDictWithIncorrectNumberOfPairs(t *testing.T) {
	_, err := dict("k1", "v1", "k2")
	if err == nil {
		t.Fatalf(`The dict should not be valid because the number of keys/values pairs are incorrect`)
	}
}

func TestDictWithInvalidKey(t *testing.T) {
	_, err := dict(1, "v1")
	if err == nil {
		t.Fatalf(`The dict should not be valid because the key is not a string`)
	}
}
