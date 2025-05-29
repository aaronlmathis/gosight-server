package telemetry

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/cache/resourcecache"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

type ResourceDiscovery struct {
	cache resourcecache.ResourceCache
}

func NewResourceDiscovery(cache resourcecache.ResourceCache) *ResourceDiscovery {
	return &ResourceDiscovery{
		cache: cache,
	}
}

func (rd *ResourceDiscovery) ProcessMetricPayload(payload *model.MetricPayload) *model.MetricPayload {
	if payload == nil || payload.Meta == nil {
		return payload
	}

	// Extract resource information from metric metadata
	ctx := context.Background()
	resource, err := rd.extractResourceFromMeta(ctx, payload.Meta)
	if err != nil {
		utils.Error("Failed to extract resource from metric payload: %v", err)
		return payload
	}

	if resource != nil {
		resource.LastSeen = time.Now()
		if err := rd.upsertResource(ctx, resource); err != nil {
			utils.Error("Failed to upsert resource: %v", err)
			return payload
		}

		// Enrich the payload with ResourceID and resource data from cache
		if payload.Meta == nil {
			payload.Meta = &model.Meta{}
		}
		payload.Meta.ResourceID = resource.ID

		// Enrich payload with existing resource data from cache
		if cachedResource, exists := rd.cache.GetResource(resource.ID); exists {
			rd.enrichPayloadFromResource(payload, cachedResource)
		}
	}

	return payload
}

func (rd *ResourceDiscovery) ProcessLogPayload(payload *model.LogPayload) *model.LogPayload {
	if payload == nil || payload.Meta == nil {
		return payload
	}

	// Extract resource information from log metadata
	ctx := context.Background()
	resource, err := rd.extractResourceFromMeta(ctx, payload.Meta)
	if err != nil {
		utils.Error("Failed to extract resource from log payload: %v", err)
		return payload
	}
	if resource != nil {
		resource.LastSeen = time.Now()
		if err := rd.upsertResource(ctx, resource); err != nil {
			utils.Error("Failed to upsert resource: %v", err)
			return payload
		}

		// Enrich the payload with ResourceID
		if payload.Meta == nil {
			payload.Meta = &model.Meta{}
		}
		payload.Meta.ResourceID = resource.ID

		// Enrich payload with existing resource data from cache
		if cachedResource, exists := rd.cache.GetResource(resource.ID); exists {
			rd.enrichLogPayloadFromResource(payload, cachedResource)
		}
	}

	return payload
}

func (rd *ResourceDiscovery) ProcessTracePayload(payload *model.TracePayload) *model.TracePayload {
	if payload == nil || payload.Meta == nil {
		return payload
	}

	// Extract resource information from trace metadata
	ctx := context.Background()
	resource, err := rd.extractResourceFromMeta(ctx, payload.Meta)
	if err != nil {
		utils.Error("Failed to extract resource from trace payload: %v", err)
		return payload
	}

	if resource != nil {
		resource.LastSeen = time.Now()
		if err := rd.upsertResource(ctx, resource); err != nil {
			utils.Error("Failed to upsert resource: %v", err)
			return payload
		}

		// Enrich the payload with ResourceID
		if payload.Meta == nil {
			payload.Meta = &model.Meta{}
		}
		payload.Meta.ResourceID = resource.ID

		// Enrich payload with existing resource data from cache
		if cachedResource, exists := rd.cache.GetResource(resource.ID); exists {
			rd.enrichTracePayloadFromResource(payload, cachedResource)
		}
	}

	return payload
}

