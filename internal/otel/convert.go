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

package otel

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

// OTLPToTraceSpans converts an OTLP ExportTraceServiceRequest to a slice of model.TraceSpan.
// It extracts spans from the request, converting them to a format suitable for GoSight's model.
func (r *OTelReceiver) OTLPToTraceSpans(req *tracepb.ExportTraceServiceRequest) []model.TraceSpan {
	var spans []model.TraceSpan

	for _, rs := range req.ResourceSpans {
		resourceAttrs := extractAttributes(rs.Resource.Attributes)

		for _, ils := range rs.ScopeSpans {
			for _, span := range ils.Spans {
				start := time.Unix(0, int64(span.StartTimeUnixNano))
				end := time.Unix(0, int64(span.EndTimeUnixNano))

				attrs := extractAttributes(span.Attributes)

				spans = append(spans, model.TraceSpan{
					TraceID:       fmt.Sprintf("%x", span.TraceId),
					SpanID:        fmt.Sprintf("%x", span.SpanId),
					ParentSpanID:  fmt.Sprintf("%x", span.ParentSpanId),
					Name:          span.Name,
					StartTime:     start,
					EndTime:       end,
					DurationMs:    end.Sub(start).Seconds() * 1000,
					Attributes:    attrs,
					ResourceAttrs: resourceAttrs,
					StatusCode:    span.Status.Code.String(),
					StatusMessage: span.Status.Message,
				})
			}
		}
	}
	return spans
}

func extractAttributes(attrs []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string)
	for _, attr := range attrs {
		if attr == nil || attr.Value == nil {
			continue
		}
		switch v := attr.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			m[attr.Key] = v.StringValue
		case *commonpb.AnyValue_IntValue:
			m[attr.Key] = fmt.Sprintf("%d", v.IntValue)
		case *commonpb.AnyValue_DoubleValue:
			m[attr.Key] = fmt.Sprintf("%f", v.DoubleValue)
		case *commonpb.AnyValue_BoolValue:
			m[attr.Key] = strconv.FormatBool(v.BoolValue)
		default:
			m[attr.Key] = "[unsupported]" // fallback or custom handling
		}
	}
	return m
}
