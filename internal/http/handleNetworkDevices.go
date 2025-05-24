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

package httpserver

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

type NetworkDeviceResponse struct {
	Devices []*model.NetworkDevice `json:"devices"`
	HasMore bool                   `json:"has_more"`
	Count   int                    `json:"count"`
	Offset  int                    `json:"offset"`
}

// HandleNetworkDevicesPage serves the network devices page.
func (s *HttpServer) HandleNetworkDevicesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
		return
	}

	user, err := s.Sys.Stores.Users.GetUserWithPermissions(ctx, userID)
	if err != nil {
		utils.Error("Failed to load user %s: %v", userID, err)
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	permissions := gosightauth.FlattenPermissions(user.Roles)

	bc := make(map[string]string, 0)
	bc["Network Devices"] = ""

	pageData := s.Tmpl.BuildPageData(user, bc, nil, r.URL.Path, "Network Devices", nil, permissions)

	err = s.Tmpl.RenderTemplate(w, "layout_dashboard", "dashboard_network_devices", pageData)

	if err != nil {
		utils.Debug("Failed to render network devices page: %v", err)
		http.Error(w, "template error", 500)
	}
}

// HandleNetworkDevicesAPI handles the network devices API.
func (s *HttpServer) HandleNetworkDevicesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		filter := parseNetworkDeviceFilterFromQuery(r)
		if filter.Limit <= 0 || filter.Limit > 1000 {
			filter.Limit = 100
		}
		devices, err := s.Sys.Stores.Data.GetNetworkDevices(s.Sys.Ctx, filter)
		if err != nil {
			utils.Error("network device query failed: %v", err)
			http.Error(w, "network device query failed", http.StatusInternalServerError)
			return
		}
		resp := NetworkDeviceResponse{
			Devices: devices,
			Offset:  filter.Offset,
			Count:   len(devices),
			HasMore: len(devices) == filter.Limit,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if r.Method == http.MethodPost {
		var device model.NetworkDevice
		if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		sanitizeDevice(&device)

		// Always set port from config based on protocol
		if device.Protocol == "UDP" {
			device.Port = s.Sys.Cfg.SyslogCollection.UDPPort
		} else if device.Protocol == "TCP" {
			device.Port = s.Sys.Cfg.SyslogCollection.TCPPort
		} else {
			device.Port = 0 // or handle as invalid protocol below
		}

		if !isValidDevice(&device) {
			http.Error(w, "invalid device data", http.StatusBadRequest)
			return
		}
		if device.ID == "" {
			device.ID = utils.NewUUID()
		}
		if device.Status == "" {
			device.Status = "enabled"
		}

		err := s.Sys.Stores.Data.UpsertNetworkDevice(s.Sys.Ctx, &device)
		if err != nil {
			utils.Error("failed to upsert device: %v", err)
			http.Error(w, "failed to upsert device", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(device)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// HandleDeleteNetworkDeviceAPI handles DELETE /api/v1/network-devices/{id}
func (s *HttpServer) HandleDeleteNetworkDeviceAPI(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if id == "" {
		http.Error(w, "missing device id", http.StatusBadRequest)
		return
	}
	err := s.Sys.Stores.Data.DeleteNetworkDeviceByID(s.Sys.Ctx, id)
	if err != nil {
		utils.Error("failed to delete device: %v", err)
		http.Error(w, "failed to delete device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleToggleNetworkDeviceStatusAPI handles POST /api/v1/network-devices/{id}/toggle
func (s *HttpServer) HandleToggleNetworkDeviceStatusAPI(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[strings.LastIndex(strings.TrimSuffix(r.URL.Path, "/toggle"), "/")+1:]
	if id == "" {
		http.Error(w, "missing device id", http.StatusBadRequest)
		return
	}
	err := s.Sys.Stores.Data.ToggleNetworkDeviceStatus(s.Sys.Ctx, id)
	if err != nil {
		utils.Error("failed to toggle device status: %v", err)
		http.Error(w, "failed to toggle device status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HandleUpdateNetworkDeviceAPI handles PUT /api/v1/network-devices/{id}
func (s *HttpServer) HandleUpdateNetworkDeviceAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	if id == "" {
		http.Error(w, "missing device id", http.StatusBadRequest)
		return
	}
	var device model.NetworkDevice
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	device.ID = id // Ensure the ID from the URL is used
	sanitizeDevice(&device)

	// Always set port from config based on protocol
	if device.Protocol == "UDP" {
		device.Port = s.Sys.Cfg.SyslogCollection.UDPPort
	} else if device.Protocol == "TCP" {
		device.Port = s.Sys.Cfg.SyslogCollection.TCPPort
	} else {
		device.Port = 0
	}

	if !isValidDevice(&device) {
		http.Error(w, "invalid device data", http.StatusBadRequest)
		return
	}
	if device.Status == "" {
		device.Status = "enabled"
	}

	err := s.Sys.Stores.Data.UpsertNetworkDevice(s.Sys.Ctx, &device)
	if err != nil {
		utils.Error("failed to update device: %v", err)
		http.Error(w, "failed to update device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(device)
}

// parseNetworkDeviceQueryParams parses the query parameters from the HTTP request
func parseNetworkDeviceFilterFromQuery(r *http.Request) *model.NetworkDeviceFilter {
	q := r.URL.Query()

	parseInt := func(key string, def int) int {
		val := q.Get(key)
		if val == "" {
			return def
		}
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		return def
	}

	filter := &model.NetworkDeviceFilter{
		Name:     q.Get("name"),
		Limit:    parseInt("limit", 100),
		Order:    q.Get("order"),
		Offset:   parseInt("offset", 0),
		Vendor:   q.Get("vendor"),
		Address:  q.Get("address"),
		Port:     parseInt("port", 0),
		Protocol: q.Get("protocol"),
		Format:   q.Get("format"),
		Facility: q.Get("facility"),
		SyslogID: q.Get("syslog_id"),
	}

	return filter
}

func sanitizeString(s string) string {
	return strings.TrimSpace(s)
}

func isValidDevice(device *model.NetworkDevice) bool {
	// Example: Validate required fields
	if device.Name == "" || device.Address == "" || device.Protocol == "" {
		return false
	}
	// Example: Validate port range
	if device.Port < 1 || device.Port > 65535 {
		return false
	}
	// Example: Validate address (basic IPv4/hostname check)
	addrPattern := `^([a-zA-Z0-9\.\-]+|\d{1,3}(\.\d{1,3}){3})$`
	matched, _ := regexp.MatchString(addrPattern, device.Address)
	return matched
}

func sanitizeDevice(device *model.NetworkDevice) {
	device.Name = sanitizeString(device.Name)
	device.Vendor = sanitizeString(device.Vendor)
	device.Address = sanitizeString(device.Address)
	device.Protocol = sanitizeString(device.Protocol)
	device.Format = sanitizeString(device.Format)
	device.Facility = sanitizeString(device.Facility)
	device.SyslogID = sanitizeString(device.SyslogID)
	// Add more as needed
}
