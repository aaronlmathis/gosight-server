<script lang="ts">
	import { onMount } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import { api } from '$lib/api';
	import { alertsWS, eventsWS, metricsWS } from '$lib/websocket';
	import {
		activeAlertsCount,
		realtimeCounters,
		alertCountStore,
		endpointCountStore,
		containerCountStore,
		eventCountStore
	} from '$lib/stores';
	import { formatNumber, getStatusBadgeClass } from '$lib/utils';
	import type { Alert, Endpoint, Event, Metric } from '$lib/types';

	let alerts: Alert[] = [];
	let endpoints: Endpoint[] = [];
	let recentEvents: Event[] = [];
	let systemMetrics: Record<string, number> = {};
	let topCpuContainers: any[] = [];
	let topMemoryContainers: any[] = [];

	// Chart instances
	let alertsChart: any;
	let agentsChart: any;
	let containersChart: any;
	let loadChart: any;
	let cpuContainersChart: any;
	let memoryContainersChart: any;

	onMount(async () => {
		await loadDashboardData();
		initializeCharts();

		// Subscribe to real-time updates
		alertsWS.messages.subscribe(handleAlertsUpdate);
		eventsWS.messages.subscribe(handleEventsUpdate);
		metricsWS.messages.subscribe(handleMetricsUpdate);
	});

	async function loadDashboardData() {
		try {
			// Load all dashboard data in parallel
			const [alertsRes, endpointsRes, eventsRes, metricsRes] = await Promise.all([
				api.alerts.getActive(),
				api.endpoints.getAll(),
				api.events.getRecent(),
				api.metrics.getSystemOverview()
			]);

			alerts = (alertsRes as any).data || [];
			endpoints = (endpointsRes as any).data || [];
			recentEvents = (eventsRes as any).data || [];
			systemMetrics = (metricsRes as any).data || {};

			// Update stores
			alertCountStore.set(alerts.length);
			endpointCountStore.set(endpoints.length);

			// Load container data
			const containersRes = await api.endpoints.getContainers();
			const containers = (containersRes as any).data || [];
			containerCountStore.set(containers.length);

			// Get top containers by resource usage
			topCpuContainers = containers
				.sort((a: any, b: any) => (b.cpuPercent || 0) - (a.cpuPercent || 0))
				.slice(0, 10);

			topMemoryContainers = containers
				.sort((a: any, b: any) => (b.memoryPercent || 0) - (a.memoryPercent || 0))
				.slice(0, 10);
		} catch (error) {
			console.error('Failed to load dashboard data:', error);
		}
	}

	function initializeCharts() {
		// Initialize ApexCharts for each metric
		initAlertsChart();
		initAgentsChart();
		initContainersChart();
		initLoadChart();
		initTopContainersCharts();
	}

	function initAlertsChart() {
		const activeAlerts = alerts.filter((a) => a.status === 'active').length;
		const resolvedAlerts = alerts.filter((a) => a.status === 'resolved').length;

		const options = {
			series: [activeAlerts],
			chart: {
				height: 180,
				type: 'radialBar'
			},
			plotOptions: {
				radialBar: {
					hollow: {
						size: '50%'
					},
					dataLabels: {
						value: {
							show: true,
							fontSize: '16px',
							fontWeight: 600
						},
						name: {
							show: false
						}
					}
				}
			},
			fill: {
				colors: ['#ef4444'] // red-500
			},
			labels: ['Active Alerts']
		};

		if (typeof ApexCharts !== 'undefined') {
			const element = document.querySelector('#radial-alerts');
			if (element) {
				alertsChart = new ApexCharts(element, options);
				alertsChart.render();
			}
		}
	}

	function initAgentsChart() {
		const onlineAgents = endpoints.filter((e) => e.status === 'online').length;
		const totalAgents = endpoints.length;
		const percentage = totalAgents > 0 ? Math.round((onlineAgents / totalAgents) * 100) : 0;

		const options = {
			series: [percentage],
			chart: {
				height: 180,
				type: 'radialBar'
			},
			plotOptions: {
				radialBar: {
					hollow: {
						size: '50%'
					},
					dataLabels: {
						value: {
							show: true,
							fontSize: '16px',
							fontWeight: 600
						},
						name: {
							show: false
						}
					}
				}
			},
			fill: {
				colors: ['#10b981'] // green-500
			},
			labels: ['Online']
		};

		if (typeof ApexCharts !== 'undefined') {
			const element = document.querySelector('#radial-agents');
			if (element) {
				agentsChart = new ApexCharts(element, options);
				agentsChart.render();
			}
		}
	}

	function initContainersChart() {
		const runningContainers = $realtimeCounters.containers.running || 0;

		const options = {
			series: [runningContainers],
			chart: {
				height: 180,
				type: 'radialBar'
			},
			plotOptions: {
				radialBar: {
					hollow: {
						size: '50%'
					},
					dataLabels: {
						value: {
							show: true,
							fontSize: '16px',
							fontWeight: 600
						},
						name: {
							show: false
						}
					}
				}
			},
			fill: {
				colors: ['#3b82f6'] // blue-500
			},
			labels: ['Running']
		};

		if (typeof ApexCharts !== 'undefined') {
			const element = document.querySelector('#radial-containers');
			if (element) {
				containersChart = new ApexCharts(element, options);
				containersChart.render();
			}
		}
	}

	function initLoadChart() {
		const cpuLoad = systemMetrics.cpu_percent || 0;

		const options = {
			series: [Math.round(cpuLoad)],
			chart: {
				height: 180,
				type: 'radialBar'
			},
			plotOptions: {
				radialBar: {
					hollow: {
						size: '50%'
					},
					dataLabels: {
						value: {
							show: true,
							fontSize: '16px',
							fontWeight: 600
						},
						name: {
							show: false
						}
					}
				}
			},
			fill: {
				colors: ['#f59e0b'] // yellow-500
			},
			labels: ['CPU %']
		};

		if (typeof ApexCharts !== 'undefined') {
			const element = document.querySelector('#radial-load');
			if (element) {
				loadChart = new ApexCharts(element, options);
				loadChart.render();
			}
		}
	}

	function initTopContainersCharts() {
		// Top CPU containers
		const cpuOptions = {
			series: [
				{
					data: topCpuContainers.map((c) => ({
						x: c.name,
						y: c.cpuPercent || 0
					}))
				}
			],
			chart: {
				type: 'bar',
				height: 200
			},
			plotOptions: {
				bar: {
					horizontal: true
				}
			},
			colors: ['#3b82f6'],
			xaxis: {
				title: {
					text: 'CPU %'
				}
			}
		};

		// Top Memory containers
		const memOptions = {
			series: [
				{
					data: topMemoryContainers.map((c) => ({
						x: c.name,
						y: c.memoryPercent || 0
					}))
				}
			],
			chart: {
				type: 'bar',
				height: 200
			},
			plotOptions: {
				bar: {
					horizontal: true
				}
			},
			colors: ['#10b981'],
			xaxis: {
				title: {
					text: 'Memory %'
				}
			}
		};

		if (typeof ApexCharts !== 'undefined') {
			const cpuElement = document.querySelector('#top-cpu-containers');
			if (cpuElement) {
				cpuContainersChart = new ApexCharts(cpuElement, cpuOptions);
				cpuContainersChart.render();
			}

			const memElement = document.querySelector('#top-mem-containers');
			if (memElement) {
				memoryContainersChart = new ApexCharts(memElement, memOptions);
				memoryContainersChart.render();
			}
		}
	}

	function handleAlertsUpdate(data: any) {
		// Update alerts and refresh chart
		alerts = data.alerts || [];
		alertCountStore.set(alerts.length);

		if (alertsChart) {
			const activeAlerts = alerts.filter((a) => a.status === 'active').length;
			alertsChart.updateSeries([activeAlerts]);
		}
	}

	function handleEventsUpdate(data: any) {
		// Add new events to the top of the list
		if (data.event) {
			recentEvents = [data.event, ...recentEvents.slice(0, 19)]; // Keep only 20 events
		}
	}

	function handleMetricsUpdate(data: any) {
		// Update system metrics and refresh charts
		if (data.metrics) {
			systemMetrics = { ...systemMetrics, ...data.metrics };

			if (loadChart && data.metrics.cpu_percent !== undefined) {
				loadChart.updateSeries([Math.round(data.metrics.cpu_percent)]);
			}
		}
	}

	function formatEventTime(timestamp: string): string {
		const now = new Date();
		const eventTime = new Date(timestamp);
		const diff = now.getTime() - eventTime.getTime();
		const minutes = Math.floor(diff / 60000);

		if (minutes < 1) return 'Just now';
		if (minutes < 60) return `${minutes}m ago`;

		const hours = Math.floor(minutes / 60);
		if (hours < 24) return `${hours}h ago`;

		const days = Math.floor(hours / 24);
		return `${days}d ago`;
	}

	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'critical':
				return 'text-red-600 dark:text-red-400';
			case 'error':
				return 'text-red-500 dark:text-red-400';
			case 'warning':
				return 'text-yellow-500 dark:text-yellow-400';
			case 'info':
				return 'text-blue-500 dark:text-blue-400';
			default:
				return 'text-gray-500 dark:text-gray-400';
		}
	}
