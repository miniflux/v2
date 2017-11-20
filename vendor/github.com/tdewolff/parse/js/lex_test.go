package js // import "github.com/tdewolff/parse/js"

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
		js       string
		expected []TokenType
	}{
		{" \t\v\f\u00A0\uFEFF\u2000", TTs{}}, // WhitespaceToken
		{"\n\r\r\n\u2028\u2029", TTs{LineTerminatorToken}},
		{"5.2 .04 0x0F 5e99", TTs{NumericToken, NumericToken, NumericToken, NumericToken}},
		{"a = 'string'", TTs{IdentifierToken, PunctuatorToken, StringToken}},
		{"/*comment*/ //comment", TTs{CommentToken, CommentToken}},
		{"{ } ( ) [ ]", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{". ; , < > <=", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{">= == != === !==", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"+ - * % ++ --", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"<< >> >>> & | ^", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"! ~ && || ? :", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"= += -= *= %= <<=", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{">>= >>>= &= |= ^= =>", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"a = /.*/g;", TTs{IdentifierToken, PunctuatorToken, RegexpToken, PunctuatorToken}},

		{"/*co\nm\u2028m/*ent*/ //co//mment\u2029//comment", TTs{CommentToken, CommentToken, LineTerminatorToken, CommentToken}},
		{"<!-", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"1<!--2\n", TTs{NumericToken, CommentToken, LineTerminatorToken}},
		{"x=y-->10\n", TTs{IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, PunctuatorToken, NumericToken, LineTerminatorToken}},
		{"  /*comment*/ -->nothing\n", TTs{CommentToken, CommentToken, LineTerminatorToken}},
		{"1 /*comment\nmultiline*/ -->nothing\n", TTs{NumericToken, CommentToken, CommentToken, LineTerminatorToken}},
		{"$ _\u200C \\u2000 \u200C", TTs{IdentifierToken, IdentifierToken, IdentifierToken, UnknownToken}},
		{">>>=>>>>=", TTs{PunctuatorToken, PunctuatorToken, PunctuatorToken}},
		{"1/", TTs{NumericToken, PunctuatorToken}},
		{"1/=", TTs{NumericToken, PunctuatorToken}},
		{"010xF", TTs{NumericToken, NumericToken, IdentifierToken}},
		{"50e+-0", TTs{NumericToken, IdentifierToken, PunctuatorToken, PunctuatorToken, NumericToken}},
		{"'str\\i\\'ng'", TTs{StringToken}},
		{"'str\\\\'abc", TTs{StringToken, IdentifierToken}},
		{"'str\\\ni\\\\u00A0ng'", TTs{StringToken}},
		{"a = /[a-z/]/g", TTs{IdentifierToken, PunctuatorToken, RegexpToken}},
		{"a=/=/g1", TTs{IdentifierToken, PunctuatorToken, RegexpToken}},
		{"a = /'\\\\/\n", TTs{IdentifierToken, PunctuatorToken, RegexpToken, LineTerminatorToken}},
		{"a=/\\//g1", TTs{IdentifierToken, PunctuatorToken, RegexpToken}},
		{"new RegExp(a + /\\d{1,2}/.source)", TTs{IdentifierToken, IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, RegexpToken, PunctuatorToken, IdentifierToken, PunctuatorToken}},

		{"0b0101 0o0707 0b17", TTs{NumericToken, NumericToken, NumericToken, NumericToken}},
		{"`template`", TTs{TemplateToken}},
		{"`a${x+y}b`", TTs{TemplateToken, IdentifierToken, PunctuatorToken, IdentifierToken, TemplateToken}},
		{"`temp\nlate`", TTs{TemplateToken}},
		{"`outer${{x: 10}}bar${ raw`nested${2}endnest` }end`", TTs{TemplateToken, PunctuatorToken, IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, TemplateToken, IdentifierToken, TemplateToken, NumericToken, TemplateToken, TemplateToken}},

		// early endings
		{"'string", TTs{StringToken}},
		{"'\n '\u2028", TTs{UnknownToken, LineTerminatorToken, UnknownToken, LineTerminatorToken}},
		{"'str\\\U00100000ing\\0'", TTs{StringToken}},
		{"'strin\\00g'", TTs{StringToken}},
		{"/*comment", TTs{CommentToken}},
		{"a=/regexp", TTs{IdentifierToken, PunctuatorToken, RegexpToken}},
		{"\\u002", TTs{UnknownToken, IdentifierToken}},

		// coverage
		{"Ø a〉", TTs{IdentifierToken, IdentifierToken, UnknownToken}},
		{"0xg 0.f", TTs{NumericToken, IdentifierToken, NumericToken, PunctuatorToken, IdentifierToken}},
		{"0bg 0og", TTs{NumericToken, IdentifierToken, NumericToken, IdentifierToken}},
		{"\u00A0\uFEFF\u2000", TTs{}},
		{"\u2028\u2029", TTs{LineTerminatorToken}},
		{"\\u0029ident", TTs{IdentifierToken}},
		{"\\u{0029FEF}ident", TTs{IdentifierToken}},
		{"\\u{}", TTs{UnknownToken, IdentifierToken, PunctuatorToken, PunctuatorToken}},
		{"\\ugident", TTs{UnknownToken, IdentifierToken}},
		{"'str\u2028ing'", TTs{UnknownToken, IdentifierToken, LineTerminatorToken, IdentifierToken, StringToken}},
		{"a=/\\\n", TTs{IdentifierToken, PunctuatorToken, PunctuatorToken, UnknownToken, LineTerminatorToken}},
		{"a=/x/\u200C\u3009", TTs{IdentifierToken, PunctuatorToken, RegexpToken, UnknownToken}},
		{"a=/x\n", TTs{IdentifierToken, PunctuatorToken, PunctuatorToken, IdentifierToken, LineTerminatorToken}},

		{"return /abc/;", TTs{IdentifierToken, RegexpToken, PunctuatorToken}},
		{"yield /abc/;", TTs{IdentifierToken, RegexpToken, PunctuatorToken}},
		{"a/b/g", TTs{IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, IdentifierToken}},
		{"{}/1/g", TTs{PunctuatorToken, PunctuatorToken, RegexpToken}},
		{"i(0)/1/g", TTs{IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, PunctuatorToken, NumericToken, PunctuatorToken, IdentifierToken}},
		{"if(0)/1/g", TTs{IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, RegexpToken}},
		{"a.if(0)/1/g", TTs{IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, PunctuatorToken, NumericToken, PunctuatorToken, IdentifierToken}},
		{"while(0)/1/g", TTs{IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, RegexpToken}},
		{"for(;;)/1/g", TTs{IdentifierToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, RegexpToken}},
		{"with(0)/1/g", TTs{IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, RegexpToken}},
		{"this/1/g", TTs{IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, IdentifierToken}},
		{"case /1/g:", TTs{IdentifierToken, RegexpToken, PunctuatorToken}},
		{"function f(){}/1/g", TTs{IdentifierToken, IdentifierToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, PunctuatorToken, RegexpToken}},
		{"this.return/1/g", TTs{IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, NumericToken, PunctuatorToken, IdentifierToken}},
		{"(a+b)/1/g", TTs{PunctuatorToken, IdentifierToken, PunctuatorToken, IdentifierToken, PunctuatorToken, PunctuatorToken, NumericToken, PunctuatorToken, IdentifierToken}},

		// go fuzz
		{"`", TTs{UnknownToken}},
	}

	for _, tt := range tokenTests {
		t.Run(tt.js, func(t *testing.T) {
			l := NewLexer(bytes.NewBufferString(tt.js))
			i := 0
			j := 0
			for {
				token, _ := l.Next()
				j++
				if token == ErrorToken {
					test.T(t, l.Err(), io.EOF)
					test.T(t, i, len(tt.expected), "when error occurred we must be at the end")
					break
				} else if token == WhitespaceToken {
					continue
				}
				if i < len(tt.expected) {
					if token != tt.expected[i] {
						test.String(t, token.String(), tt.expected[i].String(), "token types must match")
						break
					}
				} else {
					test.Fail(t, "index", i, "must not exceed expected token types size", len(tt.expected))
					break
				}
				i++
			}
		})
	}

	test.T(t, WhitespaceToken.String(), "Whitespace")
	test.T(t, TokenType(100).String(), "Invalid(100)")
}

////////////////////////////////////////////////////////////////

func ExampleNewLexer() {
	l := NewLexer(bytes.NewBufferString("var x = 'lorem ipsum';"))
	out := ""
	for {
		tt, data := l.Next()
		if tt == ErrorToken {
			break
		}
		out += string(data)
	}
	fmt.Println(out)
	// Output: var x = 'lorem ipsum';
}
