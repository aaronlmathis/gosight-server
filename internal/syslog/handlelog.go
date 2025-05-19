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

package syslog

import (
	"context"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// handleLog parses raw syslog data, builds a LogPayload, and writes it to the buffer.
func (s *SyslogServer) handleLog(ctx context.Context, raw []byte, srcAddr string) {
	ts := time.Now()

	// Simple parser stub; replace with full RFC3164/5424/CEF logic
	host, facility, severity, msg, fields := parseSyslog(raw)

	// Lookup device by source address
	dev, _ := s.sys.Stores.Data.GetNetworkDeviceByAddress(ctx, srcAddr)
	deviceID, deviceName := "", ""
	if dev != nil {
		deviceID = dev.ID
		deviceName = dev.Name
	}

	// Build the LogEntry
	entry := model.LogEntry{
		Timestamp: ts,
		Level:     severity,
		Message:   msg,
		Source:    facility,
		Category:  "network",
		Fields:    fields,
		Tags: map[string]string{
			"device_id":   deviceID,
			"device_name": deviceName,
			"src_ip":      srcAddr,
		},
		Meta: &model.LogMeta{
			Platform: "syslog",
			Service:  deviceName,
		},
	}

	// Wrap in a LogPayload
	payload := model.LogPayload{
		AgentID:    "",
		HostID:     "",
		Hostname:   host,
		EndpointID: deviceID,
		Timestamp:  ts,
		Logs:       []model.LogEntry{entry},
		Meta:       nil,
	}

	// Write into the shared log buffer
	s.sys.Buffers.Logs.WriteAny(payload)
}

// parseSyslog is a placeholder parser for raw syslog messages.
func parseSyslog(raw []byte) (host, facility, severity, message string, fields map[string]string) {
	fields = make(map[string]string)
	message = strings.TrimSpace(string(raw))
	host = ""
	facility = "syslog"
	severity = "info"
	return
}
