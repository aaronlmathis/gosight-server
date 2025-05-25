<script lang="ts">
	import { onMount } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import { api } from '$lib/api';
	import { alertsWS } from '$lib/websocket';
	import {
		Search,
		Filter,
		RefreshCw,
		ChevronDown,
		ChevronRight,
		Server,
		Package,
		Cog
	} from 'lucide-svelte';

	interface Host {
		id: string;
		endpoint_id: string;
		agent_id: string;
		hostname: string;
		ip: string;
		os: string;
		arch: string;
		version: string;
		status: string;
		last_seen: string;
		uptime_seconds: number;
		labels: Record<string, any>;
	}

	interface Container {
		id: string;
		name: string;
		image: string;
		status: string;
		heartbeat: string;
		cpu: string;
		mem: string;
		rx: string;
		tx: string;
		uptime: string;
		runtime: string;
	}

	let hosts: Host[] = [];
	let allContainers: any[] = [];
	let filteredHosts: Host[] = [];
	let searchQuery = '';
	let selectedStatus = '';
	let loading = true;
	let expandedRows = new Set<string>();
	let containerData = new Map<string, Container[]>();

	// Search dropdown state
	let showSearchDropdown = false;
	let searchSuggestions: Host[] = [];
	let searchInputElement: HTMLInputElement;

	// Summary stats
	let totalHosts = 0;
	let onlineHosts = 0;
	let runningContainers = 0;
	let totalContainers = 0;
	let observedRuntimes: string[] = [];

	onMount(async () => {
		await loadHosts();
		await loadGlobalContainerMetrics();

		// Subscribe to real-time endpoint updates
		alertsWS.messages.subscribe(handleEndpointUpdate);
	});

	async function loadHosts() {
		try {
			loading = true;
			const response = await api.endpoints.getByType('hosts');
			const rawHosts: any[] = response || [];

			hosts = rawHosts
				.filter((host) => !host.endpoint_id?.startsWith('ctr-')) // Filter out containers
				.map(
					(host): Host => ({
						id: host.agent_id || host.id,
						endpoint_id: host.endpoint_id || '',
						agent_id: host.agent_id || '',
						hostname: host.hostname || '',
						ip: host.ip || '',
						os: host.os || '',
						arch: host.arch || '',
						version: host.version || '',
						status: host.status || 'unknown',
						last_seen: host.last_seen || '',
						uptime_seconds: host.uptime_seconds || 0,
						labels: host.labels || {}
					})
				);

			updateSummaryStats();
			filterHosts();
		} catch (error) {
			console.error('Failed to load hosts:', error);
		} finally {
			loading = false;
		}
	}

	async function loadGlobalContainerMetrics() {
		try {
			const metricNames = ['cpu_percent', 'uptime_seconds'];
			const query = metricNames.map((m) => `metric=${encodeURIComponent(m)}`).join('&');

			const response = await api.request(`/query?namespace=container&${query}`);
			const rows = Array.isArray(response) ? response : [];

			const containerMap = new Map();
			const runtimeSet = new Set<string>();

			for (const row of rows) {
				const tags = row.tags || {};
				const id = tags.container_id;
				const metricName = tags['__name__'];
				const runtime = tags.runtime;

				if (!id || !metricName) continue;

				if (runtime) runtimeSet.add(runtime);

				if (!containerMap.has(id)) {
					containerMap.set(id, {
						id,
						status: 'unknown',
						runtime,
						metrics: {}
					});
				}

				const container = containerMap.get(id);
				container.metrics[metricName] = row.value;

				if (metricName === 'status') {
					container.status = row.value === 1 ? 'running' : 'exited';
				}
			}

			allContainers = Array.from(containerMap.values());
			observedRuntimes = Array.from(runtimeSet);
			updateSummaryStats();
		} catch (error) {
			console.error('Failed to load global container metrics:', error);
		}
	}

	function updateSummaryStats() {
		totalHosts = hosts.length;
		onlineHosts = hosts.filter((h) => h.status.toLowerCase() === 'online').length;
		totalContainers = allContainers.length;
		runningContainers = allContainers.filter((c) => c.status === 'running').length;
	}

	function filterHosts() {
		const query = searchQuery.toLowerCase();
		filteredHosts = hosts.filter((host) => {
			const matchesSearch =
				!query ||
				host.hostname.toLowerCase().includes(query) ||
				host.endpoint_id.toLowerCase().includes(query) ||
				host.ip.includes(query) ||
				host.os.toLowerCase().includes(query) ||
				host.arch.toLowerCase().includes(query) ||
				host.agent_id.toLowerCase().includes(query) ||
				getPlatform(host.labels).toLowerCase().includes(query);

			const matchesStatus =
				!selectedStatus || host.status.toLowerCase() === selectedStatus.toLowerCase();

			return matchesSearch && matchesStatus;
		});
	}

	function updateSearchSuggestions() {
		if (!searchQuery.trim()) {
			searchSuggestions = [];
			showSearchDropdown = false;
			return;
		}

		const query = searchQuery.toLowerCase();
		searchSuggestions = hosts
			.filter((host) => {
				return (
					host.hostname.toLowerCase().includes(query) ||
					host.endpoint_id.toLowerCase().includes(query) ||
					host.ip.includes(query) ||
					host.os.toLowerCase().includes(query) ||
					host.arch.toLowerCase().includes(query) ||
					host.agent_id.toLowerCase().includes(query) ||
					getPlatform(host.labels).toLowerCase().includes(query)
				);
			})
			.slice(0, 8); // Limit to 8 suggestions

		showSearchDropdown = searchSuggestions.length > 0;
	}

	function selectSearchSuggestion(host: Host) {
		searchQuery = host.hostname;
		showSearchDropdown = false;
		filterHosts();
	}

	function handleSearchInput() {
		updateSearchSuggestions();
		filterHosts();
	}

	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			showSearchDropdown = false;
		}
	}

	function handleSearchBlur() {
		// Delay hiding dropdown to allow for clicks
		setTimeout(() => {
			showSearchDropdown = false;
		}, 200);
	}

	async function toggleRowExpansion(agentId: string) {
		if (expandedRows.has(agentId)) {
			expandedRows.delete(agentId);
		} else {
			expandedRows.add(agentId);
			await loadContainersForHost(agentId);
		}
		expandedRows = expandedRows; // Trigger reactivity
	}

	async function loadContainersForHost(agentId: string) {
		if (containerData.has(agentId)) return; // Already loaded

		try {
			const host = hosts.find((h) => h.agent_id === agentId);
			if (!host) return;

			const hostname = host.hostname;
			const runtimes = ['podman', 'docker'];
			const metricNames = [
				'cpu_percent',
				'mem_usage_bytes',
				'net_rx_bytes',
				'net_tx_bytes',
				'uptime_seconds'
			];

			const metrics = runtimes.flatMap((rt) => metricNames.map((name) => `${name}`));
			const query =
				metrics.map((m) => `metric=${m}`).join('&') + `&hostname=${encodeURIComponent(hostname)}`;

			const response = await api.request(`/query?${query}`);
			const rows = Array.isArray(response) ? response : [];

			const containers = groupContainers(rows);
			containerData.set(agentId, containers);
			containerData = containerData; // Trigger reactivity
		} catch (error) {
			console.error('Failed to load containers for host:', agentId, error);
		}
	}

	function groupContainers(rows: any[]): Container[] {
		if (!Array.isArray(rows)) return [];
		const map: Record<string, any> = {};

		rows.forEach((row) => {
			const tags = row.tags || {};
			const id = tags.container_id;
			const metricName = tags['__name__'] || '';
			const parts = metricName.split('.');
			const runtime = parts.length >= 2 ? parts[1] : 'unknown';

			if (!id) {
				console.warn('Skipping row: missing container_id', row);
				return;
			}

			if (!map[id]) {
				map[id] = {
					id,
					name: tags.container_name || '—',
					image: tags.image || '—',
					status: tags.status || 'unknown',
					heartbeat: 'Online',
					cpu: '—',
					mem: '—',
					rx: '—',
					tx: '—',
					uptime: '—',
					runtime
				};
			}

			const value = row.value;
			const shortName = metricName.split('.').pop();

			switch (shortName) {
				case 'cpu_percent':
					map[id].cpu = `${value.toFixed(1)}%`;
					break;
				case 'mem_usage_bytes':
					map[id].mem = formatBytes(value);
					break;
				case 'net_rx_bytes':
					map[id].rx = formatBytes(value);
					break;
				case 'net_tx_bytes':
					map[id].tx = formatBytes(value);
					break;
				case 'uptime_seconds':
					map[id].uptime = formatUptime(value);
					break;
			}
		});

		return Object.values(map);
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

	function formatBytes(bytes: number): string {
		if (!bytes || isNaN(bytes)) return '—';
		const units = ['B', 'KB', 'MB', 'GB'];
		let i = 0;
		let value = bytes;
		while (value >= 1024 && i < units.length - 1) {
			value /= 1024;
			i++;
		}
		return `${value.toFixed(1)} ${units[i]}`;
	}

	function getStatusColor(status: string): string {
		const normalizedStatus = status.toLowerCase();
		switch (normalizedStatus) {
			case 'online':
				return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900';
			case 'offline':
				return 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900';
			default:
				return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900';
		}
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
		if (data.endpoint && data.endpoint.endpoint_id?.startsWith('host-')) {
			const rawHost = data.endpoint;
			const mappedHost: Host = {
				id: rawHost.agent_id || rawHost.id,
				endpoint_id: rawHost.endpoint_id || '',
				agent_id: rawHost.agent_id || '',
				hostname: rawHost.hostname || '',
				ip: rawHost.ip || '',
				os: rawHost.os || '',
				arch: rawHost.arch || '',
				version: rawHost.version || '',
				status: rawHost.status || 'unknown',
				last_seen: rawHost.last_seen || '',
				uptime_seconds: rawHost.uptime_seconds || 0,
				labels: rawHost.labels || {}
			};

			const index = hosts.findIndex((h) => h.agent_id === mappedHost.agent_id);
			if (index >= 0) {
				hosts[index] = mappedHost;
			} else {
				hosts = [...hosts, mappedHost];
			}
			updateSummaryStats();
			filterHosts();
		}
	}

	function getPlatform(labels: Record<string, any>): string {
		const platform = labels.platform || '';
		const platformVersion = labels.platform_version || '';
		return `${platform} ${platformVersion}`.trim() || '—';
	}

	// Reactive statements for filtering
	$: if (searchQuery !== undefined) {
		updateSearchSuggestions();
		filterHosts();
	}
	$: if (selectedStatus !== undefined) {
		filterHosts();
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
					Hosts and their active containers across your environment
				</p>
			</div>
			<button
				on:click={loadHosts}
				class="flex items-center space-x-2 rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
				disabled={loading}
			>
				<RefreshCw class="h-4 w-4 {loading ? 'animate-spin' : ''}" />
				<span>Refresh</span>
			</button>
		</div>

		<!-- Summary Stats -->
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="flex items-center space-x-4">
					<div class="text-blue-500 dark:text-blue-400">
						<Server class="h-6 w-6" />
					</div>
					<div>
						<div class="text-sm font-medium text-gray-500 dark:text-gray-400">Hosts Online</div>
						<div class="text-2xl font-bold text-gray-900 dark:text-white">
							{onlineHosts} / {totalHosts}
						</div>
					</div>
				</div>
			</div>
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="flex items-center space-x-4">
					<div class="text-green-500 dark:text-green-400">
						<Package class="h-6 w-6" />
					</div>
					<div>
						<div class="text-sm font-medium text-gray-500 dark:text-gray-400">
							Containers Running
						</div>
						<div class="text-2xl font-bold text-gray-900 dark:text-white">
							{runningContainers} / {totalContainers}
						</div>
					</div>
				</div>
			</div>
			<div
				class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
			>
				<div class="flex items-center space-x-4">
					<div class="text-purple-500 dark:text-purple-400">
						<Cog class="h-6 w-6" />
					</div>
					<div>
						<div class="text-sm font-medium text-gray-500 dark:text-gray-400">
							Runtimes Observed
						</div>
						<div class="text-xl font-bold text-gray-900 dark:text-white">
							{observedRuntimes.join(', ') || '—'}
						</div>
					</div>
				</div>
			</div>
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
						bind:this={searchInputElement}
						type="text"
						placeholder="Search by hostname, IP, OS, platform, agent ID..."
						bind:value={searchQuery}
						on:input={handleSearchInput}
						on:keydown={handleSearchKeydown}
						on:blur={handleSearchBlur}
						on:focus={handleSearchInput}
						class="w-full rounded-lg border border-gray-300 bg-white py-2 pr-4 pl-10 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
					/>

					<!-- Search Dropdown -->
					{#if showSearchDropdown && searchSuggestions.length > 0}
						<div
							class="absolute top-full right-0 left-0 z-50 mt-1 max-h-64 overflow-y-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-600 dark:bg-gray-800"
						>
							{#each searchSuggestions as suggestion}
								<button
									type="button"
									on:click={() => selectSearchSuggestion(suggestion)}
									class="flex w-full items-center justify-between px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700"
								>
									<div class="flex-1">
										<div class="font-medium text-gray-900 dark:text-white">
											{suggestion.hostname}
										</div>
										<div class="text-sm text-gray-500 dark:text-gray-400">
											{suggestion.endpoint_id}
										</div>
									</div>
									<div class="text-right">
										<div class="text-sm font-medium text-gray-700 dark:text-gray-300">
											{suggestion.ip}
										</div>
										<div class="text-xs text-gray-500 dark:text-gray-400">
											{suggestion.os}
										</div>
									</div>
								</button>
							{/each}
						</div>
					{/if}
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

		<!-- Endpoints Table -->
		<div
			class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-900"
		>
			{#if loading}
				<div class="p-8 text-center">
					<RefreshCw class="mx-auto h-8 w-8 animate-spin text-gray-400" />
					<p class="mt-2 text-gray-500 dark:text-gray-400">Loading hosts...</p>
				</div>
			{:else if filteredHosts.length === 0}
				<div class="p-8 text-center">
					<p class="text-gray-500 dark:text-gray-400">
						{searchQuery || selectedStatus ? 'No hosts match your filters' : 'No hosts found'}
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
							{#each filteredHosts as host}
								<!-- Host Row -->
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-800">
									<td class="px-3 py-2">
										{#if host.status.toLowerCase() === 'online'}
											<span class="font-medium text-green-500">● Online</span>
										{:else}
											<span class="font-medium text-red-500">● Offline</span>
										{/if}
									</td>
									<td class="px-3 py-2 font-medium">
										{#if host.status.toLowerCase() === 'online'}
											<a
												href="/endpoints/{host.endpoint_id}"
												class="text-blue-800 hover:underline dark:text-blue-400"
											>
												{host.hostname}
											</a>
										{:else}
											<span class="text-gray-500 dark:text-gray-400">{host.hostname}</span>
										{/if}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{host.ip}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{host.os}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{getPlatform(host.labels)}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{host.arch}
									</td>
									<td class="px-3 py-2 font-mono text-sm text-gray-900 dark:text-white">
										{host.agent_id}
									</td>
									<td class="px-3 py-2 text-sm text-gray-900 dark:text-white">
										{host.version}
									</td>
									<td class="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
										{#if host.status.toLowerCase() === 'online'}
											—
										{:else}
											{formatLastSeen(host.last_seen)}
										{/if}
									</td>
									<td class="px-3 py-2 text-sm text-gray-500 dark:text-gray-400">
										{#if host.status.toLowerCase() === 'online'}
											{formatUptime(host.uptime_seconds)}
										{:else}
											—
										{/if}
									</td>
									<td class="px-3 py-2">
										{#if host.status.toLowerCase() === 'online'}
											<button
												on:click={() => toggleRowExpansion(host.agent_id)}
												class="flex items-center text-blue-500 transition-colors hover:text-blue-700"
											>
												{#if expandedRows.has(host.agent_id)}
													<ChevronDown class="h-4 w-4" />
												{:else}
													<ChevronRight class="h-4 w-4" />
												{/if}
											</button>
										{:else}
											<span class="text-gray-400">—</span>
										{/if}
									</td>
								</tr>

								<!-- Container Row (if expanded) -->
								{#if expandedRows.has(host.agent_id)}
									<tr class="container-subtable">
										<td colspan="11" class="p-0">
											<div class="bg-gray-50 p-4 dark:bg-gray-800">
												{#if containerData.has(host.agent_id)}
													{@const hostContainers = containerData.get(host.agent_id) || []}
													{#if hostContainers.length > 0}
														<div class="overflow-x-auto">
															<table class="min-w-full text-sm">
																<thead>
																	<tr class="border-b border-gray-200 dark:border-gray-600">
																		<th
																			class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Status</th
																		>
																		<th
																			class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Heartbeat</th
																		>
																		<th
																			class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Runtime</th
																		>
																		<th
																			class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Name</th
																		>
																		<th
																			class="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Image</th
																		>
																		<th
																			class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>CPU %</th
																		>
																		<th
																			class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Memory</th
																		>
																		<th
																			class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>RX</th
																		>
																		<th
																			class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>TX</th
																		>
																		<th
																			class="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase dark:text-gray-400"
																			>Uptime</th
																		>
																	</tr>
																</thead>
																<tbody>
																	{#each hostContainers as container, i}
																		<tr
																			class="{i % 2 === 0
																				? 'bg-white dark:bg-gray-800'
																				: 'bg-gray-50 dark:bg-gray-700'} transition-colors hover:bg-gray-50 dark:hover:bg-gray-600"
																		>
																			<td class="px-3 py-2">
																				<span
																					class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {getContainerStatusColor(
																						container.status
																					)}"
																				>
																					{container.status}
																				</span>
																			</td>
																			<td class="px-3 py-2">
																				<span
																					class="inline-flex items-center rounded-sm px-2 py-0.5 text-xs font-bold {container.heartbeat ===
																					'Online'
																						? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100'
																						: 'bg-red-100 text-red-800 dark:bg-red-800 dark:text-red-100'}"
																				>
																					{container.heartbeat}
																				</span>
																			</td>
																			<td class="px-3 py-2">
																				<span
																					class="inline-block rounded-sm bg-gray-200 px-2 py-0.5 text-xs text-gray-800 dark:bg-gray-700 dark:text-gray-300"
																				>
																					{container.runtime}
																				</span>
																			</td>
																			<td
																				class="px-3 py-2 font-medium text-gray-900 dark:text-white"
																			>
																				{container.name}
																			</td>
																			<td class="px-3 py-2 text-gray-700 dark:text-gray-300">
																				{container.image}
																			</td>
																			<td
																				class="px-3 py-2 text-right text-gray-900 dark:text-white"
																			>
																				{container.cpu}
																			</td>
																			<td
																				class="px-3 py-2 text-right text-gray-900 dark:text-white"
																			>
																				{container.mem}
																			</td>
																			<td
																				class="px-3 py-2 text-right text-gray-900 dark:text-white"
																			>
																				{container.rx}
																			</td>
																			<td
																				class="px-3 py-2 text-right text-gray-900 dark:text-white"
																			>
																				{container.tx}
																			</td>
																			<td
																				class="px-3 py-2 text-right text-gray-500 dark:text-gray-400"
																			>
																				{container.uptime}
																			</td>
																		</tr>
																	{/each}
																</tbody>
															</table>
														</div>
													{:else}
														<div class="text-sm text-gray-500 dark:text-gray-400">
															No containers found for this host
														</div>
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
