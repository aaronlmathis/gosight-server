<script lang="ts">
	import { onMount } from 'svelte';
	import type { LogEntry } from '$lib/types';
	import { formatDate, getLevelBadgeClass } from '$lib/utils';
	import { api } from '$lib/api';

	export let logs: LogEntry[] = [];
	export let endpointId: string;

	let loading = false;
	let error = '';

	function getBadgeClass(level: string): string {
		return getLevelBadgeClass(level);
	}

	async function fetchInitialLogs() {
		if (!endpointId) return;

		try {
			loading = true;
			error = '';
			console.log('Fetching initial logs for endpoint:', endpointId);

			const response = await api.getLogs({
				endpoint_id: endpointId,
				limit: 50,
				order: 'desc'
			});

			if (response?.logs && Array.isArray(response.logs)) {
				// Merge initial logs with any existing logs from websocket, avoiding duplicates
				const existingIds = new Set(logs.map((log) => log.id || `${log.timestamp}-${log.message}`));
				const newLogs = response.logs.filter((log) => {
					const logId = log.id || `${log.timestamp}-${log.message}`;
					return !existingIds.has(logId);
				});

				logs = [...newLogs, ...logs].slice(0, 100); // Keep latest 100 logs
				console.log('Loaded initial logs:', newLogs.length, 'total logs:', logs.length);
			}
		} catch (err) {
			console.error('Failed to fetch initial logs:', err);
			error = err instanceof Error ? err.message : 'Failed to load logs';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchInitialLogs();
	});
</script>

<div
	class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
	id="logs"
	role="tabpanel"
	aria-labelledby="logs-tab"
>
	<div class="rounded-lg bg-white shadow dark:bg-gray-800">
		<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
			<div class="flex items-center justify-between">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">System Logs</h3>
				{#if loading}
					<div class="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
						<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-600"></div>
						<span>Loading logs...</span>
					</div>
				{/if}
			</div>
			{#if error}
				<div class="mt-2 text-sm text-red-600 dark:text-red-400">
					{error}
				</div>
			{/if}
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
					{#if loading}
						<div class="flex items-center justify-center space-x-2">
							<div class="h-5 w-5 animate-spin rounded-full border-b-2 border-blue-600"></div>
							<p class="text-gray-500 dark:text-gray-400">Loading logs...</p>
						</div>
					{:else}
						<p class="text-gray-500 dark:text-gray-400">No logs found</p>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>
