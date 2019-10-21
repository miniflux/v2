package xml // import "miniflux.app/reader/xml"

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"miniflux.app/reader/encoding"
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
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.Entity = xml.HTMLEntity
	decoder.Strict = false
	decoder.CharsetReader = encoding.CharsetReader
	err := decoder.Decode(&x)
	if _, ok := err.(*xml.SyntaxError); !ok {
		t.Errorf("Unexpected error: %v, expected xml.SyntaxError", err)
	}

	decoder = GetDecoder(strings.NewReader(data))
	err = decoder.Decode(&x)
	if err != nil {
		t.Error(err)
	}
}