func (rd *ResourceDiscovery) ProcessProcessPayload(payload *model.ProcessPayload) *model.ProcessPayload {
	if payload == nil || payload.Meta == nil {
		return payload
	}

	// Extract resource information from process metadata
	ctx := context.Background()
	resource, err := rd.extractResourceFromMeta(ctx, payload.Meta)
	if err != nil {
		utils.Error("Failed to extract resource from process payload: %v", err)
		return payload
	}

	if resource != nil {
		resource.LastSeen = time.Now()
		if err := rd.upsertResource(ctx, resource); err != nil {
			utils.Error("Failed to upsert resource: %v", err)
			return payload
		}

		// Enrich the payload with ResourceID
		if payload.Meta == nil {
			payload.Meta = &model.Meta{}
		}
		payload.Meta.ResourceID = resource.ID

		// Enrich payload with existing resource data from cache
		if cachedResource, exists := rd.cache.GetResource(resource.ID); exists {
			rd.enrichProcessPayloadFromResource(payload, cachedResource)
		}
	}

	return payload
}

func (rd *ResourceDiscovery) extractResourceFromMeta(ctx context.Context, meta *model.Meta) (*model.Resource, error) {
	if meta == nil {
		return nil, nil
	}

	// Determine kind first
	kind := meta.Kind
	if kind == "" {
		kind = rd.determineKind(meta)
		meta.Kind = kind // Update meta with determined kind
	}

	// Generate consistent resource ID using proper logic
	labels := rd.buildResourceLabels(meta, kind)
	resourceID := rd.generateResourceID(kind, labels)

	// Check if resource already exists in cache
	if existing, exists := rd.cache.GetResource(resourceID); exists {
		// Update last seen and any changed metadata
		rd.updateExistingResource(existing, meta)
		return existing, nil
	}

	// Create new resource
	resource := &model.Resource{
		ID:          resourceID,
		Kind:        kind, // Use determined kind
		Name:        rd.determineName(meta),
		DisplayName: rd.determineDisplayName(meta),
		Group:       rd.determineGroup(meta),
		Status:      model.ResourceStatusOnline,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Labels:      rd.buildResourceLabels(meta, kind), // Use proper labels
		Tags:        make(map[string]string),
		Annotations: make(map[string]string),
	}

	// Set parent relationship using the proper parent logic
	parentID := rd.determineParentID(meta)
	if parentID != "" {
		resource.ParentID = parentID
		// Ensure parent resource exists
		if err := rd.ensureParentResource(ctx, meta, parentID); err != nil {
			utils.Warn("Failed to ensure parent resource %s: %v", parentID, err)
		}
	}

	// Set additional fields from meta
	rd.populateResourceFromMeta(resource, meta)

	return resource, nil
}

func (rd *ResourceDiscovery) determineKind(meta *model.Meta) string {
	// Use explicit kind from agent if available
	if meta.Kind != "" {
		return meta.Kind
	}

	// Check if this is an agent itself
	if meta.AgentID != "" && meta.ContainerID == "" && meta.HostID == "" {
		return model.ResourceKindAgent
	}

	// Fallback to legacy detection logic
	if meta.ContainerID != "" {
		return model.ResourceKindContainer
	}
	if meta.HostID != "" || meta.Hostname != "" {
		return model.ResourceKindHost
	}
	return model.ResourceKindApp
}

// upsertResource handles creating or updating a resource using cache-only approach.
// The ResourceCache automatically handles persistence through background flushing.
func (rd *ResourceDiscovery) upsertResource(ctx context.Context, resource *model.Resource) error {
	// Use cache-only approach - ResourceCache handles persistence automatically
	rd.cache.UpsertResource(resource)
	return nil
}

// updateExistingResource updates an existing resource with new metadata
func (rd *ResourceDiscovery) updateExistingResource(existing *model.Resource, meta *model.Meta) {
	existing.LastSeen = time.Now()
	existing.UpdatedAt = time.Now()

	// Update fields that might have changed
	if meta.Environment != "" && existing.Environment != meta.Environment {
		existing.Environment = meta.Environment
	}
	if meta.IPAddress != "" && existing.IPAddress != meta.IPAddress {
		existing.IPAddress = meta.IPAddress
	}

	// Labels should be immutable - only update if kind changed
	if existing.Kind != meta.Kind {
		if existing.Labels == nil {
			existing.Labels = make(map[string]string)
		}
		for k, v := range rd.buildResourceLabels(meta, meta.Kind) {
			existing.Labels[k] = v
		}
	}

	// Update status to online if it was offline
	if existing.Status != model.ResourceStatusOnline {
		existing.Status = model.ResourceStatusOnline
	}
}

