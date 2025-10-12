// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package static // import "miniflux.app/v2/internal/ui/static"

import (
	"bytes"
	"embed"
	"log/slog"
	"slices"
	"strings"

	"miniflux.app/v2/internal/crypto"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

type asset struct {
	Data     []byte
	Checksum string
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

		BinaryBundles[name] = asset{
			Data:     data,
			Checksum: crypto.HashFromBytes(data),
		}
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

	for bundle, srcFiles := range bundles {
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

		StylesheetBundles[bundle] = asset{
			Data:     minifiedData,
			Checksum: crypto.HashFromBytes(minifiedData),
		}
	}

	return nil
}

// GenerateJavascriptBundles creates JS bundles.
func GenerateJavascriptBundles(webauthnEnabled bool) error {
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

	for bundle, srcFiles := range bundles {
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

		JavascriptBundles[bundle] = asset{
			Data:     minifiedData,
			Checksum: crypto.HashFromBytes(minifiedData),
		}
	}

	return nil
}
