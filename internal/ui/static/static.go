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
	"miniflux.app/v2/internal/http/response"

	"github.com/andybalholm/brotli"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

// LibreJS license markers wrapped around the JavaScript bundles.
const licensePrefix = "//@license magnet:?xt=urn:btih:8e4f440f4c65981c5bf93c76d35135ba5064d8b7&dn=apache-2.0.txt Apache-2.0\n"
const licenseSuffix = "\n//@license-end"

type asset struct {
	Data     []byte
	Checksum string
	// Precomputed Brotli and Gzip variants of Data, used to avoid
	// recompressing immutable assets on every request. Nil when the asset is
	// too small to be worth compressing.
	Brotli []byte
	Gzip   []byte
}

// precompressAsset computes the Brotli and Gzip variants of an immutable asset
// at maximum compression level. Both are nil for assets below the compression
// threshold.
func precompressAsset(data []byte) (brotliData, gzipData []byte) {
	if len(data) <= response.CompressionThreshold {
		return nil, nil
	}

	var brotliBuffer bytes.Buffer
	brotliWriter := brotli.NewWriterV2(&brotliBuffer, brotli.BestCompression)
	brotliWriter.Write(data)
	brotliWriter.Close()

	var gzipBuffer bytes.Buffer
	gzipWriter, _ := gzip.NewWriterLevel(&gzipBuffer, gzip.BestCompression)
	gzipWriter.Write(data)
	gzipWriter.Close()

	return brotliBuffer.Bytes(), gzipBuffer.Bytes()
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

		bundle := asset{
			Data:     data,
			Checksum: crypto.HashFromBytes(data),
		}

		// Only text-based SVG icons benefit from compression; raster images
		// (PNG, ICO) are already compressed and served without it.
		if strings.HasSuffix(name, ".svg") {
			bundle.Brotli, bundle.Gzip = precompressAsset(data)
		}

		BinaryBundles[name] = bundle
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

		brotliData, gzipData := precompressAsset(minifiedData)
		StylesheetBundles[bundleName+".css"] = asset{
			Data:     minifiedData,
			Checksum: crypto.HashFromBytes(minifiedData),
			Brotli:   brotliData,
			Gzip:     gzipData,
		}
	}

	return nil
}

// GenerateJavascriptBundles creates JS bundles.
func GenerateJavascriptBundles(webauthnEnabled bool, offlineURL string) error {
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

	// The service worker needs the offline URL, which is known at startup, so
	// it is prepended once here instead of on every request.
	serviceWorkerPrefix := fmt.Sprintf("const OFFLINE_URL=%q;", offlineURL)

	JavascriptBundles = make(map[string]asset, len(bundles))

	jsMinifier := js.Minifier{Version: 2020}

	minifier := minify.New()
	minifier.AddFunc("text/javascript", jsMinifier.Minify)

	for bundleName, srcFiles := range bundles {
		var buffer bytes.Buffer

		if bundleName == "service-worker" {
			buffer.WriteString(serviceWorkerPrefix)
		}

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

		// Wrap the minified bundle with the LibreJS license markers once, at
		// generation time, so the served bytes are stable and can be
		// precompressed.
		wrappedData := make([]byte, 0, len(licensePrefix)+len(minifiedData)+len(licenseSuffix))
		wrappedData = append(wrappedData, licensePrefix...)
		wrappedData = append(wrappedData, minifiedData...)
		wrappedData = append(wrappedData, licenseSuffix...)

		brotliData, gzipData := precompressAsset(wrappedData)
		JavascriptBundles[bundleName+".js"] = asset{
			Data:     wrappedData,
			Checksum: crypto.HashFromBytes(wrappedData),
			Brotli:   brotliData,
			Gzip:     gzipData,
		}
	}

	return nil
}
