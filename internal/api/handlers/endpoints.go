// File: server/internal/api/handlers/endpoints.go
// Description: This file contains endpoint management HTTP handlers for the GoSight server.

package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/gorilla/mux"
)

// EndpointsHandler handles endpoint management API requests
type EndpointsHandler struct {
	Sys *sys.SystemContext
}

// NewEndpointsHandler creates a new EndpointsHandler
func NewEndpointsHandler(sys *sys.SystemContext) *EndpointsHandler {
	return &EndpointsHandler{
		Sys: sys,
	}
}

// EndpointFilter represents filters for endpoint queries
type EndpointFilter struct {
	EndpointID  string
	Hostname    string
	Status      string
	HostID      string
	IP          string
	OS          string
	Arch        string
	Labels      map[string]string
	Tags        map[string]string
	LastSeenMin time.Duration
	LastSeenMax time.Duration
	Limit       int
	Sort        string
	Order       string
}

// HandleAPIEndpoints returns a JSON list of active endpoints
// It supports querying or listing all endpoints as well as /api/v1/endpoints/{endpoint_type} (hosts / containers)
func (h *EndpointsHandler) HandleAPIEndpoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := h.parseEndpointFilters(r)

	agents, err := h.Sys.Stores.Data.ListAgents(ctx)
	if err != nil {
		http.Error(w, "failed to load agents", http.StatusInternalServerError)
		return
	}
	containers, err := h.Sys.Stores.Data.ListContainers(ctx)
	if err != nil {
		http.Error(w, "failed to load containers", http.StatusInternalServerError)
		return
	}

	// Merge live status for agents
	liveMap := h.Sys.Tracker.GetAgentMap()
	for _, a := range agents {
		if live, ok := liveMap[a.AgentID]; ok {
			a.Status = live.Status
			a.Since = live.Since
			a.UptimeSeconds = live.UptimeSeconds
			a.LastSeen = live.LastSeen
		} else {
			a.Status = "Offline"
			a.Since = "—"
		}
	}

	// Apply filtering
	filteredAgents := h.filterAgents(agents, filter)
	filteredContainers := h.filterContainers(containers, filter)

	// Convert to generic structure
	var result []map[string]interface{}

	for _, a := range filteredAgents {
		result = append(result, map[string]interface{}{
			"type":      "host",
			"id":        a.EndpointID,
			"hostname":  a.Hostname,
			"status":    a.Status,
			"last_seen": a.LastSeen,
			"uptime":    a.UptimeSeconds,
			"agent_id":  a.AgentID,
			"host_id":   a.HostID,
			"labels":    a.Labels,
			"ip":        a.IP,
			"os":        a.OS,
			"arch":      a.Arch,
			"version":   a.Version,
		})
	}

	for _, c := range filteredContainers {
		result = append(result, map[string]interface{}{
			"type":         "container",
			"id":           c.EndpointID,
			"container_id": c.ContainerID,
			"name":         c.Name,
			"image":        c.ImageName,
			"image_id":     c.ImageID,
			"runtime":      c.Runtime,
			"status":       c.Status,
			"host_id":      c.HostID,
			"last_seen":    c.LastSeen,
			"labels":       c.Labels,
		})
	}

	// Sort by last_seen descending
	sort.SliceStable(result, func(i, j int) bool {
		t1, ok1 := result[i]["last_seen"].(time.Time)
		t2, ok2 := result[j]["last_seen"].(time.Time)
		return ok1 && ok2 && t1.After(t2)
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// HandleAPIEndpoint returns a specific endpoint by ID
func (h *EndpointsHandler) HandleAPIEndpoint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["id"]

	if endpointID == "" {
		http.Error(w, "endpoint ID required", http.StatusBadRequest)
		return
	}

	// Create a filter for this specific endpoint
	filter := EndpointFilter{
		EndpointID: endpointID,
		Limit:      1,
	}

	ctx := r.Context()
	agents, err := h.Sys.Stores.Data.ListAgents(ctx)
	if err != nil {
		http.Error(w, "failed to load agents", http.StatusInternalServerError)
		return
	}
	containers, err := h.Sys.Stores.Data.ListContainers(ctx)
	if err != nil {
		http.Error(w, "failed to load containers", http.StatusInternalServerError)
		return
	}

	// Merge live status for agents
	liveMap := h.Sys.Tracker.GetAgentMap()
	for _, a := range agents {
		if live, ok := liveMap[a.AgentID]; ok {
			a.Status = live.Status
			a.Since = live.Since
			a.UptimeSeconds = live.UptimeSeconds
			a.LastSeen = live.LastSeen
		} else {
			a.Status = "Offline"
			a.Since = "—"
		}
	}

	// Apply filtering
	filteredAgents := h.filterAgents(agents, filter)
	filteredContainers := h.filterContainers(containers, filter)

	// Check if we found the endpoint
	if len(filteredAgents) > 0 {
		a := filteredAgents[0]
		result := map[string]interface{}{
			"type":      "host",
			"id":        a.EndpointID,
			"hostname":  a.Hostname,
			"status":    a.Status,
			"last_seen": a.LastSeen,
			"uptime":    a.UptimeSeconds,
			"agent_id":  a.AgentID,
			"host_id":   a.HostID,
			"labels":    a.Labels,
			"ip":        a.IP,
			"os":        a.OS,
			"arch":      a.Arch,
			"version":   a.Version,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	if len(filteredContainers) > 0 {
		c := filteredContainers[0]
		result := map[string]interface{}{
			"type":         "container",
			"id":           c.EndpointID,
			"container_id": c.ContainerID,
			"name":         c.Name,
			"image":        c.ImageName,
			"image_id":     c.ImageID,
			"runtime":      c.Runtime,
			"status":       c.Status,
			"host_id":      c.HostID,
			"last_seen":    c.LastSeen,
			"labels":       c.Labels,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
		return
	}

	http.Error(w, "endpoint not found", http.StatusNotFound)
}

// HandleAPIEndpointsByType handles the /api/v1/endpoints/{endpointType} endpoint
func (h *EndpointsHandler) HandleAPIEndpointsByType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointType := strings.ToLower(vars["endpointType"])

	ctx := r.Context()
	filter := h.parseEndpointFilters(r)

	switch endpointType {
	case "hosts":
		agents, err := h.Sys.Stores.Data.ListAgents(ctx)
		if err != nil {
			http.Error(w, "failed to load agents", http.StatusInternalServerError)
			return
		}
		if agents == nil {
			agents = []*model.Agent{}
		}

		// Apply live status overlay
		liveMap := h.Sys.Tracker.GetAgentMap()
		for _, a := range agents {
			if live, ok := liveMap[a.AgentID]; ok {
				a.Status = live.Status
				a.Since = live.Since
				a.UptimeSeconds = live.UptimeSeconds
				a.LastSeen = live.LastSeen
			} else {
				a.Status = "Offline"
				a.Since = "—"
			}
		}

		filtered := h.filterAgents(agents, filter)
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Hostname < filtered[j].Hostname
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)

	case "containers":
		containers, err := h.Sys.Stores.Data.ListContainers(ctx)
		if err != nil {
			utils.Debug("Failed to load containers: %v", err)
			http.Error(w, "failed to load containers", http.StatusInternalServerError)
			return
		}
		if containers == nil {
			containers = []*model.Container{}
		}

		filtered := h.filterContainers(containers, filter)
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)

	default:
		http.Error(w, "invalid endpoint type", http.StatusBadRequest)
	}
}

// HandleAPIEndpointCreate creates a new endpoint (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointCreate(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint creation not yet implemented", http.StatusNotImplemented)
}

