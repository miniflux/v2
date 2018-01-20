// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package encoding

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
// - Feeds with charset specified only in Content-Type header and not in XML document
// - Feeds with charset specified in both places
// - Feeds with charset specified only in XML document and not in HTTP header
func CharsetReader(label string, input io.Reader) (io.Reader, error) {
	var buf1, buf2 bytes.Buffer
	w := io.MultiWriter(&buf1, &buf2)
	io.Copy(w, input)
	r := bytes.NewReader(buf2.Bytes())

	if !utf8.Valid(buf1.Bytes()) {
		// Transform document to UTF-8 from the specified XML encoding.
		return charset.NewReaderLabel(label, r)
	}

	// The document is already UTF-8, do not do anything (avoid double-encoding)
	return r, nil
}
