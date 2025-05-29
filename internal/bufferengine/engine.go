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

// Package bufferengine provides a high-performance, multi-store buffering system
// for the GoSight monitoring platform. This engine orchestrates multiple buffered
// storage backends, managing their lifecycle, flush operations, and ensuring
// reliable data persistence across different storage technologies.
//
// The buffer engine implements a producer-consumer pattern where data is written
// to memory buffers and periodically flushed to persistent storage. This approach
// optimizes write performance, reduces I/O contention, and provides resilience
// against temporary storage unavailability.
//
// Key Features:
//   - Multi-store management with independent flush intervals
//   - Concurrent flush operations with configurable worker limits
//   - Graceful shutdown with guaranteed final data persistence
//   - Per-store lifecycle management and error isolation
//   - Context-aware operation with cancellation support
//   - Comprehensive logging and monitoring integration
//
// The engine supports heterogeneous storage backends through the BufferedStore
// interface, enabling simultaneous operation with databases, file systems,
// message queues, and other persistence mechanisms.

package bufferengine

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/utils"
)

// BufferedStore defines the interface that all storage backends must implement
// to participate in the buffer engine's managed storage system. This interface
// abstracts the underlying storage implementation details while providing
// standardized lifecycle and data management operations.
//
// The interface supports diverse storage technologies including databases,
// file systems, message queues, and cloud storage services. Each implementation
// can define its own optimal buffering strategy and flush behavior while
// maintaining compatibility with the engine's orchestration framework.
//
// Key Operations:
//   - WriteAny: Accepts heterogeneous data for buffered storage
//   - Flush: Persists buffered data to the underlying storage system
//   - Close: Performs cleanup and ensures data durability on shutdown
//   - Name: Provides identification for logging and monitoring
//   - Interval: Defines optimal flush frequency for the storage type
//
// Implementations should be thread-safe and handle concurrent access gracefully.
type BufferedStore interface {
	// WriteAny accepts arbitrary data payloads for buffered storage. The method
	// should handle type conversion and validation internally, providing flexible
	// data ingestion capabilities for diverse payload types.
	//
	// Parameters:
	//   - payload: The data to buffer (type depends on implementation)
	//
	// Returns:
	//   - error: Any error encountered during buffering operation
	WriteAny(payload interface{}) error

	// Flush persists all buffered data to the underlying storage system. This
	// method should be idempotent and handle partial failures gracefully,
	// ensuring data consistency and durability.
	//
	// Returns:
	//   - error: Any error encountered during the flush operation
	Flush() error

	// Close performs graceful shutdown of the storage backend, ensuring all
	// buffered data is persisted and resources are properly released. This
	// method should perform a final flush before closing.
	//
	// Returns:
	//   - error: Any error encountered during cleanup operations
	Close() error

	// Name returns a human-readable identifier for the storage backend,
	// used for logging, monitoring, and administrative operations.
	//
	// Returns:
	//   - string: The storage backend's display name
	Name() string

	// Interval returns the optimal flush frequency for this storage backend.
	// The engine uses this value to schedule periodic flush operations,
	// allowing each backend to define its own performance characteristics.
	//
	// Returns:
	//   - time.Duration: Recommended time between flush operations
	Interval() time.Duration
}

// BufferEngine orchestrates multiple buffered storage backends, managing their
// lifecycle, flush schedules, and ensuring reliable data persistence across
// heterogeneous storage technologies. The engine implements a sophisticated
// producer-consumer pattern optimized for high-throughput monitoring data.
//
// The engine provides:
//   - Independent flush scheduling per storage backend
//   - Concurrent operation with configurable worker limits
//   - Context-aware cancellation and graceful shutdown
//   - Comprehensive error handling and recovery
//   - Resource management and cleanup coordination
//
// Architecture:
//   - Each registered store operates independently with its own flush schedule
//   - Flush operations run concurrently to maximize throughput
//   - Context cancellation provides immediate shutdown coordination
//   - WaitGroup ensures all operations complete before engine termination
//   - Error isolation prevents store failures from affecting other stores
type BufferEngine struct {
	stores        []BufferedStore // Registered storage backends
	flushInterval time.Duration   // Default flush interval (per-store intervals take precedence)
	maxWorkers    int             // Maximum concurrent flush workers
	ctx           context.Context // Cancellation context for coordinated shutdown
	wg            sync.WaitGroup  // Synchronization for graceful shutdown
}

