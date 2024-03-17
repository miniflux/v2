// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"miniflux.app/v2/internal/locale"
)

type ResponseHandler struct {
	httpResponse *http.Response
	clientErr    error
}

func NewResponseHandler(httpResponse *http.Response, clientErr error) *ResponseHandler {
	return &ResponseHandler{httpResponse: httpResponse, clientErr: clientErr}
}

func (r *ResponseHandler) EffectiveURL() string {
	return r.httpResponse.Request.URL.String()
}

func (r *ResponseHandler) ContentType() string {
	return r.httpResponse.Header.Get("Content-Type")
}

func (r *ResponseHandler) LastModified() string {
	// Ignore caching headers for feeds that do not want any cache.
	if r.httpResponse.Header.Get("Expires") == "0" {
		return ""
	}
	return r.httpResponse.Header.Get("Last-Modified")
}

func (r *ResponseHandler) ETag() string {
	// Ignore caching headers for feeds that do not want any cache.
	if r.httpResponse.Header.Get("Expires") == "0" {
		return ""
	}
	return r.httpResponse.Header.Get("ETag")
}

func (r *ResponseHandler) IsModified(lastEtagValue, lastModifiedValue string) bool {
	if r.httpResponse.StatusCode == http.StatusNotModified {
		return false
	}

	if r.ETag() != "" && r.ETag() == lastEtagValue {
		return false
	}

	if r.LastModified() != "" && r.LastModified() == lastModifiedValue {
		return false
	}

	return true
}

func (r *ResponseHandler) Close() {
	if r.httpResponse != nil && r.httpResponse.Body != nil && r.clientErr == nil {
		r.httpResponse.Body.Close()
	}
}

func (r *ResponseHandler) Body(maxBodySize int64) io.ReadCloser {
	return http.MaxBytesReader(nil, r.httpResponse.Body, maxBodySize)
}

func (r *ResponseHandler) ReadBody(maxBodySize int64) ([]byte, *locale.LocalizedErrorWrapper) {
	limitedReader := http.MaxBytesReader(nil, r.httpResponse.Body, maxBodySize)

	buffer, err := io.ReadAll(limitedReader)
	if err != nil && err != io.EOF {
		if err, ok := err.(*http.MaxBytesError); ok {
			return nil, locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: response body too large: %d bytes", err.Limit), "error.http_response_too_large")
		}

		return nil, locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: unable to read response body: %w", err), "error.http_body_read", err)
	}

	if len(buffer) == 0 {
		return nil, locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: empty response body"), "error.http_empty_response_body")
	}

	return buffer, nil
}

func (r *ResponseHandler) LocalizedError() *locale.LocalizedErrorWrapper {
	if r.clientErr != nil {
		switch {
		case isSSLError(r.clientErr):
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: %w", r.clientErr), "error.tls_error", r.clientErr)
		case isNetworkError(r.clientErr):
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: %w", r.clientErr), "error.network_operation", r.clientErr)
		case os.IsTimeout(r.clientErr):
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: %w", r.clientErr), "error.network_timeout", r.clientErr)
		case errors.Is(r.clientErr, io.EOF):
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: %w", r.clientErr), "error.http_empty_response")
		default:
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: %w", r.clientErr), "error.http_client_error", r.clientErr)
		}
	}

	switch r.httpResponse.StatusCode {
	case http.StatusUnauthorized:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: access unauthorized (401 status code)"), "error.http_not_authorized")
	case http.StatusForbidden:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: access forbidden (403 status code)"), "error.http_forbidden")
	case http.StatusTooManyRequests:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: too many requests (429 status code)"), "error.http_too_many_requests")
	case http.StatusNotFound, http.StatusGone:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: resource not found (%d status code)", r.httpResponse.StatusCode), "error.http_resource_not_found")
	case http.StatusInternalServerError:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: remote server error (%d status code)", r.httpResponse.StatusCode), "error.http_internal_server_error")
	case http.StatusBadGateway:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: bad gateway (%d status code)", r.httpResponse.StatusCode), "error.http_bad_gateway")
	case http.StatusServiceUnavailable:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: service unavailable (%d status code)", r.httpResponse.StatusCode), "error.http_service_unavailable")
	case http.StatusGatewayTimeout:
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: gateway timeout (%d status code)", r.httpResponse.StatusCode), "error.http_gateway_timeout")
	}

	if r.httpResponse.StatusCode >= 400 {
		return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: unexpected status code (%d status code)", r.httpResponse.StatusCode), "error.http_unexpected_status_code", r.httpResponse.StatusCode)
	}

	if r.httpResponse.StatusCode != 304 {
		// Content-Length = -1 when no Content-Length header is sent.
		if r.httpResponse.ContentLength == 0 {
			return locale.NewLocalizedErrorWrapper(fmt.Errorf("fetcher: empty response body"), "error.http_empty_response_body")
		}
	}

	return nil
}

func isNetworkError(err error) bool {
	if _, ok := err.(*url.Error); ok {
		return true
	}
	if err == io.EOF {
		return true
	}
	var opErr *net.OpError
	if ok := errors.As(err, &opErr); ok {
		return true
	}
	return false
}

func isSSLError(err error) bool {
	var certErr x509.UnknownAuthorityError
	if errors.As(err, &certErr) {
		return true
	}

	var hostErr x509.HostnameError
	if errors.As(err, &hostErr) {
		return true
	}

	var algErr x509.InsecureAlgorithmError
	return errors.As(err, &algErr)
}
