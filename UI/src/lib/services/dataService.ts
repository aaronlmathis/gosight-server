/**
 * Data Service for Dashboard Widgets
 * Handles live data integration, caching, and API interactions
 */
import { api } from '$lib/api';
import { websocketManager } from '$lib/websocket';
import { writable, derived, type Writable } from 'svelte/store';
import type { Widget, WidgetData, MetricDataPoint, Alert, Event } from '$lib/types/dashboard';

// Cache configuration
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
const METRICS_HISTORY_LIMIT = 100;

interface CacheEntry<T> {
	data: T;
	timestamp: number;
	stale: boolean;
}

interface MetricSubscription {
	widgetId: string;
	namespace: string;
	subnamespace: string;
	metric: string;
	endpointId?: string;
	callback: (data: MetricDataPoint[]) => void;
}

class DataService {
	private cache = new Map<string, CacheEntry<any>>();
	private metricSubscriptions = new Map<string, MetricSubscription>();
	private metricsHistory = new Map<string, MetricDataPoint[]>();
	
	// Real-time data stores
	public liveAlerts: Writable<Alert[]> = writable([]);
	public liveEvents: Writable<Event[]> = writable([]);
	public liveMetrics = writable<Record<string, MetricDataPoint[]>>({});
	
	// System overview stores
	public systemMetrics = writable<Record<string, number>>({});
	public endpointCounts = writable<{ hosts: number; containers: number; total: number }>({
		hosts: 0,
		containers: 0,
		total: 0
	});
	public alertCounts = writable<{ active: number; total: number; bySeverity: Record<string, number> }>({
		active: 0,
		total: 0,
		bySeverity: {}
	});

	constructor() {
		this.initializeWebSocketSubscriptions();
		this.loadInitialData();
	}

	/**
	 * Initialize WebSocket subscriptions for real-time updates
	 */
	private initializeWebSocketSubscriptions() {
		// Subscribe to alerts
		websocketManager.subscribeToAlerts((alert: Alert) => {
			this.liveAlerts.update(alerts => {
				const existing = alerts.findIndex(a => a.id === alert.id);
				if (existing >= 0) {
					alerts[existing] = alert;
				} else {
					alerts.unshift(alert);
				}
				return alerts.slice(0, 50); // Keep last 50 alerts
			});
			
			// Update alert counts
			this.refreshAlertCounts();
		});

		// Subscribe to events
		websocketManager.subscribeToEvents((event: Event) => {
			this.liveEvents.update(events => {
				events.unshift(event);
				return events.slice(0, 100); // Keep last 100 events
			});
		});

		// Subscribe to metrics
		websocketManager.subscribeToMetrics((metricsPayload: any) => {
			if (metricsPayload?.metrics && Array.isArray(metricsPayload.metrics)) {
				const timestamp = new Date(metricsPayload.timestamp).getTime();
				
				metricsPayload.metrics.forEach((metric: any) => {
					const metricKey = `${metric.namespace}.${metric.subnamespace}.${metric.name}`;
					const dataPoint: MetricDataPoint = {
						timestamp,
						value: parseFloat(metric.value),
						dimensions: metric.dimensions || {}
					};

					// Update metrics history
					const history = this.metricsHistory.get(metricKey) || [];
					history.unshift(dataPoint);
					this.metricsHistory.set(metricKey, history.slice(0, METRICS_HISTORY_LIMIT));

					// Notify subscribed widgets
					this.metricSubscriptions.forEach((subscription, subId) => {
						if (subscription.namespace === metric.namespace &&
							subscription.subnamespace === metric.subnamespace &&
							subscription.metric === metric.name) {
							
							// Filter by endpoint if specified
							if (!subscription.endpointId || 
								metricsPayload.endpoint_id === subscription.endpointId) {
								subscription.callback(history);
							}
						}
					});
				});

				// Update live metrics store
				this.liveMetrics.update(metrics => {
					metricsPayload.metrics.forEach((metric: any) => {
						const key = `${metric.namespace}.${metric.subnamespace}.${metric.name}`;
						if (!metrics[key]) metrics[key] = [];
						metrics[key].unshift({
							timestamp,
							value: parseFloat(metric.value),
							dimensions: metric.dimensions || {}
						});
						metrics[key] = metrics[key].slice(0, METRICS_HISTORY_LIMIT);
					});
					return metrics;
				});
			}
		});
	}

