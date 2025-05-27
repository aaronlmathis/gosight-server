<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Card, Badge } from 'flowbite-svelte';
	import { AlertTriangle, Shield, AlertCircle, XCircle, Clock, Loader } from 'lucide-svelte';
	import type { Widget, WidgetData } from '$lib/types/dashboard';
	import { dataService } from '$lib/services/dataService';

	export let widget: Widget;

	let data: WidgetData | null = null;
	let loading = true;
	let error = '';

	// Extract configuration with defaults
	$: config = widget.config || {};
	$: showActiveCount = config.showActiveCount ?? true;
	$: showBySeverity = config.showBySeverity ?? false;
	$: showRecentAlerts = config.showRecentAlerts ?? false;
	$: alertTimeRange = config.alertTimeRange || '24h';

	// Calculate status background
	$: statusBg = {
		success: 'bg-green-50 border-green-200 dark:bg-green-900/20 dark:border-green-800',
		warning: 'bg-yellow-50 border-yellow-200 dark:bg-yellow-900/20 dark:border-yellow-800',
		error: 'bg-red-50 border-red-200 dark:bg-red-900/20 dark:border-red-800',
		unknown: 'bg-gray-50 border-gray-200 dark:bg-gray-900/20 dark:border-gray-800'
	}[data?.status || 'unknown'];

	// Load widget data
	async function loadData() {
		try {
			loading = true;
			error = '';
			data = await dataService.getWidgetData(widget);
		} catch (err) {
			console.error(`Failed to load alert count data for widget ${widget.id}:`, err);
			error = err instanceof Error ? err.message : 'Failed to load alert count data';
			data = { status: 'error', value: 0, error };
		} finally {
			loading = false;
		}
	}

	// Format numbers with commas
	function formatNumber(num: number): string {
		return num.toLocaleString();
	}

	// Get severity info for display
	function getSeverityInfo(severity: string) {
		switch (severity.toLowerCase()) {
			case 'critical':
				return {
					icon: XCircle,
					color: 'text-red-500',
					bgColor: 'bg-red-100',
					badgeColor: 'red'
				};
			case 'warning':
				return {
					icon: AlertTriangle,
					color: 'text-yellow-500',
					bgColor: 'bg-yellow-100',
					badgeColor: 'yellow'
				};
			case 'info':
				return {
					icon: AlertCircle,
					color: 'text-blue-500',
					bgColor: 'bg-blue-100',
					badgeColor: 'blue'
				};
			default:
				return {
					icon: Shield,
					color: 'text-gray-500',
					bgColor: 'bg-gray-100',
					badgeColor: 'gray'
				};
		}
	}

	// Get status message
	function getStatusMessage(activeCount: number): string {
		if (activeCount === 0) {
			return 'All systems operational';
		} else if (activeCount === 1) {
			return '1 active alert';
		} else {
			return `${formatNumber(activeCount)} active alerts`;
		}
	}

	// Format time range for display
	function formatTimeRange(timeRange: string): string {
		switch (timeRange) {
			case '1h':
				return 'Last hour';
			case '6h':
				return 'Last 6 hours';
			case '12h':
				return 'Last 12 hours';
			case '24h':
				return 'Last 24 hours';
			case '7d':
				return 'Last 7 days';
			case '30d':
				return 'Last 30 days';
			default:
				return timeRange;
		}
	}

	let refreshInterval: number;

	onMount(() => {
		loadData();

		// Set up periodic refresh
		refreshInterval = setInterval(loadData, widget.config?.refreshInterval || 30000);
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});
</script>

