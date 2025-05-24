<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import type { Endpoint, Metric, Event, LogEntry, Alert } from '$lib/types';
	import {
		formatDate,
		formatBytes,
		formatDuration,
		getStatusBadgeClass,
		getLevelBadgeClass
	} from '$lib/utils';
	import {
		ChevronLeft,
		Activity,
		AlertTriangle,
		Database,
		FileText,
		Settings,
		Cpu,
		HardDrive,
		Wifi,
		Monitor,
		Terminal,
		ScrollText
	} from 'lucide-svelte';

	let endpoint: Endpoint | null = null;
	let metrics: Metric[] = [];
	let events: Event[] = [];
	let logs: LogEntry[] = [];
	let alerts: Alert[] = [];
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	const endpointId = $page.params.id;

	// Chart instances
	let overviewCharts: any = {};
	let computeCharts: any = {};
	let networkCharts: any = {};
	let diskCharts: any = {};

	let unsubscribeMetrics: (() => void) | null = null;
	let unsubscribeEvents: (() => void) | null = null;
	let unsubscribeLogs: (() => void) | null = null;
	let unsubscribeAlerts: (() => void) | null = null;

	onMount(async () => {
		await loadEndpointData();
		setupRealTimeUpdates();
		// Initialize Overview charts since it's the default active tab
		setTimeout(() => {
			initOverviewCharts();
		}, 100);
	});

	onDestroy(() => {
		if (unsubscribeMetrics) unsubscribeMetrics();
		if (unsubscribeEvents) unsubscribeEvents();
		if (unsubscribeLogs) unsubscribeLogs();
		if (unsubscribeAlerts) unsubscribeAlerts();

		// Disconnect websockets
		websocketManager.disconnect();

		// Cleanup all charts
		Object.values(overviewCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(computeCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(networkCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(diskCharts).forEach((chart: any) => chart?.destroy?.());
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
			// Handle API responses based on their actual structure:
			// - metrics: Array directly or in .data property
			// - events: Array directly (as you provided)
			// - logs: Object with logs property containing array (as you provided)
			// - alerts: Array directly or in .data/.alerts property
			metrics = Array.isArray(metricsRes)
				? (metricsRes as Metric[])
				: Array.isArray((metricsRes as any)?.data)
					? ((metricsRes as any).data as Metric[])
					: [];

			events = Array.isArray(eventsRes)
				? (eventsRes as Event[])
				: Array.isArray((eventsRes as any)?.data)
					? ((eventsRes as any).data as Event[])
					: [];

			logs = Array.isArray((logsRes as any)?.logs)
				? ((logsRes as any).logs as LogEntry[])
				: Array.isArray(logsRes)
					? (logsRes as LogEntry[])
					: Array.isArray((logsRes as any)?.data)
						? ((logsRes as any).data as LogEntry[])
						: [];

			alerts = Array.isArray(alertsRes)
				? (alertsRes as Alert[])
				: Array.isArray((alertsRes as any)?.alerts)
					? ((alertsRes as any).alerts as Alert[])
					: Array.isArray((alertsRes as any)?.data)
						? ((alertsRes as any).data as Alert[])
						: [];

			console.log('Raw API responses:', {
				endpoint: endpointRes,
				metrics: metricsRes,
				events: eventsRes,
				logs: logsRes,
				alerts: alertsRes
			});

			console.log('Loaded data:', {
				endpoint: endpoint?.id,
				metricsCount: metrics.length,
				eventsCount: events.length,
				logsCount: logs.length,
				alertsCount: alerts.length,
				sampleEvent: events[0]
			});
		} catch (err) {
			console.error('Error loading endpoint data:', err);
			error = err instanceof Error ? err.message : 'Failed to load endpoint data';
			// Initialize empty arrays on error
			metrics = [];
			events = [];
			logs = [];
			alerts = [];
		} finally {
			loading = false;
		}
	}

	function switchTab(tabId: string) {
		activeTab = tabId;
		// Initialize charts for the specific tab
		setTimeout(() => {
			if (tabId === 'overview') initOverviewCharts();
			else if (tabId === 'compute') initComputeCharts();
			else if (tabId === 'network') initNetworkCharts();
			else if (tabId === 'disk') initDiskCharts();
			else if (tabId === 'console') initConsole();
		}, 100);
	}

	function initOverviewCharts() {
		console.log('Initializing overview charts...');
		if (typeof window === 'undefined' || !window.ApexCharts) {
			console.log('ApexCharts not available, skipping chart initialization');
			return;
		}

		// Initialize chart data storage
		if (!(window as any).chartData) {
			(window as any).chartData = {
				cpu: [],
				memory: [],
				swap: [],
				cpu_mini: [],
				memory_mini: [],
				swap_mini: [],
				compute_cpu: []
			};
		}

		// Main Performance Metrics Chart
		if (!overviewCharts.main) {
			console.log('Creating main performance chart...');

			// Check if container exists
			const container = document.querySelector('#metrics-chart');
			if (!container) {
				console.error('Main chart container #metrics-chart not found');
				return;
			}

			// Prepare initial data from existing metrics
			const initialCpuData = metrics
				.filter((m) => m.name === 'cpu_usage')
				.slice(-20) // Last 20 points
				.map((m) => ({
					x: new Date(m.timestamp).getTime(),
					y: m.value
				}));

			const initialMemoryData = metrics
				.filter((m) => m.name === 'memory_usage')
				.slice(-20) // Last 20 points
				.map((m) => ({
					x: new Date(m.timestamp).getTime(),
					y: m.value
				}));

			// Store initial data
			(window as any).chartData.cpu = initialCpuData;
			(window as any).chartData.memory = initialMemoryData;

			const mainOptions = {
				chart: {
					type: 'line',
					height: 300,
					toolbar: { show: false },
					animations: {
						enabled: true,
						easing: 'linear',
						speed: 800,
						animateGradually: {
							enabled: true,
							delay: 150
						},
						dynamicAnimation: {
							enabled: true,
							speed: 350
						}
					}
				},
				series: [
					{
						name: 'CPU Usage %',
						data: initialCpuData
					},
					{
						name: 'Memory Usage %',
						data: initialMemoryData
					}
				],
				xaxis: {
					type: 'datetime',
					labels: { format: 'HH:mm' },
					range: 600000 // Show last 10 minutes
				},
				yaxis: {
					labels: { formatter: (val: number) => val.toFixed(1) + '%' },
					min: 0,
					max: 100
				},
				colors: ['#3b82f6', '#10b981'],
				stroke: { curve: 'smooth', width: 2 },
				legend: { show: true },
				tooltip: { shared: true, intersect: false },
				grid: {
					show: true,
					borderColor: '#374151',
					strokeDashArray: 0,
					position: 'back',
					xaxis: {
						lines: {
							show: false
						}
					},
					yaxis: {
						lines: {
							show: true
						}
					}
				}
			};

			try {
				overviewCharts.main = new window.ApexCharts(container, mainOptions);
				overviewCharts.main.render();
				console.log('Main performance chart created and rendered');
			} catch (err) {
				console.error('Error creating main chart:', err);
			}
		} else {
			console.log('Main chart already exists');
		}

		// Mini CPU Chart
		if (!overviewCharts.cpu) {
			const cpuContainer = document.querySelector('#miniCpuChart');
			if (!cpuContainer) {
				console.error('CPU mini chart container #miniCpuChart not found');
				return;
			}

			const cpuOptions = {
				chart: {
					type: 'area',
					height: 80,
					toolbar: { show: false },
					animations: { enabled: true },
					sparkline: { enabled: true }
				},
				series: [{ name: 'CPU', data: [] }],
				xaxis: { type: 'datetime', labels: { show: false } },
				yaxis: { show: false, min: 0, max: 100 },
				colors: ['#3b82f6'],
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.7,
						opacityTo: 0.1,
						stops: [0, 100]
					}
				},
				grid: { show: false },
				tooltip: { enabled: false }
			};

			try {
				overviewCharts.cpu = new window.ApexCharts(cpuContainer, cpuOptions);
				overviewCharts.cpu.render();
				console.log('Mini CPU chart created');
			} catch (err) {
				console.error('Error creating mini CPU chart:', err);
			}
		}

		// Mini Memory Chart
		if (!overviewCharts.memory) {
			const memContainer = document.querySelector('#miniMemoryChart');
			if (!memContainer) {
				console.error('Memory mini chart container #miniMemoryChart not found');
				return;
			}

			const memOptions = {
				chart: {
					type: 'area',
					height: 80,
					toolbar: { show: false },
					animations: { enabled: true },
					sparkline: { enabled: true }
				},
				series: [{ name: 'Memory', data: [] }],
				xaxis: { type: 'datetime', labels: { show: false } },
				yaxis: { show: false, min: 0, max: 100 },
				colors: ['#10b981'],
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.7,
						opacityTo: 0.1,
						stops: [0, 100]
					}
				},
				grid: { show: false },
				tooltip: { enabled: false }
			};

			try {
				overviewCharts.memory = new window.ApexCharts(memContainer, memOptions);
				overviewCharts.memory.render();
				console.log('Mini Memory chart created');
			} catch (err) {
				console.error('Error creating mini Memory chart:', err);
			}
		}

		// Mini Swap Chart
		if (!overviewCharts.swap) {
			const swapContainer = document.querySelector('#miniSwapChart');
			if (!swapContainer) {
				console.error('Swap mini chart container #miniSwapChart not found');
				return;
			}

			const swapOptions = {
				chart: {
					type: 'area',
					height: 80,
					toolbar: { show: false },
					animations: { enabled: true },
					sparkline: { enabled: true }
				},
				series: [{ name: 'Swap', data: [] }],
				xaxis: { type: 'datetime', labels: { show: false } },
				yaxis: { show: false, min: 0, max: 100 },
				colors: ['#f87171'],
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.7,
						opacityTo: 0.1,
						stops: [0, 100]
					}
				},
				grid: { show: false },
				tooltip: { enabled: false }
			};

			try {
				overviewCharts.swap = new window.ApexCharts(swapContainer, swapOptions);
				overviewCharts.swap.render();
				console.log('Mini Swap chart created');
			} catch (err) {
				console.error('Error creating mini Swap chart:', err);
			}
		}

		console.log('Overview charts initialization complete');
	}

	function initComputeCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// CPU Usage Chart
		if (!computeCharts.cpuUsage) {
			const cpuUsageOptions = {
				chart: {
					type: 'area',
					height: 250,
					toolbar: { show: false },
					animations: { enabled: true }
				},
				series: [{ name: 'CPU Usage %', data: [] }],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm' } },
				yaxis: { labels: { formatter: (val: number) => val.toFixed(1) + '%' } },
				colors: ['#3b82f6'],
				stroke: { curve: 'smooth', width: 2 },
				fill: { type: 'gradient' }
			};
			computeCharts.cpuUsage = new window.ApexCharts(
				document.querySelector('#cpuUsageChart'),
				cpuUsageOptions
			);
			computeCharts.cpuUsage.render();
		}

		// CPU Load Chart
		if (!computeCharts.cpuLoad) {
			const cpuLoadOptions = {
				chart: {
					type: 'line',
					height: 250,
					toolbar: { show: false },
					animations: { enabled: true }
				},
				series: [
					{ name: '1m', data: [] },
					{ name: '5m', data: [] },
					{ name: '15m', data: [] }
				],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm' } },
				colors: ['#3b82f6', '#10b981', '#f59e0b'],
				stroke: { curve: 'smooth', width: 2 }
			};
			computeCharts.cpuLoad = new window.ApexCharts(
				document.querySelector('#cpuLoadChart'),
				cpuLoadOptions
			);
			computeCharts.cpuLoad.render();
		}
	}

	function initNetworkCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// Network Traffic Chart
		if (!networkCharts.traffic) {
			const trafficOptions = {
				chart: {
					type: 'area',
					height: 300,
					toolbar: { show: false },
					animations: { enabled: true }
				},
				series: [
					{ name: 'Upload (Mbps)', data: [] },
					{ name: 'Download (Mbps)', data: [] }
				],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm' } },
				colors: ['#ef4444', '#3b82f6'],
				stroke: { curve: 'smooth', width: 2 },
				fill: { type: 'gradient' }
			};
			networkCharts.traffic = new window.ApexCharts(
				document.querySelector('#networkTrafficChart'),
				trafficOptions
			);
			networkCharts.traffic.render();
		}
	}

	function initDiskCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// Disk Usage Donut Chart
		if (!diskCharts.usage) {
			const diskUsageOptions = {
				chart: { type: 'donut', height: 400 },
				series: [0, 100],
				labels: ['Used', 'Free'],
				colors: ['#ef4444', '#10b981'],
				plotOptions: {
					pie: {
						donut: {
							size: '70%',
							labels: {
								show: true,
								total: {
									show: true,
									label: 'Total Space',
									formatter: () => 'Loading...'
								}
							}
						}
					}
				}
			};
			diskCharts.usage = new window.ApexCharts(
				document.querySelector('#diskUsageDonutChart'),
				diskUsageOptions
			);
			diskCharts.usage.render();
		}
	}

	function setupRealTimeUpdates() {
		console.log('Setting up real-time updates for endpoint:', endpointId);

		// Connect websockets with endpoint filtering (like the original implementation)
		websocketManager.connect(endpointId);

		// Subscribe to metrics updates
		unsubscribeMetrics = websocketManager.subscribeToMetrics((metricsPayload) => {
			console.log('Received metric update:', metricsPayload);
			if (metricsPayload && metricsPayload.endpoint_id === endpointId) {
				// The payload contains an array of metrics in the 'metrics' property
				if (Array.isArray(metricsPayload.metrics)) {
					// Ensure metrics is an array before using slice
					const existingMetrics = Array.isArray(metrics) ? metrics : [];
					metrics = [metricsPayload, ...existingMetrics.slice(0, 99)];
					updateCharts(metricsPayload);
				}
			}
		});

		// Subscribe to events updates
		unsubscribeEvents = websocketManager.subscribeToEvents((latestEvent) => {
			console.log('Received event update:', latestEvent);
			if (latestEvent && latestEvent.endpoint_id === endpointId) {
				// Ensure events is an array before using slice
				const existingEvents = Array.isArray(events) ? events : [];
				events = [latestEvent, ...existingEvents.slice(0, 49)];
			}
		});

		// Subscribe to logs updates
		unsubscribeLogs = websocketManager.subscribeToLogs((latestLog) => {
			console.log('Received log update:', latestLog);
			if (latestLog && latestLog.endpoint_id === endpointId) {
				// Ensure logs is an array before using slice
				const existingLogs = Array.isArray(logs) ? logs : [];
				logs = [latestLog, ...existingLogs.slice(0, 49)];
			}
		});

		// Subscribe to alerts updates
		unsubscribeAlerts = websocketManager.subscribeToAlerts((latestAlert) => {
			console.log('Received alert update:', latestAlert);
			if (latestAlert && latestAlert.endpoint_id === endpointId) {
				// Ensure alerts is an array before using slice
				const existingAlerts = Array.isArray(alerts) ? alerts : [];
				alerts = [latestAlert, ...existingAlerts.slice(0, 19)];
			}
		});
	}

	function updateCharts(metricsPayload: any) {
		if (!metricsPayload || !Array.isArray(metricsPayload.metrics)) {
			console.log('No metrics array found in payload');
			return;
		}

		const timestamp = new Date(metricsPayload.timestamp).getTime();
		console.log('Updating charts with timestamp:', timestamp);
		console.log('Available chart instances:', {
			main: !!overviewCharts.main,
			cpu: !!overviewCharts.cpu,
			memory: !!overviewCharts.memory,
			swap: !!overviewCharts.swap
		});

		// Store current data for charts (using global state)
		if (!(window as any).chartData) {
			(window as any).chartData = {
				cpu: [],
				memory: [],
				swap: []
			};
		}

		// Process each metric in the array
		metricsPayload.metrics.forEach((metric: any) => {
			const metricValue = parseFloat(metric.value);
			const metricName = metric.name;

			console.log(`Processing metric: ${metricName} = ${metricValue}`);

			// Update percentage labels in the UI
			if (metricName === 'cpu_usage') {
				const label = document.getElementById('cpu-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
			}
			if (metricName === 'memory_usage') {
				const label = document.getElementById('mem-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
			}
			if (metricName === 'swap_usage') {
				const label = document.getElementById('swap-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
			}

			// Update main performance metrics chart in overview
			if (overviewCharts.main && (metricName === 'cpu_usage' || metricName === 'memory_usage')) {
				console.log('Updating main chart for metric:', metricName);

				if (metricName === 'cpu_usage') {
					// Store CPU data
					(window as any).chartData.cpu.push({ x: timestamp, y: metricValue });
					(window as any).chartData.cpu = (window as any).chartData.cpu.slice(-50); // Keep last 50 points
				}
				if (metricName === 'memory_usage') {
					// Store Memory data
					(window as any).chartData.memory.push({ x: timestamp, y: metricValue });
					(window as any).chartData.memory = (window as any).chartData.memory.slice(-50); // Keep last 50 points
				}

				// Always update both series to maintain chart integrity
				try {
					overviewCharts.main.updateSeries(
						[
							{
								name: 'CPU Usage %',
								data: (window as any).chartData.cpu
							},
							{
								name: 'Memory Usage %',
								data: (window as any).chartData.memory
							}
						],
						false
					); // false = don't redraw each update
					console.log(
						'Updated main chart series with',
						(window as any).chartData.cpu.length,
						'CPU points and',
						(window as any).chartData.memory.length,
						'memory points'
					);
				} catch (err) {
					console.error('Error updating main chart:', err);
				}
			}

			// Update overview mini charts using updateSeries (more reliable than appendData)
			if (metricName === 'cpu_usage' && overviewCharts.cpu) {
				console.log('Updating mini CPU chart');
				try {
					(window as any).chartData.cpu_mini = (window as any).chartData.cpu_mini || [];
					(window as any).chartData.cpu_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.cpu_mini = (window as any).chartData.cpu_mini.slice(-20); // Keep last 20 points

					overviewCharts.cpu.updateSeries(
						[
							{
								name: 'CPU',
								data: (window as any).chartData.cpu_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini CPU chart:', err);
				}
			}
			if (metricName === 'memory_usage' && overviewCharts.memory) {
				console.log('Updating mini Memory chart');
				try {
					(window as any).chartData.memory_mini = (window as any).chartData.memory_mini || [];
					(window as any).chartData.memory_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.memory_mini = (window as any).chartData.memory_mini.slice(-20); // Keep last 20 points

					overviewCharts.memory.updateSeries(
						[
							{
								name: 'Memory',
								data: (window as any).chartData.memory_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini Memory chart:', err);
				}
			}
			if (metricName === 'swap_usage' && overviewCharts.swap) {
				console.log('Updating mini Swap chart');
				try {
					(window as any).chartData.swap_mini = (window as any).chartData.swap_mini || [];
					(window as any).chartData.swap_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.swap_mini = (window as any).chartData.swap_mini.slice(-20); // Keep last 20 points

					overviewCharts.swap.updateSeries(
						[
							{
								name: 'Swap',
								data: (window as any).chartData.swap_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini Swap chart:', err);
				}
			}

			// Update compute charts
			if (metricName === 'cpu_usage' && computeCharts.cpuUsage) {
				try {
					(window as any).chartData.compute_cpu = (window as any).chartData.compute_cpu || [];
					(window as any).chartData.compute_cpu.push({ x: timestamp, y: metricValue });
					(window as any).chartData.compute_cpu = (window as any).chartData.compute_cpu.slice(-50);

					computeCharts.cpuUsage.updateSeries(
						[
							{
								name: 'CPU Usage %',
								data: (window as any).chartData.compute_cpu
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating compute CPU chart:', err);
				}
			}
		});

		// Force chart redraws with proper ApexCharts methods
		setTimeout(() => {
			try {
				// Use updateOptions to force a refresh instead of redrawPaths
				if (overviewCharts.main) {
					overviewCharts.main.updateOptions({}, true, true); // redrawPaths: true, animate: true
				}
				if (overviewCharts.cpu) {
					overviewCharts.cpu.updateOptions({}, true, true);
				}
				if (overviewCharts.memory) {
					overviewCharts.memory.updateOptions({}, true, true);
				}
				if (overviewCharts.swap) {
					overviewCharts.swap.updateOptions({}, true, true);
				}
				if (computeCharts.cpuUsage) {
					computeCharts.cpuUsage.updateOptions({}, true, true);
				}
			} catch (err) {
				console.error('Error refreshing charts:', err);
			}
		}, 50);
	}

	async function runCommand(command: string) {
		try {
			await api.sendCommand(endpointId, { command, args: [] });
		} catch (err) {
			console.error('Failed to run command:', err);
		}
	}

	function getBadgeClass(value: string): string {
		// Try status class first, then level class
		const statusClass = getStatusBadgeClass(value);
		if (statusClass !== 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200') {
			return statusClass;
		}
		return getLevelBadgeClass(value);
	}

	// Initialize console when tab is switched
	function initConsole() {
		setTimeout(() => {
			const input = document.getElementById('console-command') as HTMLInputElement;
			const responsesEl = document.getElementById('console-responses');

			if (!input || !responsesEl || input.dataset.bound) return;
			input.dataset.bound = 'true';

			input.addEventListener('keydown', async (e) => {
				if (e.key !== 'Enter') return;

				const cmd = input.value.trim();
				if (!cmd) return;

				// Echo command
				const echo = document.createElement('div');
				echo.innerHTML = `<span class="text-blue-400">user</span>@<span class="text-purple-400">${endpoint?.hostname || 'host'}</span>:<span class="text-red-400">~</span>$ <span class="text-green-400">${cmd}</span>`;
				responsesEl.appendChild(echo);

				// Show placeholder while waiting
				const pending = document.createElement('div');
				pending.className = 'text-gray-500 whitespace-pre-wrap';
				pending.textContent = '[executing...]';
				responsesEl.appendChild(pending);

				input.value = '';

				try {
					await runCommand(cmd);
					// Response will arrive via websocket
				} catch (err) {
					pending.className = 'text-red-400';
					pending.textContent = '❌ ' + (err instanceof Error ? err.message : 'Command failed');
				}

				// Scroll to bottom
				const output = document.getElementById('console-output');
				if (output) output.scrollTop = output.scrollHeight;
			});
		}, 100);
	}
</script>

<svelte:head>
	<title>Endpoint {endpoint?.name || endpointId} - GoSight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white shadow dark:bg-gray-800">
		<div class="px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 items-center justify-between">
				<div class="flex items-center">
					<a href="/endpoints" class="mr-4 rounded-md p-2 hover:bg-gray-100 dark:hover:bg-gray-700">
						<ChevronLeft size={20} />
					</a>
					<div>
						<h1 class="text-xl font-semibold text-gray-900 dark:text-white">
							{endpoint?.hostname || 'Loading...'}
						</h1>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{endpoint?.ip || ''}
						</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span
						class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getBadgeClass(
							endpoint?.status || 'unknown'
						)}"
					>
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
		<div class="flex h-64 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="p-6">
			<div
				class="rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
			>
				<p class="text-red-800 dark:text-red-200">{error}</p>
			</div>
		</div>
	{:else if endpoint}
		<!-- Tabs -->
		<div class="border-b border-gray-200 dark:border-gray-700">
			<nav class="-mb-px flex space-x-8 px-4 sm:px-6 lg:px-8" id="dashboardTabs" role="tablist">
				{#each [{ id: 'overview', label: 'Overview', icon: Activity }, { id: 'compute', label: 'Compute', icon: Cpu }, { id: 'disk', label: 'Disk', icon: HardDrive }, { id: 'network', label: 'Network', icon: Wifi }, { id: 'activity', label: 'Activity', icon: Monitor }, { id: 'logs', label: 'Logs', icon: ScrollText }, { id: 'console', label: 'Console', icon: Terminal }] as tab}
					<button
						class="border-b-2 px-1 py-4 text-sm font-medium {activeTab === tab.id
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'}"
						on:click={() => switchTab(tab.id)}
						role="tab"
						aria-controls={tab.id}
						aria-selected={activeTab === tab.id}
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
		<div class="bg-gray-50 dark:bg-gray-800" id="dashboardTabContent">
			<!-- OVERVIEW TAB -->
			{#if activeTab === 'overview'}
				<div class="p-4" id="overview" role="tabpanel" aria-labelledby="overview-tab">
					<!-- System Info and Metrics Section -->
					<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
						<!-- Info Cards -->
						<div class="space-y-6 lg:col-span-2">
							<!-- Basic Info -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									System Information
								</h3>
								<dl class="grid grid-cols-1 gap-4 sm:grid-cols-2">
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Hostname</dt>
										<dd class="text-sm text-gray-900 dark:text-white">
											{endpoint.hostname || 'N/A'}
										</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">IP Address</dt>
										<dd class="text-sm text-gray-900 dark:text-white">{endpoint.ip}</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
											Operating System
										</dt>
										<dd class="text-sm text-gray-900 dark:text-white">{endpoint.os || 'N/A'}</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
											Agent Version
										</dt>
										<dd class="text-sm text-gray-900 dark:text-white">
											{endpoint.agent_version || 'N/A'}
										</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Seen</dt>
										<dd class="text-sm text-gray-900 dark:text-white">
											{endpoint.last_seen ? formatDate(endpoint.last_seen) : 'N/A'}
										</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Uptime</dt>
										<dd class="text-sm text-gray-900 dark:text-white">
											{endpoint.uptime ? formatDuration(endpoint.uptime) : 'N/A'}
										</dd>
									</div>
								</dl>
							</div>

							<!-- Metrics Chart -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Performance Metrics
								</h3>
								<div id="metrics-chart"></div>
							</div>
						</div>

						<!-- Sidebar -->
						<div class="space-y-6">
							<!-- Quick Actions -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Quick Actions
								</h3>
								<div class="space-y-2">
									<button
										class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
										on:click={() => runCommand('restart')}
									>
										Restart Service
									</button>
									<button
										class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
										on:click={() => runCommand('status')}
									>
										Check Status
									</button>
									<button
										class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
										on:click={() => runCommand('update')}
									>
										Update Agent
									</button>
								</div>
							</div>

							<!-- Recent Alerts -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Recent Alerts
								</h3>
								<div class="space-y-3">
									{#each Array.isArray(alerts) ? alerts.slice(0, 5) : [] as alert}
										<div
											class="flex items-center space-x-3 rounded-lg p-2 {alert.level === 'critical'
												? 'bg-red-50 dark:bg-red-900/20'
												: alert.level === 'warning'
													? 'bg-yellow-50 dark:bg-yellow-900/20'
													: 'bg-blue-50 dark:bg-blue-900/20'}"
										>
											<AlertTriangle
												size={16}
												class={alert.level === 'critical'
													? 'text-red-500'
													: alert.level === 'warning'
														? 'text-yellow-500'
														: 'text-blue-500'}
											/>
											<div class="min-w-0 flex-1">
												<p class="truncate text-xs font-medium text-gray-900 dark:text-white">
													{alert.message || alert.title || alert.name}
												</p>
												<p class="text-xs text-gray-500 dark:text-gray-400">
													{formatDate(alert.last_fired || alert.created_at || new Date())}
												</p>
											</div>
										</div>
									{:else}
										<p class="text-sm text-gray-500 dark:text-gray-400">No recent alerts</p>
									{/each}
								</div>
							</div>
						</div>
					</div>

					<!-- Metrics Row: CPU, Memory, Swap -->
					<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
						<div
							class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 dark:text-gray-400">CPU Usage</p>
								<p
									class="text-2xl font-bold text-indigo-600 dark:text-blue-400"
									id="cpu-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniCpuChart" class="h-20 w-full"></div>
							</div>
						</div>

						<div
							class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 dark:text-gray-400">Memory Used</p>
								<p
									class="text-2xl font-bold text-green-600 dark:text-green-400"
									id="mem-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniMemoryChart" class="h-20 w-full"></div>
							</div>
						</div>

						<div
							class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 dark:text-gray-400">Swap Used</p>
								<p
									class="text-2xl font-bold text-yellow-500 dark:text-yellow-400"
									id="swap-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniSwapChart" class="h-20 w-full"></div>
							</div>
						</div>
					</div>

					<!-- Live Logs Section -->
					<div class="mb-6 grid grid-cols-1 gap-4">
						<div
							class="flex h-96 flex-col rounded-lg border border-gray-200 bg-white p-4 shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-900"
						>
							<div class="mb-3 flex items-center justify-between">
								<h3 class="text-base font-semibold text-gray-800 dark:text-white">Live Logs</h3>
								<span class="text-xs text-gray-500 dark:text-gray-400">Last 10 entries</span>
							</div>
							<div
								id="log-stream"
								class="h-full space-y-2 overflow-y-auto rounded-md border border-gray-200 bg-gray-50 p-3 font-mono text-sm break-words whitespace-pre-wrap text-gray-800 shadow-inner dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
							>
								{#each Array.isArray(logs) ? logs.slice(0, 10) : [] as log}
									<div class="text-xs">
										<span class="text-gray-500 dark:text-gray-400"
											>[{formatDate(log.timestamp)}]</span
										>
										<span class="font-medium {getBadgeClass(log.level)}">{log.level}</span>
										<span class="ml-2">{log.message}</span>
									</div>
								{:else}
									<div class="text-gray-500 dark:text-gray-400">No logs available</div>
								{/each}
							</div>
						</div>
					</div>

					<!-- Top Processes -->
					<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2">
						<div
							class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-700 dark:bg-gray-900"
						>
							<h3 class="text-md mb-2 font-semibold text-gray-800 dark:text-white">
								Top 5 Running Processes by CPU
							</h3>
							<div class="overflow-x-auto">
								<table
									id="cpu-table"
									class="w-full text-left text-sm text-gray-700 dark:text-gray-200"
								>
									<thead
										class="bg-gray-100 text-xs text-gray-700 uppercase dark:bg-gray-700 dark:text-gray-300"
									>
										<tr>
											<th class="px-3 py-2 text-center" scope="col">PID</th>
											<th class="px-3 py-2" scope="col">User</th>
											<th class="px-3 py-2 text-right" scope="col">CPU %</th>
											<th class="px-3 py-2" scope="col">Command</th>
										</tr>
									</thead>
									<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
										<tr
											><td colspan="4" class="px-3 py-4 text-center text-gray-500"
												>Loading processes...</td
											></tr
										>
									</tbody>
								</table>
							</div>
						</div>

						<div
							class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-700 dark:bg-gray-900"
						>
							<h3 class="text-md mb-2 font-semibold text-gray-800 dark:text-white">
								Top 5 Running Processes by Memory
							</h3>
							<div class="overflow-x-auto">
								<table
									id="mem-table"
									class="w-full text-left text-sm text-gray-700 dark:text-gray-200"
								>
									<thead
										class="bg-gray-100 text-xs text-gray-700 uppercase dark:bg-gray-700 dark:text-gray-300"
									>
										<tr>
											<th class="px-3 py-2 text-center" scope="col">PID</th>
											<th class="px-3 py-2" scope="col">User</th>
											<th class="px-3 py-2 text-right" scope="col">MEM %</th>
											<th class="px-3 py-2" scope="col">Command</th>
										</tr>
									</thead>
									<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
										<tr
											><td colspan="4" class="px-3 py-4 text-center text-gray-500"
												>Loading processes...</td
											></tr
										>
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>

				<!-- COMPUTE TAB -->
			{:else if activeTab === 'compute'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="compute"
					role="tabpanel"
					aria-labelledby="compute-tab"
				>
					<div class="space-y-8">
						<div class="grid w-full grid-cols-4 gap-4">
							<!-- Left side: CPU charts (3/4 width) -->
							<div class="col-span-3 flex flex-col gap-4">
								<!-- CPU Usage Chart -->
								<div
									class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
								>
									<div class="mb-2 flex items-start justify-between">
										<h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-300">
											CPU Usage Over Time
										</h3>
										<div id="cpuUsageLabel" class="text-right">
											<p
												id="label-cpu-percent"
												class="text-2xl leading-tight font-bold text-blue-600 dark:text-blue-400"
											>
												0.0%
											</p>
											<p class="text-xs text-gray-500 dark:text-gray-400">CPU Usage</p>
										</div>
									</div>
									<div class="relative w-full">
										<div id="cpuUsageChart" class="h-[250px] w-full"></div>
									</div>
								</div>

								<!-- CPU Load Chart -->
								<div
									class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
								>
									<div class="mb-2 flex items-start justify-between">
										<h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-300">
											CPU Load Over Time
										</h3>
										<div id="cpuLoadLabel" class="space-y-1 text-right">
											<p class="text-sm font-semibold text-gray-600 dark:text-gray-300">
												Load Averages
											</p>
											<p class="text-xs text-gray-500 dark:text-gray-400">
												1m: <span
													id="label-load-1"
													class="font-bold text-blue-600 dark:text-blue-400">0.00</span
												>
												· 5m:
												<span
													id="label-load-5"
													class="font-bold text-emerald-600 dark:text-emerald-400">0.00</span
												>
												· 15m:
												<span
													id="label-load-15"
													class="font-bold text-amber-600 dark:text-amber-400">0.00</span
												>
											</p>
										</div>
									</div>
									<div class="relative w-full">
										<div id="cpuLoadChart" class="h-[250px] w-full"></div>
									</div>
								</div>
							</div>

							<!-- Right side: CPU Time Counters (1/4 width) -->
							<div class="col-span-1">
								<div
									class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
								>
									<h4
										class="mb-4 border-b border-gray-200 pb-1 text-sm font-semibold text-gray-700 dark:border-gray-700 dark:text-gray-100"
									>
										⏱ CPU Time Counters
									</h4>
									<dl
										class="divide-y divide-gray-200 text-sm text-gray-700 dark:divide-gray-700 dark:text-gray-300"
									>
										<div class="grid grid-cols-2 gap-x-4 px-2 py-2">
											<dt class="font-medium text-gray-500 dark:text-gray-400">User</dt>
											<dd class="font-semibold text-gray-800 dark:text-white" id="cpu-time-user">
												--
											</dd>
										</div>
										<div class="grid grid-cols-2 gap-x-4 bg-gray-50 px-2 py-2 dark:bg-gray-800">
											<dt class="font-medium text-gray-500 dark:text-gray-400">System</dt>
											<dd class="font-semibold text-gray-800 dark:text-white" id="cpu-time-system">
												--
											</dd>
										</div>
										<div class="grid grid-cols-2 gap-x-4 px-2 py-2">
											<dt class="font-medium text-gray-500 dark:text-gray-400">Idle</dt>
											<dd class="font-semibold text-gray-800 dark:text-white" id="cpu-time-idle">
												--
											</dd>
										</div>
										<div class="grid grid-cols-2 gap-x-4 bg-gray-50 px-2 py-2 dark:bg-gray-800">
											<dt class="font-medium text-gray-500 dark:text-gray-400">Nice</dt>
											<dd class="font-semibold text-gray-800 dark:text-white" id="cpu-time-nice">
												--
											</dd>
										</div>
										<div class="grid grid-cols-2 gap-x-4 px-2 py-2">
											<dt class="font-medium text-gray-500 dark:text-gray-400">Iowait</dt>
											<dd class="font-semibold text-gray-800 dark:text-white" id="cpu-time-iowait">
												--
											</dd>
										</div>
									</dl>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- DISK TAB -->
			{:else if activeTab === 'disk'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="disk"
					role="tabpanel"
					aria-labelledby="disk-tab"
				>
					<div class="space-y-6">
						<!-- Disk Usage Summary -->
						<div class="grid grid-cols-1 gap-4 md:grid-cols-4">
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Space</p>
									<p class="text-2xl font-bold text-gray-900 dark:text-white" id="disk-total">--</p>
								</div>
							</div>
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Used Space</p>
									<p class="text-2xl font-bold text-red-600 dark:text-red-400" id="disk-used">--</p>
								</div>
							</div>
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Free Space</p>
									<p class="text-2xl font-bold text-green-600 dark:text-green-400" id="disk-free">
										--
									</p>
								</div>
							</div>
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Usage %</p>
									<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
										<span id="disk-percent">--</span>%
									</p>
								</div>
							</div>
						</div>

						<!-- Disk Usage Chart -->
						<div
							class="rounded-lg border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-900"
						>
							<h3 class="mb-4 text-lg font-semibold text-gray-800 dark:text-white">
								Disk Usage Distribution
							</h3>
							<div id="diskUsageDonutChart" class="h-96 w-full"></div>
						</div>
					</div>
				</div>

				<!-- NETWORK TAB -->
			{:else if activeTab === 'network'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="network"
					role="tabpanel"
					aria-labelledby="network-tab"
				>
					<div class="space-y-6">
						<!-- Network Interface Stats -->
						<div
							class="rounded-lg border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-900"
						>
							<h3 class="mb-4 text-lg font-semibold text-gray-800 dark:text-white">
								Network Traffic
							</h3>
							<div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-3">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Current Upload</p>
									<p class="text-xl font-bold text-red-600 dark:text-red-400" id="current-tx">
										-- Mbps
									</p>
								</div>
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">
										Current Download
									</p>
									<p class="text-xl font-bold text-blue-600 dark:text-blue-400" id="current-rx">
										-- Mbps
									</p>
								</div>
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Peak Bandwidth</p>
									<p class="text-sm text-gray-600 dark:text-gray-300" id="peak-bandwidth">
										↑ -- / ↓ --
									</p>
								</div>
							</div>
							<div id="networkTrafficChart" class="h-80 w-full"></div>
						</div>
					</div>
				</div>

				<!-- ACTIVITY TAB -->
			{:else if activeTab === 'activity'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="activity"
					role="tabpanel"
					aria-labelledby="activity-tab"
				>
					<div class="rounded-lg bg-white shadow dark:bg-gray-800">
						<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">Recent Events</h3>
						</div>
						<div class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each Array.isArray(events) ? events : [] as event}
								<div class="px-6 py-4">
									<div class="flex items-center justify-between">
										<div>
											<h4 class="text-sm font-medium text-gray-900 dark:text-white">
												{event.category || event.type}
											</h4>
											<p class="text-sm text-gray-500 dark:text-gray-400">
												{event.message}
											</p>
										</div>
										<div class="text-right">
											<p class="text-xs text-gray-500 dark:text-gray-400">
												{formatDate(event.timestamp)}
											</p>
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
				</div>

				<!-- LOGS TAB -->
			{:else if activeTab === 'logs'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="logs"
					role="tabpanel"
					aria-labelledby="logs-tab"
				>
					<div class="rounded-lg bg-white shadow dark:bg-gray-800">
						<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">System Logs</h3>
						</div>
						<div class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each Array.isArray(logs) ? logs : [] as log}
								<div class="px-6 py-3">
									<div class="flex items-start space-x-3">
										<span
											class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {getBadgeClass(
												log.level
											)}"
										>
											{log.level}
										</span>
										<div class="min-w-0 flex-1">
											<p class="text-sm break-words text-gray-900 dark:text-white">{log.message}</p>
											<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
												{formatDate(log.timestamp)}
											</p>
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
				</div>

				<!-- CONSOLE TAB -->
			{:else if activeTab === 'console'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="console"
					role="tabpanel"
					aria-labelledby="console-tab"
				>
					<div class="rounded-lg bg-black p-4 font-mono text-sm text-green-400 shadow-lg">
						<div class="mb-4">
							<div id="console-prompt" class="mb-2">
								<span class="text-blue-400">user</span>@<span class="text-purple-400"
									>{endpoint?.hostname || 'host'}</span
								>:<span class="text-red-400">~</span><span class="text-white">$</span><span
									class="blink-cursor"
								></span>
							</div>
						</div>
						<div id="console-output" class="mb-4 h-96 overflow-y-auto">
							<div id="console-responses" class="space-y-1">
								<div class="text-gray-400">
									Welcome to {endpoint?.hostname || 'remote'} console. Type commands below:
								</div>
							</div>
						</div>
						<div class="flex items-center">
							<span class="text-blue-400">user</span>@<span class="text-purple-400"
								>{endpoint?.hostname || 'host'}</span
							>:<span class="text-red-400">~</span><span class="text-white">$ </span>
							<input
								id="console-command"
								class="ml-1 flex-1 border-none bg-transparent text-green-400 outline-none"
								type="text"
								placeholder="Enter command..."
								autocomplete="off"
							/>
						</div>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
