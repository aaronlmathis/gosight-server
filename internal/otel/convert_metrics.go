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
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	otlpcolpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

// OTLPToMetrics converts an OTLP ExportMetricsServiceRequest into GoSight’s []*model.Metric.
func (r *OTelReceiver) OTLPToMetrics(req *otlpcolpb.ExportMetricsServiceRequest) []*model.Metric {
    var out []*model.Metric

    for _, rm := range req.ResourceMetrics {
		resourceAttrs := convertKeyValueToStringMap(rm.Resource.Attributes)
		meta := buildMetaFromResourceAttrs(resourceAttrs)
        for _, ilm := range rm.ScopeMetrics {
            for _, om := range ilm.Metrics {
                m := &model.Metric{
                    Name:                   om.Name,
                    Description:            om.Description,
                    Unit:                   om.Unit,
                    DataType:               "",
                    AggregationTemporality: "",
                    DataPoints:             nil,
                    StorageResolution:      1,
                    Source:                 "otlp",
					Meta:				  meta,
                }

                switch data := om.Data.(type) {

                case *metricspb.Metric_Gauge:
                    m.DataType = "gauge"
                    for _, od := range data.Gauge.DataPoints {
                        value := extractNumberDataPointValue(od)
                        dp := model.DataPoint{
                            Attributes: convertKeyValueToMap(od.Attributes),
                            Timestamp:  time.Unix(0, int64(od.TimeUnixNano)),
                            Value:      value,
                            Exemplars:  convertOtelExemplars(od.Exemplars),
                        }
                        m.DataPoints = append(m.DataPoints, dp)
                    }

                case *metricspb.Metric_Sum:
                    m.DataType = "sum"
                    m.AggregationTemporality = data.Sum.AggregationTemporality.String()
                    for _, od := range data.Sum.DataPoints {
                        value := extractNumberDataPointValue(od)
                        dp := model.DataPoint{
                            Attributes:     convertKeyValueToMap(od.Attributes),
                            StartTimestamp: time.Unix(0, int64(od.StartTimeUnixNano)),
                            Timestamp:      time.Unix(0, int64(od.TimeUnixNano)),
                            Value:          value,
                            Exemplars:      convertOtelExemplars(od.Exemplars),
                        }
                        m.DataPoints = append(m.DataPoints, dp)
                    }

                case *metricspb.Metric_Histogram:
                    m.DataType = "histogram"
                    m.AggregationTemporality = data.Histogram.AggregationTemporality.String()
                    for _, od := range data.Histogram.DataPoints {
                        dp := model.DataPoint{
                            Attributes:     convertKeyValueToMap(od.Attributes),
                            StartTimestamp: time.Unix(0, int64(od.StartTimeUnixNano)),
                            Timestamp:      time.Unix(0, int64(od.TimeUnixNano)),
                            Count:          od.GetCount(),
                            Sum:            od.GetSum(),
                            BucketCounts:   od.BucketCounts,
                            ExplicitBounds: od.ExplicitBounds,
                            Exemplars:      convertOtelExemplars(od.Exemplars),
                        }
                        m.DataPoints = append(m.DataPoints, dp)
                    }

                case *metricspb.Metric_Summary:
                    m.DataType = "summary"
                    for _, od := range data.Summary.DataPoints {
                        var qvs []model.QuantileValue
                        for _, qt := range od.QuantileValues {
                            qvs = append(qvs, model.QuantileValue{
                                Quantile: qt.GetQuantile(),
                                Value:    qt.GetValue(),
                            })
                        }
                        dp := model.DataPoint{
                            Attributes:     convertKeyValueToMap(od.Attributes),
                            StartTimestamp: time.Unix(0, int64(od.StartTimeUnixNano)),
                            Timestamp:      time.Unix(0, int64(od.TimeUnixNano)),
                            Count:          od.GetCount(),
                            Sum:            od.GetSum(),
                            QuantileValues: qvs,

                        }
                        m.DataPoints = append(m.DataPoints, dp)
                    }

                default:
                    // Unknown or unsupported metric type—skip it.
                    continue
                }

                out = append(out, m)
            }
        }
    }

    return out
}