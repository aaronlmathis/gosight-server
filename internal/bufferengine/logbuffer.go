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

// File: gosight-server/internal/bufferengine/logbuffer.go
// Description: Package bufferengine provides a buffered log store implementation.
// The BufferedLogStore buffers log entries before writing them to the underlying log store.
// It supports a maximum buffer size and a flush interval.
// The buffer is protected by a mutex to ensure thread safety.
// The BufferedLogStore is designed to improve performance by reducing the number of write operations
// to the underlying log store.
package bufferengine

import (
	"fmt"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// LogStore is an interface that defines the methods for writing log entries.
// It is used to abstract the underlying log store implementation,
// allowing for different storage engines to be used (e.g., file, database).
type LogStore interface {
	Write(entries []model.LogPayload) error
}

// BufferedLogStore is a buffered implementation of the LogStore interface.
// It buffers log entries in memory and flushes them to the underlying log store
// when the buffer reaches a certain size or after a specified interval.
type BufferedLogStore struct {
	name          string
	underlying    LogStore
	buffer        []model.LogPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
}

// NewBufferedLogStore creates a new BufferedLogStore instance.
// It initializes the buffer with a specified maximum size and flush interval.
// The flush interval determines how often the buffer is flushed to the underlying log store.
// The maximum size determines when the buffer is flushed.
// The BufferedLogStore is designed to improve performance by reducing the number of write operations
// to the underlying log store.
func NewBufferedLogStore(name string, store LogStore, maxSize int, flushInterval time.Duration) *BufferedLogStore {
	return &BufferedLogStore{
		name:          name,
		underlying:    store,
		buffer:        make([]model.LogPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
	}
}

// Name returns the name of the BufferedLogStore.
// It is used to identify the store in logs and metrics.
func (b *BufferedLogStore) Name() string {
	return b.name
}

// Interval returns the flush interval of the BufferedLogStore.
// This is the time duration after which the buffer will be flushed to the underlying log store,
// even if the buffer size has not reached the maximum.
func (b *BufferedLogStore) Interval() time.Duration {
	return b.flushInterval
}

// WriteAny writes a log entry to the BufferedLogStore.
// It takes an interface{} as a parameter and attempts to convert it to a LogPayload.
// If the conversion fails, it returns an error.
func (b *BufferedLogStore) WriteAny(payload interface{}) error {
	p, ok := payload.(model.LogPayload)
	if !ok {
		return fmt.Errorf("BufferedLogStore: invalid payload type %T", payload)
	}
	return b.Write(p)
}

// Write writes a log entry to the BufferedLogStore.
// It appends the entry to the buffer and checks if the buffer size has reached the maximum.
// If the buffer size exceeds the maximum, it flushes the buffer to the underlying log store.
// The Write method is thread-safe and uses a mutex to protect the buffer.
func (b *BufferedLogStore) Write(payload model.LogPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

// Flush flushes the buffer to the underlying log store.
// It is called to ensure that all buffered log entries are written to the store.
// The Flush method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
func (b *BufferedLogStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// flushLocked is a helper method that performs the actual flush operation.
// It is called when the buffer size exceeds the maximum or when the Flush method is called.
// It clears the buffer and writes the buffered log entries to the underlying log store.
// The flushLocked method is thread-safe and uses a mutex to protect the buffer.
// It returns an error if the flush operation fails.
func (b *BufferedLogStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]model.LogPayload, 0, b.maxSize)
	utils.Debug("Flushing %d log payloads from buffer", len(toFlush))
	return b.underlying.Write(toFlush)
}

// Close closes the BufferedLogStore and flushes any remaining log entries in the buffer.
// It is called to ensure that all buffered log entries are written to the underlying log store
// before the BufferedLogStore is closed. The Close method is thread-safe and uses a mutex
// to protect the buffer. It returns an error if the flush operation fails.
func (b *BufferedLogStore) Close() error {
	return b.Flush()
}
