<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { Input, Label, Select, Toggle } from 'flowbite-svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { fade } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';
	import type { Widget, WidgetConfig, QuickLink } from '$lib/types/dashboard';
	import { api } from '$lib/api';

	// Props
	export let isOpen: boolean = false;
	export let widget: Widget | null = null;

	// Debug logging for modal state
	$: {
		console.log('WidgetConfigModal isOpen changed:', isOpen);
		console.log('WidgetConfigModal widget:', widget);
	}

	const dispatch = createEventDispatcher<{
		save: { config: WidgetConfig };
		cancel: void;
	}>();

	// Enhanced configuration state
	let localConfig: WidgetConfig = {};
	let errors: Record<string, string> = {};
	let warningThreshold: number | undefined;
	let criticalThreshold: number | undefined;

	// Chart configuration state
	interface MetricOption {
		label: string;
		namespace: string;
		subnamespace: string;
		name: string;
	}

	let dataSource = 'metrics'; // metrics, logs, events, alerts
	let namespaces: string[] = [];
	let subnamespaces: string[] = [];
	let metrics: string[] = [];
	let selectedNamespace = '';
	let selectedSubnamespace = '';
	let selectedMetric = '';
	let selectedMetrics: MetricOption[] = [];
	let availableTags: Record<string, Set<string>> = {};
	let selectedTags: Record<string, string> = {};
	let chartType = 'area';
	let timeFrame = '1h';
	let stepInterval = '1m';
	let aggregateBy = '';
	let groupBy = '';
	let isStacked = false;

	// Loading states
	let loadingNamespaces = false;
	let loadingSubnamespaces = false;
	let loadingMetrics = false;
	let loadingTags = false;

	// Chart type options from ApexCharts
	const chartTypes = [
		{ value: 'area', name: 'Area Chart' },
		{ value: 'line', name: 'Line Chart' },
		{ value: 'bar', name: 'Bar Chart' },
		{ value: 'column', name: 'Column Chart' },
		{ value: 'donut', name: 'Donut Chart' },
		{ value: 'pie', name: 'Pie Chart' },
		{ value: 'radialBar', name: 'Radial Bar' },
		{ value: 'scatter', name: 'Scatter Plot' },
		{ value: 'heatmap', name: 'Heatmap' },
		{ value: 'candlestick', name: 'Candlestick' }
	];

	const timeRangeOptions = [
		{ value: '5m', name: '5 minutes' },
		{ value: '10m', name: '10 minutes' },
		{ value: '15m', name: '15 minutes' },
		{ value: '30m', name: '30 minutes' },
		{ value: '1h', name: '1 hour' },
		{ value: '3h', name: '3 hours' },
		{ value: '6h', name: '6 hours' },
		{ value: '12h', name: '12 hours' },
		{ value: '1d', name: '1 day' },
		{ value: '3d', name: '3 days' },
		{ value: '1w', name: '1 week' }
	];

	const stepIntervalOptions = [
		{ value: '5s', name: '5 seconds' },
		{ value: '15s', name: '15 seconds' },
		{ value: '30s', name: '30 seconds' },
		{ value: '1m', name: '1 minute' },
		{ value: '2m', name: '2 minutes' },
		{ value: '5m', name: '5 minutes' },
		{ value: '10m', name: '10 minutes' },
		{ value: '15m', name: '15 minutes' },
		{ value: '30m', name: '30 minutes' },
		{ value: '1h', name: '1 hour' }
	];

	const aggregationOptions = [
		{ value: '', name: 'None' },
		{ value: 'avg', name: 'Average' },
		{ value: 'sum', name: 'Sum' },
		{ value: 'min', name: 'Minimum' },
		{ value: 'max', name: 'Maximum' },
		{ value: 'count', name: 'Count' },
		{ value: 'rate', name: 'Rate' }
	];

	const dataSourceOptions = [
		{ value: 'metrics', name: 'Metrics' },
		{ value: 'logs', name: 'Logs' },
		{ value: 'events', name: 'Events' },
		{ value: 'alerts', name: 'Alerts' }
	];

	const refreshIntervalOptions = [
		{ value: 5000, name: '5 seconds' },
		{ value: 10000, name: '10 seconds' },
		{ value: 30000, name: '30 seconds' },
		{ value: 60000, name: '1 minute' },
		{ value: 300000, name: '5 minutes' },
		{ value: 600000, name: '10 minutes' },
		{ value: 1800000, name: '30 minutes' },
		{ value: 3600000, name: '1 hour' }
	];

	// Watch for widget changes to reset form
	$: if (widget && isOpen) {
		resetForm();
	}

	onMount(() => {
		if (isOpen && dataSource === 'metrics') {
			loadNamespaces();
		}
	});

	// Watch for data source changes
	$: if (dataSource === 'metrics' && isOpen) {
		loadNamespaces();
	}

	// Watch for namespace changes
	$: if (selectedNamespace && isOpen) {
		loadSubnamespaces();
	}

	// Watch for subnamespace changes
	$: if (selectedSubnamespace && isOpen) {
		loadMetrics();
	}

	// Watch for metric changes
	$: if (selectedMetric && isOpen) {
		loadMetricTags();
	}

	function resetForm() {
		localConfig = { ...widget?.config || {} };
		warningThreshold = widget?.config?.threshold?.warning;
		criticalThreshold = widget?.config?.threshold?.critical;
		errors = {};

		// Reset chart configuration to defaults or saved values
		const chartConfig = localConfig.chartConfig || {};
		dataSource = chartConfig.dataSource || 'metrics';
		selectedNamespace = chartConfig.namespace || '';
		selectedSubnamespace = chartConfig.subnamespace || '';
		selectedMetric = chartConfig.metric || '';
		selectedMetrics = chartConfig.selectedMetrics || [];
		selectedTags = chartConfig.tags || {};
		chartType = chartConfig.chartType || 'area';
		timeFrame = chartConfig.timeFrame || '1h';
		stepInterval = chartConfig.stepInterval || '1m';
		aggregateBy = chartConfig.aggregateBy || '';
		groupBy = chartConfig.groupBy || '';
		isStacked = chartConfig.isStacked || false;
	}

	async function loadNamespaces() {
		if (loadingNamespaces) return;
		loadingNamespaces = true;
		console.log('DEBUG: Loading namespaces from /api/v1/');
		try {
			namespaces = await api.request('/api/v1/') as string[];
			console.log('DEBUG: Namespaces loaded:', namespaces);
		} catch (err) {
			console.error('Failed to load namespaces:', err);
			errors.namespaces = 'Failed to load namespaces';
		} finally {
			loadingNamespaces = false;
		}
	}

	async function loadSubnamespaces() {
		if (!selectedNamespace || loadingSubnamespaces) return;
		loadingSubnamespaces = true;
		subnamespaces = [];
		selectedSubnamespace = '';
		try {
			subnamespaces = await api.request(`/api/v1/${selectedNamespace}`) as string[];
		} catch (err) {
			console.error('Failed to load subnamespaces:', err);
			errors.subnamespaces = 'Failed to load subnamespaces';
		} finally {
			loadingSubnamespaces = false;
		}
	}

	async function loadMetrics() {
		if (!selectedSubnamespace || loadingMetrics) return;
		loadingMetrics = true;
		metrics = [];
		selectedMetric = '';
		try {
			metrics = await api.request(`/api/v1/${selectedNamespace}/${selectedSubnamespace}`) as string[];
		} catch (err) {
			console.error('Failed to load metrics:', err);
			errors.metrics = 'Failed to load metrics';
		} finally {
			loadingMetrics = false;
		}
	}

	async function loadMetricTags() {
		if (!selectedMetric || loadingTags) return;
		loadingTags = true;
		try {
			// Load tag suggestions similar to metrics page
			const [hostsRes, containersRes] = await Promise.all([
				api.endpoints.getHosts(),
				api.endpoints.getContainers()
			]);

			const hosts = (hostsRes as any).data || [];
			const containers = (containersRes as any).data || [];
			const allEndpoints = [...hosts, ...containers];

			availableTags = {};
			const blacklist = new Set(['agent_start_time', '_cmdline', '_uid', '_exe']);

			for (const ep of allEndpoints) {
				const tags: Record<string, string> = { ...ep.labels };
				if (ep.Hostname) tags['hostname'] = ep.Hostname;
				if (ep.Name) tags['container_name'] = ep.Name;
				if (ep.container_name) tags['container_name'] = ep.container_name;
				if (ep.EndpointID) tags['endpoint_id'] = ep.EndpointID;
				if (ep.ImageName) tags['image_name'] = ep.ImageName;

				for (const [rawKey, val] of Object.entries(tags)) {
					const k = rawKey.toLowerCase();
					if (!val || blacklist.has(k)) continue;
					if (!availableTags[k]) availableTags[k] = new Set();
					availableTags[k].add(val);
				}
			}
		} catch (err) {
			console.error('Failed to load tags:', err);
			errors.tags = 'Failed to load tags';
		} finally {
			loadingTags = false;
		}
	}

	function addMetric() {
		if (!selectedMetric || !selectedNamespace || !selectedSubnamespace) return;
		
		const metricOption: MetricOption = {
			label: `${selectedNamespace}.${selectedSubnamespace}.${selectedMetric}`,
			namespace: selectedNamespace,
			subnamespace: selectedSubnamespace,
			name: selectedMetric
		};

		// Check if metric already exists
		const exists = selectedMetrics.some(m => m.label === metricOption.label);
		if (!exists) {
			selectedMetrics = [...selectedMetrics, metricOption];
		}

		// Reset selections
		selectedMetric = '';
	}

	function removeMetric(index: number) {
		selectedMetrics = selectedMetrics.filter((_, i) => i !== index);
	}

	function addTag(key: string, value: string) {
		selectedTags = { ...selectedTags, [key]: value };
	}

	function removeTag(key: string) {
		const newTags = { ...selectedTags };
		delete newTags[key];
		selectedTags = newTags;
	}

	// Update thresholds in config when local values change
	$: if (warningThreshold !== undefined || criticalThreshold !== undefined) {
		if (!localConfig.threshold) {
			localConfig.threshold = { warning: 0, critical: 0 };
		}
		if (warningThreshold !== undefined) {
			localConfig.threshold.warning = warningThreshold;
		}
		if (criticalThreshold !== undefined) {
			localConfig.threshold.critical = criticalThreshold;
		}
	}

	function validateConfig(): boolean {
		errors = {};
		
		// Validate chart configuration for chart widgets
		if (widget?.type.includes('chart')) {
			if (dataSource === 'metrics' && selectedMetrics.length === 0) {
				errors.metrics = 'Please select at least one metric';
			}
		}

		// Validate metric thresholds
		if (warningThreshold !== undefined && criticalThreshold !== undefined) {
			if (warningThreshold >= criticalThreshold) {
				errors.threshold = 'Warning threshold must be less than critical threshold';
			}
		}

		// Validate refresh interval
		if (localConfig.refreshInterval && localConfig.refreshInterval < 5000) {
			errors.refreshInterval = 'Refresh interval must be at least 5 seconds';
		}

		return Object.keys(errors).length === 0;
	}

	function handleSave() {
		if (!validateConfig()) {
			return;
		}

		// Include chart configuration in the saved config
		localConfig.chartConfig = {
			dataSource,
			namespace: selectedNamespace,
			subnamespace: selectedSubnamespace,
			metric: selectedMetric,
			selectedMetrics,
			tags: selectedTags,
			chartType,
			timeFrame,
			stepInterval,
			aggregateBy,
			groupBy,
			isStacked
		};

		dispatch('save', { config: localConfig });
		isOpen = false;
	}

	function handleCancel() {
		isOpen = false;
		dispatch('cancel');
	}

	function addQuickLink() {
		if (!localConfig.links) {
			localConfig.links = [];
		}
		localConfig.links = [
			...localConfig.links,
			{
				id: crypto.randomUUID(),
				title: '',
				url: '',
				icon: '',
				description: ''
			}
		];
	}

	function removeQuickLink(index: number) {
		if (localConfig.links) {
			localConfig.links = localConfig.links.filter((_, i) => i !== index);
		}
	}
