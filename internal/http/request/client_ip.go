// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package request // import "miniflux.app/v2/internal/http/request"

import (
	"net"
	"net/http"
	"strings"
)

// IsTrustedIP reports whether the given remote IP address belongs to one of the trusted networks.
func IsTrustedIP(remoteIP string, trustedNetworks []string) bool {
	if len(trustedNetworks) == 0 {
		return false
	}

	ip := net.ParseIP(remoteIP)
	if ip == nil {
		return false
	}

	for _, cidr := range trustedNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}

		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// FindClientIP returns the real client IP address using trusted reverse-proxy headers when allowed.
func FindClientIP(r *http.Request, isTrustedProxyClient bool, trustedNetworks []string) string {
	if isTrustedProxyClient {
		headers := [...]string{"X-Forwarded-For", "X-Real-Ip"}
		for _, header := range headers {
			value := strings.Join(r.Header.Values(header), ",")

			if value != "" {
				// Proxies append to the list, so skip trusted hops from the right.
				addresses := strings.Split(value, ",")
				i := len(addresses) - 1
				for i > 0 && IsTrustedIP(dropIPv6zone(strings.TrimSpace(addresses[i])), trustedNetworks) {
					i--
				}
				address := strings.TrimSpace(addresses[i])
				address = dropIPv6zone(address)

				if net.ParseIP(address) != nil {
					return address
				}
			}
		}
	}

	// Fallback to TCP/IP source IP address.
	return FindRemoteIP(r)
}

// FindRemoteIP returns the parsed remote IP address from the request,
// falling back to 127.0.0.1 if the address is empty, a unix socket, or invalid.
func FindRemoteIP(r *http.Request) string {
	if r.RemoteAddr == "@" || r.RemoteAddr == "" {
		return "127.0.0.1"
	}

	// If it looks like it has a port (IPv4:port or [IPv6]:port), try to split it.
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// No port — could be a bare IPv4, IPv6, or IPv6 with zone.
		ip = r.RemoteAddr
	}

	// Strip IPv6 zone identifier if present (e.g., %eth0).
	ip = dropIPv6zone(ip)

	// Validate the IP address.
	if net.ParseIP(ip) == nil {
		return "127.0.0.1"
	}

	return ip
}

func dropIPv6zone(address string) string {
	idx := strings.IndexByte(address, '%')
	if idx != -1 {
		address = address[:idx]
	}
	return address
}
