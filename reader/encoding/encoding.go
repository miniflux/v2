// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package encoding // import "miniflux.app/reader/encoding"

import (
	"bytes"
	"io"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

// CharsetReader is used when the XML encoding is specified for the input document.
//
// The document is converted in UTF-8 only if a different encoding is specified
// and the document is not already UTF-8.
//
// Several edge cases could exists:
//
// - Feeds with encoding specified only in Content-Type header and not in XML document
// - Feeds with encoding specified in both places
// - Feeds with encoding specified only in XML document and not in HTTP header
// - Feeds with wrong encoding defined and already in UTF-8
func CharsetReader(label string, input io.Reader) (io.Reader, error) {
	buffer, _ := io.ReadAll(input)
	r := bytes.NewReader(buffer)

	// The document is already UTF-8, do not do anything (avoid double-encoding).
	// That means the specified encoding in XML prolog is wrong.
	if utf8.Valid(buffer) {
		return r, nil
	}

	// Transform document to UTF-8 from the specified encoding in XML prolog.
	return charset.NewReaderLabel(label, r)
}
