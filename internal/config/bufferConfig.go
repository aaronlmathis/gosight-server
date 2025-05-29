/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// File: gosight-server/internal/config/bufferConfig.go
// Description: This file contains the configuration for the buffer engine.
// It includes settings for metrics, logs, data, events, and alerts buffering.

package config

import "time"

// BufferEngineConfig represents the comprehensive configuration for the buffer engine subsystem.
// The buffer engine provides high-performance data buffering and batching capabilities
// to optimize throughput and reduce load on downstream storage systems.
//
// Key features:
//   - Multi-type data buffering (metrics, logs, events, alerts, data)
//   - Configurable flush intervals and buffer sizes
//   - Graceful shutdown with timeout-controlled flushing
//   - Worker pool management for parallel processing
//   - Per-data-type configuration granularity
//
// The buffer engine operates as a write-through cache layer that:
//  1. Accepts high-frequency data ingestion
//  2. Batches data efficiently in memory
//  3. Flushes to persistent storage on intervals or size thresholds
//  4. Provides reliability through disk fallback options
//
// Performance characteristics:
//   - Reduces database write operations through batching
//   - Smooths out traffic spikes and burst writes
//   - Provides configurable trade-offs between latency and throughput
//   - Supports graceful degradation under high load
//
// Example configuration:
//
//	buffer_engine:
//	  enabled: true
//	  flush_interval: "30s"
//	  max_workers: 4
//	  metrics:
//	    enabled: true
//	    buffer_size: 10000
//	    flush_interval: "15s"
type BufferEngineConfig struct {
	Enabled              bool               `yaml:"enabled"`
	FlushInterval        time.Duration      `yaml:"flush_interval"`
	ShutdownFlushTimeout time.Duration      `yaml:"shutdown_flush_timeout"`
	MaxWorkers           int                `yaml:"max_workers"`
	Metrics              MetricBufferConfig `yaml:"metrics"`
	Logs                 LogBufferConfig    `yaml:"logs"`
	Data                 DataBufferConfig   `yaml:"data"`
	Events               EventBufferConfig  `yaml:"events"`
	Alerts               AlertBufferConfig  `yaml:"alerts"`
}

// MetricBufferConfig represents the configuration for metrics data buffering.
// Metrics are typically high-frequency, time-series data points that benefit
// significantly from batching to reduce storage system load and improve
// write throughput.
//
// Configuration options:
//   - BufferSize: Maximum number of metric entries to buffer before forced flush
//   - FlushInterval: Time-based flush trigger for ensuring data freshness
//   - DropOnOverflow: Behavior when buffer capacity is exceeded
//   - RetryFailedFlush: Retry policy for failed storage operations
//   - FlushOnDisconnect: Ensure data persistence during network issues
//   - FallbackDisk: Disk-based backup when memory buffers are full
//
// Metrics buffering strategies:
//   - Time-based flushing ensures data staleness limits
//   - Size-based flushing prevents memory exhaustion
//   - Overflow handling prevents data loss under high load
//   - Disk fallback provides durability guarantees
//
// Performance considerations:
//   - Larger buffers improve batching efficiency but increase memory usage
//   - Shorter intervals improve data freshness but reduce batching benefits
//   - Disk fallback adds durability at the cost of I/O overhead
type MetricBufferConfig struct {
	Enabled           bool             `yaml:"enabled"`
	BufferSize        int              `yaml:"buffer_size"`
	FlushInterval     time.Duration    `yaml:"flush_interval"`
	DropOnOverflow    bool             `yaml:"drop_on_overflow"`
	RetryFailedFlush  bool             `yaml:"retry_failed_flush"`
	FlushOnDisconnect bool             `yaml:"flush_on_disconnect"`
	FallbackDisk      DiskBufferConfig `yaml:"fallback_disk"`
}

// LogBufferConfig represents the configuration for log message buffering.
// Log buffering optimizes the ingestion and storage of log messages from
// various sources including applications, system logs, and syslog feeds.
//
// Log-specific considerations:
//   - Variable message sizes require flexible buffer management
//   - Log bursts during incidents need overflow protection
//   - Critical logs may require immediate flushing
//   - Disk fallback ensures no log message loss
//
// Configuration features:
//   - BufferSize: Maximum number of log entries to buffer
//   - FlushInterval: Maximum time logs remain in buffer
//   - DropOnOverflow: Policy for handling buffer overflow
//   - RetryFailedFlush: Retry mechanism for storage failures
//   - FallbackDisk: Disk-based overflow protection
//
// Buffering benefits for logs:
//   - Reduces storage I/O operations during log bursts
//   - Enables efficient batch processing of log data
//   - Provides smoothing for irregular log patterns
//   - Maintains log ordering within flush intervals
type LogBufferConfig struct {
	Enabled          bool             `yaml:"enabled"`
	BufferSize       int              `yaml:"buffer_size"`
	FlushInterval    time.Duration    `yaml:"flush_interval"`
	DropOnOverflow   bool             `yaml:"drop_on_overflow"`
	RetryFailedFlush bool             `yaml:"retry_failed_flush"`
	FallbackDisk     DiskBufferConfig `yaml:"fallback_disk"`
}

