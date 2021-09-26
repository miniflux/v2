// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package client // import "miniflux.app/http/client"

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/errors"
	"miniflux.app/logger"
	"miniflux.app/timer"
)

const (
	defaultHTTPClientTimeout     = 20
	defaultHTTPClientMaxBodySize = 15 * 1024 * 1024
)

var (
	errInvalidCertificate        = "Invalid SSL certificate (original error: %q)"
	errTemporaryNetworkOperation = "This website is temporarily unreachable (original error: %q)"
	errPermanentNetworkOperation = "This website is permanently unreachable (original error: %q)"
	errRequestTimeout            = "Website unreachable, the request timed out after %d seconds"
)

// Client builds and executes HTTP requests.
type Client struct {
	inputURL string

	requestEtagHeader          string
	requestLastModifiedHeader  string
	requestAuthorizationHeader string
	requestUsername            string
	requestPassword            string
	requestUserAgent           string
	requestCookie              string

	useProxy             bool
	doNotFollowRedirects bool

	ClientTimeout               int
	ClientMaxBodySize           int64
	ClientProxyURL              string
	AllowSelfSignedCertificates bool
}

// New initializes a new HTTP client.
func New(url string) *Client {
	return &Client{
		inputURL:          url,
		ClientTimeout:     defaultHTTPClientTimeout,
		ClientMaxBodySize: defaultHTTPClientMaxBodySize,
	}
}

// NewClientWithConfig initializes a new HTTP client with application config options.
func NewClientWithConfig(url string, opts *config.Options) *Client {
	return &Client{
		inputURL:          url,
		requestUserAgent:  opts.HTTPClientUserAgent(),
		ClientTimeout:     opts.HTTPClientTimeout(),
		ClientMaxBodySize: opts.HTTPClientMaxBodySize(),
		ClientProxyURL:    opts.HTTPClientProxy(),
	}
}

func (c *Client) String() string {
	etagHeader := c.requestEtagHeader
	if c.requestEtagHeader == "" {
		etagHeader = "None"
	}

	lastModifiedHeader := c.requestLastModifiedHeader
	if c.requestLastModifiedHeader == "" {
		lastModifiedHeader = "None"
	}

	return fmt.Sprintf(
		`InputURL=%q ETag=%s LastMod=%s Auth=%v UserAgent=%q Verify=%v`,
		c.inputURL,
		etagHeader,
		lastModifiedHeader,
		c.requestAuthorizationHeader != "" || (c.requestUsername != "" && c.requestPassword != ""),
		c.requestUserAgent,
		!c.AllowSelfSignedCertificates,
	)
}

// WithCredentials defines the username/password for HTTP Basic authentication.
func (c *Client) WithCredentials(username, password string) *Client {
	if username != "" && password != "" {
		c.requestUsername = username
		c.requestPassword = password
	}
	return c
}

// WithAuthorization defines the authorization HTTP header value.
func (c *Client) WithAuthorization(authorization string) *Client {
	c.requestAuthorizationHeader = authorization
	return c
}

// WithCacheHeaders defines caching headers.
func (c *Client) WithCacheHeaders(etagHeader, lastModifiedHeader string) *Client {
	c.requestEtagHeader = etagHeader
	c.requestLastModifiedHeader = lastModifiedHeader
	return c
}

// WithProxy enables proxy for the current HTTP request.
func (c *Client) WithProxy() *Client {
	c.useProxy = true
	return c
}

// WithoutRedirects disables HTTP redirects.
func (c *Client) WithoutRedirects() *Client {
	c.doNotFollowRedirects = true
	return c
}

// WithUserAgent defines the User-Agent header to use for HTTP requests.
func (c *Client) WithUserAgent(userAgent string) *Client {
	if userAgent != "" {
		c.requestUserAgent = userAgent
	}
	return c
}

// WithCookie defines the Cookies to use for HTTP requests.
func (c *Client) WithCookie(cookie string) *Client {
	if cookie != "" {
		c.requestCookie = cookie
	}
	return c
}