// HandleAPIEndpointUpdate updates an endpoint (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointUpdate(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint update not yet implemented", http.StatusNotImplemented)
}

// HandleAPIEndpointDelete deletes an endpoint (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointDelete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint deletion not yet implemented", http.StatusNotImplemented)
}

// HandleAPIEndpointTest tests an endpoint connection (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointTest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint testing not yet implemented", http.StatusNotImplemented)
}

// HandleAPIEndpointStatus gets endpoint status
func (h *EndpointsHandler) HandleAPIEndpointStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	endpointID := vars["id"]

	if endpointID == "" {
		http.Error(w, "endpoint ID required", http.StatusBadRequest)
		return
	}

	// Get live status from tracker
	liveMap := h.Sys.Tracker.GetAgentMap()

	// Try to find agent by endpoint ID
	for agentID, live := range liveMap {
		if live.EndpointID == endpointID {
			result := map[string]interface{}{
				"endpoint_id": endpointID,
				"agent_id":    agentID,
				"status":      live.Status,
				"last_seen":   live.LastSeen,
				"uptime":      live.UptimeSeconds,
				"since":       live.Since,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(result)
			return
		}
	}

	// If not found in live map, return offline status
	result := map[string]interface{}{
		"endpoint_id": endpointID,
		"status":      "Offline",
		"last_seen":   nil,
		"uptime":      0,
		"since":       "—",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// HandleAPIEndpointEnable enables an endpoint (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointEnable(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint enable not yet implemented", http.StatusNotImplemented)
}

// HandleAPIEndpointDisable disables an endpoint (placeholder)
func (h *EndpointsHandler) HandleAPIEndpointDisable(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "endpoint disable not yet implemented", http.StatusNotImplemented)
}