// DataBufferConfig represents the configuration for general data buffering.
// This buffer handles miscellaneous data types that don't fit into specific
// categories like metrics or logs, providing a flexible buffering solution
// for custom data ingestion workflows.
//
// Use cases:
//   - Custom telemetry data from applications
//   - Third-party integration data feeds
//   - Batch processing of API responses
//   - Temporary data aggregation before analysis
//
// Configuration capabilities:
//   - BufferSize: Maximum entries before forced flush
//   - FlushInterval: Time-based flush frequency
//   - DropOnOverflow: Overflow handling strategy
//   - RetryFailedFlush: Error recovery mechanism
//   - FlushOnDisconnect: Network resilience feature
//   - FallbackDisk: Persistent overflow storage
//
// The data buffer provides generic buffering capabilities while
// maintaining the same reliability and performance characteristics
// as specialized buffers for metrics and logs.
type DataBufferConfig struct {
	Enabled           bool             `yaml:"enabled"`
	BufferSize        int              `yaml:"buffer_size"`
	FlushInterval     time.Duration    `yaml:"flush_interval"`
	DropOnOverflow    bool             `yaml:"drop_on_overflow"`
	RetryFailedFlush  bool             `yaml:"retry_failed_flush"`
	FlushOnDisconnect bool             `yaml:"flush_on_disconnect"`
	FallbackDisk      DiskBufferConfig `yaml:"fallback_disk"`
}

// EventBufferConfig represents the configuration for event data buffering.
// Events represent discrete occurrences or state changes in the system,
// often requiring different buffering strategies than continuous metrics
// or log streams.
//
// Event characteristics:
//   - Typically lower frequency than metrics but higher importance
//   - May require ordered processing for correlation
//   - Often trigger downstream processing workflows
//   - Critical events may need immediate processing
//
// Configuration options:
//   - BufferSize: Maximum number of events to buffer
//   - FlushInterval: Maximum event retention time in buffer
//   - RetryFailedFlush: Retry policy for storage failures
//
// Event buffering benefits:
//   - Enables batch processing of related events
//   - Reduces storage transaction overhead
//   - Provides event ordering within flush windows
//   - Smooths irregular event generation patterns
//
// Note: Events typically don't use overflow dropping or disk fallback
// to ensure critical event data is never lost.
type EventBufferConfig struct {
	Enabled          bool          `yaml:"enabled"`
	BufferSize       int           `yaml:"buffer_size"`
	FlushInterval    time.Duration `yaml:"flush_interval"`
	RetryFailedFlush bool          `yaml:"retry_failed_flush"`
}

// AlertBufferConfig represents the configuration for alert buffering.
// Alerts are critical notifications that require careful handling to ensure
// timely delivery while avoiding notification storms and duplicate alerts.
//
// Alert-specific requirements:
//   - Time-sensitive delivery for critical incidents
//   - Deduplication to prevent alert storms
//   - Ordering preservation for alert correlation
//   - Minimal latency for high-severity alerts
//
// Configuration features:
//   - BufferSize: Maximum alerts to buffer before flush
//   - FlushInterval: Maximum alert delay tolerance
//   - DropOnOverflow: Overflow handling (typically disabled for alerts)
//   - RetryFailedFlush: Retry mechanism for delivery failures
//
// Alert buffering strategies:
//   - Short flush intervals to minimize notification delays
//   - Conservative overflow policies to prevent alert loss
//   - Retry mechanisms to ensure delivery reliability
//   - Size-based flushing to handle alert bursts
//
// Performance vs. reliability trade-offs:
//   - Smaller buffers reduce latency but increase overhead
//   - Retry logic ensures delivery but may increase delay
//   - No disk fallback to maintain alert delivery speed
type AlertBufferConfig struct {
	Enabled          bool          `yaml:"enabled"`
	BufferSize       int           `yaml:"buffer_size"`
	FlushInterval    time.Duration `yaml:"flush_interval"`
	DropOnOverflow   bool          `yaml:"drop_on_overflow"`
	RetryFailedFlush bool          `yaml:"retry_failed_flush"`
}

// DiskBufferConfig represents the configuration for disk-based buffer fallback.
// When memory buffers reach capacity or during system stress, the disk buffer
// provides persistent storage to prevent data loss while maintaining system
// stability and performance.
//
// Disk buffering features:
//   - Persistent storage for overflow data
//   - Configurable size limits to prevent disk exhaustion
//   - Automatic cleanup when memory buffers have capacity
//   - Sequential write optimization for performance
//
// Configuration options:
//   - Enabled: Controls whether disk fallback is active
//   - Path: Directory path for storing buffered data files
//   - MaxDiskSizeMB: Maximum disk space allocation in megabytes
//
// Operational behavior:
//   - Activates when memory buffers reach capacity
//   - Uses sequential file writes for optimal I/O performance
//   - Automatically rotates files when size limits are reached
//   - Provides data recovery during system restarts
//
// Performance considerations:
//   - Disk I/O is slower than memory but provides durability
//   - Sequential writes minimize disk seek time
//   - Size limits prevent disk space exhaustion
//   - File rotation maintains manageable file sizes
//
// Example configuration:
//
//	fallback_disk:
//	  enabled: true
//	  path: "/var/lib/gosight/buffer"
//	  max_disk_size_mb: 1024
type DiskBufferConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Path          string `yaml:"path"`
	MaxDiskSizeMB int    `yaml:"max_disk_size_mb"`
}
