<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Metric } from '$lib/types';

	export let metrics: Metric[];
	export let cpuInfo: Record<string, string>;
	export let cpuTimeCounters: Record<string, string>;
	export let perCoreData: Record<string, { usage?: number; clock?: number }>;
	export let processes: any[] = [];

	// Element bindings for chart containers
	let cpuUsageChartEl!: HTMLElement;
	let cpuLoadChartEl!: HTMLElement;
	let memoryUsageChartEl!: HTMLElement;
	let swapUsageChartEl!: HTMLElement;

	let computeCharts: any = {};

	function initComputeCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// CPU Usage Chart
		if (!computeCharts.cpuUsage) {
			const options = {
				chart: {
					type: 'area',
					height: 250,
					zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
					toolbar: { show: false },
					animations: { enabled: true }
				},
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				series: [{ name: 'CPU Usage %', data: [] }],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm:ss' } },
				yaxis: { labels: { formatter: (val: number) => `${val.toFixed(1)}%` }, min: 0, max: 100 },
				colors: ['#3b82f6'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => `${val.toFixed(1)}%` }
				}
			};
			computeCharts.cpuUsage = new window.ApexCharts(cpuUsageChartEl, options);
			computeCharts.cpuUsage.render();
		}

		// CPU Load Average Chart
		if (!computeCharts.cpuLoad) {
			const options = {
				chart: {
					type: 'area',
					height: 250,
					toolbar: { show: false },
					animations: { enabled: true }
				},
				stroke: { curve: 'smooth', width: 3 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.5, opacityTo: 0.2, stops: [0, 90, 100] }
				},
				series: [
					{ name: '1m', data: [] },
					{ name: '5m', data: [] },
					{ name: '15m', data: [] }
				],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm:ss' } },
				yaxis: {
					min: 0,
					max: 4,
					tickAmount: 4,
					labels: { formatter: (val: number) => val.toFixed(2) },
					title: { text: 'Load Avg' }
				},
				colors: ['#3b82f6', '#10b981', '#f59e0b'],
				tooltip: { x: { format: 'HH:mm:ss' }, y: { formatter: (val: number) => val.toFixed(2) } },
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
				}
			};
			computeCharts.cpuLoad = new window.ApexCharts(cpuLoadChartEl, options);
			computeCharts.cpuLoad.render();
		}

		// Memory Usage Chart
		if (!computeCharts.memoryUsage) {
			const options = {
				chart: {
					type: 'area',
					height: 250,
					zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
					toolbar: { show: false },
					animations: { enabled: true }
				},
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				series: [{ name: 'Memory Usage %', data: [] }],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm:ss' } },
				yaxis: { labels: { formatter: (val: number) => `${val.toFixed(1)}%` }, min: 0, max: 100 },
				colors: ['#10b981'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => `${val.toFixed(1)}%` }
				}
			};
			computeCharts.memoryUsage = new window.ApexCharts(memoryUsageChartEl, options);
			computeCharts.memoryUsage.render();
		}

		// Swap Usage Chart
		if (!computeCharts.swapUsage) {
			const options = {
				chart: {
					type: 'area',
					height: 250,
					zoom: { type: 'x', enabled: true, autoScaleYaxis: true },
					toolbar: { show: false },
					animations: { enabled: true }
				},
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				series: [{ name: 'Swap Usage %', data: [] }],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm:ss' } },
				yaxis: { labels: { formatter: (val: number) => `${val.toFixed(1)}%` }, min: 0, max: 100 },
				colors: ['#ef4444'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => `${val.toFixed(1)}%` }
				}
			};
			computeCharts.swapUsage = new window.ApexCharts(swapUsageChartEl, options);
			computeCharts.swapUsage.render();
		}
	}

	onMount(async () => {
		await tick();
		initComputeCharts();
	});

	// Reactive updates: update compute charts when metrics change
	$: if (metrics.length && computeCharts.cpuUsage) {
		const timestamps = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.cpu_usage]);
		computeCharts.cpuUsage.updateSeries([{ data: timestamps }], false);
		const load1 = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.load1 || 0]);
		const load5 = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.load5 || 0]);
		const load15 = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.load15 || 0]);
		computeCharts.cpuLoad.updateSeries([{ data: load1 }, { data: load5 }, { data: load15 }], false);
		const memory = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.memory_usage || 0]);
		computeCharts.memoryUsage.updateSeries([{ data: memory }], false);
		const swap = metrics.map((m: any) => [new Date(m.timestamp).getTime(), m.swap_usage || 0]);
		computeCharts.swapUsage.updateSeries([{ data: swap }], false);
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
				<div bind:this={cpuUsageChartEl} class="mt-2 h-32"></div>
			</div>

			<!-- CPU Load Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">CPU Load Average</h3>
				<div bind:this={cpuLoadChartEl} class="mt-2 h-32"></div>
			</div>

			<!-- Memory Usage Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">Memory Usage Over Time</h3>
				<div bind:this={memoryUsageChartEl} class="mt-2 h-32"></div>
			</div>

			<!-- Swap Usage Chart -->
			<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
				<h3 class="text-xs font-semibold text-gray-900 dark:text-white">Swap Usage Over Time</h3>
				<div bind:this={swapUsageChartEl} class="mt-2 h-32"></div>
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
