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
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/errors"
	"miniflux.app/logger"
	"miniflux.app/timer"
	"miniflux.app/version"
)

var (
	// DefaultUserAgent sets the User-Agent header used for any requests by miniflux.
	DefaultUserAgent = "Mozilla/5.0 (compatible; Miniflux/" + version.Version + "; +https://miniflux.app)"

	errInvalidCertificate        = "Invalid SSL certificate (original error: %q)"
	errTemporaryNetworkOperation = "This website is temporarily unreachable (original error: %q)"
	errPermanentNetworkOperation = "This website is permanently unreachable (original error: %q)"
	errRequestTimeout            = "Website unreachable, the request timed out after %d seconds"
)

// Client is a HTTP Client :)
type Client struct {
	cfg                 *config.Config
	url                 string
	etagHeader          string
	lastModifiedHeader  string
	authorizationHeader string
	username            string
	password            string
	userAgent           string
	Insecure            bool
}

// WithCredentials defines the username/password for HTTP Basic authentication.
func (c *Client) WithCredentials(username, password string) *Client {
	if username != "" && password != "" {
		c.username = username
		c.password = password
	}
	return c
}

// WithAuthorization defines authorization header value.
func (c *Client) WithAuthorization(authorization string) *Client {
	c.authorizationHeader = authorization
	return c
}

// WithCacheHeaders defines caching headers.
func (c *Client) WithCacheHeaders(etagHeader, lastModifiedHeader string) *Client {
	c.etagHeader = etagHeader
	c.lastModifiedHeader = lastModifiedHeader
	return c
}

// WithUserAgent defines the User-Agent header to use for outgoing requests.
func (c *Client) WithUserAgent(userAgent string) *Client {
	if userAgent != "" {
		c.userAgent = userAgent
	}
	return c
}

// Get execute a GET HTTP request.
func (c *Client) Get() (*Response, error) {
	request, err := c.buildRequest(http.MethodGet, nil)
	if err != nil {
		return nil, err
	}

	return c.executeRequest(request)
}

// PostForm execute a POST HTTP request with form values.
func (c *Client) PostForm(values url.Values) (*Response, error) {
	request, err := c.buildRequest(http.MethodPost, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.executeRequest(request)
}

// PostJSON execute a POST HTTP request with JSON payload.
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
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[HttpClient] url=%s", c.url))

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
					err = errors.NewLocalizedError(errRequestTimeout, c.cfg.HTTPClientTimeout())
				} else if nerr.Temporary() {
					err = errors.NewLocalizedError(errTemporaryNetworkOperation, nerr)
				}
			}
		}

		return nil, err
	}

	if resp.ContentLength > int64(c.cfg.HTTPClientMaxSize() * 1024 * 1024) {
		return nil, fmt.Errorf("client: response too large (%d bytes)", resp.ContentLength)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: error while reading body %v", err)
	}

	response := &Response{
		Body:          bytes.NewReader(buf),
		StatusCode:    resp.StatusCode,
		EffectiveURL:  resp.Request.URL.String(),
		LastModified:  resp.Header.Get("Last-Modified"),
		ETag:          resp.Header.Get("ETag"),
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: resp.ContentLength,
	}

	logger.Debug("[HttpClient:%s] URL=%s, EffectiveURL=%s, Code=%d, Length=%d, Type=%s, ETag=%s, LastMod=%s, Expires=%s, Auth=%v",
		request.Method,
		c.url,
		response.EffectiveURL,
		response.StatusCode,
		resp.ContentLength,
		response.ContentType,
		response.ETag,
		response.LastModified,
		resp.Header.Get("Expires"),
		c.username != "",
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
	request, err := http.NewRequest(method, c.url, body)
	if err != nil {
		return nil, err
	}

	request.Header = c.buildHeaders()

	if c.username != "" && c.password != "" {
		request.SetBasicAuth(c.username, c.password)
	}

	return request, nil
}

func (c *Client) buildClient() http.Client {
	client := http.Client{Timeout: time.Duration(c.cfg.HTTPClientTimeout()) * time.Second}
	if c.Insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client
}

func (c *Client) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", c.userAgent)
	headers.Add("Accept", "*/*")

	if c.etagHeader != "" {
		headers.Add("If-None-Match", c.etagHeader)
	}

	if c.lastModifiedHeader != "" {
		headers.Add("If-Modified-Since", c.lastModifiedHeader)
	}

	if c.authorizationHeader != "" {
		headers.Add("Authorization", c.authorizationHeader)
	}

	headers.Add("Connection", "close")
	return headers
}

// New returns a new HTTP client.
func New(url string) *Client {
	return &Client{url: url, userAgent: DefaultUserAgent, Insecure: false}
}
