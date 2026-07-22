// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package worker // import "miniflux.app/v2/internal/worker"

import (
	"testing"
	"time"

	"miniflux.app/v2/internal/model"
)

func TestPushAfterShutdownDiscardsJobs(t *testing.T) {
	pool := NewPool(nil, 2)
	pool.Shutdown()

	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.Push(model.JobList{{FeedID: 1}, {FeedID: 2}})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Push blocked after Shutdown instead of discarding jobs")
	}
}

func TestShutdownUnblocksPendingPush(t *testing.T) {
	pool := NewPool(nil, 0)

	pushed := make(chan struct{})
	go func() {
		defer close(pushed)
		pool.Push(model.JobList{{FeedID: 1}})
	}()

	// Give Push time to block on the unbuffered queue before shutting down.
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.Shutdown()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown deadlocked while a Push was pending")
	}

	select {
	case <-pushed:
	case <-time.After(5 * time.Second):
		t.Fatal("Push remained blocked after Shutdown")
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	pool := NewPool(nil, 1)
	pool.Shutdown()
	pool.Shutdown()
}
