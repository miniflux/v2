package css // import "github.com/tdewolff/parse/css"

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/tdewolff/test"
)

type TTs []TokenType

func TestTokens(t *testing.T) {
	var tokenTests = []struct {
		css      string
		expected []TokenType
	}{
		{" ", TTs{}},
		{"5.2 .4", TTs{NumberToken, NumberToken}},
		{"color: red;", TTs{IdentToken, ColonToken, IdentToken, SemicolonToken}},
		{"background: url(\"http://x\");", TTs{IdentToken, ColonToken, URLToken, SemicolonToken}},
		{"background: URL(x.png);", TTs{IdentToken, ColonToken, URLToken, SemicolonToken}},
		{"color: rgb(4, 0%, 5em);", TTs{IdentToken, ColonToken, FunctionToken, NumberToken, CommaToken, PercentageToken, CommaToken, DimensionToken, RightParenthesisToken, SemicolonToken}},
		{"body { \"string\" }", TTs{IdentToken, LeftBraceToken, StringToken, RightBraceToken}},
		{"body { \"str\\\"ing\" }", TTs{IdentToken, LeftBraceToken, StringToken, RightBraceToken}},
		{".class { }", TTs{DelimToken, IdentToken, LeftBraceToken, RightBraceToken}},
		{"#class { }", TTs{HashToken, LeftBraceToken, RightBraceToken}},
		{"#class\\#withhash { }", TTs{HashToken, LeftBraceToken, RightBraceToken}},
		{"@media print { }", TTs{AtKeywordToken, IdentToken, LeftBraceToken, RightBraceToken}},
		{"/*comment*/", TTs{CommentToken}},
		{"/*com* /ment*/", TTs{CommentToken}},
		{"~= |= ^= $= *=", TTs{IncludeMatchToken, DashMatchToken, PrefixMatchToken, SuffixMatchToken, SubstringMatchToken}},
		{"||", TTs{ColumnToken}},
		{"<!-- -->", TTs{CDOToken, CDCToken}},
		{"U+1234", TTs{UnicodeRangeToken}},
		{"5.2 .4 4e-22", TTs{NumberToken, NumberToken, NumberToken}},
		{"--custom-variable", TTs{CustomPropertyNameToken}},

		// unexpected ending
		{"ident", TTs{IdentToken}},
		{"123.", TTs{NumberToken, DelimToken}},
		{"\"string", TTs{StringToken}},
		{"123/*comment", TTs{NumberToken, CommentToken}},
		{"U+1-", TTs{IdentToken, NumberToken, DelimToken}},

		// unicode
		{"fooδbar􀀀", TTs{IdentToken}},
		{"foo\\æ\\†", TTs{IdentToken}},
		// {"foo\x00bar", TTs{IdentToken}},
		{"'foo\u554abar'", TTs{StringToken}},
		{"\\000026B", TTs{IdentToken}},
		{"\\26 B", TTs{IdentToken}},

		// hacks
		{`\-\mo\z\-b\i\nd\in\g:\url(//business\i\nfo.co.uk\/labs\/xbl\/xbl\.xml\#xss);`, TTs{IdentToken, ColonToken, URLToken, SemicolonToken}},
		{"width/**/:/**/ 40em;", TTs{IdentToken, CommentToken, ColonToken, CommentToken, DimensionToken, SemicolonToken}},
		{":root *> #quince", TTs{ColonToken, IdentToken, DelimToken, DelimToken, HashToken}},
		{"html[xmlns*=\"\"]:root", TTs{IdentToken, LeftBracketToken, IdentToken, SubstringMatchToken, StringToken, RightBracketToken, ColonToken, IdentToken}},
		{"body:nth-of-type(1)", TTs{IdentToken, ColonToken, FunctionToken, NumberToken, RightParenthesisToken}},
		{"color/*\\**/: blue\\9;", TTs{IdentToken, CommentToken, ColonToken, IdentToken, SemicolonToken}},
		{"color: blue !ie;", TTs{IdentToken, ColonToken, IdentToken, DelimToken, IdentToken, SemicolonToken}},

		// escapes, null and replacement character
		{"c\\\x00olor: white;", TTs{IdentToken, ColonToken, IdentToken, SemicolonToken}},
		{"null\\0", TTs{IdentToken}},
		{"eof\\", TTs{IdentToken}},
		{"\"a\x00b\"", TTs{StringToken}},
		{"a\\\x00b", TTs{IdentToken}},
		{"url(a\x00b)", TTs{BadURLToken}}, // null character cannot be unquoted
		{"/*a\x00b*/", TTs{CommentToken}},

		// coverage
		{"  \n\r\n\r\"\\\r\n\\\r\"", TTs{StringToken}},
		{"U+?????? U+ABCD?? U+ABC-DEF", TTs{UnicodeRangeToken, UnicodeRangeToken, UnicodeRangeToken}},
		{"U+? U+A?", TTs{IdentToken, DelimToken, DelimToken, IdentToken, DelimToken, IdentToken, DelimToken}},
		{"-5.23 -moz", TTs{NumberToken, IdentToken}},
		{"()", TTs{LeftParenthesisToken, RightParenthesisToken}},
		{"url( //url  )", TTs{URLToken}},
		{"url( ", TTs{URLToken}},
		{"url( //url", TTs{URLToken}},
		{"url(\")a", TTs{URLToken}},
		{"url(a'\\\n)a", TTs{BadURLToken, IdentToken}},
		{"url(\"\n)a", TTs{BadURLToken, IdentToken}},
		{"url(a h)a", TTs{BadURLToken, IdentToken}},
		{"<!- | @4 ## /2", TTs{DelimToken, DelimToken, DelimToken, DelimToken, DelimToken, NumberToken, DelimToken, DelimToken, DelimToken, NumberToken}},
		{"\"s\\\n\"", TTs{StringToken}},
		{"\"a\\\"b\"", TTs{StringToken}},
		{"\"s\n", TTs{BadStringToken}},

		// small
		{"\"abcd", TTs{StringToken}},
		{"/*comment", TTs{CommentToken}},
		{"U+A-B", TTs{UnicodeRangeToken}},
		{"url((", TTs{BadURLToken}},
		{"id\u554a", TTs{IdentToken}},
	}
	for _, tt := range tokenTests {
		t.Run(tt.css, func(t *testing.T) {
			l := NewLexer(bytes.NewBufferString(tt.css))
			i := 0
			for {
				token, _ := l.Next()
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if token == WhitespaceToken {
					continue
				}
				test.That(t, i < len(tt.expected), "index", i, "must not exceed expected token types size", len(tt.expected))
				if i < len(tt.expected) {
					test.T(t, token, tt.expected[i], "token types must match")
				}
				i++
			}
		})
	}

	test.T(t, WhitespaceToken.String(), "Whitespace")
	test.T(t, EmptyToken.String(), "Empty")
	test.T(t, CustomPropertyValueToken.String(), "CustomPropertyValue")
	test.T(t, TokenType(100).String(), "Invalid(100)")
	test.T(t, NewLexer(bytes.NewBufferString("x")).consumeBracket(), ErrorToken, "consumeBracket on 'x' must return error")
}

////////////////////////////////////////////////////////////////

func ExampleNewLexer() {
	l := NewLexer(bytes.NewBufferString("color: red;"))
	out := ""
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			break
		} else if tt == WhitespaceToken || tt == CommentToken {
			continue
		}
		out += string(data)
	}
	fmt.Println(out)
	// Output: color:red;
}
