<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { formatDate, formatBytes, formatDuration } from '$lib/utils';
	import { BarChart, FileText, Download, Calendar, Filter, TrendingUp } from 'lucide-svelte';
	import DataTable from '$lib/components/DataTable.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import Loading from '$lib/components/Loading.svelte';

	let activeTab = 'overview';
	let loading = false;
	let error = '';

	// Report data
	let systemSummary: any = null;
	let alertsReport: any[] = [];
	let metricsReport: any[] = [];
	let eventsReport: any[] = [];

	// Filters
	let dateRange = '7d';
	let selectedEndpoints: string[] = [];
	let reportType = 'all';

	// Modal state
	let showExportModal = false;
	let exportLoading = false;

	// Chart variables
	let alertsChart: any;
	let metricsChart: any;

	const reportColumns = [
		{ key: 'name', label: 'Name', sortable: true },
		{ key: 'type', label: 'Type', sortable: true },
		{ key: 'created_at', label: 'Created', sortable: true, format: (value: string) => formatDate(value) },
		{ key: 'size', label: 'Size', sortable: true, format: (value: number) => formatBytes(value) }
	];

	onMount(async () => {
		await loadReportsData();
		initCharts();
	});

	async function loadReportsData() {
		try {
			loading = true;
			error = '';

			const [summaryRes, alertsRes, metricsRes, eventsRes] = await Promise.all([
				api.getSystemSummary({ range: dateRange }),
				api.getAlertsReport({ range: dateRange, endpoints: selectedEndpoints }),
				api.getMetricsReport({ range: dateRange, endpoints: selectedEndpoints }),
				api.getEventsReport({ range: dateRange, endpoints: selectedEndpoints })
			]);

			systemSummary = summaryRes.data;
			alertsReport = alertsRes.data;
			metricsReport = metricsRes.data;
			eventsReport = eventsRes.data;

			updateCharts();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load reports data';
		} finally {
			loading = false;
		}
	}

	function initCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// Alerts chart
		const alertsOptions = {
			chart: {
				type: 'donut',
				height: 300
			},
			series: [],
			labels: [],
			colors: ['#ef4444', '#f59e0b', '#10b981', '#3b82f6'],
			legend: { position: 'bottom' },
			theme: { mode: 'light' }
		};

		alertsChart = new window.ApexCharts(document.querySelector('#alerts-chart'), alertsOptions);
		alertsChart.render();

		// Metrics chart
		const metricsOptions = {
			chart: {
				type: 'line',
				height: 300,
				toolbar: { show: false }
			},
			series: [],
			xaxis: { type: 'datetime' },
			yaxis: { labels: { formatter: (val: number) => val.toFixed(1) + '%' } },
			colors: ['#3b82f6', '#10b981', '#f59e0b'],
			stroke: { curve: 'smooth', width: 2 },
			theme: { mode: 'light' }
		};

		metricsChart = new window.ApexCharts(document.querySelector('#metrics-chart'), metricsOptions);
		metricsChart.render();
	}

	function updateCharts() {
		if (!alertsChart || !metricsChart) return;

		// Update alerts chart
		if (systemSummary?.alerts_by_severity) {
			const alertsData = Object.entries(systemSummary.alerts_by_severity);
			alertsChart.updateSeries(alertsData.map(([_, count]) => count));
			alertsChart.updateOptions({
				labels: alertsData.map(([severity]) => severity.charAt(0).toUpperCase() + severity.slice(1))
			});
		}

		// Update metrics chart
		if (metricsReport.length > 0) {
			const cpuData = metricsReport.filter(m => m.name === 'cpu_usage').map(m => ({
				x: new Date(m.timestamp).getTime(),
				y: m.value
			}));
			const memoryData = metricsReport.filter(m => m.name === 'memory_usage').map(m => ({
				x: new Date(m.timestamp).getTime(),
				y: m.value
			}));

			metricsChart.updateSeries([
				{ name: 'CPU Usage', data: cpuData },
				{ name: 'Memory Usage', data: memoryData }
			]);
		}
	}

	async function exportReport(format: 'pdf' | 'csv' | 'excel') {
		try {
			exportLoading = true;
			
			const response = await api.exportReport({
				type: reportType,
				format,
				range: dateRange,
				endpoints: selectedEndpoints
			});

			// Trigger download
			const blob = new Blob([response.data], { 
				type: format === 'pdf' ? 'application/pdf' : 'application/octet-stream' 
			});
			const url = window.URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = `gosight-report-${dateRange}.${format}`;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
			window.URL.revokeObjectURL(url);

			showExportModal = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to export report';
		} finally {
			exportLoading = false;
		}
	}

	function applyFilters() {
		loadReportsData();
	}
</script>

