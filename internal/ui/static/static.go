// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package static // import "miniflux.app/v2/internal/ui/static"

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"miniflux.app/v2/internal/crypto"

	"github.com/andybalholm/brotli"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

const licensePrefix = "//@license magnet:?xt=urn:btih:8e4f440f4c65981c5bf93c76d35135ba5064d8b7&dn=apache-2.0.txt Apache-2.0\n"
const licenseSuffix = "\n//@license-end"

type asset struct {
	Data       []byte
	Checksum   string
	BrotliData []byte
	GzipData   []byte
}

// Negotiate selects the best pre-compressed representation of the
// asset based on the Accept-Encoding request header value.  It returns
// the bytes to write and the Content-Encoding value ("br", "gzip", or
// "" for identity).
func (a asset) Negotiate(acceptEncoding string) (body []byte, encoding string) {
	switch {
	case a.BrotliData != nil && strings.Contains(acceptEncoding, "br"):
		return a.BrotliData, "br"
	case a.GzipData != nil && strings.Contains(acceptEncoding, "gzip"):
		return a.GzipData, "gzip"
	default:
		return a.Data, ""
	}
}

// Static assets.
var (
	StylesheetBundles map[string]asset
	JavascriptBundles map[string]asset
	BinaryBundles     map[string]asset
)

//go:embed bin/*
var binaryFiles embed.FS

//go:embed css/*.css
var stylesheetFiles embed.FS

//go:embed js/*.js
var javascriptFiles embed.FS

// precompress produces brotli and gzip compressed variants of data.
// Best compression levels are used because this only runs once at
// startup; the resulting byte slices are served directly on every
// request, avoiding any per-request compression work.
func precompress(data []byte) (brotliData, gzipData []byte) {
	var br bytes.Buffer
	bw := brotli.NewWriterV2(&br, brotli.BestCompression)
	bw.Write(data)
	bw.Close()

	var gz bytes.Buffer
	gw, _ := gzip.NewWriterLevel(&gz, gzip.BestCompression)
	gw.Write(data)
	gw.Close()

	return br.Bytes(), gz.Bytes()
}

func GenerateBinaryBundles() error {
	dirEntries, err := binaryFiles.ReadDir("bin")
	if err != nil {
		return err
	}
	BinaryBundles = make(map[string]asset, len(dirEntries))

	minifier := minify.New()
	minifier.Add("image/svg+xml", &svg.Minifier{
		KeepComments: true, // needed to keep the license
	})

	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		data, err := binaryFiles.ReadFile("bin/" + name)
		if err != nil {
			return err
		}

		if strings.HasSuffix(name, ".svg") {
			// minifier.Bytes returns the data unchanged in case of error.
			data, err = minifier.Bytes("image/svg+xml", data)
			if err != nil {
				slog.Error("Unable to minimize the svg file", slog.String("filename", name), slog.Any("error", err))
			}
		}

		a := asset{
			Data:     data,
			Checksum: crypto.HashFromBytes(data),
		}

		if strings.HasSuffix(name, ".svg") {
			a.BrotliData, a.GzipData = precompress(data)
		}

		BinaryBundles[name] = a
	}

	return nil
}

// GenerateStylesheetsBundles creates CSS bundles.
func GenerateStylesheetsBundles() error {
	var bundles = map[string][]string{
		"light_serif":       {"css/light.css", "css/serif.css", "css/common.css"},
		"light_sans_serif":  {"css/light.css", "css/sans_serif.css", "css/common.css"},
		"dark_serif":        {"css/dark.css", "css/serif.css", "css/common.css"},
		"dark_sans_serif":   {"css/dark.css", "css/sans_serif.css", "css/common.css"},
		"system_serif":      {"css/system.css", "css/serif.css", "css/common.css"},
		"system_sans_serif": {"css/system.css", "css/sans_serif.css", "css/common.css"},
	}

	StylesheetBundles = make(map[string]asset, len(bundles))

	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)

	for bundleName, srcFiles := range bundles {
		var buffer bytes.Buffer

		for _, srcFile := range srcFiles {
			fileData, err := stylesheetFiles.ReadFile(srcFile)
			if err != nil {
				return err
			}

			buffer.Write(fileData)
		}

		minifiedData, err := minifier.Bytes("text/css", buffer.Bytes())
		if err != nil {
			return err
		}

		br, gz := precompress(minifiedData)
		StylesheetBundles[bundleName+".css"] = asset{
			Data:       minifiedData,
			Checksum:   crypto.HashFromBytes(minifiedData),
			BrotliData: br,
			GzipData:   gz,
		}
	}

	return nil
}

// GenerateJavascriptBundles creates JS bundles.
// basePath is prepended to route paths embedded in the service worker bundle.
func GenerateJavascriptBundles(webauthnEnabled bool, basePath string) error {
	var bundles = map[string][]string{
		"app": {
			"js/touch_handler.js",
			"js/keyboard_handler.js",
			"js/app.js",
		},
		"service-worker": {
			"js/service_worker.js",
		},
	}

	if webauthnEnabled {
		bundles["app"] = slices.Insert(bundles["app"], 1, "js/webauthn_handler.js")
	}

	JavascriptBundles = make(map[string]asset, len(bundles))

	jsMinifier := js.Minifier{Version: 2020}

	minifier := minify.New()
	minifier.AddFunc("text/javascript", jsMinifier.Minify)

	for bundleName, srcFiles := range bundles {
		var buffer bytes.Buffer

		for _, srcFile := range srcFiles {
			fileData, err := javascriptFiles.ReadFile(srcFile)
			if err != nil {
				return err
			}

			buffer.Write(fileData)
		}

		minifiedData, err := minifier.Bytes("text/javascript", buffer.Bytes())
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		buf.WriteString(licensePrefix)
		if bundleName == "service-worker" {
			fmt.Fprintf(&buf, "const OFFLINE_URL=%q;", basePath+"/offline")
		}
		buf.Write(minifiedData)
		buf.WriteString(licenseSuffix)

		contents := buf.Bytes()
		br, gz := precompress(contents)

		JavascriptBundles[bundleName+".js"] = asset{
			Data:       contents,
			Checksum:   crypto.HashFromBytes(contents),
			BrotliData: br,
			GzipData:   gz,
		}
	}

	return nil
}