	/**
	 * Load initial data for dashboard
	 */
	private async loadInitialData() {
		try {
			await Promise.all([
				this.refreshSystemMetrics(),
				this.refreshEndpointCounts(),
				this.refreshAlertCounts(),
				this.loadRecentAlerts(),
				this.loadRecentEvents()
			]);
		} catch (error) {
			console.error('Failed to load initial dashboard data:', error);
		}
	}

	/**
	 * Get widget data with caching
	 */
	async getWidgetData(widget: Widget): Promise<WidgetData> {
		const cacheKey = this.generateCacheKey(widget);
		const cached = this.cache.get(cacheKey);
		
		// Return cached data if still valid
		if (cached && !this.isCacheStale(cached)) {
			return cached.data;
		}

		try {
			let data: WidgetData;

			switch (widget.type) {
				case 'metric':
					data = await this.getMetricData(widget);
					break;
				case 'chart':
					data = await this.getChartData(widget);
					break;
				case 'alerts_list':
					data = await this.getAlertsData(widget);
					break;
				case 'events_list':
					data = await this.getEventsData(widget);
					break;
				case 'endpoint_count':
					data = await this.getEndpointCountData();
					break;
				case 'alert_count':
					data = await this.getAlertCountData();
					break;
				case 'system_overview':
					data = await this.getSystemOverviewData();
					break;
				case 'quick_links':
					data = { links: widget.config?.links || [] };
					break;
				default:
					data = { value: 0, status: 'unknown' };
			}

			// Cache the result
			this.cache.set(cacheKey, {
				data,
				timestamp: Date.now(),
				stale: false
			});

			return data;
		} catch (error) {
			console.error(`Failed to get data for widget ${widget.id}:`, error);
			
			// Return cached data if available, even if stale
			if (cached) {
				return cached.data;
			}
			
			// Return default data
			return { value: 0, status: 'error', error: error instanceof Error ? error.message : 'Unknown error' };
		}
	}

	/**
	 * Subscribe to real-time metric updates for a widget
	 */
	subscribeToMetric(
		widgetId: string,
		namespace: string,
		subnamespace: string,
		metric: string,
		callback: (data: MetricDataPoint[]) => void,
		endpointId?: string
	): () => void {
		const subscriptionId = `${widgetId}-${namespace}-${subnamespace}-${metric}`;
		
		this.metricSubscriptions.set(subscriptionId, {
			widgetId,
			namespace,
			subnamespace,
			metric,
			endpointId,
			callback
		});

		// Send existing data if available
		const metricKey = `${namespace}.${subnamespace}.${metric}`;
		const history = this.metricsHistory.get(metricKey);
		if (history) {
			callback(history);
		}

		// Return unsubscribe function
		return () => {
			this.metricSubscriptions.delete(subscriptionId);
		};
	}

