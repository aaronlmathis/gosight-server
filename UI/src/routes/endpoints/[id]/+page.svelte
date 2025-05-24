<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import { metricsWS, eventsWS, logsWS, alertsWS } from '$lib/websocket';
	import type { Endpoint, Metric, Event, Log, Alert } from '$lib/types';
	import { formatDate, formatBytes, formatDuration } from '$lib/utils';
	import { ChevronLeft, Activity, AlertTriangle, Database, FileText, Settings } from 'lucide-svelte';

	let endpoint: Endpoint | null = null;
	let metrics: Metric[] = [];
	let events: Event[] = [];
	let logs: Log[] = [];
	let alerts: Alert[] = [];
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	const endpointId = $page.params.id;

	let metricsChart: any;
	let unsubscribeMetrics: (() => void) | null = null;
	let unsubscribeEvents: (() => void) | null = null;
	let unsubscribeLogs: (() => void) | null = null;
	let unsubscribeAlerts: (() => void) | null = null;

	onMount(async () => {
		await loadEndpointData();
		setupRealTimeUpdates();
		initMetricsChart();
	});

	onDestroy(() => {
		if (unsubscribeMetrics) unsubscribeMetrics();
		if (unsubscribeEvents) unsubscribeEvents();
		if (unsubscribeLogs) unsubscribeLogs();
		if (unsubscribeAlerts) unsubscribeAlerts();
		if (metricsChart) metricsChart.destroy();
	});

	async function loadEndpointData() {
		try {
			loading = true;
			const [endpointRes, metricsRes, eventsRes, logsRes, alertsRes] = await Promise.all([
				api.getEndpoint(endpointId),
				api.getMetrics({ endpoint_id: endpointId, limit: 100 }),
				api.getEvents({ endpoint_id: endpointId, limit: 50 }),
				api.getLogs({ endpoint_id: endpointId, limit: 50 }),
				api.getAlerts({ endpoint_id: endpointId, limit: 20 })
			]);

			endpoint = endpointRes.data;
			metrics = metricsRes.data;
			events = eventsRes.data;
			logs = logsRes.data;
			alerts = alertsRes.data;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load endpoint data';
		} finally {
			loading = false;
		}
	}

	function setupRealTimeUpdates() {
		unsubscribeMetrics = metricsWS.messages.subscribe((messages) => {
			const latest = messages[messages.length - 1];
			if (latest && latest.endpoint_id === endpointId) {
				metrics = [latest, ...metrics.slice(0, 99)];
				updateMetricsChart(latest);
			}
		});

		unsubscribeEvents = eventsWS.messages.subscribe((messages) => {
			const latest = messages[messages.length - 1];
			if (latest && latest.endpoint_id === endpointId) {
				events = [latest, ...events.slice(0, 49)];
			}
		});

		unsubscribeLogs = logsWS.messages.subscribe((messages) => {
			const latest = messages[messages.length - 1];
			if (latest && latest.endpoint_id === endpointId) {
				logs = [latest, ...logs.slice(0, 49)];
			}
		});

		unsubscribeAlerts = alertsWS.messages.subscribe((messages) => {
			const latest = messages[messages.length - 1];
			if (latest && latest.endpoint_id === endpointId) {
				alerts = [latest, ...alerts.slice(0, 19)];
			}
		});
	}

	function initMetricsChart() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		const options = {
			chart: {
				type: 'line',
				height: 300,
				animations: { enabled: true },
				toolbar: { show: false }
			},
			series: [
				{
					name: 'CPU Usage',
					data: metrics.filter(m => m.name === 'cpu_usage').map(m => ({
						x: new Date(m.timestamp).getTime(),
						y: m.value
					}))
				},
				{
					name: 'Memory Usage',
					data: metrics.filter(m => m.name === 'memory_usage').map(m => ({
						x: new Date(m.timestamp).getTime(),
						y: m.value
					}))
				}
			],
			xaxis: {
				type: 'datetime',
				labels: { format: 'HH:mm' }
			},
			yaxis: {
				labels: { formatter: (val: number) => val.toFixed(1) + '%' }
			},
			colors: ['#3b82f6', '#10b981'],
			stroke: { curve: 'smooth', width: 2 },
			theme: { mode: 'light' }
		};

		metricsChart = new window.ApexCharts(document.querySelector('#metrics-chart'), options);
		metricsChart.render();
	}

	function updateMetricsChart(metric: Metric) {
		if (!metricsChart) return;

		const seriesIndex = metric.name === 'cpu_usage' ? 0 : metric.name === 'memory_usage' ? 1 : -1;
		if (seriesIndex >= 0) {
			metricsChart.appendData([{
				data: [{
					x: new Date(metric.timestamp).getTime(),
					y: metric.value
				}]
			}], seriesIndex);
		}
	}

	async function runCommand(command: string) {
		try {
			await api.sendCommand(endpointId, { command, args: [] });
		} catch (err) {
			console.error('Failed to run command:', err);
		}
	}
