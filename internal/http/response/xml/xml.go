// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package xml // import "miniflux.app/v2/internal/http/response/xml"

import (
	"net/http"

	"miniflux.app/v2/internal/http/response"
)

// OK writes a standard XML response with a status 200 OK.
func OK[T []byte | string](w http.ResponseWriter, r *http.Request, body T) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithBody(body)
	builder.Write()
}

// Attachment forces the XML document to be downloaded by the web browser.
func Attachment[T []byte | string](w http.ResponseWriter, r *http.Request, filename string, body T) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithAttachment(filename)
	builder.WithBody(body)
	builder.Write()
}
