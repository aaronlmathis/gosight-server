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

package rulestore

import (
	"context"
	"sync"

	"github.com/aaronlmathis/gosight-shared/model"
)

// MemoryRuleStore is an in-memory implementation of the RuleStore interface.
// It provides methods for adding, updating, deleting, and retrieving rules.
type MemoryRuleStore struct {
	rules map[string]model.AlertRule
	lock  sync.RWMutex
}

// NewMemoryStore creates a new MemoryRuleStore.
// It initializes an empty rules map.
// This store is not persistent and will lose data on application restart.
func NewMemoryStore() *MemoryRuleStore {
	return &MemoryRuleStore{rules: make(map[string]model.AlertRule)}
}

// AddRule adds a new rule to the store.
// If a rule with the same ID already exists, it will be overwritten.
func (s *MemoryRuleStore) AddRule(ctx context.Context, r model.AlertRule) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.rules[r.ID] = r
	return nil
}

// UpdateRule updates an existing rule in the store.
// If the rule does not exist, it will be added as a new rule.
func (s *MemoryRuleStore) UpdateRule(ctx context.Context, r model.AlertRule) error {
	return s.AddRule(ctx, r)
}

// DeleteRule deletes a rule from the store by its ID.
// If the rule does not exist, it will be ignored.
func (s *MemoryRuleStore) DeleteRule(ctx context.Context, id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.rules, id)
	return nil
}

// ListRules returns a list of all rules in the store.
// It returns an empty list if no rules are present.
func (s *MemoryRuleStore) ListRules(ctx context.Context) ([]model.AlertRule, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var out []model.AlertRule
	for _, rule := range s.rules {
		out = append(out, rule)
	}
	return out, nil
}

// GetActiveRules returns a list of all active rules in the store.
// An active rule is one that has its Enabled field set to true.
// It returns an empty list if no active rules are present.
func (s *MemoryRuleStore) GetActiveRules(ctx context.Context) ([]model.AlertRule, error) {
	all, _ := s.ListRules(ctx)
	var filtered []model.AlertRule
	for _, r := range all {
		if r.Enabled {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}