// determineName generates a name for the resource based on metadata
func (rd *ResourceDiscovery) determineName(meta *model.Meta) string {
	// Generate name based on kind and available metadata
	switch meta.Kind {
	case model.ResourceKindContainer:
		if meta.ContainerName != "" {
			return meta.ContainerName
		}
		if meta.ContainerID != "" {
			// Use short container ID (first 12 chars)
			if len(meta.ContainerID) > 12 {
				return meta.ContainerID[:12]
			}
			return meta.ContainerID
		}
		return "unknown-container"

	case model.ResourceKindHost:
		if meta.Hostname != "" {
			return meta.Hostname
		}
		if meta.HostID != "" {
			return meta.HostID
		}
		if meta.AgentID != "" {
			return meta.AgentID
		}
		return "unknown-host"

	default:
		if meta.Hostname != "" {
			return meta.Hostname
		}
		return "unknown-resource"
	}
}

// determineDisplayName generates a display name for the resource
func (rd *ResourceDiscovery) determineDisplayName(meta *model.Meta) string {
	// Generate display name based on kind and available metadata
	switch meta.Kind {
	case model.ResourceKindContainer:
		if meta.ContainerName != "" {
			if meta.Hostname != "" {
				return fmt.Sprintf("%s on %s", meta.ContainerName, meta.Hostname)
			}
			return meta.ContainerName
		}
		if meta.ContainerID != "" {
			containerID := meta.ContainerID
			if len(containerID) > 12 {
				containerID = containerID[:12]
			}
			if meta.Hostname != "" {
				return fmt.Sprintf("Container %s on %s", containerID, meta.Hostname)
			}
			return fmt.Sprintf("Container %s", containerID)
		}
		return "Unknown Container"

	case model.ResourceKindHost:
		if meta.Hostname != "" {
			return meta.Hostname
		}
		if meta.HostID != "" {
			return fmt.Sprintf("Host %s", meta.HostID)
		}
		if meta.AgentID != "" {
			return fmt.Sprintf("Host %s", meta.AgentID)
		}
		return "Unknown Host"

	default:
		return rd.determineName(meta)
	}
}

// determineGroup determines the resource group based on metadata
func (rd *ResourceDiscovery) determineGroup(meta *model.Meta) string {
	// Use environment as group if available
	if meta.Environment != "" {
		return meta.Environment
	}

	// Default group based on kind
	switch meta.Kind {
	case model.ResourceKindContainer:
		return "containers"
	case model.ResourceKindHost:
		return "hosts"
	default:
		return "default"
	}
}

