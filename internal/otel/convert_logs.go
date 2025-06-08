// SPDX-License-Identifier: GPL-3.0-or-later

// Copyright (C) 2025 Aaron Mathis <aaron.mathis@gmail.com>

// This file is part of GoSight.

// GoSight is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// GoSight is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with GoSight. If not, see https://www.gnu.org/licenses/.
//

package otel

import (
	"encoding/hex"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	logpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

// OTLPToLogEntries converts an OTLP ExportLogsServiceRequest into GoSight’s []*model.LogEntry.
func (r *OTelReceiver) OTLPToLogEntries(req *logpb.ExportLogsServiceRequest) []*model.LogEntry {
	var entries []*model.LogEntry

	for _, resourceLogs := range req.ResourceLogs {
		// Convert resource-level attributes (e.g., service.name, k8s.pod.name, etc.)
		resourceAttrs := convertKeyValueToStringMap(resourceLogs.Resource.Attributes)
		// Build a Meta object from those resource attributes
		meta := buildMetaFromResourceAttrs(resourceAttrs)

		for _, scopeLogs := range resourceLogs.ScopeLogs {
			// You could also record scopeLogs.Scope.Name / Version if desired
			for _, lr := range scopeLogs.LogRecords {
				// Event timestamp
				timestamp := time.Unix(0, int64(lr.TimeUnixNano))
				// Observed timestamp (when collector saw it)
				observed := time.Unix(0, int64(lr.ObservedTimeUnixNano))
				// Convert record-level attributes into a map[string]interface{}
				attrs := convertAnyValueMap(lr.Attributes)

				// Build the LogEntry
				entry := &model.LogEntry{
					Timestamp:         timestamp,
					ObservedTimestamp: observed,
					SeverityText:      lr.SeverityText,
					SeverityNumber:    int32(lr.SeverityNumber),

					Body:              lr.Body.GetStringValue(),
					Flags:             lr.Flags,

					Level:    lr.SeverityText,
					Message:  lr.Body.GetStringValue(),
					Source:   resourceAttrs["service.name"], // e.g., service name if set
					Category: "",                            // populate if you have a “category” attribute

					Fields:     nil, // populate if you parse JSON‐style fields inside Attributes
					Labels:     nil, // populate if you have any labels to attach separately
					Attributes: attrs,
					Meta:       meta,
				}

				if len(lr.TraceId) == 16 {
					entry.TraceID = hex.EncodeToString(lr.TraceId)
				}
				if len(lr.SpanId) == 8 {
					entry.SpanID = hex.EncodeToString(lr.SpanId)
				}

				// If a “pid” attribute exists and is numeric, set entry.PID
				if pidStr, ok := attrs["pid"].(string); ok {
					if pid, err := strconv.Atoi(pidStr); err == nil {
						entry.PID = pid
					}
				}

				entries = append(entries, entry)
			}
		}
	}

	return entries
}