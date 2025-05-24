<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { alertsWS } from '$lib/websocket';
	import { getStatusBadgeClass, formatDate } from '$lib/utils';
	import type { Endpoint } from '$lib/types';
	import { Search, Filter, RefreshCw } from 'lucide-svelte';

	let endpoints: Endpoint[] = [];
	let filteredEndpoints: Endpoint[] = [];
	let searchQuery = '';
	let selectedStatus = '';
	let loading = true;

	onMount(async () => {
		await loadEndpoints();

		// Subscribe to real-time endpoint updates
		alertsWS.messages.subscribe(handleEndpointUpdate);
	});

	async function loadEndpoints() {
		try {
			loading = true;
			const response = await api.endpoints.getAll();
			endpoints = (response as any).data || [];
			filterEndpoints();
		} catch (error) {
			console.error('Failed to load endpoints:', error);
		} finally {
			loading = false;
		}
	}

	function filterEndpoints() {
		filteredEndpoints = endpoints.filter((endpoint) => {
			const matchesSearch =
				!searchQuery ||
				endpoint.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
				endpoint.hostname.toLowerCase().includes(searchQuery.toLowerCase()) ||
				(endpoint.ipAddress && endpoint.ipAddress.includes(searchQuery));

			const matchesStatus = !selectedStatus || endpoint.status === selectedStatus;

			return matchesSearch && matchesStatus;
		});
	}

	function handleEndpointUpdate(data: any) {
		if (data.endpoint) {
			const index = endpoints.findIndex((e) => e.id === data.endpoint.id);
			if (index >= 0) {
				endpoints[index] = data.endpoint;
			} else {
				endpoints = [...endpoints, data.endpoint];
			}
			filterEndpoints();
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'online':
				return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900';
			case 'offline':
				return 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900';
			default:
				return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900';
		}
	}

	// Reactive statements for filtering
	$: {
		filterEndpoints();
	}
</script>

<svelte:head>
	<title>Endpoints - GoSight</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Endpoints</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				Monitor and manage your infrastructure endpoints
			</p>
		</div>
		<button
			on:click={loadEndpoints}
			class="flex items-center space-x-2 rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
			disabled={loading}
		>
			<RefreshCw class="h-4 w-4 {loading ? 'animate-spin' : ''}" />
			<span>Refresh</span>
		</button>
	</div>

	<!-- Filters -->
	<div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900">
		<div
			class="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0"
		>
			<!-- Search -->
			<div class="relative max-w-md flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-400" />
				<input
					type="text"
					placeholder="Search endpoints..."
					bind:value={searchQuery}
					class="w-full rounded-lg border border-gray-300 bg-white py-2 pr-4 pl-10 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
				/>
			</div>

			<!-- Status Filter -->
			<div class="flex items-center space-x-2">
				<Filter class="h-4 w-4 text-gray-400" />
				<select
					bind:value={selectedStatus}
					class="rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
				>
					<option value="">All Status</option>
					<option value="online">Online</option>
					<option value="offline">Offline</option>
					<option value="unknown">Unknown</option>
				</select>
			</div>
		</div>
	</div>

	<!-- Stats -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
		<div
			class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Endpoints</div>
			<div class="text-2xl font-bold text-gray-900 dark:text-white">{endpoints.length}</div>
		</div>
		<div
			class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Online</div>
			<div class="text-2xl font-bold text-green-600 dark:text-green-400">
				{endpoints.filter((e) => e.status === 'online').length}
			</div>
		</div>
		<div
			class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Offline</div>
			<div class="text-2xl font-bold text-red-600 dark:text-red-400">
				{endpoints.filter((e) => e.status === 'offline').length}
			</div>
		</div>
	</div>

	<!-- Endpoints Table -->
	<div
		class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900"
	>
		{#if loading}
			<div class="p-8 text-center">
				<RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
				<p class="mt-2 text-gray-500 dark:text-gray-400">Loading endpoints...</p>
			</div>
		{:else if filteredEndpoints.length === 0}
			<div class="p-8 text-center">
				<p class="text-gray-500 dark:text-gray-400">
					{searchQuery || selectedStatus ? 'No endpoints match your filters' : 'No endpoints found'}
				</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
					<thead class="bg-gray-50 dark:bg-gray-800">
						<tr>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								Name
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								Status
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								IP Address
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								OS
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								Last Seen
							</th>
							<th
								class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
							>
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-900">
						{#each filteredEndpoints as endpoint}
							<tr class="hover:bg-gray-50 dark:hover:bg-gray-800">
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="flex items-center">
										<div>
											<div class="text-sm font-medium text-gray-900 dark:text-white">
												{endpoint.name}
											</div>
											<div class="text-sm text-gray-500 dark:text-gray-400">
												{endpoint.hostname}
											</div>
										</div>
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<span
										class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStatusColor(
											endpoint.status
										)}"
									>
										{endpoint.status}
									</span>
								</td>
								<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-900 dark:text-white">
									{endpoint.ipAddress}
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="text-sm text-gray-900 dark:text-white">
										{endpoint.os}
									</div>
									<div class="text-sm text-gray-500 dark:text-gray-400">
										{endpoint.architecture}
									</div>
								</td>
								<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-500 dark:text-gray-400">
									{formatDate(endpoint.lastSeen || '')}
								</td>
								<td class="px-6 py-4 text-sm font-medium whitespace-nowrap">
									<a
										href="/endpoints/{endpoint.id}"
										class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
									>
										View Details
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
