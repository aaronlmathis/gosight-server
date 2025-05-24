<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import { formatNumber, formatDate } from '$lib/utils';
	import type { Metric, Endpoint } from '$lib/types';

	let endpoints: Endpoint[] = [];
	let metrics: Metric[] = [];
	let selectedEndpoint = '';
	let selectedMetrics: string[] = [];
	let timeRange = '1h';
	let loading = false;
	let error = '';
	let chartElement: HTMLElement;
	let chart: any;

	// Chart data
	let chartData: any = null;
	let availableMetrics: string[] = [];

	// Time range options
	const timeRanges = [
		{ value: '15m', label: '15 minutes' },
		{ value: '1h', label: '1 hour' },
		{ value: '6h', label: '6 hours' },
		{ value: '24h', label: '24 hours' },
		{ value: '7d', label: '7 days' },
		{ value: '30d', label: '30 days' }
	];

	onMount(async () => {
		await loadEndpoints();
		await loadAvailableMetrics();
		initializeChart();
		
		// Subscribe to real-time metric updates
		websocketManager.connect();
		websocketManager.subscribeToMetrics((metric: Metric) => {
			if (selectedEndpoint && metric.endpoint_id === selectedEndpoint) {
				updateChartWithNewData(metric);
			}
		});
	});

	async function loadEndpoints() {
		try {
			const response = await api.getEndpoints();
			endpoints = response.endpoints || [];
		} catch (err) {
			console.error('Failed to load endpoints:', err);
		}
	}

	async function loadAvailableMetrics() {
		try {
			const response = await api.getMetrics();
			const allMetrics = response.metrics || [];
			availableMetrics = [...new Set(allMetrics.map(m => m.name))];
		} catch (err) {
			console.error('Failed to load available metrics:', err);
		}
	}

	async function loadMetrics() {
		if (!selectedEndpoint || selectedMetrics.length === 0) return;

		try {
			loading = true;
			error = '';
			
			const response = await api.getMetrics({
				endpoint_id: selectedEndpoint,
				metric_names: selectedMetrics,
				time_range: timeRange
			});
			
			metrics = response.metrics || [];
			updateChart();
		} catch (err) {
			error = 'Failed to load metrics: ' + (err as Error).message;
			console.error('Failed to load metrics:', err);
		} finally {
			loading = false;
		}
	}

	function initializeChart() {
		if (typeof window !== 'undefined' && window.ApexCharts && chartElement) {
			const options = {
				chart: {
					type: 'line',
					height: 400,
					animations: {
						enabled: true,
						easing: 'linear',
						dynamicAnimation: {
							speed: 1000
						}
					},
					toolbar: {
						show: true,
						tools: {
							download: true,
							selection: true,
							zoom: true,
							zoomin: true,
							zoomout: true,
							pan: true,
							reset: true
						}
					},
					background: 'transparent'
				},
				theme: {
					mode: document.documentElement.classList.contains('dark') ? 'dark' : 'light'
				},
				stroke: {
					curve: 'smooth',
					width: 2
				},
				xaxis: {
					type: 'datetime',
					labels: {
						datetimeUTC: false
					}
				},
				yaxis: {
					labels: {
						formatter: function(value: number) {
							return formatNumber(value);
						}
					}
				},
				tooltip: {
					x: {
						format: 'dd MMM yyyy HH:mm:ss'
					},
					y: {
						formatter: function(value: number) {
							return formatNumber(value);
						}
					}
				},
				legend: {
					position: 'top'
				},
				grid: {
					show: true,
					borderColor: '#e0e4e7'
				},
				colors: ['#3B82F6', '#EF4444', '#10B981', '#F59E0B', '#8B5CF6', '#EC4899']
			};

			chart = new window.ApexCharts(chartElement, options);
			chart.render();
		}
	}

	function updateChart() {
		if (!chart || metrics.length === 0) return;

		// Group metrics by name
		const metricGroups: { [key: string]: Metric[] } = {};
		metrics.forEach(metric => {
			if (!metricGroups[metric.name]) {
				metricGroups[metric.name] = [];
			}
			metricGroups[metric.name].push(metric);
		});

		// Prepare chart series
		const series = Object.keys(metricGroups).map(metricName => ({
			name: metricName,
			data: metricGroups[metricName]
				.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime())
				.map(metric => ({
					x: new Date(metric.timestamp).getTime(),
					y: metric.value
				}))
		}));

		chart.updateSeries(series);
	}

	function updateChartWithNewData(metric: Metric) {
		if (!chart || !selectedMetrics.includes(metric.name)) return;

		const newDataPoint = {
			x: new Date(metric.timestamp).getTime(),
			y: metric.value
		};

		// Find the series for this metric
		const seriesIndex = chart.w.config.series.findIndex((s: any) => s.name === metric.name);
		if (seriesIndex !== -1) {
			chart.appendData([{
				data: [newDataPoint]
			}], seriesIndex);
		}
	}

	function toggleMetric(metricName: string) {
		if (selectedMetrics.includes(metricName)) {
			selectedMetrics = selectedMetrics.filter(m => m !== metricName);
		} else {
			selectedMetrics = [...selectedMetrics, metricName];
		}
	}

	function onEndpointChange() {
		selectedMetrics = [];
		metrics = [];
		if (chart) {
			chart.updateSeries([]);
		}
	}

	$: if (selectedEndpoint && selectedMetrics.length > 0) {
		loadMetrics();
	}
