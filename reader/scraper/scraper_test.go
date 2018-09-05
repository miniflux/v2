// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper // import "miniflux.app/reader/scraper"

import "testing"

func TestGetPredefinedRules(t *testing.T) {
	if getPredefinedScraperRules("http://www.phoronix.com/") == "" {
		t.Error("Unable to find rule for phoronix.com")
	}

	if getPredefinedScraperRules("https://www.linux.com/") == "" {
		t.Error("Unable to find rule for linux.com")
	}

	if getPredefinedScraperRules("https://example.org/") != "" {
		t.Error("A rule not defined should not return anything")
	}
}
