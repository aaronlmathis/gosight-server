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

package telemetry

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/proto"
	collogpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logpb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// ConvertToModelPayload converts a protobuf MetricPayload to a model.MetricPayload.
func ConvertToModelPayload(pbPayload *proto.MetricPayload) model.MetricPayload {
	metrics := make([]model.Metric, 0, len(pbPayload.Metrics))
	var modelMeta *model.Meta

	for _, m := range pbPayload.Metrics {
		// Create metric with new OTLP structure
		metric := model.Metric{
			Namespace:    m.Namespace,
			SubNamespace: m.Subnamespace,
			Name:         m.Name,
			Unit:         m.Unit,
			DataType:     m.Type, // Map old Type to new DataType
			Source:       "agent",
		}

		// Create a single DataPoint from the old metric structure
		dataPoint := model.DataPoint{
			Timestamp:  m.Timestamp.AsTime(),
			Value:      m.Value,
			Attributes: m.Dimensions, // Map old Dimensions to new Attributes
		}

		// Handle StatisticValues if present (convert to histogram-like structure)
		if m.StatisticValues != nil {
			dataPoint.Count = uint64(m.StatisticValues.SampleCount)
			dataPoint.Sum = m.StatisticValues.Sum
			// For backwards compatibility, you might want to create bucket bounds
			// or handle this as a summary type metric
			metric.DataType = "summary"
		}

		metric.DataPoints = []model.DataPoint{dataPoint}

		if pbPayload.Meta != nil {
			modelMeta = convertProtoMetaToModelMeta(pbPayload.Meta)
		}

		metrics = append(metrics, metric)
	}

	return model.MetricPayload{
		AgentID:    pbPayload.AgentId,
		HostID:     pbPayload.HostId,
		Hostname:   pbPayload.Hostname,
		EndpointID: pbPayload.EndpointId,
		Timestamp:  pbPayload.Timestamp.AsTime(),
		Metrics:    metrics,
		Meta:       modelMeta,
	}
}

// ConvertProtoProcessPayload converts a protobuf ProcessPayload to a model.ProcessPayload.
func ConvertProtoProcessPayload(pb *proto.ProcessPayload) model.ProcessPayload {

	processes := make([]model.ProcessInfo, 0, len(pb.Processes))
	for _, p := range pb.Processes {
		labels := make(map[string]string, len(p.Labels))
		for k, v := range p.Labels {
			labels[k] = v
		}

		processes = append(processes, model.ProcessInfo{
			PID:        int(p.Pid),
			PPID:       int(p.Ppid),
			User:       p.User,
			Executable: p.Executable,
			Cmdline:    p.Cmdline,
			CPUPercent: p.CpuPercent,
			MemPercent: p.MemPercent,
			Threads:    int(p.Threads),
			StartTime:  p.StartTime.AsTime(),
			Labels:     labels,
		})
	}

	return model.ProcessPayload{
		AgentID:    pb.AgentId,
		HostID:     pb.HostId,
		Hostname:   pb.Hostname,
		EndpointID: pb.EndpointId,
		Timestamp:  pb.Timestamp.AsTime(),
		Processes:  processes,
		Meta:       convertProtoMetaToModelMeta(pb.Meta),
	}
}

// convertProtoMetaToModelMeta converts a protobuf Meta to a model.Meta.
func convertProtoMetaToModelMeta(pbMeta *proto.Meta) *model.Meta {
	if pbMeta == nil {
		return nil
	}
	return &model.Meta{
		AgentID:              pbMeta.AgentId,
		AgentVersion:         pbMeta.AgentVersion,
		HostID:               pbMeta.HostId,
		EndpointID:           pbMeta.EndpointId,
		Hostname:             pbMeta.Hostname,
		IPAddress:            pbMeta.IpAddress,
		OS:                   pbMeta.Os,
		OSVersion:            pbMeta.OsVersion,
		Platform:             pbMeta.Platform,
		PlatformFamily:       pbMeta.PlatformFamily,
		PlatformVersion:      pbMeta.PlatformVersion,
		KernelArchitecture:   pbMeta.KernelArchitecture,
		VirtualizationSystem: pbMeta.VirtualizationSystem,
		VirtualizationRole:   pbMeta.VirtualizationRole,
		KernelVersion:        pbMeta.KernelVersion,
		Architecture:         pbMeta.Architecture,
		CloudProvider:        pbMeta.CloudProvider,
		Region:               pbMeta.Region,
		AvailabilityZone:     pbMeta.AvailabilityZone,
		InstanceID:           pbMeta.InstanceId,
		InstanceType:         pbMeta.InstanceType,
		AccountID:            pbMeta.AccountId,
		ProjectID:            pbMeta.ProjectId,
		ResourceGroup:        pbMeta.ResourceGroup,
		VPCID:                pbMeta.VpcId,
		SubnetID:             pbMeta.SubnetId,
		ImageID:              pbMeta.ImageId,
		ServiceID:            pbMeta.ServiceId,
		ContainerID:          pbMeta.ContainerId,
		ContainerName:        pbMeta.ContainerName,
		PodName:              pbMeta.PodName,
		Namespace:            pbMeta.Namespace,
		ClusterName:          pbMeta.ClusterName,
		NodeName:             pbMeta.NodeName,
		ContainerImageID:     pbMeta.ContainerImageId,
		ContainerImageName:   pbMeta.ContainerImageName,
		Application:          pbMeta.Application,
		Environment:          pbMeta.Environment,
		Service:              pbMeta.Service,
		Version:              pbMeta.Version,
		// Add this missing field
		DeploymentID:     pbMeta.DeploymentId,
		PublicIP:         pbMeta.PublicIp,
		PrivateIP:        pbMeta.PrivateIp,
		MACAddress:       pbMeta.MacAddress,
		NetworkInterface: pbMeta.NetworkInterface,
		Labels:           pbMeta.Labels,
		Tags:             pbMeta.Tags, // Add this if it exists in proto
	}
}

