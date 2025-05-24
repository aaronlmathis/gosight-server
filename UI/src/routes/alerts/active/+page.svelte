<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
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
		TableHeadCell
	} from 'flowbite-svelte';
	import { RefreshOutline, SearchOutline } from 'flowbite-svelte-icons';

	interface ActiveAlert {
		id: string;
		rule_id: string;
		state: string;
		level: string;
		target: string;
		scope: string;
		message: string;
		first_fired: string;
		last_fired: string;
		value?: number;
		endpoint_id?: string;
	}

	let alerts: ActiveAlert[] = [];
	let filteredAlerts: ActiveAlert[] = [];
	let loading = true;
	let error = '';
	let searchTerm = '';
	let stateFilter = '';
	let levelFilter = '';
	let expandedRows: Set<string> = new Set();
	let alertContext: { [alertId: string]: any } = {};
	let loadingContext: Set<string> = new Set();

	// Auto-refresh functionality
	let refreshInterval: number;
	const REFRESH_INTERVAL = 30000; // 30 seconds

	onMount(async () => {
		await loadActiveAlerts();
		// Set up auto-refresh
		refreshInterval = setInterval(loadActiveAlerts, REFRESH_INTERVAL);
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});

	async function loadActiveAlerts() {
		try {
			loading = true;
			error = '';

			// Get active alerts from the API
			const response = await api.alerts.getActive();
			alerts = Array.isArray(response) ? response : [];

			filterAlerts();
		} catch (err) {
			error = 'Failed to load active alerts: ' + (err as Error).message;
			console.error('Error loading active alerts:', err);
		} finally {
			loading = false;
		}
	}

	function filterAlerts() {
		filteredAlerts = alerts.filter((alert) => {
			const matchesSearch =
				!searchTerm ||
				alert.rule_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
				alert.target?.toLowerCase().includes(searchTerm.toLowerCase()) ||
				alert.message?.toLowerCase().includes(searchTerm.toLowerCase());

			const matchesState = !stateFilter || alert.state === stateFilter;
			const matchesLevel = !levelFilter || alert.level === levelFilter;

			return matchesSearch && matchesState && matchesLevel;
		});
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

	async function acknowledgeAlert(alertId: string) {
		try {
			await api.alerts.acknowledge(alertId);
			await loadActiveAlerts(); // Refresh
		} catch (err) {
			console.error('Failed to acknowledge alert:', err);
		}
	}

	async function resolveAlert(alertId: string) {
		try {
			await api.alerts.resolve(alertId);
			await loadActiveAlerts(); // Refresh
		} catch (err) {
			console.error('Failed to resolve alert:', err);
		}
	}

	// Reactive statements
	$: {
		searchTerm, stateFilter, levelFilter;
		filterAlerts();
	}
</script>

<svelte:head>
	<title>Active Alerts - GoSight</title>
</svelte:head>

<PermissionGuard requiredPermission="gosight:dashboard:view">
	<div class="space-y-6 p-6">
		<!-- Header -->
		<div class="mb-6">
			<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Active Alerts</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">
				Unresolved alerts with incident investigation tools.
			</p>
		</div>

		<!-- Filters Row -->
		<div class="mb-4 flex flex-wrap items-center gap-4">
			<div class="relative">
				<SearchOutline
					class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-500"
				/>
				<Input bind:value={searchTerm} placeholder="Search alerts..." class="w-64 pl-10" />
			</div>

			<Select bind:value={stateFilter} class="w-40">
				<option value="">All States</option>
				<option value="firing">Firing</option>
				<option value="pending">Pending</option>
				<option value="resolved">Resolved</option>
			</Select>

			<Select bind:value={levelFilter} class="w-40">
				<option value="">All Levels</option>
				<option value="info">Info</option>
				<option value="warning">Warning</option>
				<option value="critical">Critical</option>
			</Select>

			<Button on:click={loadActiveAlerts} color="light" class="ml-auto">
				<RefreshOutline class="mr-2 h-4 w-4" />
				Refresh
			</Button>
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

		<!-- Alerts Table -->
		{#if !loading && !error}
			<div class="overflow-x-auto rounded-lg border border-gray-200 shadow dark:border-gray-700">
				<Table>
					<TableHead>
						<TableHeadCell>Rule</TableHeadCell>
						<TableHeadCell>State</TableHeadCell>
						<TableHeadCell>Severity</TableHeadCell>
						<TableHeadCell>Scope</TableHeadCell>
						<TableHeadCell>Target</TableHeadCell>
						<TableHeadCell>First Fired</TableHeadCell>
						<TableHeadCell>Last Fired</TableHeadCell>
						<TableHeadCell>Actions</TableHeadCell>
					</TableHead>
					<TableBody>
						{#each filteredAlerts as alert (alert.id)}
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
								<TableBodyCell>{alert.scope || '-'}</TableBodyCell>
								<TableBodyCell>{alert.target || '-'}</TableBodyCell>
								<TableBodyCell>{formatDate(alert.first_fired)}</TableBodyCell>
								<TableBodyCell>{formatDate(alert.last_fired)}</TableBodyCell>
								<TableBodyCell>
									<div class="flex gap-2">
										<Button size="xs" color="light" on:click={() => acknowledgeAlert(alert.id)}>
											Ack
										</Button>
										<Button size="xs" color="green" on:click={() => resolveAlert(alert.id)}>
											Resolve
										</Button>
										<Button
											size="xs"
											color="alternative"
											on:click={() => toggleExpandRow(alert.id)}
										>
											{expandedRows.has(alert.id) ? '[-]' : '[+]'}
										</Button>
									</div>
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
															<span class="font-medium">Last Fired:</span>
															{formatDate(alert.last_fired)}
														</p>
													</div>
												</div>
											</div>

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
													<div class="rounded-lg bg-white p-3 text-sm dark:bg-gray-900">
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

						{#if filteredAlerts.length === 0 && !loading}
							<TableBodyRow>
								<TableBodyCell colspan="8" class="py-6 text-center text-gray-500">
									{searchTerm || stateFilter || levelFilter
										? 'No alerts match your filters'
										: 'No active alerts'}
								</TableBodyCell>
							</TableBodyRow>
						{/if}
					</TableBody>
				</Table>
			</div>
		{/if}
	</div>
</PermissionGuard>
