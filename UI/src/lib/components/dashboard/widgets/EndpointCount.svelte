<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Card, Badge } from 'flowbite-svelte';
	import { Server, Container, Globe, CheckCircle, XCircle, Loader } from 'lucide-svelte';
	import type { Widget, WidgetData } from '$lib/types/dashboard';
	import { dataService } from '$lib/services/dataService';

	export let widget: Widget;

	let data: WidgetData | null = null;
	let loading = true;
	let error = '';

	// Extract configuration with defaults
	$: config = widget.config || {};
	$: showHostCount = config.showHostCount ?? true;
	$: showContainerCount = config.showContainerCount ?? true;
	$: showTotalCount = config.showTotalCount ?? true;
	$: showOnlineStatus = config.showOnlineStatus ?? false;

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
			console.error(`Failed to load endpoint count data for widget ${widget.id}:`, err);
			error = err instanceof Error ? err.message : 'Failed to load endpoint count data';
			data = { status: 'error', value: 0, error };
		} finally {
			loading = false;
		}
	}

	// Format numbers with commas
	function formatNumber(num: number): string {
		return num.toLocaleString();
	}

	// Get status info for display
	function getStatusInfo(count: number) {
		if (count === 0) {
			return {
				icon: XCircle,
				color: 'text-red-500',
				bgColor: 'bg-red-100',
				label: 'None'
			};
		}
		return {
			icon: CheckCircle,
			color: 'text-green-500',
			bgColor: 'bg-green-100',
			label: 'Active'
		};
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
		<Globe class="h-5 w-5 text-gray-500" />
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
			<!-- Total Count -->
			{#if showTotalCount}
				<div class="text-center">
					<div class="text-3xl font-bold text-gray-900 dark:text-white">
						{formatNumber(data.value || 0)}
					</div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Total Endpoints</div>
				</div>
			{/if}

			<!-- Detailed Counts -->
			{#if data.details && (showHostCount || showContainerCount)}
				<div class="grid grid-cols-1 gap-3">
					{#if showHostCount && data.details.hosts !== undefined}
						<div
							class="flex items-center justify-between rounded-lg bg-white/50 p-3 dark:bg-gray-800/50"
						>
							<div class="flex items-center space-x-2">
								<Server class="h-4 w-4 text-blue-500" />
								<span class="text-sm font-medium text-gray-700 dark:text-gray-300"> Hosts </span>
							</div>
							<div class="flex items-center space-x-2">
								<span class="text-sm font-bold text-gray-900 dark:text-white">
									{formatNumber(data.details.hosts)}
								</span>
								{#if showOnlineStatus}
									{@const statusInfo = getStatusInfo(data.details.hosts)}
									<Badge color={data.details.hosts > 0 ? 'green' : 'red'} class="text-xs">
										<svelte:component this={statusInfo.icon} class="mr-1 h-3 w-3" />
										{statusInfo.label}
									</Badge>
								{/if}
							</div>
						</div>
					{/if}

					{#if showContainerCount && data.details.containers !== undefined}
						<div
							class="flex items-center justify-between rounded-lg bg-white/50 p-3 dark:bg-gray-800/50"
						>
							<div class="flex items-center space-x-2">
								<Container class="h-4 w-4 text-purple-500" />
								<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
									Containers
								</span>
							</div>
							<div class="flex items-center space-x-2">
								<span class="text-sm font-bold text-gray-900 dark:text-white">
									{formatNumber(data.details.containers)}
								</span>
								{#if showOnlineStatus}
									{@const statusInfo = getStatusInfo(data.details.containers)}
									<Badge color={data.details.containers > 0 ? 'green' : 'red'} class="text-xs">
										<svelte:component this={statusInfo.icon} class="mr-1 h-3 w-3" />
										{statusInfo.label}
									</Badge>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Status Summary -->
			{#if showOnlineStatus && data.value !== undefined}
				{@const statusInfo = getStatusInfo(data.value)}
				<div class="border-t border-gray-200 pt-2 dark:border-gray-700">
					<div class="flex items-center justify-center space-x-2">
						<svelte:component this={statusInfo.icon} class="h-4 w-4 {statusInfo.color}" />
						<span class="text-sm text-gray-600 dark:text-gray-400">
							{data.value > 0
								? `${formatNumber(data.value)} endpoints online`
								: 'No endpoints available'}
						</span>
					</div>
				</div>
			{/if}
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
			<Globe class="mb-2 h-8 w-8 text-gray-400" />
			<p class="text-sm text-gray-500 dark:text-gray-400">No endpoint data available</p>
		</div>
	{/if}
</Card>
