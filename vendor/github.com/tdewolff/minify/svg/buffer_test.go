package svg // import "github.com/tdewolff/minify/svg"

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/tdewolff/parse/svg"
	"github.com/tdewolff/parse/xml"
	"github.com/tdewolff/test"
)

func TestBuffer(t *testing.T) {
	//    0   12     3            4 5   6   7 8   9    01
	s := `<svg><path d="M0 0L1 1z"/>text<tag/>text</svg>`
	z := NewTokenBuffer(xml.NewLexer(bytes.NewBufferString(s)))

	tok := z.Shift()
	test.That(t, tok.Hash == svg.Svg, "first token is <svg>")
	test.That(t, z.pos == 0, "shift first token and restore position")
	test.That(t, len(z.buf) == 0, "shift first token and restore length")

	test.That(t, z.Peek(2).Hash == svg.D, "third token is d")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 3, "mtwo tokens after peeking")

	test.That(t, z.Peek(8).Hash == svg.Svg, "ninth token is <svg>")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 9, "nine tokens after peeking")

	test.That(t, z.Peek(9).TokenType == xml.ErrorToken, "tenth token is an error")
	test.That(t, z.Peek(9) == z.Peek(10), "tenth and eleventh token are EOF")
	test.That(t, len(z.buf) == 10, "ten tokens after peeking")

	_ = z.Shift()
	tok = z.Shift()
	test.That(t, tok.Hash == svg.Path, "third token is <path>")
	test.That(t, z.pos == 2, "don't change position after peeking")
}

func TestAttributes(t *testing.T) {
	r := bytes.NewBufferString(`<rect x="0" y="1" width="2" height="3" rx="4" ry="5"/>`)
	l := xml.NewLexer(r)
	tb := NewTokenBuffer(l)
	tb.Shift()
	for k := 0; k < 2; k++ { // run twice to ensure similar results
		attrs, _ := tb.Attributes(svg.X, svg.Y, svg.Width, svg.Height, svg.Rx, svg.Ry)
		for i := 0; i < 6; i++ {
			test.That(t, attrs[i] != nil, "attr must not be nil")
			val := string(attrs[i].AttrVal)
			j, _ := strconv.ParseInt(val, 10, 32)
			test.That(t, int(j) == i, "attr data is bad at position", i)
		}
	}
}

////////////////////////////////////////////////////////////////

func BenchmarkAttributes(b *testing.B) {
	r := bytes.NewBufferString(`<rect x="0" y="1" width="2" height="3" rx="4" ry="5"/>`)
	l := xml.NewLexer(r)
	tb := NewTokenBuffer(l)
	tb.Shift()
	tb.Peek(6)
	for i := 0; i < b.N; i++ {
		tb.Attributes(svg.X, svg.Y, svg.Width, svg.Height, svg.Rx, svg.Ry)
	}
}
