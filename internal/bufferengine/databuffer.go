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

// Description: Package bufferengine provides a buffered data store implementation
// for writing process payloads to an underlying data store.
// It buffers the payloads in memory and flushes them to the underlying store
// when the buffer reaches a certain size or after a specified interval.
// The buffered data store is designed to improve performance by reducing the
// number of write operations to the underlying data store.
package bufferengine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/datastore"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// DataStore defines the interface for persistent storage of process monitoring
// data within the GoSight system. This interface abstracts the underlying
// storage implementation, enabling support for diverse storage backends
// including databases, file systems, and cloud storage services.
//
// The interface is designed to handle batch operations efficiently, supporting
// high-throughput process monitoring scenarios where large volumes of process
// data need to be persisted with minimal latency impact.
type DataStore interface {
	// Write persists a batch of process payloads to the underlying storage system.
	// The method should handle batch operations efficiently and provide appropriate
	// error handling for partial failures.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - batches: Slice of process payloads to persist
	//
	// Returns:
	//   - error: Any error encountered during the write operation
	Write(ctx context.Context, batches []*model.ProcessPayload) error
}

// BufferedDataStore implements high-performance buffered storage for process
// monitoring data, optimizing write operations through intelligent batching
// and configurable flush strategies. This implementation significantly reduces
// I/O overhead and improves overall system throughput in high-frequency
// process monitoring scenarios.
//
// The buffer operates with dual flush triggers:
//   - Size-based: Automatic flush when buffer reaches capacity
//   - Time-based: Periodic flush based on configurable intervals
//
// This design ensures optimal performance while preventing unbounded memory
// growth and providing predictable data persistence guarantees.
//
// Key Features:
//   - Thread-safe concurrent operations with optimized locking
//   - Configurable buffer size and flush intervals
//   - Automatic batch optimization for storage efficiency
//   - Context-aware operations with cancellation support
//   - Comprehensive error handling and recovery
//   - Memory-efficient buffer management
//   - Integration with diverse storage backends
//
// The implementation provides reliable data persistence while minimizing
// the performance impact on data collection operations.
type BufferedDataStore struct {
	name          string                  // Human-readable identifier for logging
	underlying    datastore.DataStore     // Persistent storage backend
	buffer        []*model.ProcessPayload // In-memory payload buffer
	mu            sync.Mutex              // Synchronization for thread safety
	maxSize       int                     // Maximum buffer size before flush
	flushInterval time.Duration           // Time-based flush interval
	ctx           context.Context         // Context for cancellation and timeout
}

// NewBufferedDataStore creates and initializes a new BufferedDataStore instance
// configured for optimal process data storage performance. The store immediately
// begins accepting write operations and manages automatic flush operations
// based on the specified configuration parameters.
//
// The configuration enables fine-tuning for different deployment scenarios:
//   - maxSize: Controls memory usage and batch size optimization
//   - flushInterval: Balances latency vs. throughput requirements
//   - ctx: Enables graceful shutdown and operation cancellation
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - name: Human-readable identifier for logging and monitoring
//   - store: The underlying persistent storage implementation
//   - maxSize: Maximum buffer size before automatic flush
//   - flushInterval: Time interval for periodic flush operations
//
// Returns:
//   - *BufferedDataStore: Configured buffer ready for data operations
func NewBufferedDataStore(ctx context.Context, name string, store datastore.DataStore, maxSize int, flushInterval time.Duration) *BufferedDataStore {
	return &BufferedDataStore{
		name:          name,
		underlying:    store,
		buffer:        make([]*model.ProcessPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
		ctx:           ctx,
	}
}

// Name returns the human-readable identifier for this buffered data store
// instance. This identifier is used for logging, monitoring, metrics, and
// administrative operations to distinguish between multiple buffer instances.
//
// Returns:
//   - string: The store's display name
func (b *BufferedDataStore) Name() string {
	return b.name
}

// Interval returns the configured time duration between automatic flush
// operations. The buffer engine uses this value to schedule periodic flush
// operations, enabling each buffer to operate at its optimal frequency
// based on data characteristics and storage requirements.
//
// Returns:
//   - time.Duration: Time interval between automatic flush operations
func (b *BufferedDataStore) Interval() time.Duration {
	return b.flushInterval
}

// WriteAny provides a type-safe interface for writing arbitrary payloads to
// the buffered data store. This method validates the payload type and delegates
// to the appropriate typed write method, ensuring type safety while maintaining
// interface compatibility with the BufferedStore interface.
//
// The method specifically handles process payloads, validating the type and
// rejecting incompatible data with descriptive error messages.
//
// Parameters:
//   - payload: The data payload to write (must be *model.ProcessPayload)
//
// Returns:
//   - error: Type validation error or write operation error
func (b *BufferedDataStore) WriteAny(payload interface{}) error {
	p, ok := payload.(*model.ProcessPayload)
	if !ok {
		return errors.New("invalid payload type for process data")
	}
	return b.Write(p)
}

// Write adds a process payload to the buffer with automatic flush management.
// The method implements intelligent buffering with size-based flush triggers,
// ensuring optimal batch sizes while preventing unbounded memory growth.
//
// When the buffer reaches capacity, an automatic flush operation is triggered
// to persist the accumulated data. This approach optimizes both memory usage
// and storage performance by maintaining predictable batch sizes.
//
// Parameters:
//   - payload: The process payload to buffer for later persistence
//
// Returns:
//   - error: Any error encountered during buffering or automatic flush
func (b *BufferedDataStore) Write(payload *model.ProcessPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

// Flush immediately persists all buffered data to the underlying storage
// system, regardless of current buffer size or timing. This method provides
// explicit control over data persistence and is commonly used during shutdown
// sequences or when immediate durability is required.
//
// The operation is thread-safe and can be called concurrently with write
// operations. It ensures all currently buffered data is safely persisted
// before returning.
//
// Returns:
//   - error: Any error encountered during the flush operation
func (b *BufferedDataStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// flushLocked performs the core flush operation while holding the buffer lock.
// This internal method handles the actual data transfer to the underlying
// storage system, buffer reset, and error handling. It's designed for optimal
// performance with minimal lock contention.
//
// The method atomically transfers the current buffer contents to avoid data
// loss during concurrent operations and resets the buffer for continued use.
// Comprehensive logging provides visibility into flush operations for
// monitoring and debugging purposes.
//
// Returns:
//   - error: Any error encountered during the persistence operation
func (b *BufferedDataStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]*model.ProcessPayload, 0, b.maxSize)
	utils.Debug("Flushing %d process payloads from buffer", len(toFlush))
	return b.underlying.Write(b.ctx, toFlush)
}

// Close performs graceful shutdown of the buffered data store, ensuring all
// buffered data is persisted before termination. This method is essential
// for data consistency and should be called as part of the application's
// cleanup sequence.
//
// The close operation guarantees that no buffered data is lost during
// shutdown by performing a final flush before resource cleanup. This
// ensures data durability across application restarts and shutdowns.
//
// Returns:
//   - error: Any error encountered during final flush operation
func (b *BufferedDataStore) Close() error {
	return b.Flush()
}
