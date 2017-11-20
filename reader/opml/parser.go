// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package opml

import (
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

func Parse(data io.Reader) (SubcriptionList, error) {
	opml := new(Opml)
	decoder := xml.NewDecoder(data)
	decoder.CharsetReader = charset.NewReaderLabel

	err := decoder.Decode(opml)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse OPML file: %v\n", err)
	}

	return opml.Transform(), nil
}
