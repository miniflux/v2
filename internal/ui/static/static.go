// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package static // import "miniflux.app/v2/internal/ui/static"

import (
	"bytes"
	"embed"
	"fmt"

	"miniflux.app/v2/internal/crypto"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

// Static assets.
var (
	StylesheetBundleChecksums map[string]string
	StylesheetBundles         map[string][]byte
	JavascriptBundleChecksums map[string]string
	JavascriptBundles         map[string][]byte
)

//go:embed bin/*
var binaryFiles embed.FS

//go:embed css/*.css
var stylesheetFiles embed.FS

//go:embed js/*.js
var javascriptFiles embed.FS

var binaryFileChecksums map[string]string

// CalculateBinaryFileChecksums generates hash of embed binary files.
func CalculateBinaryFileChecksums() error {
	binaryFileChecksums = make(map[string]string)

	dirEntries, err := binaryFiles.ReadDir("bin")
	if err != nil {
		return err
	}

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
	return binaryFiles.ReadFile(fmt.Sprintf(`bin/%s`, filename))
}

// GetBinaryFileChecksum returns a binary file checksum.
func GetBinaryFileChecksum(filename string) (string, error) {
	if _, found := binaryFileChecksums[filename]; !found {
		return "", fmt.Errorf(`static: unable to find checksum for %q`, filename)
	}
	return binaryFileChecksums[filename], nil
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

	StylesheetBundles = make(map[string][]byte)
	StylesheetBundleChecksums = make(map[string]string)

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

		StylesheetBundles[bundle] = minifiedData
		StylesheetBundleChecksums[bundle] = crypto.HashFromBytes(minifiedData)
	}

	return nil
}

// GenerateJavascriptBundles creates JS bundles.
func GenerateJavascriptBundles() error {
	var bundles = map[string][]string{
		"app": {
			"js/dom_helper.js",
			"js/touch_handler.js",
			"js/keyboard_handler.js",
			"js/request_builder.js",
			"js/modal_handler.js",
			"js/app.js",
			"js/webauthn_handler.js",
			"js/bootstrap.js",
		},
		"service-worker": {
			"js/service_worker.js",
		},
	}

	var prefixes = map[string]string{
		"app": "(function(){'use strict';",
	}

	var suffixes = map[string]string{
		"app": "})();",
	}

	JavascriptBundles = make(map[string][]byte)
	JavascriptBundleChecksums = make(map[string]string)

	jsMinifier := js.Minifier{Version: 2017}

	minifier := minify.New()
	minifier.AddFunc("text/javascript", jsMinifier.Minify)

	for bundle, srcFiles := range bundles {
		var buffer bytes.Buffer

		if prefix, found := prefixes[bundle]; found {
			buffer.WriteString(prefix)
		}

		for _, srcFile := range srcFiles {
			fileData, err := javascriptFiles.ReadFile(srcFile)
			if err != nil {
				return err
			}

			buffer.Write(fileData)
		}

		if suffix, found := suffixes[bundle]; found {
			buffer.WriteString(suffix)
		}

		minifiedData, err := minifier.Bytes("text/javascript", buffer.Bytes())
		if err != nil {
			return err
		}

		JavascriptBundles[bundle] = minifiedData
		JavascriptBundleChecksums[bundle] = crypto.HashFromBytes(minifiedData)
	}

	return nil
}
