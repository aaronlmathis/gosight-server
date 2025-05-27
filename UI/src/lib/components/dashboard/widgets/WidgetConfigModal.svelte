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
		console.log('WidgetConfigModal dataSource:', dataSource);
		console.log('WidgetConfigModal loadingNamespaces:', loadingNamespaces);
		console.log('WidgetConfigModal namespaces:', namespaces);
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
		tags: Record<string, string>; // Required tags for this metric
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
	let pendingMetricTags: Record<string, string> = {}; // Tags being configured for the current metric addition
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
		// Load namespaces when modal opens with metrics data source
		if (dataSource === 'metrics') {
			loadNamespaces();
		}
	}

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

	// Watch for metric changes - load tags when selectedMetric changes or when selectedMetrics array changes
	$: if (selectedMetric && isOpen) {
		loadMetricTags();
	}

	// Watch for selectedMetrics changes to load tags for the first metric
	$: if (selectedMetrics.length > 0 && isOpen) {
		loadTagsForSelectedMetrics();
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
		selectedMetrics = (chartConfig.selectedMetrics || []).map((metric: any) => ({
			...metric,
			tags: metric.tags || {} // Ensure tags property exists for backward compatibility
		}));
		chartType = chartConfig.chartType || 'area';
		timeFrame = chartConfig.timeFrame || '1h';
		stepInterval = chartConfig.stepInterval || '1m';
		aggregateBy = chartConfig.aggregateBy || '';
		groupBy = chartConfig.groupBy || '';
		isStacked = chartConfig.isStacked || false;
		
		// Reset arrays and loading states
		namespaces = [];
		subnamespaces = [];
		metrics = [];
		availableTags = {};
		pendingMetricTags = {};
		loadingNamespaces = false;
		loadingSubnamespaces = false;
		loadingMetrics = false;
		loadingTags = false;
	}

	async function loadNamespaces() {
		if (loadingNamespaces) {
			console.log('DEBUG: loadNamespaces already in progress, skipping');
			return;
		}
		loadingNamespaces = true;
		console.log('DEBUG: Starting loadNamespaces - Loading namespaces via api.metrics.getNamespaces()');
		try {
			const response = await api.metrics.getNamespaces();
			console.log('DEBUG: Raw namespaces response:', response);
			
			if (Array.isArray(response)) {
				namespaces = response;
			} else {
				console.warn('DEBUG: Response is not an array:', response);
				namespaces = [];
			}
			console.log('DEBUG: Namespaces loaded successfully:', namespaces);
			// Clear any previous errors
			if (errors.namespaces) {
				delete errors.namespaces;
				errors = { ...errors };
			}
		} catch (err) {
			console.error('Failed to load namespaces:', err);
			namespaces = [];
			errors.namespaces = 'Failed to load namespaces';
			errors = { ...errors };
		} finally {
			loadingNamespaces = false;
			console.log('DEBUG: loadNamespaces completed, loadingNamespaces:', loadingNamespaces);
		}
	}

	async function loadSubnamespaces() {
		if (!selectedNamespace || loadingSubnamespaces) return;
		loadingSubnamespaces = true;
		subnamespaces = [];
		selectedSubnamespace = '';
		console.log('DEBUG: Loading subnamespaces via api.metrics.getSubNamespaces(' + selectedNamespace + ')');
		try {
			subnamespaces = await api.metrics.getSubNamespaces(selectedNamespace) as string[];
			console.log('DEBUG: Subnamespaces loaded:', subnamespaces);
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
		// Clear pending tags when changing metrics selection
		pendingMetricTags = {};
		availableTags = {};
		console.log('DEBUG: Loading metrics via api.metrics.getMetricNames(' + selectedNamespace + ', ' + selectedSubnamespace + ')');
		try {
			metrics = await api.metrics.getMetricNames(selectedNamespace, selectedSubnamespace) as string[];
			console.log('DEBUG: Metrics loaded:', metrics);
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
		console.log('DEBUG: Loading metric dimensions via api.metrics.getMetricDimensions(' + selectedNamespace + ', ' + selectedSubnamespace + ', ' + selectedMetric + ')');
		try {
			// Load metric dimensions from the correct endpoint
			const dimensionsResponse = await api.metrics.getMetricDimensions(selectedNamespace, selectedSubnamespace, selectedMetric);
			console.log('DEBUG: Dimensions response:', dimensionsResponse);
			
			// Handle the case where API returns an array of dimension keys
			availableTags = {};
			if (Array.isArray(dimensionsResponse)) {
				// API returns array of tag keys, we need to load values for each key
				console.log('DEBUG: API returned array of tag keys:', dimensionsResponse);
				
				// For now, we'll need to get tag values from endpoints since the API only returns keys
				try {
					const [hostsRes, containersRes] = await Promise.all([
						api.endpoints.getHosts(),
						api.endpoints.getContainers()
					]);

					const hosts = (hostsRes as any).data || hostsRes || [];
					const containers = (containersRes as any).data || containersRes || [];
					const allEndpoints = [...hosts, ...containers];

					// Initialize tag structure for the keys returned by the API
					const tagKeysToLoad = new Set(dimensionsResponse.filter(key => 
						!['agent_start_time', '_cmdline', '_uid', '_exe'].includes(key)
					));

					for (const ep of allEndpoints) {
						const tags: Record<string, string> = { ...ep.labels };
						if (ep.Hostname) tags['hostname'] = ep.Hostname;
						if (ep.Name) tags['container_name'] = ep.Name;
						if (ep.container_name) tags['container_name'] = ep.container_name;
						if (ep.EndpointID) tags['endpoint_id'] = ep.EndpointID;
						if (ep.ImageName) tags['image_name'] = ep.ImageName;
						if (ep.host_id) tags['host_id'] = ep.host_id;
						if (ep.agent_id) tags['agent_id'] = ep.agent_id;
						if (ep.environment) tags['environment'] = ep.environment;
						if (ep.department) tags['department'] = ep.department;
						if (ep.cost_center) tags['cost_center'] = ep.cost_center;

						for (const [rawKey, val] of Object.entries(tags)) {
							const k = rawKey.toLowerCase();
							if (!val || !tagKeysToLoad.has(k)) continue;
							if (!availableTags[k]) availableTags[k] = new Set();
							availableTags[k].add(String(val));
						}
					}
				} catch (endpointErr) {
					console.error('Failed to load endpoint data for tag values:', endpointErr);
				}
			} else if (dimensionsResponse && typeof dimensionsResponse === 'object') {
				// Handle object response format (key -> values)
				for (const [key, values] of Object.entries(dimensionsResponse)) {
					if (Array.isArray(values)) {
						availableTags[key] = new Set(values as string[]);
					}
				}
			}
			console.log('DEBUG: Available tags processed:', availableTags);
		} catch (err) {
			console.error('Failed to load metric dimensions:', err);
			errors.tags = 'Failed to load metric dimensions';
			
			// Fallback to the old method if the dimensions endpoint fails
			console.log('DEBUG: Falling back to hosts/containers method');
			try {
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
				console.log('DEBUG: Fallback tags loaded:', availableTags);
			} catch (fallbackErr) {
				console.error('Fallback method also failed:', fallbackErr);
			}
		} finally {
			loadingTags = false;
		}
	}

	async function loadTagsForSelectedMetrics() {
		if (selectedMetrics.length === 0 || loadingTags) return;
		
		// Load tags for the first selected metric to get the available dimensions
		const firstMetric = selectedMetrics[0];
		if (!firstMetric) return;
		
		loadingTags = true;
		console.log('DEBUG: Loading tags for selected metrics, using first metric:', firstMetric.label);
		
		try {
			const dimensionsResponse = await api.metrics.getMetricDimensions(
				firstMetric.namespace, 
				firstMetric.subnamespace, 
				firstMetric.name
			);
			console.log('DEBUG: Dimensions response for selected metrics:', dimensionsResponse);
			
			// Convert the dimensions response to the expected format
			availableTags = {};
			if (dimensionsResponse && typeof dimensionsResponse === 'object') {
				for (const [key, values] of Object.entries(dimensionsResponse)) {
					if (Array.isArray(values)) {
						availableTags[key] = new Set(values as string[]);
					}
				}
			}
			console.log('DEBUG: Available tags for selected metrics:', availableTags);
		} catch (err) {
			console.error('Failed to load metric dimensions for selected metrics:', err);
			// Don't set error here since this is a fallback attempt
		} finally {
			loadingTags = false;
		}
	}

	function addMetric() {
		if (!selectedMetric || !selectedNamespace || !selectedSubnamespace) return;
		
		// Require at least one tag to be selected
		if (Object.keys(pendingMetricTags).length === 0) {
			errors.tags = 'Please select at least one tag/filter for the metric to ensure proper data scoping';
			return;
		}
		
		const metricOption: MetricOption = {
			label: `${selectedNamespace}.${selectedSubnamespace}.${selectedMetric}`,
			namespace: selectedNamespace,
			subnamespace: selectedSubnamespace,
			name: selectedMetric,
			tags: { ...pendingMetricTags } // Include the required tags
		};

		// Check if metric already exists (with same tags)
		const exists = selectedMetrics.some(m => 
			m.label === metricOption.label && 
			JSON.stringify(m.tags) === JSON.stringify(metricOption.tags)
		);
		if (!exists) {
			selectedMetrics = [...selectedMetrics, metricOption];
		}

		// Reset selections and pending tags
		selectedMetric = '';
		pendingMetricTags = {};
		availableTags = {};
		// Clear the tags error if it was set
		if (errors.tags) {
			delete errors.tags;
			errors = { ...errors };
		}
	}

	function removeMetric(index: number) {
		selectedMetrics = selectedMetrics.filter((_, i) => i !== index);
	}

	function addPendingTag(key: string, value: string) {
		pendingMetricTags = { ...pendingMetricTags, [key]: value };
	}

	function removePendingTag(key: string) {
		const newTags = { ...pendingMetricTags };
		delete newTags[key];
		pendingMetricTags = newTags;
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
			selectedMetrics, // Now includes tags embedded in each metric
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
									disabled={loadingNamespaces || namespaces.length === 0}
								/>
								{#if loadingNamespaces}
									<p class="mt-1 text-sm text-blue-600">Loading namespaces...</p>
								{/if}
								{#if !loadingNamespaces && namespaces.length === 0}
									<p class="mt-1 text-sm text-orange-600">No namespaces available</p>
								{/if}
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
									<Select
										id="metric"
										bind:value={selectedMetric}
										items={metrics.map(m => ({ value: m, name: m }))}
										placeholder={loadingMetrics ? "Loading..." : "Select metric"}
										disabled={loadingMetrics}
									/>
									{#if errors.metrics}
										<p class="mt-1 text-sm text-red-600">{errors.metrics}</p>
									{/if}
								</div>
								
								<!-- Tag Selection for Current Metric (Required) -->
								{#if selectedMetric}
									<div class="border-l-4 border-blue-500 bg-blue-50 p-4 dark:bg-blue-900/20">
										<Label class="mb-2 text-blue-800 dark:text-blue-200">Required: Select Tags/Filters for {selectedMetric}</Label>
										<p class="mb-3 text-sm text-blue-700 dark:text-blue-300">
											You must select at least one tag to scope the metric data and enable meaningful comparisons.
										</p>
										
										{#if loadingTags}
											<p class="mt-1 text-sm text-blue-600">Loading available tags...</p>
										{:else if Object.keys(availableTags).length > 0}
											<div class="space-y-3">
												<!-- Available Tag Selection -->
												<div class="space-y-2">
													<Label class="text-sm font-medium text-gray-700 dark:text-gray-300">Available Tag Filters:</Label>
													{#each Object.entries(availableTags) as [tagKey, tagValues]}
														<div class="flex gap-2 items-center">
															<span class="min-w-0 w-1/3 text-sm font-medium text-gray-700 dark:text-gray-300">{tagKey}:</span>
															<select
																value=""
																on:change={(e) => {
																	const target = e.target as HTMLSelectElement;
																	if (target.value) {
																		addPendingTag(tagKey, target.value);
																		target.value = ''; // Reset selection
																	}
																}}
																class="min-w-0 flex-1 block rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
															>
																<option value="">Add {tagKey} filter...</option>
																{#each Array.from(tagValues) as value}
																	<option value={value}>{value}</option>
																{/each}
															</select>
														</div>
													{/each}
												</div>
												
												<!-- Show currently selected tags for this metric -->
												{#if Object.keys(pendingMetricTags).length > 0}
													<div class="mt-3">
														<Label class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">Selected Filters:</Label>
														<div class="flex flex-wrap gap-2">
															{#each Object.entries(pendingMetricTags) as [key, value]}
																<span class="inline-flex items-center px-3 py-1 text-sm font-medium text-blue-800 bg-blue-100 rounded-full dark:bg-blue-900 dark:text-blue-300">
																	<span class="font-semibold">{key}:</span>
																	<span class="ml-1">{value}</span>
																	<button 
																		on:click={() => removePendingTag(key)}
																		class="ml-2 text-blue-400 hover:text-blue-600 focus:outline-none"
																		title="Remove filter"
																	>
																		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
																			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
																		</svg>
																	</button>
																</span>
															{/each}
														</div>
													</div>
												{/if}
												
												<!-- Add Metric Button (only enabled when tags are selected) -->
												<div class="pt-3 border-t border-blue-200 dark:border-blue-700">
													<button 
														on:click={addMetric} 
														disabled={!selectedMetric || Object.keys(pendingMetricTags).length === 0} 
														class="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800 transition-colors"
													>
														{#if Object.keys(pendingMetricTags).length === 0}
															Select at least one tag filter to continue
														{:else}
															Add {selectedMetric} with {Object.keys(pendingMetricTags).length} filter{Object.keys(pendingMetricTags).length > 1 ? 's' : ''}
														{/if}
													</button>
												</div>
											</div>
										{:else}
											<div class="text-center py-4">
												<p class="text-sm text-orange-600">No tags available for this metric.</p>
												<p class="text-xs text-gray-500 mt-1">Tags are required for data scoping. Please try a different metric.</p>
											</div>
										{/if}
										
										{#if errors.tags}
											<p class="mt-2 text-sm text-red-600">{errors.tags}</p>
										{/if}
									</div>
								{/if}
							{/if}

							<!-- Selected Metrics -->
							{#if selectedMetrics.length > 0}
								<div>
									<Label class="mb-2">Selected Metrics</Label>
									<div class="space-y-3">
										{#each selectedMetrics as metric, index}
											<div class="rounded-md bg-gray-100 p-3 dark:bg-gray-800">
												<div class="flex items-center justify-between mb-2">
													<span class="text-sm font-medium">{metric.label}</span>
													<button 
														on:click={() => removeMetric(index)} 
														class="px-2 py-1 text-xs font-medium text-white bg-red-600 rounded hover:bg-red-700 focus:ring-2 focus:ring-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-800"
													>
														Remove
													</button>
												</div>
												{#if metric.tags && Object.keys(metric.tags).length > 0}
													<div class="flex flex-wrap gap-1">
														<span class="text-xs text-gray-600 dark:text-gray-400 mr-2">Filters:</span>
														{#each Object.entries(metric.tags) as [key, value]}
															<span class="inline-flex items-center px-2 py-1 text-xs font-medium text-green-800 bg-green-100 rounded-full dark:bg-green-900 dark:text-green-300">
																{key}:{value}
															</span>
														{/each}
													</div>
												{/if}
											</div>
										{/each}
									</div>
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