// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package filter // import "miniflux.app/filter"

import (
	"net/http"
	"os"
	"testing"

	"miniflux.app/config"

	"github.com/gorilla/mux"
)

func TestProxyFilterWithHttpDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "http-only")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "http-only")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "none")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "none")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "invalid")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "invalid")
	c := config.NewConfig()

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyFilter(r, c, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}
