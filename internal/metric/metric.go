// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metric // import "miniflux.app/v2/internal/metric"

import (
	"context"
	"log/slog"
	"time"

	"miniflux.app/v2/internal/storage"

	"github.com/prometheus/client_golang/prometheus"
)

// Status label values for histogram metrics.
const (
	StatusSuccess = "success"
	StatusError   = "error"
)

// Prometheus Metrics.
var (
	BackgroundFeedRefreshDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "miniflux",
			Name:      "background_feed_refresh_duration",
			Help:      "Processing time to refresh feeds from the background workers",
			Buckets:   prometheus.LinearBuckets(1, 2, 15),
		},
		[]string{"status"},
	)

	ScraperRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "miniflux",
			Name:      "scraper_request_duration",
			Help:      "Web scraper request duration",
			Buckets:   prometheus.LinearBuckets(1, 2, 25),
		},
		[]string{"status"},
	)

	ArchiveEntriesDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "miniflux",
			Name:      "archive_entries_duration",
			Help:      "Archive entries duration",
			Buckets:   prometheus.LinearBuckets(1, 2, 30),
		},
		[]string{"status"},
	)

	usersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "users",
			Help:      "Number of users",
		},
	)

	feedsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "feeds",
			Help:      "Number of feeds by status",
		},
		[]string{"status"},
	)

	brokenFeedsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "broken_feeds",
			Help:      "Number of broken feeds",
		},
	)

	entriesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "entries",
			Help:      "Number of entries by status",
		},
		[]string{"status"},
	)

	dbOpenConnectionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_open_connections",
			Help:      "The number of established connections both in use and idle",
		},
	)

	dbConnectionsInUseGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_in_use",
			Help:      "The number of connections currently in use",
		},
	)

	dbConnectionsIdleGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_idle",
			Help:      "The number of idle connections",
		},
	)

	dbConnectionsWaitCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_wait_count",
			Help:      "The total number of connections waited for",
		},
	)

	dbConnectionsMaxIdleClosedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_max_idle_closed",
			Help:      "The total number of connections closed due to SetMaxIdleConns",
		},
	)

	dbConnectionsMaxIdleTimeClosedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_max_idle_time_closed",
			Help:      "The total number of connections closed due to SetConnMaxIdleTime",
		},
	)

	dbConnectionsMaxLifetimeClosedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "miniflux",
			Name:      "db_connections_max_lifetime_closed",
			Help:      "The total number of connections closed due to SetConnMaxLifetime",
		},
	)
)

// collector represents a metric collector.
type collector struct {
	store           *storage.Storage
	refreshInterval time.Duration
}

// NewCollector initializes a new metric collector.
func NewCollector(store *storage.Storage, refreshInterval time.Duration) *collector {
	prometheus.MustRegister(BackgroundFeedRefreshDuration)
	prometheus.MustRegister(ScraperRequestDuration)
	prometheus.MustRegister(ArchiveEntriesDuration)
	prometheus.MustRegister(usersGauge)
	prometheus.MustRegister(feedsGauge)
	prometheus.MustRegister(brokenFeedsGauge)
	prometheus.MustRegister(entriesGauge)
	prometheus.MustRegister(dbOpenConnectionsGauge)
	prometheus.MustRegister(dbConnectionsInUseGauge)
	prometheus.MustRegister(dbConnectionsIdleGauge)
	prometheus.MustRegister(dbConnectionsWaitCountGauge)
	prometheus.MustRegister(dbConnectionsMaxIdleClosedGauge)
	prometheus.MustRegister(dbConnectionsMaxIdleTimeClosedGauge)
	prometheus.MustRegister(dbConnectionsMaxLifetimeClosedGauge)

	return &collector{store, refreshInterval}
}

// GatherStorageMetrics polls the database to fetch metrics.
func (c *collector) GatherStorageMetrics(ctx context.Context) {
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Debug("Stopping metric collector")
			return
		case <-ticker.C:
		}
		slog.Debug("Collecting metrics from the database")

		if usersCount, err := c.store.CountUsers(); err != nil {
			slog.Warn("Unable to collect users metric", slog.Any("error", err))
		} else {
			usersGauge.Set(float64(usersCount))
		}

		if brokenFeedsCount, err := c.store.CountAllFeedsWithErrors(); err != nil {
			slog.Warn("Unable to collect broken feeds metric", slog.Any("error", err))
		} else {
			brokenFeedsGauge.Set(float64(brokenFeedsCount))
		}

		if feedsCount, err := c.store.CountAllFeeds(); err != nil {
			slog.Warn("Unable to collect feeds metric", slog.Any("error", err))
		} else {
			for status, count := range feedsCount {
				feedsGauge.WithLabelValues(status).Set(float64(count))
			}
		}

		if entriesCount, err := c.store.CountAllEntries(); err != nil {
			slog.Warn("Unable to collect entries metric", slog.Any("error", err))
		} else {
			for status, count := range entriesCount {
				entriesGauge.WithLabelValues(status).Set(float64(count))
			}
		}

		dbStats := c.store.DBStats()
		dbOpenConnectionsGauge.Set(float64(dbStats.OpenConnections))
		dbConnectionsInUseGauge.Set(float64(dbStats.InUse))
		dbConnectionsIdleGauge.Set(float64(dbStats.Idle))
		dbConnectionsWaitCountGauge.Set(float64(dbStats.WaitCount))
		dbConnectionsMaxIdleClosedGauge.Set(float64(dbStats.MaxIdleClosed))
		dbConnectionsMaxIdleTimeClosedGauge.Set(float64(dbStats.MaxIdleTimeClosed))
		dbConnectionsMaxLifetimeClosedGauge.Set(float64(dbStats.MaxLifetimeClosed))
	}
}
