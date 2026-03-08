// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration // import "miniflux.app/v2/internal/integration"

import (
	"log/slog"
	"time"

	"miniflux.app/v2/internal/integration/ai"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
)

const (
	// aiWorkerScanInterval is the pause between full user scans.
	aiWorkerScanInterval = 1 * time.Minute

	// aiWorkerBatchSize is the number of entries processed per user per cycle.
	// Smaller than backfill's 50 to spread API load evenly over time.
	aiWorkerBatchSize = 10
)

// StartAISummaryWorker starts a background goroutine that continuously scans
// for unread entries without AI summaries and generates them.
// It processes all users with AI enabled, in batches of 10 entries per user per cycle.
// The worker stops when the stop channel is closed.
func StartAISummaryWorker(store *storage.Storage, stop <-chan struct{}) {
	slog.Info("AI summary worker started")

	for {
		usersProcessed, entriesSummarized := runAISummaryCycle(store)

		if entriesSummarized > 0 {
			slog.Info("AI summary worker cycle completed",
				slog.Int("users_processed", usersProcessed),
				slog.Int("entries_summarized", entriesSummarized),
			)
		}

		select {
		case <-stop:
			slog.Info("AI summary worker stopped")
			return
		case <-time.After(aiWorkerScanInterval):
		}
	}
}

// runAISummaryCycle iterates over all users, generates AI summaries for unsummarized
// unread entries, and returns how many users were processed and entries summarized.
func runAISummaryCycle(store *storage.Storage) (int, int) {
	users, err := store.Users()
	if err != nil {
		slog.Error("AI summary worker: failed to fetch users",
			slog.Any("error", err),
		)
		return 0, 0
	}

	usersProcessed := 0
	totalEntriesSummarized := 0

	for _, user := range users {
		integration, err := store.Integration(user.ID)
		if err != nil {
			slog.Warn("AI summary worker: failed to load integration config",
				slog.Int64("user_id", user.ID),
				slog.Any("error", err),
			)
			continue
		}

		if !integration.AIEnabled || integration.AIProviderURL == "" || integration.AIAPIKey == "" || integration.AIModel == "" {
			continue
		}

		summarized := processUserEntries(store, user, integration)
		if summarized > 0 {
			totalEntriesSummarized += summarized
		}
		usersProcessed++
	}

	return usersProcessed, totalEntriesSummarized
}

// processUserEntries queries and summarizes unsummarized unread entries for a single user.
// Returns the number of entries successfully summarized.
func processUserEntries(store *storage.Storage, user *model.User, integration *model.Integration) int {
	client := ai.NewClient(
		integration.AIProviderURL,
		integration.AIAPIKey,
		integration.AIModel,
	)

	entryBuilder := store.NewEntryQueryBuilder(user.ID)
	entryBuilder.WithStatus(model.EntryStatusUnread)
	entryBuilder.WithoutAISummary()
	entryBuilder.WithSorting("published_at", "desc")
	entryBuilder.WithLimit(aiWorkerBatchSize)
	entries, err := entryBuilder.GetEntries()
	if err != nil {
		slog.Warn("AI summary worker: failed to query entries",
			slog.Int64("user_id", user.ID),
			slog.Any("error", err),
		)
		return 0
	}

	summarized := 0
	for _, entry := range entries {
		result, err := client.SummarizeEntry(entry.Title, entry.Content, entry.AISummary, user.Language)
		if err != nil {
			slog.Warn("AI summary worker: failed to summarize entry",
				slog.Int64("user_id", user.ID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
				slog.Any("error", err),
			)
			continue
		}

		if result == nil {
			continue
		}

		now := time.Now()
		entry.AISummary = result.Summary
		entry.AIScore = result.Score
		entry.AISummarizedAt = &now

		if err := store.UpdateEntryAISummary(entry); err != nil {
			slog.Warn("AI summary worker: failed to save summary",
				slog.Int64("user_id", user.ID),
				slog.Int64("entry_id", entry.ID),
				slog.Any("error", err),
			)
			continue
		}

		summarized++
	}

	return summarized
}
