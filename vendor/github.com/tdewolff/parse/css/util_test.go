package css // import "github.com/tdewolff/parse/css"

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestIsIdent(t *testing.T) {
	test.That(t, IsIdent([]byte("color")))
	test.That(t, !IsIdent([]byte("4.5")))
}

func TestIsURLUnquoted(t *testing.T) {
	test.That(t, IsURLUnquoted([]byte("http://x")))
	test.That(t, !IsURLUnquoted([]byte(")")))
}

func TestHsl2Rgb(t *testing.T) {
	r, g, b := HSL2RGB(0.0, 1.0, 0.5)
	test.T(t, r, 1.0)
	test.T(t, g, 0.0)
	test.T(t, b, 0.0)

	r, g, b = HSL2RGB(1.0, 1.0, 0.5)
	test.T(t, r, 1.0)
	test.T(t, g, 0.0)
	test.T(t, b, 0.0)

	r, g, b = HSL2RGB(0.66, 0.0, 1.0)
	test.T(t, r, 1.0)
	test.T(t, g, 1.0)
	test.T(t, b, 1.0)
}
