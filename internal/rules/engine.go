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

// Package rules provides the core logic for evaluating alert rules
// and emitting events based on metric data.
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aaronlmathis/gosight-server/internal/alerts"
	"github.com/aaronlmathis/gosight-server/internal/store/rulestore"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// Evaluator is responsible for evaluating alert rules against incoming metrics and logs.
// It uses a RuleStore to manage the rules and an AlertManager to handle
// the state of alerts. The Evaluator maintains a history of metrics
// for each rule and endpoint combination, allowing it to track the state
// of alerts over time. The firing map is used to track which rules are currently
// firing for each endpoint, preventing duplicate alerts from being emitted.
type Evaluator struct {
	store    rulestore.RuleStore
	AlertMgr *alerts.Manager
	history  map[string][]model.Metric
	firing   map[string]bool // ruleID + endpointID
}

// NewEvaluator creates a new Evaluator instance.
// It takes a RuleStore for rule management and an *events.emitter
// for emitting events. The history map is used to track metrics
// for each rule and endpoint combination.
// The history is keyed by rule ID and endpoint ID, allowing for
// efficient tracking of metrics over time.

func NewEvaluator(store rulestore.RuleStore, alertMgr *alerts.Manager) *Evaluator {
	return &Evaluator{
		store:    store,
		AlertMgr: alertMgr,
		history:  make(map[string][]model.Metric),
		firing:   make(map[string]bool),
	}
}

// EvaluateMetric processes the given metrics and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the metrics.
// The evaluation is done in the context of the provided context.Context.
// The metrics are expected to be in the format of model.Metric,
// and the metadata is expected to be in the format of model.Meta.
func (e *Evaluator) EvaluateMetric(ctx context.Context, metrics []model.Metric, meta *model.Meta) {

	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, rule := range activeRules {

		if !rule.Enabled {
			continue
		}

		if rule.Type != "metric" {
			continue
		}

		if !ruleMatchLabels(rule.Match, meta) {
			continue
		}

		metricName := fmt.Sprintf("%s.%s.%s", rule.Scope.Namespace, rule.Scope.SubNamespace, rule.Scope.Metric)

		var matched *model.Metric
		for _, m := range metrics {
			full := strings.ToLower(fmt.Sprintf("%s.%s.%s", m.Namespace, m.SubNamespace, m.Name))

			if full == metricName {
				matched = &m
				break
			}
		}

		if matched == nil {
			continue
		}

		// Use helper function to extract value from DataPoints
		matchedValue := getMetricValue(matched)
		firing := evaluateExpression(rule.Expression, matched)
		key := rule.ID + "|" + meta.EndpointID

		if firing {
			if !e.firing[key] {
				e.firing[key] = true
				// Use extracted value instead of matched.Value
				e.AlertMgr.HandleState(ctx, rule, meta, matchedValue, true)
			}
		} else {
			if e.firing[key] {
				delete(e.firing, key)
				// Use extracted value instead of matched.Value
				e.AlertMgr.HandleState(ctx, rule, meta, matchedValue, false)
			}
		}
	}
}

// EvaluateLogs processes the given logs and metadata,
// checking them against active rules in the store.
// It emits events when rules are triggered based on the logs.
// The evaluation is done in the context of the provided context.Context.
// The logs are expected to be in the format of model.LogEntry,
// and the metadata is expected to be in the format of model.Meta.
// Logs are point-in-time events, so they are always evaluated immediately.
func (e *Evaluator) EvaluateLogs(ctx context.Context, logs []model.LogEntry, meta *model.Meta) {
	activeRules, err := e.store.GetActiveRules(ctx)
	if err != nil {
		utils.Error("Failed to fetch active rules: %v", err)
		return
	}

	for _, log := range logs {
		for _, rule := range activeRules {
			utils.Debug("Evaluating log rule: %s", rule.ID)
			if !rule.Enabled {
				continue
			}
			if rule.Type != "log" {
				continue
			}
			if !ruleMatchLabels(rule.Match, meta) {
				continue
			}
			if rule.Match.Category != "" && rule.Match.Category != log.Category {
				continue
			}
			if rule.Match.Source != "" && rule.Match.Source != log.Source {
				continue
			}
			firing := evaluateLogExpression(rule.Expression, log)
			if firing {
				e.AlertMgr.HandleLogState(ctx, rule, meta, log, true)
			}
		}
	}
}

// evaluateLogExpression evaluates the log entry against the rule's expression.
func evaluateLogExpression(expr model.Expression, log model.LogEntry) bool {
	val := ""
	if expr.Datatype == "level" {
		val = log.Level
	} else if expr.Datatype == "message" {
		val = log.Message
	} else {
		val = log.Source
	}

	switch expr.Operator {
	case "contains":
		return strings.Contains(val, toString(expr.Value))
	case "=", "==":
		return val == toString(expr.Value)
	case "!=":
		return val != toString(expr.Value)
	case "regex":
		re, err := regexp.Compile(toString(expr.Value))
		if err != nil {
			return false
		}
		return re.MatchString(val)
	default:
		return false
	}
}

// evaluateLogExpression evaluates the log entry against the rule's expression.
func evaluateExpression(expr model.Expression, m *model.Metric) bool {
	// Extract value using helper function instead of m.Value
	metricValue := getMetricValue(m)

	switch expr.Operator {
	case ">":
		return metricValue > toFloat(expr.Value)
	case "<":
		return metricValue < toFloat(expr.Value)
	case ">=":
		return metricValue >= toFloat(expr.Value)
	case "<=":
		return metricValue <= toFloat(expr.Value)
	case "=", "==":
		return metricValue == toFloat(expr.Value)
	case "!=":
		return metricValue != toFloat(expr.Value)
	case "contains":
		return strings.Contains(toString(metricValue), toString(expr.Value))
	case "regex":
		re, err := regexp.Compile(toString(expr.Value))
		if err != nil {
			return false
		}
		return re.MatchString(toString(metricValue))
	default:
		return false
	}
}

// ruleMatchLabels checks if the rule's match labels match the given metadata labels.

func ruleMatchLabels(match model.MatchCriteria, meta *model.Meta) bool {
	if len(match.EndpointIDs) > 0 {
		found := false
		for _, id := range match.EndpointIDs {
			if id == meta.EndpointID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for k, v := range match.Labels {
		if meta.Labels[k] != v {
			return false
		}
	}
	return true
}

// toFloat converts various types to float64.
// It handles float64, int, and string types.
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0.0
	}
}

// toString converts various types to string.
// It handles string, float64, and int types.
// It returns the string representation of the value.
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.2f", val)
	case int:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// getMetricValue extracts the first data point value from a metric
func getMetricValue(metric *model.Metric) float64 {
	if len(metric.DataPoints) == 0 {
		return 0
	}

	// For most metrics, use the first data point's value
	dp := metric.DataPoints[0]

	// Handle different metric types
	switch metric.DataType {
	case "gauge", "sum":
		return dp.Value
	case "histogram":
		// For histograms, you might want count, sum, or average
		if dp.Count > 0 {
			return dp.Sum / float64(dp.Count) // Return average
		}
		return dp.Sum
	case "summary":
		// For summaries, return sum or calculate from quantiles
		return dp.Sum
	default:
		return dp.Value
	}
}

// getMetricDimensions extracts the first data point attributes from a metric
func getMetricDimensions(metric *model.Metric) map[string]string {
	if len(metric.DataPoints) == 0 {
		return make(map[string]string)
	}
	return metric.DataPoints[0].Attributes
}