<Card class="h-full {statusBg} relative border-2">
	<!-- Loading Overlay -->
	{#if loading}
		<div
			class="absolute inset-0 z-10 flex items-center justify-center bg-white/50 backdrop-blur-sm dark:bg-gray-900/50"
		>
			<Loader class="h-6 w-6 animate-spin text-blue-500" />
		</div>
	{/if}

	<!-- Widget Header -->
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white">
			{widget.title}
		</h3>
		<AlertTriangle class="h-5 w-5 text-gray-500" />
	</div>

	{#if error}
		<!-- Error State -->
		<div class="flex h-32 flex-col items-center justify-center text-center">
			<XCircle class="mb-2 h-8 w-8 text-red-500" />
			<p class="text-sm text-red-600 dark:text-red-400">{error}</p>
		</div>
	{:else if data}
		<!-- Content -->
		<div class="space-y-4">
			<!-- Active Count -->
			{#if showActiveCount}
				<div class="text-center">
					<div
						class="text-3xl font-bold {(data.value || 0) > 0
							? 'text-red-600'
							: 'text-green-600'} dark:text-white"
					>
						{formatNumber(data.value || 0)}
					</div>
					<div class="text-sm text-gray-500 dark:text-gray-400">
						{getStatusMessage(data.value || 0)}
					</div>
				</div>
			{/if}

			<!-- Severity Breakdown -->
			{#if showBySeverity && data.details}
				<div class="space-y-2">
					<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">By Severity</h4>
					<div class="grid grid-cols-1 gap-2">
						{#each Object.entries(data.details) as [severity, count]}
							{@const severityInfo = getSeverityInfo(severity)}
							<div
								class="flex items-center justify-between rounded-lg bg-white/50 p-2 dark:bg-gray-800/50"
							>
								<div class="flex items-center space-x-2">
									<svelte:component this={severityInfo.icon} class="h-4 w-4 {severityInfo.color}" />
									<span class="text-sm font-medium text-gray-700 capitalize dark:text-gray-300">
										{severity}
									</span>
								</div>
								<Badge color="primary" class="text-xs">
									{formatNumber(count)}
								</Badge>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Recent Alerts Info -->
			{#if showRecentAlerts}
				<div class="border-t border-gray-200 pt-2 dark:border-gray-700">
					<div class="flex items-center justify-between text-sm">
						<div class="flex items-center space-x-1 text-gray-500 dark:text-gray-400">
							<Clock class="h-3 w-3" />
							<span>{formatTimeRange(alertTimeRange)}</span>
						</div>
						{#if (data.value || 0) > 0}
							<button
								type="button"
								class="inline-flex items-center rounded bg-red-600 px-2 py-1 text-xs font-medium text-white hover:bg-red-700 focus:ring-4 focus:ring-red-300 focus:outline-none dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
							>
								View Details
							</button>
						{/if}
					</div>
				</div>
			{/if}

			<!-- Status Indicator -->
			<div
				class="flex items-center justify-center rounded-lg p-2 {(data.value || 0) > 0
					? 'bg-red-50 dark:bg-red-900/20'
					: 'bg-green-50 dark:bg-green-900/20'}"
			>
				{#if (data.value || 0) === 0}
					<div class="flex items-center space-x-2 text-green-600 dark:text-green-400">
						<Shield class="h-4 w-4" />
						<span class="text-sm font-medium">System Health: Good</span>
					</div>
				{:else if (data.value || 0) <= 5}
					<div class="flex items-center space-x-2 text-yellow-600 dark:text-yellow-400">
						<AlertTriangle class="h-4 w-4" />
						<span class="text-sm font-medium">System Health: Attention Needed</span>
					</div>
				{:else}
					<div class="flex items-center space-x-2 text-red-600 dark:text-red-400">
						<XCircle class="h-4 w-4" />
						<span class="text-sm font-medium">System Health: Critical</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- Timestamp -->
		{#if data.timestamp}
			<div class="absolute right-2 bottom-2 text-xs text-gray-400">
				{new Date(data.timestamp).toLocaleTimeString()}
			</div>
		{/if}
	{:else}
		<!-- Empty State -->
		<div class="flex h-32 flex-col items-center justify-center text-center">
			<Shield class="mb-2 h-8 w-8 text-gray-400" />
			<p class="text-sm text-gray-500 dark:text-gray-400">No alert data available</p>
		</div>
	{/if}
</Card>
