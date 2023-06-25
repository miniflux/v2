// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cli // import "miniflux.app/cli"

import (
	"time"

	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/model"
	"miniflux.app/storage"
)

func runCleanupTasks(store *storage.Storage) {
	nbSessions := store.CleanOldSessions(config.Opts.CleanupRemoveSessionsDays())
	nbUserSessions := store.CleanOldUserSessions(config.Opts.CleanupRemoveSessionsDays())
	logger.Info("[Sessions] Removed %d application sessions and %d user sessions", nbSessions, nbUserSessions)

	startTime := time.Now()
	if rowsAffected, err := store.ArchiveEntries(model.EntryStatusRead, config.Opts.CleanupArchiveReadDays(), config.Opts.CleanupArchiveBatchSize()); err != nil {
		logger.Error("[ArchiveReadEntries] %v", err)
	} else {
		logger.Info("[ArchiveReadEntries] %d entries changed", rowsAffected)

		if config.Opts.HasMetricsCollector() {
			metric.ArchiveEntriesDuration.WithLabelValues(model.EntryStatusRead).Observe(time.Since(startTime).Seconds())
		}
	}

	startTime = time.Now()
	if rowsAffected, err := store.ArchiveEntries(model.EntryStatusUnread, config.Opts.CleanupArchiveUnreadDays(), config.Opts.CleanupArchiveBatchSize()); err != nil {
		logger.Error("[ArchiveUnreadEntries] %v", err)
	} else {
		logger.Info("[ArchiveUnreadEntries] %d entries changed", rowsAffected)

		if config.Opts.HasMetricsCollector() {
			metric.ArchiveEntriesDuration.WithLabelValues(model.EntryStatusUnread).Observe(time.Since(startTime).Seconds())
		}
	}
}
