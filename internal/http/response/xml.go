// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import "net/http"

// XML writes a standard XML response with a status 200 OK.
func XML[T []byte | string](w http.ResponseWriter, r *http.Request, body T) {
	builder := New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithBody(body)
	builder.Write()
}

// XMLAttachment forces the XML document to be downloaded by the web browser.
func XMLAttachment[T []byte | string](w http.ResponseWriter, r *http.Request, filename string, body T) {
	builder := New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithAttachment(filename)
	builder.WithBody(body)
	builder.Write()
}
