// Package json minifies JSON following the specifications at http://json.org/.
package json // import "github.com/tdewolff/minify/json"

import (
	"io"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/parse/json"
)

var (
	commaBytes = []byte(",")
	colonBytes = []byte(":")
)

////////////////////////////////////////////////////////////////

// DefaultMinifier is the default minifier.
var DefaultMinifier = &Minifier{}

// Minifier is a JSON minifier.
type Minifier struct{}

// Minify minifies JSON data, it reads from r and writes to w.
func Minify(m *minify.M, w io.Writer, r io.Reader, params map[string]string) error {
	return DefaultMinifier.Minify(m, w, r, params)
}

// Minify minifies JSON data, it reads from r and writes to w.
func (o *Minifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	skipComma := true

	p := json.NewParser(r)
	defer p.Restore()

	for {
		state := p.State()
		gt, text := p.Next()
		if gt == json.ErrorGrammar {
			if p.Err() != io.EOF {
				return p.Err()
			}
			return nil
		}

		if !skipComma && gt != json.EndObjectGrammar && gt != json.EndArrayGrammar {
			if state == json.ObjectKeyState || state == json.ArrayState {
				if _, err := w.Write(commaBytes); err != nil {
					return err
				}
			} else if state == json.ObjectValueState {
				if _, err := w.Write(colonBytes); err != nil {
					return err
				}
			}
		}
		skipComma = gt == json.StartObjectGrammar || gt == json.StartArrayGrammar

		if _, err := w.Write(text); err != nil {
			return err
		}
	}
}
