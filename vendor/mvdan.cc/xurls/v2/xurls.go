// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

// Package xurls extracts urls from plain text using regular expressions.
package xurls

import (
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:generate go run ./generate/tldsgen
//go:generate go run ./generate/schemesgen
//go:generate go run ./generate/unicodegen

const (
	// pathCont is based on https://www.rfc-editor.org/rfc/rfc3987#section-2.2
	// but does not match separators anywhere or most puncutation in final position,
	// to avoid creating asymmetries like
	// `Did you know that **<a href="...">https://example.com/**</a> is reserved for documentation?`
	// from `Did you know that **https://example.com/** is reserved for documentation?`.
	unreservedChar      = `a-zA-Z0-9\-._~`
	endUnreservedChar   = `a-zA-Z0-9\-_~`
	midSubDelimChar     = `!$&'*+,;=`
	endSubDelimChar     = `$&+=`
	midIPathSegmentChar = unreservedChar + `%` + midSubDelimChar + `:@` + allowedUcsChar
	endIPathSegmentChar = endUnreservedChar + `%` + endSubDelimChar + allowedUcsCharMinusPunc
	iPrivateChar        = `\x{E000}-\x{F8FF}\x{F0000}-\x{FFFFD}\x{100000}-\x{10FFFD}`
	midIChar            = `/?#\\` + midIPathSegmentChar + iPrivateChar
	endIChar            = `/#` + endIPathSegmentChar + iPrivateChar
	wellParen           = `\((?:[` + midIChar + `]|\([` + midIChar + `]*\))*\)`
	wellBrack           = `\[(?:[` + midIChar + `]|\[[` + midIChar + `]*\])*\]`
	wellBrace           = `\{(?:[` + midIChar + `]|\{[` + midIChar + `]*\})*\}`
	wellAll             = wellParen + `|` + wellBrack + `|` + wellBrace
	pathCont            = `(?:[` + midIChar + `]*(?:` + wellAll + `|[` + endIChar + `]))+`

	letter    = `\p{L}`
	mark      = `\p{M}`
	number    = `\p{N}`
	iriChar   = letter + mark + number
	iri       = `[` + iriChar + `](?:[` + iriChar + `\-]*[` + iriChar + `])?`
	subdomain = `(?:` + iri + `\.)+`
	octet     = `(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9][0-9]|[0-9])`
	ipv4Addr  = octet + `\.` + octet + `\.` + octet + `\.` + octet

	// ipv6Addr is based on https://datatracker.ietf.org/doc/html/rfc4291#section-2.2
	// with a specific alternative for each valid count of leading 16-bit hexadecimal "chomps"
	// that have not been replaced with a `::` elision.
	h4                 = `[0-9a-fA-F]{1,4}`
	ipv6AddrMinusEmpty = `(?:` +
		// 7 colon-terminated chomps, followed by a final chomp or the rest of an elision.
		`(?:` + h4 + `:){7}(?:` + h4 + `|:)|` +
		// 6 chomps, followed by an IPv4 address or elision with final chomp or final elision.
		`(?:` + h4 + `:){6}(?:` + ipv4Addr + `|:` + h4 + `|:)|` +
		// 5 chomps, followed by an elision with optional IPv4 or up to 2 final chomps.
		`(?:` + h4 + `:){5}(?::` + ipv4Addr + `|(?::` + h4 + `){1,2}|:)|` +
		// 4 chomps, followed by an elision with optional IPv4 (optionally preceded by a chomp) or
		// up to 3 final chomps.
		`(?:` + h4 + `:){4}(?:(?::` + h4 + `){0,1}:` + ipv4Addr + `|(?::` + h4 + `){1,3}|:)|` +
		// 3 chomps, followed by an elision with optional IPv4 (preceded by up to 2 chomps) or
		// up to 4 final chomps.
		`(?:` + h4 + `:){3}(?:(?::` + h4 + `){0,2}:` + ipv4Addr + `|(?::` + h4 + `){1,4}|:)|` +
		// 2 chomps, followed by an elision with optional IPv4 (preceded by up to 3 chomps) or
		// up to 5 final chomps.
		`(?:` + h4 + `:){2}(?:(?::` + h4 + `){0,3}:` + ipv4Addr + `|(?::` + h4 + `){1,5}|:)|` +
		// 1 chomp, followed by an elision with optional IPv4 (preceded by up to 4 chomps) or
		// up to 6 final chomps.
		`(?:` + h4 + `:){1}(?:(?::` + h4 + `){0,4}:` + ipv4Addr + `|(?::` + h4 + `){1,6}|:)|` +
		// elision, followed by optional IPv4 (preceded by up to 5 chomps) or
		// up to 7 final chomps.
		// `:` is an intentionally omitted alternative, to avoid matching `::`.
		`:(?:(?::` + h4 + `){0,5}:` + ipv4Addr + `|(?::` + h4 + `){1,7})` +
		`)`
	ipv6Addr         = `(?:` + ipv6AddrMinusEmpty + `|::)`
	ipAddrMinusEmpty = `(?:` + ipv6AddrMinusEmpty + `|\b` + ipv4Addr + `\b)`
	port             = `(?::[0-9]*)?`
)

// AnyScheme can be passed to StrictMatchingScheme to match any possibly valid
// scheme, and not just the known ones.
var AnyScheme = `(?:[a-zA-Z][a-zA-Z.\-+]*://|` + anyOf(SchemesNoAuthority...) + `:)`

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
	`gemini`,        // gemini
	`jdbc`,          // Java database Connectivity
	`moz-extension`, // Firefox extension
	`postgres`,      // PostgreSQL (short form)
	`postgresql`,    // PostgreSQL
	`slack`,         // Slack
	`zoommtg`,       // Zoom (desktop)
	`zoomus`,        // Zoom (mobile)
}

