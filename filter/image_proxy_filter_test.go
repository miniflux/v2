// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestProxyFilterWithHttp(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttps(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}