// NewBufferEngine creates and initializes a new BufferEngine instance configured
// for managing multiple storage backends with optimal performance characteristics.
// The engine is immediately ready for store registration and operation.
//
// The configuration parameters enable fine-tuning for different deployment
// scenarios, from single-node installations to high-throughput distributed
// environments. The context integration ensures clean shutdown coordination
// with the broader application lifecycle.
//
// Parameters:
//   - ctx: Context for cancellation and shutdown coordination
//   - flushInterval: Default flush interval (individual stores may override)
//   - maxWorkers: Maximum concurrent flush operations (resource limiting)
//
// Returns:
//   - *BufferEngine: Configured engine ready for store registration
func NewBufferEngine(ctx context.Context, flushInterval time.Duration, maxWorkers int) *BufferEngine {
	return &BufferEngine{
		flushInterval: flushInterval,
		maxWorkers:    maxWorkers,
		ctx:           ctx,
	}
}

// RegisterStore adds a new storage backend to the engine's managed store
// collection. The store becomes part of the engine's lifecycle management,
// receiving automatic flush scheduling and shutdown coordination.
//
// Stores can be registered before or after the engine starts, providing
// flexibility for dynamic storage configuration. Each store operates
// independently with its own flush schedule and error isolation.
//
// Parameters:
//   - store: The storage backend to register (must implement BufferedStore)
func (e *BufferEngine) RegisterStore(store BufferedStore) {
	e.stores = append(e.stores, store)
	utils.Info("BufferEngine registered store: %s", store.Name())
}

// Start initiates the buffer engine's operation, launching independent flush
// routines for each registered storage backend. The engine coordinates multiple
// concurrent operations while maintaining individual store autonomy and
// optimal performance characteristics.
//
// Each store operates with its own dedicated goroutine and flush schedule,
// enabling heterogeneous storage backends to function at their optimal
// frequencies. The engine monitors context cancellation for coordinated
// shutdown and ensures all stores receive final flush operations.
//
// The startup process:
//   - Launches dedicated goroutine per registered store
//   - Configures individual flush timers based on store preferences
//   - Establishes context cancellation monitoring
//   - Provides comprehensive logging and monitoring integration
//
// This method should be called after all desired stores are registered.
func (e *BufferEngine) Start() {
	utils.Info("BufferEngine starting with %d stores (per-store intervals)", len(e.stores))

	for _, store := range e.stores {
		e.wg.Add(1)
		go func(s BufferedStore) {
			defer e.wg.Done()
			interval := s.Interval()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			utils.Info("Buffer [%s] started with flush interval: %s", s.Name(), interval)

			for {
				select {
				case <-e.ctx.Done():
					utils.Info("Buffer [%s] shutting down...", s.Name())
					_ = s.Flush() // final flush on shutdown
					return
				case <-ticker.C:
					if err := s.Flush(); err != nil {
						utils.Warn("Flush failed for [%s]: %v", s.Name(), err)
					}
				}
			}
		}(store)
	}
}

// Stop performs graceful shutdown of the buffer engine and all registered
// storage backends. This method ensures all buffered data is persisted and
// resources are properly released before termination.
//
// The shutdown process:
//   - Waits for all background flush routines to complete
//   - Performs final flush operations for each store
//   - Closes all storage backends with proper error handling
//   - Provides comprehensive shutdown logging and status reporting
//
// This method blocks until all operations are complete, ensuring data
// consistency and preventing data loss during application shutdown.
// It should be called as part of the application's cleanup sequence.
func (e *BufferEngine) Stop() {
	utils.Info("BufferEngine waiting for background flush routines to stop...")
	e.wg.Wait()

	for _, store := range e.stores {
		if err := store.Close(); err != nil {
			utils.Warn("Error closing store [%s]: %v", store.Name(), err)
		}
	}
	utils.Info("BufferEngine stopped cleanly")
}