// The regular expressions are compiled when the API is first called.
// Any subsequent calls will use the same regular expression pointers.
//
// We do not need to make a copy of them for each API call,
// as Copy is now only useful if one copy calls Longest but not another,
// and we always call Longest after compiling the regular expression.
var (
	strictRe   *regexp.Regexp
	strictInit sync.Once

	relaxedRe   *regexp.Regexp
	relaxedInit sync.Once
)

func anyOf(strs ...string) string {
	var b strings.Builder
	b.WriteString("(?:")
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
	schemes := `(?:(?i)(?:` + anyOf(Schemes...) + `|` + anyOf(SchemesUnofficial...) + `)://|` + anyOf(SchemesNoAuthority...) + `:)`
	return schemes + pathCont
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
	tlds := `(?:(?i)` + punycode + `|` + anyOf(append(asciiTLDs, PseudoTLDs...)...) + `\b|` + anyOf(unicodeTLDs...) + `)`
	domain := subdomain + tlds

	hostName := `(?:` + domain + `|\[` + ipv6Addr + `\]|\b` + ipv4Addr + `\b)`
	webURL := hostName + port + `(?:/` + pathCont + `|/)?`
	email := `[a-zA-Z0-9._%\-+]+@` + domain
	return strictExp() + `|` + webURL + `|` + email + `|` + ipv6AddrMinusEmpty
}

// Strict produces a regexp that matches any URL with a scheme in either the
// Schemes or SchemesNoAuthority lists.
func Strict() *regexp.Regexp {
	strictInit.Do(func() {
		strictRe = regexp.MustCompile(strictExp())
		strictRe.Longest()
	})
	return strictRe
}

// Relaxed produces a regexp that matches any URL matched by Strict, plus any
// URL with no scheme or email address.
func Relaxed() *regexp.Regexp {
	relaxedInit.Do(func() {
		relaxedRe = regexp.MustCompile(relaxedExp())
		relaxedRe.Longest()
	})
	return relaxedRe
}

// StrictMatchingScheme produces a regexp similar to Strict, but requiring that
// the scheme match the given regular expression. See AnyScheme too.
func StrictMatchingScheme(exp string) (*regexp.Regexp, error) {
	strictMatching := `(?i)(?:` + exp + `)(?-i)` + pathCont
	re, err := regexp.Compile(strictMatching)
	if err != nil {
		return nil, err
	}
	re.Longest()
	return re, nil
}