</script>

<svelte:head>
	<title>Metrics Explorer - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Metrics Explorer</h1>
		<p class="text-gray-600 dark:text-gray-400">Visualize and analyze system metrics</p>
	</div>

	<!-- Controls -->
	<div class="mb-6 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<!-- Endpoint Selection -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
					Endpoint
				</label>
				<select
					bind:value={selectedEndpoint}
					on:change={onEndpointChange}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					<option value="">Select an endpoint</option>
					{#each endpoints as endpoint}
						<option value={endpoint.id}>{endpoint.name}</option>
					{/each}
				</select>
			</div>

			<!-- Time Range -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
					Time Range
				</label>
				<select
					bind:value={timeRange}
					on:change={loadMetrics}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					{#each timeRanges as range}
						<option value={range.value}>{range.label}</option>
					{/each}
				</select>
			</div>

			<!-- Refresh Button -->
			<div class="flex items-end">
				<button
					on:click={loadMetrics}
					disabled={!selectedEndpoint || selectedMetrics.length === 0 || loading}
					class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{#if loading}
						<i class="fas fa-spinner fa-spin mr-2"></i>
						Loading...
					{:else}
						<i class="fas fa-sync-alt mr-2"></i>
						Refresh
					{/if}
				</button>
			</div>
		</div>
	</div>

	<!-- Metric Selection -->
	{#if selectedEndpoint}
		<div class="mb-6 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
			<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-3">Select Metrics</h3>
			<div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
				{#each availableMetrics as metricName}
					<label class="flex items-center">
						<input
							type="checkbox"
							checked={selectedMetrics.includes(metricName)}
							on:change={() => toggleMetric(metricName)}
							class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
						/>
						<span class="ml-2 text-sm text-gray-900 dark:text-white">{metricName}</span>
					</label>
				{/each}
			</div>
		</div>
	{/if}

	{#if error}
		<div class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
			<div class="flex">
				<i class="fas fa-exclamation-triangle text-red-500 mr-3 mt-0.5"></i>
				<div>
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<p class="text-sm text-red-600 dark:text-red-300 mt-1">{error}</p>
				</div>
			</div>
		</div>
	{/if}

	<!-- Chart -->
	<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
		<div class="mb-4 flex justify-between items-center">
			<h3 class="text-lg font-medium text-gray-900 dark:text-white">
				{selectedEndpoint ? endpoints.find(e => e.id === selectedEndpoint)?.name || 'Unknown Endpoint' : 'Select an endpoint to view metrics'}
			</h3>
			{#if selectedMetrics.length > 0}
				<div class="text-sm text-gray-500 dark:text-gray-400">
					{selectedMetrics.length} metric{selectedMetrics.length !== 1 ? 's' : ''} selected
				</div>
			{/if}
		</div>

		<div bind:this={chartElement} class="w-full"></div>

		{#if !selectedEndpoint}
			<div class="text-center py-12">
				<i class="fas fa-chart-line text-4xl text-gray-400 mb-4"></i>
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No Endpoint Selected</h3>
				<p class="text-gray-600 dark:text-gray-400">Choose an endpoint to start exploring metrics.</p>
			</div>
		{:else if selectedMetrics.length === 0}
			<div class="text-center py-12">
				<i class="fas fa-check-square text-4xl text-gray-400 mb-4"></i>
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No Metrics Selected</h3>
				<p class="text-gray-600 dark:text-gray-400">Select one or more metrics to visualize.</p>
			</div>
		{/if}
	</div>

	<!-- Current Values -->
	{#if metrics.length > 0}
		<div class="mt-6 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			{#each selectedMetrics as metricName}
				{@const latestMetric = metrics.filter(m => m.name === metricName).sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())[0]}
				{#if latestMetric}
					<div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-sm font-medium text-gray-600 dark:text-gray-400">{metricName}</p>
								<p class="text-2xl font-bold text-gray-900 dark:text-white">
									{formatNumber(latestMetric.value)}
									{#if latestMetric.unit}
										<span class="text-sm font-normal text-gray-500 dark:text-gray-400">{latestMetric.unit}</span>
									{/if}
								</p>
							</div>
							<div class="text-right">
								<p class="text-xs text-gray-500 dark:text-gray-400">
									{formatDate(latestMetric.timestamp)}
								</p>
							</div>
						</div>
					</div>
				{/if}
			{/each}
		</div>
	{/if}
</div>
