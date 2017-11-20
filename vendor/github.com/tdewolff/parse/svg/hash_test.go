package svg // import "github.com/tdewolff/parse/svg"

import (
	"testing"

	"github.com/tdewolff/test"
)

func TestHashTable(t *testing.T) {
	test.T(t, ToHash([]byte("svg")), Svg, "'svg' must resolve to hash.Svg")
	test.T(t, ToHash([]byte("width")), Width, "'width' must resolve to hash.Width")
	test.T(t, Svg.String(), "svg")
	test.T(t, ToHash([]byte("")), Hash(0), "empty string must resolve to zero")
	test.T(t, Hash(0xffffff).String(), "")
	test.T(t, ToHash([]byte("svgs")), Hash(0), "'svgs' must resolve to zero")
	test.T(t, ToHash([]byte("uopi")), Hash(0), "'uopi' must resolve to zero")
}
