// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"miniflux.app/config"
)

func TestProxyFilterWithHttpDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "http-only")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "http-only")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "none")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "none")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsAlwaysAndCustomProxyServer(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")
	os.Setenv("PROXY_IMAGE_URL", "https://proxy-example/proxy")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="https://proxy-example/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "invalid")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "invalid")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ImageProxyRewriter(r, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithSrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" srcset="http://website/folder/image2.png 656w, http://website/folder/image3.png 360w" alt="test"></p>`
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w" alt="test"/></p>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithEmptySrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" srcset="" alt="test"></p>`
	expected := `<p><img src="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="" alt="test"/></p>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="http://website/folder/image2.png 656w,   http://website/folder/image3.png 360w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w, /proxy/aHR0cHM6Ly93ZWJzaXRlL3NvbWUsaW1hZ2UucG5n 2x"/></picture>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterOnlyNonHTTPWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "https")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="http://website/folder/image2.png 656w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, https://website/some,image.png 2x"/></picture>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestImageProxyWithImageDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<img src="data:image/gif;base64,test">`
	expected := `<img src="data:image/gif;base64,test"/>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestImageProxyWithImageSourceDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	expected := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	output := ImageProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}