// convertOTLPToModelMetricPayloads converts OTLP ExportMetricsServiceRequest to GoSight model.MetricPayload(s).
// This function extracts resource attributes and converts OTLP metrics to GoSight model format,
// preserving all metadata and metric data for proper processing by GoSight business logic.
func convertOTLPToModelMetricPayloads(req *colmetricpb.ExportMetricsServiceRequest) []model.MetricPayload {
	if req == nil || len(req.ResourceMetrics) == 0 {
		return []model.MetricPayload{}
	}

	var payloads []model.MetricPayload

	for _, resourceMetric := range req.ResourceMetrics {
		// Extract resource attributes to build Meta
		meta := convertOTLPResourceToMeta(resourceMetric.Resource)

		for _, scopeMetric := range resourceMetric.ScopeMetrics {
			// Collect metrics for this scope
			var metrics []model.Metric

			for _, otlpMetric := range scopeMetric.Metrics {
				convertedMetrics := convertOTLPMetricToModelMetrics(otlpMetric, scopeMetric.Scope)
				metrics = append(metrics, convertedMetrics...)
			}

			if len(metrics) > 0 {
				// Create payload with current timestamp if not available
				timestamp := time.Now()
				if len(metrics) > 0 && len(metrics[0].DataPoints) > 0 {
					timestamp = metrics[0].DataPoints[0].Timestamp
				}

				payload := model.MetricPayload{
					AgentID:    meta.AgentID,
					HostID:     meta.HostID,
					Hostname:   meta.Hostname,
					EndpointID: meta.EndpointID,
					Timestamp:  timestamp,
					Metrics:    metrics,
					Meta:       meta,
				}

				payloads = append(payloads, payload)
			}
		}
	}

	return payloads
}

// convertOTLPResourceToMeta converts OTLP Resource to GoSight Meta
func convertOTLPResourceToMeta(resource *resourcepb.Resource) *model.Meta {
	if resource == nil {
		return &model.Meta{}
	}

	meta := &model.Meta{
		Labels: make(map[string]string),
	}

	// Process resource attributes
	for _, attr := range resource.Attributes {
		if attr.Value == nil {
			continue
		}

		key := attr.Key
		var value string

		// Extract string value from AnyValue
		switch v := attr.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			value = v.StringValue
		case *commonpb.AnyValue_IntValue:
			value = fmt.Sprintf("%d", v.IntValue)
		case *commonpb.AnyValue_DoubleValue:
			value = fmt.Sprintf("%.6f", v.DoubleValue)
		case *commonpb.AnyValue_BoolValue:
			if v.BoolValue {
				value = "true"
			} else {
				value = "false"
			}
		default:
			continue
		}

		// Map OTLP standard attributes to GoSight Meta fields
		switch key {
		case "host.id":
			meta.HostID = value
		case "agent.id", "service.instance.id":
			meta.AgentID = value
		case "host.name":
			meta.Hostname = value
		case "endpoint.id":
			meta.EndpointID = value
		case "service.name":
			meta.Service = value
		case "service.version":
			meta.Version = value
		case "host.ip":
			meta.IPAddress = value
		case "os.type":
			meta.OS = value
		case "os.version":
			meta.OSVersion = value
		case "host.arch":
			meta.Architecture = value
		case "cloud.provider":
			meta.CloudProvider = value
		case "cloud.region":
			meta.Region = value
		case "cloud.availability_zone":
			meta.AvailabilityZone = value
		case "cloud.instance.id":
			meta.InstanceID = value
		case "cloud.instance.type":
			meta.InstanceType = value
		case "cloud.account.id":
			meta.AccountID = value
		case "cloud.project.id":
			meta.ProjectID = value
		case "container.id":
			meta.ContainerID = value
		case "container.name":
			meta.ContainerName = value
		case "container.image.id":
			meta.ContainerImageID = value
		case "container.image.name":
			meta.ContainerImageName = value
		case "k8s.pod.name":
			meta.PodName = value
		case "k8s.namespace.name":
			meta.Namespace = value
		case "k8s.cluster.name":
			meta.ClusterName = value
		case "k8s.node.name":
			meta.NodeName = value
		default:
			// Store unknown attributes as labels
			meta.Labels[key] = value
		}
	}

	return meta
}

