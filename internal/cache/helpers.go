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

// File: gosight-server/internal/cache/helpers.go
// Description: helpers.go contains utility functions for the cache package.
// These functions are used to manipulate and query the cache data structures.
package cache

import "strings"

// StringSet is a set of strings, implemented as a map with empty struct values.
// This allows for efficient membership testing and uniqueness.
// The empty struct consumes no memory, so this is a memory-efficient way to store a set of strings.
// The zero value of StringSet is an empty set.
type StringSet map[string]struct{}

// addToSet adds a value to a set, creating the set if it doesn't exist.
func addToSet(m map[string]StringSet, key, val string) {
	if m[key] == nil {
		m[key] = make(StringSet)
	}
	m[key][val] = struct{}{}
}

// containsMatch checks if a value contains a substring, ignoring case.
// This is a case-insensitive check, so it will return true if the value contains the substring
func containsMatch(value, substr string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(substr))
}
