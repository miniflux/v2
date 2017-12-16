package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestXML(t *testing.T) {
	xmlTests := []struct {
		xml      string
		expected string
	}{
		{"<!-- comment -->", ""},
		{"<A>x</A>", "<A>x</A>"},
		{"<a><b>x</b></a>", "<a><b>x</b></a>"},
		{"<a><b>x\ny</b></a>", "<a><b>x\ny</b></a>"},
		{"<a> <![CDATA[ a ]]> </a>", "<a>a</a>"},
		{"<a >a</a >", "<a>a</a>"},
		{"<?xml  version=\"1.0\" ?>", "<?xml version=\"1.0\"?>"},
		{"<x></x>", "<x/>"},
		{"<x> </x>", "<x/>"},
		{"<x a=\"b\"></x>", "<x a=\"b\"/>"},
		{"<x a=\"\"></x>", "<x a=\"\"/>"},
		{"<x a=a></x>", "<x a=a/>"},
		{"<x a=\" a \n\r\t b \"/>", "<x a=\" a     b \"/>"},
		{"<x a=\"&apos;b&quot;\"></x>", "<x a=\"'b&#34;\"/>"},
		{"<x a=\"&quot;&quot;'\"></x>", "<x a='\"\"&#39;'/>"},
		{"<!DOCTYPE foo SYSTEM \"Foo.dtd\">", "<!DOCTYPE foo SYSTEM \"Foo.dtd\">"},
		{"text <!--comment--> text", "text text"},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"<!doctype html>", "<!doctype html=>"}, // bad formatted, doctype must be uppercase and html must have attribute value
		{"<x>\n<!--y-->\n</x>", "<x></x>"},
		{"<style>lala{color:red}</style>", "<style>lala{color:red}</style>"},
		{`cats  and 	dogs `, `cats and dogs`},

		// go fuzz
		{`</0`, `</0`},
		{`<!DOCTYPE`, `<!DOCTYPE`},
		{`<![CDATA[`, ``},
	}

	m := minify.New()
	for _, tt := range xmlTests {
		t.Run(tt.xml, func(t *testing.T) {
			r := bytes.NewBufferString(tt.xml)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.xml, err, w.String(), tt.expected)
		})
	}
}

func TestXMLKeepWhitespace(t *testing.T) {
	xmlTests := []struct {
		xml      string
		expected string
	}{
		{`cats  and 	dogs `, `cats and dogs`},
		{` <div> <i> test </i> <b> test </b> </div> `, `<div> <i> test </i> <b> test </b> </div>`},
		{"text\n<!--comment-->\ntext", "text\ntext"},
		{"text\n<!--comment-->text<!--comment--> text", "text\ntext text"},
		{"<x>\n<!--y-->\n</x>", "<x>\n</x>"},
		{"<style>lala{color:red}</style>", "<style>lala{color:red}</style>"},
		{"<x> <?xml?> </x>", "<x><?xml?> </x>"},
		{"<x> <![CDATA[ x ]]> </x>", "<x> x </x>"},
		{"<x> <![CDATA[ <<<<< ]]> </x>", "<x><![CDATA[ <<<<< ]]></x>"},
	}

	m := minify.New()
	xmlMinifier := &Minifier{KeepWhitespace: true}
	for _, tt := range xmlTests {
		t.Run(tt.xml, func(t *testing.T) {
			r := bytes.NewBufferString(tt.xml)
			w := &bytes.Buffer{}
			err := xmlMinifier.Minify(m, w, r, nil)
			test.Minify(t, tt.xml, err, w.String(), tt.expected)
		})
	}
}

func TestReaderErrors(t *testing.T) {
	r := test.NewErrorReader(0)
	w := &bytes.Buffer{}
	m := minify.New()
	err := Minify(m, w, r, nil)
	test.T(t, err, test.ErrPlain, "return error at first read")
}

func TestWriterErrors(t *testing.T) {
	errorTests := []struct {
		xml string
		n   []int
	}{
		{`<!DOCTYPE foo>`, []int{0}},
		{`<?xml?>`, []int{0, 1}},
		{`<a x=y z="val">`, []int{0, 1, 2, 3, 4, 8, 9}},
		{`<foo/>`, []int{1}},
		{`</foo>`, []int{0}},
		{`<foo></foo>`, []int{1}},
		{`<![CDATA[data<<<<<]]>`, []int{0}},
		{`text`, []int{0}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.xml, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.xml)
				w := test.NewErrorWriter(n)
				err := Minify(m, w, r, nil)
				test.T(t, err, test.ErrPlain)
			})
		}
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), Minify)

	if err := m.Minify("text/xml", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