// convertOTLPMetricToModelMetrics converts a single OTLP Metric to one or more GoSight model.Metric
func convertOTLPMetricToModelMetrics(otlpMetric *metricpb.Metric, scope *commonpb.InstrumentationScope) []model.Metric {
	if otlpMetric == nil {
		return []model.Metric{}
	}

	// Create base metric with new OTLP structure
	metric := model.Metric{
		Name:        otlpMetric.Name,
		Description: otlpMetric.Description,
		Unit:        otlpMetric.Unit,
		Source:      "otlp",
	}

	// Determine namespace and subnamespace from scope
	if scope != nil && scope.Name != "" {
		parts := strings.Split(scope.Name, ".")
		if len(parts) >= 1 {
			metric.Namespace = parts[0]
		}
		if len(parts) >= 2 {
			metric.SubNamespace = strings.Join(parts[1:], ".")
		}
	} else {
		metric.Namespace = "metrics"
	}

	switch data := otlpMetric.Data.(type) {
	case *metricpb.Metric_Gauge:
		metric.DataType = "gauge"
		for _, dp := range data.Gauge.DataPoints {
			dataPoint := model.DataPoint{
				Timestamp:  time.Unix(0, int64(dp.TimeUnixNano)),
				Value:      extractNumberDataPointValue(dp),
				Attributes: convertOTLPAttributes(dp.Attributes),
			}
			metric.DataPoints = append(metric.DataPoints, dataPoint)
		}

	case *metricpb.Metric_Sum:
		metric.DataType = "sum"
		metric.AggregationTemporality = data.Sum.AggregationTemporality.String()
		for _, dp := range data.Sum.DataPoints {
			dataPoint := model.DataPoint{
				StartTimestamp: time.Unix(0, int64(dp.StartTimeUnixNano)),
				Timestamp:      time.Unix(0, int64(dp.TimeUnixNano)),
				Value:          extractNumberDataPointValue(dp),
				Attributes:     convertOTLPAttributes(dp.Attributes),
			}
			metric.DataPoints = append(metric.DataPoints, dataPoint)
		}

	case *metricpb.Metric_Histogram:
		metric.DataType = "histogram"
		metric.AggregationTemporality = data.Histogram.AggregationTemporality.String()
		for _, dp := range data.Histogram.DataPoints {
			dataPoint := model.DataPoint{
				StartTimestamp: time.Unix(0, int64(dp.StartTimeUnixNano)),
				Timestamp:      time.Unix(0, int64(dp.TimeUnixNano)),
				Count:          dp.Count,
				Sum:            dp.GetSum(),
				BucketCounts:   dp.BucketCounts,
				ExplicitBounds: dp.ExplicitBounds,
				Attributes:     convertOTLPAttributes(dp.Attributes),
			}
			metric.DataPoints = append(metric.DataPoints, dataPoint)
		}

	case *metricpb.Metric_Summary:
		metric.DataType = "summary"
		for _, dp := range data.Summary.DataPoints {
			var quantiles []model.QuantileValue
			for _, qv := range dp.QuantileValues {
				quantiles = append(quantiles, model.QuantileValue{
					Quantile: qv.Quantile,
					Value:    qv.Value,
				})
			}

			dataPoint := model.DataPoint{
				StartTimestamp: time.Unix(0, int64(dp.StartTimeUnixNano)),
				Timestamp:      time.Unix(0, int64(dp.TimeUnixNano)),
				Count:          dp.Count,
				Sum:            dp.Sum,
				QuantileValues: quantiles,
				Attributes:     convertOTLPAttributes(dp.Attributes),
			}
			metric.DataPoints = append(metric.DataPoints, dataPoint)
		}
	}

	return []model.Metric{metric}
}

