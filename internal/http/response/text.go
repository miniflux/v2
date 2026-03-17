// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import "net/http"

// Text writes a standard text response with a status 200 OK.
func Text(w http.ResponseWriter, r *http.Request, body string) {
	builder := NewBuilder(w, r)
	builder.WithHeader("Content-Type", `text/plain; charset=utf-8`)
	builder.WithBodyAsString(body)
	builder.Write()
}
