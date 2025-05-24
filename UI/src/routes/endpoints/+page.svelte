<script lang="ts">
	import { onMount } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import { api } from '$lib/api';
	import { alertsWS } from '$lib/websocket';
	import { getStatusBadgeClass, formatDate } from '$lib/utils';
	import type { Endpoint, EndpointApiResponse } from '$lib/types';
	import { Search, Filter, RefreshCw, ChevronDown, ChevronRight } from 'lucide-svelte';

	let endpoints: Endpoint[] = [];
	let containers: any[] = [];
	let filteredEndpoints: Endpoint[] = [];
	let searchQuery = '';
	let selectedStatus = '';
	let loading = true;
	let expandedRows = new Set<string>();
	let containerData = new Map<string, any[]>();

	onMount(async () => {
		await loadEndpoints();

		// Subscribe to real-time endpoint updates
		alertsWS.messages.subscribe(handleEndpointUpdate);
	});

	async function loadEndpoints() {
		try {
			loading = true;
			const [hostsResponse, containersResponse] = await Promise.all([
				api.endpoints.getAll(),
				api.endpoints.getByType('containers')
			]);
			
			// API returns array directly, not wrapped in data object
			const rawEndpoints: any[] = hostsResponse || [];
			const rawContainers: any[] = containersResponse || [];
			
			endpoints = rawEndpoints.map(
				(endpoint): Endpoint => ({
					id: endpoint.id,
					name: endpoint.hostname || endpoint.id,
					hostname: endpoint.hostname || '',
					ipAddress: endpoint.ip || '',
					status: (endpoint.status?.toLowerCase() as 'online' | 'offline' | 'unknown') || 'unknown',
					lastSeen: endpoint.last_seen || '',
					os: endpoint.os || '',
					architecture: endpoint.arch || '',
					port: 0, // Default port, not provided by API
					tags: endpoint.labels ? Object.entries(endpoint.labels).map(([k, v]) => `${k}:${v}`) : [],
					uptime: endpoint.uptime || 0,
					agentVersion: endpoint.version || '',
					// Add additional fields from API
					agentId: endpoint.agent_id || '',
					hostId: endpoint.host_id || '',
					uptimeSeconds: endpoint.uptime_seconds || 0
				})
			);
			
			containers = rawContainers;
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

	function toggleRowExpansion(endpointId: string) {
		if (expandedRows.has(endpointId)) {
			expandedRows.delete(endpointId);
		} else {
			expandedRows.add(endpointId);
			loadContainersForEndpoint(endpointId);
		}
		expandedRows = expandedRows; // Trigger reactivity
	}

	async function loadContainersForEndpoint(endpointId: string) {
		if (containerData.has(endpointId)) return; // Already loaded

		try {
			// Find the endpoint to get hostname
			const endpoint = endpoints.find(e => e.id === endpointId);
			if (!endpoint) return;

			// Filter containers for this host
			const hostContainers = containers.filter(c => c.host_id === endpoint.hostId);
			containerData.set(endpointId, hostContainers);
			containerData = containerData; // Trigger reactivity
		} catch (error) {
			console.error('Failed to load containers for endpoint:', endpointId, error);
		}
	}

	function formatUptime(seconds: number): string {
		if (!seconds) return '—';
		const s = Math.floor(seconds);
		const d = Math.floor(s / 86400);
		const h = Math.floor((s % 86400) / 3600);
		const m = Math.floor((s % 3600) / 60);
		return `${d > 0 ? d + 'd ' : ''}${h}h ${m}m`;
	}

	function formatLastSeen(isoTime: string): string {
		if (!isoTime) return '—';
		const last = new Date(isoTime).getTime();
		const now = Date.now();
		const diff = Math.floor((now - last) / 1000);

		if (diff < 60) return `${diff}s ago`;
		if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
		if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
		return `${Math.floor(diff / 86400)}d ago`;
	}

	function getContainerStatusColor(status: string): string {
		switch (status?.toLowerCase()) {
			case 'running':
				return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900';
			case 'exited':
			case 'stopped':
				return 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900';
			default:
				return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900';
		}
	}

	function handleEndpointUpdate(data: any) {
		if (data.endpoint) {
			// Map the websocket data to the expected format
			const rawEndpoint: EndpointApiResponse = data.endpoint;
			const mappedEndpoint: Endpoint = {
				id: rawEndpoint.id,
				name: rawEndpoint.hostname || rawEndpoint.id,
				hostname: rawEndpoint.hostname || '',
				ipAddress: rawEndpoint.ip || '',
				status:
					(rawEndpoint.status?.toLowerCase() as 'online' | 'offline' | 'unknown') || 'unknown',
				lastSeen: rawEndpoint.last_seen || '',
				os: rawEndpoint.os || '',
				architecture: rawEndpoint.arch || '',
				port: 0,
				tags: rawEndpoint.labels
					? Object.entries(rawEndpoint.labels).map(([k, v]) => `${k}:${v}`)
					: [],
				uptime: rawEndpoint.uptime || 0,
				agentVersion: rawEndpoint.version || ''
			};

			const index = endpoints.findIndex((e) => e.id === mappedEndpoint.id);
			if (index >= 0) {
				endpoints[index] = mappedEndpoint;
			} else {
				endpoints = [...endpoints, mappedEndpoint];
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

<PermissionGuard requiredPermission="gosight:dashboard:view">
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
		<div
			class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
		>
			<div
				class="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0"
			>
				<!-- Search -->
				<div class="relative max-w-md flex-1">
					<Search
						class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-400"
					/>
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
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-4">
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Endpoints</div>
				<div class="text-2xl font-bold text-gray-900 dark:text-white">{endpoints.length}</div>
			</div>
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Hosts Online</div>
				<div class="text-2xl font-bold text-green-600 dark:text-green-400">
					{endpoints.filter((e) => e.status === 'online').length} / {endpoints.length}
				</div>
			</div>
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Containers Running</div>
				<div class="text-2xl font-bold text-blue-600 dark:text-blue-400">
					{containers.filter((c) => c.status?.toLowerCase() === 'running').length} / {containers.length}
				</div>
			</div>
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Runtimes</div>
				<div class="text-xl font-bold text-gray-900 dark:text-white">
					{[...new Set(containers.map(c => c.runtime).filter(Boolean))].join(', ') || '—'}
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
						{searchQuery || selectedStatus
							? 'No endpoints match your filters'
							: 'No endpoints found'}
					</p>
				</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
						<thead class="bg-gray-50 dark:bg-gray-800">
							<tr>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Status
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Hostname
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									IP Address
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									OS
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Platform
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Architecture
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Agent ID
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Version
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Last Seen
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Uptime
								</th>
								<th
									class="px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-400"
								>
									Containers
								</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-900">
							{#each filteredEndpoints as endpoint}
								<!-- Host Row -->
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-800">
									<td class="px-3 py-2">
										{#if endpoint.status === 'online'}
											<span class="text-green-500 font-medium">● Online</span>
										{:else}
											<span class="text-red-500 font-medium">● Offline</span>
										{/if}
									</td>
									<td class="px-3 py-2 font-medium">
										{#if endpoint.status === 'online'}
											<a
												href="/endpoints/{endpoint.id}"
												class="text-blue-800 dark:text-blue-400 hover:underline"
											>
												{endpoint.hostname}
											</a>
										{:else}
											<span class="text-gray-500 dark:text-gray-400">{endpoint.hostname}</span>
										{/if}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{endpoint.ipAddress}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{endpoint.os}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{endpoint.tags.find(t => t.startsWith('platform:'))?.split(':')[1] || '—'}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{endpoint.architecture}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white font-mono">
										{endpoint.agentId}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{endpoint.agentVersion}
									</td>
									<td class="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
										{#if endpoint.status === 'online'}
											—
										{:else}
											{formatLastSeen(endpoint.lastSeen)}
										{/if}
									</td>
									<td class="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
										{#if endpoint.status === 'online'}
											{formatUptime(endpoint.uptimeSeconds)}
										{:else}
											—
										{/if}
									</td>
									<td class="px-3 py-2">
										{#if endpoint.status === 'online'}
											<button
												on:click={() => toggleRowExpansion(endpoint.id)}
												class="text-blue-500 hover:text-blue-700 flex items-center"
											>
												{#if expandedRows.has(endpoint.id)}
													<ChevronDown class="w-4 h-4" />
												{:else}
													<ChevronRight class="w-4 h-4" />
												{/if}
											</button>
										{:else}
											<span class="text-gray-400">—</span>
										{/if}
									</td>
								</tr>

								<!-- Container Row (if expanded) -->
								{#if expandedRows.has(endpoint.id)}
									<tr class="container-subtable">
										<td colspan="11" class="p-0">
											<div class="p-4 bg-gray-50 dark:bg-gray-800">
												{#if containerData.has(endpoint.id)}
													{@const hostContainers = containerData.get(endpoint.id) || []}
													{#if hostContainers.length > 0}
														<div class="overflow-x-auto">
															<table class="min-w-full text-sm">
																<thead>
																	<tr class="border-b border-gray-200 dark:border-gray-600">
																		<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Container</th>
																		<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Image</th>
																		<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Status</th>
																		<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Runtime</th>
																		<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Last Seen</th>
																	</tr>
																</thead>
																<tbody>
																	{#each hostContainers as container}
																		<tr class="border-b border-gray-100 dark:border-gray-700">
																			<td class="px-3 py-2 font-medium text-gray-900 dark:text-white">
																				{container.name || container.Name || '—'}
																			</td>
																			<td class="px-3 py-2 text-gray-600 dark:text-gray-300">
																				{container.image || container.ImageName || '—'}
																			</td>
																			<td class="px-3 py-2">
																				<span
																					class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {getContainerStatusColor(
																						container.status || container.Status
																					)}"
																				>
																					{container.status || container.Status || 'unknown'}
																				</span>
																			</td>
																			<td class="px-3 py-2 text-gray-600 dark:text-gray-300">
																				{container.runtime || container.Runtime || '—'}
																			</td>
																			<td class="px-3 py-2 text-gray-500 dark:text-gray-400">
																				{formatLastSeen(container.last_seen || container.LastSeen || '')}
																			</td>
																		</tr>
																	{/each}
																</tbody>
															</table>
														</div>
													{:else}
														<div class="text-sm text-gray-500 dark:text-gray-400">No containers found for this host</div>
													{/if}
												{:else}
													<div class="text-sm text-gray-400">Loading containers...</div>
												{/if}
											</div>
										</td>
									</tr>
								{/if}
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	</div>
</PermissionGuard>
