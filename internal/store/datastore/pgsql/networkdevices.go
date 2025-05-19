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

package pgstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
)

// GetNetworkDevices returns all network devices from the database
func (s *PGDataStore) GetNetworkDevices(ctx context.Context) ([]*model.NetworkDevice, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, created_at, updated_at
		FROM network_devices
	`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var networkDevices []*model.NetworkDevice

	for rows.Next() {
		var networkDevice model.NetworkDevice
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&networkDevice.ID,
			&networkDevice.Name,
			&networkDevice.Vendor,
			&networkDevice.Address,
			&networkDevice.Port,
			&networkDevice.Protocol,
			&networkDevice.Format,
			&networkDevice.Facility,
			&networkDevice.SyslogID,
			&networkDevice.RateLimit,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		networkDevice.CreatedAt = createdAt
		networkDevice.UpdatedAt = updatedAt
		networkDevices = append(networkDevices, &networkDevice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return networkDevices, nil
}

// GetNetworkDeviceByAddress looks up a single NetworkDevice by its IP/hostname.
// Returns (nil, nil) if no matching device is found.
func (s *PGDataStore) GetNetworkDeviceByAddress(ctx context.Context, address string) (*model.NetworkDevice, error) {
	const q = `
        SELECT id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, created_at, updated_at
        FROM network_devices
        WHERE address = $1
        LIMIT 1
    `

	row := s.db.QueryRowContext(ctx, q, address)

	var nd model.NetworkDevice
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&nd.ID,
		&nd.Name,
		&nd.Vendor,
		&nd.Address,
		&nd.Port,
		&nd.Protocol,
		&nd.Format,
		&nd.Facility,
		&nd.SyslogID,
		&nd.RateLimit,
		&createdAt,
		&updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetNetworkDeviceByAddress scan failed: %w", err)
	}

	nd.CreatedAt = createdAt
	nd.UpdatedAt = updatedAt
	return &nd, nil
}
