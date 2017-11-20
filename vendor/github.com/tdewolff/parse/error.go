package parse

import (
	"fmt"
	"io"

	"github.com/tdewolff/parse/buffer"
)

type Error struct {
	Message string
	Line    int
	Col     int
	Context string
}

func NewError(msg string, r io.Reader, offset int) *Error {
	line, col, context, _ := Position(r, offset)
	return &Error{
		msg,
		line,
		col,
		context,
	}
}

func NewErrorLexer(msg string, l *buffer.Lexer) *Error {
	r := buffer.NewReader(l.Bytes())
	offset := l.Offset()
	return NewError(msg, r, offset)
}

func (e *Error) Error() string {
	return fmt.Sprintf("parse error:%d:%d: %s\n%s", e.Line, e.Col, e.Message, e.Context)
}
