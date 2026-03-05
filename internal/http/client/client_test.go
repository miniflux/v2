// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientWithoutBlockingPrivateNetworks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithOptions(Options{Timeout: 5 * time.Second})
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestBlockPrivateNetworksBlocksLoopback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithOptions(Options{Timeout: 5 * time.Second, BlockPrivateNetworks: true})
	_, err := client.Get(server.URL)
	if err == nil {
		t.Fatal("Expected an error when connecting to loopback address, got nil")
	}

	if !errors.Is(err, ErrPrivateNetwork) {
		t.Fatalf("Expected ErrPrivateNetwork, got %v", err)
	}
}

func TestBlockPrivateNetworksAllowsPublicIPs(t *testing.T) {
	client := NewClientWithOptions(Options{Timeout: 5 * time.Second, BlockPrivateNetworks: true})
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected custom http.Transport when blockPrivateNetworks is true")
	}
	if transport.DialContext == nil {
		t.Fatal("Expected custom DialContext when blockPrivateNetworks is true")
	}
}

func TestNoCustomTransportWhenNotBlocking(t *testing.T) {
	client := NewClientWithOptions(Options{Timeout: 5 * time.Second})
	if client.Transport != nil {
		t.Fatal("Expected nil transport when blockPrivateNetworks is false")
	}
}

func TestBlockPrivateNetworksBlocksPrivateIP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Listener = listener
	server.Start()
	defer server.Close()

	client := NewClientWithOptions(Options{Timeout: 5 * time.Second, BlockPrivateNetworks: true})
	_, err = client.Get(server.URL)
	if err == nil {
		t.Fatal("Expected error when connecting to private IP")
	}

	if !errors.Is(err, ErrPrivateNetwork) {
		t.Fatalf("Expected ErrPrivateNetwork, got: %v", err)
	}
}

func TestBlockPrivateNetworksAllowsLoopbackWhenDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClientWithOptions(Options{Timeout: 5 * time.Second})
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Expected no error when blockPrivateNetworks is false, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
}
