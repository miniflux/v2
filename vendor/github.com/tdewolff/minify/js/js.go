// Package js minifies ECMAScript5.1 following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js // import "github.com/tdewolff/minify/js"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse"
	"github.com/tdewolff/parse/js"
)

var (
	spaceBytes   = []byte(" ")
	newlineBytes = []byte("\n")
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JS minifier.
type Minifier struct{}

// Minify minifies JS data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies JS data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	prev := js.LineTerminatorToken
	prevLast := byte(' ')
	lineTerminatorQueued := false
	whitespaceQueued := false

	l := js.NewLexer(r)
	defer l.Restore()

	for {
		tt, data := l.Next()
		if tt == js.ErrorToken {
			if l.Err() != io.EOF {
				return l.Err()
			}
			return nil
		} else if tt == js.LineTerminatorToken {
			lineTerminatorQueued = true
		} else if tt == js.WhitespaceToken {
			whitespaceQueued = true
		} else if tt == js.CommentToken {
			if len(data) > 5 && data[1] == '*' && data[2] == '!' {
				if _, err := w.Write(data[:3]); err != nil {
					return err
				}
				comment := parse.TrimWhitespace(parse.ReplaceMultipleWhitespace(data[3 : len(data)-2]))
				if _, err := w.Write(comment); err != nil {
					return err
				}
				if _, err := w.Write(data[len(data)-2:]); err != nil {
					return err
				}
			}
		} else {
			first := data[0]
			if (prev == js.IdentifierToken || prev == js.NumericToken || prev == js.PunctuatorToken || prev == js.StringToken || prev == js.RegexpToken) &&
				(tt == js.IdentifierToken || tt == js.NumericToken || tt == js.StringToken || tt == js.PunctuatorToken || tt == js.RegexpToken) {
				if lineTerminatorQueued && (prev != js.PunctuatorToken || prevLast == '}' || prevLast == ']' || prevLast == ')' || prevLast == '+' || prevLast == '-' || prevLast == '"' || prevLast == '\'') &&
					(tt != js.PunctuatorToken || first == '{' || first == '[' || first == '(' || first == '+' || first == '-' || first == '!' || first == '~') {
					if _, err := w.Write(newlineBytes); err != nil {
						return err
					}
				} else if whitespaceQueued && (prev != js.StringToken && prev != js.PunctuatorToken && tt != js.PunctuatorToken || (prevLast == '+' || prevLast == '-') && first == prevLast) {
					if _, err := w.Write(spaceBytes); err != nil {
						return err
					}
				}
			}
			if _, err := w.Write(data); err != nil {
				return err
			}
			prev = tt
			prevLast = data[len(data)-1]
			lineTerminatorQueued = false
			whitespaceQueued = false
		}
	}
}
