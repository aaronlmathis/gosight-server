<script lang="ts">
	import type { Widget, WidgetData, ChartSeries } from '$lib/types/dashboard';
	import { Card } from 'flowbite-svelte';
	import { onMount, onDestroy } from 'svelte';
	import { dataService } from '$lib/services/dataService';

	export let widget: Widget;

	let chartContainer: HTMLElement;
	let chart: any;
	let widgetData: WidgetData | null = null;
	let loading = true;
	let error = '';
	let unsubscribeFunctions: (() => void)[] = [];

	// Initialize chart options with empty data
	let chartOptions = {
		chart: {
			type: widget.config?.chartConfig?.chartType || 'line',
			height: '100%',
			toolbar: { show: false },
			animations: { enabled: true, easing: 'easeinout', speed: 800 }
		},
		series: [],
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'HH:mm',
				style: {
					fontSize: '10px',
					colors: '#6B7280'
				}
			}
		},
		yaxis: {
			labels: {
				style: {
					fontSize: '10px',
					colors: '#6B7280'
				}
			}
		},
		stroke: {
			curve: 'smooth',
			width: 2
		},
		colors: ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'],
		grid: {
			show: true,
			strokeDashArray: 3,
			borderColor: '#E5E7EB',
			xaxis: { lines: { show: false } }
		},
		tooltip: {
			x: { format: 'HH:mm:ss' },
			theme: 'light'
		},
		legend: {
			show: true,
			position: 'top',
			horizontalAlign: 'left',
			fontSize: '12px'
		},
		dataLabels: {
			enabled: false
		}
	};

	// Load widget data and setup real-time updates
	async function loadChartData() {
		try {
			loading = true;
			error = '';

			const data = await dataService.getWidgetData(widget);
			widgetData = data;

			if (data.status === 'error') {
				error = data.error || 'Failed to load chart data';
				return;
			}

			if (data.series && Array.isArray(data.series)) {
				updateChartSeries(data.series);
			}
		} catch (err) {
			console.error('Failed to load chart data:', err);
			error = err instanceof Error ? err.message : 'Failed to load chart data';
		} finally {
			loading = false;
		}
	}

	// Update chart with new series data
	function updateChartSeries(series: ChartSeries[]) {
		if (chart && series.length > 0) {
			chart.updateSeries(series);
		} else if (chart) {
			// No data available, show empty chart
			chart.updateSeries([]);
		}
	}

	// Setup real-time subscriptions for configured metrics
	function setupRealTimeSubscriptions() {
		const chartConfig = widget.config?.chartConfig;
		const selectedMetrics = chartConfig?.selectedMetrics || [];

		selectedMetrics.forEach((metricOption: any) => {
			const { namespace, subnamespace, name: metric, tags } = metricOption;

			if (namespace && subnamespace && metric) {
				const unsubscribe = dataService.subscribeToMetric(
					widget.id,
					namespace,
					subnamespace,
					metric,
					(dataPoints) => {
						// Transform metric data points to chart format
						const chartData = dataPoints.map((point) => ({
							x: point.timestamp,
							y: point.value
						}));

						// Update the specific series
						if (chart) {
							const seriesIndex = selectedMetrics.findIndex(
								(m: any) =>
									m.namespace === namespace &&
									m.subnamespace === subnamespace &&
									m.name === metric
							);

							if (seriesIndex >= 0) {
								// Create series name with tags for better identification
								const tagString = Object.entries(tags || {})
									.map(([key, value]) => `${key}=${value}`)
									.join(',');
								const seriesName = tagString 
									? `${metricOption.label} (${tagString})`
									: metricOption.label;

								chart.updateSeries(
									[
										{
											name: seriesName,
											data: chartData
										}
									],
									false,
									true
								);
							}
						}
					},
					widget.config?.endpointId,
					tags // Pass tags for filtering
				);

				unsubscribeFunctions.push(unsubscribe);
			}
		});
	}

	onMount(async () => {
		if (typeof window !== 'undefined') {
			try {
				// Dynamically import ApexCharts to avoid SSR issues
				const ApexCharts = (await import('apexcharts')).default;

				chart = new ApexCharts(chartContainer, chartOptions);
				await chart.render();

				// Load initial data
				await loadChartData();

				// Setup real-time subscriptions
				setupRealTimeSubscriptions();
			} catch (err) {
				console.error('Failed to initialize chart:', err);
				error = 'Failed to initialize chart';
			}
		}
	});

	onDestroy(() => {
		// Cleanup chart
		if (chart) {
			chart.destroy();
		}

		// Cleanup subscriptions
		unsubscribeFunctions.forEach((fn) => fn());
		unsubscribeFunctions = [];
	});
</script>

<Card class="relative h-full">
	<div class="flex h-full flex-col">
		<!-- Header -->
		{#if widget.config?.showTitle !== false}
			<div
				class="mb-3 flex items-center justify-between text-sm font-medium text-gray-900 dark:text-gray-100"
			>
				<span>{widget.title}</span>
				{#if loading}
					<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-500"></div>
				{/if}
			</div>
		{/if}

		<!-- Error State -->
		{#if error}
			<div class="flex flex-1 items-center justify-center">
				<div class="text-center">
					<div class="mb-2 text-sm text-red-500 dark:text-red-400">
						<svg class="mx-auto mb-2 h-8 w-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							></path>
						</svg>
						Chart Error
					</div>
					<div class="text-xs text-gray-500 dark:text-gray-400">{error}</div>
				</div>
			</div>
		{:else}
			<!-- Chart Container -->
			<div class="relative min-h-0 flex-1">
				{#if loading}
					<div
						class="bg-opacity-50 absolute inset-0 z-10 flex items-center justify-center bg-white dark:bg-gray-800"
					>
						<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-500"></div>
					</div>
				{/if}
				<div bind:this={chartContainer} class="h-full w-full"></div>
			</div>
		{/if}

		<!-- Footer with metadata -->
		{#if widgetData?.timestamp && !error}
			<div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
				Last updated: {new Date(widgetData.timestamp).toLocaleTimeString()}
			</div>
		{/if}
	</div>
</Card>
