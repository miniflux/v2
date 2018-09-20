// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package template // import "miniflux.app/template"

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	fm := funcMap{}
	if f, ok := fm.Map()["truncate"]; ok {
		if truncate := f.(func(str string, max int) string); ok {
			shortEnglishText := "Short text"
			shortUnicodeText := "Короткий текст"

			// edge case
			if truncate(shortEnglishText, len(shortEnglishText)) != shortEnglishText {
				t.Fatal("Invalid truncation")
			}
			// real case
			if truncate(shortEnglishText, 25) != shortEnglishText {
				t.Fatal("Invalid truncation")
			}
			if truncate(shortUnicodeText, len(shortUnicodeText)) != shortUnicodeText {
				t.Fatal("Invalid truncation")
			}
			if truncate(shortUnicodeText, 25) != shortUnicodeText {
				t.Fatal("Invalid truncation")
			}

			longEnglishText := "This is really pretty long English text"
			longRussianText := "Это реально очень длинный русский текст"

			if truncate(longEnglishText, 25) != "This is really pretty lon…" {
				t.Fatal("Invalid truncation")
			}
			if truncate(longRussianText, 25) != "Это реально очень длинный…" {
				t.Fatal("Invalid truncation")
			}
		} else {
			t.Fatal("Type assetion for this func is failed, check func, maybe it was changed")
		}
	} else {
		t.Fatal("There is no such function in this map, check key, maybe it was changed")
	}
}
