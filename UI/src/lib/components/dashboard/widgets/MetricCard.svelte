<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Card } from 'flowbite-svelte';
	import { TrendingUp, TrendingDown, Minus, Loader } from 'lucide-svelte';
	import { dataService } from '$lib/services/dataService';
	import type { Widget, WidgetData } from '$lib/types/dashboard';

	export let widget: Widget;

	let data: WidgetData = { status: 'unknown', value: 0 };
	let loading = true;
	let error = '';
	let unsubscribe: (() => void) | null = null;

	// Format value with appropriate precision and unit
	function formatValue(value: number, unit?: string): string {
		if (typeof value !== 'number' || isNaN(value)) return '0';

		let formatted: string;
		if (value >= 1000000) {
			formatted = (value / 1000000).toFixed(1) + 'M';
		} else if (value >= 1000) {
			formatted = (value / 1000).toFixed(1) + 'K';
		} else if (value >= 100) {
			formatted = value.toFixed(0);
		} else if (value >= 10) {
			formatted = value.toFixed(1);
		} else {
			formatted = value.toFixed(2);
		}

		return unit ? `${formatted}${unit}` : formatted;
	}

	// Get status styles
	$: statusColor =
		{
			success: 'text-green-600',
			warning: 'text-yellow-600',
			error: 'text-red-600',
			unknown: 'text-gray-600'
		}[data.status] || 'text-gray-600';

	$: statusBg =
		{
			success: 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800',
			warning: 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800',
			error: 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800',
			unknown: 'bg-gray-50 border-gray-200 dark:bg-gray-900/20 dark:border-gray-800'
		}[data.status] || 'bg-gray-50 border-gray-200';

	// Load widget data
	async function loadData() {
		try {
			loading = true;
			error = '';
			data = await dataService.getWidgetData(widget);
		} catch (err) {
			console.error(`Failed to load data for widget ${widget.id}:`, err);
			error = err instanceof Error ? err.message : 'Failed to load data';
			data = { status: 'error', value: 0, error };
		} finally {
			loading = false;
		}
	}

	// Subscribe to real-time updates if it's a metric widget
	function subscribeToUpdates() {
		const config = widget.config || {};

		if (widget.type === 'metric' && config.namespace && config.subnamespace && config.metric) {
			unsubscribe = dataService.subscribeToMetric(
				widget.id,
				config.namespace,
				config.subnamespace,
				config.metric,
				(metricData) => {
					if (metricData && metricData.length > 0) {
						const latest = metricData[0];
						data = {
							...data,
							value: latest.value,
							timestamp: new Date(latest.timestamp).toISOString()
						};
					}
				},
				config.endpointId
			);
		}
	}

	let refreshInterval: number;

	onMount(async () => {
		await loadData();
		subscribeToUpdates();

		// Set up periodic refresh for non-metric widgets
		if (widget.type !== 'metric') {
			refreshInterval = setInterval(loadData, widget.config?.refreshInterval || 30000);
		}
	});

	onDestroy(() => {
		if (unsubscribe) {
			unsubscribe();
		}
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});
</script>

<Card class="h-full {statusBg} relative border-2">
	<!-- Loading Overlay -->
	{#if loading}
		<div
			class="absolute inset-0 z-10 flex items-center justify-center rounded-lg bg-white/50 dark:bg-gray-900/50"
		>
			<Loader class="h-6 w-6 animate-spin text-blue-600" />
		</div>
	{/if}

	<div class="flex h-full flex-col justify-between">
		<!-- Header -->
		{#if widget.config?.showTitle !== false}
			<div class="mb-2 text-sm font-medium text-gray-600 dark:text-gray-400">
				{widget.title}
			</div>
		{/if}

		<!-- Main Value -->
		<div class="flex flex-1 flex-col justify-center">
			{#if error}
				<div class="text-center">
					<div class="mb-2 text-red-600 dark:text-red-400">Error</div>
					<div class="text-xs text-gray-500 dark:text-gray-400">{error}</div>
				</div>
			{:else}
				<div class="text-3xl font-bold {statusColor} mb-1">
					{formatValue(data.value || 0, data.unit || widget.config?.unit)}
				</div>

				<!-- Trend -->
				{#if data.trend}
					<div class="flex items-center text-sm">
						{#if data.trend === 'up'}
							<TrendingUp class="mr-1 h-4 w-4 text-green-500" />
							<span class="text-green-500">Trending up</span>
						{:else if data.trend === 'down'}
							<TrendingDown class="mr-1 h-4 w-4 text-red-500" />
							<span class="text-red-500">Trending down</span>
						{:else}
							<Minus class="mr-1 h-4 w-4 text-gray-400" />
							<span class="text-gray-500">Stable</span>
						{/if}
					</div>
				{/if}
			{/if}
		</div>

		<!-- Status Indicator -->
		<div class="text-xs text-gray-500 capitalize dark:text-gray-400">
			{data.status}
		</div>

		<!-- Last Updated -->
		{#if data.timestamp}
			<div class="mt-1 text-xs text-gray-400 dark:text-gray-500">
				Updated: {new Date(data.timestamp).toLocaleTimeString()}
			</div>
		{/if}
	</div>
</Card>
