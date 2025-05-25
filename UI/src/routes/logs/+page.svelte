<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import PermissionGuard from '$lib/components/PermissionGuard.svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import { formatDate } from '$lib/utils';
	import type { LogEntry, Endpoint } from '$lib/types';

	let logs: LogEntry[] = [];
	let endpoints: Endpoint[] = [];
	let loading = true;
	let error = '';
	let autoRefresh = false;

	// Advanced search filters
	let searchTerm = '';
	let selectedEndpoint = '';
	let selectedSource = '';
	let selectedContainer = '';
	let selectedApp = '';
	let startTime = '';
	let endTime = '';

	// Cursor-based pagination like old system
	let currentCursor: string | null = null;
	let nextCursor: string | null = null;
	let hasMore = false;
	let cursorStack: string[] = [];
	let pageSize = 50;
	let logCount = 0;
	let firstVisibleTime: string | null = null;

	// Multi-select filters
	let selectedLevels: string[] = [];
	let selectedCategories: string[] = [];
	let activeFilters: { key: string; value: string; isCustomTag?: boolean }[] = [];
	let expandedLogs = new Set<string>();

	// Dropdown states
	let showLevelDropdown = false;
	let showCategoryDropdown = false;
	let showEndpointDropdown = false;
	let endpointSearchTerm = '';

	// Log levels and categories matching old system
	const logLevels = [
		{
			value: 'critical',
			label: 'Critical',
			color: 'text-red-800 bg-red-100 dark:text-red-100 dark:bg-red-900'
		},
		{
			value: 'error',
			label: 'Error',
			color: 'text-red-700 bg-red-100 dark:text-red-100 dark:bg-red-800'
		},
		{
			value: 'warning',
			label: 'Warning',
			color: 'text-yellow-700 bg-yellow-100 dark:text-yellow-100 dark:bg-yellow-800'
		},
		{
			value: 'info',
			label: 'Info',
			color: 'text-blue-700 bg-blue-100 dark:text-blue-100 dark:bg-blue-800'
		},
		{
			value: 'debug',
			label: 'Debug',
			color: 'text-gray-700 bg-gray-100 dark:text-gray-100 dark:bg-gray-800'
		}
	];

	const logCategories = [
		{ value: 'system', label: 'System' },
		{ value: 'application', label: 'Application' },
		{ value: 'security', label: 'Security' },
		{ value: 'performance', label: 'Performance' },
		{ value: 'auth', label: 'Auth' },
		{ value: 'network', label: 'Network' },
		{ value: 'container', label: 'Container' },
		{ value: 'metric', label: 'Metric' },
		{ value: 'gosight', label: 'GoSight' },
		{ value: 'scheduler', label: 'Scheduler' },
		{ value: 'config', label: 'Config' },
		{ value: 'audit', label: 'Audit' },
		{ value: 'alert', label: 'Alert' }
	];

	// Filtered endpoints for dropdown
	$: filteredEndpoints = endpoints.filter((endpoint) =>
		endpoint.name.toLowerCase().includes(endpointSearchTerm.toLowerCase())
	);

	// Click outside handler
	function handleClickOutside(event: MouseEvent) {
		const target = event.target as Element;

		// Close level dropdown
		if (showLevelDropdown && !target.closest('#level-dropdown-container')) {
			showLevelDropdown = false;
		}

		// Close category dropdown
		if (showCategoryDropdown && !target.closest('#category-dropdown-container')) {
			showCategoryDropdown = false;
		}

		// Close endpoint dropdown
		if (showEndpointDropdown && !target.closest('#endpoint-dropdown-container')) {
			showEndpointDropdown = false;
		}
	}

	onMount(async () => {
		if (browser) {
			document.addEventListener('click', handleClickOutside);

			// Load URL parameters
			loadFiltersFromURL();

			await loadEndpoints();
			await loadLogs();
		}
	});

	onDestroy(() => {
		if (browser) {
			document.removeEventListener('click', handleClickOutside);
			if (autoRefresh) {
				websocketManager.disconnect();
			}
		}
	});

	function loadFiltersFromURL() {
		const params = $page.url.searchParams;

		searchTerm = params.get('contains') || '';
		selectedEndpoint = params.get('endpoint') || '';
		selectedSource = params.get('source') || '';
		selectedContainer = params.get('container') || '';
		selectedApp = params.get('app') || '';
		startTime = params.get('start') || '';
		endTime = params.get('end') || '';

		// Handle multiple levels
		selectedLevels = params.getAll('level');

		// Handle multiple categories
		selectedCategories = params.getAll('category');

		// Handle cursor
		currentCursor = params.get('cursor');

		updateActiveFilters();
	}

	function updateURL() {
		if (!browser) return;

		const params = new URLSearchParams();

		if (searchTerm) params.set('contains', searchTerm);
		if (selectedEndpoint) params.set('endpoint', selectedEndpoint);
		if (selectedSource) params.set('source', selectedSource);
		if (selectedContainer) params.set('container', selectedContainer);
		if (selectedApp) params.set('app', selectedApp);
		if (startTime) params.set('start', startTime);
		if (endTime) params.set('end', endTime);

		selectedLevels.forEach((level) => params.append('level', level));
		selectedCategories.forEach((category) => params.append('category', category));

		if (currentCursor) params.set('cursor', currentCursor);

		// Add custom tag filters
		activeFilters.filter((f) => f.isCustomTag).forEach((f) => params.set(`tag_${f.key}`, f.value));

		const newUrl = params.toString() ? `?${params.toString()}` : '';
		goto(newUrl, { replaceState: true, noScroll: true });
	}

	function updateActiveFilters() {
		activeFilters = [];

		if (searchTerm) activeFilters.push({ key: 'contains', value: searchTerm });
		if (selectedEndpoint) {
			const endpoint = endpoints.find((e) => e.id === selectedEndpoint);
			activeFilters.push({ key: 'endpoint', value: endpoint?.name || selectedEndpoint });
		}
		if (selectedSource) activeFilters.push({ key: 'source', value: selectedSource });
		if (selectedContainer) activeFilters.push({ key: 'container', value: selectedContainer });
		if (selectedApp) activeFilters.push({ key: 'app', value: selectedApp });
		if (startTime)
			activeFilters.push({ key: 'start', value: new Date(startTime).toLocaleString() });
		if (endTime) activeFilters.push({ key: 'end', value: new Date(endTime).toLocaleString() });

		selectedLevels.forEach((level) => {
			const levelObj = logLevels.find((l) => l.value === level);
			activeFilters.push({ key: 'level', value: levelObj?.label || level });
		});

		selectedCategories.forEach((category) => {
			const categoryObj = logCategories.find((c) => c.value === category);
			activeFilters.push({ key: 'category', value: categoryObj?.label || category });
		});
	}

	async function loadEndpoints() {
		try {
			const response = await api.getEndpoints();
			endpoints = response || [];
		} catch (err) {
			console.error('Failed to load endpoints:', err);
		}
	}

	async function loadLogs(isNextPage = false) {
		try {
			loading = true;
			error = '';

			const params = new URLSearchParams();

			// Basic filters
			if (searchTerm) params.set('contains', searchTerm);
			if (selectedEndpoint) params.set('endpoint', selectedEndpoint);
			if (selectedSource) params.set('source', selectedSource);
			if (selectedContainer) params.set('container_name', selectedContainer);
			if (selectedApp) params.set('app_name', selectedApp);
			if (startTime) {
				const startDate = new Date(startTime);
				if (!isNaN(startDate.getTime())) {
					params.set('start', startDate.toISOString());
				}
			}
			if (endTime) {
				const endDate = new Date(endTime);
				if (!isNaN(endDate.getTime())) {
					params.set('end', endDate.toISOString());
				}
			}

			// Multi-select filters
			selectedLevels.forEach((level) => params.append('level', level));
			selectedCategories.forEach((category) => params.append('category', category));

			// Custom tag filters
			activeFilters
				.filter((f) => f.isCustomTag)
				.forEach((f) => params.set(`tag_${f.key}`, f.value));

			// Pagination
			const cursorToUse = isNextPage ? nextCursor : currentCursor;
			if (cursorToUse) {
				params.set('cursor', cursorToUse);
			}

			params.set('limit', String(pageSize));
			params.set('order', 'desc');

			const response = await fetch(`/api/v1/logs?${params.toString()}`);

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data = await response.json();
			logs = data.logs || [];
			hasMore = data.has_more || false;
			nextCursor = data.next_cursor || null;
			logCount = data.count || 0;

			// Update cursor tracking
			if (isNextPage && logs.length > 0) {
				if (currentCursor) {
					cursorStack.push(currentCursor);
				}
				currentCursor = nextCursor;
			} else if (!isNextPage) {
				// Reset for new search
				cursorStack = [];
				currentCursor = cursorToUse;
			}

			// Update first visible time for display
			if (logs.length > 0) {
				firstVisibleTime = logs[0].timestamp;
			}
		} catch (err) {
			error = 'Failed to load logs: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	function handleFilterChange() {
		// Reset pagination
		currentCursor = null;
		nextCursor = null;
		cursorStack = [];
		hasMore = false;

		updateActiveFilters();
		updateURL();
		loadLogs();
	}

	function resetFilters() {
		searchTerm = '';
		selectedEndpoint = '';
		selectedSource = '';
		selectedContainer = '';
		selectedApp = '';
		startTime = '';
		endTime = '';
		selectedLevels = [];
		selectedCategories = [];
		currentCursor = null;
		nextCursor = null;
		cursorStack = [];
		hasMore = false;

		// Remove custom tag filters
		activeFilters = activeFilters.filter((f) => !f.isCustomTag);

		updateActiveFilters();
		updateURL();
		loadLogs();
	}

	function removeFilter(filter: { key: string; value: string; isCustomTag?: boolean }) {
		if (filter.isCustomTag) {
			activeFilters = activeFilters.filter(
				(f) => !(f.key === filter.key && f.value === filter.value && f.isCustomTag)
			);
		} else {
			switch (filter.key) {
				case 'contains':
					searchTerm = '';
					break;
				case 'endpoint':
					selectedEndpoint = '';
					break;
				case 'source':
					selectedSource = '';
					break;
				case 'container':
					selectedContainer = '';
					break;
				case 'app':
					selectedApp = '';
					break;
				case 'start':
					startTime = '';
					break;
				case 'end':
					endTime = '';
					break;
				case 'level':
					const levelObj = logLevels.find((l) => l.label === filter.value);
					if (levelObj) {
						selectedLevels = selectedLevels.filter((l) => l !== levelObj.value);
					}
					break;
				case 'category':
					const categoryObj = logCategories.find((c) => c.label === filter.value);
					if (categoryObj) {
						selectedCategories = selectedCategories.filter((c) => c !== categoryObj.value);
					}
					break;
			}
		}

		handleFilterChange();
	}

	function addCustomTag(key: string, value: string) {
		// Check if tag already exists
		if (!activeFilters.some((f) => f.key === key && f.value === value && f.isCustomTag)) {
			activeFilters = [...activeFilters, { key, value, isCustomTag: true }];
			updateURL();
			loadLogs();
		}
	}

	function handlePrevPage() {
		if (cursorStack.length > 0) {
			const prevCursor = cursorStack.pop();
			currentCursor = prevCursor || null;
			nextCursor = null;
			loadLogs();
		}
	}

	function handleNextPage() {
		if (hasMore && nextCursor) {
			loadLogs(true);
		}
	}

	function toggleAutoRefresh() {
		autoRefresh = !autoRefresh;
		if (autoRefresh) {
			websocketManager.connect();
			websocketManager.subscribeToLogs((log: LogEntry) => {
				// Add new log to the beginning if it matches current filters
				logs = [log, ...logs];
				// Keep only recent logs to prevent memory issues
				if (logs.length > pageSize * 3) {
					logs = logs.slice(0, pageSize * 3);
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
		const csvContent = logs
			.map(
				(log) =>
					`"${formatDate(log.timestamp)}","${log.level}","${log.source || ''}","${log.message.replace(/"/g, '""')}"`
			)
			.join('\n');

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
		const levelObj = logLevels.find((l) => l.value === level);
		return levelObj?.color || 'text-gray-600 bg-gray-100 dark:text-gray-300 dark:bg-gray-700';
	}

	function getLogLevelIcon(level: string): string {
		switch (level) {
			case 'critical':
			case 'error':
				return 'fas fa-exclamation-circle';
			case 'warning':
				return 'fas fa-exclamation-triangle';
			case 'info':
				return 'fas fa-info-circle';
			case 'debug':
				return 'fas fa-bug';
			default:
				return 'fas fa-file-alt';
		}
	}

	function toggleLogExpanded(logId: string) {
		if (expandedLogs.has(logId)) {
			expandedLogs.delete(logId);
		} else {
			expandedLogs.add(logId);
		}
		expandedLogs = expandedLogs; // Trigger reactivity
	}

	function handleLevelChange(level: string, checked: boolean) {
		if (checked) {
			selectedLevels = [...selectedLevels, level];
		} else {
			selectedLevels = selectedLevels.filter((l) => l !== level);
		}
		handleFilterChange();
	}

	function handleCategoryChange(category: string, checked: boolean) {
		if (checked) {
			selectedCategories = [...selectedCategories, category];
		} else {
			selectedCategories = selectedCategories.filter((c) => c !== category);
		}
		handleFilterChange();
	}

	function selectEndpoint(endpoint: Endpoint) {
		selectedEndpoint = endpoint.id;
		endpointSearchTerm = endpoint.name;
		showEndpointDropdown = false;
		handleFilterChange();
	}

	// Get level button text
	$: levelButtonText =
		selectedLevels.length === 0
			? 'Select levels'
			: selectedLevels.length === 1
				? logLevels.find((l) => l.value === selectedLevels[0])?.label || 'Selected'
				: `${selectedLevels.length} selected`;

	// Get category button text
	$: categoryButtonText =
		selectedCategories.length === 0
			? 'Select categories'
			: selectedCategories.length === 1
				? logCategories.find((c) => c.value === selectedCategories[0])?.label || 'Selected'
				: `${selectedCategories.length} selected`;
</script>

<svelte:head>
	<title>Log Explorer - GoSight</title>
</svelte:head>

<PermissionGuard requiredPermission="gosight:dashboard:view">
	<div class="space-y-6 p-4">
		<!-- Header -->
		<div class="mb-6 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">Log Explorer</h1>
				<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
					Search and filter logs from your endpoints.
				</p>
			</div>
			<div class="flex gap-2">
				<button
					on:click={toggleAutoRefresh}
					class="rounded-lg border px-4 py-2 text-sm transition-colors {autoRefresh
						? 'border-green-300 bg-green-100 text-green-800 dark:border-green-800 dark:bg-green-900/20 dark:text-green-400'
						: 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'}"
				>
					<i class="fas fa-{autoRefresh ? 'pause' : 'play'} mr-2"></i>
					{autoRefresh ? 'Pause' : 'Start'} Auto-refresh
				</button>
				<button
					on:click={exportLogs}
					disabled={logs.length === 0}
					class="rounded-lg border border-transparent bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
				>
					<i class="fas fa-download mr-2"></i>
					Export CSV
				</button>
			</div>
		</div>

		<!-- Advanced Log Search Form -->
		<form on:submit|preventDefault={handleFilterChange} class="space-y-6">
			<div
				class="rounded-lg border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-900"
			>
				<h2 class="mb-6 text-lg font-semibold text-gray-900 dark:text-white">
					Advanced Log Search
				</h2>

				<div class="space-y-4">
					<!-- Row 1: Keyword | Endpoint | Source -->
					<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
						<div>
							<label
								for="filter-keyword"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300">Keyword</label
							>
							<input
								type="text"
								id="filter-keyword"
								bind:value={searchTerm}
								on:input={handleFilterChange}
								placeholder="Search logs..."
								class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							/>
						</div>

						<div id="endpoint-dropdown-container">
							<label
								for="endpoint-name"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300"
								>Endpoint Name</label
							>
							<div class="relative mt-1">
								<input
									type="text"
									id="endpoint-name"
									bind:value={endpointSearchTerm}
									on:focus={() => (showEndpointDropdown = true)}
									on:input={() => (showEndpointDropdown = true)}
									placeholder="Search endpoints..."
									autocomplete="off"
									class="min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
								/>
								{#if showEndpointDropdown && filteredEndpoints.length > 0}
									<div
										class="absolute z-10 mt-1 max-h-60 w-full overflow-y-auto rounded-md border border-gray-300 bg-white text-sm shadow-lg dark:border-gray-700 dark:bg-gray-900"
									>
										{#each filteredEndpoints as endpoint}
											<button
												type="button"
												on:click={() => selectEndpoint(endpoint)}
												class="w-full px-4 py-2 text-left text-gray-900 hover:bg-gray-50 dark:text-white dark:hover:bg-gray-700"
											>
												{endpoint.name}
											</button>
										{/each}
									</div>
								{/if}
							</div>
						</div>

						<div>
							<label
								for="filter-source"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300">Source</label
							>
							<input
								type="text"
								id="filter-source"
								bind:value={selectedSource}
								on:input={handleFilterChange}
								placeholder="e.g. docker, systemd"
								class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							/>
						</div>
					</div>

					<!-- Row 2: Container Name | App Name -->
					<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
						<div>
							<label
								for="container-name"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300"
								>Container Name</label
							>
							<input
								type="text"
								id="container-name"
								bind:value={selectedContainer}
								on:input={handleFilterChange}
								placeholder="Filter by container name"
								class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							/>
						</div>

						<div>
							<label
								for="app-name"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300"
								>Application Name</label
							>
							<input
								type="text"
								id="app-name"
								bind:value={selectedApp}
								on:input={handleFilterChange}
								placeholder="Filter by application name"
								class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-800 dark:text-white"
							/>
						</div>
					</div>

					<!-- Row 3: Log Level | Category | Start/End Time -->
					<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
						<!-- Log Level Multi-select -->
						<div id="level-dropdown-container">
							<label
								for="filter-level"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300">Log Level</label
							>
							<div class="relative">
								<button
									type="button"
									on:click={() => (showLevelDropdown = !showLevelDropdown)}
									class="mt-1 inline-flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-900 shadow-sm hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
								>
									{levelButtonText}
									<svg class="ml-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 9l-7 7-7-7"
										/>
									</svg>
								</button>
								{#if showLevelDropdown}
									<div
										class="absolute z-10 mt-1 w-full divide-y divide-gray-100 rounded-lg bg-white shadow dark:divide-gray-700 dark:bg-gray-800"
									>
										<ul class="space-y-1 p-3 text-sm text-gray-700 dark:text-gray-200">
											{#each logLevels as level}
												<li>
													<label class="flex items-center">
														<input
															type="checkbox"
															checked={selectedLevels.includes(level.value)}
															on:change={(e) =>
																handleLevelChange(level.value, e.currentTarget.checked)}
															class="h-4 w-4 rounded border-gray-300 bg-gray-100 text-blue-600 dark:border-gray-600 dark:bg-gray-700"
														/>
														<span class="ml-2">{level.label}</span>
													</label>
												</li>
											{/each}
										</ul>
									</div>
								{/if}
							</div>
						</div>

						<!-- Category Multi-select -->
						<div id="category-dropdown-container">
							<label
								for="filter-category"
								class="block text-sm font-medium text-gray-700 dark:text-gray-300">Category</label
							>
							<div class="relative">
								<button
									type="button"
									on:click={() => (showCategoryDropdown = !showCategoryDropdown)}
									class="mt-1 inline-flex w-full items-center justify-between rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-900 shadow-sm hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
								>
									{categoryButtonText}
									<svg class="ml-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 9l-7 7-7-7"
										/>
									</svg>
								</button>
								{#if showCategoryDropdown}
									<div
										class="absolute z-10 mt-1 w-full divide-y divide-gray-100 rounded-lg bg-white shadow dark:divide-gray-700 dark:bg-gray-800"
									>
										<ul class="space-y-1 p-3 text-sm text-gray-700 dark:text-gray-200">
											{#each logCategories as category}
												<li>
													<label class="flex items-center">
														<input
															type="checkbox"
															checked={selectedCategories.includes(category.value)}
															on:change={(e) =>
																handleCategoryChange(category.value, e.currentTarget.checked)}
															class="h-4 w-4 rounded border-gray-300 bg-gray-100 text-blue-600 dark:border-gray-600 dark:bg-gray-700"
														/>
														<span class="ml-2">{category.label}</span>
													</label>
												</li>
											{/each}
										</ul>
									</div>
								{/if}
							</div>
						</div>

						<!-- Time Range -->
						<div class="flex flex-col gap-4 md:flex-row">
							<div class="w-full">
								<label
									for="start-time"
									class="block text-sm font-medium text-gray-700 dark:text-gray-300"
									>Start Time</label
								>
								<input
									type="datetime-local"
									id="start-time"
									bind:value={startTime}
									on:change={handleFilterChange}
									class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
								/>
							</div>
							<div class="w-full">
								<label
									for="end-time"
									class="block text-sm font-medium text-gray-700 dark:text-gray-300">End Time</label
								>
								<input
									type="datetime-local"
									id="end-time"
									bind:value={endTime}
									on:change={handleFilterChange}
									class="mt-1 min-h-[2.5rem] w-full rounded-md border-gray-300 bg-white text-sm text-gray-900 shadow-sm dark:border-gray-600 dark:bg-gray-800 dark:text-white"
								/>
							</div>
						</div>
					</div>
				</div>

				<!-- Active Filters -->
				{#if activeFilters.length > 0}
					<div class="mt-6">
						<div class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
							Active Filters
						</div>
						<div class="flex flex-wrap items-center gap-2">
							{#each activeFilters as filter}
								<span
									class="inline-flex items-center rounded-sm bg-blue-100 px-3 py-1 text-sm font-medium text-blue-800 dark:bg-blue-900 dark:text-blue-100"
								>
									{filter.key}:{filter.value}
									<button
										type="button"
										on:click={() => removeFilter(filter)}
										class="ml-1 text-xs hover:text-blue-600 dark:hover:text-blue-300"
									>
										&times;
									</button>
								</span>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Buttons -->
				<div class="mt-6 flex justify-end gap-3">
					<button
						type="button"
						on:click={resetFilters}
						class="rounded border border-gray-300 bg-gray-50 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:hover:bg-gray-600"
					>
						Reset
					</button>
					<button
						type="submit"
						class="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
					>
						Search
					</button>
				</div>
			</div>
		</form>

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

		<!-- Results Table -->
		<div
			class="rounded-xl border border-gray-100 bg-white shadow-sm dark:border-gray-800 dark:bg-gray-900"
		>
			<div class="overflow-x-auto">
				<table class="min-w-full table-fixed text-left text-sm text-gray-700 dark:text-gray-200">
					<colgroup>
						<col style="width: 10%" />
						<col style="width: 8%" />
						<col style="width: 8%" />
						<col style="width: 8%" />
						<col style="width: 46%" />
						<col style="width: 10%" />
						<col style="width: 10%" />
					</colgroup>
					<thead
						class="bg-gray-50 text-xs text-gray-500 uppercase dark:bg-gray-800 dark:text-gray-300"
					>
						<tr>
							<th scope="col" class="px-4 py-3">Time</th>
							<th scope="col" class="px-4 py-3">Level</th>
							<th scope="col" class="px-4 py-3">Source</th>
							<th scope="col" class="px-4 py-3">Endpoint</th>
							<th scope="col" class="px-4 py-3">Message</th>
							<th scope="col" class="px-4 py-3">User</th>
							<th scope="col" class="px-4 py-3">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#if loading}
							<tr>
								<td colspan="7" class="py-8 text-center">
									<div
										class="inline-block h-8 w-8 animate-spin rounded-full border-b-2 border-gray-900 dark:border-white"
									></div>
									<div class="mt-2 text-sm text-gray-600 dark:text-gray-400">Loading logs...</div>
								</td>
							</tr>
						{:else if logs.length === 0}
							<tr>
								<td colspan="7" class="py-4 text-center">
									<div class="text-gray-500 dark:text-gray-400">No logs found</div>
								</td>
							</tr>
						{:else}
							{#each logs as log, index (log.id || index)}
								<tr
									class="border-b border-gray-100 hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800"
								>
									<!-- Time -->
									<td class="px-4 py-3 text-xs">
										{formatDate(log.timestamp)}
									</td>

									<!-- Level -->
									<td class="px-4 py-3">
										<span
											class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getLogLevelColor(
												log.level
											)}"
										>
											<i class="{getLogLevelIcon(log.level)} mr-1"></i>
											{log.level}
										</span>
									</td>

									<!-- Source -->
									<td class="px-4 py-3 font-mono text-xs">
										{log.source || '-'}
									</td>

									<!-- Endpoint -->
									<td class="px-4 py-3 text-xs">
										{log.endpoint_name || '-'}
									</td>

									<!-- Message -->
									<td class="px-4 py-3">
										<div class="max-w-md font-mono text-xs break-words whitespace-pre-wrap">
											{log.message}
										</div>
									</td>

									<!-- User -->
									<td class="px-4 py-3 text-xs">
										{log.user || '-'}
									</td>

									<!-- Actions -->
									<td class="px-4 py-3">
										<div class="flex gap-2">
											<button
												type="button"
												on:click={() => toggleLogExpanded(log.id || String(index))}
												class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
												title="Toggle details"
											>
												<i
													class="fas fa-{expandedLogs.has(log.id || String(index))
														? 'chevron-up'
														: 'chevron-down'}"
												></i>
											</button>
											<button
												type="button"
												on:click={() => navigator.clipboard.writeText(log.message)}
												class="text-gray-600 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-300"
												title="Copy message"
											>
												📋
											</button>
										</div>
									</td>
								</tr>

								<!-- Expanded details row -->
								{#if expandedLogs.has(log.id || String(index))}
									<tr class="bg-gray-50 dark:bg-gray-800">
										<td colspan="7" class="px-4 py-3">
											<div class="space-y-2">
												<div class="text-sm font-medium text-gray-900 dark:text-white">
													Log Details
												</div>

												<div class="grid grid-cols-2 gap-4 text-xs">
													<div>
														<span class="font-medium text-gray-600 dark:text-gray-400">ID:</span>
														<span class="ml-2 font-mono">{log.id || 'N/A'}</span>
													</div>
													<div>
														<span class="font-medium text-gray-600 dark:text-gray-400"
															>Timestamp:</span
														>
														<span class="ml-2">{log.timestamp}</span>
													</div>
													<div>
														<span class="font-medium text-gray-600 dark:text-gray-400">Level:</span>
														<span class="ml-2">{log.level}</span>
													</div>
													<div>
														<span class="font-medium text-gray-600 dark:text-gray-400">Source:</span
														>
														<span class="ml-2 font-mono">{log.source || 'N/A'}</span>
													</div>
													{#if log.container_name}
														<div>
															<span class="font-medium text-gray-600 dark:text-gray-400"
																>Container:</span
															>
															<span class="ml-2 font-mono">{log.container_name}</span>
														</div>
													{/if}
													{#if log.app_name}
														<div>
															<span class="font-medium text-gray-600 dark:text-gray-400">App:</span>
															<span class="ml-2">{log.app_name}</span>
														</div>
													{/if}
												</div>

												{#if log.metadata && Object.keys(log.metadata).length > 0}
													<div class="mt-3">
														<div class="mb-2 text-sm font-medium text-gray-900 dark:text-white">
															Metadata
														</div>
														<pre
															class="rounded bg-gray-100 p-3 font-mono text-xs whitespace-pre-wrap text-gray-800 dark:bg-gray-700 dark:text-gray-100">{JSON.stringify(
																log.metadata,
																null,
																2
															)}</pre>
													</div>
												{/if}
											</div>
										</td>
									</tr>
								{/if}
							{/each}
						{/if}
					</tbody>
				</table>

				<!-- Pagination Controls -->
				<div
					class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6 dark:border-gray-700 dark:bg-gray-900"
				>
					<div class="flex flex-1 justify-between sm:hidden">
						<button
							on:click={handlePrevPage}
							disabled={cursorStack.length === 0}
							class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
						>
							Previous
						</button>
						<button
							on:click={handleNextPage}
							disabled={!hasMore}
							class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200 dark:hover:bg-gray-700"
						>
							Next
						</button>
					</div>
					<div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
						<div class="flex items-center gap-4">
							<div class="text-sm text-gray-700 dark:text-gray-300">
								Showing <span class="font-medium">{logCount}</span> results
							</div>
							{#if firstVisibleTime}
								<div class="text-sm text-gray-700 dark:text-gray-300">
									from <span class="font-medium">{new Date(firstVisibleTime).toLocaleString()}</span
									>
								</div>
							{/if}
						</div>
						<div>
							<nav class="isolate inline-flex -space-x-px rounded-md shadow-sm">
								<button
									on:click={handlePrevPage}
									disabled={cursorStack.length === 0}
									class="relative inline-flex items-center gap-1 rounded-l-md px-3 py-2 text-gray-400 ring-1 ring-gray-300 ring-inset hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-500 dark:ring-gray-600 dark:hover:bg-gray-700"
								>
									<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
										<path
											fill-rule="evenodd"
											d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
											clip-rule="evenodd"
										/>
									</svg>
									Previous
								</button>
								<button
									on:click={handleNextPage}
									disabled={!hasMore}
									class="relative inline-flex items-center gap-1 rounded-r-md px-3 py-2 text-gray-400 ring-1 ring-gray-300 ring-inset hover:bg-gray-50 focus:z-20 focus:outline-offset-0 disabled:cursor-not-allowed disabled:opacity-50 dark:text-gray-500 dark:ring-gray-600 dark:hover:bg-gray-700"
								>
									Next
									<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
										<path
											fill-rule="evenodd"
											d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
											clip-rule="evenodd"
										/>
									</svg>
								</button>
							</nav>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</PermissionGuard>
