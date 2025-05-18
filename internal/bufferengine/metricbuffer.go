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

// File: gosight-server/internal/bufferengine/metricbuffer.go
// Description: Package bufferengine provides a buffered metric store implementation.
// It buffers metric payloads before writing them to the underlying metric store.
// The BufferedMetricStore is designed to improve performance by reducing the number of write operations
// to the underlying metric store.
// The buffer is protected by a mutex to ensure thread safety.
// The BufferedMetricStore supports a maximum buffer size and a flush interval.
package bufferengine

import (
	"errors"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// BufferedMetricStore is a buffered implementation of the MetricStore interface.
// It buffers metric payloads in memory and flushes them to the underlying metric store
// when the buffer reaches a certain size or after a specified interval.
// The buffer is protected by a mutex to ensure thread safety.
// The BufferedMetricStore is designed to improve performance by reducing the number of write operations
type BufferedMetricStore struct {
	name          string
	underlying    MetricStore
	buffer        []model.MetricPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
}

// MetricStore is an interface that defines the methods for writing metric payloads.
// It is used to abstract the underlying metric store implementation,
// allowing for different storage engines to be used (e.g., file, database).
// The MetricStore interface defines methods for writing metric payloads,
// flushing the buffer, closing the store, and retrieving the store name and flush interval.
type MetricStore interface {
	Write(payloads []model.MetricPayload) error
}

// NewBufferedMetricStore creates a new BufferedMetricStore instance.
// It initializes the buffer with a specified maximum size and flush interval.
// The flush interval determines how often the buffer is flushed to the underlying metric store.
// The maximum size determines when the buffer is flushed.
// The BufferedMetricStore is designed to improve performance by reducing the number of write operations
// to the underlying metric store.
func NewBufferedMetricStore(name string, store MetricStore, maxSize int, flushInterval time.Duration) *BufferedMetricStore {
	return &BufferedMetricStore{
		name:          name,
		underlying:    store,
		buffer:        make([]model.MetricPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
	}
}

// Name returns the name of the BufferedMetricStore.
// It is used to identify the store in logs and metrics.
func (b *BufferedMetricStore) Name() string {
	return b.name
}

// Interval returns the flush interval of the BufferedMetricStore.
// This is the time duration after which the buffer will be flushed to the underlying metric store,
// even if the buffer size has not reached the maximum.
func (b *BufferedMetricStore) Interval() time.Duration {
	return b.flushInterval
}

// WriteAny writes a payload to the buffered metric store.
// It takes an interface{} as a parameter and attempts to convert it to a MetricPayload.
// If the conversion is successful, it writes the payload to the buffer.
// If the conversion fails, it returns an error.
func (b *BufferedMetricStore) WriteAny(payload interface{}) error {
	p, ok := payload.(model.MetricPayload)
	if !ok {
		return errors.New("invalid payload type for metrics")
	}
	return b.Write(p)
}

// Write writes a metric payload to the buffered metric store.
// It appends the payload to the buffer and checks if the buffer size has reached the maximum.
// If the buffer size exceeds the maximum, it flushes the buffer to the underlying metric store.
// The Write method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
func (b *BufferedMetricStore) Write(payload model.MetricPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

// Flush flushes the buffer to the underlying metric store.
// It is called to ensure that all buffered metric payloads are written to the store.
// The Flush method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
func (b *BufferedMetricStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// flushLocked flushes the buffer to the underlying metric store.
// It is called when the buffer size exceeds the maximum size or when the Flush method is called.
// The flushLocked method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
func (b *BufferedMetricStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]model.MetricPayload, 0, b.maxSize)
	//utils.Debug("Flushing %d metric payloads from buffer", len(toFlush))
	return b.underlying.Write(toFlush)
}

// Close closes the BufferedMetricStore and flushes any remaining buffered metric payloads.
// It is called to ensure that all buffered metric payloads are written to the underlying metric store.
// The Close method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
// The Close method is typically called when the application is shutting down
func (b *BufferedMetricStore) Close() error {
	return b.Flush()
}
