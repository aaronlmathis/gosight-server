<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Endpoint, Metric } from '$lib/types';
	import { formatDate, formatDuration, getBadgeClass } from '$lib/utils';
	import { AlertTriangle } from 'lucide-svelte';

	export let endpoint: Endpoint;
	export let hostInfo: {
		summary: Record<string, string>;
		procs: string;
		uptime: string;
		users_loggedin: string;
	};
	export let metrics: Metric[];
	export let processes: any[];
	export let alerts: any[] = [];
	export let logs: any[] = [];
	export let runCommand: (command: string) => void;

	// Element bindings
	let metricsChartEl!: HTMLElement;
	let miniCpuChartEl!: HTMLElement;
	let miniMemoryChartEl!: HTMLElement;
	let miniSwapChartEl!: HTMLElement;
	let mainCpuTableBody!: HTMLTableSectionElement;
	let mainMemTableBody!: HTMLTableSectionElement;
	let cpuPercentLabel!: HTMLElement;
	let memPercentLabel!: HTMLElement;
	let swapPercentLabel!: HTMLElement;

	// Chart instances and data
	let overviewCharts: any = {};
	let chartData = { cpu: [] as any[], memory: [] as any[], swap: [] as any[] };

	function getTopProcessesTooltip(type: 'cpu' | 'memory', limit = 5) {
		const key = type === 'cpu' ? 'cpu_percent' : 'memory_percent';
		if (!processes || processes.length === 0) return '';

		const validProcesses = processes.filter(
			(p) => p && typeof parseFloat(p?.[key] || 0) === 'number'
		);
		const top = [...validProcesses]
			.sort((a, b) => parseFloat(b?.[key] || 0) - parseFloat(a?.[key] || 0))
			.slice(0, limit);
		const rows = top
			.map(
				(p) =>
					`<tr><td>${p.pid || 'N/A'}</td><td>${p.user || 'N/A'}</td><td>${parseFloat(p?.[key] || 0).toFixed(1)}%</td><td>${p.command || 'N/A'}</td></tr>`
			)
			.join('');
		return `<table class="text-xs w-full"><thead><tr><th>PID</th><th>User</th><th>%</th><th>Cmd</th></tr></thead><tbody>${rows}</tbody></table>`;
	}

	export function initCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;
		if (!metricsChartEl || !miniCpuChartEl || !miniMemoryChartEl || !miniSwapChartEl) {
			console.log('Chart containers not ready yet');
			return;
		}

		// Initialize chart data storage
		if (!(window as any).chartData) {
			(window as any).chartData = {
				cpu: [],
				memory: [],
				swap: [],
				cpu_mini: [],
				memory_mini: [],
				swap_mini: []
			};
		}

		console.log('Initializing overview charts...');

		// Main performance chart
		if (!overviewCharts.main) {
			const mainOptions = {
				chart: {
					type: 'line',
					height: 350,
					animations: { enabled: false },
					toolbar: { show: false },
					zoom: { enabled: false }
				},
				series: [
					{ name: 'CPU Usage %', data: chartData.cpu || [] },
					{ name: 'Memory Usage %', data: chartData.memory || [] }
				],
				xaxis: {
					type: 'datetime',
					labels: { format: 'HH:mm:ss' }
				},
				yaxis: {
					min: 0,
					max: 100,
					labels: { formatter: (val: number) => `${val.toFixed(1)}%` }
				},
				stroke: { curve: 'smooth', width: 2 },
				colors: ['#3B82F6', '#10B981'],
				grid: { strokeDashArray: 4 },
				legend: { position: 'top' },
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => `${val.toFixed(2)}%` }
				}
			};
			overviewCharts.main = new window.ApexCharts(metricsChartEl, mainOptions);
			overviewCharts.main.render();
		}

		// Mini charts
		if (!overviewCharts.cpu) {
			const cpuOptions = {
				chart: {
					type: 'area',
					height: 60,
					sparkline: { enabled: true },
					animations: { enabled: false }
				},
				series: [{ data: chartData.cpu || [] }],
				stroke: { curve: 'smooth', width: 2 },
				fill: { opacity: 0.3, gradient: { enabled: true } },
				colors: ['#3B82F6'],
				tooltip: {
					fixed: { enabled: false },
					x: { show: false },
					y: {
						formatter: () => getTopProcessesTooltip('cpu'),
						title: { formatter: () => 'Top CPU Processes' }
					},
					marker: { show: false }
				}
			};
			overviewCharts.cpu = new window.ApexCharts(miniCpuChartEl, cpuOptions);
			overviewCharts.cpu.render();
		}

		if (!overviewCharts.memory) {
			const memOptions = {
				chart: {
					type: 'area',
					height: 60,
					sparkline: { enabled: true },
					animations: { enabled: false }
				},
				series: [{ data: chartData.memory || [] }],
				stroke: { curve: 'smooth', width: 2 },
				fill: { opacity: 0.3, gradient: { enabled: true } },
				colors: ['#10B981'],
				tooltip: {
					fixed: { enabled: false },
					x: { show: false },
					y: {
						formatter: () => getTopProcessesTooltip('memory'),
						title: { formatter: () => 'Top Memory Processes' }
					},
					marker: { show: false }
				}
			};
			overviewCharts.memory = new window.ApexCharts(miniMemoryChartEl, memOptions);
			overviewCharts.memory.render();
		}

		if (!overviewCharts.swap) {
			const swapOptions = {
				chart: {
					type: 'area',
					height: 60,
					sparkline: { enabled: true },
					animations: { enabled: false }
				},
				series: [{ data: chartData.swap || [] }],
				stroke: { curve: 'smooth', width: 2 },
				fill: { opacity: 0.3, gradient: { enabled: true } },
				colors: ['#F59E0B'],
				tooltip: {
					fixed: { enabled: false },
					x: { show: false },
					y: { formatter: (val: number) => `${val.toFixed(1)}%` },
					marker: { show: false }
				}
			};
			overviewCharts.swap = new window.ApexCharts(miniSwapChartEl, swapOptions);
			overviewCharts.swap.render();
		}
	}

	// Reactive update: when new metrics arrive, push into charts
	$: if (metrics.length && overviewCharts.main) {
		const latest = metrics[metrics.length - 1];
		const ts = new Date(latest.timestamp).getTime();

		// Parse metrics to extract percentages - metrics don't have direct cpu_percent properties
		// Instead we need to look for specific metric names in the flat metrics array
		let cpuPercent = 0;
		let memoryPercent = 0;
		let swapPercent = 0;

		// Since this reactive statement expects a single latest metric but we need to parse
		// multiple metrics, we should handle this differently. For now, let's skip the chart updates
		// and focus on fixing the undefined errors
		console.log('Metric structure:', latest);

		// Skip chart updates for now to prevent errors
		// chartData.cpu.push([ts, latest.cpu_percent]);
		// chartData.memory.push([ts, latest.memory_percent]);
		// chartData.swap.push([ts, latest.swap_percent]);
		// overviewCharts.main.updateSeries([
		// 	{ name: 'CPU Usage %', data: chartData.cpu },
		// 	{ name: 'Memory Usage %', data: chartData.memory }
		// ]);
		// overviewCharts.cpu.updateSeries([{ data: chartData.cpu.map(([t, v]) => [t, v]) }]);
		// overviewCharts.memory.updateSeries([{ data: chartData.memory.map(([t, v]) => [t, v]) }]);
		// overviewCharts.swap.updateSeries([{ data: chartData.swap.map(([t, v]) => [t, v]) }]);
		// // Update percent labels
		// cpuPercentLabel.textContent = `${latest.cpu_percent.toFixed(1)}%`;
		// memPercentLabel.textContent = `${latest.memory_percent.toFixed(1)}%`;
		// swapPercentLabel.textContent = `${latest.swap_percent.toFixed(1)}%`;
	}

	// Reactive update: when processes change, update process tables
	$: if (processes && mainCpuTableBody) {
		const validProcesses = processes.filter(
			(p) => p && typeof parseFloat(p?.cpu_percent || 0) === 'number'
		);
		const topCpu = [...validProcesses]
			.sort((a, b) => parseFloat(b?.cpu_percent || 0) - parseFloat(a?.cpu_percent || 0))
			.slice(0, 5);
		mainCpuTableBody.innerHTML = topCpu
			.map(
				(p) =>
					`<tr><td class="px-3 py-2 text-center">${p.pid || 'N/A'}</td><td class="px-3 py-2">${p.user || 'N/A'}</td><td class="px-3 py-2 text-right">${parseFloat(p?.cpu_percent || 0).toFixed(1)}</td><td class="px-3 py-2">${p.command || 'N/A'}</td></tr>`
			)
			.join('');
	}
	$: if (processes && mainMemTableBody) {
		const validProcesses = processes.filter(
			(p) => p && typeof parseFloat(p?.memory_percent || 0) === 'number'
		);
		const topMem = [...validProcesses]
			.sort((a, b) => parseFloat(b?.memory_percent || 0) - parseFloat(a?.memory_percent || 0))
			.slice(0, 5);
		mainMemTableBody.innerHTML = topMem
			.map(
				(p) =>
					`<tr><td class="px-3 py-2 text-center">${p.pid || 'N/A'}</td><td class="px-3 py-2">${p.user || 'N/A'}</td><td class="px-3 py-2 text-right">${parseFloat(p?.memory_percent || 0).toFixed(1)}</td><td class="px-3 py-2">${p.command || 'N/A'}</td></tr>`
			)
			.join('');
	}

	onMount(async () => {
		await tick();
		initCharts();
	});
