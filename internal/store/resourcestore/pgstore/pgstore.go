package pgstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/store/resourcestore"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/lib/pq"
)

type PGResourceStore struct {
	db *sql.DB
}

func NewPGResourceStore(db *sql.DB) resourcestore.ResourceStore {
	return &PGResourceStore{db: db}
}

func (s *PGResourceStore) Create(ctx context.Context, resource *model.Resource) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert main resource record
	query := `
        INSERT INTO resources (
            id, kind, name, display_name, group_name, parent_id,
            status, last_seen, first_seen, location, environment,
            owner, platform, runtime, version, os, arch, ip_address,
            resource_type, cluster, namespace
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
            $14, $15, $16, $17, $18, $19, $20, $21
        )`

	_, err = tx.ExecContext(ctx, query,
		resource.ID, resource.Kind, resource.Name, resource.DisplayName,
		resource.Group, nullString(resource.ParentID), resource.Status,
		resource.LastSeen, resource.FirstSeen, resource.Location,
		resource.Environment, resource.Owner, resource.Platform,
		resource.Runtime, resource.Version, resource.OS, resource.Arch,
		resource.IPAddress, resource.ResourceType, resource.Cluster,
		resource.Namespace,
	)
	if err != nil {
		return err
	}

	// Insert labels
	if err := s.insertKeyValues(ctx, tx, "resource_labels", resource.ID, resource.Labels); err != nil {
		return err
	}

	// Insert tags
	if err := s.insertKeyValues(ctx, tx, "resource_tags", resource.ID, resource.Tags); err != nil {
		return err
	}

	// Insert annotations
	if err := s.insertKeyValues(ctx, tx, "resource_annotations", resource.ID, resource.Annotations); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGResourceStore) CreateBatch(ctx context.Context, resources []*model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare batch insert for main resource records
	query := `
        INSERT INTO resources (
            id, kind, name, display_name, group_name, parent_id,
            status, last_seen, first_seen, location, environment,
            owner, platform, runtime, version, os, arch, ip_address,
            resource_type, cluster, namespace
        ) VALUES `

	values := make([]interface{}, 0, len(resources)*21)
	placeholders := make([]string, 0, len(resources))

	i := 1
	for _, resource := range resources {
		placeholder := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i, i+1, i+2, i+3, i+4, i+5, i+6, i+7, i+8, i+9, i+10, i+11, i+12, i+13, i+14, i+15, i+16, i+17, i+18, i+19, i+20)
		placeholders = append(placeholders, placeholder)

		values = append(values,
			resource.ID, resource.Kind, resource.Name, resource.DisplayName,
			resource.Group, nullString(resource.ParentID), resource.Status,
			resource.LastSeen, resource.FirstSeen, resource.Location,
			resource.Environment, resource.Owner, resource.Platform,
			resource.Runtime, resource.Version, resource.OS, resource.Arch,
			resource.IPAddress, resource.ResourceType, resource.Cluster,
			resource.Namespace,
		)
		i += 21
	}

	query += strings.Join(placeholders, ", ")
	_, err = tx.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	// Insert labels, tags, and annotations for each resource
	for _, resource := range resources {
		if err := s.insertKeyValues(ctx, tx, "resource_labels", resource.ID, resource.Labels); err != nil {
			return err
		}
		if err := s.insertKeyValues(ctx, tx, "resource_tags", resource.ID, resource.Tags); err != nil {
			return err
		}
		if err := s.insertKeyValues(ctx, tx, "resource_annotations", resource.ID, resource.Annotations); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PGResourceStore) Get(ctx context.Context, id string) (*model.Resource, error) {
	resource := &model.Resource{}

	query := `
        SELECT id, kind, name, display_name, group_name, parent_id,
               status, last_seen, first_seen, created_at, updated_at,
               location, environment, owner, platform, runtime, version,
               os, arch, ip_address, resource_type, cluster, namespace
        FROM resources WHERE id = $1`

	var parentID sql.NullString
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
		&resource.Group, &parentID, &resource.Status, &resource.LastSeen,
		&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
		&resource.Location, &resource.Environment, &resource.Owner,
		&resource.Platform, &resource.Runtime, &resource.Version,
		&resource.OS, &resource.Arch, &resource.IPAddress,
		&resource.ResourceType, &resource.Cluster, &resource.Namespace,
	)
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		resource.ParentID = parentID.String
	}

	// Load labels, tags, and annotations
	resource.Labels, _ = s.getKeyValues(ctx, "resource_labels", id)
	resource.Tags, _ = s.getKeyValues(ctx, "resource_tags", id)
	resource.Annotations, _ = s.getKeyValues(ctx, "resource_annotations", id)

	return resource, nil
}

