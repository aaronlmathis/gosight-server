export interface Resource {
    id: string;
    kind: string;
    name: string;
    display_name?: string;
    group?: string;
    parent_id?: string;
    
    labels: Record<string, string>;
    tags: Record<string, string>;
    
    status: ResourceStatus;
    last_seen: string;
    first_seen: string;
    created_at: string;
    updated_at: string;
    
    location?: string;
    environment?: string;
    owner?: string;
    platform?: string;
    runtime?: string;
    version?: string;
    os?: string;
    arch?: string;
    ip_address?: string;
    
    resource_type?: string;
    cluster?: string;
    namespace?: string;
    annotations: Record<string, string>;
}

export enum ResourceStatus {
    Online = 'online',
    Offline = 'offline',
    Idle = 'idle',
    Unknown = 'unknown'
}

export enum ResourceKind {
    Host = 'host',
    Container = 'container',
    App = 'app',
    Device = 'device',
    Syslog = 'syslog',
    Otel = 'otel'
}

export interface ResourceFilter {
    kinds?: string[];
    groups?: string[];
    status?: string[];
    labels?: Record<string, string>;
    tags?: Record<string, string>;
    environment?: string[];
    owner?: string[];
    last_seen_since?: string;
}

export interface ResourceSummary {
    total: number;
    by_kind: Record<string, number>;
    by_status: Record<string, number>;
    by_environment: Record<string, number>;
}