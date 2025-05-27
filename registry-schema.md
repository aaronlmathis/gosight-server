
# GoSight Resource Registry Schema Project

Full gosight documentation can be found at ./gosight-server/docs - Please review these before answering.



## Resource Model

A **Resource** represents any entity that produces telemetry in GoSight.

### Go Struct

```go
type Resource struct {
	ID          string            // Unique, stable resource ID (UUID or hash)
	Kind        string            // "host", "container", "app", "device", "syslog", "otel"
	Name        string            // Friendly name (e.g., "web-01", "api-prod")
	DisplayName string            // Optional override shown in UI (vs. system name)
	Group       string            // Logical group (e.g., team, project, stack)
	ParentID    string            // Optional parent resource ID (e.g., host for container)

	Labels      map[string]string // System-defined identity fields (from meta/telemetry)
	Tags        map[string]string // User-defined tags (editable from UI)

	Status      string            // Online, Offline, Idle, Unknown
	LastSeen    time.Time         // Timestamp of last signal
	FirstSeen   time.Time         // When resource was first registered
	CreatedAt   time.Time         // Persisted creation timestamp
	UpdatedAt   time.Time         // Persisted update timestamp

	Location    string            // Optional physical or logical location (e.g., us-east-1a, rack-12)
	Environment string            // "prod", "staging", "dev" (can also be a tag)
	Owner       string            // User/team/service owner
	Platform    string            // "linux", "windows", "aws", "gke", "azure" etc.
	Runtime     string            // "docker", "podman", "kubernetes", "systemd", etc.
	Version     string            // Agent version, service version, etc.
	OS          string            // OS details (e.g., "ubuntu 22.04", "rhel 9.2")
	Arch        string            // "amd64", "arm64"
	IPAddress   string            // Optional primary IP address

	ResourceType string           // Optional refinement of Kind (e.g., "vm", "ec2", "ecs-task", "lambda")
	Cluster      string           // Kubernetes/OpenShift cluster name (if applicable)
	Namespace    string           // Kubernetes namespace or logical grouping
	Annotations  map[string]string // For internal use / UI / integrations

	Updated     bool              // Dirty flag (needs sync to store)
}
```

## Term Definitions

| Term        | Definition |
|-------------|------------|
| Resource    | A unit that produces telemetry (metrics, logs, traces). Replaces the "endpoint" concept. |
| Kind        | The resource type — e.g., host, container, app, device, otel. |
| ID          | A stable internal identifier (like a UUID or generated hash). |
| Name        | A user-friendly name shown in the UI (e.g., web-01, nginx-ctr-123). |
| Group       | A logical grouping (e.g., by team, service, or business unit). |
| Labels      | System-generated identity metadata — e.g., hostname, agent_id, container_id. |
| Tags        | User-defined metadata used for filtering, scoping, and UI classification — e.g., env=prod, role=db. |
| ParentID    | ID of the parent resource, if nested (e.g., container on a host). |
| Status      | Live state (e.g., Online, Idle, Offline). |
| Updated     | Indicates if the resource should be flushed to the database. |

## Tags vs Labels

| Attribute   | Labels                              | Tags                                          |
|-------------|-------------------------------------|-----------------------------------------------|
| Origin      | Set by agent/system                 | Set by user or enriched by server             |
| Use         | Identity, joins, metric/log scoping | UI filters, dashboards, alert groupings       |
| Examples    | hostname, container_id, job         | env=prod, team=infra, critical=true           |
| Format      | Key-value                           | Key-value or key-only                         |
| Mutability  | Immutable/stable                    | Editable/flexible                             |

## PostgreSQL Schema

### Table: `resources`

```sql
CREATE TABLE resources (
    id           UUID PRIMARY KEY,
    kind         TEXT NOT NULL,
    name         TEXT NOT NULL,
    group_name   TEXT,
    parent_id    UUID REFERENCES resources(id),
    status       TEXT NOT NULL DEFAULT 'Offline',
    last_seen    TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Table: `resource_labels`

```sql
CREATE TABLE resource_labels (
    resource_id UUID REFERENCES resources(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (resource_id, key)
);
```

### Table: `resource_tags`

```sql
CREATE TABLE resource_tags (
    resource_id UUID REFERENCES resources(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    PRIMARY KEY (resource_id, key)
);
```

### Indexes

```sql
CREATE INDEX idx_resources_kind ON resources(kind);
CREATE INDEX idx_resources_group_name ON resources(group_name);
CREATE INDEX idx_resources_last_seen ON resources(last_seen);

CREATE INDEX idx_labels_key_value ON resource_labels(key, value);
CREATE INDEX idx_tags_key_value ON resource_tags(key, value);
```

## Optional: Materialized View

```sql
CREATE MATERIALIZED VIEW resource_summary AS
SELECT
    r.id,
    r.name,
    r.kind,
    r.group_name,
    r.status,
    r.last_seen,
    jsonb_object_agg(l.key, l.value) AS labels,
    jsonb_object_agg(t.key, t.value) AS tags
FROM resources r
LEFT JOIN resource_labels l ON r.id = l.resource_id
LEFT JOIN resource_tags t ON r.id = t.resource_id
GROUP BY r.id;
```

This view supports fast lookups for dashboards and filtering.
