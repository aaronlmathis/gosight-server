// Dashboard widget types and configurations

export interface WidgetPosition {
	x: number;
	y: number;
	width: number;
	height: number;
}

export interface Widget {
	id: string;
	type: WidgetType;
	title: string;
	position: WidgetPosition;
	config: WidgetConfig;
	createdAt: string;
	updatedAt: string;
}

export type WidgetType = 
	| 'metric'
	| 'chart'
	| 'alerts_list'
	| 'events_list'
	| 'endpoint_count'
	| 'alert_count'
	| 'system_overview'
	| 'quick_links'
	| 'metric-card'
	| 'chart-line' 
	| 'chart-donut'
	| 'chart-bar'
	| 'recent-events'
	| 'active-alerts'
	| 'system-status'
	| 'cpu-usage'
	| 'memory-usage'
	| 'network-stats'
	| 'disk-usage'
	| 'container-stats'
	| 'endpoint-health'
	| 'alert-summary'
	| 'log-stream';

export interface WidgetConfig {
	// Common config
	refreshInterval?: number;
	showTitle?: boolean;
	
	// Metric card config
	metricType?: string;
	unit?: string;
	threshold?: {
		warning: number;
		critical: number;
	};
	
	// Metric specific properties
	namespace?: string;
	subnamespace?: string;
	metric?: string;
	endpointId?: string;
	
	// Event/Alert filtering
	category?: string;
	type?: string;
	severity?: string;
	status?: string;
	
	// Chart config
	chartType?: 'line' | 'area' | 'bar' | 'donut' | 'radial';
	timeRange?: string;
	dataSource?: string;
	metrics?: string[];
	
	// Quick links config
	links?: QuickLink[];
	
	// Alert config
	alertLevels?: string[];
	maxItems?: number;
	
	// List config
	limit?: number;
	sortBy?: string;
	sortOrder?: 'asc' | 'desc';
	
	// System overview config
	showCpuUsage?: boolean;
	showMemoryUsage?: boolean;
	showDiskUsage?: boolean;
	showUptime?: boolean;
	showLoadAverage?: boolean;
	
	// Endpoint count config
	showHostCount?: boolean;
	showContainerCount?: boolean;
	showTotalCount?: boolean;
	showOnlineStatus?: boolean;
	
	// Alert count config
	showActiveCount?: boolean;
	showBySeverity?: boolean;
	showRecentAlerts?: boolean;
	alertTimeRange?: string;
	
	// Colors and styling
	colors?: string[];
	backgroundColor?: string;
	textColor?: string;
}

export interface QuickLink {
	id: string;
	title: string;
	url: string;
	icon?: string;
	description?: string;
}

export interface Dashboard {
	id: string;
	name: string;
	description?: string;
	isDefault: boolean;
	widgets: Widget[];
	layout: DashboardLayout;
	createdAt: string;
	updatedAt: string;
}

export interface DashboardLayout {
	columns: number;
	rowHeight: number;
	margin: [number, number];
	padding: [number, number];
}

export interface WidgetTemplate {
	type: WidgetType;
	name: string;
	description: string;
	icon: string;
	category: string;
	defaultSize: Pick<WidgetPosition, 'width' | 'height'>;
	defaultConfig: WidgetConfig;
	preview?: string;
}

// Drag and drop types
export interface DragItem {
	type: 'widget' | 'template';
	widget?: Widget;
	template?: WidgetTemplate;
}

export interface DropResult {
	position: WidgetPosition;
	targetIndex?: number;
}

// Dashboard preferences that extend existing user preferences
export interface DashboardPreferences {
	dashboards: Dashboard[];
	defaultDashboardId: string;
	globalSettings: {
		autoRefresh: boolean;
		refreshInterval: number;
		showGrid: boolean;
		snapToGrid: boolean;
		theme: 'auto' | 'light' | 'dark';
	};
}

// Widget data types for real-time integration
export interface MetricDataPoint {
	timestamp: number;
	value: number;
	dimensions?: Record<string, string>;
}

export interface ChartSeries {
	name: string;
	data: Array<{ x: number; y: number }>;
	color?: string;
}

export interface Alert {
	id: string;
	rule_id: string;
	level: 'info' | 'warning' | 'error' | 'critical';
	message: string;
	timestamp: string;
	endpoint_id?: string;
	state: 'firing' | 'resolved';
	details?: Record<string, any>;
}

export interface Event {
	id: string;
	type: string;
	category: string;
	level: string;
	message: string;
	timestamp: string;
	endpoint_id?: string;
	details?: Record<string, any>;
}

export interface WidgetData {
	// Common fields
	status: 'success' | 'warning' | 'error' | 'unknown';
	error?: string;
	timestamp?: string;
	
	// Metric widget data
	value?: number;
	trend?: 'up' | 'down' | 'stable';
	unit?: string;
	thresholds?: {
		warning?: number;
		critical?: number;
	};
	
	// Chart widget data
	series?: ChartSeries[];
	
	// List widget data
	alerts?: Alert[];
	events?: Event[];
	
	// Count widget data
	details?: Record<string, any>;
	
	// System overview data
	metrics?: Record<string, number>;
	
	// Quick links data
	links?: QuickLink[];
}
