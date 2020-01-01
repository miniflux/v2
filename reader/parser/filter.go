package parser // import "miniflux.app/reader/parser"

import (
	"strings"
	"unicode"
)

func StripInvalidCharacter(data string) string {
	return strings.Map(func(s rune) rune {
		if unicode.IsPrint(s) {
			return s
		}
		return -1
	}, data)
}
