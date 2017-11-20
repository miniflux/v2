package css // import "github.com/tdewolff/minify/css"

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestCSS(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", "/*!bang comment*/"},
		{"i{}/*! bang  comment */", "i{}/*!bang comment*/"},
		{"i { key: value; key2: value; }", "i{key:value;key2:value}"},
		{".cla .ss > #id { x:y; }", ".cla .ss>#id{x:y}"},
		{".cla[id ^= L] { x:y; }", ".cla[id^=L]{x:y}"},
		{"area:focus { outline : 0;}", "area:focus{outline:0}"},
		{"@import 'file';", "@import 'file'"},
		{"@font-face { x:y; }", "@font-face{x:y}"},

		{"input[type=\"radio\"]{x:y}", "input[type=radio]{x:y}"},
		{"DIV{margin:1em}", "div{margin:1em}"},
		{".CLASS{margin:1em}", ".CLASS{margin:1em}"},
		{"@MEDIA all{}", "@media all{}"},
		{"@media only screen and (max-width : 800px){}", "@media only screen and (max-width:800px){}"},
		{"@media (-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}", "@media(-webkit-min-device-pixel-ratio:1.5),(min-resolution:1.5dppx){}"},
		{"[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}", "[class^=icon-] i[class^=icon-],i[class*=\" icon-\"]{x:y}"},
		{"html{line-height:1;}html{line-height:1;}", "html{line-height:1}html{line-height:1}"},
		{"a { b: 1", "a{b:1}"},

		{":root { --custom-variable:0px; }", ":root{--custom-variable:0px}"},

		// case sensitivity
		{"@counter-style Ident{}", "@counter-style Ident{}"},

		// coverage
		{"a, b + c { x:y; }", "a,b+c{x:y}"},

		// bad declaration
		{".clearfix { *zoom: 1px; }", ".clearfix{*zoom:1px}"},
		{".clearfix { *zoom: 1px }", ".clearfix{*zoom:1px}"},
		{".clearfix { color:green; *zoom: 1px; color:red; }", ".clearfix{color:green;*zoom:1px;color:red}"},

		// go-fuzz
		{"input[type=\"\x00\"] {  a: b\n}.a{}", "input[type=\"\x00\"]{a:b}.a{}"},
		{"a{a:)'''", "a{a:)'''}"},
	}

	m := minify.New()
	for _, tt := range cssTests {
		t.Run(tt.css, func(t *testing.T) {
			r := bytes.NewBufferString(tt.css)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.css, err, w.String(), tt.expected)
		})
	}
}