// parseEndpointFilters parses the query parameters from the request and returns an EndpointFilter struct
func (h *EndpointsHandler) parseEndpointFilters(r *http.Request) EndpointFilter {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	return EndpointFilter{
		EndpointID:  q.Get("endpointID"),
		Hostname:    q.Get("hostname"),
		Status:      q.Get("status"),
		HostID:      q.Get("hostID"),
		IP:          q.Get("ip"),
		OS:          q.Get("os"),
		Arch:        q.Get("arch"),
		Labels:      utils.ParseTagString(q.Get("labels")),
		Tags:        utils.ParseTagString(q.Get("tags")),
		LastSeenMin: h.parseDuration(q.Get("lastSeenMin")),
		LastSeenMax: h.parseDuration(q.Get("lastSeenMax")),
		Limit:       limit,
		Sort:        q.Get("sort"),
		Order:       strings.ToLower(q.Get("order")),
	}
}

// parseDuration safely parses a duration string
func (h *EndpointsHandler) parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// filterAgents filters a list of agents based on the provided filter criteria
func (h *EndpointsHandler) filterAgents(list []*model.Agent, filter EndpointFilter) []*model.Agent {
	var out []*model.Agent
	now := time.Now()

	for _, a := range list {
		if filter.EndpointID != "" && a.EndpointID != filter.EndpointID {
			continue
		}
		if filter.Hostname != "" && a.Hostname != filter.Hostname {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(a.Status, filter.Status) {
			continue
		}
		if filter.HostID != "" && a.HostID != filter.HostID {
			continue
		}
		if filter.IP != "" && a.IP != filter.IP {
			continue
		}
		if filter.OS != "" && a.OS != filter.OS {
			continue
		}
		if filter.Arch != "" && a.Arch != filter.Arch {
			continue
		}
		if !utils.MatchAllLabels(filter.Labels, a.Labels) {
			continue
		}

		age := now.Sub(a.LastSeen)
		if filter.LastSeenMin > 0 && age < filter.LastSeenMin {
			continue
		}
		if filter.LastSeenMax > 0 && age > filter.LastSeenMax {
			continue
		}
		out = append(out, a)
	}

	return h.sortAndLimitAgents(out, filter.Sort, filter.Order, filter.Limit)
}

// filterContainers filters a list of containers based on the provided filter criteria
func (h *EndpointsHandler) filterContainers(list []*model.Container, filter EndpointFilter) []*model.Container {
	var out []*model.Container
	now := time.Now()

	for _, c := range list {
		if filter.EndpointID != "" && c.EndpointID != filter.EndpointID {
			continue
		}
		if filter.Hostname != "" && c.Name != filter.Hostname {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(c.Status, filter.Status) {
			continue
		}
		if filter.HostID != "" && c.HostID != filter.HostID {
			continue
		}
		if !utils.MatchAllLabels(filter.Labels, c.Labels) {
			continue
		}
		age := now.Sub(c.LastSeen)
		if filter.LastSeenMin > 0 && age < filter.LastSeenMin {
			continue
		}
		if filter.LastSeenMax > 0 && age > filter.LastSeenMax {
			continue
		}
		out = append(out, c)
	}

	return h.sortAndLimitContainers(out, filter.Sort, filter.Order, filter.Limit)
}

// sortAndLimitAgents sorts and limits the agents list
func (h *EndpointsHandler) sortAndLimitAgents(list []*model.Agent, sortField, order string, limit int) []*model.Agent {
	if sortField != "" {
		switch sortField {
		case "hostname":
			sort.SliceStable(list, func(i, j int) bool {
				if order == "desc" {
					return list[i].Hostname > list[j].Hostname
				}
				return list[i].Hostname < list[j].Hostname
			})
		case "last_seen":
			sort.SliceStable(list, func(i, j int) bool {
				if order == "desc" {
					return list[i].LastSeen.After(list[j].LastSeen)
				}
				return list[i].LastSeen.Before(list[j].LastSeen)
			})
		}
	}

	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}

// sortAndLimitContainers sorts and limits the containers list
func (h *EndpointsHandler) sortAndLimitContainers(list []*model.Container, sortField, order string, limit int) []*model.Container {
	if sortField != "" {
		switch sortField {
		case "name":
			sort.SliceStable(list, func(i, j int) bool {
				if order == "desc" {
					return list[i].Name > list[j].Name
				}
				return list[i].Name < list[j].Name
			})
		case "last_seen":
			sort.SliceStable(list, func(i, j int) bool {
				if order == "desc" {
					return list[i].LastSeen.After(list[j].LastSeen)
				}
				return list[i].LastSeen.Before(list[j].LastSeen)
			})
		}
	}

	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list
}
