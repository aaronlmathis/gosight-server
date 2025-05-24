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
	name?: string;
	title?: string;
	description?: string;
	severity: 'low' | 'medium' | 'high';
	status: 'active' | 'resolved' | 'acknowledged';
	createdAt?: string;
	created_at?: string;
	updatedAt?: string;
	updated_at?: string;
	resolvedAt?: string;
	resolved_at?: string;
	source?: string;
	endpoint_id?: string;
	endpoint_name?: string;
	endpointId?: string;
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
export interface Endpoint {
	id: string;
	name: string;
	hostname: string;
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
	id: string;
	name: string;
	type: 'gauge' | 'counter' | 'histogram';
	value: number;
	unit: string;
	labels: Record<string, string>;
	timestamp: string;
	endpointId?: string;
	endpoint_id?: string;
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
	type: string;
	source: string;
	message: string;
	description?: string;
	severity: 'info' | 'warning' | 'error' | 'critical';
	timestamp: string;
	endpointId?: string;
	endpoint_id?: string;
	endpoint_name?: string;
	user_name?: string;
	metadata?: Record<string, any>;
}

// Log types
export interface LogEntry {
	id: string;
	timestamp: string;
	level: 'debug' | 'info' | 'warning' | 'error' | 'fatal' | 'critical';
	message: string;
	source: string;
	category?: string;
	endpointId?: string;
	endpoint_id?: string;
	endpoint_name?: string;
	metadata?: Record<string, any>;
	meta?: Record<string, any>;
	tags?: Record<string, string>;
	target?: string;
	unit?: string;
	app_name?: string;
	service?: string;
	event_id?: string;
	user?: string;
	container_id?: string;
	container_name?: string;
	platform?: string;
	fields?: Record<string, any>;
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
