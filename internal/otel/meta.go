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
	"strconv"
	"strings"

	"github.com/aaronlmathis/gosight-shared/model"
)

// buildMetaFromResourceAttrs constructs a *model.Meta by mapping known OTLP resource‐level
// attributes (in attrs) into the corresponding fields of model.Meta. Any attribute not
// explicitly mapped here will be ignored; you can extend this function to cover more keys
// as needed.
func buildMetaFromResourceAttrs(attrs map[string]string) *model.Meta {
    meta := &model.Meta{
        // Agent / Host / Endpoint context (if your collector sets these)
        HostID:     attrs["host.id"],     // e.g., “i-0123456789abcdef0”
        Hostname:   attrs["host.name"],   // e.g., “ip-10-0-1-23.ec2.internal”
        IPAddress:  attrs["host.ip"],     // if you populate host.ip
        // OS / Platform / Kernel
        OS:                   attrs["host.os"],               // e.g., “linux”
        OSVersion:            attrs["host.os.version"],       // e.g., “5.10.0-8-amd64”
        Platform:             attrs["host.platform"],         // e.g., “ubuntu”
        PlatformFamily:       attrs["host.platform.family"],  // e.g., “debian”
        PlatformVersion:      attrs["host.platform.version"], // e.g., “20.04”
        KernelArchitecture:   attrs["host.arch"],             // e.g., “x86_64”
        KernelVersion:        attrs["host.kernel.version"],   // e.g., “5.10.0-8-amd64”
        Architecture:         attrs["host.arch"],             // (duplicate of host.arch)
        VirtualizationSystem: attrs["host.container"],        // e.g., “docker” or “k8s”
        VirtualizationRole:   attrs["host.container.role"],   // e.g., “guest”

        // Cloud / IaaS context
        CloudProvider:    attrs["cloud.provider"],        // “aws”, “azure”, “gcp”
        Region:           attrs["cloud.region"],          // e.g., “us-west-2”
        AvailabilityZone: attrs["cloud.zone"],            // e.g., “us-west-2a”
        AccountID:        attrs["cloud.account.id"],      // e.g., AWS Account ID
        ProjectID:        attrs["cloud.account.id"],      // GCP puts project_id here
        ResourceGroup:    attrs["cloud.resource.group"],  // Azure resource group
        InstanceID:       attrs["cloud.instance.id"],     // e.g., EC2 instance ID
        InstanceType:     attrs["cloud.instance.type"],   // e.g., “t3.medium”
        VPCID:            attrs["cloud.vpc.id"],          // AWS/GCP VPC ID
        SubnetID:         attrs["cloud.subnet.id"],       // AWS/GCP/Azure subnet ID
        ImageID:          attrs["cloud.image.id"],        // AMI ID or equivalent
        ServiceID:        attrs["cloud.service.id"],      // managed service ID, if any

        // Kubernetes / Container context
        ContainerID:        attrs["container.id"],               // container runtime ID
        ContainerName:      attrs["container.name"],             // e.g., Docker container name
        ContainerImageID:   attrs["container.image.id"],         // full image digest/ID
        ContainerImageName: attrs["container.image.name"],       // e.g., “nginx:1.21.3”
        PodName:            attrs["k8s.pod.name"],               // Kubernetes Pod name
        PodUID:             attrs["k8s.pod.uid"],                // Pod UID
        Namespace:          attrs["k8s.namespace.name"],         // K8s namespace
        NamespaceUID:       attrs["k8s.namespace.uid"],          // Namespace UID
        DeploymentName:     attrs["k8s.deployment.name"],         // Deployment or higher‐level owner
        OwnerKind:          attrs["k8s.deployment.kind"],         // e.g., “Deployment”
        OwnerName:          attrs["k8s.deployment.name"],         // owner name (same as DeploymentName)
        ClusterName:        attrs["k8s.cluster.name"],           // cluster identifier
        ClusterUID:         attrs["k8s.cluster.uid"],            // cluster UID if present
        NodeName:           attrs["k8s.node.name"],              // Kubernetes node name
        ServiceAccount:     attrs["k8s.service_account.name"],    // Pod’s service account

        // Application / OTel Resource / Service context
        ServiceName:               attrs["service.name"],               // OTLP: service.name
        ServiceNamespace:          attrs["service.namespace"],          // OTLP: service.namespace
        ServiceInstanceID:         attrs["service.instance.id"],        // OTLP: service.instance.id
        ServiceVersion:            attrs["service.version"],            // OTLP: service.version
        TelemetrySDKName:          attrs["telemetry.sdk.name"],         // OTLP: telemetry.sdk.name
        TelemetrySDKVersion:       attrs["telemetry.sdk.version"],      // OTLP: telemetry.sdk.version
        TelemetrySDKLanguage:      attrs["telemetry.sdk.language"],     // OTLP: telemetry.sdk.language
        InstrumentationLibrary:    attrs["otel.library.name"],          // OTLP: instrumentation library name
        InstrumentationLibVersion: attrs["otel.library.version"],       // OTLP: instrumentation library version

        // Process / Runtime context (if emitted as resource attrs)
        ProcessID:      0,                          // process.id as int, if needed
        ProcessName:    attrs["process.executable"], // e.g., “java”, “nginx”
        RuntimeName:    attrs["process.runtime.name"],    // e.g., “go”, “java”
        RuntimeVersion: attrs["process.runtime.version"], // e.g., “go1.20”

        // Networking / Mesh / Security context
        PublicIP:         attrs["net.host.ip"],             // public IP
        PrivateIP:        attrs["net.host.private_ip"],     // private IP
        MACAddress:       attrs["net.host.mac"],            // MAC address
        NetworkInterface: attrs["net.host.interface"],      // interface name (e.g., “eth0”)
        MeshPeerVersion:  attrs["mesh.peer.version"],       // service mesh proxy version
        MTLSEnabled:      attrs["security.tls.enabled"] == "true", // true/false
        TLSVersion:       attrs["security.tls.version"],    // e.g., “TLSv1.3”
        CipherSuite:      attrs["security.tls.cipher"],     // e.g., “TLS_AES_128_GCM_SHA256”
        AuthMethod:       attrs["rpc.auth.method"],         // e.g., “oauth2”, “mtls”
        User:             attrs["enduser.id"],              // end‐user ID if available

        // Deployment / CI-CD context
        DeploymentID:   attrs["deployment.id"],    // e.g., CI/CD pipeline ID or deployment UUID
        GitCommitHash:  attrs["git.commit"],       // git commit SHA if injected
        BuildTimestamp: attrs["build.time"],       // build timestamp (e.g., “2023-07-28T15:04:05Z”)

        // Log‐specific fields (included here if you rely on them for traces/logs)
        AppName:    attrs["app.name"],      // e.g., “nginx”, “orders-api”
        AppVersion: attrs["app.version"],   // e.g., “1.4.2”
        Unit:       attrs["systemd.unit"],  // journald unit name if used
        EventID:    attrs["event.id"],      // Windows event ID, etc.
        Executable: attrs["process.executable"], // path to binary
        Path:       attrs["log.file.path"], // file path if tailing logs
        Extra:      nil,                    // any custom fields you want to set

        // Custom / User‐defined tags & labels (you can merge all remaining attrs here)
        Labels: nil,
        Tags:   nil,
    }

    // Parse ProcessID if present
    if pidStr, ok := attrs["process.pid"]; ok {
        if pid, err := strconv.Atoi(pidStr); err == nil {
            meta.ProcessID = pid
        }
    }

    // Collect any Kubernetes pod labels prefixed with "k8s.pod.label."
    meta.PodLabels = make(map[string]string)
    for key, val := range attrs {
        if strings.HasPrefix(key, "k8s.pod.label.") {
            labelKey := strings.TrimPrefix(key, "k8s.pod.label.")
            meta.PodLabels[labelKey] = val
        }
    }

    // Collect any Kubernetes node labels prefixed with "k8s.node.label."
    meta.NodeLabels = make(map[string]string)
    for key, val := range attrs {
        if strings.HasPrefix(key, "k8s.node.label.") {
            labelKey := strings.TrimPrefix(key, "k8s.node.label.")
            meta.NodeLabels[labelKey] = val
        }
    }

    // If resource attributes include arbitrary user tags, add them under Tags
    meta.Labels = make(map[string]string)
    for key, val := range attrs {
        switch key {
        // Skip keys already consumed above
        case "service.name", "service.version", "k8s.pod.name", "k8s.pod.uid",
            "k8s.namespace.name", "k8s.namespace.uid", "k8s.cluster.name", "k8s.cluster.uid",
            "k8s.node.name", "host.id", "host.name", "cloud.provider", "cloud.region",
            "cloud.zone", "cloud.account.id", "cloud.instance.id", "cloud.instance.type",
            "container.id", "container.name", "container.image.name", "container.image.id",
            "deployment.id", "git.commit", "build.time", "process.pid", "process.executable",
            "process.runtime.name", "process.runtime.version", "net.host.ip", "net.host.mac",
            "security.tls.version", "security.tls.cipher", "rpc.auth.method", "enduser.id":
            // Already mapped; do nothing

        default:
            // Preserve as a “tag” for custom queries
            meta.Labels[key] = val
        }
    }

    return meta
}