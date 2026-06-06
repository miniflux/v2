package storage

import "strings"

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