// populateResourceFromMeta populates resource fields from metadata
func (rd *ResourceDiscovery) populateResourceFromMeta(resource *model.Resource, meta *model.Meta) {
	// Set basic fields
	if meta.IPAddress != "" {
		resource.IPAddress = meta.IPAddress
	}
	if meta.Environment != "" {
		resource.Environment = meta.Environment
	}
	if meta.OS != "" {
		resource.OS = meta.OS
	}
	if meta.Platform != "" {
		resource.Platform = meta.Platform
	}
	if meta.Architecture != "" {
		resource.Arch = meta.Architecture
	}

	// Set container-specific fields
	if meta.Kind == model.ResourceKindContainer {
		if meta.ContainerImageName != "" {
			resource.Annotations["container_image"] = meta.ContainerImageName
		}
		if meta.ContainerImageID != "" {
			resource.Annotations["container_image_id"] = meta.ContainerImageID
		}
	}

	// Set host-specific fields
	if meta.Kind == model.ResourceKindHost {
		if meta.PlatformFamily != "" {
			resource.Annotations["platform_family"] = meta.PlatformFamily
		}
		if meta.PlatformVersion != "" {
			resource.Annotations["platform_version"] = meta.PlatformVersion
		}
		if meta.VirtualizationSystem != "" {
			resource.Annotations["virtualization_system"] = meta.VirtualizationSystem
		}
		if meta.VirtualizationRole != "" {
			resource.Annotations["virtualization_role"] = meta.VirtualizationRole
		}
	}

	// Classify meta.Labels into proper Labels vs Tags based on their origin/purpose
	if meta.Labels != nil {
		// System-generated metadata that should be Labels (not user-editable)
		systemLabels := []string{
			"namespace", "subnamespace", "job", "instance", "agent_start_time",
			"container_id", "device_id", "runtime", "source", "type",
		}

		labelCount := 0
		tagCount := 0

		// User/operational metadata that should remain as Tags
		for k, v := range meta.Labels {
			isSystemLabel := false
			for _, sysLabel := range systemLabels {
				if k == sysLabel {
					isSystemLabel = true
					break
				}
			}

			if isSystemLabel {
				// Add to Labels (system-generated, immutable)
				if resource.Labels == nil {
					resource.Labels = make(map[string]string)
				}
				resource.Labels[k] = v
				labelCount++
			} else {
				// Add to Tags (user-defined, mutable)
				if resource.Tags == nil {
					resource.Tags = make(map[string]string)
				}
				resource.Tags[k] = v
				tagCount++
			}
		}

		utils.Debug("Resource %s: classified meta.Labels - %d -> Labels, %d -> Tags", resource.ID, labelCount, tagCount)
	}
}

func (rd *ResourceDiscovery) determineParentID(meta *model.Meta) string {
	kind := meta.Kind
	if kind == "" {
		kind = rd.determineKind(meta)
	}

	switch kind {
	case model.ResourceKindContainer:
		// Container parent is host
		parentLabels := rd.buildResourceLabels(meta, model.ResourceKindHost)
		if len(parentLabels) > 0 {
			return rd.generateResourceID(model.ResourceKindHost, parentLabels)
		}

	case model.ResourceKindHost:
		// Host parent is agent
		parentLabels := rd.buildResourceLabels(meta, model.ResourceKindAgent)
		if len(parentLabels) > 0 {
			return rd.generateResourceID(model.ResourceKindAgent, parentLabels)
		}

	case model.ResourceKindApp:
		// App parent can be either container or host
		if meta.ContainerID != "" {
			// Running in container
			parentLabels := rd.buildResourceLabels(meta, model.ResourceKindContainer)
			if len(parentLabels) > 0 {
				return rd.generateResourceID(model.ResourceKindContainer, parentLabels)
			}
		} else {
			// Running on host
			parentLabels := rd.buildResourceLabels(meta, model.ResourceKindHost)
			if len(parentLabels) > 0 {
				return rd.generateResourceID(model.ResourceKindHost, parentLabels)
			}
		}
	}

	return "" // No parent for agents and syslog
}

// determineRuntime determines the runtime from metadata
func (rd *ResourceDiscovery) determineRuntime(meta *model.Meta) string {
	// Check tags first
	if runtime := meta.Tags["runtime"]; runtime != "" {
		return runtime
	}

	// Check for container runtimes based on container ID format
	if meta.ContainerID != "" {
		if strings.HasPrefix(meta.ContainerID, "docker://") {
			return "docker"
		}
		if strings.HasPrefix(meta.ContainerID, "containerd://") {
			return "containerd"
		}
		if strings.HasPrefix(meta.ContainerID, "cri-o://") {
			return "cri-o"
		}
		// Default to docker if we can't determine
		return "docker"
	}

	return ""
}

// ensureParentResource ensures the parent resource exists
func (rd *ResourceDiscovery) ensureParentResource(ctx context.Context, meta *model.Meta, parentID string) error {
	// Check if parent already exists in cache
	if _, exists := rd.cache.GetResource(parentID); exists {
		return nil
	}

	// Create parent resource based on the parent ID and current metadata
	return rd.createParentResource(ctx, meta, parentID)
}

