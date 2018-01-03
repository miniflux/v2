// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
)

// Middleware represents a HTTP middleware.
type Middleware func(http.Handler) http.Handler

// Chain handles a list of middlewares.
type Chain struct {
	middlewares []Middleware
}

// Wrap adds a HTTP handler into the chain.
func (m *Chain) Wrap(h http.Handler) http.Handler {
	for i := range m.middlewares {
		h = m.middlewares[len(m.middlewares)-1-i](h)
	}

	return h
}

// WrapFunc adds a HTTP handler function into the chain.
func (m *Chain) WrapFunc(fn http.HandlerFunc) http.Handler {
	return m.Wrap(fn)
}

// NewChain returns a new Chain.
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{append(([]Middleware)(nil), middlewares...)}
}
