package css // import "github.com/tdewolff/parse/css"

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestHashTable(t *testing.T) {
	test.T(t, ToHash([]byte("font")), Font, "'font' must resolve to hash.Font")
	test.T(t, Font.String(), "font")
	test.T(t, Margin_Left.String(), "margin-left")
	test.T(t, ToHash([]byte("")), Hash(0), "empty string must resolve to zero")
	test.T(t, Hash(0xffffff).String(), "")
	test.T(t, ToHash([]byte("fonts")), Hash(0), "'fonts' must resolve to zero")
}
