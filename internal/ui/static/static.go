// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package static // import "influxeed-engine/v2/internal/ui/static"

import (
	"bytes"
	"embed"
	"fmt"

	"influxeed-engine/v2/internal/crypto"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type asset struct {
	Data     []byte
	Checksum string
}

// Static assets.
var (
	StylesheetBundles   map[string]asset
	JavascriptBundles   map[string]asset
	binaryFileChecksums map[string]string
)

//go:embed bin/*
var binaryFiles embed.FS

//go:embed css/*.css
var stylesheetFiles embed.FS

//go:embed js/*.js
var javascriptFiles embed.FS

// CalculateBinaryFileChecksums generates hash of embed binary files.
func CalculateBinaryFileChecksums() error {
	dirEntries, err := binaryFiles.ReadDir("bin")
	if err != nil {
		return err
	}
	binaryFileChecksums = make(map[string]string, len(dirEntries))

	for _, dirEntry := range dirEntries {
		data, err := LoadBinaryFile(dirEntry.Name())
		if err != nil {
			return err
		}

		binaryFileChecksums[dirEntry.Name()] = crypto.HashFromBytes(data)
	}

	return nil
}

// LoadBinaryFile loads an embed binary file.
func LoadBinaryFile(filename string) ([]byte, error) {
	return binaryFiles.ReadFile("bin/" + filename)
}

// GetBinaryFileChecksum returns a binary file checksum.
func GetBinaryFileChecksum(filename string) (string, error) {
	data, found := binaryFileChecksums[filename]
	if !found {
		return "", fmt.Errorf(`static: unable to find checksum for %q`, filename)
	}
	return data, nil
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
func GenerateJavascriptBundles() error {
	var bundles = map[string][]string{
		"app": {
			"js/tt.js", // has to be first
			"js/touch_handler.js",
			"js/keyboard_handler.js",
			"js/modal_handler.js",
			"js/webauthn_handler.js",
			"js/app.js",
		},
		"service-worker": {
			"js/service_worker.js",
		},
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