func (s *PGResourceStore) Update(ctx context.Context, resource *model.Resource) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update main resource record
	query := `
        UPDATE resources SET
            kind = $2, name = $3, display_name = $4, group_name = $5, parent_id = $6,
            status = $7, last_seen = $8, location = $9, environment = $10,
            owner = $11, platform = $12, runtime = $13, version = $14, os = $15,
            arch = $16, ip_address = $17, resource_type = $18, cluster = $19,
            namespace = $20, updated_at = now()
        WHERE id = $1`

	_, err = tx.ExecContext(ctx, query,
		resource.ID, resource.Kind, resource.Name, resource.DisplayName,
		resource.Group, nullString(resource.ParentID), resource.Status,
		resource.LastSeen, resource.Location, resource.Environment,
		resource.Owner, resource.Platform, resource.Runtime, resource.Version,
		resource.OS, resource.Arch, resource.IPAddress, resource.ResourceType,
		resource.Cluster, resource.Namespace,
	)
	if err != nil {
		return err
	}

	// Delete existing key-value pairs
	for _, table := range []string{"resource_labels", "resource_tags", "resource_annotations"} {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE resource_id = $1", table), resource.ID)
		if err != nil {
			return err
		}
	}

	// Insert updated key-value pairs
	if err := s.insertKeyValues(ctx, tx, "resource_labels", resource.ID, resource.Labels); err != nil {
		return err
	}
	if err := s.insertKeyValues(ctx, tx, "resource_tags", resource.ID, resource.Tags); err != nil {
		return err
	}
	if err := s.insertKeyValues(ctx, tx, "resource_annotations", resource.ID, resource.Annotations); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGResourceStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM resources WHERE id = $1", id)
	return err
}

