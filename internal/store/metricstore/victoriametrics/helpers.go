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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/victoriametrics.go

package victoriametricstore

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// buildPrometheusFormat converts a batch of MetricPayloads into a Prometheus-compatible string format.
// Each metric is formatted with its name, labels, value, and timestamp.
func buildPrometheusFormat(batch []model.MetricPayload) string {
	var sb strings.Builder

	for _, payload := range batch {

		// Core labels from Meta + Tags
		baseLabels := BuildPromLabels(payload.Meta)

		for _, m := range payload.Metrics {
			fullName := normalizeMetricName(m.Name)

			// Process each data point in the metric
			for _, dp := range m.DataPoints {
				// Start with base Meta + Tags
				labels := make(map[string]string, len(baseLabels)+len(dp.Attributes))
				for k, v := range baseLabels {
					labels[k] = v
				}

				// Apply metric-specific dimensions (override any key)
				// Use dp.Attributes instead of m.Dimensions
				for k, v := range dp.Attributes {
					labels[k] = v
				}

				// Use dp.Value instead of m.Value
				// Use dp.Timestamp instead of payload.Timestamp
				dpTs := dp.Timestamp.UnixNano() / 1e6

				sb.WriteString(fmt.Sprintf("%s{%s} %f %d\n",
					fullName,
					formatLabelMap(labels),
					dp.Value,
					dpTs,
				))
			}
		}
	}

	return sb.String()
}

// normalizeMetricName normalizes the metric name for Prometheus compatibility.
func normalizeMetricName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, "/", "."))
}

// formatLabelMap prepares potential labels for Prometheus scraping.
// It combines payload.Meta tags and metric dimensions into a single map.
// It converts the labels map to a string in the format: key1="value1",key2="value2",...
// It allows Dimensions to override Meta tags if two keys are the same.

func formatLabelMap(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// BuildPromLabels constructs Prometheus-compatible labels from the given Meta object.
// It filters out any labels that are already present in the Meta object to avoid duplication.
// The resulting labels are returned as a map of key-value pairs.

func BuildPromLabels(meta *model.Meta) map[string]string {
	if meta == nil {
		return map[string]string{}
	}

	labels := map[string]string{}

	// Identity and system labels from Meta
	if meta.AgentID != "" {
		labels["agent_id"] = meta.AgentID
	}
	if meta.AgentVersion != "" {
		labels["agent_version"] = meta.AgentVersion
	}
	if meta.HostID != "" {
		labels["host_id"] = meta.HostID
	}
	if meta.EndpointID != "" {
		labels["endpoint_id"] = meta.EndpointID
	}
	if meta.Hostname != "" {
		labels["hostname"] = meta.Hostname
	}
	if meta.IPAddress != "" {
		labels["ip_address"] = meta.IPAddress
	}
	if meta.OS != "" {
		labels["os"] = meta.OS
	}
	if meta.OSVersion != "" {
		labels["os_version"] = meta.OSVersion
	}
	if meta.Platform != "" {
		labels["platform"] = meta.Platform
	}
	if meta.PlatformFamily != "" {
		labels["platform_family"] = meta.PlatformFamily
	}
	if meta.PlatformVersion != "" {
		labels["platform_version"] = meta.PlatformVersion
	}
	if meta.KernelArchitecture != "" {
		labels["kernel_architecture"] = meta.KernelArchitecture
	}
	if meta.KernelVersion != "" {
		labels["kernel_version"] = meta.KernelVersion
	}
	if meta.VirtualizationSystem != "" {
		labels["virtualization_system"] = meta.VirtualizationSystem
	}
	if meta.VirtualizationRole != "" {
		labels["virtualization_role"] = meta.VirtualizationRole
	}
	if meta.Architecture != "" {
		labels["architecture"] = meta.Architecture
	}
	if meta.Environment != "" {
		labels["environment"] = meta.Environment
	}
	if meta.Region != "" {
		labels["region"] = meta.Region
	}
	if meta.AvailabilityZone != "" {
		labels["availability_zone"] = meta.AvailabilityZone
	}
	if meta.InstanceID != "" {
		labels["instance_id"] = meta.InstanceID
	}
	if meta.InstanceType != "" {
		labels["instance_type"] = meta.InstanceType
	}
	if meta.AccountID != "" {
		labels["account_id"] = meta.AccountID
	}
	if meta.ProjectID != "" {
		labels["project_id"] = meta.ProjectID
	}
	if meta.ResourceGroup != "" {
		labels["resource_group"] = meta.ResourceGroup
	}
	if meta.VPCID != "" {
		labels["vpc_id"] = meta.VPCID
	}
	if meta.SubnetID != "" {
		labels["subnet_id"] = meta.SubnetID
	}
	if meta.ImageID != "" {
		labels["image_id"] = meta.ImageID
	}
	if meta.ServiceID != "" {
		labels["service_id"] = meta.ServiceID
	}
	if meta.ContainerID != "" {
		labels["container_id"] = meta.ContainerID
	}
	if meta.ContainerName != "" {
		labels["container_name"] = meta.ContainerName
	}
	if meta.PodName != "" {
		labels["pod_name"] = meta.PodName
	}
	if meta.ClusterName != "" {
		labels["cluster_name"] = meta.ClusterName
	}
	if meta.NodeName != "" {
		labels["node_name"] = meta.NodeName
	}
	if meta.ContainerImageID != "" {
		labels["container_image_id"] = meta.ContainerImageID
	}
	if meta.ContainerImageName != "" {
		labels["container_image_name"] = meta.ContainerImageName
	}

	if meta.Application != "" {
		labels["application"] = meta.Application
	}
	if meta.Service != "" {
		labels["service"] = meta.Service
	}
	if meta.Version != "" {
		labels["version"] = meta.Version
	}
	if meta.DeploymentID != "" {
		labels["deployment_id"] = meta.DeploymentID
	}
	if meta.PublicIP != "" {
		labels["public_ip"] = meta.PublicIP
	}
	if meta.PrivateIP != "" {
		labels["private_ip"] = meta.PrivateIP
	}
	if meta.MACAddress != "" {
		labels["mac_address"] = meta.MACAddress
	}
	if meta.NetworkInterface != "" {
		labels["network_interface"] = meta.NetworkInterface
	}

	// Labels (pre-filtered to avoid duplication)
	for k, v := range meta.Labels {
		if _, exists := labels[k]; !exists {
			labels[k] = v
		}
	}

	return labels
}

// BuildPromQL constructs a Prometheus Query Language (PromQL) query string
func BuildPromQL(metric string, filters map[string]string) string {
	if len(filters) == 0 {
		return metric
	}
	var parts []string
	for k, v := range filters {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf(`%s{%s}`, metric, strings.Join(parts, ","))
}

// parseDurationToSeconds converts a duration string (e.g., "5s", "1m", "2h") to seconds.
func parseDurationToSeconds(step string) (int, error) {
	d, err := time.ParseDuration(step)
	if err != nil {
		return 0, err
	}
	return int(d.Seconds()), nil
}
