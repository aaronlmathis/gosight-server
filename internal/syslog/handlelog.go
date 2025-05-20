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
	"fmt"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// handleLog parses raw syslog data, builds a LogPayload, and writes it to the buffer.
func (s *SyslogServer) handleLog(ctx context.Context, raw []byte, srcAddr string) {
	utils.Debug("Handling log from %s", srcAddr)
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
	meta := &model.Meta{
		Hostname:   host,
		EndpointID: deviceID,
		Platform:   "syslog",
		Service:    deviceName,
		IPAddress:  srcAddr,
	}
	// Wrap in a LogPayload
	payload := model.LogPayload{
		Hostname:   deviceName,
		EndpointID: deviceID,
		Timestamp:  ts,
		Logs:       []model.LogEntry{entry},
		Meta:       meta,
	}

	// Write into the shared log buffer
	s.sys.Buffers.Logs.WriteAny(payload)
}

// SyslogFormat represents the detected format of a syslog message
type SyslogFormat int

const (
	FormatUnknown SyslogFormat = iota
	FormatRFC3164
	FormatRFC5424
	FormatCEF
)

// parseSyslog parses raw syslog messages in RFC3164, RFC5424, or CEF format
func parseSyslog(raw []byte) (host, facility, severity, message string, fields map[string]string) {
	fields = make(map[string]string)
	msg := strings.TrimSpace(string(raw))

	// Default values
	host = ""
	facility = "syslog"
	severity = "info"
	message = msg

	// Detect format
	format := detectFormat(msg)

	switch format {
	case FormatRFC5424:
		return parseRFC5424(msg)
	case FormatRFC3164:
		return parseRFC3164(msg)
	case FormatCEF:
		return parseCEF(msg)
	}

	return
}

// detectFormat determines the syslog format based on message characteristics
func detectFormat(msg string) SyslogFormat {
	if strings.HasPrefix(msg, "<") {
		if strings.Contains(msg, "1 ") && strings.Count(msg, "|") == 0 { // RFC5424 has version "1 "
			return FormatRFC5424
		}
		return FormatRFC3164 // Assume RFC3164 if it starts with PRI but isn't RFC5424
	}
	if strings.Contains(msg, "CEF:") {
		return FormatCEF
	}
	return FormatUnknown
}

// parseRFC5424 parses a message in RFC5424 format
// Example: <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - MSG
func parseRFC5424(msg string) (host, facility, severity, message string, fields map[string]string) {
	fields = make(map[string]string)

	// Extract PRI
	priEnd := strings.Index(msg, ">")
	if priEnd == -1 || priEnd > 5 {
		return "", "syslog", "info", msg, fields
	}

	pri := msg[1:priEnd]
	priNum := 0
	fmt.Sscanf(pri, "%d", &priNum)

	facilityNum := priNum / 8
	severityNum := priNum % 8

	facility = getFacilityString(facilityNum)
	severity = getSeverityString(severityNum)

	// Split remaining parts
	parts := strings.SplitN(msg[priEnd+1:], " ", 7)
	if len(parts) < 7 {
		return "", facility, severity, msg, fields
	}

	// Extract components
	fields["version"] = parts[0]
	fields["timestamp"] = parts[1]
	host = parts[2]
	fields["app_name"] = parts[3]
	fields["proc_id"] = parts[4]
	fields["msg_id"] = parts[5]
	message = parts[6]

	return
}

// parseRFC3164 parses a message in RFC3164 (BSD) format
// Example: <13>Feb  5 17:32:18 10.0.0.99 myproc[10]: Use the BFG!
func parseRFC3164(msg string) (host, facility, severity, message string, fields map[string]string) {
	fields = make(map[string]string)

	// Extract PRI
	priEnd := strings.Index(msg, ">")
	if priEnd == -1 || priEnd > 5 {
		return "", "syslog", "info", msg, fields
	}

	pri := msg[1:priEnd]
	priNum := 0
	fmt.Sscanf(pri, "%d", &priNum)

	facilityNum := priNum / 8
	severityNum := priNum % 8

	facility = getFacilityString(facilityNum)
	severity = getSeverityString(severityNum)

	// Parse timestamp and hostname
	parts := strings.SplitN(msg[priEnd+1:], " ", 4)
	if len(parts) < 4 {
		return "", facility, severity, msg, fields
	}

	timestamp := strings.Join(parts[0:3], " ")
	fields["timestamp"] = timestamp

	// Extract hostname and message
	remainder := parts[3]
	hostEnd := strings.Index(remainder, " ")
	if hostEnd == -1 {
		return "", facility, severity, remainder, fields
	}

	host = remainder[:hostEnd]
	message = strings.TrimSpace(remainder[hostEnd+1:])

	// Extract process info if present
	if procStart := strings.Index(message, "["); procStart != -1 {
		if procEnd := strings.Index(message[procStart:], "]"); procEnd != -1 {
			fields["process"] = message[:procStart]
			fields["pid"] = message[procStart+1 : procStart+procEnd]
			message = strings.TrimSpace(message[procStart+procEnd+2:]) // +2 to skip "]: "
		}
	}

	return
}

// parseCEF parses a message in CEF format
// Example: CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
func parseCEF(msg string) (host, facility, severity, message string, fields map[string]string) {
	fields = make(map[string]string)

	// Check for CEF prefix
	cefStart := strings.Index(msg, "CEF:")
	if cefStart == -1 {
		return "", "syslog", "info", msg, fields
	}

	// Extract header if present (everything before CEF:)
	if cefStart > 0 {
		header := msg[:cefStart]
		// Try to parse header as RFC3164
		host, facility, severity, _, _ = parseRFC3164(header)
	}

	// Parse CEF parts
	parts := strings.Split(msg[cefStart+4:], "|")
	if len(parts) < 7 {
		return host, facility, severity, msg, fields
	}

	// Standard CEF fields
	fields["cef_version"] = parts[0]
	fields["device_vendor"] = parts[1]
	fields["device_product"] = parts[2]
	fields["device_version"] = parts[3]
	fields["signature_id"] = parts[4]
	fields["name"] = parts[5]
	fields["severity"] = parts[6]

	// Parse extension fields if present
	if len(parts) > 7 {
		message = parts[7]
		// Parse key-value pairs in extensions
		pairs := strings.Split(parts[7], " ")
		for _, pair := range pairs {
			if strings.Contains(pair, "=") {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					fields[kv[0]] = kv[1]
				}
			}
		}
	}

	return
}

// getFacilityString converts a facility number to its string representation
func getFacilityString(facility int) string {
	facilities := map[int]string{
		0:  "kern",
		1:  "user",
		2:  "mail",
		3:  "daemon",
		4:  "auth",
		5:  "syslog",
		6:  "lpr",
		7:  "news",
		8:  "uucp",
		9:  "cron",
		10: "authpriv",
		11: "ftp",
		12: "ntp",
		13: "security",
		14: "console",
		15: "mark",
		16: "local0",
		17: "local1",
		18: "local2",
		19: "local3",
		20: "local4",
		21: "local5",
		22: "local6",
		23: "local7",
	}
	if name, ok := facilities[facility]; ok {
		return name
	}
	return "unknown"
}

// getSeverityString converts a severity number to its string representation
func getSeverityString(severity int) string {
	severities := map[int]string{
		0: "emergency",
		1: "alert",
		2: "critical",
		3: "error",
		4: "warning",
		5: "notice",
		6: "info",
		7: "debug",
	}
	if name, ok := severities[severity]; ok {
		return name
	}
	return "unknown"
}
