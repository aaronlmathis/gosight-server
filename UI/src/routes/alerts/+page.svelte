<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { alertsWS } from '$lib/websocket';
	import { activeAlerts, globalSearchQuery } from '$lib/stores';
	import { formatDate, getStatusBadgeClass } from '$lib/utils';
	import type { Alert } from '$lib/types';

	let alerts: Alert[] = [];
	let filteredAlerts: Alert[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let statusFilter = 'all';
	let severityFilter = 'all';
	let sortBy = 'created_at';
	let sortOrder = 'desc';

	// Pagination
	let currentPage = 1;
	let pageSize = 20;
	let totalAlerts = 0;

	onMount(async () => {
		await loadAlerts();
		
		// Subscribe to real-time alert updates
		websocketManager.connect();
		websocketManager.subscribeToAlerts((alert: Alert) => {
			alerts = [alert, ...alerts];
			filterAndSortAlerts();
		});

		// Subscribe to search store
		searchStore.subscribe(term => {
			searchTerm = term;
			filterAndSortAlerts();
		});
	});

	async function loadAlerts() {
		try {
			loading = true;
			const response = await api.getAlerts({
				page: currentPage,
				limit: pageSize,
				sort: `${sortBy}:${sortOrder}`
			});
			alerts = response.alerts || [];
			totalAlerts = response.total || 0;
			filterAndSortAlerts();
		} catch (err) {
			error = 'Failed to load alerts: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	function filterAndSortAlerts() {
		let filtered = [...alerts];

		// Search filter
		if (searchTerm) {
			filtered = filtered.filter(alert =>
				alert.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				alert.description?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				alert.endpoint_name?.toLowerCase().includes(searchTerm.toLowerCase())
			);
		}

		// Status filter
		if (statusFilter !== 'all') {
			filtered = filtered.filter(alert => alert.status === statusFilter);
		}

		// Severity filter
		if (severityFilter !== 'all') {
			filtered = filtered.filter(alert => alert.severity === severityFilter);
		}

		// Sort
		filtered.sort((a, b) => {
			let aValue = a[sortBy as keyof Alert];
			let bValue = b[sortBy as keyof Alert];
			
			if (typeof aValue === 'string') aValue = aValue.toLowerCase();
			if (typeof bValue === 'string') bValue = bValue.toLowerCase();
			
			if (sortOrder === 'asc') {
				return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
			} else {
				return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
			}
		});

		filteredAlerts = filtered;
	}

	function handleSort(field: string) {
		if (sortBy === field) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortBy = field;
			sortOrder = 'desc';
		}
		filterAndSortAlerts();
	}

	async function acknowledgeAlert(alertId: string) {
		try {
			await api.acknowledgeAlert(alertId);
			alerts = alerts.map(alert =>
				alert.id === alertId ? { ...alert, status: 'acknowledged' } : alert
			);
			filterAndSortAlerts();
		} catch (err) {
			console.error('Failed to acknowledge alert:', err);
		}
	}

	async function resolveAlert(alertId: string) {
		try {
			await api.resolveAlert(alertId);
			alerts = alerts.map(alert =>
				alert.id === alertId ? { ...alert, status: 'resolved' } : alert
			);
			filterAndSortAlerts();
		} catch (err) {
			console.error('Failed to resolve alert:', err);
		}
	}

	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'critical': return 'text-red-600 bg-red-100';
			case 'high': return 'text-orange-600 bg-orange-100';
			case 'medium': return 'text-yellow-600 bg-yellow-100';
			case 'low': return 'text-blue-600 bg-blue-100';
			default: return 'text-gray-600 bg-gray-100';
		}
	}

	$: totalPages = Math.ceil(totalAlerts / pageSize);
</script>

