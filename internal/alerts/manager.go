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

// Package alerts provides a manager for handling alert instances.
// It manages the state of alerts, including firing and resolving them,
// and dispatches events related to alert state changes.
// It also provides a way to list active alerts and handle log-based alerts.
// The manager uses a store to persist alert instances and a dispatcher
// to trigger actions based on alert state changes.
package alerts

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/core/events/dispatcher"
	"github.com/aaronlmathis/gosight-server/internal/events"
	"github.com/aaronlmathis/gosight-server/internal/store/alertstore"
	"github.com/aaronlmathis/gosight-server/internal/websocket"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/google/uuid"
)

// Manager is responsible for managing alert instances.
// It handles the state of alerts, including firing and resolving them,
// and dispatches events related to alert state changes.
type Manager struct {
	lock       sync.RWMutex
	active     map[string]*model.AlertInstance // key: ruleID|endpointID
	emitter    *events.Emitter
	dispatcher *dispatcher.Dispatcher
	store      alertstore.AlertStore
	hub        *websocket.AlertsHub
}

// NewManager creates a new Manager instance.
// It takes an emitter for emitting events, a dispatcher for triggering actions,
// a store for persisting alert instances, and a hub for broadcasting alerts.
// The manager maintains a map of active alert instances, keyed by rule ID and endpoint ID.
// It also provides methods for handling alert state changes and listing active alerts.
func NewManager(emitter *events.Emitter, dispatcher *dispatcher.Dispatcher, store alertstore.AlertStore, hub *websocket.AlertsHub) *Manager {
	return &Manager{
		active:     make(map[string]*model.AlertInstance),
		emitter:    emitter,
		dispatcher: dispatcher,
		store:      store,
		hub:        hub,
	}
}

// key generates a unique key for the alert instance based on the rule ID and endpoint ID.
// This key is used to store and retrieve alert instances from the active map.
func key(ruleID, endpointID string) string {
	return ruleID + "|" + endpointID
}

// HandleState processes the state of an alert based on the given rule, metadata, value, and triggered status.
func (m *Manager) HandleState(ctx context.Context, rule model.AlertRule, meta *model.Meta, value float64, triggered bool) {
	k := key(rule.ID, meta.EndpointID)
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	current := m.active[k]

	if triggered {
		if current != nil {
			if rule.Options.RepeatInterval != "" {
				repeatDur, _ := time.ParseDuration(rule.Options.RepeatInterval)
				if now.Sub(current.LastFired) < repeatDur {
					return
				}
			}
			current.LastFired = now
			current.LastValue = value
			_ = m.store.UpsertAlert(ctx, current)
			m.hub.Broadcast(*current)
			m.emitAlertFiringEvent(ctx, rule, meta, now)
			m.dispatchFiringEvent(ctx, rule, meta, current.Target, now, current.Message)
			return
		}

		scope, target := inferScopeAndTarget(meta)

		inst := &model.AlertInstance{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			State:      "firing",
			Previous:   "ok",
			Scope:      scope,
			Target:     target,
			FirstFired: now,
			LastFired:  now,
			LastOK:     now,
			LastValue:  value,
			Level:      rule.Level,
			Message:    rule.Message,
			Labels:     utils.SafeCopyLabels(meta),
		}

		event := model.EventEntry{
			Timestamp: now,
			Level:     rule.Level,
			Category:  "alert",
			Message:   rule.Message,
			Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
			Scope:     scope,
			Target:    target,
			Meta:      utils.SafeCopyLabels(meta),
		}
		event.Meta["rule_id"] = rule.ID

		if len(rule.Actions) > 0 {
			for _, actionID := range rule.Actions {
				utils.Debug("Triggering action ID: %s for alert rule: %s", actionID, rule.ID)
				m.dispatcher.TriggerActionByID(ctx, actionID, event)
			}
			m.active[k] = inst
			_ = m.store.UpsertAlert(ctx, inst)
			m.hub.Broadcast(*inst)
			m.emitter.Emit(ctx, event)
			return
		}

		m.active[k] = inst
		_ = m.store.UpsertAlert(ctx, inst)
		m.hub.Broadcast(*inst)
		m.emitAlertFiringEvent(ctx, rule, meta, now)
		m.dispatchFiringEvent(ctx, rule, meta, target, now, rule.Message)
	} else {
		if current != nil {
			delete(m.active, k)
			m.emitAlertResolvedEvent(ctx, rule, meta, now)
			if rule.Options.NotifyOnResolve {
				m.dispatchResolvedEvent(ctx, rule, meta, current.Target, now, "Resolved: "+rule.Message)
			}
			_ = m.store.ResolveAlert(ctx, rule.ID, current.Target, now)
			m.hub.Broadcast(*current)
		}
	}
}

