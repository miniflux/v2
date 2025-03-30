// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package response // import "miniflux.app/v2/internal/http/response"

// ContentSecurityPolicyForUntrustedContent is the default CSP for untrusted content.
// default-src 'none' disables all content sources
// form-action 'none' disables all form submissions
// sandbox enables a sandbox for the requested resource
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/form-action
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/sandbox
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy/default-src
const ContentSecurityPolicyForUntrustedContent = `default-src 'none'; form-action 'none'; sandbox;`
