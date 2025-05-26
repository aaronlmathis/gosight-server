<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Metric } from '$lib/types';
	import { chart } from 'svelte-apexcharts';

	export let metrics: Metric[];
	export let cpuInfo: Record<string, string>;
	export let cpuTimeCounters: Record<string, string>;
	export let perCoreData: Record<string, { usage?: number; clock?: number }>;
	export let processes: any[] = [];

	// Chart instances
	let cpuChart: ApexCharts;
	let cpuLoadChart: ApexCharts;
	let memoryChart: ApexCharts;
	let swapChart: ApexCharts;

	// Chart data - keep accumulated points up to limit
	let cpuUsageData: Array<[number, number]> = [];
	let cpuLoadData = {
		load1: [] as Array<[number, number]>,
		load5: [] as Array<[number, number]>,
		load15: [] as Array<[number, number]>
	};
	let memoryUsageData: Array<[number, number]> = [];
	let swapUsageData: Array<[number, number]> = [];

	// Process history for tooltips
	let processHistory: Array<{ timestamp: number; processes: any[] }> = [];
	let latestCpuPercent = 0;
	let latestMemUsedPercent = 0;
	let lastProcessedTimestamp = 0; // Track last processed metric timestamp

	const MAX_DATA_POINTS = 50; // Keep last 50 points per chart

	// Reactive chart options
	$: isDark = typeof window !== 'undefined' && document.documentElement.classList.contains('dark');
	$: textColor = isDark ? '#d1d5db' : '#374151';
	$: gridColor = isDark ? '#374151' : '#e5e7eb';
	$: theme = isDark ? 'dark' : 'light';

	// Helper function to find closest process snapshot
	function findClosestSnapshot(ts: number) {
		let closest = null;
		let minDiff = Infinity;
		for (const snap of processHistory) {
			const diff = Math.abs(snap.timestamp - ts);
			if (diff < minDiff) {
				closest = snap;
				minDiff = diff;
			}
		}
		return closest;
	}

	// Helper function to truncate command names
	function truncateCommand(cmd: string, max: number = 30): string {
		if (!cmd) return '(?)';
		return cmd.length > max ? cmd.slice(0, max - 1) + '…' : cmd;
	}

	// Helper function to get current usage
	function getCurrentUsage(type: 'cpu' | 'memory'): number {
		if (type === 'cpu') return latestCpuPercent;
		if (type === 'memory') return latestMemUsedPercent;
		return 0;
	}

	// Generate process tooltip for charts
	function generateProcessTooltip(isMem: boolean) {
		return function ({ series, seriesIndex, dataPointIndex, w }: any) {
			const point = w.config.series[seriesIndex].data[dataPointIndex];
			let hoverTime;

			if (Array.isArray(point)) {
				hoverTime = point[0];
			} else if (point && typeof point === 'object' && 'x' in point) {
				hoverTime = point.x;
			} else if (typeof point === 'number') {
				hoverTime = point;
			} else {
				return 'No process data';
			}
			const snapshot = findClosestSnapshot(hoverTime);
			if (!snapshot || !snapshot.processes) return 'No process data';

			const labelKey = isMem ? 'mem_percent' : 'cpu_percent';
			const processes = snapshot.processes;

			const rows = processes
				.sort((a: any, b: any) => (parseFloat(b[labelKey]) || 0) - (parseFloat(a[labelKey]) || 0))
				.slice(0, 5)
				.map((p: any) => {
					const full = p.cmdline || p.exe || '(?)';
					const short = truncateCommand(full, 30);
					const value = (parseFloat(p[labelKey]) || 0).toFixed(1);
					const color = isMem ? '#10b981' : '#3b82f6';

					return `
						<tr style="border-bottom: 1px solid #e5e7eb;">
							<td style="padding:4px 6px; font-size:11px; color:#6b7280;">${p.pid || '?'}</td>
							<td title="${full}" style="max-width:150px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; padding:4px 6px; font-size:11px;">
								${short}
							</td>
							<td style="text-align:right; padding:4px 6px; font-weight:500; font-size:11px; color:${color};">${value}%</td>
						</tr>`;
				})
				.join('');

			return `
				<div >
					<div style="font-weight:600; font-size:12px; margin-bottom:6px; display:flex; justify-content:space-between;">
						<span>Top Processes (${isMem ? 'Memory' : 'CPU'})</span>
						<span>Total: ${getCurrentUsage(isMem ? 'memory' : 'cpu').toFixed(1)}%</span>
					</div>
					<table style="width:100%; border-collapse:collapse;">
						<thead>
							<tr style="text-align:left; font-size:11px; color:#9ca3af;">
								<th style="padding:4px 6px;">PID</th>
								<th style="padding:4px 6px;">Command</th>
								<th style="padding:4px 6px; text-align:right;">Usage</th>
							</tr>
						</thead>
						<tbody>
							${rows}
						</tbody>
					</table>
				</div>`;
		};
	}
	// Chart options without series data for binding
	$: cpuUsageChartOptions = {
		chart: {
			type: 'area',
			height: 250,
			zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent',
			events: {
				mounted: (chartContext: any, config: any) => {
					cpuChart = chartContext;
				}
			}
		},
		series: [], // Empty series - will be populated via updateSeries
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'hh:mm:ss A',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			}
		},
		colors: ['#3b82f6'],
		tooltip: {
			x: { format: 'hh:mm:ss A' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` },
			custom: generateProcessTooltip(false),
			cssClass: 'custom-process-tooltip'
		},
		grid: { borderColor: gridColor },
		dataLabels: { enabled: false },
		theme: { mode: theme }
	};

	// CPU Load Chart Options
	$: cpuLoadChartOptions = {
		chart: {
			type: 'line',
			height: 250,
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent',
			events: {
				mounted: (chartContext: any, config: any) => {
					cpuLoadChart = chartContext;
				}
			}
		},
		series: [], // Empty series - will be populated via updateSeries
		stroke: { curve: 'smooth', width: 3 },

		xaxis: {
			type: 'datetime',
			labels: {
				format: 'hh:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
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
			x: { format: 'hh:mm:ss' },
			y: { formatter: (val: number) => val.toFixed(2) },
			custom: generateProcessTooltip(false),
			cssClass: 'custom-process-tooltip'
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
		dataLabels: { enabled: false },
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
			background: 'transparent',
			events: {
				mounted: (chartContext: any, config: any) => {
					memoryChart = chartContext;
				}
			}
		},
		series: [], // Empty series - will be populated via updateSeries
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'hh:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			}
		},
		colors: ['#10b981'],
		tooltip: {
			x: { format: 'hh:mm:ss' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` },
			custom: generateProcessTooltip(true),
			cssClass: 'custom-process-tooltip'
		},
		grid: { borderColor: gridColor },
		dataLabels: { enabled: false },
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
			background: 'transparent',
			events: {
				mounted: (chartContext: any, config: any) => {
					swapChart = chartContext;
				}
			}
		},
		series: [], // Empty series - will be populated via updateSeries
		stroke: { curve: 'smooth', width: 2 },
		fill: {
			type: 'gradient',
			gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
		},
		xaxis: {
			type: 'datetime',
			labels: {
				format: 'hh:mm:ss',
				style: { colors: textColor }
			}
		},
		yaxis: {
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			}
		},
		colors: ['#ef4444'],
		tooltip: {
			x: { format: 'hh:mm:ss' },
			y: { formatter: (val: number) => `${val.toFixed(1)}%` }
		},
		grid: { borderColor: gridColor },
		dataLabels: { enabled: false },
		theme: { mode: theme }
	};

	const SWAP_BUCKET_MS = 1000; // 1s for tolerant pairing

	function dedupeSeries(series: Array<[number, number]>): Array<[number, number]> {
		const seen = new Set();
		return series.filter(([ts]) => {
			if (seen.has(ts)) return false;
			seen.add(ts);
			return true;
		});
	}

	function norm(str: string | undefined) {
		return (str || '').toLowerCase();
	}

	// Process metrics data when they change
	function processMetrics(allMetrics: Metric[]) {
		// Filter to only process new metrics (timestamps greater than last processed)
		const newMetrics = allMetrics.filter((m) => {
			const metricTime = new Date(m.timestamp).getTime();
			return metricTime > lastProcessedTimestamp;
		});

		if (newMetrics.length === 0) {
			console.log('ComputeTab: No new metrics to process');
			return;
		}

		const now = Date.now();

		// --- CPU Usage ---
		const totalCpuMetrics = newMetrics.filter(
			(m) =>
				norm(m.namespace) === 'system' &&
				norm(m.subnamespace) === 'cpu' &&
				m.name === 'usage_percent' &&
				m.dimensions?.scope === 'total'
		);
		if (totalCpuMetrics.length > 0) {
			const newCpuData: Array<[number, number]> = totalCpuMetrics.map((m) => [
				new Date(m.timestamp).getTime(),
				m.value
			]);
			cpuUsageData = dedupeSeries([...cpuUsageData, ...newCpuData]).slice(-MAX_DATA_POINTS);
			latestCpuPercent = totalCpuMetrics[totalCpuMetrics.length - 1].value;
		}

		// --- CPU Load ---
		const load1Metrics = newMetrics.filter(
			(m) =>
				norm(m.namespace) === 'system' && norm(m.subnamespace) === 'cpu' && m.name === 'load_avg_1'
		);
		const load5Metrics = newMetrics.filter(
			(m) =>
				norm(m.namespace) === 'system' && norm(m.subnamespace) === 'cpu' && m.name === 'load_avg_5'
		);
		const load15Metrics = newMetrics.filter(
			(m) =>
				norm(m.namespace) === 'system' && norm(m.subnamespace) === 'cpu' && m.name === 'load_avg_15'
		);
		if (load1Metrics.length > 0 || load5Metrics.length > 0 || load15Metrics.length > 0) {
			const newLoad1Data: Array<[number, number]> = load1Metrics.map((m) => [
				new Date(m.timestamp).getTime(),
				m.value
			]);
			const newLoad5Data: Array<[number, number]> = load5Metrics.map((m) => [
				new Date(m.timestamp).getTime(),
				m.value
			]);
			const newLoad15Data: Array<[number, number]> = load15Metrics.map((m) => [
				new Date(m.timestamp).getTime(),
				m.value
			]);
			cpuLoadData = {
				load1: dedupeSeries([...cpuLoadData.load1, ...newLoad1Data]).slice(-MAX_DATA_POINTS),
				load5: dedupeSeries([...cpuLoadData.load5, ...newLoad5Data]).slice(-MAX_DATA_POINTS),
				load15: dedupeSeries([...cpuLoadData.load15, ...newLoad15Data]).slice(-MAX_DATA_POINTS)
			};
		}

		// --- Memory Usage ---
		const memoryMetrics = newMetrics.filter(
			(m) =>
				norm(m.namespace) === 'system' &&
				norm(m.subnamespace) === 'memory' &&
				m.name === 'used_percent'
		);
		if (memoryMetrics.length > 0) {
			const newMemoryData: Array<[number, number]> = memoryMetrics.map((m) => [
				new Date(m.timestamp).getTime(),
				m.value
			]);
			memoryUsageData = dedupeSeries([...memoryUsageData, ...newMemoryData]).slice(
				-MAX_DATA_POINTS
			);
			latestMemUsedPercent = memoryMetrics[memoryMetrics.length - 1].value;
		}

		// --- Swap Usage (bucketed pairing, tolerant to ms mismatches) ---
		const swapBuckets: Record<number, { total?: number; free?: number }> = {};
		newMetrics.forEach((m) => {
			if (
				norm(m.namespace) === 'system' &&
				norm(m.subnamespace) === 'memory' &&
				(m.name === 'swap_total' || m.name === 'swap_free')
			) {
				const ts = Math.round(new Date(m.timestamp).getTime() / SWAP_BUCKET_MS) * SWAP_BUCKET_MS;
				if (!swapBuckets[ts]) swapBuckets[ts] = {};
				if (m.name === 'swap_total') swapBuckets[ts].total = m.value;
				if (m.name === 'swap_free') swapBuckets[ts].free = m.value;
			}
		});
		const newSwapData: Array<[number, number]> = [];
		for (const tsStr of Object.keys(swapBuckets)) {
			const ts = Number(tsStr);
			const { total, free } = swapBuckets[ts];
			if (typeof total === 'number' && typeof free === 'number' && total > 0) {
				const usedPercent = (100 * (total - free)) / total;
				newSwapData.push([ts, usedPercent]);
			}
		}
		if (newSwapData.length > 0) {
			swapUsageData = dedupeSeries([...swapUsageData, ...newSwapData]).slice(-MAX_DATA_POINTS);
			console.log('Swap Usage % points:', newSwapData);
		}

		// --- Process History ---
		if (processes && processes.length > 0) {
			processHistory = [...processHistory, { timestamp: now, processes }].slice(-50);
		}

		// --- Last processed timestamp ---
		if (newMetrics.length > 0) {
			const timestamps = newMetrics.map((m) => new Date(m.timestamp).getTime());
			lastProcessedTimestamp = Math.max(...timestamps);
		}
	}

	// Reactive statements to update chart series when data changes
	$: if (cpuChart && cpuUsageData.length > 0) {
		cpuChart.updateSeries([{ name: 'CPU Usage %', data: cpuUsageData }]);
	}

	$: if (
		cpuLoadChart &&
		(cpuLoadData.load1.length > 0 || cpuLoadData.load5.length > 0 || cpuLoadData.load15.length > 0)
	) {
		cpuLoadChart.updateSeries([
			{ name: '1m', data: cpuLoadData.load1 },
			{ name: '5m', data: cpuLoadData.load5 },
			{ name: '15m', data: cpuLoadData.load15 }
		]);
	}

	$: if (memoryChart && memoryUsageData.length > 0) {
		memoryChart.updateSeries([{ name: 'Memory Usage %', data: memoryUsageData }]);
	}

	$: if (swapChart && swapUsageData.length > 0) {
		swapChart.updateSeries([{ name: 'Swap Usage %', data: swapUsageData }]);
	}

	// Reactive metrics processing
	$: if (metrics.length > 0) {
		console.log('[ComputeTab] Metrics sample:', metrics.slice(0, 2));
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
								{parseFloat(proc?.mem_percent || 0).toFixed(1)}%
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
