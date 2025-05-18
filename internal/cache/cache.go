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

// File: gosight-server/internal/cache/cache.go
// Description: Package cache provides a unified cache for the GoSight server.
// It includes caches for processes, metrics, tags, logs, and other components.
// The cache is used to store and retrieve data efficiently.
package cache

// Cache is a struct that holds all the caches used in the GoSight server.
// It includes caches for processes, metrics, tags, logs, and other components.
// Each cache is represented by a specific type, such as ProcessCache, MetricCache, TagCache, and LogCache.
type Cache struct {
	Processes ProcessCache
	Metrics   MetricCache
	Tags      TagCache
	Logs      LogCache
	/*
		Agents    AgentCache
		Endpoints EndpointCache
		Alerts    AlertCache
		Events    EventCache
	*/
}
