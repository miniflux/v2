// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
)

type trackingReadCloser struct {
	io.Reader
	closed bool
}

func (t *trackingReadCloser) Close() error {
	t.closed = true
	return nil
}

func TestZstdReadCloser(t *testing.T) {
	original := []byte("The quick brown fox jumps over the lazy dog")

	var compressed bytes.Buffer
	encoder, err := zstd.NewWriter(&compressed)
	if err != nil {
		t.Fatalf("Failed to create zstd encoder: %v", err)
	}
	if _, err := encoder.Write(original); err != nil {
		t.Fatalf("Failed to write zstd data: %v", err)
	}
	if err := encoder.Close(); err != nil {
		t.Fatalf("Failed to close zstd encoder: %v", err)
	}

	body := &trackingReadCloser{Reader: bytes.NewReader(compressed.Bytes())}
	reader := NewZstdReadCloser(body)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read zstd data: %v", err)
	}

	if !bytes.Equal(result, original) {
		t.Errorf("Expected %q, got %q", original, result)
	}
}

func TestZstdReadCloserInvalidData(t *testing.T) {
	body := &trackingReadCloser{Reader: bytes.NewReader([]byte("not zstd data"))}
	reader := NewZstdReadCloser(body)

	_, err := io.ReadAll(reader)
	if err == nil {
		t.Error("Expected error reading invalid zstd data")
	}
}

func TestZstdReadCloserClose(t *testing.T) {
	body := &trackingReadCloser{Reader: bytes.NewReader(nil)}
	reader := NewZstdReadCloser(body)
	reader.Close()

	if !body.closed {
		t.Error("Expected underlying body to be closed")
	}
}

func TestGzipReadCloser(t *testing.T) {
	original := []byte("The quick brown fox jumps over the lazy dog")

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(original); err != nil {
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	body := &trackingReadCloser{Reader: bytes.NewReader(compressed.Bytes())}
	reader := NewGzipReadCloser(body)

	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip data: %v", err)
	}

	if !bytes.Equal(result, original) {
		t.Errorf("Expected %q, got %q", original, result)
	}
}

func TestGzipReadCloserClose(t *testing.T) {
	body := &trackingReadCloser{Reader: bytes.NewReader(nil)}
	reader := NewGzipReadCloser(body)
	reader.Close()

	if !body.closed {
		t.Error("Expected underlying body to be closed")
	}
}
