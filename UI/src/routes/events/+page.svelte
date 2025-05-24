<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import { formatDate } from '$lib/utils';
	import type { Event, Endpoint } from '$lib/types';

	let events: Event[] = [];
	let endpoints: Endpoint[] = [];
	let loading = true;
	let error = '';
	let autoRefresh = true;
	let searchTerm = '';
	let selectedEndpoint = '';
	let selectedType = '';
	let selectedTimeRange = '24h';

	// Pagination
	let currentPage = 1;
	let pageSize = 30;
	let totalEvents = 0;

	// Event types with colors and icons
	const eventTypes = [
		{ value: '', label: 'All Types', color: '', icon: '' },
		{
			value: 'endpoint_up',
			label: 'Endpoint Up',
			color: 'text-green-600 bg-green-100',
			icon: 'fas fa-arrow-up'
		},
		{
			value: 'endpoint_down',
			label: 'Endpoint Down',
			color: 'text-red-600 bg-red-100',
			icon: 'fas fa-arrow-down'
		},
		{
			value: 'alert_triggered',
			label: 'Alert Triggered',
			color: 'text-orange-600 bg-orange-100',
			icon: 'fas fa-exclamation-triangle'
		},
		{
			value: 'alert_resolved',
			label: 'Alert Resolved',
			color: 'text-blue-600 bg-blue-100',
			icon: 'fas fa-check-circle'
		},
		{
			value: 'maintenance_start',
			label: 'Maintenance Started',
			color: 'text-purple-600 bg-purple-100',
			icon: 'fas fa-tools'
		},
		{
			value: 'maintenance_end',
			label: 'Maintenance Ended',
			color: 'text-purple-600 bg-purple-100',
			icon: 'fas fa-tools'
		},
		{
			value: 'configuration_change',
			label: 'Configuration Change',
			color: 'text-yellow-600 bg-yellow-100',
			icon: 'fas fa-cog'
		},
		{
			value: 'user_action',
			label: 'User Action',
			color: 'text-indigo-600 bg-indigo-100',
			icon: 'fas fa-user'
		}
	];

	const timeRanges = [
		{ value: '1h', label: '1 hour' },
		{ value: '6h', label: '6 hours' },
		{ value: '24h', label: '24 hours' },
		{ value: '7d', label: '7 days' },
		{ value: '30d', label: '30 days' }
	];

	onMount(async () => {
		await loadEndpoints();
		await loadEvents();

		// Subscribe to real-time event updates
		if (autoRefresh) {
			websocketManager.connect();
			websocketManager.subscribeToEvents((event: Event) => {
				if (matchesFilters(event)) {
					events = [event, ...events];
					// Keep only recent events to prevent memory issues
					if (events.length > pageSize * 3) {
						events = events.slice(0, pageSize * 3);
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

	async function loadEvents() {
		try {
			loading = true;
			error = '';

			const params: any = {
				page: currentPage,
				limit: pageSize
			};

			if (searchTerm) params.search = searchTerm;
			if (selectedEndpoint) params.endpoint_id = selectedEndpoint;
			if (selectedType) params.type = selectedType;
			if (selectedTimeRange) params.time_range = selectedTimeRange;

			const response = await api.getEvents(params);
			events = response.events || [];
			totalEvents = response.total || 0;
		} catch (err) {
			error = 'Failed to load events: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	function matchesFilters(event: Event): boolean {
		if (selectedEndpoint && event.endpoint_id !== selectedEndpoint) return false;
		if (selectedType && event.type !== selectedType) return false;
		if (searchTerm && !event.description?.toLowerCase().includes(searchTerm.toLowerCase()))
			return false;
		return true;
	}

	function handleFilterChange() {
		currentPage = 1;
		loadEvents();
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		if (autoRefresh) {
			websocketManager.connect();
			websocketManager.subscribeToEvents((event: Event) => {
				if (matchesFilters(event)) {
					events = [event, ...events];
				}
			});
		} else {
			websocketManager.disconnect();
		}
	}

	function getEventTypeConfig(type: string) {
		return eventTypes.find((t) => t.value === type) || eventTypes[0];
	}

	function getTimeAgo(timestamp: string): string {
		const now = new Date();
		const eventTime = new Date(timestamp);
		const diffMs = now.getTime() - eventTime.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 30) return `${diffDays}d ago`;
		return formatDate(timestamp);
	}

	$: totalPages = Math.ceil(totalEvents / pageSize);
</script>

<svelte:head>
	<title>Events - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Events</h1>
			<p class="text-gray-600 dark:text-gray-400">Activity stream and system events</p>
		</div>
		<button
			on:click={toggleAutoRefresh}
			class="rounded-lg border px-4 py-2 text-sm transition-colors {autoRefresh
				? 'border-green-300 bg-green-100 text-green-800 dark:border-green-800 dark:bg-green-900/20 dark:text-green-400'
				: 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
		>
			<i class="fas fa-{autoRefresh ? 'pause' : 'play'} mr-2"></i>
			{autoRefresh ? 'Pause' : 'Start'} Auto-refresh
		</button>
	</div>

	<!-- Filters -->
	<div class="mb-6 rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
			<!-- Search -->
			<div>
				<label
					for="search-events"
					class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
				>
					Search
				</label>
				<div class="relative">
					<i class="fas fa-search absolute top-1/2 left-3 -translate-y-1/2 transform text-gray-400"
					></i>
					<input
						id="search-events"
						type="text"
						placeholder="Search events..."
						bind:value={searchTerm}
						on:input={handleFilterChange}
						class="w-full rounded-lg border border-gray-300 bg-white py-2 pr-4 pl-10 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
					/>
				</div>
			</div>

			<!-- Endpoint Filter -->
			<div>
				<label
					for="endpoint-filter"
					class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
				>
					Endpoint
				</label>
				<select
					id="endpoint-filter"
					bind:value={selectedEndpoint}
					on:change={handleFilterChange}
					class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
				>
					<option value="">All Endpoints</option>
					{#each endpoints as endpoint}
						<option value={endpoint.id}>{endpoint.name}</option>
					{/each}
				</select>
			</div>

			<!-- Type Filter -->
			<div>
				<label
					for="type-filter"
					class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
				>
					Type
				</label>
				<select
					id="type-filter"
					bind:value={selectedType}
					on:change={handleFilterChange}
					class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
				>
					{#each eventTypes as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
			</div>

			<!-- Time Range -->
			<div>
				<label
					for="time-range-filter"
					class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
				>
					Time Range
				</label>
				<select
					id="time-range-filter"
					bind:value={selectedTimeRange}
					on:change={handleFilterChange}
					class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
				>
					{#each timeRanges as range}
						<option value={range.value}>{range.label}</option>
					{/each}
				</select>
			</div>
		</div>
	</div>

	{#if error}
		<div
			class="mb-6 rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
		>
			<div class="flex">
				<i class="fas fa-exclamation-triangle mt-0.5 mr-3 text-red-500"></i>
				<div>
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<p class="mt-1 text-sm text-red-600 dark:text-red-300">{error}</p>
				</div>
			</div>
		</div>
	{/if}

	<!-- Events Timeline -->
	<div class="rounded-lg bg-white shadow dark:bg-gray-800">
		<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
			<h3 class="text-lg font-medium text-gray-900 dark:text-white">
				Activity Timeline
				{#if totalEvents > 0}
					<span class="text-sm font-normal text-gray-500 dark:text-gray-400">
						({totalEvents} total)
					</span>
				{/if}
			</h3>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
			</div>
		{:else if events.length === 0}
			<div class="py-12 text-center">
				<i class="fas fa-calendar-alt mb-4 text-4xl text-gray-400"></i>
				<h3 class="mb-2 text-lg font-medium text-gray-900 dark:text-white">No Events Found</h3>
				<p class="text-gray-600 dark:text-gray-400">No events match your current filters.</p>
			</div>
		{:else}
			<div class="divide-y divide-gray-200 dark:divide-gray-700">
				{#each events as event, index (event.id)}
					{@const typeConfig = getEventTypeConfig(event.type)}
					<div class="px-6 py-4 hover:bg-gray-50 dark:hover:bg-gray-700/50">
						<div class="flex items-start gap-4">
							<!-- Event Icon -->
							<div class="mt-1 flex-shrink-0">
								<div
									class="h-10 w-10 rounded-full {typeConfig.color} flex items-center justify-center"
								>
									<i class="{typeConfig.icon || 'fas fa-info-circle'} text-sm"></i>
								</div>
							</div>

							<!-- Event Content -->
							<div class="min-w-0 flex-1">
								<div class="mb-1 flex items-center justify-between">
									<div class="flex items-center gap-3">
										<span
											class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {typeConfig.color}"
										>
											{typeConfig.label}
										</span>
										{#if event.endpoint_name}
											<span class="text-sm font-medium text-blue-600 dark:text-blue-400">
												{event.endpoint_name}
											</span>
										{/if}
									</div>
									<div class="text-sm text-gray-500 dark:text-gray-400">
										{getTimeAgo(event.timestamp)}
									</div>
								</div>

								<div class="mb-2 text-sm text-gray-900 dark:text-white">
									{event.description || 'No description available'}
								</div>

								{#if event.metadata && Object.keys(event.metadata).length > 0}
									<details class="text-xs">
										<summary
											class="cursor-pointer text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
										>
											<i class="fas fa-info-circle mr-1"></i>
											Show details
										</summary>
										<div class="mt-2 rounded bg-gray-50 p-3 font-mono dark:bg-gray-700">
											<pre>{JSON.stringify(event.metadata, null, 2)}</pre>
										</div>
									</details>
								{/if}

								<div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
									{formatDate(event.timestamp)}
									{#if event.user_name}
										• by {event.user_name}
									{/if}
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div
					class="flex items-center justify-between border-t border-gray-200 px-6 py-4 dark:border-gray-700"
				>
					<div class="text-sm text-gray-700 dark:text-gray-300">
						Showing {(currentPage - 1) * pageSize + 1} to {Math.min(
							currentPage * pageSize,
							totalEvents
						)} of {totalEvents} events
					</div>
					<div class="flex gap-2">
						<button
							on:click={() => currentPage > 1 && (currentPage--, loadEvents())}
							disabled={currentPage === 1}
							class="rounded border border-gray-300 px-3 py-1 text-sm text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
						>
							<i class="fas fa-chevron-left mr-1"></i>
							Previous
						</button>
						<span class="px-3 py-1 text-sm text-gray-700 dark:text-gray-300">
							Page {currentPage} of {totalPages}
						</span>
						<button
							on:click={() => currentPage < totalPages && (currentPage++, loadEvents())}
							disabled={currentPage === totalPages}
							class="rounded border border-gray-300 px-3 py-1 text-sm text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
						>
							Next
							<i class="fas fa-chevron-right ml-1"></i>
						</button>
					</div>
				</div>
			{/if}
		{/if}
	</div>

	<!-- Event Type Legend -->
	<div class="mt-6 rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<h4 class="mb-3 text-sm font-medium text-gray-900 dark:text-white">Event Types</h4>
		<div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
			{#each eventTypes.slice(1) as type}
				<div class="flex items-center gap-2">
					<div class="h-6 w-6 rounded-full {type.color} flex items-center justify-center">
						<i class="{type.icon} text-xs"></i>
					</div>
					<span class="text-sm text-gray-700 dark:text-gray-300">{type.label}</span>
				</div>
			{/each}
		</div>
	</div>
</div>
