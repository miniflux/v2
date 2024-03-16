// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package parser // import "miniflux.app/v2/internal/reader/parser"

import (
	"strings"
	"testing"
)

func TestDetectRDF(t *testing.T) {
	data := `<?xml version="1.0"?><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns="http://my.netscape.com/rdf/simple/0.9/"></rdf:RDF>`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatRDF {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatRDF)
	}
}

func TestDetectRSS(t *testing.T) {
	data := `<?xml version="1.0"?><rss version="2.0"><channel></channel></rss>`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatRSS {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatRSS)
	}
}

func TestDetectAtom10(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?><feed xmlns="http://www.w3.org/2005/Atom"></feed>`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatAtom {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatAtom)
	}
}

func TestDetectAtom03(t *testing.T) {
	data := `<?xml version="1.0" encoding="utf-8"?><feed version="0.3" xmlns="http://purl.org/atom/ns#" xmlns:dc="http://purl.org/dc/elements/1.1/" xml:lang="en"></feed>`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatAtom {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatAtom)
	}
}

func TestDetectAtomWithISOCharset(t *testing.T) {
	data := `<?xml version="1.0" encoding="ISO-8859-15"?><feed xmlns="http://www.w3.org/2005/Atom"></feed>`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatAtom {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatAtom)
	}
}

func TestDetectJSON(t *testing.T) {
	data := `
	{
		"version" : "https://jsonfeed.org/version/1",
		"title" : "Example"
	}
	`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatJSON {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatJSON)
	}
}

func TestDetectUnknown(t *testing.T) {
	data := `
	<!DOCTYPE html> <html> </html>
	`
	format, _ := DetectFeedFormat(strings.NewReader(data))

	if format != FormatUnknown {
		t.Errorf(`Wrong format detected: %q instead of %q`, format, FormatUnknown)
	}
}
