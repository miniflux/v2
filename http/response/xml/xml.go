// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package xml // import "miniflux.app/http/response/xml"

import (
	"net/http"

	"miniflux.app/http/response"
)

// OK writes a standard XML response with a status 200 OK.
func OK(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithBody(body)
	builder.Write()
}

// Attachment forces the XML document to be downloaded by the web browser.
func Attachment(w http.ResponseWriter, r *http.Request, filename string, body interface{}) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", "text/xml; charset=utf-8")
	builder.WithAttachment(filename)
	builder.WithBody(body)
	builder.Write()
}
