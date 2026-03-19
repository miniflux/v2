// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

import "net/http"

// ContentSecurityPolicyForUntrustedContent is the default CSP for untrusted content.
// default-src 'none' disables all content sources
// form-action 'none' disables all form submissions
// sandbox enables a sandbox for the requested resource
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/form-action
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/sandbox
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/default-src
const ContentSecurityPolicyForUntrustedContent = `default-src 'none'; form-action 'none'; sandbox;`

// NoContent sends a no content response to the client.
func NoContent(w http.ResponseWriter, r *http.Request) {
	builder := NewBuilder(w, r)
	builder.WithStatus(http.StatusNoContent)
	builder.Write()
}