</script>

<svelte:head>
	<title>Dashboard - GoSight</title>
</svelte:head>

<PermissionGuard requiredPermission="gosight:dashboard:view">
	<section class="space-y-6 p-4">
		<div class="mb-6">
			<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Welcome to GoSight</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				Monitoring real-time metrics, logs, and alerts across your infrastructure.
			</p>
		</div>

		<!-- Summary Row -->
		<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
			<!-- LEFT: Summary Cards in 2x3 Subgrid -->
			<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
				<!-- Alerts -->
				<a href="/alerts" class="block rounded-lg transition hover:ring-2 hover:ring-blue-400">
					<div
						class="flex h-44 items-center justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm transition duration-200 hover:border-gray-200 hover:shadow-md dark:border-gray-700 dark:bg-gray-900 dark:hover:border-gray-600"
					>
						<div class="flex w-full flex-col justify-center truncate pl-3 text-sm">
							<div class="font-medium whitespace-nowrap text-gray-700 dark:text-gray-200">
								Active Alerts
							</div>
							<div class="mt-1 text-base font-semibold text-red-600 dark:text-red-400">
								{alerts.filter((a) => a.status === 'active').length || 'None'}
							</div>
						</div>
						<div class="h-[180px] w-[220px]">
							<div id="radial-alerts" class="h-full w-full"></div>
						</div>
					</div>
				</a>

				<!-- Agents -->
				<a href="/endpoints" class="block rounded-lg transition hover:ring-2 hover:ring-blue-400">
					<div
						class="flex h-44 items-center justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm transition duration-200 hover:border-gray-200 hover:shadow-md dark:border-gray-700 dark:bg-gray-900 dark:hover:border-gray-600"
					>
						<div class="flex w-full flex-col justify-center truncate pl-3 text-sm">
							<div class="font-medium whitespace-nowrap text-gray-700 dark:text-gray-200">
								Host Endpoints
							</div>
							<div class="mt-1 text-base font-semibold text-gray-600 dark:text-gray-400">
								{endpoints.filter((e) => e.status === 'online').length}/{endpoints.length} Online
							</div>
						</div>
						<div class="h-[180px] w-[220px]">
							<div id="radial-agents" class="h-full w-full"></div>
						</div>
					</div>
				</a>

				<!-- Containers -->
				<a
					href="/endpoints#containers"
					class="block rounded-lg transition hover:ring-2 hover:ring-blue-400"
				>
					<div
						class="flex h-44 items-center justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm transition duration-200 hover:border-gray-200 hover:shadow-md dark:border-gray-700 dark:bg-gray-900 dark:hover:border-gray-600"
					>
						<div class="flex w-full flex-col justify-center truncate pl-3 text-sm">
							<div class="font-medium whitespace-nowrap text-gray-700 dark:text-gray-200">
								Container Endpoints
							</div>
							<div class="mt-1 text-base font-semibold text-gray-600 dark:text-gray-400">
								{$realtimeCounters.containers.running} Running
							</div>
						</div>
						<div class="h-[180px] w-[220px]">
							<div id="radial-containers" class="h-full w-full"></div>
						</div>
					</div>
				</a>

				<!-- CPU Load -->
				<a href="/metrics" class="block rounded-lg transition hover:ring-2 hover:ring-blue-400">
					<div
						class="flex h-44 items-center justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm transition duration-200 hover:border-gray-200 hover:shadow-md dark:border-gray-700 dark:bg-gray-900 dark:hover:border-gray-600"
					>
						<div class="flex w-full flex-col justify-center truncate pl-3 text-sm">
							<div class="font-medium whitespace-nowrap text-gray-700 dark:text-gray-200">
								System Load
							</div>
							<div class="mt-1 text-base font-semibold text-gray-600 dark:text-gray-400">
								{Math.round(systemMetrics.cpu_percent || 0)}% CPU
							</div>
						</div>
						<div class="h-[180px] w-[220px]">
							<div id="radial-load" class="h-full w-full"></div>
						</div>
					</div>
				</a>

				<!-- Top container by CPU -->
				<div
					class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
				>
					<div class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-200">
						Top Containers by CPU Usage
					</div>
					<div id="top-cpu-containers" class="h-52"></div>
				</div>

				<!-- Top container by Memory -->
				<div
					class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
				>
					<div class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-200">
						Top Containers by Memory Usage
					</div>
					<div id="top-mem-containers" class="h-52"></div>
				</div>
			</div>

			<!-- Recent Events -->
			<div
				class="h-full rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
			>
				<div
					class="mb-2 border-b border-gray-100 pb-2 font-semibold text-gray-800 dark:border-gray-700 dark:text-gray-100"
				>
					Recent Events
				</div>
				<div
					class="min-h-[300px] space-y-2 overflow-y-auto text-sm text-gray-800 dark:text-gray-300"
				>
					{#each recentEvents as event}
						<div
							class="flex items-start space-x-2 rounded p-2 hover:bg-gray-50 dark:hover:bg-gray-800"
						>
							<div
								class="mt-2 h-2 w-2 flex-shrink-0 rounded-full {getSeverityColor(event.severity)}"
							></div>
							<div class="min-w-0 flex-1">
								<div class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{event.message}
								</div>
								<div class="flex items-center space-x-2 text-xs text-gray-500 dark:text-gray-400">
									<span>{event.source}</span>
									<span>•</span>
									<span>{formatEventTime(event.timestamp)}</span>
								</div>
							</div>
						</div>
					{:else}
						<div class="text-center text-gray-500 dark:text-gray-400 py-8">No recent events</div>
					{/each}
				</div>
			</div>
		</div>
	</section>
</PermissionGuard>

<style>
	:global(.apexcharts-datalabel-value) {
		font-size: 1em !important;
	}
</style>
