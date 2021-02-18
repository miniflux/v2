// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package static // import "miniflux.app/ui/static"

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"fmt"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
)

// Static assets.
var (
	StylesheetBundleChecksums map[string]string
	StylesheetBundles         map[string][]byte
)

//go:embed bin/*
var binaryFiles embed.FS

//go:embed css/*.css
var stylesheetFiles embed.FS

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

		binaryFileChecksums[dirEntry.Name()] = fmt.Sprintf("%x", sha256.Sum256(data))
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
		StylesheetBundleChecksums[bundle] = fmt.Sprintf("%x", sha256.Sum256(minifiedData))
	}

	return nil
}
