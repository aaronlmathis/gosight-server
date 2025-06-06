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

// Package pgstore provides PostgreSQL storage implementation for GoSight process data.
//
// This file implements process-specific storage operations including:
//   - Bulk insertion of process snapshots for efficient batch processing
//   - Individual process snapshot and detail insertion
//   - Flexible querying with filtering, sorting, and pagination
//   - JSON handling for process labels and metadata
//
// The storage model separates process snapshots (point-in-time system state)
// from individual process details to optimize for both write performance
// and query flexibility. Process snapshots contain host metadata and timing
// information, while process_info records contain the detailed process data.
package pgstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
)

// Write implements the BufferedDataStore interface for process payloads.
// It performs bulk insertion of process snapshots for efficient batch processing.
func (s *PGDataStore) Write(ctx context.Context, batches []*model.ProcessPayload) error {
	if len(batches) == 0 {
		return nil
	}
	return s.bulkInsertSnapshots(ctx, batches)
}

// bulkInsertSnapshots performs efficient bulk insertion of process snapshots into the database.
// It constructs a single INSERT statement with multiple value sets to minimize database roundtrips.
// Only process metadata is stored at the snapshot level; individual process details are handled separately.
func (s *PGDataStore) bulkInsertSnapshots(ctx context.Context, batches []*model.ProcessPayload) error {
	var (
		args         []interface{}
		placeholders []string
	)

	for i, snap := range batches {
		metaJSON, err := json.Marshal(snap.Meta)
		if err != nil {
			utils.Warn("Failed to marshal metadata for snapshot from endpoint %s: %v", snap.EndpointID, err)
			continue
		}

		idx := i * 6
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)",
			idx+1, idx+2, idx+3, idx+4, idx+5, idx+6))

		args = append(args,
			snap.Timestamp,
			snap.HostID,
			snap.EndpointID,
			snap.AgentID,
			snap.Hostname,
			metaJSON,
		)
	}

	if len(args) == 0 {
		utils.Warn("No valid process snapshots to insert")
		return nil
	}

	query := `
		INSERT INTO process_snapshots (timestamp, host_id, endpoint_id, agent_id, hostname, meta)
		VALUES ` + strings.Join(placeholders, ",")

	start := time.Now()
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		utils.Error("Bulk insert of %d process snapshots failed: %v", len(batches), err)
		return err
	}

	utils.Debug("Inserted %d process snapshots in %s", len(batches), time.Since(start))
	return nil
}

// InsertFullProcessPayload inserts a complete process payload into the database.
// This is a convenience method that combines snapshot creation with process detail insertion
// in a single operation. It first creates a process snapshot and then inserts all
// associated process information records.
func (s *PGDataStore) InsertFullProcessPayload(ctx context.Context, payload *model.ProcessPayload) error {
	snapshotID, err := s.InsertProcessSnapshot(ctx, payload)
	if err != nil {
		return err
	}
	return s.InsertProcessInfos(ctx, snapshotID, payload)
}

// InsertProcessSnapshot inserts a new process snapshot into the database.
// A process snapshot represents a point-in-time capture of system process state,
// including metadata about the host, agent, and collection timestamp.
// Returns the generated snapshot ID for linking associated process details.
func (s *PGDataStore) InsertProcessSnapshot(ctx context.Context, snap *model.ProcessPayload) (int64, error) {
	metaJSON, err := json.Marshal(snap.Meta)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal meta: %w", err)
	}

	var snapshotID int64
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO process_snapshots (timestamp, host_id, endpoint_id, agent_id, hostname, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		snap.Timestamp,
		snap.HostID,
		snap.EndpointID,
		snap.AgentID,
		snap.Hostname,
		metaJSON,
	).Scan(&snapshotID)
	if err != nil {
		return 0, fmt.Errorf("insert snapshot failed: %w", err)
	}

	return snapshotID, nil
}

