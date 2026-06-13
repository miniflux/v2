// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"slices"
	"strconv"
	"strings"
)

type acceptEncodingParser struct {
	// accepted contains all encoding that particular parser instance advertises.
	accepted []string
}

// AcceptEncoding creates parser instance for "Accept-Encoding" header values.
// It accepts list of encodings recognized by user of this parser instance.
func AcceptEncoding(accepted ...string) *acceptEncodingParser {
	return &acceptEncodingParser{accepted: accepted}
}

// Parse parses input string according to [HTTP Semantics]
// and returns first encoding that can be understood by us.
//
// Currently this function ignores set weights other than q=0.
// Encodings with q=0 will not be considered.
//
// If string is empty or no encoding was accepted function returns "identity".
//
// For "identity;q=0" and "*;q=0" function returns an empty string. In that case,
// if no other encoding was accepted, 406 Not Acceptable should be returned.
//
// [HTTP Semantics]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept-Encoding.
func (p *acceptEncodingParser) Parse(acceptEncoding string) string {
	accepted := "identity"

	for enc := range strings.SplitSeq(acceptEncoding, ",") {
		enc = strings.TrimSpace(enc)

		if qi := strings.IndexByte(enc, ';'); qi > -1 {
			qstr := strings.TrimPrefix(enc[qi:], ";")
			qstr = strings.TrimSpace(qstr)
			qstr = strings.TrimPrefix(qstr, "q=")

			q, err := strconv.ParseFloat(qstr, 64)
			if err != nil {
				continue // Ignore weird float values.
			}

			enc = strings.TrimSpace(enc[:qi])

			if q == 0 && slices.Contains([]string{"identity", "*"}, enc) {
				accepted = "" // Explicitly disabled, so can't be used as fallback.
				continue
			}

			if q == 0 {
				continue // Skipping unwanted.
			}
		}

		if !slices.Contains(p.accepted, enc) {
			continue // Skipping unsupported.
		}

		accepted = enc
		break
	}

	return accepted
}