<svelte:head>
	<title>Alerts - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Alerts</h1>
		<p class="text-gray-600 dark:text-gray-400">Monitor and manage system alerts</p>
	</div>

	<!-- Filters and Controls -->
	<div class="mb-6 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
		<div class="flex flex-wrap gap-4 items-center justify-between">
			<div class="flex flex-wrap gap-4 items-center">
				<!-- Search -->
				<div class="relative">
					<i class="fas fa-search absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400"></i>
					<input
						type="text"
						placeholder="Search alerts..."
						bind:value={searchTerm}
						on:input={filterAndSortAlerts}
						class="pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
					/>
				</div>

				<!-- Status Filter -->
				<select
					bind:value={statusFilter}
					on:change={filterAndSortAlerts}
					class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					<option value="all">All Status</option>
					<option value="active">Active</option>
					<option value="acknowledged">Acknowledged</option>
					<option value="resolved">Resolved</option>
				</select>

				<!-- Severity Filter -->
				<select
					bind:value={severityFilter}
					on:change={filterAndSortAlerts}
					class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
				>
					<option value="all">All Severity</option>
					<option value="critical">Critical</option>
					<option value="high">High</option>
					<option value="medium">Medium</option>
					<option value="low">Low</option>
				</select>
			</div>

			<div class="flex gap-2">
				<a
					href="/alerts/rules"
					class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
				>
					<i class="fas fa-cog mr-2"></i>
					Alert Rules
				</a>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="flex justify-center items-center py-12">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
			<div class="flex">
				<i class="fas fa-exclamation-triangle text-red-500 mr-3 mt-0.5"></i>
				<div>
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<p class="text-sm text-red-600 dark:text-red-300 mt-1">{error}</p>
				</div>
			</div>
		</div>
	{:else}
		<!-- Alerts Table -->
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
					<thead class="bg-gray-50 dark:bg-gray-700">
						<tr>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600"
								on:click={() => handleSort('severity')}
							>
								Severity
								{#if sortBy === 'severity'}
									<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
								{/if}
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600"
								on:click={() => handleSort('name')}
							>
								Alert
								{#if sortBy === 'name'}
									<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
								{/if}
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600"
								on:click={() => handleSort('endpoint_name')}
							>
								Endpoint
								{#if sortBy === 'endpoint_name'}
									<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
								{/if}
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600"
								on:click={() => handleSort('status')}
							>
								Status
								{#if sortBy === 'status'}
									<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
								{/if}
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600"
								on:click={() => handleSort('created_at')}
							>
								Created
								{#if sortBy === 'created_at'}
									<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
								{/if}
							</th>
							<th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
						{#each filteredAlerts as alert (alert.id)}
							<tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
								<td class="px-6 py-4 whitespace-nowrap">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getSeverityColor(alert.severity)}">
										{alert.severity}
									</span>
								</td>
								<td class="px-6 py-4">
									<div class="text-sm font-medium text-gray-900 dark:text-white">
										{alert.name}
									</div>
									<div class="text-sm text-gray-500 dark:text-gray-400">
										{alert.description}
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
									{alert.endpoint_name || '-'}
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(alert.status)}">
										{alert.status}
									</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
									{formatDate(alert.created_at)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
									<div class="flex justify-end gap-2">
										{#if alert.status === 'active'}
											<button
												on:click={() => acknowledgeAlert(alert.id)}
												class="text-yellow-600 hover:text-yellow-900 dark:text-yellow-400 dark:hover:text-yellow-300"
												title="Acknowledge"
											>
												<i class="fas fa-check"></i>
											</button>
										{/if}
										{#if alert.status !== 'resolved'}
											<button
												on:click={() => resolveAlert(alert.id)}
												class="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300"
												title="Resolve"
											>
												<i class="fas fa-check-double"></i>
											</button>
										{/if}
										<a
											href="/alerts/{alert.id}"
											class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
											title="View Details"
										>
											<i class="fas fa-eye"></i>
										</a>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="bg-white dark:bg-gray-800 px-4 py-3 flex items-center justify-between border-t border-gray-200 dark:border-gray-700 sm:px-6">
					<div class="flex-1 flex justify-between sm:hidden">
						<button
							on:click={() => currentPage > 1 && (currentPage--, loadAlerts())}
							disabled={currentPage === 1}
							class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Previous
						</button>
						<button
							on:click={() => currentPage < totalPages && (currentPage++, loadAlerts())}
							disabled={currentPage === totalPages}
							class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							Next
						</button>
					</div>
					<div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
						<div>
							<p class="text-sm text-gray-700 dark:text-gray-300">
								Showing
								<span class="font-medium">{(currentPage - 1) * pageSize + 1}</span>
								to
								<span class="font-medium">{Math.min(currentPage * pageSize, totalAlerts)}</span>
								of
								<span class="font-medium">{totalAlerts}</span>
								results
							</p>
						</div>
						<div>
							<nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
								<button
									on:click={() => currentPage > 1 && (currentPage--, loadAlerts())}
									disabled={currentPage === 1}
									class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
								>
									<i class="fas fa-chevron-left"></i>
								</button>
								{#each Array(totalPages) as _, i}
									{#if i + 1 === currentPage || i + 1 === 1 || i + 1 === totalPages || Math.abs(i + 1 - currentPage) <= 2}
										<button
											on:click={() => (currentPage = i + 1, loadAlerts())}
											class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 {currentPage === i + 1 ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-500 text-blue-600 dark:text-blue-400' : ''}"
										>
											{i + 1}
										</button>
									{:else if i + 1 === currentPage - 3 || i + 1 === currentPage + 3}
										<span class="relative inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-700 dark:text-gray-300">
											...
										</span>
									{/if}
								{/each}
								<button
									on:click={() => currentPage < totalPages && (currentPage++, loadAlerts())}
									disabled={currentPage === totalPages}
									class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
								>
									<i class="fas fa-chevron-right"></i>
								</button>
							</nav>
						</div>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
