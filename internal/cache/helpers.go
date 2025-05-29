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

package cache

import "strings"

// StringSet represents a memory-efficient set of unique strings implemented using
// a map with empty struct values. This design provides O(1) membership testing
// and insertion operations while minimizing memory overhead.
//
// The empty struct{} type consumes zero bytes of memory, making this implementation
// optimal for storing large sets of strings where only uniqueness matters.
// StringSet supports typical set operations like membership testing, insertion,
// and iteration over elements.
//
// Usage:
//
//	set := make(StringSet)
//	set["example"] = struct{}{}
//	if _, exists := set["example"]; exists {
//	    // Element is in the set
//	}
type StringSet map[string]struct{}

// addToSet is a utility function that adds a value to a string set within a nested map structure.
// It automatically initializes the inner StringSet if it doesn't exist for the given key.
// This function is commonly used in cache implementations to maintain indexes of
// related string values grouped by categories.
//
// Parameters:
//   - m: The outer map containing StringSet values indexed by string keys
//   - key: The key in the outer map to access or create the StringSet
//   - val: The string value to add to the StringSet
//
// The function ensures thread-safety must be handled by the caller if concurrent
// access is required.
func addToSet(m map[string]StringSet, key, val string) {
	if m[key] == nil {
		m[key] = make(StringSet)
	}
	m[key][val] = struct{}{}
}

// containsMatch performs a case-insensitive substring search to determine if a value
// contains a given substring. This function is commonly used in cache filtering
// and search operations where case sensitivity is not desired.
//
// The function converts both the value and substring to lowercase before performing
// the comparison, ensuring consistent matching behavior regardless of the original
// case of the input strings.
//
// Parameters:
//   - value: The string to search within
//   - substr: The substring to search for
//
// Returns:
//   - bool: true if the value contains the substring (case-insensitive), false otherwise
//
// Example:
//
//	containsMatch("SystemMetric", "metric") // returns true
//	containsMatch("process.cpu", "CPU")     // returns true
//	containsMatch("memory", "disk")         // returns false
func containsMatch(value, substr string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(substr))
}
