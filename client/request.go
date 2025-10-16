// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client // import "miniflux.app/v2/client"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	userAgent      = "Miniflux Client Library"
	defaultTimeout = 80 * time.Second
)

// List of exposed errors.
var (
	ErrNotAuthorized = errors.New("miniflux: unauthorized (bad credentials)")
	ErrForbidden     = errors.New("miniflux: access forbidden")
	ErrServerError   = errors.New("miniflux: internal server error")
	ErrNotFound      = errors.New("miniflux: resource not found")
	ErrBadRequest    = errors.New("miniflux: bad request")
	ErrEmptyEndpoint = errors.New("miniflux: empty endpoint provided")
)

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

type request struct {
	endpoint string
	username string
	password string
	apiKey   string
	client   *http.Client
}

func (r *request) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return r.execute(ctx, http.MethodGet, path, nil)
}

func (r *request) Post(ctx context.Context, path string, data any) (io.ReadCloser, error) {
	return r.execute(ctx, http.MethodPost, path, data)
}

func (r *request) PostFile(ctx context.Context, path string, f io.ReadCloser) (io.ReadCloser, error) {
	return r.execute(ctx, http.MethodPost, path, f)
}

func (r *request) Put(ctx context.Context, path string, data any) (io.ReadCloser, error) {
	return r.execute(ctx, http.MethodPut, path, data)
}

func (r *request) Delete(ctx context.Context, path string) error {
	_, err := r.execute(ctx, http.MethodDelete, path, nil)
	return err
}

func (r *request) execute(
	ctx context.Context,
	method string,
	path string,
	data any,
) (io.ReadCloser, error) {
	if r.endpoint == "" {
		return nil, ErrEmptyEndpoint
	}
	if r.endpoint[len(r.endpoint)-1:] == "/" {
		r.endpoint = r.endpoint[:len(r.endpoint)-1]
	}

	u, err := url.Parse(r.endpoint + path)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header = r.buildHeaders()

	if r.username != "" && r.password != "" {
		request.SetBasicAuth(r.username, r.password)
	}

	if data != nil {
		switch data := data.(type) {
		case io.ReadCloser:
			request.Body = data
		default:
			request.Body = io.NopCloser(bytes.NewBuffer(r.toJSON(data)))
		}
	}

	client := r.client
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusUnauthorized:
		response.Body.Close()
		return nil, ErrNotAuthorized
	case http.StatusForbidden:
		response.Body.Close()
		return nil, ErrForbidden
	case http.StatusInternalServerError:
		defer response.Body.Close()

		var resp errorResponse
		decoder := json.NewDecoder(response.Body)
		// If we failed to decode, just return a generic ErrServerError
		if err := decoder.Decode(&resp); err != nil {
			return nil, ErrServerError
		}
		return nil, errors.New("miniflux: internal server error: " + resp.ErrorMessage)
	case http.StatusNotFound:
		response.Body.Close()
		return nil, ErrNotFound
	case http.StatusNoContent:
		response.Body.Close()
		return nil, nil
	case http.StatusBadRequest:
		defer response.Body.Close()

		var resp errorResponse
		decoder := json.NewDecoder(response.Body)
		if err := decoder.Decode(&resp); err != nil {
			return nil, fmt.Errorf("%w (%v)", ErrBadRequest, err)
		}

		return nil, fmt.Errorf("%w (%s)", ErrBadRequest, resp.ErrorMessage)
	}

	if response.StatusCode > 400 {
		response.Body.Close()
		return nil, fmt.Errorf("miniflux: status code=%d", response.StatusCode)
	}

	return response.Body, nil
}

func (r *request) buildHeaders() http.Header {
	headers := make(http.Header)
	headers.Add("User-Agent", userAgent)
	headers.Add("Content-Type", "application/json")
	headers.Add("Accept", "application/json")
	if r.apiKey != "" {
		headers.Add("X-Auth-Token", r.apiKey)
	}
	return headers
}

func (r *request) toJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("Unable to convert interface to JSON:", err)
		return []byte("")
	}

	return b
}