// createParentResource creates a parent resource from available metadata
func (rd *ResourceDiscovery) createParentResource(ctx context.Context, meta *model.Meta, parentID string) error {
	// Determine parent kind from ID
	var parentKind string
	if strings.HasPrefix(parentID, "host-") {
		parentKind = model.ResourceKindHost
	} else if strings.HasPrefix(parentID, "agent-") {
		parentKind = model.ResourceKindAgent
	} else {
		return fmt.Errorf("unknown parent resource type for ID: %s", parentID)
	}

	// Create parent meta based on current meta
	parentMeta := &model.Meta{
		Kind:         parentKind,
		AgentID:      meta.AgentID,
		HostID:       meta.HostID,
		Hostname:     meta.Hostname,
		IPAddress:    meta.IPAddress,
		OS:           meta.OS,
		Platform:     meta.Platform,
		Architecture: meta.Architecture,
		Environment:  meta.Environment,
		Tags:         meta.Tags,
	}

	// Create parent resource
	parentResource := &model.Resource{
		ID:          parentID,
		Kind:        parentKind,
		Name:        rd.determineName(parentMeta),
		DisplayName: rd.determineDisplayName(parentMeta),
		Group:       rd.determineGroup(parentMeta),
		Status:      model.ResourceStatusOnline,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Labels:      rd.buildResourceLabels(parentMeta, parentKind),
		Tags:        make(map[string]string),
		Annotations: make(map[string]string),
	}

	// Set parent's parent if needed
	grandParentID := rd.determineParentID(parentMeta)
	if grandParentID != "" {
		parentResource.ParentID = grandParentID
		// Recursively ensure grandparent exists
		if err := rd.ensureParentResource(ctx, parentMeta, grandParentID); err != nil {
			utils.Warn("Failed to ensure grandparent resource %s: %v", grandParentID, err)
		}
	}

	rd.populateResourceFromMeta(parentResource, parentMeta)
	return rd.upsertResource(ctx, parentResource)
}

// rd.generateResourceID creates a deterministic ID for a resource kind using a fingerprint of its identifying labels.
// The resulting ID is in the form <kind>-<shortsha1>.
func (rd *ResourceDiscovery) generateResourceID(kind string, labels map[string]string) string {
	if kind == "" {
		return ""
	}

	// Sort keys for consistent fingerprint
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build canonical fingerprint string: key=value|key=value...
	var sb strings.Builder
	sb.WriteString(kind)
	sb.WriteString(":")
	for _, k := range keys {
		v := labels[k]
		if v != "" {
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sb.WriteString("|")
		}
	}
	fingerprint := sb.String()

	// Hash fingerprint
	hash := sha1.Sum([]byte(fingerprint))
	shortHash := hex.EncodeToString(hash[:6]) // 12 hex chars (~48-bit)

	return fmt.Sprintf("%s-%s", kind, shortHash)
}

func (rd *ResourceDiscovery) determineResourceStatus(meta *model.Meta) string {
	// For containers, check explicit status from tags/metadata
	if meta.Kind == model.ResourceKindContainer {
		if status := meta.Tags["container_status"]; status != "" {
			switch strings.ToLower(status) {
			case "running":
				return model.ResourceStatusOnline
			case "exited", "stopped", "dead":
				return model.ResourceStatusOffline
			case "paused":
				return model.ResourceStatusIdle
			default:
				return model.ResourceStatusUnknown
			}
		}
	}

	// Default to online for actively sending telemetry
	// LastSeen will be updated separately to track heartbeat
	return model.ResourceStatusOnline
}

