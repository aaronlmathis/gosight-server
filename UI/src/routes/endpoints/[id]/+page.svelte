<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy, tick } from 'svelte';
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
	import OverviewTab from './OverviewTab.svelte';
	import ComputeTab from './ComputeTab.svelte';
	import NetworkTab from './NetworkTab.svelte';
	import DiskTab from './DiskTab.svelte';
	import ActivityTab from './ActivityTab.svelte';
	import LogsTab from './LogsTab.svelte';
	import ConsoleTab from './ConsoleTab.svelte';

	let endpoint: Endpoint | null = null;
	let metrics: Metric[] = [];
	let events: Event[] = [];
	let logs: LogEntry[] = [];
	let alerts: Alert[] = [];
	let processes: any[] = []; // Add processes array
	let processHistory: Array<{ timestamp: number; processes: any[] }> = []; // Process history for tooltips
	let latestCpuPercent = 0; // Latest CPU percentage for tooltips
	let latestMemUsedPercent = 0; // Latest memory percentage for tooltips
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	// CPU Info state
	let cpuInfo = {
		model: '--',
		vendor: '--',
		cores: '--',
		threads: '--',
		baseClock: '--',
		cache: '--',
		family: '--',
		stepping: '--',
		physical: '--'
	};

	// CPU Time Counters state
	let cpuTimeCounters = {
		user: '--',
		system: '--',
		idle: '--',
		nice: '--',
		iowait: '--',
		irq: '--',
		softirq: '--',
		steal: '--',
		guest: '--',
		guest_nice: '--'
	};

	// Per-Core Usage state
	let perCoreData: Record<string, { usage?: number; clock?: number }> = {};

	// Host Info state
	let hostInfo = {
		summary: {} as Record<string, string>, // Will hold tags from 'info' metric
		procs: '--',
		uptime: '--',
		users_loggedin: '--'
	};

	const endpointId = $page.params.id;

	// Component references
	let overviewTabRef: any;
	let computeTabRef: any;
	let networkTabRef: any;
	let diskTabRef: any;

	let unsubscribeMetrics: (() => void) | null = null;
	let unsubscribeEvents: (() => void) | null = null;
	let unsubscribeLogs: (() => void) | null = null;
	let unsubscribeAlerts: (() => void) | null = null;
	let unsubscribeProcesses: (() => void) | null = null; // Add processes unsubscribe

	// Remove old element bindings since they're now in individual components

	onMount(async () => {
		await loadEndpointData();
		setupRealTimeUpdates();
		// Charts will be initialized by their respective components
	});

	onDestroy(() => {
		if (unsubscribeMetrics) unsubscribeMetrics();
		if (unsubscribeEvents) unsubscribeEvents();
		if (unsubscribeLogs) unsubscribeLogs();
		if (unsubscribeAlerts) unsubscribeAlerts();
		if (unsubscribeProcesses) unsubscribeProcesses(); // Add processes cleanup

		// Disconnect websockets
		websocketManager.disconnect();

		// Charts will be cleaned up by their respective components
	});
	async function loadEndpointData() {
		try {
			loading = true;
			// Only load endpoint information, not historical data
			// Charts will start empty and populate with live websocket data
			const endpointRes = await api.getEndpoint(endpointId);
			endpoint = endpointRes.data || endpointRes;

			// Initialize empty arrays - charts will populate with live data only
			metrics = [];
			events = [];
			logs = [];
			alerts = [];
		} catch (err) {
			console.error('Error loading endpoint data:', err);
			error = err instanceof Error ? err.message : 'Failed to load endpoint data';
			// Initialize empty arrays on error
			endpoint = null;
			metrics = [];
			events = [];
			logs = [];
			alerts = [];
		} finally {
			loading = false;
		}
	}

	async function switchTab(tabId: string) {
		activeTab = tabId;
		// Wait for DOM update then initialize charts for the specific tab
		await tick();
		if (tabId === 'overview' && overviewTabRef?.initCharts) {
			overviewTabRef.initCharts();
		} else if (tabId === 'compute' && computeTabRef?.initCharts) {
			computeTabRef.initCharts();
		} else if (tabId === 'network' && networkTabRef?.initCharts) {
			networkTabRef.initCharts();
		} else if (tabId === 'disk' && diskTabRef?.initCharts) {
			diskTabRef.initCharts();
		}
		// ConsoleTab handles console initialization within component
	}

	// Chart initialization is now handled by individual tab components
	function setupRealTimeUpdates() {
		console.log('Setting up real-time updates for endpoint:', endpointId);

		// Connect websockets with endpoint filtering (like the original implementation)
		websocketManager.connect(endpointId);

		// Subscribe to metrics updates
		unsubscribeMetrics = websocketManager.subscribeToMetrics((metricsPayload) => {
			console.log('Parent: Received metric update:', metricsPayload);
			if (metricsPayload && metricsPayload.endpoint_id === endpointId) {
				// The payload contains an array of metrics in the 'metrics' property
				if (Array.isArray(metricsPayload.metrics)) {
					// Add individual metrics to our metrics array for display
					const existingMetrics = Array.isArray(metrics) ? metrics : [];
					const newMetrics = metricsPayload.metrics.map((m: any) => ({
						...m,
						endpoint_id: metricsPayload.endpoint_id,
						timestamp: metricsPayload.timestamp
					}));
					metrics = [...newMetrics, ...existingMetrics.slice(0, 99)];
					console.log('Parent: Updated metrics array, total length:', metrics.length);
					console.log(
						'Parent: Metric names in array:',
						metrics.slice(0, 10).map((m) => m.name)
					);

					updateCharts(metricsPayload);
				}
			}
		});

		// Subscribe to events updates
		unsubscribeEvents = websocketManager.subscribeToEvents((latestEvent) => {
			//console.log('Received event update:', latestEvent);
			if (latestEvent && latestEvent.endpoint_id === endpointId) {
				// Ensure events is an array before using slice
				const existingEvents = Array.isArray(events) ? events : [];
				events = [latestEvent, ...existingEvents.slice(0, 49)];
			}
		});

		// Subscribe to logs updates
		unsubscribeLogs = websocketManager.subscribeToLogs((latestLog) => {
			//console.log('Received log update:', latestLog);
			if (latestLog && latestLog.endpoint_id === endpointId) {
				// Ensure logs is an array before using slice
				const existingLogs = Array.isArray(logs) ? logs : [];
				logs = [latestLog, ...existingLogs.slice(0, 49)];
			}
		});

		// Subscribe to alerts updates
		unsubscribeAlerts = websocketManager.subscribeToAlerts((latestAlert) => {
			//console.log('Received alert update:', latestAlert);
			if (latestAlert && latestAlert.endpoint_id === endpointId) {
				// Ensure alerts is an array before using slice
				const existingAlerts = Array.isArray(alerts) ? alerts : [];
				alerts = [latestAlert, ...existingAlerts.slice(0, 19)];
			}
		});

		// Subscribe to processes updates
		unsubscribeProcesses = websocketManager.subscribeToProcesses((processData) => {
			console.log('Received process update:', processData);
			if (processData && processData.endpoint_id === endpointId) {
				if (Array.isArray(processData.processes)) {
					processes = processData.processes;

					// Add to process history for tooltips (similar to vanilla JS)
					const ts = new Date(processData.timestamp).getTime();
					processHistory.push({ timestamp: ts, processes: processData.processes });

					// Keep only last 30 minutes of process history
					const cutoff = Date.now() - 30 * 60 * 1000;
					while (processHistory.length > 0 && processHistory[0].timestamp < cutoff) {
						processHistory.shift();
					}

					// Process data updated - individual tab components will handle their own updates
				}
			}
		});
	}

	async function updateCharts(metricsPayload: any) {
		if (!metricsPayload || !Array.isArray(metricsPayload.metrics)) {
			console.log('No metrics array found in payload');
			return;
		}

		const timestamp = new Date(metricsPayload.timestamp).getTime();

		// Process each metric in the array
		metricsPayload.metrics.forEach((metric: any) => {
			const metricValue = parseFloat(metric.value);
			const metricName = metric.name;
			const namespace = metric.namespace?.toLowerCase() || '';
			const subNamespace = metric.subnamespace?.toLowerCase() || '';
			const dimensions = metric.dimensions || {};

			//console.log(`Processing metric: ${namespace}.${subNamespace}.${metricName} = ${metricValue}`);

			// Create a full metric identifier for better matching
			const fullMetricName = `${namespace}.${subNamespace}.${metricName}`;

			// Update percentage labels in the UI based on various possible metric names
			const isCpuMetric =
				metricName === 'usage_percent' && subNamespace === 'cpu' && dimensions?.scope === 'total';

			const isMemoryMetric = metricName === 'used_percent' && subNamespace === 'memory';

			const isSwapMetric =
				['swap_used_percent', 'swap_total', 'swap_free', 'swap_used'].includes(metricName) &&
				subNamespace === 'memory';

			if (isCpuMetric) {
				latestCpuPercent = metricValue; // Update latest CPU percentage
			}

			if (isMemoryMetric) {
				latestMemUsedPercent = metricValue; // Update latest Memory percentage
			}

			// Process CPU Info metrics (based on original compute.js)
			if (namespace === 'system' && subNamespace === 'cpu') {
				if (metricName === 'count_logical') {
					cpuInfo.threads = metricValue.toString();
				} else if (metricName === 'count_physical') {
					cpuInfo.cores = metricValue.toString();
				} else if (metricName === 'clock_mhz' && dimensions?.core === 'core0') {
					// Get CPU info from core0 dimensions (like original)
					cpuInfo.baseClock = `${metricValue.toFixed(0)} MHz`;
					cpuInfo.vendor = dimensions.vendor || '--';
					cpuInfo.model = dimensions.model || '--';
					cpuInfo.family = dimensions.family || '--';
					cpuInfo.stepping = dimensions.stepping || '--';
					cpuInfo.cache = dimensions.cache
						? `${parseInt(dimensions.cache).toLocaleString()} KB`
						: '--';
					cpuInfo.physical = dimensions.physical === 'true' ? 'Yes' : 'No';
				}
			}

			// Process CPU Time Counter metrics
			if (namespace === 'system' && subNamespace === 'cpu' && metricName.startsWith('time_')) {
				const timeType = metricName.replace('time_', '');
				if (timeType in cpuTimeCounters) {
					// Convert to seconds and format (like original)
					const seconds = Math.round(metricValue);
					cpuTimeCounters[timeType as keyof typeof cpuTimeCounters] = seconds.toLocaleString();
				}
			}

			// Process Per-Core Usage metrics
			if (
				namespace === 'system' &&
				subNamespace === 'cpu' &&
				metricName === 'usage_percent' &&
				dimensions?.scope === 'per_core'
			) {
				const core = dimensions.core;
				if (core) {
					if (!perCoreData[core]) {
						perCoreData[core] = {};
					}
					perCoreData[core].usage = metricValue;
				}
			}

			// Process Per-Core Clock metrics
			if (
				namespace === 'system' &&
				subNamespace === 'cpu' &&
				metricName === 'clock_mhz' &&
				dimensions?.core &&
				dimensions.core !== 'core0'
			) {
				const core = dimensions.core;
				if (core) {
					if (!perCoreData[core]) {
						perCoreData[core] = {};
					}
					perCoreData[core].clock = metricValue;
				}
			}

			// Process Host Info metrics (system.host namespace)
			if (namespace === 'system' && subNamespace === 'host') {
				if (metricName === 'procs') {
					hostInfo.procs = metricValue.toString();
				} else if (metricName === 'uptime') {
					// Format uptime from seconds to readable format
					const seconds = Math.floor(metricValue);
					const days = Math.floor(seconds / 86400);
					const hours = Math.floor((seconds % 86400) / 3600);
					const minutes = Math.floor((seconds % 3600) / 60);

					if (days > 0) {
						hostInfo.uptime = `${days}d ${hours}h ${minutes}m`;
					} else if (hours > 0) {
						hostInfo.uptime = `${hours}h ${minutes}m`;
					} else {
						hostInfo.uptime = `${minutes}m`;
					}
				} else if (metricName === 'users_loggedin') {
					hostInfo.users_loggedin = metricValue.toString();
				} else if (metricName === 'info') {
					// Extract host info from dimensions/tags
					hostInfo.summary = { ...dimensions };
				}
			}

			// Chart updates are now handled by individual tab components
			// This main function only processes and stores the metric data
		});

		// Data processing complete - individual tab components will handle their own chart updates
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
		<!-- Tabs navigation -->
		<div class="border-b border-gray-200 dark:border-gray-700">
			<nav class="-mb-px flex space-x-8 px-4 sm:px-6 lg:px-8" id="dashboardTabs" role="tablist">
				{#each [{ id: 'overview', label: 'Overview', icon: Activity }, { id: 'compute', label: 'Compute', icon: Cpu }, { id: 'disk', label: 'Disk', icon: HardDrive }, { id: 'network', label: 'Network', icon: Wifi }, { id: 'activity', label: 'Activity', icon: Monitor }, { id: 'logs', label: 'Logs', icon: ScrollText }, { id: 'console', label: 'Console', icon: Terminal }] as tab}
					<button
						class="border-b-2 px-1 py-4 text-sm font-medium {activeTab === tab.id
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'}"
						on:click={() => switchTab(tab.id)}
						role="tab"
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

		<!-- Tabs content -->
		<div class="p-4" id="tab-content">
			{#if activeTab === 'overview'}
				<OverviewTab
					bind:this={overviewTabRef}
					{endpoint}
					{hostInfo}
					{metrics}
					{processes}
					{runCommand}
				/>
			{:else if activeTab === 'compute'}
				<ComputeTab
					bind:this={computeTabRef}
					{metrics}
					{cpuInfo}
					{cpuTimeCounters}
					{perCoreData}
					{processes}
					{hostInfo}
				/>
			{:else if activeTab === 'network'}
				<NetworkTab bind:this={networkTabRef} {metrics} />
			{:else if activeTab === 'disk'}
				<DiskTab bind:this={diskTabRef} {metrics} />
			{:else if activeTab === 'activity'}
				<ActivityTab {events} />
			{:else if activeTab === 'logs'}
				<LogsTab {logs} {endpointId} />
			{:else if activeTab === 'console'}
				<ConsoleTab {endpoint} on:run={(e) => runCommand(e.detail)} />
			{/if}
		</div>
	{/if}
</div>
