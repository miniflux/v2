// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package static // import "miniflux.app/ui/static"

import (
	"crypto/sha256"
	"embed"
	"fmt"
)

//go:embed bin/*
var binaryFiles embed.FS

var binaryFileChecksums map[string]string

func init() {
	binaryFileChecksums = make(map[string]string)

	dirEntries, err := binaryFiles.ReadDir("bin")
	if err != nil {
		panic(err)
	}

	for _, dirEntry := range dirEntries {
		data, err := LoadBinaryFile(dirEntry.Name())
		if err != nil {
			panic(err)
		}

		binaryFileChecksums[dirEntry.Name()] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
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