func TestCSSInline(t *testing.T) {
	cssTests := []struct {
		css      string
		expected string
	}{
		{"/*comment*/", ""},
		{"/*! bang  comment */", ""},
		{";", ""},
		{"empty:", "empty:"},
		{"key: value;", "key:value"},
		{"margin: 0 1; padding: 0 1;", "margin:0 1;padding:0 1"},
		{"color: #FF0000;", "color:red"},
		{"color: #000000;", "color:#000"},
		{"color: black;", "color:#000"},
		{"color: rgb(255,255,255);", "color:#fff"},
		{"color: rgb(100%,100%,100%);", "color:#fff"},
		{"color: rgba(255,0,0,1);", "color:red"},
		{"color: rgba(255,0,0,2);", "color:red"},
		{"color: rgba(255,0,0,0.5);", "color:rgba(255,0,0,.5)"},
		{"color: rgba(255,0,0,-1);", "color:transparent"},
		{"color: rgba(0%,15%,25%,0.2);", "color:rgba(0%,15%,25%,.2)"},
		{"color: rgba(0,0,0,0.5);", "color:rgba(0,0,0,.5)"},
		{"color: hsla(5,0%,10%,0.75);", "color:hsla(5,0%,10%,.75)"},
		{"color: hsl(0,100%,50%);", "color:red"},
		{"color: hsla(1,2%,3%,1);", "color:#080807"},
		{"color: hsla(1,2%,3%,0);", "color:transparent"},
		{"color: hsl(48,100%,50%);", "color:#fc0"},
		{"font-weight: bold; font-weight: normal;", "font-weight:700;font-weight:400"},
		{"font: bold \"Times new Roman\",\"Sans-Serif\";", "font:700 times new roman,\"sans-serif\""},
		{"outline: none;", "outline:0"},
		{"outline: none !important;", "outline:0!important"},
		{"border-left: none;", "border-left:0"},
		{"margin: 1 1 1 1;", "margin:1"},
		{"margin: 1 2 1 2;", "margin:1 2"},
		{"margin: 1 2 3 2;", "margin:1 2 3"},
		{"margin: 1 2 3 4;", "margin:1 2 3 4"},
		{"margin: 1 1 1 a;", "margin:1 1 1 a"},
		{"margin: 1 1 1 1 !important;", "margin:1!important"},
		{"padding:.2em .4em .2em", "padding:.2em .4em"},
		{"margin: 0em;", "margin:0"},
		{"font-family:'Arial', 'Times New Roman';", "font-family:arial,times new roman"},
		{"background:url('http://domain.com/image.png');", "background:url(http://domain.com/image.png)"},
		{"filter: progid : DXImageTransform.Microsoft.BasicImage(rotation=1);", "filter:progid:DXImageTransform.Microsoft.BasicImage(rotation=1)"},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity=0);", "filter:alpha(opacity=0)"},
		{"content: \"a\\\nb\";", "content:\"ab\""},
		{"content: \"a\\\r\nb\\\r\nc\";", "content:\"abc\""},
		{"content: \"\";", "content:\"\""},

		{"font:27px/13px arial,sans-serif", "font:27px/13px arial,sans-serif"},
		{"text-decoration: none !important", "text-decoration:none!important"},
		{"color:#fff", "color:#fff"},
		{"border:2px rgb(255,255,255);", "border:2px #fff"},
		{"margin:-1px", "margin:-1px"},
		{"margin:+1px", "margin:1px"},
		{"margin:0.5em", "margin:.5em"},
		{"margin:-0.5em", "margin:-.5em"},
		{"margin:05em", "margin:5em"},
		{"margin:.50em", "margin:.5em"},
		{"margin:5.0em", "margin:5em"},
		{"margin:5000em", "margin:5e3em"},
		{"color:#c0c0c0", "color:silver"},
		{"-ms-filter: \"progid:DXImageTransform.Microsoft.Alpha(Opacity=80)\";", "-ms-filter:\"alpha(opacity=80)\""},
		{"filter: progid:DXImageTransform.Microsoft.Alpha(Opacity = 80);", "filter:alpha(opacity=80)"},
		{"MARGIN:1EM", "margin:1em"},
		//{"color:CYAN", "color:cyan"}, // TODO
		{"width:attr(Name em)", "width:attr(Name em)"},
		{"content:CounterName", "content:CounterName"},
		{"background:URL(x.PNG);", "background:url(x.PNG)"},
		{"background:url(/*nocomment*/)", "background:url(/*nocomment*/)"},
		{"background:url(data:,text)", "background:url(data:,text)"},
		{"background:url('data:text/xml; version = 2.0,content')", "background:url(data:text/xml;version=2.0,content)"},
		{"background:url('data:\\'\",text')", "background:url('data:\\'\",text')"},
		{"margin:0 0 18px 0;", "margin:0 0 18px"},
		{"background:none", "background:0 0"},
		{"background:none 1 1", "background:none 1 1"},
		{"z-index:1000", "z-index:1000"},

		{"any:0deg 0s 0ms 0dpi 0dpcm 0dppx 0hz 0khz", "any:0 0s 0ms 0dpi 0dpcm 0dppx 0hz 0khz"},
		{"--custom-variable:0px;", "--custom-variable:0px"},
		{"--foo: if(x > 5) this.width = 10", "--foo: if(x > 5) this.width = 10"},
		{"--foo: ;", "--foo: "},

		// case sensitivity
		{"animation:Ident", "animation:Ident"},
		{"animation-name:Ident", "animation-name:Ident"},

		// coverage
		{"margin: 1 1;", "margin:1"},
		{"margin: 1 2;", "margin:1 2"},
		{"margin: 1 1 1;", "margin:1"},
		{"margin: 1 2 1;", "margin:1 2"},
		{"margin: 1 2 3;", "margin:1 2 3"},
		{"margin: 0%;", "margin:0"},
		{"color: rgb(255,64,64);", "color:#ff4040"},
		{"color: rgb(256,-34,2342435);", "color:#f0f"},
		{"color: rgb(120%,-45%,234234234%);", "color:#f0f"},
		{"color: rgb(0, 1, ident);", "color:rgb(0,1,ident)"},
		{"color: rgb(ident);", "color:rgb(ident)"},
		{"margin: rgb(ident);", "margin:rgb(ident)"},
		{"filter: progid:b().c.Alpha(rgba(x));", "filter:progid:b().c.Alpha(rgba(x))"},

		// go-fuzz
		{"FONT-FAMILY: ru\"", "font-family:ru\""},
	}

	m := minify.New()
	params := map[string]string{"inline": "1"}
	for _, tt := range cssTests {
		t.Run(tt.css, func(t *testing.T) {
			r := bytes.NewBufferString(tt.css)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, params)
			test.Minify(t, tt.css, err, w.String(), tt.expected)
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
		css string
		n   []int
	}{
		{`@import 'file'`, []int{0, 2}},
		{`@media all{}`, []int{0, 2, 3, 4}},
		{`a[id^="L"]{margin:2in!important;color:red}`, []int{0, 4, 6, 7, 8, 9, 10, 11}},
		{`a{color:rgb(255,0,0)}`, []int{4}},
		{`a{color:rgb(255,255,255)}`, []int{4}},
		{`a{color:hsl(0,100%,50%)}`, []int{4}},
		{`a{color:hsl(360,100%,100%)}`, []int{4}},
		{`a{color:f(arg)}`, []int{4}},
		{`<!--`, []int{0}},
		{`/*!comment*/`, []int{0, 1, 2}},
		{`a{--var:val}`, []int{2, 3, 4}},
		{`a{*color:0}`, []int{2, 3}},
		{`a{color:0;baddecl 5}`, []int{5}},
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.css, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.css)
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
	m.AddFunc("text/css", Minify)

	if err := m.Minify("text/css", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
