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

// File: gosight-server/internal/bufferengine/engine.go
// Description: Package bufferengine provides a buffered store engine for GoSight.

package bufferengine

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-shared/utils"
)

// BufferedStore is an interface that defines the methods for a buffered store.
// It is used to abstract the underlying store implementation,
// allowing for different storage engines to be used (e.g., file, database).
// The BufferedStore interface defines methods for writing data, flushing the buffer,
// closing the store, and retrieving the store name and flush interval.
type BufferedStore interface {
	WriteAny(payload interface{}) error
	Flush() error
	Close() error
	Name() string
	Interval() time.Duration
}

// BufferEngine is a struct that manages multiple buffered stores.
// It is responsible for starting and stopping the stores, as well as flushing the data at regular intervals.
type BufferEngine struct {
	stores        []BufferedStore
	flushInterval time.Duration
	maxWorkers    int
	ctx           context.Context
	wg            sync.WaitGroup
}

// NewBufferEngine creates a new BufferEngine instance.
// It takes a context, flush interval, and maximum number of workers as parameters.
// The flush interval is used to determine how often the data should be flushed to the stores.
// The maximum number of workers is used to limit the number of concurrent goroutines
// that can be used for flushing the data.
func NewBufferEngine(ctx context.Context, flushInterval time.Duration, maxWorkers int) *BufferEngine {
	return &BufferEngine{
		flushInterval: flushInterval,
		maxWorkers:    maxWorkers,
		ctx:           ctx,
	}
}

// RegisterStore registers a new buffered store with the BufferEngine.
// It adds the store to the list of stores managed by the engine.
// The store must implement the BufferedStore interface.
// The engine will manage the lifecycle of the store, including starting and stopping it.
// The store will be flushed at regular intervals as specified by the flush interval.
func (e *BufferEngine) RegisterStore(store BufferedStore) {
	e.stores = append(e.stores, store)
	utils.Info("BufferEngine registered store: %s", store.Name())
}

// Start starts the BufferEngine and its registered stores.
// It launches a goroutine for each store that will flush the data at regular intervals.
// The flush interval is determined by the store's Interval method.
// The engine will also listen for a cancellation signal from the context.
// When the context is cancelled, the engine will stop all stores and wait for them to finish flushing.
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

// Stop stops the BufferEngine and all its registered stores.
// It waits for all background flush routines to finish before closing the stores.
// The engine will also log any errors encountered while closing the stores.
// This method should be called when the engine is no longer needed,
// such as during application shutdown.
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