func (s *PGResourceStore) UpdateBatch(ctx context.Context, resources []*model.Resource) error {
	if len(resources) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update each resource individually for simplicity
	for _, resource := range resources {
		query := `
            UPDATE resources SET
                kind = $2, name = $3, display_name = $4, group_name = $5, parent_id = $6,
                status = $7, last_seen = $8, location = $9, environment = $10,
                owner = $11, platform = $12, runtime = $13, version = $14, os = $15,
                arch = $16, ip_address = $17, resource_type = $18, cluster = $19,
                namespace = $20, updated_at = now()
            WHERE id = $1`

		_, err = tx.ExecContext(ctx, query,
			resource.ID, resource.Kind, resource.Name, resource.DisplayName,
			resource.Group, nullString(resource.ParentID), resource.Status,
			resource.LastSeen, resource.Location, resource.Environment,
			resource.Owner, resource.Platform, resource.Runtime, resource.Version,
			resource.OS, resource.Arch, resource.IPAddress, resource.ResourceType,
			resource.Cluster, resource.Namespace,
		)
		if err != nil {
			return err
		}

		// Delete and re-insert key-value pairs
		for _, table := range []string{"resource_labels", "resource_tags", "resource_annotations"} {
			_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE resource_id = $1", table), resource.ID)
			if err != nil {
				return err
			}
		}

		if err := s.insertKeyValues(ctx, tx, "resource_labels", resource.ID, resource.Labels); err != nil {
			return err
		}
		if err := s.insertKeyValues(ctx, tx, "resource_tags", resource.ID, resource.Tags); err != nil {
			return err
		}
		if err := s.insertKeyValues(ctx, tx, "resource_annotations", resource.ID, resource.Annotations); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PGResourceStore) List(ctx context.Context, filter *model.ResourceFilter, limit, offset int) ([]*model.Resource, error) {
	query := `
        SELECT id, kind, name, display_name, group_name, parent_id,
               status, last_seen, first_seen, created_at, updated_at,
               location, environment, owner, platform, runtime, version,
               os, arch, ip_address, resource_type, cluster, namespace
        FROM resources WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if len(filter.Kinds) > 0 {
			query += fmt.Sprintf(" AND kind = ANY($%d)", argIndex)
			args = append(args, pq.Array(filter.Kinds))
			argIndex++
		}
		if len(filter.Groups) > 0 {
			query += fmt.Sprintf(" AND group_name = ANY($%d)", argIndex)
			args = append(args, pq.Array(filter.Groups))
			argIndex++
		}
		if len(filter.Status) > 0 {
			query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			args = append(args, pq.Array(filter.Status))
			argIndex++
		}
		if len(filter.Environment) > 0 {
			query += fmt.Sprintf(" AND environment = ANY($%d)", argIndex)
			args = append(args, pq.Array(filter.Environment))
			argIndex++
		}
		if len(filter.Owner) > 0 {
			query += fmt.Sprintf(" AND owner = ANY($%d)", argIndex)
			args = append(args, pq.Array(filter.Owner))
			argIndex++
		}
		if filter.LastSeenSince != nil {
			query += fmt.Sprintf(" AND last_seen >= $%d", argIndex)
			args = append(args, *filter.LastSeenSince)
			argIndex++
		}
	}

	query += fmt.Sprintf(" ORDER BY last_seen DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		resource := &model.Resource{}
		var parentID sql.NullString
		err := rows.Scan(
			&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
			&resource.Group, &parentID, &resource.Status, &resource.LastSeen,
			&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
			&resource.Location, &resource.Environment, &resource.Owner,
			&resource.Platform, &resource.Runtime, &resource.Version,
			&resource.OS, &resource.Arch, &resource.IPAddress,
			&resource.ResourceType, &resource.Cluster, &resource.Namespace,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			resource.ParentID = parentID.String
		}

		// Load labels, tags, and annotations
		resource.Labels, _ = s.getKeyValues(ctx, "resource_labels", resource.ID)
		resource.Tags, _ = s.getKeyValues(ctx, "resource_tags", resource.ID)
		resource.Annotations, _ = s.getKeyValues(ctx, "resource_annotations", resource.ID)

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

func (s *PGResourceStore) GetChildren(ctx context.Context, parentID string) ([]*model.Resource, error) {
	return s.List(ctx, &model.ResourceFilter{}, 1000, 0) // Simple implementation, filter by parent in query later
}

func (s *PGResourceStore) GetParent(ctx context.Context, childID string) (*model.Resource, error) {
	child, err := s.Get(ctx, childID)
	if err != nil {
		return nil, err
	}
	if child.ParentID == "" {
		return nil, nil
	}
	return s.Get(ctx, child.ParentID)
}

func (s *PGResourceStore) GetByLabels(ctx context.Context, labels map[string]string) ([]*model.Resource, error) {
	if len(labels) == 0 {
		return []*model.Resource{}, nil
	}

	// Build query to find resources with matching labels
	query := `
        SELECT DISTINCT r.id, r.kind, r.name, r.display_name, r.group_name, r.parent_id,
               r.status, r.last_seen, r.first_seen, r.created_at, r.updated_at,
               r.location, r.environment, r.owner, r.platform, r.runtime, r.version,
               r.os, r.arch, r.ip_address, r.resource_type, r.cluster, r.namespace
        FROM resources r
        JOIN resource_labels l ON r.id = l.resource_id
        WHERE `

	conditions := make([]string, 0, len(labels))
	args := make([]interface{}, 0, len(labels)*2)
	argIndex := 1

	for key, value := range labels {
		conditions = append(conditions, fmt.Sprintf("(l.key = $%d AND l.value = $%d)", argIndex, argIndex+1))
		args = append(args, key, value)
		argIndex += 2
	}

	query += strings.Join(conditions, " OR ")

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		resource := &model.Resource{}
		var parentID sql.NullString
		err := rows.Scan(
			&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
			&resource.Group, &parentID, &resource.Status, &resource.LastSeen,
			&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
			&resource.Location, &resource.Environment, &resource.Owner,
			&resource.Platform, &resource.Runtime, &resource.Version,
			&resource.OS, &resource.Arch, &resource.IPAddress,
			&resource.ResourceType, &resource.Cluster, &resource.Namespace,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			resource.ParentID = parentID.String
		}

		// Load labels, tags, and annotations
		resource.Labels, _ = s.getKeyValues(ctx, "resource_labels", resource.ID)
		resource.Tags, _ = s.getKeyValues(ctx, "resource_tags", resource.ID)
		resource.Annotations, _ = s.getKeyValues(ctx, "resource_annotations", resource.ID)

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

func (s *PGResourceStore) GetByTags(ctx context.Context, tags map[string]string) ([]*model.Resource, error) {
	if len(tags) == 0 {
		return []*model.Resource{}, nil
	}

	// Build query to find resources with matching tags
	query := `
        SELECT DISTINCT r.id, r.kind, r.name, r.display_name, r.group_name, r.parent_id,
               r.status, r.last_seen, r.first_seen, r.created_at, r.updated_at,
               r.location, r.environment, r.owner, r.platform, r.runtime, r.version,
               r.os, r.arch, r.ip_address, r.resource_type, r.cluster, r.namespace
        FROM resources r
        JOIN resource_tags t ON r.id = t.resource_id
        WHERE `

	conditions := make([]string, 0, len(tags))
	args := make([]interface{}, 0, len(tags)*2)
	argIndex := 1

	for key, value := range tags {
		conditions = append(conditions, fmt.Sprintf("(t.key = $%d AND t.value = $%d)", argIndex, argIndex+1))
		args = append(args, key, value)
		argIndex += 2
	}

	query += strings.Join(conditions, " OR ")

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		resource := &model.Resource{}
		var parentID sql.NullString
		err := rows.Scan(
			&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
			&resource.Group, &parentID, &resource.Status, &resource.LastSeen,
			&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
			&resource.Location, &resource.Environment, &resource.Owner,
			&resource.Platform, &resource.Runtime, &resource.Version,
			&resource.OS, &resource.Arch, &resource.IPAddress,
			&resource.ResourceType, &resource.Cluster, &resource.Namespace,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			resource.ParentID = parentID.String
		}

		// Load labels, tags, and annotations
		resource.Labels, _ = s.getKeyValues(ctx, "resource_labels", resource.ID)
		resource.Tags, _ = s.getKeyValues(ctx, "resource_tags", resource.ID)
		resource.Annotations, _ = s.getKeyValues(ctx, "resource_annotations", resource.ID)

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

func (s *PGResourceStore) UpdateLabels(ctx context.Context, resourceID string, labels map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing labels
	_, err = tx.ExecContext(ctx, "DELETE FROM resource_labels WHERE resource_id = $1", resourceID)
	if err != nil {
		return err
	}

	// Insert new labels
	if err := s.insertKeyValues(ctx, tx, "resource_labels", resourceID, labels); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGResourceStore) UpdateTags(ctx context.Context, resourceID string, tags map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing tags
	_, err = tx.ExecContext(ctx, "DELETE FROM resource_tags WHERE resource_id = $1", resourceID)
	if err != nil {
		return err
	}

	// Insert new tags
	if err := s.insertKeyValues(ctx, tx, "resource_tags", resourceID, tags); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PGResourceStore) UpdateStatus(ctx context.Context, resourceID string, status string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE resources SET status = $1, updated_at = now() WHERE id = $2",
		status, resourceID)
	return err
}

func (s *PGResourceStore) UpdateLastSeen(ctx context.Context, resourceID string, lastSeen time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE resources SET last_seen = $1, updated_at = now() WHERE id = $2",
		lastSeen, resourceID)
	return err
}

func (s *PGResourceStore) GetStaleResources(ctx context.Context, threshold time.Duration) ([]*model.Resource, error) {
	cutoff := time.Now().Add(-threshold)
	query := `
        SELECT id, kind, name, display_name, group_name, parent_id,
               status, last_seen, first_seen, created_at, updated_at,
               location, environment, owner, platform, runtime, version,
               os, arch, ip_address, resource_type, cluster, namespace
        FROM resources 
        WHERE last_seen < $1 
        ORDER BY last_seen ASC`

	rows, err := s.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		resource := &model.Resource{}
		var parentID sql.NullString
		err := rows.Scan(
			&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
			&resource.Group, &parentID, &resource.Status, &resource.LastSeen,
			&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
			&resource.Location, &resource.Environment, &resource.Owner,
			&resource.Platform, &resource.Runtime, &resource.Version,
			&resource.OS, &resource.Arch, &resource.IPAddress,
			&resource.ResourceType, &resource.Cluster, &resource.Namespace,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			resource.ParentID = parentID.String
		}

		// Load labels, tags, and annotations
		resource.Labels, _ = s.getKeyValues(ctx, "resource_labels", resource.ID)
		resource.Tags, _ = s.getKeyValues(ctx, "resource_tags", resource.ID)
		resource.Annotations, _ = s.getKeyValues(ctx, "resource_annotations", resource.ID)

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

func (s *PGResourceStore) GetResourceSummary(ctx context.Context) (map[string]int, error) {
	query := "SELECT kind, COUNT(*) FROM resources GROUP BY kind"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := make(map[string]int)
	for rows.Next() {
		var kind string
		var count int
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, err
		}
		summary[kind] = count
	}

	return summary, rows.Err()
}

func (s *PGResourceStore) GetResourcesByKind(ctx context.Context, kind string) ([]*model.Resource, error) {
	filter := &model.ResourceFilter{
		Kinds: []string{kind},
	}
	return s.List(ctx, filter, 1000, 0)
}

// Helper methods
func (s *PGResourceStore) insertKeyValues(ctx context.Context, tx *sql.Tx, table, resourceID string, kvs map[string]string) error {
	if len(kvs) == 0 {
		return nil
	}

	query := fmt.Sprintf("INSERT INTO %s (resource_id, key, value) VALUES ", table)
	values := make([]interface{}, 0, len(kvs)*3)
	placeholders := make([]string, 0, len(kvs))

	i := 1
	for k, v := range kvs {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", i, i+1, i+2))
		values = append(values, resourceID, k, v)
		i += 3
	}

	query += strings.Join(placeholders, ", ")
	_, err := tx.ExecContext(ctx, query, values...)
	return err
}

func (s *PGResourceStore) getKeyValues(ctx context.Context, table, resourceID string) (map[string]string, error) {
	query := fmt.Sprintf("SELECT key, value FROM %s WHERE resource_id = $1", table)
	rows, err := s.db.QueryContext(ctx, query, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, rows.Err()
}

func (s *PGResourceStore) Count(ctx context.Context, filter *model.ResourceFilter) (int, error) {
	query := "SELECT COUNT(*) FROM resources WHERE 1=1"
	args := []interface{}{}

	if filter != nil {
		if len(filter.Kinds) > 0 {
			query += " AND kind = ANY($1)"
			args = append(args, pq.Array(filter.Kinds))
		}
		if len(filter.Groups) > 0 {
			query += " AND group_name = ANY($2)"
			args = append(args, pq.Array(filter.Groups))
		}
		if len(filter.Status) > 0 {
			query += " AND status = ANY($3)"
			args = append(args, pq.Array(filter.Status))
		}
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (s *PGResourceStore) Search(ctx context.Context, query *model.ResourceSearchQuery) ([]*model.Resource, error) {
	var args []interface{}
	argCounter := 0

	sqlQuery := `
		SELECT id, kind, name, display_name, COALESCE(group_name, '') as group_name,
			   COALESCE(parent_id, '') as parent_id, status, last_seen, first_seen,
			   created_at, updated_at, COALESCE(location, '') as location,
			   COALESCE(environment, '') as environment, COALESCE(owner, '') as owner,
			   COALESCE(platform, '') as platform, COALESCE(runtime, '') as runtime,
			   COALESCE(version, '') as version, COALESCE(os, '') as os,
			   COALESCE(arch, '') as arch, COALESCE(ip_address, '') as ip_address,
			   COALESCE(resource_type, '') as resource_type, COALESCE(cluster, '') as cluster,
			   COALESCE(namespace, '') as namespace
		FROM resources WHERE 1=1`

	// Add text search on name and display_name if query is provided
	if query.Query != "" {
		argCounter++
		sqlQuery += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d)", argCounter, argCounter)
		args = append(args, "%"+query.Query+"%")
	}

	// Add kind filter
	if len(query.Kinds) > 0 {
		argCounter++
		sqlQuery += fmt.Sprintf(" AND kind = ANY($%d)", argCounter)
		args = append(args, pq.Array(query.Kinds))
	}

	// Add group filter
	if len(query.Groups) > 0 {
		argCounter++
		sqlQuery += fmt.Sprintf(" AND group_name = ANY($%d)", argCounter)
		args = append(args, pq.Array(query.Groups))
	}

	// Add status filter
	if len(query.Status) > 0 {
		argCounter++
		sqlQuery += fmt.Sprintf(" AND status = ANY($%d)", argCounter)
		args = append(args, pq.Array(query.Status))
	}

	// Add ordering and limits
	sqlQuery += " ORDER BY created_at DESC"
	
	if query.Limit > 0 {
		argCounter++
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argCounter)
		args = append(args, query.Limit)
	}

	if query.Offset > 0 {
		argCounter++
		sqlQuery += fmt.Sprintf(" OFFSET $%d", argCounter)
		args = append(args, query.Offset)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		resource := &model.Resource{}
		err := rows.Scan(
			&resource.ID, &resource.Kind, &resource.Name, &resource.DisplayName,
			&resource.Group, &resource.ParentID, &resource.Status, &resource.LastSeen,
			&resource.FirstSeen, &resource.CreatedAt, &resource.UpdatedAt,
			&resource.Location, &resource.Environment, &resource.Owner,
			&resource.Platform, &resource.Runtime, &resource.Version,
			&resource.OS, &resource.Arch, &resource.IPAddress,
			&resource.ResourceType, &resource.Cluster, &resource.Namespace,
		)
		if err != nil {
			return nil, err
		}

		// Initialize empty maps for labels, tags, and annotations
		resource.Labels = make(map[string]string)
		resource.Tags = make(map[string]string)
		resource.Annotations = make(map[string]string)

		resources = append(resources, resource)
	}

	return resources, rows.Err()
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
