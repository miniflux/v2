// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/text/internal"
	"golang.org/x/text/message/pipeline"
)

// TODO:
// - merge information into existing files
// - handle different file formats (PO, XLIFF)
// - handle features (gender, plural)
// - message rewriting

var (
	lang *string
)

func init() {
	lang = cmdExtract.Flag.String("lang", "en-US", "comma-separated list of languages to process")
}

var cmdExtract = &Command{
	Run:       runExtract,
	UsageLine: "extract <package>*",
	Short:     "extracts strings to be translated from code",
}

func runExtract(cmd *Command, config *pipeline.Config, args []string) error {
	config.Packages = args
	state, err := pipeline.Extract(config)
	if err != nil {
		return wrap(err, "extract failed")
	}
	out := state.Extracted

	langs := append(getLangs(), config.SourceLanguage)
	langs = internal.UniqueTags(langs)
	for _, tag := range langs {
		// TODO: inject translations from existing files to avoid retranslation.
		out.Language = tag
		data, err := json.MarshalIndent(out, "", "    ")
		if err != nil {
			return wrap(err, "JSON marshal failed")
		}
		file := filepath.Join(*dir, tag.String(), outFile)
		if err := os.MkdirAll(filepath.Dir(file), 0750); err != nil {
			return wrap(err, "dir create failed")
		}
		if err := ioutil.WriteFile(file, data, 0740); err != nil {
			return wrap(err, "write failed")
		}
	}
	return nil
}
