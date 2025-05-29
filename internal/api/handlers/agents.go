// File: server/internal/api/handlers/agents.go
// Description: This file contains the agents and containers handlers for the GoSight server.

package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
)

// AgentsHandler provides handlers for agents and containers API endpoints
type AgentsHandler struct {
	Sys *sys.SystemContext
}

// NewAgentsHandler creates a new AgentsHandler
func NewAgentsHandler(sys *sys.SystemContext) *AgentsHandler {
	return &AgentsHandler{
		Sys: sys,
	}
}

// HandleAgentsAPI returns a JSON list of active agents
func (h *AgentsHandler) HandleAgentsAPI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get stored agents from database
	storedAgents, err := h.Sys.Stores.Data.ListAgents(ctx)
	if err != nil {
		http.Error(w, "failed to load agents", http.StatusInternalServerError)
		return
	}

	if storedAgents == nil {
		storedAgents = []*model.Agent{}
	}

	// 2. Apply live status overlay
	liveMap := h.Sys.Tracker.GetAgentMap()

	// 3. Update status based on live data
	for i, agent := range storedAgents {
		if live, ok := liveMap[agent.AgentID]; ok {
			storedAgents[i].Status = live.Status
			storedAgents[i].Since = live.Since
			storedAgents[i].UptimeSeconds = live.UptimeSeconds
			storedAgents[i].LastSeen = live.LastSeen
		} else {
			storedAgents[i].Status = "Offline"
			storedAgents[i].Since = "—"
		}
	}

	// 4. Sort by hostname
	sort.SliceStable(storedAgents, func(i, j int) bool {
		return storedAgents[i].Hostname < storedAgents[j].Hostname
	})

	// 5. Respond
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(storedAgents)
}

// ContainerMetrics represents container metrics
type ContainerMetrics struct {
	Host   string            `json:"host"`
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Status string            `json:"status"`
	CPU    *float64          `json:"cpu,omitempty"`
	Mem    *float64          `json:"mem,omitempty"`
	RX     *float64          `json:"rx,omitempty"`
	TX     *float64          `json:"tx,omitempty"`
	Uptime *float64          `json:"uptime,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Ports  string            `json:"ports,omitempty"`
}