// Add the missing buildResourceLabels function per your specification
func (rd *ResourceDiscovery) buildResourceLabels(meta *model.Meta, kind string) map[string]string {
	labels := make(map[string]string)

	switch kind {
	case model.ResourceKindAgent:
		// Agent: use just agent_id
		if meta.AgentID != "" {
			labels["agent_id"] = meta.AgentID
		}

	case model.ResourceKindHost:
		// Host: use host_id, agent_id
		if meta.HostID != "" {
			labels["host_id"] = meta.HostID
		}
		if meta.AgentID != "" {
			labels["agent_id"] = meta.AgentID
		}

	case model.ResourceKindContainer:
		// Container: use container_id, runtime
		if meta.ContainerID != "" {
			labels["container_id"] = meta.ContainerID
		}
		runtime := rd.determineRuntime(meta)
		if runtime != "" {
			labels["runtime"] = runtime
		}

	case model.ResourceKindApp:
		// App: complex logic - need to determine parent type
		if meta.Service != "" {
			labels["service"] = meta.Service
		}
		if meta.Application != "" {
			labels["application"] = meta.Application
		}
		// If no service/app identifiers, use hostname as fallback
		if len(labels) == 0 && meta.Hostname != "" {
			labels["hostname"] = meta.Hostname
		}

	case model.ResourceKindSyslog:
		// Syslog: hostname + device_id (if present), or mac_address, or ip, or just hostname
		if meta.Hostname != "" {
			labels["hostname"] = meta.Hostname
		}
		if deviceID := meta.Labels["device_id"]; deviceID != "" {
			labels["device_id"] = deviceID
		} else if meta.MACAddress != "" {
			labels["mac_address"] = meta.MACAddress
		} else if meta.IPAddress != "" {
			labels["ip_address"] = meta.IPAddress
		}
	}

	return labels
}

// enrichPayloadFromResource enriches a metric payload's metadata with information
// from a cached resource. This merges resource tags and labels into the payload
// metadata to provide comprehensive context for telemetry processing.
func (rd *ResourceDiscovery) enrichPayloadFromResource(payload *model.MetricPayload, resource *model.Resource) {
	if payload == nil || payload.Meta == nil || resource == nil {
		return
	}

	// Merge resource tags into payload meta tags (user-defined, mutable)
	if resource.Tags != nil {
		if payload.Meta.Tags == nil {
			payload.Meta.Tags = make(map[string]string)
		}
		utils.Debug("Enriching payload with resource tags: %v", resource.Tags)
		// Add resource tags to payload, but don't override existing ones
		for key, value := range resource.Tags {
			if _, exists := payload.Meta.Tags[key]; !exists {
				payload.Meta.Tags[key] = value
			}
		}
	}

	// Merge resource labels into payload meta labels (system-generated, immutable)
	if resource.Labels != nil {
		if payload.Meta.Labels == nil {
			payload.Meta.Labels = make(map[string]string)
		}
		utils.Debug("Enriching payload with resource labels: %v", resource.Labels)
		// Add resource labels to payload, but don't override existing ones
		for key, value := range resource.Labels {
			if _, exists := payload.Meta.Labels[key]; !exists {
				payload.Meta.Labels[key] = value
			}
		}
	}

	// Optionally enrich with some resource metadata if not already present
	if payload.Meta.Environment == "" && resource.Environment != "" {
		payload.Meta.Environment = resource.Environment
	}

	if payload.Meta.Version == "" && resource.Version != "" {
		payload.Meta.Version = resource.Version
	}
}

