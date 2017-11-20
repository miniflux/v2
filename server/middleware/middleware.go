// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

type MiddlewareChain struct {
	middlewares []Middleware
}

func (m *MiddlewareChain) Wrap(h http.Handler) http.Handler {
	for i := range m.middlewares {
		h = m.middlewares[len(m.middlewares)-1-i](h)
	}

	return h
}

func (m *MiddlewareChain) WrapFunc(fn http.HandlerFunc) http.Handler {
	return m.Wrap(fn)
}

func NewMiddlewareChain(middlewares ...Middleware) *MiddlewareChain {
	return &MiddlewareChain{append(([]Middleware)(nil), middlewares...)}
}