// Add helper function to extract value from NumberDataPoint
func extractNumberDataPointValue(dp *metricpb.NumberDataPoint) float64 {
	switch value := dp.Value.(type) {
	case *metricpb.NumberDataPoint_AsDouble:
		return value.AsDouble
	case *metricpb.NumberDataPoint_AsInt:
		return float64(value.AsInt)
	default:
		return 0
	}
}

// convertSeverityToLevel converts OTLP severity numbers to GoSight log levels
func convertSeverityToLevel(severity logpb.SeverityNumber) string {
	switch {
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_FATAL:
		return "critical"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_ERROR:
		return "error"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_WARN:
		return "warning"
	case severity >= logpb.SeverityNumber_SEVERITY_NUMBER_INFO:
		return "info"
	default:
		return "debug"
	}
}

// convertAttributesToMap converts OTLP attributes to string map
func convertAttributesToMap(attributes []*commonpb.KeyValue) map[string]string {
	result := make(map[string]string)
	for _, attr := range attributes {
		result[attr.Key] = attr.Value.GetStringValue()
	}
	return result
}

// convertOTLPToModelLogPayloads converts OTLP logs to GoSight model.LogPayload format
// This function extracts resource attributes and converts OTLP logs to GoSight model format,
// preserving all metadata and log data for proper processing by GoSight business logic.
func convertOTLPToModelLogPayloads(req *collogpb.ExportLogsServiceRequest) []model.LogPayload {
	if req == nil || len(req.ResourceLogs) == 0 {
		return []model.LogPayload{}
	}

	var payloads []model.LogPayload

	for _, resourceLogs := range req.ResourceLogs {
		// Extract resource attributes to build Meta
		baseMeta := convertOTLPResourceToMeta(resourceLogs.Resource)

		for _, scopeLogs := range resourceLogs.ScopeLogs {
			var logs []model.LogEntry

			for _, logRecord := range scopeLogs.LogRecords {
				// Convert OTLP timestamp (nanoseconds) to Go time
				timestamp := time.Unix(0, int64(logRecord.TimeUnixNano))
				observedTimestamp := time.Unix(0, int64(logRecord.ObservedTimeUnixNano))
				if timestamp.IsZero() {
					timestamp = time.Now()
				}
				if observedTimestamp.IsZero() {
					observedTimestamp = timestamp
				}

				// Extract message from body
				message := ""
				if logRecord.Body != nil {
					switch body := logRecord.Body.Value.(type) {
					case *commonpb.AnyValue_StringValue:
						message = body.StringValue
					case *commonpb.AnyValue_IntValue:
						message = fmt.Sprintf("%d", body.IntValue)
					case *commonpb.AnyValue_DoubleValue:
						message = fmt.Sprintf("%.6f", body.DoubleValue)
					case *commonpb.AnyValue_BoolValue:
						message = fmt.Sprintf("%t", body.BoolValue)
					case *commonpb.AnyValue_BytesValue:
						message = string(body.BytesValue)
					default:
						message = logRecord.Body.String()
					}
				}

				// Convert severity to level
				level := convertSeverityToLevel(logRecord.SeverityNumber)
				severityText := logRecord.SeverityText
				if severityText == "" {
					severityText = level
				}

				// Extract attributes for log entry
				fields := make(map[string]string)
				labels := make(map[string]string)
				attributes := make(map[string]interface{})

				// Create a copy of base Meta for this log entry
				logMeta := *baseMeta

				var source, category string
				var pid int

				for _, attr := range logRecord.Attributes {
					if attr.Value == nil {
						continue
					}

					key := attr.Key
					value := extractStringFromAnyValue(attr.Value)

					// Map specific attributes to Meta fields or log-specific fields
					switch key {
					case "service.name", "app.name":
						logMeta.Service = value
					case "service.version", "app.version":
						logMeta.AppVersion = value
					case "container.id":
						logMeta.ContainerID = value
					case "container.name":
						logMeta.ContainerName = value
					case "container.image.id":
						logMeta.ContainerImageID = value
					case "container.image.name":
						logMeta.ContainerImageName = value
					case "k8s.pod.name":
						logMeta.PodName = value
					case "k8s.namespace.name":
						logMeta.Namespace = value
					case "k8s.cluster.name":
						logMeta.ClusterName = value
					case "k8s.node.name":
						logMeta.NodeName = value
					case "source", "log.source":
						source = value
					case "category", "log.category":
						category = value
					case "pid", "process.pid":
						if intVal, err := strconv.Atoi(value); err == nil {
							pid = intVal
						}
					case "thread.id", "thread_id":
						attributes["thread.id"] = value
					case "logger.name":
						attributes["logger.name"] = value
					case "log.file.path", "path":
						attributes["log.file.path"] = value
					case "user", "process.owner":
						attributes["user"] = value
					case "executable", "process.executable":
						attributes["process.executable"] = value
					default:
						// Check if it's a tag/label
						if strings.HasPrefix(key, "tag.") ||
							key == "environment" || key == "deployment" ||
							key == "region" || key == "zone" {
							labels[key] = value
						} else {
							// Everything else goes to fields
							fields[key] = value
						}
						// Also add to generic attributes
						attributes[key] = value
					}
				}

				// Use scope name as source if not found in attributes
				if source == "" && scopeLogs.Scope != nil {
					source = scopeLogs.Scope.Name
				}
				if source == "" {
					source = "otlp"
				}

				// Use severity text as category if not found
				if category == "" {
					category = severityText
				}

				// Handle trace context if present
				var traceID, spanID string
				var flags uint32
				if len(logRecord.TraceId) == 16 {
					traceID = fmt.Sprintf("%x", logRecord.TraceId)
				}
				if len(logRecord.SpanId) == 8 {
					spanID = fmt.Sprintf("%x", logRecord.SpanId)
				}
				flags = logRecord.Flags

				// Create the log entry with new structure
				logEntry := model.LogEntry{
					Timestamp:         timestamp,
					ObservedTimestamp: observedTimestamp,
					SeverityText:      severityText,
					SeverityNumber:    int32(logRecord.SeverityNumber),
					Level:             level,
					Body:              message,
					Message:           message, // Keep for compatibility
					Source:            source,
					Category:          category,
					PID:               pid,
					Fields:            fields,
					Labels:            labels,
					Attributes:        attributes,
					TraceID:           traceID,
					SpanID:            spanID,
					Flags:             flags,
					Meta:              &logMeta, // Use the enriched Meta copy
				}

				logs = append(logs, logEntry)
			}

			if len(logs) > 0 {
				// Create payload with current timestamp if not available
				timestamp := time.Now()
				if len(logs) > 0 {
					timestamp = logs[0].Timestamp
				}

				payload := model.LogPayload{
					AgentID:    baseMeta.AgentID,
					HostID:     baseMeta.HostID,
					Hostname:   baseMeta.Hostname,
					EndpointID: baseMeta.EndpointID,
					Timestamp:  timestamp,
					Logs:       logs,
					Meta:       baseMeta,
				}

				payloads = append(payloads, payload)
			}
		}
	}

	return payloads
}