// InsertProcessInfos inserts multiple process information records associated with a snapshot.
// Each process record includes detailed information such as PID, CPU/memory usage,
// command line, executable path, and custom labels. The insertion is performed
// within a transaction to ensure data consistency.
func (s *PGDataStore) InsertProcessInfos(ctx context.Context, snapshotID int64, payload *model.ProcessPayload) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO process_info (
			snapshot_id, pid, ppid, username, exe, cmdline, cpu_percent, mem_percent,
			threads, start_time, tags, timestamp, endpoint_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	for _, p := range payload.Processes {
		labelsJSON, _ := json.Marshal(p.Labels)
		_, err := stmt.ExecContext(ctx,
			snapshotID,
			p.PID,
			p.PPID,
			p.User,
			p.Executable,
			p.Cmdline,
			p.CPUPercent,
			p.MemPercent,
			p.Threads,
			p.StartTime,
			labelsJSON,
			// new fields:
			payload.Timestamp,
			payload.EndpointID,
		)
		if err != nil {
			return fmt.Errorf("exec insert: %w", err)
		}
	}

	return tx.Commit()
}

// QueryProcessInfos retrieves process information based on the provided filter criteria.
// It supports filtering by endpoint, time range, resource usage thresholds, user,
// process IDs, and text searches in executable paths and command lines.
// Results can be sorted by various fields and paginated for large datasets.
func (s *PGDataStore) QueryProcessInfos(ctx context.Context, filter *model.ProcessQueryFilter) ([]model.ProcessInfo, error) {
	var args []interface{}
	var where []string

	where = append(where, "ps.endpoint_id = ?")
	args = append(args, filter.EndpointID)

	if !filter.Start.IsZero() {
		where = append(where, "ps.timestamp >= ?")
		args = append(args, filter.Start)
	}
	if !filter.End.IsZero() {
		where = append(where, "ps.timestamp <= ?")
		args = append(args, filter.End)
	}
	if filter.MinCPU > 0 {
		where = append(where, "pi.cpu_percent >= ?")
		args = append(args, filter.MinCPU)
	}
	if filter.MaxCPU > 0 {
		where = append(where, "pi.cpu_percent <= ?")
		args = append(args, filter.MaxCPU)
	}
	if filter.MinMemory > 0 {
		where = append(where, "pi.mem_percent >= ?")
		args = append(args, filter.MinMemory)
	}
	if filter.MaxMemory > 0 {
		where = append(where, "pi.mem_percent <= ?")
		args = append(args, filter.MaxMemory)
	}
	if filter.User != "" {
		where = append(where, "pi.user = ?")
		args = append(args, filter.User)
	}
	if filter.PID > 0 {
		where = append(where, "pi.pid = ?")
		args = append(args, filter.PID)
	}
	if filter.PPID > 0 {
		where = append(where, "pi.ppid = ?")
		args = append(args, filter.PPID)
	}
	if filter.ExeContains != "" {
		where = append(where, "pi.exe ILIKE ?")
		args = append(args, "%"+filter.ExeContains+"%")
	}
	if filter.CmdlineContains != "" {
		where = append(where, "pi.cmdline ILIKE ?")
		args = append(args, "%"+filter.CmdlineContains+"%")
	}

	// Base query
	query := `
		SELECT pi.pid, pi.ppid, pi.user, pi.exe, pi.cmdline, pi.cpu_percent, pi.mem_percent,
			pi.threads, pi.start_time, pi.tags
		FROM process_info pi
		JOIN process_snapshots ps ON pi.snapshot_id = ps.id
		WHERE ` + strings.Join(where, " AND ")

	// Sorting
	sortField := "pi.cpu_percent"
	if filter.SortBy != "" {
		sortField = "pi." + filter.SortBy
	}
	direction := "DESC"
	if !filter.SortDesc {
		direction = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortField, direction)

	// Pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query process infos: %w", err)
	}
	defer rows.Close()

	var results []model.ProcessInfo
	for rows.Next() {
		var p model.ProcessInfo
		var startTime time.Time
		var labelsData []byte

		err := rows.Scan(
			&p.PID, &p.PPID, &p.User, &p.Executable, &p.Cmdline,
			&p.CPUPercent, &p.MemPercent, &p.Threads, &startTime, &labelsData,
		)
		if err != nil {
			return nil, fmt.Errorf("scan process row: %w", err)
		}
		p.StartTime = startTime
		p.Labels = parseLabelsJSON(labelsData)
		results = append(results, p)
	}

	return results, nil
}

// parseLabelsJSON is a helper function to parse JSONB labels from the database.
// It safely unmarshals JSON data into a string map, returning an empty map
// if parsing fails to prevent data corruption errors from affecting queries.
func parseLabelsJSON(data []byte) map[string]string {
	var m map[string]string
	_ = json.Unmarshal(data, &m)
	return m
}
