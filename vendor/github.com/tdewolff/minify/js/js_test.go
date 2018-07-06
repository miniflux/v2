package js // import "github.com/tdewolff/minify/js"

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/test"
)

func TestJS(t *testing.T) {
	jsTests := []struct {
		js       string
		expected string
	}{
		{"/*comment*/", ""},
		{"// comment\na", "a"},
		{"/*! bang  comment */", "/*!bang comment*/"},
		{"function x(){}", "function x(){}"},
		{"function x(a, b){}", "function x(a,b){}"},
		{"a  b", "a b"},
		{"a\n\nb", "a\nb"},
		{"a// comment\nb", "a\nb"},
		{"''\na", "''\na"},
		{"''\n''", "''\n''"},
		{"]\n0", "]\n0"},
		{"a\n{", "a\n{"},
		{";\na", ";a"},
		{",\na", ",a"},
		{"}\na", "}\na"},
		{"+\na", "+\na"},
		{"+\n(", "+\n("},
		{"+\n\"\"", "+\n\"\""},
		{"a + ++b", "a+ ++b"},                                          // JSMin caution
		{"var a=/\\s?auto?\\s?/i\nvar", "var a=/\\s?auto?\\s?/i\nvar"}, // #14
		{"var a=0\n!function(){}", "var a=0\n!function(){}"},           // #107
		{"function(){}\n\"string\"", "function(){}\n\"string\""},       // #109
		{"false\n\"string\"", "false\n\"string\""},                     // #109
		{"`\n", "`"},                                                   // go fuzz
		{"a\n~b", "a\n~b"},                                             // #132
		{"x / /\\d+/.exec(s)[0]", "x/ /\\d+/.exec(s)[0]"},              // #183

		{"function(){}\n`string`", "function(){}\n`string`"}, // #181
		{"false\n`string`", "false\n`string`"},               // #181
		{"`string`\nwhatever()", "`string`\nwhatever()"},     // #181

		{"x+/**/++y", "x+ ++y"},                          // #185
		{"x+\n++y", "x+\n++y"},                           // #185
		{"f()/*!com\nment*/g()", "f()/*!com\nment*/g()"}, // #185
		{"f()/*com\nment*/g()", "f()\ng()"},              // #185
		{"f()/*!\n*/g()", "f()/*!\n*/g()"},               // #185

		// go-fuzz
		{`/\`, `/\`},
	}

	m := minify.New()
	for _, tt := range jsTests {
		t.Run(tt.js, func(t *testing.T) {
			r := bytes.NewBufferString(tt.js)
			w := &bytes.Buffer{}
			err := Minify(m, w, r, nil)
			test.Minify(t, tt.js, err, w.String(), tt.expected)
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
		js string
		n  []int
	}{
		{"a\n{5 5", []int{0, 1, 4}},
		{`/*!comment*/`, []int{0, 1, 2}},
		{"false\n\"string\"", []int{1}}, // #109
	}

	m := minify.New()
	for _, tt := range errorTests {
		for _, n := range tt.n {
			t.Run(fmt.Sprint(tt.js, " ", tt.n), func(t *testing.T) {
				r := bytes.NewBufferString(tt.js)
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
	m.AddFunc("text/javascript", Minify)

	if err := m.Minify("text/javascript", os.Stdout, os.Stdin); err != nil {
		panic(err)
	}
}