</script>

<svelte:head>
	<title>Endpoint {endpoint?.name || endpointId} - GoSight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 shadow">
		<div class="px-4 sm:px-6 lg:px-8">
			<div class="flex items-center justify-between h-16">
				<div class="flex items-center">
					<a href="/endpoints" class="mr-4 p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700">
						<ChevronLeft size={20} />
					</a>
					<div>
						<h1 class="text-xl font-semibold text-gray-900 dark:text-white">
							{endpoint?.name || 'Loading...'}
						</h1>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{endpoint?.ip_address || ''}
						</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getBadgeClass(endpoint?.status || 'unknown')}">
						{endpoint?.status || 'Unknown'}
					</span>
					{#if endpoint?.os}
						<span class="text-sm text-gray-500 dark:text-gray-400">{endpoint.os}</span>
					{/if}
				</div>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="p-6">
			<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
				<p class="text-red-800 dark:text-red-200">{error}</p>
			</div>
		</div>
	{:else if endpoint}
		<!-- Tabs -->
		<div class="border-b border-gray-200 dark:border-gray-700">
			<nav class="px-4 sm:px-6 lg:px-8 -mb-px flex space-x-8">
				{#each [
					{ id: 'overview', label: 'Overview', icon: Activity },
					{ id: 'metrics', label: 'Metrics', icon: Database },
					{ id: 'events', label: 'Events', icon: FileText },
					{ id: 'logs', label: 'Logs', icon: FileText },
					{ id: 'alerts', label: 'Alerts', icon: AlertTriangle },
					{ id: 'settings', label: 'Settings', icon: Settings }
				] as tab}
					<button
						class="py-4 px-1 border-b-2 font-medium text-sm {activeTab === tab.id 
							? 'border-blue-500 text-blue-600 dark:text-blue-400' 
							: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
						on:click={() => activeTab = tab.id}
					>
						<div class="flex items-center">
							<svelte:component this={tab.icon} size={16} class="mr-2" />
							{tab.label}
						</div>
					</button>
				{/each}
			</nav>
		</div>

		<!-- Content -->
		<div class="p-6">
			{#if activeTab === 'overview'}
				<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
					<!-- Info Cards -->
					<div class="lg:col-span-2 space-y-6">
						<!-- Basic Info -->
						<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">System Information</h3>
							<dl class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Hostname</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.hostname || 'N/A'}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">IP Address</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.ip_address}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Operating System</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.os || 'N/A'}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Agent Version</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.agent_version || 'N/A'}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Seen</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.last_seen ? formatDate(endpoint.last_seen) : 'N/A'}</dd>
								</div>
								<div>
									<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Uptime</dt>
									<dd class="text-sm text-gray-900 dark:text-white">{endpoint.uptime ? formatDuration(endpoint.uptime) : 'N/A'}</dd>
								</div>
							</dl>
						</div>

						<!-- Metrics Chart -->
						<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Performance Metrics</h3>
							<div id="metrics-chart"></div>
						</div>
					</div>

					<!-- Sidebar -->
					<div class="space-y-6">
						<!-- Quick Actions -->
						<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Quick Actions</h3>
							<div class="space-y-2">
								<button 
									class="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
									on:click={() => runCommand('restart')}
								>
									Restart Service
								</button>
								<button 
									class="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
									on:click={() => runCommand('status')}
								>
									Check Status
								</button>
								<button 
									class="w-full text-left px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
									on:click={() => runCommand('update')}
								>
									Update Agent
								</button>
							</div>
						</div>

						<!-- Recent Alerts -->
						<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Recent Alerts</h3>
							<div class="space-y-3">
								{#each alerts.slice(0, 5) as alert}
									<div class="flex items-center space-x-3 p-2 rounded-lg {alert.severity === 'critical' ? 'bg-red-50 dark:bg-red-900/20' : alert.severity === 'warning' ? 'bg-yellow-50 dark:bg-yellow-900/20' : 'bg-blue-50 dark:bg-blue-900/20'}">
										<AlertTriangle size={16} class="{alert.severity === 'critical' ? 'text-red-500' : alert.severity === 'warning' ? 'text-yellow-500' : 'text-blue-500'}" />
										<div class="flex-1 min-w-0">
											<p class="text-xs font-medium text-gray-900 dark:text-white truncate">{alert.title}</p>
											<p class="text-xs text-gray-500 dark:text-gray-400">{formatDate(alert.created_at)}</p>
										</div>
									</div>
								{:else}
									<p class="text-sm text-gray-500 dark:text-gray-400">No recent alerts</p>
								{/each}
							</div>
						</div>
					</div>
				</div>

			{:else if activeTab === 'metrics'}
				<!-- Metrics content would go here -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Detailed Metrics</h3>
					<p class="text-gray-500 dark:text-gray-400">Detailed metrics view coming soon...</p>
				</div>

			{:else if activeTab === 'events'}
				<!-- Events content -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white">Recent Events</h3>
					</div>
					<div class="divide-y divide-gray-200 dark:divide-gray-700">
						{#each events as event}
							<div class="px-6 py-4">
								<div class="flex items-center justify-between">
									<div>
										<h4 class="text-sm font-medium text-gray-900 dark:text-white">{event.type}</h4>
										<p class="text-sm text-gray-500 dark:text-gray-400">{event.description}</p>
									</div>
									<div class="text-right">
										<p class="text-xs text-gray-500 dark:text-gray-400">{formatDate(event.timestamp)}</p>
									</div>
								</div>
							</div>
						{:else}
							<div class="px-6 py-8 text-center">
								<p class="text-gray-500 dark:text-gray-400">No events found</p>
							</div>
						{/each}
					</div>
				</div>

			{:else if activeTab === 'logs'}
				<!-- Logs content -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white">Recent Logs</h3>
					</div>
					<div class="divide-y divide-gray-200 dark:divide-gray-700">
						{#each logs as log}
							<div class="px-6 py-3">
								<div class="flex items-start space-x-3">
									<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium {getBadgeClass(log.level)}">
										{log.level}
									</span>
									<div class="flex-1 min-w-0">
										<p class="text-sm text-gray-900 dark:text-white break-words">{log.message}</p>
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{formatDate(log.timestamp)}</p>
									</div>
								</div>
							</div>
						{:else}
							<div class="px-6 py-8 text-center">
								<p class="text-gray-500 dark:text-gray-400">No logs found</p>
							</div>
						{/each}
					</div>
				</div>

			{:else if activeTab === 'alerts'}
				<!-- Alerts content -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white">Alerts</h3>
					</div>
					<div class="divide-y divide-gray-200 dark:divide-gray-700">
						{#each alerts as alert}
							<div class="px-6 py-4">
								<div class="flex items-center justify-between">
									<div class="flex-1">
										<div class="flex items-center space-x-2">
											<h4 class="text-sm font-medium text-gray-900 dark:text-white">{alert.title}</h4>
											<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium {getBadgeClass(alert.severity)}">
												{alert.severity}
											</span>
											<span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium {getBadgeClass(alert.status)}">
												{alert.status}
											</span>
										</div>
										<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{alert.description}</p>
									</div>
									<div class="text-right">
										<p class="text-xs text-gray-500 dark:text-gray-400">{formatDate(alert.created_at)}</p>
									</div>
								</div>
							</div>
						{:else}
							<div class="px-6 py-8 text-center">
								<p class="text-gray-500 dark:text-gray-400">No alerts found</p>
							</div>
						{/each}
					</div>
				</div>

			{:else if activeTab === 'settings'}
				<!-- Settings content -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Endpoint Settings</h3>
					<p class="text-gray-500 dark:text-gray-400">Settings panel coming soon...</p>
				</div>
			{/if}
		</div>
	{/if}
</div>