// Get performs a GET HTTP request.
func (c *Client) Get() (*Response, error) {
	request, err := c.buildRequest(http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest(request)
}

// PostForm performs a POST HTTP request with form encoded values.
func (c *Client) PostForm(values url.Values) (*Response, error) {
	request, err := c.buildRequest(http.MethodPost, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.executeRequest(request)
}

// PostJSON performs a POST HTTP request with a JSON payload.
func (c *Client) PostJSON(data interface{}) (*Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	request, err := c.buildRequest(http.MethodPost, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	return c.executeRequest(request)
}

func (c *Client) executeRequest(request *http.Request) (*Response, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[HttpClient] inputURL=%s", c.inputURL))

	logger.Debug("[HttpClient:Before] Method=%s %s",
		request.Method,
		c.String(),
	)

	client := c.buildClient()
	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		if uerr, ok := err.(*url.Error); ok {
			switch uerr.Err.(type) {
			case x509.CertificateInvalidError, x509.HostnameError:
				err = errors.NewLocalizedError(errInvalidCertificate, uerr.Err)
			case *net.OpError:
				if uerr.Err.(*net.OpError).Temporary() {
					err = errors.NewLocalizedError(errTemporaryNetworkOperation, uerr.Err)
				} else {
					err = errors.NewLocalizedError(errPermanentNetworkOperation, uerr.Err)
				}
			case net.Error:
				nerr := uerr.Err.(net.Error)
				if nerr.Timeout() {
					err = errors.NewLocalizedError(errRequestTimeout, c.ClientTimeout)
				} else if nerr.Temporary() {
					err = errors.NewLocalizedError(errTemporaryNetworkOperation, nerr)
				}
			}
		}

		return nil, err
	}

	if resp.ContentLength > c.ClientMaxBodySize {
		return nil, fmt.Errorf("client: response too large (%d bytes)", resp.ContentLength)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: error while reading body %v", err)
	}

	response := &Response{
		Body:          bytes.NewReader(buf),
		StatusCode:    resp.StatusCode,
		EffectiveURL:  resp.Request.URL.String(),
		LastModified:  resp.Header.Get("Last-Modified"),
		ETag:          resp.Header.Get("ETag"),
		Expires:       resp.Header.Get("Expires"),
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
	}

	logger.Debug("[HttpClient:After] Method=%s %s; Response => %s",
		request.Method,
		c.String(),
		response,
	)

	// Ignore caching headers for feeds that do not want any cache.
	if resp.Header.Get("Expires") == "0" {
		logger.Debug("[HttpClient] Ignore caching headers for %q", response.EffectiveURL)
		response.ETag = ""
		response.LastModified = ""
	}

	return response, err
}

func (c *Client) buildRequest(method string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, c.inputURL, body)
	if err != nil {
		return nil, err
	}

	request.Header = c.buildHeaders()

	if c.requestUsername != "" && c.requestPassword != "" {
		request.SetBasicAuth(c.requestUsername, c.requestPassword)
	}

	return request, nil
}

func (c *Client) buildClient() http.Client {
	client := http.Client{
		Timeout: time.Duration(c.ClientTimeout) * time.Second,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			// Default is 30s.
			Timeout: 10 * time.Second,

			// Default is 30s.
			KeepAlive: 15 * time.Second,
		}).DialContext,

		// Default is 100.
		MaxIdleConns: 50,

		// Default is 90s.
		IdleConnTimeout: 10 * time.Second,
	}

	if c.AllowSelfSignedCertificates {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if c.doNotFollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if c.useProxy && c.ClientProxyURL != "" {
		proxyURL, err := url.Parse(c.ClientProxyURL)
		if err != nil {
			logger.Error("[HttpClient] Proxy URL error: %v", err)
		} else {
			logger.Debug("[HttpClient] Use proxy: %s", proxyURL)
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client.Transport = transport

	return client
}

func (c *Client) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("Accept", "*/*")

	if c.requestUserAgent != "" {
		headers.Add("User-Agent", c.requestUserAgent)
	}

	if c.requestEtagHeader != "" {
		headers.Add("If-None-Match", c.requestEtagHeader)
	}

	if c.requestLastModifiedHeader != "" {
		headers.Add("If-Modified-Since", c.requestLastModifiedHeader)
	}

	if c.requestAuthorizationHeader != "" {
		headers.Add("Authorization", c.requestAuthorizationHeader)
	}

	if c.requestCookie != "" {
		headers.Add("Cookie", c.requestCookie)
	}

	headers.Add("Connection", "close")
	return headers
}
