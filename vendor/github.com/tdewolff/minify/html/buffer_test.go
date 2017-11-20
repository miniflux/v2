package html // import "github.com/tdewolff/minify/html"

import (
	"bytes"
	"testing"

	"github.com/tdewolff/parse/html"
	"github.com/tdewolff/test"
)

func TestBuffer(t *testing.T) {
	//    0 12  3           45   6   7   8             9   0
	s := `<p><a href="//url">text</a>text<!--comment--></p>`
	z := NewTokenBuffer(html.NewLexer(bytes.NewBufferString(s)))

	tok := z.Shift()
	test.That(t, tok.Hash == html.P, "first token is <p>")
	test.That(t, z.pos == 0, "shift first token and restore position")
	test.That(t, len(z.buf) == 0, "shift first token and restore length")

	test.That(t, z.Peek(2).Hash == html.Href, "third token is href")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 3, "two tokens after peeking")

	test.That(t, z.Peek(8).Hash == html.P, "ninth token is <p>")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 9, "nine tokens after peeking")

	test.That(t, z.Peek(9).TokenType == html.ErrorToken, "tenth token is an error")
	test.That(t, z.Peek(9) == z.Peek(10), "tenth and eleventh tokens are EOF")
	test.That(t, len(z.buf) == 10, "ten tokens after peeking")

	_ = z.Shift()
	tok = z.Shift()
	test.That(t, tok.Hash == html.A, "third token is <a>")
	test.That(t, z.pos == 2, "don't change position after peeking")
}
