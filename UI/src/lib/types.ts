// User types
export interface User {
	id: string;
	email: string;
	firstName?: string;
	first_name?: string;
	lastName?: string;
	last_name?: string;
	username?: string;
	avatar?: string;
	role: string;
	permissions: string[];
	createdAt?: string;
	created_at?: string;
	updatedAt?: string;
	updated_at?: string;
	last_login?: string;
	profile?: UserProfile;
}

// User Profile types
export interface UserProfile {
	full_name?: string;
	phone?: string;
	avatar_url?: string;
}

export interface ProfileUpdateRequest {
	full_name?: string;
	phone?: string;
}

export interface PasswordChangeRequest {
	current_password: string;
	new_password: string;
	confirm_password: string;
}

export interface UserPreferences {
	theme?: string;
	notifications?: boolean | {
		email_alerts?: boolean;
		push_alerts?: boolean;
		alert_frequency?: string;
	};
	dashboard?: {
		refresh_interval?: number;
		default_time_range?: string;
		show_system_metrics?: boolean;
	};
}

// Layout data type
export interface LayoutData {
	title?: string;
	user: User | null;
	permissions: Record<string, boolean>;
	meta: Record<string, any>;
}

// Alert types
export interface Alert {
	id: string;
	rule_id: string;
	endpoint_id?: string;
	state: string; // "ok", "firing", "resolved", "no_data"
	previous: string; // previous state
	scope: string; // "global", "endpoint", "agent", "user", "cloud" etc
	target: string; // e.g. "endpoint_id", "agent_id", "user_id"
	first_fired: string; // when it first started firing
	last_fired: string; // when it last evaluated as firing
	last_ok: string; // last time condition returned OK
	last_value: number; // most recent value
	level: string; // from rule (info/warning/critical)
	message: string; // expanded from template
	labels?: Record<string, string>;
	resolved_at?: string; // when it was resolved
	timestamp?: string; // alert timestamp
	
	// Widget-expected fields for compatibility
	name?: string;
	title?: string;
	description?: string;
	severity?: 'critical' | 'warning' | 'info' | 'success'; // Updated to match widget expectations
	status?: 'active' | 'resolved' | 'acknowledged';
	createdAt?: string;
	created_at?: string;
	updatedAt?: string;
	updated_at?: string;
	resolvedAt?: string;
	source?: string;
	endpointId?: string;
	endpoint_name?: string;
	conditions?: AlertCondition[];
	notifications?: AlertNotification[];
}

// API Response types
export interface ApiResponse<T = any> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
}

export interface AlertRulesResponse {
	rules?: AlertRule[];
}

export interface EndpointsResponse {
	endpoints?: Endpoint[];
}

// Backend AlertRule structure (matches Go model)
export interface AlertRule {
	id: string;
	name: string;
	description?: string;
	message: string;
	level: 'info' | 'warning' | 'critical';
	enabled: boolean;
	type: 'metric' | 'log' | 'event' | 'composite';
	match: MatchCriteria;
	scope: Scope;
	expression: Expression;
	actions: string[];
	options: Options;
}

export interface Scope {
	namespace?: string;
	subnamespace?: string;
	metric?: string;
}

export interface Expression {
	operator: string; // >, <, =, !=, contains, regex
	value: number | string;
	datatype?: string; // percent, numeric, status
}

export interface Options {
	cooldown?: string;
	eval_interval?: string;
	repeat_interval?: string;
	notify_on_resolve?: boolean;
}

export interface MatchCriteria {
	endpoint_ids?: string[];
	labels?: Record<string, string>;
	category?: string;
	source?: string;
	scope?: string;
}

// Alert summary for combining with rules
export interface AlertSummary {
	rule_id: string;
	state: string; // "firing", "resolved", etc.
	last_change: string; // ISO date string
}

// Combined alert table data (like old frontend)
export interface AlertTableData {
	id: string;
	name: string;
	state: string;
	last_state_change: string;
	conditions_summary: string;
	actions: string[];
}

// Frontend form data (simplified structure for UI)
export interface AlertRuleFormData {
	name: string;
	description?: string;
	severity: 'info' | 'warning' | 'critical';
	metric_name: string;
	operator: 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte';
	threshold: number;
	duration: number;
	endpoint_id?: string;
	enabled: boolean;
}

export interface AlertCondition {
	id: string;
	metric: string;
	operator: 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte';
	threshold: number;
	duration: number;
}

export interface AlertNotification {
	id: string;
	type: 'email' | 'webhook' | 'slack';
	target: string;
	enabled: boolean;
}

// Endpoint types
// Endpoint from API (raw format)
export interface EndpointApiResponse {
	id: string;
	hostname: string;
	ip?: string;
	arch?: string;
	last_seen?: string;
	os: string;
	status: string;
	type: string;
	uptime?: number;
	version?: string;
	agent_id?: string;
	host_id?: string;
	labels?: Record<string, string>;
}

