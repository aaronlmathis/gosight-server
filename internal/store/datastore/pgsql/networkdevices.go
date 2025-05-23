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
func (s *PGDataStore) GetAllNetworkDevices(ctx context.Context) ([]*model.NetworkDevice, error) {
	rows, err := s.db.QueryContext(ctx, `
	SELECT id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, status, created_at, updated_at
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
			&networkDevice.Status, // added
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

func (s *PGDataStore) GetNetworkDevices(ctx context.Context, filter *model.NetworkDeviceFilter) ([]*model.NetworkDevice, error) {
	var (
		query = `
			SELECT id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, status, created_at, updated_at
			FROM network_devices
			WHERE 1=1
		`
		args []any
		argN = 1
	)

	if filter.Name != "" {
		query += fmt.Sprintf(" AND name = $%d", argN)
		args = append(args, filter.Name)
		argN++
	}
	if filter.Vendor != "" {
		query += fmt.Sprintf(" AND vendor = $%d", argN)
		args = append(args, filter.Vendor)
		argN++
	}
	if filter.Address != "" {
		query += fmt.Sprintf(" AND address = $%d", argN)
		args = append(args, filter.Address)
		argN++
	}
	if filter.Port > 0 {
		query += fmt.Sprintf(" AND port = $%d", argN)
		args = append(args, filter.Port)
		argN++
	}
	if filter.Protocol != "" {
		query += fmt.Sprintf(" AND protocol = $%d", argN)
		args = append(args, filter.Protocol)
		argN++
	}
	if filter.Format != "" {
		query += fmt.Sprintf(" AND format = $%d", argN)
		args = append(args, filter.Format)
		argN++
	}
	if filter.Facility != "" {
		query += fmt.Sprintf(" AND facility = $%d", argN)
		args = append(args, filter.Facility)
		argN++
	}
	if filter.SyslogID != "" {
		query += fmt.Sprintf(" AND syslog_id = $%d", argN)
		args = append(args, filter.SyslogID)
		argN++
	}
	if filter.RateLimit > 0 {
		query += fmt.Sprintf(" AND rate_limit = $%d", argN)
		args = append(args, filter.RateLimit)
		argN++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, filter.Status)
		argN++
	}
	// Pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var networkDevices []*model.NetworkDevice

	for rows.Next() {
		var device model.NetworkDevice
		if err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.Vendor,
			&device.Address,
			&device.Port,
			&device.Protocol,
			&device.Format,
			&device.Facility,
			&device.SyslogID,
			&device.RateLimit,
			&device.Status, // added
			&device.CreatedAt,
			&device.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		networkDevices = append(networkDevices, &device)
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
        SELECT id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, status, created_at, updated_at
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
		&nd.Status,
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

// UpsertNetworkDevice inserts a new network device or updates an existing one
func (s *PGDataStore) UpsertNetworkDevice(ctx context.Context, device *model.NetworkDevice) error {
	const q = `
		INSERT INTO network_devices
			(id, name, vendor, address, port, protocol, format, facility, syslog_id, rate_limit, status, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			vendor = EXCLUDED.vendor,
			address = EXCLUDED.address,
			port = EXCLUDED.port,
			protocol = EXCLUDED.protocol,
			format = EXCLUDED.format,
			facility = EXCLUDED.facility,
			syslog_id = EXCLUDED.syslog_id,
			rate_limit = EXCLUDED.rate_limit,
			status = EXCLUDED.status,
			updated_at = NOW()
	`
	_, err := s.db.ExecContext(ctx, q,
		device.ID,
		device.Name,
		device.Vendor,
		device.Address,
		device.Port,
		device.Protocol,
		device.Format,
		device.Facility,
		device.SyslogID,
		device.RateLimit,
		device.Status,
	)
	return err
}

// DeleteNetworkDeviceByID deletes a network device by its ID
func (s *PGDataStore) DeleteNetworkDeviceByID(ctx context.Context, id string) error {
	const q = `DELETE FROM network_devices WHERE id = $1`
	_, err := s.db.ExecContext(ctx, q, id)
	return err
}

// ToggleNetworkDeviceStatus toggles the status (enabled/disabled) of a device by ID
func (s *PGDataStore) ToggleNetworkDeviceStatus(ctx context.Context, id string) error {
	const q = `
		UPDATE network_devices
		SET status = CASE WHEN status = 'enabled' THEN 'disabled' ELSE 'enabled' END,
		    updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, q, id)
	return err
}
