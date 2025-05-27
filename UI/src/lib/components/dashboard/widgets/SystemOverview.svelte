<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { Widget } from '$lib/types/dashboard';
	import { dataService } from '$lib/services/dataService';
	import { Card, Badge, Progressbar } from 'flowbite-svelte';
	import { Activity, HardDrive, MemoryStick, Clock, Zap } from 'lucide-svelte';

	export let widget: Widget;

	// Component state
	let loading = true;
	let error: string | null = null;
	let systemData: any = {};
	let refreshInterval: number;

	// Widget configuration
	$: config = widget.config || {};
	$: showCpuUsage = config.showCpuUsage !== false;
	$: showMemoryUsage = config.showMemoryUsage !== false;
	$: showDiskUsage = config.showDiskUsage !== false;
	$: showUptime = config.showUptime === true;
	$: showLoadAverage = config.showLoadAverage === true;
	$: refreshRate = (config.refreshInterval || 30) * 1000;

	// Load system overview data
	async function loadSystemData() {
		try {
			loading = true;
			error = null;

			const data = await dataService.getWidgetData(widget);

			if (data.status === 'error') {
				error = data.error || 'Failed to load system data';
				return;
			}

			systemData = data.metrics || {};
		} catch (err) {
			console.error('Failed to load system data:', err);
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}

	// Format bytes to human readable
	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
	}

	// Format uptime
	function formatUptime(seconds: number): string {
		const days = Math.floor(seconds / 86400);
		const hours = Math.floor((seconds % 86400) / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);

		if (days > 0) {
			return `${days}d ${hours}h ${minutes}m`;
		} else if (hours > 0) {
			return `${hours}h ${minutes}m`;
		} else {
			return `${minutes}m`;
		}
	}

	// Get progress bar color based on usage percentage
	function getProgressColor(percentage: number): 'green' | 'yellow' | 'red' {
		if (percentage < 70) return 'green';
		if (percentage < 90) return 'yellow';
		return 'red';
	}

	// Component lifecycle
	onMount(() => {
		loadSystemData();

		// Set up refresh interval
		refreshInterval = setInterval(loadSystemData, refreshRate);
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});

	// Update refresh interval when config changes
	$: if (refreshInterval) {
		clearInterval(refreshInterval);
		refreshInterval = setInterval(loadSystemData, refreshRate);
	}
</script>

<Card class="h-full p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium text-gray-900 dark:text-white">
			{widget.title}
		</h3>
		<Activity class="h-5 w-5 text-blue-600" />
	</div>

	{#if loading}
		<div class="flex h-32 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="flex h-32 items-center justify-center">
			<div class="text-center">
				<div class="mb-2 text-sm text-red-500">⚠️ Error</div>
				<p class="text-xs text-gray-500">{error}</p>
			</div>
		</div>
	{:else}
		<div class="space-y-4">
			<!-- CPU Usage -->
			{#if showCpuUsage && systemData.cpu}
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<Zap class="h-4 w-4 text-blue-600" />
							<span class="text-sm font-medium">CPU</span>
						</div>
						<Badge color={getProgressColor(systemData.cpu.usage_percent || 0)}>
							{(systemData.cpu.usage_percent || 0).toFixed(1)}%
						</Badge>
					</div>
					<Progressbar
						progress={systemData.cpu.usage_percent || 0}
						color={getProgressColor(systemData.cpu.usage_percent || 0)}
						size="sm"
					/>
				</div>
			{/if}

			<!-- Memory Usage -->
			{#if showMemoryUsage && systemData.memory}
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<MemoryStick class="h-4 w-4 text-green-600" />
							<span class="text-sm font-medium">Memory</span>
						</div>
						<div class="text-right">
							<Badge color={getProgressColor(systemData.memory.used_percent || 0)}>
								{(systemData.memory.used_percent || 0).toFixed(1)}%
							</Badge>
							<div class="text-xs text-gray-500">
								{formatBytes(systemData.memory.used || 0)} / {formatBytes(
									systemData.memory.total || 0
								)}
							</div>
						</div>
					</div>
					<Progressbar
						progress={systemData.memory.used_percent || 0}
						color={getProgressColor(systemData.memory.used_percent || 0)}
						size="sm"
					/>
				</div>
			{/if}

			<!-- Disk Usage -->
			{#if showDiskUsage && systemData.disk}
				<div class="space-y-2">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<HardDrive class="h-4 w-4 text-purple-600" />
							<span class="text-sm font-medium">Disk</span>
						</div>
						<div class="text-right">
							<Badge color={getProgressColor(systemData.disk.used_percent || 0)}>
								{(systemData.disk.used_percent || 0).toFixed(1)}%
							</Badge>
							<div class="text-xs text-gray-500">
								{formatBytes(systemData.disk.used || 0)} / {formatBytes(systemData.disk.total || 0)}
							</div>
						</div>
					</div>
					<Progressbar
						progress={systemData.disk.used_percent || 0}
						color={getProgressColor(systemData.disk.used_percent || 0)}
						size="sm"
					/>
				</div>
			{/if}

			<!-- Uptime -->
			{#if showUptime && systemData.host?.uptime}
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<Clock class="h-4 w-4 text-orange-600" />
						<span class="text-sm font-medium">Uptime</span>
					</div>
					<Badge color="blue">
						{formatUptime(systemData.host.uptime)}
					</Badge>
				</div>
			{/if}

			<!-- Load Average -->
			{#if showLoadAverage && systemData.cpu?.load_avg}
				<div class="space-y-1">
					<div class="flex items-center gap-2">
						<Activity class="h-4 w-4 text-indigo-600" />
						<span class="text-sm font-medium">Load Average</span>
					</div>
					<div class="grid grid-cols-3 gap-2 text-center">
						<div class="text-xs">
							<div class="font-medium">{(systemData.cpu.load_avg.load1 || 0).toFixed(2)}</div>
							<div class="text-gray-500">1m</div>
						</div>
						<div class="text-xs">
							<div class="font-medium">{(systemData.cpu.load_avg.load5 || 0).toFixed(2)}</div>
							<div class="text-gray-500">5m</div>
						</div>
						<div class="text-xs">
							<div class="font-medium">{(systemData.cpu.load_avg.load15 || 0).toFixed(2)}</div>
							<div class="text-gray-500">15m</div>
						</div>
					</div>
				</div>
			{/if}

			<!-- No data message -->
			{#if !showCpuUsage && !showMemoryUsage && !showDiskUsage && !showUptime && !showLoadAverage}
				<div class="py-8 text-center text-gray-500">
					<p class="text-sm">No metrics selected</p>
					<p class="text-xs">Configure this widget to display system metrics</p>
				</div>
			{/if}
		</div>
	{/if}
</Card>
