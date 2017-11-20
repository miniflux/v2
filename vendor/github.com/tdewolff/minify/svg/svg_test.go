package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/test"
)

func TestSVG(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<!-- comment -->`, ``},
		{`<!DOCTYPE svg SYSTEM "foo.dtd">`, ``},
		{`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`, `<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`},
		{`<!DOCTYPE svg SYSTEM "foo.dtd">`, ``},
		{`<?xml version="1.0" ?>`, ``},
		{`<style> <![CDATA[ x ]]> </style>`, `<style>x</style>`},
		{`<style> <![CDATA[ <<<< ]]> </style>`, `<style>&lt;&lt;&lt;&lt;</style>`},
		{`<style> <![CDATA[ <<<<< ]]> </style>`, `<style><![CDATA[ <<<<< ]]></style>`},
		{`<style/><![CDATA[ <<<<< ]]>`, `<style/><![CDATA[ <<<<< ]]>`},
		{`<svg version="1.0"></svg>`, `<svg version="1.0"/>`},
		{`<svg version="1.1" x="0" y="0px" width="100%" height="100%"><path/></svg>`, `<svg><path/></svg>`},
		{`<path x="a"> </path>`, `<path x="a"/>`},
		{`<path x=" a "/>`, `<path x="a"/>`},
		{"<path x=\" a \n b \"/>", `<path x="a b"/>`},
		{`<path x="5.0px" y="0%"/>`, `<path x="5" y="0"/>`},
		{`<svg viewBox="5.0px 5px 240IN px"><path/></svg>`, `<svg viewBox="5 5 240in px"><path/></svg>`},
		{`<svg viewBox="5.0!5px"><path/></svg>`, `<svg viewBox="5!5px"><path/></svg>`},
		{`<path d="M 100 100 L 300 100 L 200 100 z"/>`, `<path d="M1e2 1e2H3e2 2e2z"/>`},
		{`<path d="M100 -100M200 300z"/>`, `<path d="M1e2-1e2M2e2 3e2z"/>`},
		{`<path d="M0.5 0.6 M -100 0.5z"/>`, `<path d="M.5.6M-1e2.5z"/>`},
		{`<path d="M01.0 0.6 z"/>`, `<path d="M1 .6z"/>`},
		{`<path d="M20 20l-10-10z"/>`, `<path d="M20 20 10 10z"/>`},
		{`<?xml version="1.0" encoding="utf-8"?>`, ``},
		{`<svg viewbox="0 0 16 16"><path/></svg>`, `<svg viewbox="0 0 16 16"><path/></svg>`},
		{`<g></g>`, ``},
		{`<g><path/></g>`, `<path/>`},
		{`<g id="a"><g><path/></g></g>`, `<g id="a"><path/></g>`},
		{`<path fill="#ffffff"/>`, `<path fill="#fff"/>`},
		{`<path fill="#fff"/>`, `<path fill="#fff"/>`},
		{`<path fill="white"/>`, `<path fill="#fff"/>`},
		{`<path fill="#ff0000"/>`, `<path fill="red"/>`},
		{`<line x1="5" y1="10" x2="20" y2="40"/>`, `<path d="M5 10 20 40z"/>`},
		{`<rect x="5" y="10" width="20" height="40"/>`, `<path d="M5 10h20v40H5z"/>`},
		{`<rect x="-5.669" y="147.402" fill="#843733" width="252.279" height="14.177"/>`, `<path fill="#843733" d="M-5.669 147.402h252.279v14.177H-5.669z"/>`},
		{`<rect x="5" y="10" rx="2" ry="3"/>`, `<rect x="5" y="10" rx="2" ry="3"/>`},
		{`<rect x="5" y="10" height="40"/>`, ``},
		{`<rect x="5" y="10" width="30" height="0"/>`, ``},
		{`<polygon points="1,2 3,4"/>`, `<path d="M1 2 3 4z"/>`},
		{`<polyline points="1,2 3,4"/>`, `<path d="M1 2 3 4"/>`},
		{`<svg contentStyleType="text/json ; charset=iso-8859-1"><style>{a : true}</style></svg>`, `<svg contentStyleType="text/json;charset=iso-8859-1"><style>{a : true}</style></svg>`},
		{`<metadata><dc:title /></metadata>`, ``},

		// from SVGO
		{`<!DOCTYPE bla><?xml?><!-- comment --><metadata/>`, ``},

		{`<polygon fill="none" stroke="#000" points="-0.1,"/>`, `<polygon fill="none" stroke="#000" points="-0.1,"/>`}, // #45
		{`<path stroke="url(#UPPERCASE)"/>`, `<path stroke="url(#UPPERCASE)"/>`},                                       // #117

		// go fuzz
		{`<0 d=09e9.6e-9e0`, `<0 d=""`},
		{`<line`, `<line`},
	}

	m := minify.New()
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
		})
	}
}

func TestSVGStyle(t *testing.T) {
	svgTests := []struct {
		svg      string
		expected string
	}{
		{`<style> a > b {} </style>`, `<style>a>b{}</style>`},
		{`<style> <![CDATA[ @media x < y {} ]]> </style>`, `<style>@media x &lt; y{}</style>`},
		{`<style> <![CDATA[ * { content: '<<<<<'; } ]]> </style>`, `<style><![CDATA[*{content:'<<<<<'}]]></style>`},
		{`<style/><![CDATA[ * { content: '<<<<<'; ]]>`, `<style/><![CDATA[ * { content: '<<<<<'; ]]>`},
		{`<path style="fill: black; stroke: #ff0000;"/>`, `<path style="fill:#000;stroke:red"/>`},
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
		})
	}
}

func TestSVGDecimals(t *testing.T) {
	var svgTests = []struct {
		svg      string
		expected string
	}{
		{`<svg x="1.234" y="0.001" width="1.001"><path/></svg>`, `<svg x="1.2" width="1"><path/></svg>`},
	}

	m := minify.New()
	o := &Minifier{Decimals: 1}
	for _, tt := range svgTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := o.Minify(m, w, r, nil)
			test.Minify(t, tt.svg, err, w.String(), tt.expected)
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
		svg string
		n   []int
	}{
		{`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "foo.dtd" [ <!ENTITY x "bar"> ]>`, []int{0}},
		{`abc`, []int{0}},
		{`<style>abc</style>`, []int{2}},
		{`<![CDATA[ <<<< ]]>`, []int{0}},
		{`<![CDATA[ <<<<< ]]>`, []int{0}},
		{`<path d="x"/>`, []int{0, 1, 2, 3, 4, 5}},
		{`<path></path>`, []int{1}},
		{`<svg>x</svg>`, []int{1, 3}},
		{`<svg>x</svg >`, []int{3}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.svg, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.svg)
				w := test.NewErrorWriter(n)
				err := Minify(m, w, r, nil)
				test.T(t, err, test.ErrPlain)
			})
		}
	}
}

func TestMinifyErrors(t *testing.T) {
	errorTests := []struct {
		svg string
		err error
	}{
		{`<style>abc</style>`, test.ErrPlain},
		{`<style><![CDATA[abc]]></style>`, test.ErrPlain},
		{`<path style="abc"/>`, test.ErrPlain},
	}

	m := minify.New()
	m.AddFunc("text/css", func(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
		return test.ErrPlain
	})
	for _, tt := range errorTests {
		t.Run(tt.svg, func(t *testing.T) {
			r := bytes.NewBufferString(tt.svg)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.T(t, err, tt.err)
		})
	}
}

////////////////////////////////////////////////////////////////

func ExampleMinify() {
	m := minify.New()
	m.AddFunc("image/svg+xml", Minify)
	m.AddFunc("text/css", css.Minify)

	if err := m.Minify("image/svg+xml", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
