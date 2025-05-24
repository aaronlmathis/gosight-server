<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { formatDate } from '$lib/utils';
	import {
		Button,
		Badge,
		Spinner,
		Input,
		Select,
		Table,
		TableBody,
		TableBodyCell,
		TableBodyRow,
		TableHead,
		TableHeadCell,
		Pagination
	} from 'flowbite-svelte';
	import {
		SearchOutline,
		DownloadOutline,
		FilterOutline,
		CloseOutline
	} from 'flowbite-svelte-icons';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';

	interface AlertHistoryItem {
		id: string;
		rule_id: string;
		state: string;
		level: string;
		target: string;
		scope: string;
		message: string;
		first_fired: string;
		last_ok: string;
		value?: number;
		endpoint_id?: string;
		tags?: { [key: string]: string };
	}

	let alerts: AlertHistoryItem[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let stateFilter = '';
	let levelFilter = '';
	let expandedRows: Set<string> = new Set();
	let alertContext: { [alertId: string]: any } = {};
	let loadingContext: Set<string> = new Set();

	// Pagination
	let currentPage = 1;
	let pageSize = 20;
	let totalCount = 0;
	let totalPages = 0;

	// Sorting
	let sortField = 'first_fired';
	let sortOrder: 'asc' | 'desc' = 'desc';

	// Tag filters
	let activeTagFilters: string[] = [];

	onMount(async () => {
		// Load initial parameters from URL
		loadParamsFromURL();
		await loadAlertHistory();
	});

	function loadParamsFromURL() {
		const urlParams = $page.url.searchParams;

		searchTerm = urlParams.get('rule_id') || '';
		stateFilter = urlParams.get('state') || '';
		levelFilter = urlParams.get('level') || '';
		currentPage = parseInt(urlParams.get('page') || '1');
		sortField = urlParams.get('sort') || 'first_fired';
		sortOrder = (urlParams.get('order') as 'asc' | 'desc') || 'desc';
		activeTagFilters = urlParams.getAll('tag');
	}

	function updateURL() {
		const params = new URLSearchParams();

		if (searchTerm) params.set('rule_id', searchTerm);
		if (stateFilter) params.set('state', stateFilter);
		if (levelFilter) params.set('level', levelFilter);
		if (currentPage > 1) params.set('page', currentPage.toString());
		if (sortField !== 'first_fired') params.set('sort', sortField);
		if (sortOrder !== 'desc') params.set('order', sortOrder);
		activeTagFilters.forEach((tag) => params.append('tag', tag));

		const newURL = `${$page.url.pathname}?${params.toString()}`;
		goto(newURL, { replaceState: true, noScroll: true });
	}

	async function loadAlertHistory() {
		try {
			loading = true;
			error = '';

			const params: any = {
				limit: pageSize,
				page: currentPage,
				sort: sortField,
				order: sortOrder
			};

			if (searchTerm) params.rule_id = searchTerm;
			if (stateFilter) params.state = stateFilter;
			if (levelFilter) params.level = levelFilter;

			const response = await api.alerts.getAll(params);
			alerts = Array.isArray(response) ? response : [];

			// Note: In a real implementation, you'd need the total count from headers
			// For now, we'll estimate based on page size
			totalCount =
				alerts.length === pageSize
					? currentPage * pageSize + 1
					: (currentPage - 1) * pageSize + alerts.length;
			totalPages = Math.ceil(totalCount / pageSize);
		} catch (err) {
			error = 'Failed to load alert history: ' + (err as Error).message;
			console.error('Error loading alert history:', err);
		} finally {
			loading = false;
		}
	}

	function getStateBadge(state: string) {
		switch (state) {
			case 'firing':
				return 'red';
			case 'pending':
				return 'yellow';
			case 'resolved':
				return 'green';
			default:
				return 'gray';
		}
	}

	function getSeverityBadge(level: string) {
		switch (level) {
			case 'critical':
				return 'red';
			case 'warning':
				return 'yellow';
			case 'info':
				return 'blue';
			default:
				return 'gray';
		}
	}

	async function toggleExpandRow(alertId: string) {
		if (expandedRows.has(alertId)) {
			expandedRows.delete(alertId);
			expandedRows = new Set(expandedRows);
		} else {
			expandedRows.add(alertId);
			expandedRows = new Set(expandedRows);

			// Load context if not already loaded
			if (!alertContext[alertId] && !loadingContext.has(alertId)) {
				loadingContext.add(alertId);
				loadingContext = new Set(loadingContext);

				try {
					const context = await api.alerts.getContext(alertId, '1h');
					alertContext[alertId] = context;
					alertContext = { ...alertContext };
				} catch (err) {
					console.error('Failed to load alert context:', err);
				} finally {
					loadingContext.delete(alertId);
					loadingContext = new Set(loadingContext);
				}
			}
		}
	}

	function handleSort(field: string) {
		if (sortField === field) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortOrder = 'desc';
		}
		currentPage = 1;
		updateURL();
		loadAlertHistory();
	}

	function handlePageChange(event: CustomEvent) {
		currentPage = event.detail;
		updateURL();
		loadAlertHistory();
	}

	function clearFilters() {
		searchTerm = '';
		stateFilter = '';
		levelFilter = '';
		activeTagFilters = [];
		currentPage = 1;
		updateURL();
		loadAlertHistory();
	}

	function removeTagFilter(tag: string) {
		activeTagFilters = activeTagFilters.filter((t) => t !== tag);
		updateURL();
		loadAlertHistory();
	}

	async function exportData(format: 'json' | 'csv' | 'yaml') {
		try {
			// For export, get all data without pagination
			const params: any = {
				limit: 10000,
				page: 1,
				sort: sortField,
				order: sortOrder
			};

			if (searchTerm) params.rule_id = searchTerm;
			if (stateFilter) params.state = stateFilter;
			if (levelFilter) params.level = levelFilter;

			const exportData = await api.alerts.getAll(params);

			let content = '';
			let filename = `alerts_history_${new Date().toISOString().split('T')[0]}`;
			let mimeType = '';

			switch (format) {
				case 'json':
					content = JSON.stringify(exportData, null, 2);
					filename += '.json';
					mimeType = 'application/json';
					break;
				case 'csv':
					// Simple CSV export
					const headers = [
						'rule_id',
						'state',
						'level',
						'target',
						'scope',
						'first_fired',
						'last_ok'
					];
					const csvRows = [headers.join(',')];
					exportData.forEach((alert: AlertHistoryItem) => {
						const row = headers.map((header) => {
							const value = alert[header as keyof AlertHistoryItem] || '';
							return `"${String(value).replace(/"/g, '""')}"`;
						});
						csvRows.push(row.join(','));
					});
					content = csvRows.join('\n');
					filename += '.csv';
					mimeType = 'text/csv';
					break;
				case 'yaml':
					// Simple YAML export
					content =
						'---\n' +
						exportData
							.map((alert) =>
								Object.entries(alert)
									.map(([key, value]) => `${key}: ${JSON.stringify(value)}`)
									.join('\n')
							)
							.join('\n---\n');
					filename += '.yaml';
					mimeType = 'application/yaml';
					break;
			}

			// Download the file
			const blob = new Blob([content], { type: mimeType });
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err) {
			console.error('Failed to export data:', err);
		}
	}

	// Reactive statements for search debouncing
	let searchTimeout: NodeJS.Timeout;
	$: {
		if (searchTerm !== undefined || stateFilter !== undefined || levelFilter !== undefined) {
			// Clear existing timeout
			if (searchTimeout) {
				clearTimeout(searchTimeout);
			}
			// Debounce search
			searchTimeout = setTimeout(() => {
				currentPage = 1;
				updateURL();
				loadAlertHistory();
			}, 300);
		}
	}
