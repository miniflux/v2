// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxy // import "miniflux.app/v2/internal/proxy"

import (
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"miniflux.app/v2/internal/config"
)

func TestProxyFilterWithHttpDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "http-only")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsDefault(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "http-only")
	os.Setenv("PROXY_MEDIA_TYPES", "image")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "none")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsNever(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "none")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := input

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsAlways(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="/proxy/LdPNR1GBDigeeNp2ArUQRyZsVqT_PWLfHGjYFrrWWIY=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsAlwaysAndCustomProxyServer(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_URL", "https://proxy-example/proxy")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="https://proxy-example/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "invalid")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithHttpsInvalid(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "invalid")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := ProxyRewriter(r, input)
	expected := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got "%s" instead of "%s"`, output, expected)
	}
}

func TestProxyFilterWithSrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" srcset="http://website/folder/image2.png 656w, http://website/folder/image3.png 360w" alt="test"></p>`
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/QgAmrJWiAud_nNAsz3F8OTxaIofwAiO36EDzH_YfMzo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w" alt="test"/></p>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithEmptySrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<p><img src="http://website/folder/image.png" srcset="" alt="test"></p>`
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="" alt="test"/></p>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="http://website/folder/image2.png 656w,   http://website/folder/image3.png 360w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/QgAmrJWiAud_nNAsz3F8OTxaIofwAiO36EDzH_YfMzo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w, /proxy/ZIw0hv8WhSTls5aSqhnFaCXlUrKIqTnBRaY0-NaLnds=/aHR0cHM6Ly93ZWJzaXRlL3NvbWUsaW1hZ2UucG5n 2x"/></picture>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterOnlyNonHTTPWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "https")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="http://website/folder/image2.png 656w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, https://website/some,image.png 2x"/></picture>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyWithImageDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<img src="data:image/gif;base64,test">`
	expected := `<img src="data:image/gif;base64,test"/>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyWithImageSourceDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	expected := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithVideo(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "video")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<video poster="https://example.com/img.png" src="https://example.com/video.mp4"></video>`
	expected := `<video poster="/proxy/aDFfroYL57q5XsojIzATT6OYUCkuVSPXYJQAVrotnLw=/aHR0cHM6Ly9leGFtcGxlLmNvbS9pbWcucG5n" src="/proxy/0y3LR8zlx8S8qJkj1qWFOO6x3a-5yf2gLWjGIJV5yyc=/aHR0cHM6Ly9leGFtcGxlLmNvbS92aWRlby5tcDQ="></video>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterVideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_OPTION", "all")
	os.Setenv("PROXY_MEDIA_TYPES", "image")
	os.Setenv("PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/proxy/{encodedDigest}/{encodedURL}", func(w http.ResponseWriter, r *http.Request) {}).Name("proxy")

	input := `<video poster="https://example.com/img.png" src="https://example.com/video.mp4"></video>`
	expected := `<video poster="/proxy/aDFfroYL57q5XsojIzATT6OYUCkuVSPXYJQAVrotnLw=/aHR0cHM6Ly9leGFtcGxlLmNvbS9pbWcucG5n" src="https://example.com/video.mp4"></video>`
	output := ProxyRewriter(r, input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}
