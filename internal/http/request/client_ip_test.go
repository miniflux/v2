// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net/http"
	"testing"
)

func TestFindClientIPWithoutHeaders(t *testing.T) {
	r := &http.Request{RemoteAddr: "192.168.0.1:4242"}
	if ip := FindClientIP(r); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "192.168.0.1"}
	if ip := FindClientIP(r); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "fe80::14c2:f039:edc7:edc7"}
	if ip := FindClientIP(r); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "fe80::14c2:f039:edc7:edc7%eth0"}
	if ip := FindClientIP(r); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	r = &http.Request{RemoteAddr: "[fe80::14c2:f039:edc7:edc7%eth0]:4242"}
	if ip := FindClientIP(r); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestFindClientIPWithXFFHeader(t *testing.T) {
	// Test with multiple IPv4 addresses.
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "203.0.113.195" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv6 address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "2001:db8:85a3:8d3:1319:8a2e:370:7348")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "2001:db8:85a3:8d3:1319:8a2e:370:7348" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv6 address with zone
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "fe80::14c2:f039:edc7:edc7%eth0")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "fe80::14c2:f039:edc7:edc7" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with single IPv4 address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "70.41.3.18")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "70.41.3.18" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}

	// Test with invalid IP address.
	headers = http.Header{}
	headers.Set("X-Forwarded-For", "fake IP")
	r = &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "192.168.0.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithXRealIPHeader(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Real-Ip", "192.168.122.1")
	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "192.168.122.1" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithBothHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	headers.Set("X-Real-Ip", "192.168.122.1")

	r := &http.Request{RemoteAddr: "192.168.0.1:4242", Header: headers}

	if ip := FindClientIP(r); ip != "203.0.113.195" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithUnixSocketRemoteAddress(t *testing.T) {
	r := &http.Request{RemoteAddr: "@"}

	if ip := FindClientIP(r); ip != "@" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}

func TestClientIPWithUnixSocketRemoteAddrAndBothHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
	headers.Set("X-Real-Ip", "192.168.122.1")

	r := &http.Request{RemoteAddr: "@", Header: headers}

	if ip := FindClientIP(r); ip != "203.0.113.195" {
		t.Fatalf(`Unexpected result, got: %q`, ip)
	}
}