</script>

<svelte:head>
	<title>Alert History - GoSight</title>
</svelte:head>

<div class="space-y-6 p-4">
	<!-- Header -->
	<div class="mb-6">
		<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Alert History</h1>
		<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">View history of alert instances...</p>
	</div>

	<!-- Filters and Actions -->
	<div class="mb-6 space-y-4">
		<!-- Row: Search + Filter + Export -->
		<div class="flex flex-wrap items-center justify-between gap-4">
			<!-- Search Input -->
			<div class="relative">
				<SearchOutline
					class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-500"
				/>
				<Input
					bind:value={searchTerm}
					placeholder="Search message, target, rule ID..."
					class="w-full pl-10 md:w-80"
				/>
			</div>

			<!-- Filters + Export Buttons -->
			<div class="ml-auto flex flex-wrap items-center gap-2">
				<Select bind:value={stateFilter} class="w-32">
					<option value="">All States</option>
					<option value="firing">Firing</option>
					<option value="resolved">Resolved</option>
				</Select>

				<Select bind:value={levelFilter} class="w-32">
					<option value="">All Levels</option>
					<option value="info">Info</option>
					<option value="warning">Warning</option>
					<option value="error">Error</option>
				</Select>

				<Button color="blue" size="sm" on:click={() => exportData('json')}>
					<DownloadOutline class="mr-1 h-3 w-3" />
					JSON
				</Button>
				<Button color="yellow" size="sm" on:click={() => exportData('yaml')}>
					<DownloadOutline class="mr-1 h-3 w-3" />
					YAML
				</Button>
				<Button color="green" size="sm" on:click={() => exportData('csv')}>
					<DownloadOutline class="mr-1 h-3 w-3" />
					CSV
				</Button>
				<Button color="gray" size="sm" on:click={clearFilters}>
					<FilterOutline class="mr-1 h-3 w-3" />
					Clear
				</Button>
			</div>
		</div>

		<!-- Active Tag Filters -->
		{#if activeTagFilters.length > 0}
			<div class="flex flex-wrap gap-2">
				<span class="text-sm text-gray-600 dark:text-gray-400">Active filters:</span>
				{#each activeTagFilters as tag}
					<Badge color="blue" class="cursor-pointer">
						{tag}
						<button on:click={() => removeTagFilter(tag)} class="ml-2">
							<CloseOutline class="h-3 w-3" />
						</button>
					</Badge>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Loading State -->
	{#if loading}
		<div class="flex items-center justify-center py-12">
			<Spinner size="8" />
		</div>
	{/if}

	<!-- Error State -->
	{#if error}
		<div
			class="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
		>
			<p class="text-red-800 dark:text-red-200">{error}</p>
		</div>
	{/if}

	<!-- Alert History Table -->
	{#if !loading && !error}
		<div class="overflow-x-auto rounded-lg border border-gray-200 shadow dark:border-gray-700">
			<Table>
				<TableHead>
					<TableHeadCell class="cursor-pointer" on:click={() => handleSort('rule_id')}>
						Rule {sortField === 'rule_id' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
					</TableHeadCell>
					<TableHeadCell class="cursor-pointer" on:click={() => handleSort('state')}>
						State {sortField === 'state' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
					</TableHeadCell>
					<TableHeadCell class="cursor-pointer" on:click={() => handleSort('level')}>
						Severity {sortField === 'level' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
					</TableHeadCell>
					<TableHeadCell>Target</TableHeadCell>
					<TableHeadCell>Scope</TableHeadCell>
					<TableHeadCell class="cursor-pointer" on:click={() => handleSort('first_fired')}>
						First Fired {sortField === 'first_fired' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
					</TableHeadCell>
					<TableHeadCell class="cursor-pointer" on:click={() => handleSort('last_ok')}>
						Last OK {sortField === 'last_ok' ? (sortOrder === 'asc' ? '↑' : '↓') : ''}
					</TableHeadCell>
					<TableHeadCell>Expand</TableHeadCell>
				</TableHead>
				<TableBody>
					{#each alerts as alert (alert.id)}
						<TableBodyRow class="hover:bg-gray-50 dark:hover:bg-gray-800">
							<TableBodyCell>
								<button
									class="font-medium text-blue-600 hover:underline dark:text-blue-400"
									on:click={() => toggleExpandRow(alert.id)}
								>
									{alert.rule_id}
								</button>
							</TableBodyCell>
							<TableBodyCell>
								<Badge color={getStateBadge(alert.state)}>{alert.state}</Badge>
							</TableBodyCell>
							<TableBodyCell>
								<Badge color={getSeverityBadge(alert.level)}>{alert.level}</Badge>
							</TableBodyCell>
							<TableBodyCell>{alert.target || '-'}</TableBodyCell>
							<TableBodyCell>{alert.scope || '-'}</TableBodyCell>
							<TableBodyCell>{formatDate(alert.first_fired)}</TableBodyCell>
							<TableBodyCell>{formatDate(alert.last_ok)}</TableBodyCell>
							<TableBodyCell>
								<Button size="xs" color="alternative" on:click={() => toggleExpandRow(alert.id)}>
									{expandedRows.has(alert.id) ? '[-]' : '[+]'}
								</Button>
							</TableBodyCell>
						</TableBodyRow>

						<!-- Expanded Row Content -->
						{#if expandedRows.has(alert.id)}
							<TableBodyRow class="bg-gray-50 dark:bg-gray-800/50">
								<TableBodyCell colspan="8">
									<div class="space-y-4 p-4">
										<!-- Alert Details -->
										<div class="grid grid-cols-2 gap-4">
											<div>
												<h4 class="mb-2 font-semibold text-gray-800 dark:text-white">
													Alert Details
												</h4>
												<div class="space-y-1 text-sm">
													<p>
														<span class="font-medium">Message:</span>
														{alert.message || 'No message'}
													</p>
													<p><span class="font-medium">Value:</span> {alert.value || 'N/A'}</p>
													<p>
														<span class="font-medium">Endpoint:</span>
														{alert.endpoint_id || 'N/A'}
													</p>
												</div>
											</div>
											<div>
												<h4 class="mb-2 font-semibold text-gray-800 dark:text-white">Timeline</h4>
												<div class="space-y-1 text-sm">
													<p>
														<span class="font-medium">First Fired:</span>
														{formatDate(alert.first_fired)}
													</p>
													<p>
														<span class="font-medium">Last OK:</span>
														{formatDate(alert.last_ok)}
													</p>
												</div>
											</div>
										</div>

										<!-- Tags -->
										{#if alert.tags && Object.keys(alert.tags).length > 0}
											<div>
												<h4 class="mb-2 font-semibold text-gray-800 dark:text-white">Tags</h4>
												<div class="flex flex-wrap gap-2">
													{#each Object.entries(alert.tags) as [key, value]}
														<Badge color="gray" class="text-xs">{key}: {value}</Badge>
													{/each}
												</div>
											</div>
										{/if}

										<!-- Context Loading/Display -->
										{#if loadingContext.has(alert.id)}
											<div class="flex items-center gap-2">
												<Spinner size="4" />
												<span class="text-sm text-gray-500">Loading context...</span>
											</div>
										{:else if alertContext[alert.id]}
											<div class="border-t pt-4">
												<h4 class="mb-2 font-semibold text-gray-800 dark:text-white">
													Related Context
												</h4>
												<div
													class="max-h-60 overflow-y-auto rounded-lg bg-white p-3 text-sm dark:bg-gray-900"
												>
													<pre class="whitespace-pre-wrap">{JSON.stringify(
															alertContext[alert.id],
															null,
															2
														)}</pre>
												</div>
											</div>
										{/if}
									</div>
								</TableBodyCell>
							</TableBodyRow>
						{/if}
					{/each}

					{#if alerts.length === 0 && !loading}
						<TableBodyRow>
							<TableBodyCell colspan="8" class="py-6 text-center text-gray-500">
								{searchTerm || stateFilter || levelFilter
									? 'No alerts match your filters'
									: 'No alert history found'}
							</TableBodyCell>
						</TableBodyRow>
					{/if}
				</TableBody>
			</Table>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="mt-6 flex justify-center">
				<Pagination
					pages={[
						{ name: '1', href: '#', active: currentPage === 1 },
						...(totalPages > 1
							? Array.from({ length: Math.min(totalPages - 1, 4) }, (_, i) => ({
									name: String(i + 2),
									href: '#',
									active: currentPage === i + 2
								}))
							: [])
					]}
					on:previous={() => currentPage > 1 && (currentPage--, updateURL(), loadAlertHistory())}
					on:next={() =>
						currentPage < totalPages && (currentPage++, updateURL(), loadAlertHistory())}
					on:click={handlePageChange}
				/>
			</div>
		{/if}

		<!-- Results summary -->
		<div class="text-center text-sm text-gray-500 dark:text-gray-400">
			Showing {(currentPage - 1) * pageSize + 1} to {Math.min(currentPage * pageSize, totalCount)} of
			{totalCount} alerts
		</div>
	{/if}
</div>
