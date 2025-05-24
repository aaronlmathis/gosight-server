<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import { formatDate } from '$lib/utils';
	import type { LogEntry, Endpoint } from '$lib/types';

	let logs: LogEntry[] = [];
	let endpoints: Endpoint[] = [];
	let loading = true;
	let error = '';
	let autoRefresh = true;
	let searchTerm = '';
	let selectedEndpoint = '';
	let selectedLevel = '';
	let selectedTimeRange = '1h';
	let sortOrder = 'desc';

	// Pagination
	let currentPage = 1;
	let pageSize = 50;
	let totalLogs = 0;

	// Log levels with colors
	const logLevels = [
		{ value: '', label: 'All Levels', color: '' },
		{ value: 'error', label: 'Error', color: 'text-red-600 bg-red-100' },
		{ value: 'warn', label: 'Warning', color: 'text-yellow-600 bg-yellow-100' },
		{ value: 'info', label: 'Info', color: 'text-blue-600 bg-blue-100' },
		{ value: 'debug', label: 'Debug', color: 'text-gray-600 bg-gray-100' }
	];

	const timeRanges = [
		{ value: '15m', label: '15 minutes' },
		{ value: '1h', label: '1 hour' },
		{ value: '6h', label: '6 hours' },
		{ value: '24h', label: '24 hours' },
		{ value: '7d', label: '7 days' }
	];

	onMount(async () => {
		await loadEndpoints();
		await loadLogs();
		
		// Subscribe to real-time log updates
		if (autoRefresh) {
			websocketManager.connect();
			websocketManager.subscribeToLogs((log: LogEntry) => {
				// Add new log to the beginning if it matches current filters
				if (matchesFilters(log)) {
					logs = [log, ...logs];
					// Keep only the latest logs to prevent memory issues
					if (logs.length > pageSize * 3) {
						logs = logs.slice(0, pageSize * 3);
					}
				}
			});
		}
	});

	async function loadEndpoints() {
		try {
			const response = await api.getEndpoints();
			endpoints = response.endpoints || [];
		} catch (err) {
			console.error('Failed to load endpoints:', err);
		}
	}

	async function loadLogs() {
		try {
			loading = true;
			error = '';
			
			const params: any = {
				page: currentPage,
				limit: pageSize,
				sort: sortOrder
			};

			if (searchTerm) params.search = searchTerm;
			if (selectedEndpoint) params.endpoint_id = selectedEndpoint;
			if (selectedLevel) params.level = selectedLevel;
			if (selectedTimeRange) params.time_range = selectedTimeRange;

			const response = await api.getLogs(params);
			logs = response.logs || [];
			totalLogs = response.total || 0;
		} catch (err) {
			error = 'Failed to load logs: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	function matchesFilters(log: LogEntry): boolean {
		if (selectedEndpoint && log.endpoint_id !== selectedEndpoint) return false;
		if (selectedLevel && log.level !== selectedLevel) return false;
		if (searchTerm && !log.message.toLowerCase().includes(searchTerm.toLowerCase())) return false;
		return true;
	}

	function handleFilterChange() {
		currentPage = 1;
		loadLogs();
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		if (autoRefresh) {
			websocketManager.connect();
			websocketManager.subscribeToLogs((log: LogEntry) => {
				if (matchesFilters(log)) {
					logs = [log, ...logs];
				}
			});
		} else {
			websocketManager.disconnect();
		}
	}

	function clearLogs() {
		logs = [];
	}

	function exportLogs() {
		const csvContent = logs.map(log => 
			`"${formatDate(log.timestamp)}","${log.level}","${log.source || ''}","${log.message.replace(/"/g, '""')}"`
		).join('\n');
		
		const header = 'Timestamp,Level,Source,Message\n';
		const blob = new Blob([header + csvContent], { type: 'text/csv' });
		const url = window.URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `logs-${new Date().toISOString().split('T')[0]}.csv`;
		a.click();
		window.URL.revokeObjectURL(url);
	}

	function getLogLevelColor(level: string): string {
		const levelObj = logLevels.find(l => l.value === level);
		return levelObj?.color || 'text-gray-600 bg-gray-100';
	}

	function getLogLevelIcon(level: string): string {
		switch (level) {
			case 'error': return 'fas fa-exclamation-circle';
			case 'warn': return 'fas fa-exclamation-triangle';
			case 'info': return 'fas fa-info-circle';
			case 'debug': return 'fas fa-bug';
			default: return 'fas fa-file-alt';
		}
	}

	$: totalPages = Math.ceil(totalLogs / pageSize);
</script>

<svelte:head>
	<title>Logs - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6 flex justify-between items-center">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Logs</h1>
			<p class="text-gray-600 dark:text-gray-400">View and analyze system logs</p>
		</div>
		<div class="flex gap-2">
			<button
				on:click={toggleAutoRefresh}
				class="px-4 py-2 text-sm border rounded-lg transition-colors {autoRefresh ? 'bg-green-100 text-green-800 border-green-300 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800' : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600 dark:hover:bg-gray-600'}"
			>
				<i class="fas fa-{autoRefresh ? 'pause' : 'play'} mr-2"></i>
				{autoRefresh ? 'Pause' : 'Start'} Auto-refresh
			</button>
			<button
				on:click={exportLogs}
				disabled={logs.length === 0}
				class="px-4 py-2 text-sm bg-blue-600 text-white border border-transparent rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				<i class="fas fa-download mr-2"></i>
				Export CSV
			</button>
		</div>
	</div>

	<!-- Filters -->
	<div class="mb-6 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
			<!-- Search -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
					Search
				</label>
				<div class="relative">
					<i class="fas fa-search absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"></i>
					<input
						type="text"
						placeholder="Search logs..."
						bind:value={searchTerm}
						on:input={handleFilterChange}
						class="pl-10 pr-4 py-2 w-full border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
					/>
				</div>
			</div>

			<!-- Endpoint Filter -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
					Endpoint
				</label>
				<select
					bind:value={selectedEndpoint}
					on:change={handleFilterChange}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					<option value="">All Endpoints</option>
					{#each endpoints as endpoint}
						<option value={endpoint.id}>{endpoint.name}</option>
					{/each}
				</select>
			</div>

			<!-- Level Filter -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
					Level
				</label>
				<select
					bind:value={selectedLevel}
					on:change={handleFilterChange}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					{#each logLevels as level}
						<option value={level.value}>{level.label}</option>
					{/each}
				</select>
			</div>

			<!-- Time Range -->
			<div>
				<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
					Time Range
				</label>
				<select
					bind:value={selectedTimeRange}
					on:change={handleFilterChange}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					{#each timeRanges as range}
						<option value={range.value}>{range.label}</option>
					{/each}
				</select>
			</div>

			<!-- Actions -->
			<div class="flex items-end gap-2">
				<button
					on:click={loadLogs}
					disabled={loading}
					class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				>
					{#if loading}
						<i class="fas fa-spinner fa-spin mr-2"></i>
					{:else}
						<i class="fas fa-sync-alt mr-2"></i>
					{/if}
					Refresh
				</button>
				<button
					on:click={clearLogs}
					disabled={logs.length === 0}
					class="px-4 py-2 text-gray-600 dark:text-gray-400 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					title="Clear current logs"
				>
					<i class="fas fa-trash"></i>
				</button>
			</div>
		</div>
	</div>

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

	<!-- Logs Display -->
	<div class="bg-white dark:bg-gray-800 rounded-lg shadow">
		<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
			<h3 class="text-lg font-medium text-gray-900 dark:text-white">
				Log Entries
				{#if totalLogs > 0}
					<span class="text-sm font-normal text-gray-500 dark:text-gray-400">
						({totalLogs} total)
					</span>
				{/if}
			</h3>
			<div class="flex items-center gap-4">
				<label class="flex items-center text-sm text-gray-600 dark:text-gray-400">
					Sort:
					<select
						bind:value={sortOrder}
						on:change={loadLogs}
						class="ml-2 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
					>
						<option value="desc">Newest first</option>
						<option value="asc">Oldest first</option>
					</select>
				</label>
			</div>
		</div>

		{#if loading}
			<div class="flex justify-center items-center py-12">
				<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
			</div>
		{:else if logs.length === 0}
			<div class="text-center py-12">
				<i class="fas fa-file-alt text-4xl text-gray-400 mb-4"></i>
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">No Logs Found</h3>
				<p class="text-gray-600 dark:text-gray-400">No log entries match your current filters.</p>
			</div>
		{:else}
			<div class="overflow-hidden">
				{#each logs as log, index (log.id || index)}
					<div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700 last:border-b-0 hover:bg-gray-50 dark:hover:bg-gray-700/50">
						<div class="flex items-start gap-4">
							<div class="flex-shrink-0 mt-1">
								<i class="{getLogLevelIcon(log.level)} text-lg {getLogLevelColor(log.level).split(' ')[0]}"></i>
							</div>
							
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-3 mb-2">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getLogLevelColor(log.level)}">
										{log.level}
									</span>
									<span class="text-sm text-gray-500 dark:text-gray-400">
										{formatDate(log.timestamp)}
									</span>
									{#if log.source}
										<span class="text-sm text-gray-600 dark:text-gray-300 font-mono">
											{log.source}
										</span>
									{/if}
									{#if log.endpoint_name}
										<span class="text-sm text-blue-600 dark:text-blue-400">
											{log.endpoint_name}
										</span>
									{/if}
								</div>
								
								<div class="text-sm text-gray-900 dark:text-white font-mono whitespace-pre-wrap break-words">
									{log.message}
								</div>
								
								{#if log.metadata && Object.keys(log.metadata).length > 0}
									<details class="mt-2">
										<summary class="text-xs text-gray-500 dark:text-gray-400 cursor-pointer hover:text-gray-700 dark:hover:text-gray-300">
											<i class="fas fa-info-circle mr-1"></i>
											Show metadata
										</summary>
										<div class="mt-2 p-3 bg-gray-50 dark:bg-gray-700 rounded text-xs font-mono">
											<pre>{JSON.stringify(log.metadata, null, 2)}</pre>
										</div>
									</details>
								{/if}
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
					<div class="text-sm text-gray-700 dark:text-gray-300">
						Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, totalLogs)} of {totalLogs} entries
					</div>
					<div class="flex gap-2">
						<button
							on:click={() => currentPage > 1 && (currentPage--, loadLogs())}
							disabled={currentPage === 1}
							class="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-gray-700 dark:text-gray-300"
						>
							<i class="fas fa-chevron-left mr-1"></i>
							Previous
						</button>
						<span class="px-3 py-1 text-sm text-gray-700 dark:text-gray-300">
							Page {currentPage} of {totalPages}
						</span>
						<button
							on:click={() => currentPage < totalPages && (currentPage++, loadLogs())}
							disabled={currentPage === totalPages}
							class="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed text-gray-700 dark:text-gray-300"
						>
							Next
							<i class="fas fa-chevron-right ml-1"></i>
						</button>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>
