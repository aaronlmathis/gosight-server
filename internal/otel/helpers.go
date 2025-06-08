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
	"fmt"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

// convertKeyValueToStringMap converts a slice of OTLP KeyValue into a map[string]string.
func convertKeyValueToStringMap(kvs []*commonpb.KeyValue) map[string]string {
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		switch v := kv.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			m[kv.Key] = v.StringValue
		case *commonpb.AnyValue_DoubleValue:
			m[kv.Key] = fmt.Sprintf("%v", v.DoubleValue)
		case *commonpb.AnyValue_IntValue:
			m[kv.Key] = fmt.Sprintf("%v", v.IntValue)
		case *commonpb.AnyValue_BoolValue:
			m[kv.Key] = fmt.Sprintf("%v", v.BoolValue)
		default:
			m[kv.Key] = fmt.Sprintf("%v", kv.Value)
		}
	}
	return m
}

// convertAnyValueMap converts OTLP attributes into map[string]interface{}.
func convertAnyValueMap(kvs []*commonpb.KeyValue) map[string]interface{} {
	m := make(map[string]interface{}, len(kvs))
	for _, kv := range kvs {
		switch v := kv.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			m[kv.Key] = v.StringValue
		case *commonpb.AnyValue_DoubleValue:
			m[kv.Key] = v.DoubleValue
		case *commonpb.AnyValue_IntValue:
			m[kv.Key] = v.IntValue
		case *commonpb.AnyValue_BoolValue:
			m[kv.Key] = v.BoolValue
		default:
			m[kv.Key] = fmt.Sprintf("%v", kv.Value)
		}
	}
	return m
}



// extractNumberDataPointValue handles the one‐of Value in NumberDataPoint, returning a float64.
func extractNumberDataPointValue(ndp *metricspb.NumberDataPoint) float64 {
    switch v := ndp.Value.(type) {
    case *metricspb.NumberDataPoint_AsDouble:
        return v.AsDouble
    case *metricspb.NumberDataPoint_AsInt:
        return float64(v.AsInt)
    default:
        // Fallback if neither is set (rare)
        return 0
    }
}

// convertKeyValueToMap turns an OTLP KeyValue slice into a map[string]string.
func convertKeyValueToMap(kvs []*commonpb.KeyValue) map[string]string {
    attrs := make(map[string]string, len(kvs))
    for _, kv := range kvs {
        switch v := kv.Value.Value.(type) {
        case *commonpb.AnyValue_StringValue:
            attrs[kv.Key] = v.StringValue
        case *commonpb.AnyValue_DoubleValue:
            attrs[kv.Key] = fmt.Sprintf("%v", v.DoubleValue)
        case *commonpb.AnyValue_IntValue:
            attrs[kv.Key] = fmt.Sprintf("%v", v.IntValue)
        case *commonpb.AnyValue_BoolValue:
            attrs[kv.Key] = fmt.Sprintf("%v", v.BoolValue)
        default:
            attrs[kv.Key] = fmt.Sprintf("%v", kv.Value)
        }
    }
    return attrs
}

// convertOtelExemplars converts OTLP Exemplars into our model.Exemplar slice.
func convertOtelExemplars(otlpExs []*metricspb.Exemplar) []model.Exemplar {
    var exs []model.Exemplar
    for _, ox := range otlpExs {
        var value float64
        switch v := ox.Value.(type) {
        case *metricspb.Exemplar_AsDouble:
            value = v.AsDouble
        case *metricspb.Exemplar_AsInt:
            value = float64(v.AsInt)
        default:
            value = 0
        }		
        ex := model.Exemplar{
            Value:     value,
            Timestamp: time.Unix(0, int64(ox.TimeUnixNano)),
        }
        if len(ox.TraceId) == 16 {
            ex.TraceID = hex.EncodeToString(ox.TraceId)
        }
        if len(ox.SpanId) == 8 {
            ex.SpanID = hex.EncodeToString(ox.SpanId)
        }
        fa := make(map[string]string, len(ox.FilteredAttributes))
        for _, fkv := range ox.FilteredAttributes {
            fa[fkv.Key] = fkv.Value.GetStringValue()
        }
        ex.FilteredAttributes = fa
        exs = append(exs, ex)
    }
    return exs
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