	/**
	 * Get metric data for metric widgets
	 */
	private async getMetricData(widget: Widget): Promise<WidgetData> {
		const { namespace, subnamespace, metric, endpointId } = widget.config || {};
		
		if (!namespace || !subnamespace || !metric) {
			return { value: 0, status: 'error', error: 'Missing metric configuration' };
		}

		try {
			// Get latest value
			const latest = await api.metrics.getMetricLatest(namespace, subnamespace, metric);
			const value = parseFloat(latest.value || 0);
			
			// Get recent history for trend calculation
			const historyData = await api.metrics.getMetricData(namespace, subnamespace, metric, {
				start: new Date(Date.now() - 60 * 60 * 1000).toISOString(), // Last hour
				endpointID: endpointId
			});

			const trend = this.calculateTrend(Array.isArray(historyData) ? historyData : []);
			const status = this.determineStatus(value, widget.config?.thresholds);

			return {
				value,
				trend,
				status,
				unit: widget.config?.unit || '',
				timestamp: latest.timestamp
			};
		} catch (error) {
			console.error(`Failed to get metric data for ${namespace}.${subnamespace}.${metric}:`, error);
			return { value: 0, status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch metric' };
		}
	}

	/**
	 * Get chart data for chart widgets
	 */
	private async getChartData(widget: Widget): Promise<WidgetData> {
		const { metrics, timeRange, endpointId } = widget.config || {};
		
		if (!metrics || !Array.isArray(metrics) || metrics.length === 0) {
			return { series: [], status: 'error', error: 'No metrics configured' };
		}

		try {
			const endTime = new Date();
			const startTime = new Date(endTime.getTime() - this.parseTimeRange(timeRange || '1h'));
			
			const seriesData = await Promise.all(
				metrics.map(async (metricConfig: any) => {
					const { namespace, subnamespace, metric, label } = metricConfig;
					
					const data = await api.metrics.getMetricData(namespace, subnamespace, metric, {
						start: startTime.toISOString(),
						end: endTime.toISOString(),
						endpointID: endpointId
					});

					const points = Array.isArray(data) ? data.map((point: any) => ({
						x: new Date(point.timestamp).getTime(),
						y: parseFloat(point.value || 0)
					})) : [];

					return {
						name: label || `${namespace}.${subnamespace}.${metric}`,
						data: points
					};
				})
			);

			return {
				series: seriesData,
				status: 'success'
			};
		} catch (error) {
			console.error('Failed to get chart data:', error);
			return { series: [], status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch chart data' };
		}
	}

	/**
	 * Get alerts data
	 */
	private async getAlertsData(widget: Widget): Promise<WidgetData> {
		try {
			const { limit = 10, severity, endpointId } = widget.config || {};
			
			const alerts = await api.alerts.getAll({
				limit,
				level: severity,
				endpoint_id: endpointId,
				sort: 'timestamp',
				order: 'desc'
			});

			return {
				alerts: Array.isArray(alerts) ? alerts : [],
				status: 'success'
			};
		} catch (error) {
			console.error('Failed to get alerts data:', error);
			return { alerts: [], status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch alerts' };
		}
	}

	/**
	 * Get events data
	 */
	private async getEventsData(widget: Widget): Promise<WidgetData> {
		try {
			const { limit = 10, category, endpointId } = widget.config || {};
			
			const events = await api.events.getAll({
				limit,
				category,
				endpoint_id: endpointId,
				sort: 'timestamp'
			});

			return {
				events: Array.isArray(events) ? events : [],
				status: 'success'
			};
		} catch (error) {
			console.error('Failed to get events data:', error);
			return { events: [], status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch events' };
		}
	}

	/**
	 * Get endpoint count data
	 */
	private async getEndpointCountData(): Promise<WidgetData> {
		try {
			const [hosts, containers] = await Promise.all([
				api.endpoints.getHosts(),
				api.endpoints.getContainers()
			]);

			const hostCount = Array.isArray(hosts) ? hosts.length : 0;
			const containerCount = Array.isArray(containers) ? containers.length : 0;
			const total = hostCount + containerCount;

			this.endpointCounts.set({ hosts: hostCount, containers: containerCount, total });

			return {
				value: total,
				details: { hosts: hostCount, containers: containerCount },
				status: 'success'
			};
		} catch (error) {
			console.error('Failed to get endpoint counts:', error);
			return { value: 0, status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch endpoints' };
		}
	}

	/**
	 * Get alert count data
	 */
	private async getAlertCountData(): Promise<WidgetData> {
		try {
			const [active, summary] = await Promise.all([
				api.alerts.getActive(),
				api.alerts.getSummary()
			]);

			const activeCount = Array.isArray(active) ? active.length : 0;
			const bySeverity = summary?.by_severity || {};
			const total = Object.values(bySeverity).reduce((sum, count) => sum + (count as number), 0);

			this.alertCounts.set({ active: activeCount, total, bySeverity });

			return {
				value: activeCount,
				details: bySeverity,
				status: activeCount > 0 ? 'warning' : 'success'
			};
		} catch (error) {
			console.error('Failed to get alert counts:', error);
			return { value: 0, status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch alerts' };
		}
	}

	/**
	 * Get system overview data
	 */
	private async getSystemOverviewData(): Promise<WidgetData> {
		try {
			const overview = await api.metrics.getSystemOverview();
			
			this.systemMetrics.set(overview);

			return {
				metrics: overview,
				status: 'success'
			};
		} catch (error) {
			console.error('Failed to get system overview:', error);
			return { metrics: {}, status: 'error', error: error instanceof Error ? error.message : 'Failed to fetch system overview' };
		}
	}

	/**
	 * Refresh system metrics
	 */
	private async refreshSystemMetrics() {
		try {
			const overview = await api.metrics.getSystemOverview();
			this.systemMetrics.set(overview);
		} catch (error) {
			console.error('Failed to refresh system metrics:', error);
		}
	}

	/**
	 * Refresh endpoint counts
	 */
	private async refreshEndpointCounts() {
		try {
			const data = await this.getEndpointCountData();
			if (data.details) {
				this.endpointCounts.set({
					hosts: data.details.hosts,
					containers: data.details.containers,
					total: data.value
				});
			}
		} catch (error) {
			console.error('Failed to refresh endpoint counts:', error);
		}
	}

	/**
	 * Refresh alert counts
	 */
	private async refreshAlertCounts() {
		try {
			const data = await this.getAlertCountData();
			if (data.details) {
				this.alertCounts.set({
					active: data.value,
					total: Object.values(data.details).reduce((sum, count) => sum + (count as number), 0),
					bySeverity: data.details
				});
			}
		} catch (error) {
			console.error('Failed to refresh alert counts:', error);
		}
	}

	/**
	 * Load recent alerts
	 */
	private async loadRecentAlerts() {
		try {
			const alerts = await api.alerts.getAll({ limit: 50, sort: 'timestamp', order: 'desc' });
			this.liveAlerts.set(Array.isArray(alerts) ? alerts : []);
		} catch (error) {
			console.error('Failed to load recent alerts:', error);
		}
	}

	/**
	 * Load recent events
	 */
	private async loadRecentEvents() {
		try {
			const events = await api.events.getRecent(100);
			this.liveEvents.set(Array.isArray(events) ? events : []);
		} catch (error) {
			console.error('Failed to load recent events:', error);
		}
	}

	/**
	 * Invalidate cache for a widget
	 */
	invalidateCache(widget: Widget) {
		const cacheKey = this.generateCacheKey(widget);
		this.cache.delete(cacheKey);
	}

	/**
	 * Clear all cache
	 */
	clearCache() {
		this.cache.clear();
	}

	/**
	 * Save widget configuration to backend
	 */
	async saveWidgetConfig(widgetId: string, config: any): Promise<void> {
		try {
			// Try to save to backend API
			const response = await fetch(`/api/v1/dashboard/widgets/${widgetId}/config`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify(config)
			});

			if (!response.ok) {
				console.warn(`Failed to save widget config to backend: ${response.status}`);
			}
		} catch (error) {
			console.warn('Failed to save widget config to backend:', error);
			// Configuration is still saved in the store, just not persisted to backend
		}
	}

	/**
	 * Save dashboard configuration to backend
	 */
	async saveDashboardConfig(dashboardId: string, config: any): Promise<void> {
		try {
			// Try to save to backend API
			const response = await fetch(`/api/v1/dashboard/dashboards/${dashboardId}/config`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify(config)
			});

			if (!response.ok) {
				console.warn(`Failed to save dashboard config to backend: ${response.status}`);
			}
		} catch (error) {
			console.warn('Failed to save dashboard config to backend:', error);
			// Configuration is still saved in the store, just not persisted to backend
		}
	}

	/**
	 * Clear dashboard configuration from backend
	 */
	async clearDashboardConfig(dashboardId: string): Promise<void> {
		try {
			// Try to clear from backend API
			const response = await fetch(`/api/v1/dashboard/dashboards/${dashboardId}/config`, {
				method: 'DELETE',
				credentials: 'include'
			});

			if (!response.ok) {
				console.warn(`Failed to clear dashboard config from backend: ${response.status}`);
			}
		} catch (error) {
			console.warn('Failed to clear dashboard config from backend:', error);
		}
	}

	/**
	 * Get available metrics for widget configuration
	 */
	async getAvailableMetrics(): Promise<Array<{ namespace: string; subnamespace: string; metric: string; description?: string }>> {
		try {
			const response = await fetch('/api/v1/metrics/available', {
				credentials: 'include'
			});

			if (response.ok) {
				return await response.json();
			}
			
			console.warn('Failed to fetch available metrics, using fallback');
			return this.getFallbackMetrics();
		} catch (error) {
			console.warn('Failed to fetch available metrics:', error);
			return this.getFallbackMetrics();
		}
	}

	/**
	 * Get available endpoints for widget configuration
	 */
	async getAvailableEndpoints(): Promise<Array<{ id: string; name: string; type: string; status: string }>> {
		try {
			const response = await fetch('/api/v1/endpoints', {
				credentials: 'include'
			});

			if (response.ok) {
				const data = await response.json();
				return data.endpoints || [];
			}
			
			console.warn('Failed to fetch available endpoints, using fallback');
			return this.getFallbackEndpoints();
		} catch (error) {
			console.warn('Failed to fetch available endpoints:', error);
			return this.getFallbackEndpoints();
		}
	}

	/**
	 * Fallback metrics when API is unavailable
	 */
	private getFallbackMetrics() {
		return [
			{ namespace: 'system', subnamespace: 'cpu', metric: 'usage_percent', description: 'CPU Usage Percentage' },
			{ namespace: 'system', subnamespace: 'memory', metric: 'usage_percent', description: 'Memory Usage Percentage' },
			{ namespace: 'system', subnamespace: 'disk', metric: 'usage_percent', description: 'Disk Usage Percentage' },
			{ namespace: 'system', subnamespace: 'network', metric: 'bytes_sent', description: 'Network Bytes Sent' },
			{ namespace: 'system', subnamespace: 'network', metric: 'bytes_received', description: 'Network Bytes Received' },
			{ namespace: 'application', subnamespace: 'response', metric: 'time_ms', description: 'Response Time (ms)' },
			{ namespace: 'application', subnamespace: 'requests', metric: 'per_second', description: 'Requests per Second' },
			{ namespace: 'container', subnamespace: 'cpu', metric: 'usage_percent', description: 'Container CPU Usage' },
			{ namespace: 'container', subnamespace: 'memory', metric: 'usage_bytes', description: 'Container Memory Usage' }
		];
	}

	/**
	 * Fallback endpoints when API is unavailable
	 */
	private getFallbackEndpoints() {
		return [
			{ id: 'localhost', name: 'Local Host', type: 'host', status: 'online' },
			{ id: 'web-server-01', name: 'Web Server 01', type: 'container', status: 'online' },
			{ id: 'database-01', name: 'Database 01', type: 'container', status: 'online' }
		];
	}

	/**
	 * Generate cache key for widget
	 */
	private generateCacheKey(widget: Widget): string {
		const config = widget.config || {};
		const configStr = JSON.stringify(config);
		return `${widget.type}-${widget.id}-${configStr}`;
	}

	/**
	 * Check if cache entry is stale
	 */
	private isCacheStale(entry: CacheEntry<any>): boolean {
		return entry.stale || (Date.now() - entry.timestamp > CACHE_DURATION);
	}

	/**
	 * Calculate trend from historical data
	 */
	private calculateTrend(data: MetricDataPoint[]): 'up' | 'down' | 'stable' {
		if (!data || data.length < 2) return 'stable';
		
		const recent = data.slice(0, Math.min(10, data.length));
		const older = data.slice(Math.min(10, data.length), Math.min(20, data.length));
		
		if (recent.length === 0 || older.length === 0) return 'stable';
		
		const recentAvg = recent.reduce((sum, point) => sum + point.value, 0) / recent.length;
		const olderAvg = older.reduce((sum, point) => sum + point.value, 0) / older.length;
		
		const threshold = Math.abs(olderAvg) * 0.05; // 5% threshold
		
		if (recentAvg > olderAvg + threshold) return 'up';
		if (recentAvg < olderAvg - threshold) return 'down';
		return 'stable';
	}

	/**
	 * Determine status based on value and thresholds
	 */
	private determineStatus(value: number, thresholds?: { warning?: number; critical?: number }): 'success' | 'warning' | 'error' {
		if (!thresholds) return 'success';
		
		if (thresholds.critical !== undefined && value >= thresholds.critical) {
			return 'error';
		}
		
		if (thresholds.warning !== undefined && value >= thresholds.warning) {
			return 'warning';
		}
		
		return 'success';
	}

	/**
	 * Parse time range string to milliseconds
	 */
	private parseTimeRange(timeRange: string): number {
		const match = timeRange.match(/(\d+)([smhd])/);
		if (!match) return 3600000; // Default 1 hour
		
		const value = parseInt(match[1]);
		const unit = match[2];
		
		const multipliers = {
			s: 1000,
			m: 60000,
			h: 3600000,
			d: 86400000
		};
		
		return value * (multipliers[unit as keyof typeof multipliers] || 3600000);
	}
}

// Export singleton instance
export const dataService = new DataService();
