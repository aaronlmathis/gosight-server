<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { Endpoint } from '$lib/types';

	// Types for the metric explorer
	interface MetricInfo {
		label: string;
		namespace: string;
		subnamespace: string;
		name: string;
	}

	interface ChartSlot {
		id: string;
		metrics: MetricInfo[];
		filters: Record<string, boolean>;
		availableDimensions: string[];
		period: string;
		graphType: string;
		groupBy: string;
		aggregate: string;
		chart: any;
		shouldReset: boolean;
	}

	// State variables
	let allMetrics: MetricInfo[] = [];
	let tagSuggestions: Record<string, Set<string>> = {};
	let chartSlots: ChartSlot[] = [];
	let activeSlotId: string | null = null;
	let endpoints: Endpoint[] = [];
	let loading = false;

	// Control elements
	let metricPanelsEl: HTMLElement;
	let metricInput = '';
	let tagInput = '';
	let showMetricSuggestions = false;
	let showTagSuggestions = false;
	let filteredMetrics: MetricInfo[] = [];
	let filteredTags: string[] = [];

	// Time range options
	const timeRanges = [
		{ value: '10m', label: '10 minutes' },
		{ value: '30m', label: '30 minutes' },
		{ value: '1h', label: '1 hour' },
		{ value: '3h', label: '3 hours' },
		{ value: '12h', label: '12 hours' },
		{ value: '1d', label: '1 day' },
		{ value: '3d', label: '3 days' },
		{ value: '1w', label: '1 week' }
	];
	let selectedTimeRange = '3h';

	onMount(async () => {
		initSlots(3); // Initialize 3 chart slots
		await loadMetrics();
		await loadEndpointTagSuggestions();

		// Set up ApexCharts if available
		if (typeof window !== 'undefined' && (window as any).ApexCharts) {
			console.log('ApexCharts loaded');
		}
	});

	// Initialize chart slots
	function initSlots(count: number) {
		chartSlots = [];
		for (let i = 0; i < count; i++) {
			const slotId = `chart-slot-${i}`;
			chartSlots.push({
				id: slotId,
				metrics: [],
				filters: {},
				availableDimensions: [],
				period: '5m',
				graphType: 'area',
				groupBy: '',
				aggregate: '',
				chart: null,
				shouldReset: false
			});
		}

		// Set first slot as active
		if (chartSlots.length > 0) {
			setActiveSlot(chartSlots[0].id);
		}
	}

	// Set active slot
	function setActiveSlot(slotId: string) {
		activeSlotId = slotId;
		// Reset control values to match active slot
		const panel = chartSlots.find((s) => s.id === slotId);
		if (panel) {
			// Update form controls to match panel settings
		}
	}

	// Load all metrics from API
	async function loadMetrics() {
		try {
			const namespaces = (await api.request('/api/v1/')) as string[];
			for (const ns of namespaces) {
				const subs = (await api.request(`/api/v1/${ns}`)) as string[];
				for (const sub of subs) {
					const metrics = (await api.request(`/api/v1/${ns}/${sub}`)) as string[];
					for (const m of metrics) {
						allMetrics.push({
							label: m,
							namespace: ns,
							subnamespace: sub,
							name: m.split('.').pop() || m
						});
					}
				}
			}
		} catch (err) {
			console.error('Failed to load metrics:', err);
		}
	}

	// Load endpoint tag suggestions
	async function loadEndpointTagSuggestions() {
		try {
			const [hostsRes, containersRes] = await Promise.all([
				api.endpoints.getHosts(),
				api.endpoints.getContainers()
			]);

			const hosts = (hostsRes as any).data || [];
			const containers = (containersRes as any).data || [];
			const allEndpoints = [...hosts, ...containers];

			tagSuggestions = {};
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
					if (!tagSuggestions[k]) tagSuggestions[k] = new Set();
					tagSuggestions[k].add(val);
				}
			}
		} catch (err) {
			console.error('Failed to load tag suggestions:', err);
		}
	}

	// Handle metric search input
	function handleMetricInput(event: Event) {
		const target = event.target as HTMLInputElement;
		metricInput = target.value;
		const q = metricInput.trim().toLowerCase();

		if (q.length < 2) {
			showMetricSuggestions = false;
			return;
		}

		filteredMetrics = allMetrics.filter((m) => m.label.toLowerCase().includes(q)).slice(0, 10);
		showMetricSuggestions = filteredMetrics.length > 0;
	}

	// Handle tag input
	function handleTagInput(event: Event) {
		const target = event.target as HTMLInputElement;
		tagInput = target.value;
		const q = tagInput.trim().toLowerCase();

		if (q.length < 2) {
			showTagSuggestions = false;
			return;
		}

		const allEntries: string[] = [];
		for (const [key, values] of Object.entries(tagSuggestions)) {
			for (const val of values) {
				if (val) {
					allEntries.push(`${key}:${val}`);
				}
			}
		}

		filteredTags = allEntries.filter((entry) => entry.toLowerCase().includes(q)).slice(0, 10);
		showTagSuggestions = filteredTags.length > 0;
	}

	// Add selected metric to active slot
	async function addSelectedMetric(metric: MetricInfo) {
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (!panel) return;

		panel.metrics.push(metric);

		// Load dimensions for this metric
		try {
			const [namespace, subnamespace, ...metricParts] = metric.label.split('.');
			const shortMetric = metricParts.join('.');
			const dims = await api.request(
				`/api/v1/${namespace}/${subnamespace}/${shortMetric}/dimensions`
			);
			if (Array.isArray(dims)) {
				panel.availableDimensions = dims;
			}
		} catch (err) {
			console.error('Failed to fetch dimensions for', metric.label, err);
		}

		metricInput = '';
		showMetricSuggestions = false;
		await loadData();
	}

	// Add selected filter to active slot
	function addSelectedFilter(filterStr: string) {
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (!panel) return;

		panel.filters[filterStr] = true;
		tagInput = '';
		showTagSuggestions = false;
		loadData();
	}

	// Remove metric from active slot
	function removeMetric(metric: MetricInfo) {
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (!panel) return;

		panel.metrics = panel.metrics.filter((m) => m.label !== metric.label);

		if (panel.metrics.length === 0) {
			resetSlot(panel);
		} else {
			loadData();
		}
	}

	// Remove filter from active slot
	function removeFilter(filterStr: string) {
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (!panel) return;

		delete panel.filters[filterStr];
		loadData();
	}

	// Reset a chart slot
	function resetSlot(panel: ChartSlot) {
		if (panel.chart) {
			try {
				panel.chart.destroy();
			} catch (err) {
				console.warn('Chart destroy failed', err);
			}
			panel.chart = null;
		}

		panel.metrics = [];
		panel.filters = {};
		panel.availableDimensions = [];
		panel.period = '5m';
		panel.graphType = 'area';
		panel.groupBy = '';
		panel.aggregate = '';
		panel.shouldReset = false;
	}

	// Load and render chart data
	async function loadData() {
		for (const panel of chartSlots) {
			if (panel.shouldReset || panel.metrics.length === 0) {
				resetSlot(panel);
				continue;
			}

			// Destroy old chart before reload
			if (panel.chart) {
				panel.chart.destroy();
				panel.chart = null;
			}

			const now = new Date();
			const startDate = new Date(now);
			const timeRange = panel.period || '5m';

			// Calculate time range
			if (timeRange.endsWith('m')) {
				startDate.setMinutes(now.getMinutes() - parseInt(timeRange));
			} else if (timeRange.endsWith('h')) {
				startDate.setHours(now.getHours() - parseInt(timeRange));
			} else if (timeRange.endsWith('d')) {
				startDate.setDate(now.getDate() - parseInt(timeRange));
			}

			const start = startDate.toISOString();
			const end = now.toISOString();

			// Calculate step
			let step = '15s';
			if (timeRange.endsWith('m')) {
				const minutes = parseInt(timeRange);
				if (minutes <= 15) step = '5s';
				else if (minutes <= 30) step = '15s';
				else if (minutes <= 60) step = '30s';
				else step = '60s';
			} else if (timeRange.endsWith('h')) {
				const hours = parseInt(timeRange);
				if (hours <= 6) step = '2m';
				else if (hours <= 12) step = '5m';
				else step = '10m';
			} else if (timeRange.endsWith('d')) {
				const days = parseInt(timeRange);
				if (days <= 1) step = '10m';
				else step = '30m';
			}

			const allSeries: any[] = [];

			for (const metric of panel.metrics) {
				const tagFilter = Object.keys(panel.filters)
					.map((f) => f.replace(':', '='))
					.join(',');

				const url =
					`/api/v1/query?metric=${encodeURIComponent(metric.label)}&start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}&step=${encodeURIComponent(step)}` +
					(tagFilter ? `&tags=${encodeURIComponent(tagFilter)}` : '');

				try {
					const data = await api.request(url);
					if (!data || !Array.isArray(data)) continue;

					// Build series for each metric
					const metricSeries = buildSeries(data, panel);
					allSeries.push(...metricSeries);
				} catch (err) {
					console.error('Failed to load metric', metric.label, err);
				}
			}

			// Render chart with combined series
			if (allSeries.length > 0) {
				renderChartPanel(panel, allSeries);
			}
		}
	}

	// Build chart series from data
	function buildSeries(dataArray: any[], panel: ChartSlot) {
		if (!panel) return [];

		const groupKey = panel.groupBy || '';
		const groups: Record<string, any[]> = {};

		for (const d of dataArray) {
			let id = 'unknown';

			if (groupKey && d.tags?.[groupKey]) {
				id = d.tags[groupKey];
			} else {
				id = d.tags?.endpoint_id || d.tags?.instance || d.tags?.hostname || 'unknown';
			}

			if (!groups[id]) groups[id] = [];
			groups[id].push([d.timestamp, d.value]);
		}

		return Object.entries(groups).map(([name, data]) => ({ name, data }));
	}

	// Render chart panel
	function renderChartPanel(panel: ChartSlot, series: any[]) {
		const slotEl = document.getElementById(panel.id);
		if (!slotEl) return;

		slotEl.innerHTML = '';
		slotEl.className = 'relative rounded border p-2 bg-white dark:bg-gray-900';

		// Handle stacked area correctly
		let chartType = panel.graphType;
		let stacked = false;

		if (panel.graphType === 'stacked-area') {
			chartType = 'area';
			stacked = true;
		}

		const manySeries = series.length > 5;
		const manyPoints = series.reduce((sum, s) => sum + s.data.length, 0) > 500;
		const optimized = manySeries || manyPoints;

		const chartOptions = {
			chart: {
				type: chartType || 'area',
				height: 250,
				zoom: {
					type: 'x',
					enabled: true,
					autoScaleYaxis: true
				},
				toolbar: {
					autoSelected: 'zoom'
				},
				stacked: stacked,
				animations: {
					enabled: !optimized
				}
			},
			stroke: {
				curve: manySeries ? 'straight' : 'smooth',
				width: 2
			},
			fill: {
				type: chartType === 'area' ? 'gradient' : 'solid',
				gradient:
					chartType === 'area'
						? {
								shadeIntensity: 1,
								opacityFrom: 0.4,
								opacityTo: 0,
								stops: [0, 90, 100]
							}
						: undefined
			},
			dataLabels: {
				enabled: false
			},
			markers: {
				size: 0
			},
			title: {
				text: panel.metrics.map((m) => m.name).join(', '),
				align: 'left',
				style: {
					fontSize: '14px',
					fontWeight: 600,
					color: '#263238'
				}
			},
			xaxis: {
				type: 'datetime',
				labels: {
					datetimeFormatter: {
						month: "MMM 'yy",
						day: 'dd MMM',
						hour: 'HH:mm',
						minute: 'HH:mm'
					}
				}
			},
			yaxis: {
				labels: {
					formatter: (val: number) => val.toFixed(2)
				},
				title: {
					text: 'Value'
				}
			},
			tooltip: {
				shared: true,
				intersect: false,
				x: { format: 'MMM dd HH:mm' },
				y: { formatter: (val: number) => val.toFixed(2) }
			},
			series: series
		};

		if (panel.graphType === 'bar') {
			chartOptions.chart.type = 'bar';
			chartOptions.stroke = { curve: 'smooth', width: 0 };
			chartOptions.fill = {
				type: 'solid',
				gradient: undefined
			};
			(chartOptions as any).plotOptions = {
				bar: {
					horizontal: false,
					columnWidth: '50%',
					endingShape: 'rounded'
				}
			};
		}

		if (typeof window !== 'undefined' && (window as any).ApexCharts) {
			const chart = new (window as any).ApexCharts(slotEl, chartOptions);
			chart.render();
			panel.chart = chart;
		}
	}

	// Handle period change for active slot
	function handlePeriodChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (panel) {
			panel.period = target.value;
			loadData();
		}
	}

	// Handle graph type change for active slot
	function handleGraphTypeChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (panel) {
			panel.graphType = target.value;
			loadData();
		}
	}

	// Handle group by change for active slot
	function handleGroupByChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (panel) {
			panel.groupBy = target.value;
			loadData();
		}
	}

	// Handle aggregate change for active slot
	function handleAggregateChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		const panel = chartSlots.find((s) => s.id === activeSlotId);
		if (panel) {
			panel.aggregate = target.value;
			loadData();
		}
	}

	// Get active panel for reactive UI updates
	$: activePanel = chartSlots.find((s) => s.id === activeSlotId);