// HandleLogState processes the state of a log-based alert based on the given rule, metadata, log entry, and triggered status.
// It creates a new alert instance if the alert is triggered and updates the existing instance if it is already firing.
// It also emits an event for the log alert and triggers any actions associated with the rule.
// The log entry is expected to be in the format of model.LogEntry, and the metadata is expected to be in the format of model.Meta.
func (m *Manager) HandleLogState(ctx context.Context, rule model.AlertRule, meta *model.Meta, log model.LogEntry, triggered bool) {
	k := key(rule.ID, meta.EndpointID+"|"+log.Timestamp.Format(time.RFC3339Nano))
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	if triggered {
		inst := &model.AlertInstance{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			State:      "firing",
			Previous:   "ok",
			Scope:      "endpoint",
			Target:     meta.EndpointID,
			FirstFired: now,
			LastFired:  now,
			LastOK:     now,
			LastValue:  0,
			Level:      rule.Level,
			Message:    rule.Message + ": " + log.Message,
			Labels:     utils.SafeCopyLabels(meta),
		}

		event := model.EventEntry{
			Timestamp: now,
			Level:     rule.Level,
			Category:  "log_alert",
			Message:   rule.Message + ": " + log.Message,
			Source:    log.Source,
			Scope:     "endpoint",
			Target:    meta.EndpointID,
			Meta:      utils.SafeCopyLabels(meta),
		}
		event.Meta["rule_id"] = rule.ID

		if len(rule.Actions) > 0 {
			for _, actionID := range rule.Actions {
				m.dispatcher.TriggerActionByID(ctx, actionID, event)
			}
			m.active[k] = inst
			_ = m.store.UpsertAlert(ctx, inst)
			m.hub.Broadcast(*inst)
			m.emitter.Emit(ctx, event)
			return
		}

		m.active[k] = inst
		_ = m.store.UpsertAlert(ctx, inst)
		m.hub.Broadcast(*inst)
		m.emitLogAlertFiringEvent(ctx, rule, meta, log, now)
	}
}

// emitAlertFiringEvent emits an event for a firing alert.
// It creates an EventEntry with the alert's details and dispatches it to the event emitter and dispatcher.
// The event includes the alert's timestamp, level, category, message, source, scope, target, and metadata.
// The event is also broadcasted to the websocket hub for real-time updates.
func (m *Manager) emitAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	scope, target := inferScopeAndTarget(meta)
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   rule.Message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyLabels(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

// emitAlertResolvedEvent emits an event for a resolved alert.
// It creates an EventEntry with the alert's details and dispatches it to the event emitter and dispatcher.
// The event includes the alert's timestamp, level, category, message, source, scope, target, and metadata.
// The event is also broadcasted to the websocket hub for real-time updates.
func (m *Manager) emitAlertResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	scope, target := inferScopeAndTarget(meta)
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   "Resolved: " + rule.Message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyLabels(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

// emitLogAlertFiringEvent emits an event for a log-based alert firing.
// It creates an EventEntry with the log's details and dispatches it to the event emitter and dispatcher.
// The event includes the log's timestamp, level, category, message, source, scope, target, and metadata.
// The event is also broadcasted to the websocket hub for real-time updates.
func (m *Manager) emitLogAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, log model.LogEntry, now time.Time) {
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "log_alert",
		Message:   rule.Message + ": " + log.Message,
		Source:    log.Source,
		Scope:     "endpoint",
		Target:    meta.EndpointID,
		Meta:      utils.SafeCopyLabels(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

// ListActive returns a list of all active alert instances.
// It iterates over the active map and appends each alert instance to a slice.
// The slice is then returned to the caller.
func (m *Manager) ListActive() []model.AlertInstance {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var list []model.AlertInstance
	for _, v := range m.active {
		list = append(list, *v)
	}
	return list
}

// inferScopeAndTarget determines the scope and target for the alert instance.
// It checks if the endpoint ID is present in the metadata.
// If it is, the scope is set to "endpoint" and the target is set to the endpoint ID.
// If the endpoint ID is not present, the scope is set to "global" and the target is set to "gosight-core".
func inferScopeAndTarget(meta *model.Meta) (string, string) {
	if meta.EndpointID != "" {
		return "endpoint", meta.EndpointID
	}
	return "global", "gosight-core"
}

// dispatchFiringEvent dispatches a firing event for the alert instance.
// It creates an EventEntry with the alert's details and dispatches it to the event dispatcher.
// The event includes the alert's timestamp, level, category, message, source, scope, target, and metadata.
// The event is also broadcasted to the websocket hub for real-time updates.
func (m *Manager) dispatchFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, target string, now time.Time, message string) {
	scope, target := inferScopeAndTarget(meta)
	m.dispatcher.Dispatch(ctx, model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyLabels(meta),
	})
}

// dispatchResolvedEvent dispatches a resolved event for the alert instance.
// It creates an EventEntry with the alert's details and dispatches it to the event dispatcher.
// The event includes the alert's timestamp, level, category, message, source, scope, target, and metadata.
// The event is also broadcasted to the websocket hub for real-time updates.
func (m *Manager) dispatchResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, target string, now time.Time, message string) {
	scope, target := inferScopeAndTarget(meta)
	m.dispatcher.Dispatch(ctx, model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyLabels(meta),
	})
}
