// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxyrotator // import "miniflux.app/v2/internal/proxyrotator"

import (
	"net/url"
	"sync"
)

var ProxyRotatorInstance *ProxyRotator

// ProxyRotator manages a list of proxies and rotates through them.
type ProxyRotator struct {
	proxies      []*url.URL
	currentIndex int
	mutex        sync.Mutex
}

// NewProxyRotator creates a new ProxyRotator with the given proxy URLs.
func NewProxyRotator(proxyURLs []string) (*ProxyRotator, error) {
	parsedProxies := make([]*url.URL, 0, len(proxyURLs))

	for _, p := range proxyURLs {
		proxyURL, err := url.Parse(p)
		if err != nil {
			return nil, err
		}
		parsedProxies = append(parsedProxies, proxyURL)
	}

	return &ProxyRotator{
		proxies:      parsedProxies,
		currentIndex: 0,
		mutex:        sync.Mutex{},
	}, nil
}

// GetNextProxy returns the next proxy in the rotation.
func (pr *ProxyRotator) GetNextProxy() *url.URL {
	if len(pr.proxies) == 0 {
		return nil
	}

	pr.mutex.Lock()
	proxy := pr.proxies[pr.currentIndex]
	pr.currentIndex = (pr.currentIndex + 1) % len(pr.proxies)
	pr.mutex.Unlock()

	return proxy
}

// HasProxies checks if there are any proxies available in the rotator.
func (pr *ProxyRotator) HasProxies() bool {
	return len(pr.proxies) > 0
}