</script>

<svelte:head>
	<title>Metric Explorer - GoSight</title>
	<script src="https://cdn.jsdelivr.net/npm/apexcharts"></script>
</svelte:head>

<section class="space-y-6 p-4">
	<!-- Header -->
	<div class="mb-4 flex flex-wrap items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-gray-800 dark:text-white">📊 Metric Explorer</h1>
			<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
				Explore and compare metrics with full control over aggregation, filters, and visualization.
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<select
				bind:value={selectedTimeRange}
				class="rounded border-gray-300 text-sm dark:border-gray-600"
			>
				{#each timeRanges as range}
					<option value={range.value}>{range.label}</option>
				{/each}
			</select>
			<button
				class="rounded border border-gray-300 px-2 py-1 text-sm text-gray-700 dark:border-gray-600 dark:text-gray-300"
			>
				UTC
			</button>
			<button
				on:click={loadData}
				class="text-sm text-gray-600 hover:text-blue-600 dark:text-gray-400"
			>
				↻ Refresh
			</button>
			<button class="rounded bg-orange-600 px-3 py-1.5 text-sm text-white hover:bg-orange-700">
				+ Add to Dashboard
			</button>
		</div>
	</div>

	<!-- Grid Layout -->
	<div class="grid grid-cols-4 gap-4">
		<!-- LEFT COLUMN: Control Panel -->
		<div
			class="col-span-1 max-h-[80vh] space-y-4 overflow-y-auto rounded-lg border border-gray-100 bg-white p-4 shadow dark:border-gray-800 dark:bg-gray-900"
		>
			<!-- Metrics Search -->
			<div>
				<label for="metric-search" class="text-xs font-semibold text-gray-500 dark:text-gray-400"
					>Metrics</label
				>
				<input
					id="metric-search"
					bind:value={metricInput}
					on:input={handleMetricInput}
					type="text"
					placeholder="Search metrics..."
					class="mt-1 w-full rounded border border-gray-300 bg-white px-2 py-1 text-sm dark:border-gray-600 dark:bg-gray-800"
				/>

				<!-- Metric Suggestions -->
				{#if showMetricSuggestions}
					<div class="mt-2 max-h-40 space-y-1 overflow-y-auto text-sm">
						{#each filteredMetrics as metric}
							<button
								tabindex="0"
								class="cursor-pointer rounded px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700"
								on:click={() => addSelectedMetric(metric)}
								on:keydown={(e) => e.key === 'Enter' && addSelectedMetric(metric)}
							>
								{metric.label}
							</button>
						{/each}
					</div>
				{/if}

				<!-- Selected Metrics -->
				{#if activePanel}
					<div class="mt-2 flex flex-wrap gap-1">
						{#each activePanel.metrics as metric}
							<span
								class="mr-1 mb-1 inline-flex items-center rounded-full bg-blue-100 px-2 py-1 text-xs text-blue-800 dark:bg-blue-900 dark:text-blue-200"
							>
								{metric.label}
								<button
									on:click={() => removeMetric(metric)}
									class="ml-2 text-sm text-blue-500 hover:text-red-500"
								>
									&times;
								</button>
							</span>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Filter Search -->
			<div>
				<label
					for="filter-search"
					class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">From</label
				>
				<div class="relative w-full">
					<input
						id="filter-search"
						bind:value={tagInput}
						on:input={handleTagInput}
						type="text"
						placeholder="Filter by hostname, tag, or endpoint..."
						class="w-full rounded border border-gray-300 bg-white px-3 py-2 text-sm text-gray-800 placeholder-gray-400 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200 dark:placeholder-gray-500"
					/>

					<!-- Tag Suggestions -->
					{#if showTagSuggestions}
						<div
							class="absolute z-50 mt-1 max-h-60 w-full overflow-y-auto rounded-md border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
						>
							{#each filteredTags as tag}
								<button
									class="w-full cursor-pointer px-3 py-2 text-left text-sm whitespace-nowrap hover:bg-gray-100 dark:hover:bg-gray-700"
									on:click={() => addSelectedFilter(tag)}
								>
									{tag}
								</button>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Selected Filters -->
				{#if activePanel}
					<div class="mt-2 flex flex-wrap gap-2 text-sm">
						{#each Object.keys(activePanel.filters) as filterStr}
							<span
								class="mr-1 mb-1 inline-flex items-center rounded bg-gray-200 px-2 py-1 text-xs dark:bg-gray-800"
							>
								{filterStr}
								<button
									on:click={() => removeFilter(filterStr)}
									class="ml-1 text-red-500 hover:text-red-700"
								>
									&times;
								</button>
							</span>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Aggregate By -->
			<div>
				<label for="aggregate-select" class="text-xs font-semibold text-gray-500 dark:text-gray-400"
					>Aggregate by</label
				>
				<select
					id="aggregate-select"
					on:change={handleAggregateChange}
					value={activePanel?.aggregate || ''}
					class="mt-1 w-full rounded border border-gray-300 text-sm dark:border-gray-600"
				>
					<option value="">None</option>
					<option value="sum">Sum</option>
					<option value="avg">Average</option>
					<option value="min">Minimum</option>
					<option value="max">Maximum</option>
					<option value="stddev">Std Deviation</option>
				</select>
			</div>

			<!-- Group By -->
			<div>
				<label for="groupby-select" class="text-xs font-semibold text-gray-500 dark:text-gray-400"
					>Group by</label
				>
				<select
					id="groupby-select"
					on:change={handleGroupByChange}
					value={activePanel?.groupBy || ''}
					class="mt-1 w-full rounded border border-gray-300 text-sm dark:border-gray-600"
				>
					<option value="">None</option>
					<option value="hostname">Hostname</option>
					<option value="endpoint_id">Endpoint ID</option>
					<option value="platform">Platform</option>
					<option value="os_version">OS Version</option>
					<option value="interface">Interface</option>
					<option value="container_name">Container Name</option>
					<option value="job">Job</option>
					{#if activePanel}
						{#each activePanel.availableDimensions as dim}
							<option value={dim}>{dim}</option>
						{/each}
					{/if}
				</select>
			</div>

			<!-- Graph Options -->
			<div class="border-t pt-3">
				<div class="mb-2 text-xs font-semibold text-gray-500 dark:text-gray-400">Graph Options</div>
				<div class="mt-1 space-y-2">
					<div class="flex justify-between gap-2">
						<label for="period-select" class="w-1/2 text-xs">Period</label>
						<select
							id="period-select"
							on:change={handlePeriodChange}
							value={activePanel?.period || '5m'}
							class="w-1/2 rounded border border-gray-300 text-sm dark:border-gray-600"
						>
							<option value="5m">5m</option>
							<option value="10m">10m</option>
							<option value="30m">30m</option>
							<option value="1h">1h</option>
						</select>
					</div>
					<div class="flex justify-between gap-2">
						<label for="graphtype-select" class="w-1/2 text-xs">Graph Type</label>
						<select
							id="graphtype-select"
							on:change={handleGraphTypeChange}
							value={activePanel?.graphType || 'area'}
							class="w-1/2 rounded border border-gray-300 text-sm dark:border-gray-600"
						>
							<option value="area">Area</option>
							<option value="stacked-area">Stacked Area</option>
							<option value="line">Line</option>
							<option value="bar">Bar</option>
						</select>
					</div>
				</div>
			</div>
		</div>

		<!-- RIGHT COLUMN: Chart Grid -->
		<div class="col-span-3">
			<div bind:this={metricPanelsEl} class="flex flex-col gap-4">
				{#each chartSlots as slot}
					<div
						id={slot.id}
						class="rounded border-2 {activeSlotId === slot.id
							? 'border-blue-400'
							: 'border-gray-200 dark:border-gray-700'} flex h-[250px] cursor-pointer items-center justify-center bg-white p-4 text-gray-400 dark:bg-gray-900"
						on:click={() => setActiveSlot(slot.id)}
					>
						{#if slot.metrics.length === 0}
							<span class="text-sm">➕ New Chart</span>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	</div>
</section>
