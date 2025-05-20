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
	"net/http"
	"strconv"	
	"encoding/json"
	"github.com/aaronlmathis/gosight-server/internal/contextutil"		
	gosightauth "github.com/aaronlmathis/gosight-server/internal/auth"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/aaronlmathis/gosight-shared/model"
)

type NetworkDeviceResponse struct {
	Devices    []*model.NetworkDevice `json:"devices"`
	HasMore    bool                  `json:"has_more"`
	Count      int                   `json:"count"`
	Offset     int                   `json:"offset"`
}

// HandleNetworkDevicesPage serves the network devices page.
func (s *HttpServer) HandleNetworkDevicesPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if forbidden, ok := ctx.Value("forbidden").(bool); ok && forbidden {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}		

	userID, ok := contextutil.GetUserID(ctx)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
	filter := parseNetworkDeviceFilterFromQuery(r)

	// Ensure a sane default and request one extra record to determine if there are more
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
		Name:        q.Get("name"),
		Limit:         parseInt("limit", 100),
		Order:         q.Get("order"),	
		Offset:        parseInt("offset", 0),
		Vendor:        q.Get("vendor"),
		Address:       q.Get("address"),
		Port:          parseInt("port", 0),
		Protocol:      q.Get("protocol"),
		Format:        q.Get("format"),
		Facility:      q.Get("facility"),
		SyslogID:      q.Get("syslog_id"),
	}

	return filter
}