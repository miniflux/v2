// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package metric // import "miniflux.app/metric"

import (
	"time"

	"miniflux.app/logger"
	"miniflux.app/storage"

	"github.com/prometheus/client_golang/prometheus"
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
)

// Collector represents a metric collector.
type Collector struct {
	store           *storage.Storage
	refreshInterval int
}

// NewCollector initializes a new metric collector.
func NewCollector(store *storage.Storage, refreshInterval int) *Collector {
	prometheus.MustRegister(BackgroundFeedRefreshDuration)
	prometheus.MustRegister(ScraperRequestDuration)
	prometheus.MustRegister(ArchiveEntriesDuration)
	prometheus.MustRegister(usersGauge)
	prometheus.MustRegister(feedsGauge)
	prometheus.MustRegister(brokenFeedsGauge)
	prometheus.MustRegister(entriesGauge)

	return &Collector{store, refreshInterval}
}

// GatherStorageMetrics polls the database to fetch metrics.
func (c *Collector) GatherStorageMetrics() {
	for range time.Tick(time.Duration(c.refreshInterval) * time.Second) {
		logger.Debug("[Metric] Collecting database metrics")

		usersGauge.Set(float64(c.store.CountUsers()))
		brokenFeedsGauge.Set(float64(c.store.CountAllFeedsWithErrors()))

		feedsCount := c.store.CountAllFeeds()
		for status, count := range feedsCount {
			feedsGauge.WithLabelValues(status).Set(float64(count))
		}

		entriesCount := c.store.CountAllEntries()
		for status, count := range entriesCount {
			entriesGauge.WithLabelValues(status).Set(float64(count))
		}
	}
}