// Endpoint for frontend use (normalized format)
export interface Endpoint {
	id: string;
	name: string;
	hostname: string;
	ip?: string;
	ipAddress?: string;
	ip_address?: string;
	port: number;
	status: 'online' | 'offline' | 'unknown';
	lastSeen?: string;
	last_seen?: string;
	agentVersion?: string;
	agent_version?: string;
	os: string;
	architecture: string;
	tags: string[];
	metrics?: EndpointMetric[];
	uptime?: number;
}

export interface EndpointMetric {
	name: string;
	value: number;
	unit: string;
	timestamp: string;
}

// Metric types
export interface Metric {
	id?: string;
	name: string;
	namespace?: string;
	subnamespace?: string;
	type?: 'gauge' | 'counter' | 'histogram';
	value: number;
	unit?: string;
	labels?: Record<string, string>;
	dimensions?: Record<string, string>;
	timestamp: string;
	endpointId?: string;
	endpoint_id?: string;
	stats?: {
		min: number;
		max: number;
		count: number;
		sum: number;
	};
	resolution?: number;
}

export interface MetricSeries {
	name: string;
	data: MetricDataPoint[];
	labels: Record<string, string>;
}

export interface MetricDataPoint {
	timestamp: string;
	value: number;
}

// Event types
export interface Event {
	id: string;
	level: string; // info, warning, critical
	type: string; // event type (system / alert)
	category: string; // metric, log, system, security
	message: string;
	source: string;
	scope: string; // "endpoint", "system", etc.
	target: string; // "host-123", "gosight-core", etc.
	timestamp: string;
	endpointId?: string;
	endpoint_id?: string;
	endpoint_name?: string;
	user_name?: string;
	meta?: Record<string, any>;
	
	// Widget-expected fields for compatibility
	title?: string;
	name?: string;
	description?: string;
	created_at?: string;
	metadata?: Record<string, any>;
}

// Log types
export interface LogMeta {
	platform?: string;
	app_name?: string;
	app_version?: string;
	container_id?: string;
	container_name?: string;
	unit?: string;
	service?: string;
	event_id?: string;
	user?: string;
	exe?: string;
	path?: string;
	extra?: Record<string, string>;
}

export interface LogEntry {
	id?: string; // Optional client-side ID for deduplication
	timestamp: string; // ISO timestamp
	level: string;
	message: string;
	source: string;
	category?: string;
	pid?: number;
	fields?: Record<string, string>;
	tags?: Record<string, string>;
	meta?: LogMeta;
}

export interface LogResponse {
	logs: LogEntry[];
	next_cursor?: string;
	has_more: boolean;
	count: number;
}

// Process types
export interface Process {
	id: string;
	pid: number;
	name: string;
	command: string;
	user: string;
	cpuPercent: number;
	memoryPercent: number;
	status: 'running' | 'sleeping' | 'stopped' | 'zombie';
	startTime: string;
	endpointId: string;
}

// Command types
export interface Command {
	id: string;
	command: string;
	args: string[];
	status: 'pending' | 'running' | 'completed' | 'failed';
	output?: string;
	error?: string;
	createdAt: string;
	completedAt?: string;
	endpointId: string;
}

// Network Device types
export interface NetworkDevice {
	id: string;
	name: string;
	type: 'router' | 'switch' | 'firewall' | 'access_point' | 'server';
	ipAddress: string;
	macAddress?: string;
	vendor?: string;
	model?: string;
	location?: string;
	status: 'up' | 'down' | 'unknown';
	lastSeen: string;
	ports: NetworkPort[];
}

export interface NetworkPort {
	id: string;
	number: number;
	name?: string;
	type: 'ethernet' | 'fiber' | 'wireless';
	status: 'up' | 'down';
	speed?: number;
	duplex?: 'full' | 'half';
}

// WebSocket message types
export interface WebSocketMessage {
	type: string;
	data: any;
	timestamp: string;
}



export interface PaginatedResponse<T = any> {
	data: T[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
}

// Filter and search types
export interface Filter {
	field: string;
	operator: 'eq' | 'ne' | 'gt' | 'lt' | 'gte' | 'lte' | 'contains' | 'starts_with' | 'ends_with';
	value: any;
}

export interface SearchParams {
	query?: string;
	filters?: Filter[];
	sortBy?: string;
	sortOrder?: 'asc' | 'desc';
	page?: number;
	pageSize?: number;
}

// Chart and visualization types
export interface ChartConfig {
	type: 'line' | 'bar' | 'pie' | 'area' | 'scatter';
	title?: string;
	xAxis?: AxisConfig;
	yAxis?: AxisConfig;
	series: SeriesConfig[];
}

export interface AxisConfig {
	title?: string;
	type?: 'datetime' | 'numeric' | 'category';
	min?: number;
	max?: number;
}

export interface SeriesConfig {
	name: string;
	data: any[];
	color?: string;
	type?: string;
}
