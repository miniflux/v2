// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"
	"testing"
)

var trustedProxyNetworks = []string{"127.0.0.0/8", "192.168.0.0/16"}

func TestFindClientIPWithoutHeaders(t *testing.T) {
	r := &http.Request{RemoteAddr: "192.168.0.1:4242"}
	if ip := FindClientIP(r, false, nil); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "192.168.0.1"}
	if ip := FindClientIP(r, false, nil); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "fe80::14c2:f039:edc7:edc7"}
	if ip := FindClientIP(r, false, nil); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "fe80::14c2:f039:edc7:edc7%eth0"}
	if ip := FindClientIP(r, false, nil); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "[fe80::14c2:f039:edc7:edc7%eth0]:4242"}
	if ip := FindClientIP(r, false, nil); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "@"}
	if ip := FindClientIP(r, false, nil); ip != "127.0.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: ""}
	if ip := FindClientIP(r, false, nil); ip != "127.0.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestFindClientIPWithXFFHeader(t *testing.T) {
	// Test with multiple IPv4 addresses.
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "150.172.238.178" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv6 address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "2001:db8:85a3:8d3:1319:8a2e:370:7348")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "2001:db8:85a3:8d3:1319:8a2e:370:7348" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv6 address with zone
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "fe80::14c2:f039:edc7:edc7%eth0")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv4 address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "70.41.3.18")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "70.41.3.18" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with invalid IP address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "fake IP")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestFindClientIPSkipsTrustedProxiesInXFFHeader(t *testing.T) {
	scenarios := []struct {
		values   []string
		expected string
	}{
		{[]string{"203.0.113.195, 192.168.1.2"}, "203.0.113.195"},
		{[]string{"1.2.3.4, 203.0.113.195, 192.168.1.2"}, "203.0.113.195"},
		{[]string{"1.2.3.4", "203.0.113.195, 192.168.1.2"}, "203.0.113.195"},
		{[]string{"192.168.1.3, 192.168.1.2"}, "192.168.1.3"},
		{[]string{"fake IP, 192.168.1.2"}, "192.168.0.1"},
	}

	for _, scenario := range scenarios {
		headers := http.Header{}
		for _, value := range scenario.values {
			headers.Add("X-Forwarded-For", value)
		}
		r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

		if ip := FindClientIP(r, true, trustedProxyNetworks); ip != scenario.expected {
			t.Errorf(`Expected %q for X-Forwarded-For %q, got %q`, scenario.expected, scenario.values, ip)
		}
	}
}

func TestClientIPWithXRealIPHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Real-Ip", "192.168.122.1")
	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "192.168.122.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithBothHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	headers.Set("X-Real-Ip", "192.168.122.1")

	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "150.172.238.178" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithUnixSocketRemoteAddrAndBothHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	headers.Set("X-Real-Ip", "192.168.122.1")

	r := &http.Request{RemoteAddr: "@", Header: headers}

	if ip := FindClientIP(r, true, trustedProxyNetworks); ip != "150.172.238.178" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestIsTrustedIP(t *testing.T) {
	trustedNetworks := []string{"127.0.0.1/8", "10.0.0.0/8", "::1/128", "invalid"}

	scenarios := []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"::1", true},
		{"192.168.1.1", false},
		{"invalid", false},
		{"@", false},
		{"/tmp/miniflux.sock", false},
		{"", false},
	}

	for _, scenario := range scenarios {
		result := IsTrustedIP(scenario.ip, trustedNetworks)
		if result != scenario.expected {
			t.Errorf("Expected %v for IP %s, got %v", scenario.expected, scenario.ip, result)
		}
	}

	if IsTrustedIP("127.0.0.1", nil) {
		t.Error("Expected false when no trusted networks are defined")
	}

	if IsTrustedIP("127.0.0.1", []string{}) {
		t.Error("Expected false when trusted networks list is empty")
	}
}

func TestFindRemoteIP(t *testing.T) {
	scenarios := []struct {
		ip       string
		expected string
	}{
		{"192.168.0.1:4242", "192.168.0.1"},
		{"[2001:db8::1]:4242", "2001:db8::1"},
		{"fe80::14c2:f039:edc7:edc7%eth0", "fe80::14c2:f039:edc7:edc7"},
		{"", "127.0.0.1"},
		{"@", "127.0.0.1"},
		{"invalid", "127.0.0.1"},
	}

	for _, scenario := range scenarios {
		r := &http.Request{RemoteAddr: scenario.ip}
		result := FindRemoteIP(r)
		if result != scenario.expected {
			t.Errorf("Expected %q for RemoteAddr %q, got %q", scenario.expected, scenario.ip, result)
		}
	}
}