</script>


<Modal 
	bind:show={isOpen} 
	size="xl" 
	title="Configure Widget"
	on:close={() => isOpen = false}
>
	<div style="background-color: red; color: white; padding: 20px; font-size: 24px;">
		DEBUG: Modal is open! Widget: {widget?.title || 'No widget'}
		<br>
		isOpen: {isOpen}
	</div>
	
	{#if widget}
		<div class="space-y-6 max-h-96 overflow-y-auto" transition:fade={{ duration: 300, easing: quintOut }}>
			<!-- Common Settings -->
			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">General Settings</h3>
				
				<div>
					<Label for="showTitle" class="mb-2">Display Options</Label>
					<Toggle bind:checked={localConfig.showTitle}>Show widget title</Toggle>
				</div>

				<div>
					<Label for="refreshInterval" class="mb-2">Refresh Interval</Label>
					<Select
						id="refreshInterval"
						bind:value={localConfig.refreshInterval}
						items={refreshIntervalOptions}
						placeholder="Select refresh interval"
					/>
					{#if errors.refreshInterval}
						<p class="mt-1 text-sm text-red-600">{errors.refreshInterval}</p>
					{/if}
				</div>
			</div>

			<!-- Chart Configuration for Chart Widgets -->
			{#if widget && widget.type.includes('chart')}
				<div class="space-y-4">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Chart Configuration</h3>
					
					<!-- Data Source Selection -->
					<div>
						<Label for="dataSource" class="mb-2">Data Source</Label>
						<Select
							id="dataSource"
							bind:value={dataSource}
							items={dataSourceOptions}
							placeholder="Select data source"
						/>
					</div>

					{#if dataSource === 'metrics'}
						<!-- Hierarchical Metric Selection -->
						<div class="space-y-3">
							<div>
								<Label for="namespace" class="mb-2">Namespace</Label>
								<Select
									id="namespace"
									bind:value={selectedNamespace}
									items={namespaces.map(n => ({ value: n, name: n }))}
									placeholder={loadingNamespaces ? "Loading..." : "Select namespace"}
									disabled={loadingNamespaces}
								/>
								{#if errors.namespaces}
									<p class="mt-1 text-sm text-red-600">{errors.namespaces}</p>
								{/if}
							</div>

							{#if selectedNamespace}
								<div>
									<Label for="subnamespace" class="mb-2">Sub-namespace</Label>
									<Select
										id="subnamespace"
										bind:value={selectedSubnamespace}
										items={subnamespaces.map(s => ({ value: s, name: s }))}
										placeholder={loadingSubnamespaces ? "Loading..." : "Select sub-namespace"}
										disabled={loadingSubnamespaces}
									/>
									{#if errors.subnamespaces}
										<p class="mt-1 text-sm text-red-600">{errors.subnamespaces}</p>
									{/if}
								</div>
							{/if}

							{#if selectedSubnamespace}
								<div>
									<Label for="metric" class="mb-2">Metric</Label>
									<div class="flex gap-2">
										<Select
											id="metric"
											bind:value={selectedMetric}
											items={metrics.map(m => ({ value: m, name: m }))}
											placeholder={loadingMetrics ? "Loading..." : "Select metric"}
											disabled={loadingMetrics}
											class="flex-1"
										/>
										<button 
											on:click={addMetric} 
											disabled={!selectedMetric} 
											class="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
										>
											Add
										</button>
									</div>
									{#if errors.metrics}
										<p class="mt-1 text-sm text-red-600">{errors.metrics}</p>
									{/if}
								</div>
							{/if}

							<!-- Selected Metrics -->
							{#if selectedMetrics.length > 0}
								<div>
									<Label class="mb-2">Selected Metrics</Label>
									<div class="space-y-2">
										{#each selectedMetrics as metric, index}
											<div class="flex items-center justify-between rounded-md bg-gray-100 p-2 dark:bg-gray-800">
												<span class="text-sm">{metric.label}</span>
												<button 
													on:click={() => removeMetric(index)} 
													class="px-2 py-1 text-xs font-medium text-white bg-red-600 rounded hover:bg-red-700 focus:ring-2 focus:ring-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
												>
													Remove
												</button>
											</div>
										{/each}
									</div>
								</div>
							{/if}

							<!-- Tags/Labels -->
							{#if Object.keys(availableTags).length > 0}
								<div>
									<Label class="mb-2">Tags/Labels (Filters)</Label>
									<div class="space-y-2">
										{#each Object.entries(availableTags) as [tagKey, tagValues]}
											<div class="flex gap-2 items-center">
												<span class="min-w-0 flex-1 text-sm font-medium">{tagKey}:</span>
												<select
													value={selectedTags[tagKey] || ''}
													on:change={(e) => {
														const target = e.target as HTMLSelectElement;
														if (target.value) {
															addTag(tagKey, target.value);
														} else {
															removeTag(tagKey);
														}
													}}
													class="min-w-0 flex-1 block rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
												>
													<option value="">Select value</option>
													{#each Array.from(tagValues) as value}
														<option value={value}>{value}</option>
													{/each}
												</select>
											</div>
										{/each}
									</div>

									<!-- Selected Tags Display -->
									{#if Object.keys(selectedTags).length > 0}
										<div class="mt-2">
											<Label class="mb-1">Applied Filters:</Label>
											<div class="flex flex-wrap gap-1">
												{#each Object.entries(selectedTags) as [key, value]}
													<span class="inline-flex items-center px-2 py-1 text-xs font-medium text-blue-800 bg-blue-100 rounded-full dark:bg-blue-900 dark:text-blue-300">
														{key}:{value}
														<button 
															on:click={() => removeTag(key)}
															class="ml-1 text-blue-400 hover:text-blue-600 focus:outline-none"
														>
															<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
															</svg>
														</button>
													</span>
												{/each}
											</div>
										</div>
									{/if}
								</div>
							{/if}
						</div>
					{/if}

					<!-- Chart Type Selection -->
					<div>
						<Label for="chartType" class="mb-2">Chart Type</Label>
						<Select
							id="chartType"
							bind:value={chartType}
							items={chartTypes}
							placeholder="Select chart type"
						/>
					</div>

					<!-- Time Configuration -->
					<div class="grid grid-cols-2 gap-4">
						<div>
							<Label for="timeFrame" class="mb-2">Time Frame</Label>
							<Select
								id="timeFrame"
								bind:value={timeFrame}
								items={timeRangeOptions}
								placeholder="Select time range"
							/>
						</div>
						<div>
							<Label for="stepInterval" class="mb-2">Step Interval</Label>
							<Select
								id="stepInterval"
								bind:value={stepInterval}
								items={stepIntervalOptions}
								placeholder="Select step interval"
							/>
						</div>
					</div>

					<!-- Aggregation and Grouping -->
					<div class="grid grid-cols-2 gap-4">
						<div>
							<Label for="aggregateBy" class="mb-2">Aggregate By</Label>
							<Select
								id="aggregateBy"
								bind:value={aggregateBy}
								items={aggregationOptions}
								placeholder="Select aggregation"
							/>
						</div>
						<div>
							<Label for="groupBy" class="mb-2">Group By</Label>
							<Input
								id="groupBy"
								bind:value={groupBy}
								placeholder="e.g., hostname, container_name"
							/>
						</div>
					</div>

					<!-- Chart Options -->
					<div>
						<Label for="chartOptions" class="mb-2">Chart Options</Label>
						<Toggle bind:checked={isStacked}>Stacked Chart</Toggle>
					</div>
				</div>
			{/if}

			<!-- Legacy Widget-specific settings -->
			{#if widget && (widget.type === 'metric' || widget.type === 'metric-card')}
				<div class="space-y-4">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Metric Settings</h3>
					
					<div>
						<Label for="metricType" class="mb-2">Metric Type</Label>
						<Input
							id="metricType"
							bind:value={localConfig.metricType}
							placeholder="e.g., cpu_usage, memory_usage"
						/>
					</div>

					<div>
						<Label for="unit" class="mb-2">Unit</Label>
						<Input
							id="unit"
							bind:value={localConfig.unit}
							placeholder="e.g., %, MB, seconds"
						/>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<Label for="warningThreshold" class="mb-2">Warning Threshold</Label>
							<Input
								id="warningThreshold"
								type="number"
								bind:value={warningThreshold}
								placeholder="Warning level"
							/>
						</div>
						<div>
							<Label for="criticalThreshold" class="mb-2">Critical Threshold</Label>
							<Input
								id="criticalThreshold"
								type="number"
								bind:value={criticalThreshold}
								placeholder="Critical level"
							/>
						</div>
					</div>
					{#if errors.threshold}
						<p class="mt-1 text-sm text-red-600">{errors.threshold}</p>
					{/if}
				</div>
			{/if}

			{#if widget && widget.type === 'quick_links'}
				<div class="space-y-4">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Quick Links</h3>
					
					{#if localConfig.links && localConfig.links.length > 0}
						{#each localConfig.links as link, index (link.id)}
							<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
								<div class="mb-2 flex items-center justify-between">
									<span class="text-sm font-medium">Link {index + 1}</span>
									<button 
										on:click={() => removeQuickLink(index)} 
										class="px-2 py-1 text-xs font-medium text-white bg-red-600 rounded hover:bg-red-700 focus:ring-2 focus:ring-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
									>
										Remove
									</button>
								</div>
								<div class="grid grid-cols-2 gap-2">
									<Input
										bind:value={link.title}
										placeholder="Link title"
									/>
									<Input
										bind:value={link.url}
										placeholder="URL"
									/>
									<Input
										bind:value={link.icon}
										placeholder="Icon (optional)"
									/>
									<Input
										bind:value={link.description}
										placeholder="Description (optional)"
									/>
								</div>
							</div>
						{/each}
					{/if}
					
					<button 
						on:click={addQuickLink} 
						class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
					>
						Add Quick Link
					</button>
				</div>
			{/if}
		</div>
	{/if}

	<svelte:fragment slot="footer">
		<div class="flex justify-end space-x-2">
			<button 
				on:click={handleCancel} 
				class="px-4 py-2 text-sm font-medium text-gray-900 bg-white border border-gray-200 rounded-lg hover:bg-gray-100 hover:text-blue-700 focus:ring-4 focus:ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700"
			>
				Cancel
			</button>
			<button 
				on:click={handleSave} 
				class="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
			>
				Save Changes
			</button>
		</div>
	</svelte:fragment>
</Modal>