// extractStringFromAnyValue safely extracts string value from OTLP AnyValue
func extractStringFromAnyValue(value *commonpb.AnyValue) string {
	if value == nil {
		return ""
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", v.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%.6f", v.DoubleValue)
	case *commonpb.AnyValue_BoolValue:
		if v.BoolValue {
			return "true"
		}
		return "false"
	case *commonpb.AnyValue_BytesValue:
		return string(v.BytesValue)
	case *commonpb.AnyValue_ArrayValue:
		// Convert array to JSON-like string
		var elements []string
		for _, item := range v.ArrayValue.Values {
			elements = append(elements, extractStringFromAnyValue(item))
		}
		return "[" + strings.Join(elements, ",") + "]"
	case *commonpb.AnyValue_KvlistValue:
		// Convert key-value list to JSON-like string
		var pairs []string
		for _, kv := range v.KvlistValue.Values {
			val := extractStringFromAnyValue(kv.Value)
			pairs = append(pairs, fmt.Sprintf("%s:%s", kv.Key, val))
		}
		return "{" + strings.Join(pairs, ",") + "}"
	default:
		return value.String()
	}
}

// convertOTLPAttributes converts OTLP KeyValue attributes to GoSight dimensions map
func convertOTLPAttributes(attributes []*commonpb.KeyValue) map[string]string {
	dims := make(map[string]string, len(attributes))

	for _, attr := range attributes {
		if attr.Value == nil {
			continue
		}

		var value string
		switch v := attr.Value.Value.(type) {
		case *commonpb.AnyValue_StringValue:
			value = v.StringValue
		case *commonpb.AnyValue_IntValue:
			value = fmt.Sprintf("%d", v.IntValue)
		case *commonpb.AnyValue_DoubleValue:
			value = fmt.Sprintf("%.6f", v.DoubleValue)
		case *commonpb.AnyValue_BoolValue:
			if v.BoolValue {
				value = "true"
			} else {
				value = "false"
			}
		default:
			continue
		}

		dims[attr.Key] = value
	}

	return dims
}
