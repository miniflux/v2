// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package xml // import "miniflux.app/reader/xml"

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	type myxml struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Title   string   `xml:"title"`
	}
	// Add the body contains illegal characters
	data := fmt.Sprintf(`<?xml version="1.0" encoding="windows-1251"?><rss version="2.0"><title>%s</title></rss>`, "\x10")
	var x myxml

	decoder := NewDecoder(strings.NewReader(data))
	err := decoder.Decode(&x)
	if err != nil {
		t.Error(err)
	}
}