// enrichLogPayloadFromResource enriches a log payload's metadata with information
// from a cached resource. This merges resource tags and labels into the payload
// metadata to provide comprehensive context for log processing.
func (rd *ResourceDiscovery) enrichLogPayloadFromResource(payload *model.LogPayload, resource *model.Resource) {
	if payload == nil || payload.Meta == nil || resource == nil {
		return
	}

	// Merge resource tags into payload meta tags (user-defined, mutable)
	if resource.Tags != nil {
		if payload.Meta.Tags == nil {
			payload.Meta.Tags = make(map[string]string)
		}
		// Add resource tags to payload, but don't override existing ones
		for key, value := range resource.Tags {
			if _, exists := payload.Meta.Tags[key]; !exists {
				payload.Meta.Tags[key] = value
			}
		}
	}

	// Merge resource labels into payload meta labels (system-generated, immutable)
	if resource.Labels != nil {
		if payload.Meta.Labels == nil {
			payload.Meta.Labels = make(map[string]string)
		}
		// Add resource labels to payload, but don't override existing ones
		for key, value := range resource.Labels {
			if _, exists := payload.Meta.Labels[key]; !exists {
				payload.Meta.Labels[key] = value
			}
		}
	}

	// Optionally enrich with some resource metadata if not already present
	if payload.Meta.Environment == "" && resource.Environment != "" {
		payload.Meta.Environment = resource.Environment
	}

	if payload.Meta.Version == "" && resource.Version != "" {
		payload.Meta.Version = resource.Version
	}
}

// enrichTracePayloadFromResource enriches a trace payload's metadata with information
// from a cached resource. This merges resource tags and labels into the payload
// metadata to provide comprehensive context for trace processing.
func (rd *ResourceDiscovery) enrichTracePayloadFromResource(payload *model.TracePayload, resource *model.Resource) {
	if payload == nil || payload.Meta == nil || resource == nil {
		return
	}

	// Merge resource tags into payload meta tags (user-defined, mutable)
	if resource.Tags != nil {
		if payload.Meta.Tags == nil {
			payload.Meta.Tags = make(map[string]string)
		}
		// Add resource tags to payload, but don't override existing ones
		for key, value := range resource.Tags {
			if _, exists := payload.Meta.Tags[key]; !exists {
				payload.Meta.Tags[key] = value
			}
		}
	}

	// Merge resource labels into payload meta labels (system-generated, immutable)
	if resource.Labels != nil {
		if payload.Meta.Labels == nil {
			payload.Meta.Labels = make(map[string]string)
		}
		// Add resource labels to payload, but don't override existing ones
		for key, value := range resource.Labels {
			if _, exists := payload.Meta.Labels[key]; !exists {
				payload.Meta.Labels[key] = value
			}
		}
	}

	// Optionally enrich with some resource metadata if not already present
	if payload.Meta.Environment == "" && resource.Environment != "" {
		payload.Meta.Environment = resource.Environment
	}

	if payload.Meta.Version == "" && resource.Version != "" {
		payload.Meta.Version = resource.Version
	}
}

// enrichProcessPayloadFromResource enriches a process payload's metadata with information
// from a cached resource. This merges resource tags and labels into the payload
// metadata to provide comprehensive context for process monitoring.
func (rd *ResourceDiscovery) enrichProcessPayloadFromResource(payload *model.ProcessPayload, resource *model.Resource) {
	if payload == nil || payload.Meta == nil || resource == nil {
		return
	}

	// Merge resource tags into payload meta tags (user-defined, mutable)
	if resource.Tags != nil {
		if payload.Meta.Tags == nil {
			payload.Meta.Tags = make(map[string]string)
		}
		// Add resource tags to payload, but don't override existing ones
		for key, value := range resource.Tags {
			if _, exists := payload.Meta.Tags[key]; !exists {
				payload.Meta.Tags[key] = value
			}
		}
	}

	// Merge resource labels into payload meta labels (system-generated, immutable)
	if resource.Labels != nil {
		if payload.Meta.Labels == nil {
			payload.Meta.Labels = make(map[string]string)
		}
		// Add resource labels to payload, but don't override existing ones
		for key, value := range resource.Labels {
			if _, exists := payload.Meta.Labels[key]; !exists {
				payload.Meta.Labels[key] = value
			}
		}
	}

	// Optionally enrich with some resource metadata if not already present
	if payload.Meta.Environment == "" && resource.Environment != "" {
		payload.Meta.Environment = resource.Environment
	}

	if payload.Meta.Version == "" && resource.Version != "" {
		payload.Meta.Version = resource.Version
	}
}
