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

// File: gosight-server/internal/bufferengine/databuffer.go
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

// DataStore is an interface that defines the methods for writing process payloads.
// It is used to abstract the underlying data store implementation,
// allowing for different storage engines to be used (e.g., file, database).
type DataStore interface {
	Write(ctx context.Context, batches []*model.ProcessPayload) error
}

// BufferedDataStore is a buffered implementation of the DataStore interface.
// It buffers process payloads in memory and flushes them to the underlying data store
// when the buffer reaches a certain size or after a specified interval.
// This helps to reduce the number of write operations and improve performance.
// The buffer is protected by a mutex to ensure thread safety.
type BufferedDataStore struct {
	name          string
	underlying    datastore.DataStore
	buffer        []*model.ProcessPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
	ctx           context.Context
}

// NewBufferedDataStore creates a new BufferedDataStore instance.
// It takes a context, a name for the data store, an underlying data store,
// a maximum buffer size, and a flush interval as parameters.
// The flush interval is the time duration after which the buffer will be flushed
// to the underlying data store, even if the buffer size has not reached the maximum.
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

// Name returns the name of the buffered data store.
// This is used to identify the buffered data store in logs and metrics.
func (b *BufferedDataStore) Name() string {
	return b.name
}

// Interval returns the flush interval of the buffered data store.
// This is the time duration after which the buffer will be flushed
func (b *BufferedDataStore) Interval() time.Duration {
	return b.flushInterval
}

// WriteAny writes a process payload to the buffered data store.
// It takes an interface{} as a parameter and attempts to cast it to a *model.ProcessPayload.
// If the cast is successful, it calls the Write method to add the payload to the buffer.
// If the cast fails, it returns an error indicating that the payload type is invalid.
// This method is used to provide a generic interface for writing different types of payloads,
func (b *BufferedDataStore) WriteAny(payload interface{}) error {
	p, ok := payload.(*model.ProcessPayload)
	if !ok {
		return errors.New("invalid payload type for process data")
	}
	return b.Write(p)
}

// Write writes a process payload to the buffered data store.
// It appends the payload to the buffer and checks if the buffer size has reached
// the maximum size. If it has, it calls the flushLocked method to flush the buffer
// to the underlying data store. This method is protected by a mutex to ensure
// thread safety when accessing the buffer.
func (b *BufferedDataStore) Write(payload *model.ProcessPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

// Flush flushes the buffer to the underlying data store.
// It is called to ensure that any remaining payloads in the buffer are written
// to the underlying data store before closing the buffered data store.
// This method is protected by a mutex to ensure thread safety when accessing the buffer.
func (b *BufferedDataStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// flushLocked is a helper method that flushes the buffer to the underlying data store.
// It is called when the buffer size reaches the maximum size or when the Flush method is called.
// It clears the buffer after flushing to ensure that the next batch of payloads can be written.
// This method is protected by a mutex to ensure thread safety when accessing the buffer.
func (b *BufferedDataStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]*model.ProcessPayload, 0, b.maxSize)
	utils.Debug("Flushing %d process payloads from buffer", len(toFlush))
	return b.underlying.Write(b.ctx, toFlush)
}

// Close closes the buffered data store.
// It flushes any remaining payloads in the buffer to the underlying data store
// and releases any resources held by the buffered data store.
// This method is called when the buffered data store is no longer needed.
// It is important to call this method to ensure that all data is written
// to the underlying data store before closing the application.
func (b *BufferedDataStore) Close() error {
	return b.Flush()
}
