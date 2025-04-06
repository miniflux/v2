// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxyrotator // import "miniflux.app/v2/internal/proxyrotator"

import (
	"testing"
)

func TestProxyRotator(t *testing.T) {
	proxyURLs := []string{
		"http://proxy1.example.com",
		"http://proxy2.example.com",
		"http://proxy3.example.com",
	}

	rotator, err := NewProxyRotator(proxyURLs)
	if err != nil {
		t.Fatalf("Failed to create ProxyRotator: %v", err)
	}

	if !rotator.HasProxies() {
		t.Fatalf("Expected rotator to have proxies")
	}

	seenProxies := make(map[string]bool)
	for range len(proxyURLs) * 2 {
		proxy := rotator.GetNextProxy()
		if proxy == nil {
			t.Fatalf("Expected a proxy, got nil")
		}

		seenProxies[proxy.String()] = true
	}

	if len(seenProxies) != len(proxyURLs) {
		t.Fatalf("Expected to see all proxies, but saw: %v", seenProxies)
	}
}

func TestProxyRotatorEmpty(t *testing.T) {
	rotator, err := NewProxyRotator([]string{})
	if err != nil {
		t.Fatalf("Failed to create ProxyRotator: %v", err)
	}

	if rotator.HasProxies() {
		t.Fatalf("Expected rotator to have no proxies")
	}

	proxy := rotator.GetNextProxy()
	if proxy != nil {
		t.Fatalf("Expected no proxy, got: %v", proxy)
	}
}

func TestProxyRotatorInvalidURL(t *testing.T) {
	invalidProxyURLs := []string{
		"http://validproxy.example.com",
		"test|test://invalidproxy.example.com",
	}

	rotator, err := NewProxyRotator(invalidProxyURLs)
	if err == nil {
		t.Fatalf("Expected an error when creating ProxyRotator with invalid URLs, but got none")
	}

	if rotator != nil {
		t.Fatalf("Expected rotator to be nil when initialization fails, but got: %v", rotator)
	}
}
