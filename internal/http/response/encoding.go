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

// Parse the input string according to [HTTP Semantics] and return our
// most-preferred encoding that the client accepts. Candidates are matched in
// the order of the accepted list passed to [AcceptEncoding], so the server's
// preference wins over the order (and weights) the client advertises.
//
// Currently this function ignores set weights other than q=0.
// Encodings with q=0 will not be considered.
//
// Returns "identity" if the input is empty or no accepted encoding matches.
//
// [HTTP Semantics]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept-Encoding.
func (p *acceptEncodingParser) Parse(acceptEncoding string) string {
	// bestIdx tracks the lowest index into p.accepted (i.e. the server's most
	// preferred encoding) seen so far. Tracking it inline avoids collecting the
	// client's encodings into a slice, keeping this method allocation free.
	bestIdx := -1

	for enc := range strings.SplitSeq(acceptEncoding, ",") {
		enc = strings.TrimSpace(enc)

		if qi := strings.IndexByte(enc, ';'); qi > -1 {
			qstr := strings.TrimPrefix(enc[qi:], ";")
			qstr = strings.TrimSpace(qstr)
			qstr = strings.TrimPrefix(qstr, "q=")

			q, err := strconv.ParseFloat(qstr, 64)
			if err != nil || q == 0 {
				continue // Ignore weird float values.
			}
			enc = strings.TrimSpace(enc[:qi])
		}

		if i := slices.Index(p.accepted, enc); i > -1 && (bestIdx == -1 || i < bestIdx) {
			bestIdx = i
			if bestIdx == 0 {
				break // Nothing can beat the top server preference.
			}
		}
	}

	if bestIdx == -1 {
		return "identity"
	}
	return p.accepted[bestIdx]
}