</script>

<div class="p-4" id="overview" role="tabpanel" aria-labelledby="overview-tab">
	<!-- System Info and Metrics Section -->
	<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
		<!-- Info Cards -->
		<div class="space-y-6 lg:col-span-2">
			<!-- Basic Info -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">System Information</h3>
				<dl class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Hostname</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{endpoint.hostname || 'N/A'}
						</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">IP Address</dt>
						<dd class="text-sm text-gray-900 dark:text-white">{endpoint.ip}</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Operating System</dt>
						<dd class="text-sm text-gray-900 dark:text-white">{endpoint.os || 'N/A'}</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Agent Version</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{endpoint.agent_version || 'N/A'}
						</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Seen</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{endpoint.last_seen ? formatDate(endpoint.last_seen) : 'N/A'}
						</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Uptime</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{hostInfo.uptime || (endpoint.uptime ? formatDuration(endpoint.uptime) : 'N/A')}
						</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Running Processes</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{hostInfo.procs || 'N/A'}
						</dd>
					</div>
					<div>
						<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Users Logged In</dt>
						<dd class="text-sm text-gray-900 dark:text-white">
							{hostInfo.users_loggedin || 'N/A'}
						</dd>
					</div>
					<!-- Additional host info from summary -->
					{#if hostInfo.summary && Object.keys(hostInfo.summary).length > 0}
						{#each Object.entries(hostInfo.summary) as [key, value]}
							<div>
								<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
									{key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' ')}
								</dt>
								<dd class="text-sm text-gray-900 dark:text-white">
									{value || 'N/A'}
								</dd>
							</div>
						{/each}
					{/if}
				</dl>
			</div>

			<!-- Metrics Chart -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Performance Metrics</h3>
				<div id="metrics-chart" bind:this={metricsChartEl}></div>
			</div>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Quick Actions -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Quick Actions</h3>
				<div class="space-y-2">
					<button
						class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('restart')}
					>
						Restart Service
					</button>
					<button
						class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('status')}
					>
						Check Status
					</button>
					<button
						class="w-full rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('update')}
					>
						Update Agent
					</button>
				</div>
			</div>

			<!-- Recent Alerts -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Recent Alerts</h3>
				<div class="space-y-3">
					{#each Array.isArray(alerts) ? alerts.slice(0, 5) : [] as alert}
						<div
							class="flex items-center space-x-3 rounded-lg p-2 {alert.level === 'critical'
								? 'bg-red-50 dark:bg-red-900/20'
								: alert.level === 'warning'
									? 'bg-yellow-50 dark:bg-yellow-900/20'
									: 'bg-blue-50 dark:bg-blue-900/20'}"
						>
							<AlertTriangle
								size={16}
								class={alert.level === 'critical'
									? 'text-red-500'
									: alert.level === 'warning'
										? 'text-yellow-500'
										: 'text-blue-500'}
							/>
							<div class="min-w-0 flex-1">
								<p class="truncate text-xs font-medium text-gray-900 dark:text-white">
									{alert.message || alert.title || alert.name}
								</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">
									{formatDate(alert.last_fired || alert.created_at || new Date())}
								</p>
							</div>
						</div>
					{:else}
						<p class="text-sm text-gray-500 dark:text-gray-400">No recent alerts</p>
					{/each}
				</div>
			</div>
		</div>
	</div>

	<!-- Metrics Row: CPU, Memory, Swap -->
	<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
		<div
			class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
		>
			<div class="flex items-center justify-between">
				<p class="text-sm text-gray-500 dark:text-gray-400">CPU Usage</p>
				<p
					class="text-2xl font-bold text-indigo-600 dark:text-blue-400"
					bind:this={cpuPercentLabel}
				>
					--%
				</p>
			</div>
			<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
			<div class="mt-2">
				<div bind:this={miniCpuChartEl} class="h-20 w-full"></div>
			</div>
		</div>

		<div
			class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
		>
			<div class="flex items-center justify-between">
				<p class="text-sm text-gray-500 dark:text-gray-400">Memory Used</p>
				<p
					class="text-2xl font-bold text-green-600 dark:text-green-400"
					bind:this={memPercentLabel}
				>
					--%
				</p>
			</div>
			<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
			<div class="mt-2">
				<div bind:this={miniMemoryChartEl} class="h-20 w-full"></div>
			</div>
		</div>

		<div
			class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
		>
			<div class="flex items-center justify-between">
				<p class="text-sm text-gray-500 dark:text-gray-400">Swap Used</p>
				<p
					class="text-2xl font-bold text-yellow-500 dark:text-yellow-400"
					bind:this={swapPercentLabel}
				>
					--%
				</p>
			</div>
			<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
			<div class="mt-2">
				<div bind:this={miniSwapChartEl} class="h-20 w-full"></div>
			</div>
		</div>
	</div>

	<!-- Live Logs Section -->
	<div class="mb-6 grid grid-cols-1 gap-4">
		<div
			class="flex h-96 flex-col rounded-lg border border-gray-200 bg-white p-4 shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="mb-3 flex items-center justify-between">
				<h3 class="text-base font-semibold text-gray-800 dark:text-white">Live Logs</h3>
				<span class="text-xs text-gray-500 dark:text-gray-400">Last 10 entries</span>
			</div>
			<div
				class="h-full space-y-2 overflow-y-auto rounded-md border border-gray-200 bg-gray-50 p-3 font-mono text-sm break-words whitespace-pre-wrap text-gray-800 shadow-inner dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
			>
				{#each Array.isArray(logs) ? logs.slice(0, 10) : [] as log}
					<div class="text-xs">
						<span class="text-gray-500 dark:text-gray-400">[{formatDate(log.timestamp)}]</span>
						<span class="font-medium {getBadgeClass(log.level)}">{log.level}</span>
						<span class="ml-2">{log.message}</span>
					</div>
				{:else}
					<div class="text-gray-500 dark:text-gray-400">No logs available</div>
				{/each}
			</div>
		</div>
	</div>

	<!-- Top Processes -->
	<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2">
		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-700 dark:bg-gray-900"
		>
			<h3 class="text-md mb-2 font-semibold text-gray-800 dark:text-white">
				Top 5 Running Processes by CPU
			</h3>
			<div class="overflow-x-auto">
				<table class="w-full text-left text-sm text-gray-700 dark:text-gray-200">
					<thead
						class="bg-gray-100 text-xs text-gray-700 uppercase dark:bg-gray-700 dark:text-gray-300"
					>
						<tr>
							<th class="px-3 py-2 text-center" scope="col">PID</th>
							<th class="px-3 py-2" scope="col">User</th>
							<th class="px-3 py-2 text-right" scope="col">CPU %</th>
							<th class="px-3 py-2" scope="col">Command</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 dark:divide-gray-700" bind:this={mainCpuTableBody}>
						<tr
							><td colspan="4" class="px-3 py-4 text-center text-gray-500">Loading processes...</td
							></tr
						>
					</tbody>
				</table>
			</div>
		</div>

		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm sm:p-6 dark:border-gray-700 dark:bg-gray-900"
		>
			<h3 class="text-md mb-2 font-semibold text-gray-800 dark:text-white">
				Top 5 Running Processes by Memory
			</h3>
			<div class="overflow-x-auto">
				<table class="w-full text-left text-sm text-gray-700 dark:text-gray-200">
					<thead
						class="bg-gray-100 text-xs text-gray-700 uppercase dark:bg-gray-700 dark:text-gray-300"
					>
						<tr>
							<th class="px-3 py-2 text-center" scope="col">PID</th>
							<th class="px-3 py-2" scope="col">User</th>
							<th class="px-3 py-2 text-right" scope="col">MEM %</th>
							<th class="px-3 py-2" scope="col">Command</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 dark:divide-gray-700" bind:this={mainMemTableBody}>
						<tr
							><td colspan="4" class="px-3 py-4 text-center text-gray-500">Loading processes...</td
							></tr
						>
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>
