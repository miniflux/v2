package storage

import (
	"fmt"
	"slices"
	"strings"
)

// whereBuilder constructs WHERE expression string using [strings.Builder].
type whereBuilder struct {
	sb strings.Builder
}

// String returns WHERE condidion string, including WHERE keyword.
func (b *whereBuilder) String() string {
	return b.sb.String()
}

func (b *whereBuilder) and(s string) {
	if b.sb.Len() == 0 {
		b.sb.WriteString("WHERE ")
	} else {
		b.sb.WriteString(" AND ")
	}

	b.sb.WriteString(s)
}

func (b *whereBuilder) andf(format string, args ...any) {
	if b.sb.Len() == 0 {
		b.sb.WriteString("WHERE ")
	} else {
		b.sb.WriteString(" AND ")
	}

	fmt.Fprintf(&b.sb, format, args...)
}

// orderByBuilder constructs ORDER BY expression string using [strings.Builder].
type orderByBuilder struct {
	sb strings.Builder
}

// String returns ORDER BY expression string, including ORDER BY keyword.
func (b *orderByBuilder) String() string {
	return b.sb.String()
}

func (b *orderByBuilder) asc(column string) {
	if b.sb.Len() == 0 {
		b.sb.WriteString("ORDER BY ")
	} else {
		b.sb.WriteString(", ")
	}

	b.sb.WriteString(column)
	b.sb.WriteString(" ASC")
}

func (b *orderByBuilder) desc(column string) {
	if b.sb.Len() == 0 {
		b.sb.WriteString("ORDER BY ")
	} else {
		b.sb.WriteString(", ")
	}

	b.sb.WriteString(column)
	b.sb.WriteString(" DESC")
}

// argsBuilder collects all parametrized args.
type argsBuilder struct {
	args []any
}

func (b *argsBuilder) clone() argsBuilder {
	return argsBuilder{args: slices.Clone(b.args)}
}

func (b *argsBuilder) all() []any {
	return b.args
}

func (b *argsBuilder) append(arg any) int {
	b.args = append(b.args, arg)
	return len(b.args)
}
