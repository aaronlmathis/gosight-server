<script lang="ts">
	import { onMount } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import { api } from '$lib/api';
	import type { AlertRule, AlertSummary, AlertTableData } from '$lib/types';

	let rules: AlertRule[] = [];
	let summaries: AlertSummary[] = [];
	let tableData: AlertTableData[] = [];
	let filteredData: AlertTableData[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let sortBy = 'name';
	let sortOrder: 'asc' | 'desc' = 'asc';

	// Bulk selection state
	let selectedIds: Set<string> = new Set();
	let selectAll = false;
	let bulkActionsOpen = false;

	onMount(async () => {
		await loadAlerts();
	});

	async function loadAlerts() {
		try {
			loading = true;
			error = '';

			// Load rules and summaries like the old frontend
			const [rulesResponse, summariesResponse] = await Promise.all([
				api.getAlertRules(),
				api.getSummary()
			]);

			rules = rulesResponse;
			summaries = summariesResponse;

			buildTableData();
			filterAndSortData();
		} catch (err) {
			error = 'Failed to load alerts: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	function renderExpression(expr: any): string {
		if (typeof expr === 'object' && expr !== null) {
			const left = expr.datatype || 'value';
			const op = expr.operator || '?';
			let val = expr.value;

			if (typeof val === 'number' && expr.datatype === 'percent') {
				val = `${val}%`;
			} else if (typeof val === 'string') {
				val = `"${val}"`;
			}

			return `${left} ${op} ${val}`;
		}
		return typeof expr === 'string' ? expr : '-';
	}

	function formatMatchCriteria(match: any): string {
		if (!match || Object.keys(match).length === 0) {
			return '';
		}
		let criteria: string[] = [];

		if (match.labels && typeof match.labels === 'object') {
			for (const [k, v] of Object.entries(match.labels)) {
				criteria.push(`label:${k}=${v}`);
			}
		}

		for (const [k, v] of Object.entries(match)) {
			if (k !== 'labels') {
				criteria.push(`${k}=${v}`);
			}
		}

		return '(' + criteria.join(', ') + ')';
	}

	function buildTableData() {
		tableData = rules.map((rule) => {
			const summary = summaries.find((s) => s.rule_id === rule.id);

			let state = 'Insufficient Data';
			let lastStateChange = '-';

			if (summary) {
				if (summary.state === 'firing') {
					state = 'Alarm';
				} else if (summary.state === 'resolved') {
					state = 'OK';
				}
				lastStateChange = summary.last_change;
			}

			return {
				id: rule.id,
				name: rule.name,
				state: state,
				last_state_change: lastStateChange,
				conditions_summary: `${renderExpression(rule.expression)} ${formatMatchCriteria(rule.match)}`,
				actions: rule.actions || []
			};
		});
	}

	function filterAndSortData() {
		let filtered = [...tableData];

		// Search filter
		if (searchTerm) {
			filtered = filtered.filter(
				(alert) =>
					alert.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
					alert.conditions_summary.toLowerCase().includes(searchTerm.toLowerCase())
			);
		}

		// Sort
		filtered.sort((a, b) => {
			let aValue = a[sortBy as keyof AlertTableData];
			let bValue = b[sortBy as keyof AlertTableData];

			if (typeof aValue === 'string') aValue = aValue.toLowerCase();
			if (typeof bValue === 'string') bValue = bValue.toLowerCase();

			if (sortOrder === 'asc') {
				return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
			} else {
				return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
			}
		});

		filteredData = filtered;
		updateSelectAllState();
	}

	function handleSort(field: string) {
		if (sortBy === field) {
			sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
		} else {
			sortBy = field;
			sortOrder = 'asc';
		}
		filterAndSortData();
	}

	function handleSearch() {
		filterAndSortData();
	}

	function editAlert(alertId: string) {
		window.location.href = `/alerts/edit/${alertId}`;
	}

	// Bulk selection functions
	function handleSelectAll() {
		if (selectAll) {
			selectedIds = new Set(filteredData.map((alert) => alert.id));
		} else {
			selectedIds = new Set();
		}
		updateBulkActionsState();
	}

	function handleRowSelect(alertId: string) {
		if (selectedIds.has(alertId)) {
			selectedIds.delete(alertId);
		} else {
			selectedIds.add(alertId);
		}
		selectedIds = new Set(selectedIds); // Trigger reactivity
		updateSelectAllState();
		updateBulkActionsState();
	}

	function updateSelectAllState() {
		selectAll = filteredData.length > 0 && filteredData.every((alert) => selectedIds.has(alert.id));
	}

	function updateBulkActionsState() {
		// Enable/disable bulk actions button based on selection
	}

	function handleBulkDisable() {
		console.log('Disabling selected:', Array.from(selectedIds));
		bulkActionsOpen = false;
		// TODO: Implement bulk disable API call
	}

	function handleBulkDelete() {
		if (confirm('Are you sure you want to delete the selected alerts?')) {
			console.log('Deleting selected:', Array.from(selectedIds));
			bulkActionsOpen = false;
			// TODO: Implement bulk delete API call
		}
	}

	function getStateBadgeClass(state: string): string {
		switch (state) {
			case 'Alarm':
				return 'bg-red-100 text-red-800';
			case 'OK':
				return 'bg-green-100 text-green-800';
			default:
				return 'bg-gray-300 text-gray-800';
		}
	}

	function handleRowClick(event: MouseEvent, alertId: string) {
		// Prevent row click when clicking on checkbox or buttons
		const target = event.target as HTMLElement;
		if (target.tagName === 'INPUT' || target.tagName === 'BUTTON' || target.closest('button')) {
			return;
		}
		handleRowSelect(alertId);
	}

	// Close bulk actions menu when clicking outside
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as HTMLElement;
		if (!target.closest('.bulk-actions-container')) {
			bulkActionsOpen = false;
		}
	}

	$: bulkActionsEnabled = selectedIds.size > 0;
</script>

<svelte:head>
	<title>Alerts - GoSight</title>
</svelte:head>

<svelte:window on:click={handleClickOutside} />

<PermissionGuard requiredPermission="gosight:dashboard:view">
	<div class="p-6">
		<div class="mb-6">
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Alerts</h1>
			<p class="text-gray-600 dark:text-gray-400">Monitor and manage alert rules</p>
		</div>

		<!-- Search and Controls -->
		<div class="mb-6 rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<div class="flex flex-wrap items-center gap-4">
					<!-- Search -->
					<div class="relative">
						<i
							class="fas fa-search absolute top-1/2 left-3 -translate-y-1/2 transform text-gray-400"
						></i>
						<input
							type="text"
							placeholder="Search alerts..."
							bind:value={searchTerm}
							on:input={handleSearch}
							class="rounded-lg border border-gray-300 bg-white py-2 pr-4 pl-10 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
					</div>

					<!-- Bulk Actions -->
					<div class="bulk-actions-container relative">
						<button
							on:click={() => (bulkActionsOpen = !bulkActionsOpen)}
							disabled={!bulkActionsEnabled}
							class="inline-flex items-center rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
						>
							Actions
							<i class="fas fa-chevron-down ml-2"></i>
						</button>

						{#if bulkActionsOpen}
							<div
								class="absolute left-0 z-10 mt-2 w-48 rounded-md border border-gray-200 bg-white py-1 shadow-lg dark:border-gray-600 dark:bg-gray-700"
							>
								<button
									on:click={handleBulkDisable}
									class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-600"
								>
									<i class="fas fa-pause mr-2"></i>
									Disable Selected
								</button>
								<button
									on:click={handleBulkDelete}
									class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-100 dark:text-red-400 dark:hover:bg-gray-600"
								>
									<i class="fas fa-trash mr-2"></i>
									Delete Selected
								</button>
							</div>
						{/if}
					</div>
				</div>

				<div class="flex gap-2">
					<a
						href="/alerts/add"
						class="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
					>
						<i class="fas fa-plus mr-2"></i>
						Add
					</a>
				</div>
			</div>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-12">
				<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div
				class="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
			>
				<div class="flex">
					<i class="fas fa-exclamation-triangle mt-0.5 mr-3 text-red-500"></i>
					<div>
						<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
						<p class="mt-1 text-sm text-red-600 dark:text-red-300">{error}</p>
					</div>
				</div>
			</div>
		{:else}
			<!-- Alerts Table -->
			<div class="overflow-hidden rounded-lg bg-white shadow dark:bg-gray-800">
				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
						<thead class="bg-gray-50 dark:bg-gray-700">
							<tr>
								<th
									class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
								>
									<input
										type="checkbox"
										bind:checked={selectAll}
										on:change={handleSelectAll}
										class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
								</th>
								<th
									class="cursor-pointer px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-600"
									on:click={() => handleSort('name')}
								>
									Name
									{#if sortBy === 'name'}
										<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
									{/if}
								</th>
								<th
									class="cursor-pointer px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-600"
									on:click={() => handleSort('state')}
								>
									State
									{#if sortBy === 'state'}
										<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
									{/if}
								</th>
								<th
									class="cursor-pointer px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-600"
									on:click={() => handleSort('last_state_change')}
								>
									Last Fired
									{#if sortBy === 'last_state_change'}
										<i class="fas fa-sort-{sortOrder === 'asc' ? 'up' : 'down'} ml-1"></i>
									{/if}
								</th>
								<th
									class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
								>
									Conditions
								</th>
								<th
									class="px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase dark:text-gray-300"
								>
									Actions
								</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-800">
							{#each filteredData as alert (alert.id)}
								<tr
									class="cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 {selectedIds.has(
										alert.id
									)
										? 'bg-blue-50 dark:bg-blue-900/20'
										: ''}"
									on:click={(e) => handleRowClick(e, alert.id)}
								>
									<td class="px-6 py-4 whitespace-nowrap">
										<input
											type="checkbox"
											checked={selectedIds.has(alert.id)}
											on:change={() => handleRowSelect(alert.id)}
											class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
										/>
									</td>
									<td class="px-6 py-4">
										<div class="text-sm font-medium text-gray-900 dark:text-white">
											{alert.name}
										</div>
									</td>
									<td class="px-6 py-4 whitespace-nowrap">
										<span
											class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStateBadgeClass(
												alert.state
											)}"
										>
											{alert.state}
										</span>
									</td>
									<td class="px-6 py-4 text-sm whitespace-nowrap text-gray-900 dark:text-white">
										{alert.last_state_change}
									</td>
									<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">
										{alert.conditions_summary}
									</td>
									<td class="px-6 py-4 text-sm whitespace-nowrap">
										<div class="flex gap-2">
											<button
												on:click={() => editAlert(alert.id)}
												class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300"
												title="Edit"
												aria-label="Edit alert"
											>
												<i class="fas fa-edit"></i>
											</button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if filteredData.length === 0}
					<div class="py-12 text-center">
						<i class="fas fa-exclamation-triangle mb-4 text-4xl text-gray-400"></i>
						<h3 class="mb-2 text-lg font-medium text-gray-900 dark:text-white">No alerts found</h3>
						<p class="text-gray-500 dark:text-gray-400">
							Try adjusting your search criteria or create a new alert rule.
						</p>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</PermissionGuard>
