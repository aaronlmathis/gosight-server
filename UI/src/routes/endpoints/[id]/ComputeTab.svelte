<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Metric } from '$lib/types';
	import { chart } from 'svelte-apexcharts';

	export let metrics: Metric[];
	export let cpuInfo: Record<string, string>;
	export let cpuTimeCounters: Record<string, string>;
	export let perCoreData: Record<string, { usage?: number; clock?: number }>;
	export let processes: any[] = [];

	// Chart data
	let cpuUsageData: Array<[number, number]> = [];
	let cpuLoadData = {
		load1: [] as Array<[number, number]>,
		load5: [] as Array<[number, number]>,
		load15: [] as Array<[number, number]>
	};
	let memoryUsageData: Array<[number, number]> = [];
	let swapUsageData: Array<[number, number]> = [];

	// Reactive chart options
	$: isDark = typeof window !== 'undefined' && document.documentElement.classList.contains('dark');
	$: textColor = isDark ? '#d1d5db' : '#374151';
	$: gridColor = isDark ? '#374151' : '#e5e7eb';
	$: theme = isDark ? 'dark' : 'light';

	// CPU Usage Chart Options
	$: cpuUsageChartOptions = {
		chart: {
			type: 'area',
			height: 250,
			zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent'
		},
		series: [{ name: 'CPU Usage %', data: cpuUsageData }],
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'HH:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			},
			min: 0,
			max: 100
		},
		colors: ['#3b82f6'],
		tooltip: {
			x: { format: 'HH:mm:ss' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` }
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// CPU Load Chart Options
	$: cpuLoadChartOptions = {
		chart: {
			type: 'area',
			height: 250,
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent'
		},
		series: [
			{ name: '1m', data: cpuLoadData.load1 },
			{ name: '5m', data: cpuLoadData.load5 },
			{ name: '15m', data: cpuLoadData.load15 }
		],
		stroke: { curve: 'smooth', width: 3 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.5, opacityTo: 0.2, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'HH:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			min: 0,
			max: 4,
			tickAmount: 4,
			labels: {
				formatter: (val: number) => val.toFixed(2),
				style: { colors: textColor }
			},
			title: {
				text: 'Load Avg',
				style: { color: textColor }
			}
		},
		colors: ['#3b82f6', '#10b981', '#f59e0b'],
		tooltip: {
			x: { format: 'HH:mm:ss' },
			y: { formatter: (val: number) => val.toFixed(2) }
		},
		annotations: {
			yaxis: [
				{
					y: 1.0,
					borderColor: '#facc15',
					label: { text: 'Warn ≥ 1.0', style: { background: '#facc15', color: '#000' } }
				},
				{
					y: 1.5,
					borderColor: '#f87171',
					label: { text: 'High ≥ 1.5', style: { background: '#f87171', color: '#fff' } }
				}
			]
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// Memory Usage Chart Options
	$: memoryUsageChartOptions = {
		chart: {
			type: 'area',
			height: 250,
			zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent'
		},
		series: [{ name: 'Memory Usage %', data: memoryUsageData }],
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'HH:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			},
			min: 0,
			max: 100
		},
		colors: ['#10b981'],
		tooltip: {
			x: { format: 'HH:mm:ss' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` }
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// Swap Usage Chart Options
	$: swapUsageChartOptions = {
		chart: {
			type: 'area',
			height: 250,
			zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent'
		},
		series: [{ name: 'Swap Usage %', data: swapUsageData }],
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'HH:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			},
			min: 0,
			max: 100
		},
		colors: ['#ef4444'],
		tooltip: {
			x: { format: 'HH:mm:ss' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` }
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// Process metrics data when they change
	function processMetrics(metrics: Metric[]) {
		// Extract CPU usage data
		const cpuMetrics = metrics.filter((m) => m.name === 'cpu_usage' || m.name === 'cpu_percent');
		cpuUsageData = cpuMetrics.map((m) => [new Date(m.timestamp).getTime(), m.value]);

		// Extract load average data
		const load1Metrics = metrics.filter((m) => m.name === 'load1');
		const load5Metrics = metrics.filter((m) => m.name === 'load5');
		const load15Metrics = metrics.filter((m) => m.name === 'load15');

		cpuLoadData = {
			load1: load1Metrics.map((m) => [new Date(m.timestamp).getTime(), m.value]),
			load5: load5Metrics.map((m) => [new Date(m.timestamp).getTime(), m.value]),
			load15: load15Metrics.map((m) => [new Date(m.timestamp).getTime(), m.value])
		};

		// Extract memory usage data
		const memoryMetrics = metrics.filter(
			(m) => m.name === 'memory_usage' || m.name === 'memory_percent'
		);
		memoryUsageData = memoryMetrics.map((m) => [new Date(m.timestamp).getTime(), m.value]);

		// Extract swap usage data
		const swapMetrics = metrics.filter((m) => m.name === 'swap_usage' || m.name === 'swap_percent');
		swapUsageData = swapMetrics.map((m) => [new Date(m.timestamp).getTime(), m.value]);
	}

	// Reactive metrics processing
	$: if (metrics.length > 0) {
		processMetrics(metrics);
	}
</script>

<div class="p-4" id="compute" role="tabpanel" aria-labelledby="compute-tab">
	<!-- Compute Resources Section -->
	<div class="mb-6 rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
		<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Compute Resources</h2>
		<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
			<!-- CPU Usage Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">CPU Usage Over Time</h3>
				<div use:chart={cpuUsageChartOptions} class="mt-2 h-64"></div>
			</div>

			<!-- CPU Load Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">CPU Load Average</h3>
				<div use:chart={cpuLoadChartOptions} class="mt-2 h-64"></div>
			</div>

			<!-- Memory Usage Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">Memory Usage Over Time</h3>
				<div use:chart={memoryUsageChartOptions} class="mt-2 h-64"></div>
			</div>

			<!-- Swap Usage Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">Swap Usage Over Time</h3>
				<div use:chart={swapUsageChartOptions} class="mt-2 h-64"></div>
			</div>
		</div>
	</div>

	<!-- CPU Info Cards Section -->
	<div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
		<!-- CPU Info Card -->
		<div class="rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
			<h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">CPU Information</h3>
			<div class="space-y-3">
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Model:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.model}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Vendor:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.vendor}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Cores:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.cores}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Threads:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.threads}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Base Clock:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.baseClock}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Cache:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.cache}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Family:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.family}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Stepping:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.stepping}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Physical CPUs:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuInfo.physical}</span>
				</div>
			</div>
		</div>

		<!-- CPU Time Counters Card -->
		<div class="rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
			<h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">CPU Time Counters</h3>
			<div class="space-y-3">
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">User:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.user}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">System:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.system}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Idle:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.idle}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Nice:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.nice}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">IOWait:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.iowait}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">IRQ:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.irq}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">SoftIRQ:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.softirq}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Steal:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.steal}</span>
				</div>
				<div class="flex justify-between text-xs">
					<span class="text-gray-500 dark:text-gray-400">Guest:</span>
					<span class="font-medium text-gray-900 dark:text-white">{cpuTimeCounters.guest}</span>
				</div>
			</div>
		</div>

		<!-- Per-Core Usage Card -->
		<div class="rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
			<h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">Per-Core Usage</h3>
			<div class="max-h-64 overflow-y-auto">
				{#if Object.keys(perCoreData).length > 0}
					<div class="grid grid-cols-1 gap-2">
						{#each Object.entries(perCoreData).sort(([a], [b]) => parseInt(a.replace('core', '')) - parseInt(b.replace('core', ''))) as [coreId, coreData]}
							<div class="rounded border bg-gray-50 p-2 dark:border-gray-600 dark:bg-gray-700">
								<div class="flex items-center justify-between">
									<span class="text-xs font-medium text-gray-700 dark:text-gray-300">
										{coreId.startsWith('core') ? coreId.replace('core', 'Core ') : `Core ${coreId}`}
									</span>
									<div class="flex space-x-2 text-xs">
										{#if coreData.usage !== undefined}
											<span class="text-blue-600 dark:text-blue-400">
												{coreData.usage.toFixed(1)}%
											</span>
										{/if}
										{#if coreData.clock !== undefined}
											<span class="text-green-600 dark:text-green-400">
												{coreData.clock.toFixed(0)} MHz
											</span>
										{/if}
									</div>
								</div>
								{#if coreData.usage !== undefined}
									<div class="mt-1 h-1.5 w-full rounded-full bg-gray-200 dark:bg-gray-600">
										<div
											class="h-1.5 rounded-full bg-gradient-to-r from-blue-500 to-blue-600"
											style="width: {Math.min(100, Math.max(0, coreData.usage))}%"
										></div>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				{:else}
					<div class="py-4 text-center text-xs text-gray-500 dark:text-gray-400">
						No per-core data available
					</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Processes Section -->
	<div class="mb-6 rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
		<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Running Processes</h2>
		<div class="mt-4">
			<table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
				<thead class="bg-gray-50 dark:bg-gray-800">
					<tr>
						<th
							scope="col"
							class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"
						>
							PID
						</th>
						<th
							scope="col"
							class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"
						>
							User
						</th>
						<th
							scope="col"
							class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"
						>
							CPU %
						</th>
						<th
							scope="col"
							class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"
						>
							Memory %
						</th>
						<th
							scope="col"
							class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400"
						>
							Command
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white dark:divide-gray-700 dark:bg-gray-900">
					{#each Array.isArray(processes) ? processes.slice(0, 10) : [] as proc}
						<tr class="hover:bg-gray-50 dark:hover:bg-gray-800">
							<td class="px-3 py-2 text-xs text-gray-900 dark:text-white">
								{proc?.pid || 'N/A'}
							</td>
							<td class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">
								{proc?.username || proc?.user || 'N/A'}
							</td>
							<td class="px-3 py-2 text-xs text-gray-900 dark:text-white">
								{parseFloat(proc?.cpu_percent || 0).toFixed(1)}%
							</td>
							<td class="px-3 py-2 text-xs text-gray-900 dark:text-white">
								{parseFloat(proc?.memory_percent || 0).toFixed(1)}%
							</td>
							<td class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">
								<span
									class="block max-w-xs truncate"
									title={proc?.cmdline || proc?.command || proc?.name || 'N/A'}
								>
									{proc?.cmdline || proc?.command || proc?.name || 'N/A'}
								</span>
							</td>
						</tr>
					{:else}
						<tr>
							<td
								colspan="5"
								class="px-3 py-2 text-center text-xs text-gray-500 dark:text-gray-400"
							>
								Loading processes...
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
