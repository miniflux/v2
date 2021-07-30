// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

// Package xurls extracts urls from plain text using regular expressions.
package xurls

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

//go:generate go run ./generate/tldsgen
//go:generate go run ./generate/schemesgen
//go:generate go run ./generate/unicodegen

const (
	letter    = `\p{L}`
	mark      = `\p{M}`
	number    = `\p{N}`
	iriChar   = letter + mark + number
	currency  = `\p{Sc}`
	otherSymb = `\p{So}`
	endChar   = iriChar + `/\-_+&~%=#` + currency + otherSymb
	midChar   = endChar + "_*" + otherPuncMinusDoubleQuote
	wellParen = `\([` + midChar + `]*(\([` + midChar + `]*\)[` + midChar + `]*)*\)`
	wellBrack = `\[[` + midChar + `]*(\[[` + midChar + `]*\][` + midChar + `]*)*\]`
	wellBrace = `\{[` + midChar + `]*(\{[` + midChar + `]*\}[` + midChar + `]*)*\}`
	wellAll   = wellParen + `|` + wellBrack + `|` + wellBrace
	pathCont  = `([` + midChar + `]*(` + wellAll + `|[` + endChar + `])+)+`

	iri      = `[` + iriChar + `]([` + iriChar + `\-]*[` + iriChar + `])?`
	domain   = `(` + iri + `\.)+`
	octet    = `(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])`
	ipv4Addr = `\b` + octet + `\.` + octet + `\.` + octet + `\.` + octet + `\b`
	ipv6Addr = `([0-9a-fA-F]{1,4}:([0-9a-fA-F]{1,4}:([0-9a-fA-F]{1,4}:([0-9a-fA-F]{1,4}:([0-9a-fA-F]{1,4}:[0-9a-fA-F]{0,4}|:[0-9a-fA-F]{1,4})?|(:[0-9a-fA-F]{1,4}){0,2})|(:[0-9a-fA-F]{1,4}){0,3})|(:[0-9a-fA-F]{1,4}){0,4})|:(:[0-9a-fA-F]{1,4}){0,5})((:[0-9a-fA-F]{1,4}){2}|:(25[0-5]|(2[0-4]|1[0-9]|[1-9])?[0-9])(\.(25[0-5]|(2[0-4]|1[0-9]|[1-9])?[0-9])){3})|(([0-9a-fA-F]{1,4}:){1,6}|:):[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){7}:`
	ipAddr   = `(` + ipv4Addr + `|` + ipv6Addr + `)`
	port     = `(:[0-9]*)?`
)

// AnyScheme can be passed to StrictMatchingScheme to match any possibly valid
// scheme, and not just the known ones.
var AnyScheme = `([a-zA-Z][a-zA-Z.\-+]*://|` + anyOf(SchemesNoAuthority...) + `:)`

// SchemesNoAuthority is a sorted list of some well-known url schemes that are
// followed by ":" instead of "://". The list includes both officially
// registered and unofficial schemes.
var SchemesNoAuthority = []string{
	`bitcoin`, // Bitcoin
	`cid`,     // Content-ID
	`file`,    // Files
	`magnet`,  // Torrent magnets
	`mailto`,  // Mail
	`mid`,     // Message-ID
	`sms`,     // SMS
	`tel`,     // Telephone
	`xmpp`,    // XMPP
}

// SchemesUnofficial is a sorted list of some well-known url schemes which
// aren't officially registered just yet. They tend to correspond to software.
//
// Mostly collected from https://en.wikipedia.org/wiki/List_of_URI_schemes#Unofficial_but_common_URI_schemes.
var SchemesUnofficial = []string{
	`jdbc`,       // Java database Connectivity
	`postgres`,   // PostgreSQL (short form)
	`postgresql`, // PostgreSQL
	`slack`,      // Slack
	`zoommtg`,    // Zoom (desktop)
	`zoomus`,     // Zoom (mobile)
}

func anyOf(strs ...string) string {
	var b strings.Builder
	b.WriteByte('(')
	for i, s := range strs {
		if i != 0 {
			b.WriteByte('|')
		}
		b.WriteString(regexp.QuoteMeta(s))
	}
	b.WriteByte(')')
	return b.String()
}

func strictExp() string {
	schemes := `((` + anyOf(Schemes...) + `|` + anyOf(SchemesUnofficial...) + `)://|` + anyOf(SchemesNoAuthority...) + `:)`
	return `(?i)` + schemes + `(?-i)` + pathCont
}

func relaxedExp() string {
	var asciiTLDs, unicodeTLDs []string
	for i, tld := range TLDs {
		if tld[0] >= utf8.RuneSelf {
			asciiTLDs = TLDs[:i:i]
			unicodeTLDs = TLDs[i:]
			break
		}
	}
	punycode := `xn--[a-z0-9-]+`

	// Use \b to make sure ASCII TLDs are immediately followed by a word break.
	// We can't do that with unicode TLDs, as they don't see following
	// whitespace as a word break.
	tlds := `(?i)(` + punycode + `|` + anyOf(append(asciiTLDs, PseudoTLDs...)...) + `\b|` + anyOf(unicodeTLDs...) + `)(?-i)`
	site := domain + tlds

	hostName := `(` + site + `|` + ipAddr + `)`
	webURL := hostName + port + `(/|/` + pathCont + `)?`
	email := `[a-zA-Z0-9._%\-+]+@` + site
	return strictExp() + `|` + webURL + `|` + email
}

// Strict produces a regexp that matches any URL with a scheme in either the
// Schemes or SchemesNoAuthority lists.
func Strict() *regexp.Regexp {
	re := regexp.MustCompile(strictExp())
	re.Longest()
	return re
}

// Relaxed produces a regexp that matches any URL matched by Strict, plus any
// URL with no scheme or email address.
func Relaxed() *regexp.Regexp {
	re := regexp.MustCompile(relaxedExp())
	re.Longest()
	return re
}

// StrictMatchingScheme produces a regexp similar to Strict, but requiring that
// the scheme match the given regular expression. See AnyScheme too.
func StrictMatchingScheme(exp string) (*regexp.Regexp, error) {
	strictMatching := `(?i)(` + exp + `)(?-i)` + pathCont
	re, err := regexp.Compile(strictMatching)
	if err != nil {
		return nil, err
	}
	re.Longest()
	return re, nil
}
