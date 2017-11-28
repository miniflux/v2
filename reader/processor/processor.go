// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package processor

import (
	"github.com/miniflux/miniflux2/reader/rewrite"
	"github.com/miniflux/miniflux2/reader/sanitizer"
)

// ItemContentProcessor executes a set of functions to sanitize and alter item contents.
func ItemContentProcessor(url, content string) string {
	content = sanitizer.Sanitize(url, content)
	return rewrite.Rewriter(url, content)
}
