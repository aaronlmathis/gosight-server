import type { WidgetTemplate } from '$lib/types/dashboard';

// Widget templates that users can add to their dashboard
export const WIDGET_TEMPLATES: WidgetTemplate[] = [
	// Metric Cards
	{
		type: 'metric-card',
		name: 'Metric Card',
		description: 'Display a single metric value with optional thresholds',
		icon: 'activity',
		category: 'Metrics',
		defaultSize: { width: 3, height: 2 },
		defaultConfig: {
			refreshInterval: 30,
			showTitle: true,
			unit: '',
			threshold: { warning: 75, critical: 90 }
		}
	},
	{
		type: 'cpu-usage',
		name: 'CPU Usage',
		description: 'Real-time CPU usage monitoring',
		icon: 'cpu',
		category: 'System',
		defaultSize: { width: 4, height: 3 },
		defaultConfig: {
			refreshInterval: 10,
			showTitle: true,
			chartType: 'line',
			timeRange: '15m'
		}
	},
	{
		type: 'memory-usage',
		name: 'Memory Usage',
		description: 'Memory consumption tracking',
		icon: 'memory',
		category: 'System',
		defaultSize: { width: 4, height: 3 },
		defaultConfig: {
			refreshInterval: 10,
			showTitle: true,
			chartType: 'area',
			timeRange: '15m'
		}
	},
	{
		type: 'disk-usage',
		name: 'Disk Usage',
		description: 'Storage space monitoring',
		icon: 'hard-drive',
		category: 'System',
		defaultSize: { width: 3, height: 2 },
		defaultConfig: {
			refreshInterval: 60,
			showTitle: true,
			chartType: 'donut'
		}
	},
	{
		type: 'network-stats',
		name: 'Network Stats',
		description: 'Network I/O statistics',
		icon: 'wifi',
		category: 'System',
		defaultSize: { width: 6, height: 3 },
		defaultConfig: {
			refreshInterval: 15,
			showTitle: true,
			chartType: 'line',
			timeRange: '30m'
		}
	},

	// Charts
	{
		type: 'chart-line',
		name: 'Line Chart',
		description: 'Time series data visualization',
		icon: 'trending-up',
		category: 'Charts',
		defaultSize: { width: 6, height: 4 },
		defaultConfig: {
			refreshInterval: 30,
			showTitle: true,
			chartType: 'line',
			timeRange: '1h',
			metrics: []
		}
	},
	{
		type: 'chart-bar',
		name: 'Bar Chart',
		description: 'Categorical data comparison',
		icon: 'bar-chart-3',
		category: 'Charts',
		defaultSize: { width: 6, height: 4 },
		defaultConfig: {
			refreshInterval: 60,
			showTitle: true,
			chartType: 'bar'
		}
	},
	{
		type: 'chart-donut',
		name: 'Donut Chart',
		description: 'Proportional data display',
		icon: 'pie-chart',
		category: 'Charts',
		defaultSize: { width: 4, height: 4 },
		defaultConfig: {
			refreshInterval: 60,
			showTitle: true,
			chartType: 'donut'
		}
	},

	// Alerts & Events
	{
		type: 'active-alerts',
		name: 'Active Alerts',
		description: 'List of currently firing alerts',
		icon: 'alert-triangle',
		category: 'Monitoring',
		defaultSize: { width: 6, height: 4 },
		defaultConfig: {
			refreshInterval: 15,
			showTitle: true,
			maxItems: 10,
			alertLevels: ['critical', 'warning', 'info']
		}
	},
	{
		type: 'alert-summary',
		name: 'Alert Summary',
		description: 'Overview of alert counts by severity',
		icon: 'alert-circle',
		category: 'Monitoring',
		defaultSize: { width: 4, height: 3 },
		defaultConfig: {
			refreshInterval: 30,
			showTitle: true,
			chartType: 'donut'
		}
	},
	{
		type: 'recent-events',
		name: 'Recent Events',
		description: 'Latest system events and activities',
		icon: 'activity',
		category: 'Monitoring',
		defaultSize: { width: 6, height: 5 },
		defaultConfig: {
			refreshInterval: 20,
			showTitle: true,
			limit: 15,
			sortBy: 'timestamp',
			sortOrder: 'desc'
		}
	},
	{
		type: 'log-stream',
		name: 'Log Stream',
		description: 'Real-time log entries',
		icon: 'scroll-text',
		category: 'Monitoring',
		defaultSize: { width: 8, height: 6 },
		defaultConfig: {
			refreshInterval: 5,
			showTitle: true,
			limit: 50,
			sortBy: 'timestamp',
			sortOrder: 'desc'
		}
	},

	// Infrastructure
	{
		type: 'endpoint-health',
		name: 'Endpoint Health',
		description: 'Status overview of monitored endpoints',
		icon: 'monitor',
		category: 'Infrastructure',
		defaultSize: { width: 6, height: 4 },
		defaultConfig: {
			refreshInterval: 30,
			showTitle: true,
			chartType: 'donut'
		}
	},
	{
		type: 'container-stats',
		name: 'Container Stats',
		description: 'Docker container resource usage',
		icon: 'container',
		category: 'Infrastructure',
		defaultSize: { width: 8, height: 5 },
		defaultConfig: {
			refreshInterval: 20,
			showTitle: true,
			limit: 10,
			sortBy: 'cpu_percent',
			sortOrder: 'desc'
		}
	},
	{
		type: 'system-status',
		name: 'System Status',
		description: 'Overall system health indicators',
		icon: 'shield-check',
		category: 'Infrastructure',
		defaultSize: { width: 4, height: 3 },
		defaultConfig: {
			refreshInterval: 45,
			showTitle: true
		}
	},

	// Utilities
	{
		type: 'quick-links',
		name: 'Quick Links',
		description: 'Customizable navigation shortcuts',
		icon: 'external-link',
		category: 'Utilities',
		defaultSize: { width: 3, height: 4 },
		defaultConfig: {
			showTitle: true,
			links: [
				{ id: '1', title: 'Grafana', url: '/grafana', icon: 'bar-chart-3', description: 'Metrics dashboard' },
				{ id: '2', title: 'Logs', url: '/logs', icon: 'file-text', description: 'Log viewer' },
				{ id: '3', title: 'Alerts', url: '/alerts', icon: 'alert-triangle', description: 'Alert management' }
			]
		}
	}
];

// Get templates by category
export function getTemplatesByCategory(): Record<string, WidgetTemplate[]> {
	return WIDGET_TEMPLATES.reduce((acc, template) => {
		if (!acc[template.category]) {
			acc[template.category] = [];
		}
		acc[template.category].push(template);
		return acc;
	}, {} as Record<string, WidgetTemplate[]>);
}

// Get template by type
export function getTemplate(type: string): WidgetTemplate | undefined {
	return WIDGET_TEMPLATES.find(t => t.type === type);
}
