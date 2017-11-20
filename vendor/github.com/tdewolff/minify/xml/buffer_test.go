package xml // import "github.com/tdewolff/minify/xml"

import (
	"bytes"
	"testing"

	"github.com/tdewolff/parse/xml"
	"github.com/tdewolff/test"
)

func TestBuffer(t *testing.T) {
	//    0 12  3           45   6   7   8             9   0
	s := `<p><a href="//url">text</a>text<!--comment--></p>`
	z := NewTokenBuffer(xml.NewLexer(bytes.NewBufferString(s)))

	tok := z.Shift()
	test.That(t, string(tok.Text) == "p", "first token is <p>")
	test.That(t, z.pos == 0, "shift first token and restore position")
	test.That(t, len(z.buf) == 0, "shift first token and restore length")

	test.That(t, string(z.Peek(2).Text) == "href", "third token is href")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 3, "two tokens after peeking")

	test.That(t, string(z.Peek(8).Text) == "p", "ninth token is <p>")
	test.That(t, z.pos == 0, "don't change position after peeking")
	test.That(t, len(z.buf) == 9, "nine tokens after peeking")

	test.That(t, z.Peek(9).TokenType == xml.ErrorToken, "tenth token is an error")
	test.That(t, z.Peek(9) == z.Peek(10), "tenth and eleventh token are EOF")
	test.That(t, len(z.buf) == 10, "ten tokens after peeking")

	_ = z.Shift()
	tok = z.Shift()
	test.That(t, string(tok.Text) == "a", "third token is <a>")
	test.That(t, z.pos == 2, "don't change position after peeking")
}