<svelte:head>
	<title>Reports - GoSight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white dark:bg-gray-800 shadow">
		<div class="px-4 sm:px-6 lg:px-8">
			<div class="flex items-center justify-between h-16">
				<div class="flex items-center">
					<BarChart class="h-6 w-6 text-gray-400 mr-3" />
					<h1 class="text-xl font-semibold text-gray-900 dark:text-white">Reports</h1>
				</div>
				<div class="flex items-center space-x-4">
					<!-- Filters -->
					<select 
						bind:value={dateRange} 
						on:change={applyFilters}
						class="text-sm border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-white"
					>
						<option value="24h">Last 24 hours</option>
						<option value="7d">Last 7 days</option>
						<option value="30d">Last 30 days</option>
						<option value="90d">Last 90 days</option>
					</select>

					<button
						type="button"
						class="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						on:click={() => showExportModal = true}
					>
						<Download class="h-4 w-4 mr-2" />
						Export
					</button>
				</div>
			</div>
		</div>
	</div>

	{#if error}
		<div class="mx-4 sm:mx-6 lg:mx-8 mt-4">
			<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
				<p class="text-red-800 dark:text-red-200">{error}</p>
			</div>
		</div>
	{/if}

	<!-- Tabs -->
	<div class="border-b border-gray-200 dark:border-gray-700">
		<nav class="px-4 sm:px-6 lg:px-8 -mb-px flex space-x-8">
			{#each [
				{ id: 'overview', label: 'Overview', icon: TrendingUp },
				{ id: 'alerts', label: 'Alerts Report', icon: FileText },
				{ id: 'metrics', label: 'Metrics Report', icon: BarChart },
				{ id: 'events', label: 'Events Report', icon: Calendar }
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
		{#if loading}
			<Loading size="lg" text="Loading reports..." />
		{:else if activeTab === 'overview'}
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
				<!-- System Summary -->
				{#if systemSummary}
					<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
						<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">System Summary</h3>
						<dl class="grid grid-cols-2 gap-4">
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Endpoints</dt>
								<dd class="text-2xl font-bold text-gray-900 dark:text-white">{systemSummary.total_endpoints || 0}</dd>
							</div>
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Active Alerts</dt>
								<dd class="text-2xl font-bold text-red-600 dark:text-red-400">{systemSummary.active_alerts || 0}</dd>
							</div>
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Events</dt>
								<dd class="text-2xl font-bold text-blue-600 dark:text-blue-400">{systemSummary.total_events || 0}</dd>
							</div>
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Avg Response Time</dt>
								<dd class="text-2xl font-bold text-green-600 dark:text-green-400">{systemSummary.avg_response_time || 0}ms</dd>
							</div>
						</dl>
					</div>
				{/if}

				<!-- Alerts Chart -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Alerts by Severity</h3>
					<div id="alerts-chart"></div>
				</div>

				<!-- Performance Metrics -->
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6 lg:col-span-2">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Performance Metrics</h3>
					<div id="metrics-chart"></div>
				</div>
			</div>

		{:else if activeTab === 'alerts'}
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Alerts Report</h3>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
						Summary of alerts for the selected time period
					</p>
				</div>
				<div class="p-6">
					{#if alertsReport.length > 0}
						<DataTable 
							data={alertsReport} 
							columns={reportColumns}
							itemsPerPage={25}
							searchable={true}
						/>
					{:else}
						<div class="text-center py-12">
							<FileText class="mx-auto h-12 w-12 text-gray-400" />
							<h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No alerts</h3>
							<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">No alerts found for the selected period.</p>
						</div>
					{/if}
				</div>
			</div>

		{:else if activeTab === 'metrics'}
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Metrics Report</h3>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
						Performance metrics summary
					</p>
				</div>
				<div class="p-6">
					{#if metricsReport.length > 0}
						<DataTable 
							data={metricsReport} 
							columns={reportColumns}
							itemsPerPage={25}
							searchable={true}
						/>
					{:else}
						<div class="text-center py-12">
							<BarChart class="mx-auto h-12 w-12 text-gray-400" />
							<h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No metrics</h3>
							<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">No metrics found for the selected period.</p>
						</div>
					{/if}
				</div>
			</div>

		{:else if activeTab === 'events'}
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Events Report</h3>
					<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
						System events and activities
					</p>
				</div>
				<div class="p-6">
					{#if eventsReport.length > 0}
						<DataTable 
							data={eventsReport} 
							columns={reportColumns}
							itemsPerPage={25}
							searchable={true}
						/>
					{:else}
						<div class="text-center py-12">
							<Calendar class="mx-auto h-12 w-12 text-gray-400" />
							<h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No events</h3>
							<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">No events found for the selected period.</p>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Export Modal -->
<Modal bind:show={showExportModal} title="Export Report" size="md">
	<div class="space-y-4">
		<div>
			<label for="reportType" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
				Report Type
			</label>
			<select
				id="reportType"
				bind:value={reportType}
				class="mt-1 block w-full border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
			>
				<option value="all">Complete Report</option>
				<option value="alerts">Alerts Only</option>
				<option value="metrics">Metrics Only</option>
				<option value="events">Events Only</option>
			</select>
		</div>

		<div>
			<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
				Export Format
			</label>
			<div class="grid grid-cols-3 gap-3">
				<button
					type="button"
					disabled={exportLoading}
					class="flex flex-col items-center p-4 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
					on:click={() => exportReport('pdf')}
				>
					<FileText class="h-8 w-8 text-red-500 mb-2" />
					<span class="text-sm font-medium text-gray-900 dark:text-white">PDF</span>
				</button>
				<button
					type="button"
					disabled={exportLoading}
					class="flex flex-col items-center p-4 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
					on:click={() => exportReport('csv')}
				>
					<FileText class="h-8 w-8 text-green-500 mb-2" />
					<span class="text-sm font-medium text-gray-900 dark:text-white">CSV</span>
				</button>
				<button
					type="button"
					disabled={exportLoading}
					class="flex flex-col items-center p-4 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
					on:click={() => exportReport('excel')}
				>
					<FileText class="h-8 w-8 text-blue-500 mb-2" />
					<span class="text-sm font-medium text-gray-900 dark:text-white">Excel</span>
				</button>
			</div>
		</div>

		{#if exportLoading}
			<div class="flex items-center justify-center py-4">
				<Loading size="sm" text="Generating report..." />
			</div>
		{/if}
	</div>
</Modal>
