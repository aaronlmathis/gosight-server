<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Metric } from '$lib/types';
	import { formatBytes } from '$lib/utils';
	import { HardDrive, Activity, Database } from 'lucide-svelte';
	import { chart } from 'svelte-apexcharts';
	import type { ApexOptions } from 'apexcharts';

	export let metrics: Metric[];

	// Processed data
	let diskSummary = {
		total: 0,
		used: 0,
		free: 0,
		percentage: 0
	};

	let mountpoints: Array<{
		mountpoint: string;
		fstype: string;
		device: string;
		total: number;
		used: number;
		free: number;
		used_percent: number;
		inodes_used_percent?: number;
	}> = [];

	let selectedDevice = '';
	let availableDevices: string[] = [];

	// Chart data cache for I/O
	let ioDataCache: Record<
		string,
		{
			timestamps: string[];
			readCount: number[];
			writeCount: number[];
			readBytes: number[];
			writeBytes: number[];
		}
	> = {};

	const MAX_IO_POINTS = 30;

	// Reactive chart options and data
	$: isDark = typeof window !== 'undefined' && document.documentElement.classList.contains('dark');
	$: textColor = isDark ? '#d1d5db' : '#374151';
	$: gridColor = isDark ? '#374151' : '#e5e7eb';
	$: theme = isDark ? 'dark' : 'light';

	// Disk Usage Donut Chart
	$: diskUsageSeries = [diskSummary.used, diskSummary.free];
	$: diskUsageChartOptions = {
		chart: {
			type: 'donut',
			height: 250,
			toolbar: { show: false },
			background: 'transparent'
		},
		series: diskUsageSeries,
		labels: ['Used', 'Free'],
		colors: ['#ef4444', '#10b981'],
		plotOptions: {
			pie: {
				donut: {
					size: '60%',
					labels: {
						show: true,
						name: { fontSize: '14px', color: textColor },
						value: {
							fontSize: '20px',
							color: textColor,
							formatter: (val: number) => formatBytes(val)
						},
						total: {
							show: true,
							label: 'Total Usage',
							color: textColor,
							formatter: () => `${diskSummary.percentage.toFixed(1)}%`
						}
					}
				}
			}
		},
		legend: {
			position: 'bottom',
			labels: { colors: textColor }
		},
		theme: { mode: theme }
	};

	// Partition Usage Radial Chart
	$: radialSeries = mountpoints.slice(0, 6).map((mp) => Math.round(mp.used_percent));
	$: radialLabels = mountpoints
		.slice(0, 6)
		.map((mp) => (mp.mountpoint.length > 12 ? mp.mountpoint.slice(0, 10) + '…' : mp.mountpoint));
	$: avgUsage =
		radialSeries.length > 0 ? radialSeries.reduce((a, b) => a + b, 0) / radialSeries.length : 0;

	$: radialChartOptions = {
		chart: {
			type: 'radialBar',
			height: 400,
			toolbar: { show: false }
		},
		series: radialSeries,
		plotOptions: {
			radialBar: {
				hollow: { size: '40%' },
				dataLabels: {
					name: {
						show: true,
						fontSize: '14px',
						color: textColor
					},
					value: {
						show: true,
						fontSize: '16px',
						color: textColor,
						formatter: (val: number) => `${val}%`
					},
					total: {
						show: true,
						label: 'Avg Used',
						fontSize: '16px',
						color: textColor,
						formatter: () => `${avgUsage.toFixed(1)}%`
					}
				}
			}
		},
		labels: radialLabels,
		colors: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#14b8a6'],
		legend: {
			show: true,
			position: 'bottom',
			labels: { colors: textColor }
		},
		theme: { mode: theme }
	};

	// Inode Usage Chart
	$: inodeData = mountpoints.filter((mp) => mp.inodes_used_percent !== undefined).slice(0, 8);
	$: inodeCategories = inodeData.map((mp) => mp.mountpoint);
	$: inodeSeries = [
		{ name: 'Inodes Used %', data: inodeData.map((mp) => mp.inodes_used_percent || 0) }
	];

	$: inodeChartOptions = {
		chart: {
			type: 'bar',
			height: 250,
			toolbar: { show: false }
		},
		series: inodeSeries,
		plotOptions: {
			bar: {
				horizontal: true,
				borderRadius: 4
			}
		},
		dataLabels: { enabled: false },
		xaxis: {
			categories: inodeCategories,
			labels: {
				formatter: (val: number) => `${val.toFixed(1)}%`,
				style: { colors: textColor }
			},
			title: {
				text: 'Inodes Used %',
				style: { color: textColor }
			},
			max: 100
		},
		yaxis: {
			labels: { style: { colors: textColor } }
		},
		colors: ['#6366f1'],
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// I/O Charts data
	$: selectedDeviceCache = selectedDevice ? ioDataCache[selectedDevice] : null;
	$: ioTimestamps = selectedDeviceCache?.timestamps || [];

	// IOPS Chart
	$: iopsSeries = [
		{ name: 'Read Count', data: selectedDeviceCache?.readCount || [] },
		{ name: 'Write Count', data: selectedDeviceCache?.writeCount || [] }
	];

	$: iopsChartOptions = {
		chart: {
			type: 'line',
			height: 250,
			toolbar: { show: false }
		},
		series: iopsSeries,
		stroke: { curve: 'smooth', width: 2 },
		dataLabels: { enabled: false },
		xaxis: {
			categories: ioTimestamps,
			labels: { style: { colors: textColor } }
		},
		yaxis: {
			title: { text: 'Ops/sec', style: { color: textColor } },
			labels: { style: { colors: textColor } }
		},
		colors: ['#3b82f6', '#10b981'],
		legend: { labels: { colors: textColor } },
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	// Throughput Chart
	$: throughputSeries = [
		{ name: 'Read Bytes', data: selectedDeviceCache?.readBytes || [] },
		{ name: 'Write Bytes', data: selectedDeviceCache?.writeBytes || [] }
	];

	$: throughputChartOptions = {
		chart: {
			type: 'line',
			height: 250,
			toolbar: { show: false }
		},
		series: throughputSeries,
		stroke: { curve: 'smooth', width: 2 },
		dataLabels: { enabled: false },
		xaxis: {
			categories: ioTimestamps,
			labels: { style: { colors: textColor } }
		},
		yaxis: {
			title: { text: 'Throughput', style: { color: textColor } },
			labels: {
				style: { colors: textColor },
				formatter: (val: number) => formatBytes(val)
			}
		},
		colors: ['#f59e0b', '#ef4444'],
		legend: { labels: { colors: textColor } },
		tooltip: {
			y: { formatter: (val: number) => formatBytes(val) }
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	function formatIO(val: number): string {
		if (val >= 1024 ** 3) return (val / 1024 ** 3).toFixed(2) + ' GB/s';
		if (val >= 1024 ** 2) return (val / 1024 ** 2).toFixed(1) + ' MB/s';
		if (val >= 1024) return (val / 1024).toFixed(1) + ' KB/s';
		return val.toFixed(0) + ' B/s';
	}

	function processMetrics(metrics: Metric[]) {
		const usageByMount: Record<string, any> = {};
		const ioByDevice: Record<string, any> = {};

		// Process disk usage metrics
		for (const metric of metrics) {
			const dims = metric.dimensions || {};
			const mp = dims.mountpoint;
			const dev = dims.device;

			if (metric.subnamespace === 'Disk' && mp) {
				if (!usageByMount[mp]) usageByMount[mp] = {};
				usageByMount[mp][metric.name] = metric.value;
				if (dims.device) {
					usageByMount[mp].device = dims.device.replace('/dev/', '');
				}
				if (dims.fstype) usageByMount[mp].fstype = dims.fstype;
			}

			if (metric.subnamespace === 'DiskIO' && dev) {
				if (!ioByDevice[dev]) ioByDevice[dev] = {};
				ioByDevice[dev][metric.name] = metric.value;
			}
		}

		// Update summary
		let totalSum = 0,
			usedSum = 0;
		const mountpointData: typeof mountpoints = [];

		for (const [mp, data] of Object.entries(usageByMount)) {
			const total = data.total || 0;
			const used = data.used || 0;
			const free = data.free || total - used;
			const usedPercent = total > 0 ? (used / total) * 100 : 0;

			totalSum += total;
			usedSum += used;

			mountpointData.push({
				mountpoint: mp,
				fstype: data.fstype || '—',
				device: data.device || '—',
				total,
				used,
				free,
				used_percent: usedPercent,
				inodes_used_percent: data.inodes_used_percent
			});
		}

		mountpoints = mountpointData.sort((a, b) => b.used_percent - a.used_percent);

		diskSummary = {
			total: totalSum,
			used: usedSum,
			free: totalSum - usedSum,
			percentage: totalSum > 0 ? (usedSum / totalSum) * 100 : 0
		};

		// Update device selector
		availableDevices = Object.keys(ioByDevice).sort();
		if (!selectedDevice && availableDevices.length > 0) {
			selectedDevice = availableDevices[0];
		}

		// Update I/O cache
		for (const [dev, data] of Object.entries(ioByDevice)) {
			updateIOCache(dev, data);
		}
	}

	function updateIOCache(device: string, data: any) {
		if (!ioDataCache[device]) {
			ioDataCache[device] = {
				timestamps: [],
				readCount: [],
				writeCount: [],
				readBytes: [],
				writeBytes: []
			};
		}

		const cache = ioDataCache[device];
		const ts = new Date().toLocaleTimeString([], {
			hour12: false,
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});

		cache.timestamps.push(ts);
		cache.readCount.push(data.read_count || 0);
		cache.writeCount.push(data.write_count || 0);
		cache.readBytes.push(data.read_bytes || 0);
		cache.writeBytes.push(data.write_bytes || 0);

		// Trim to max points
		if (cache.timestamps.length > MAX_IO_POINTS) {
			(Object.keys(cache) as Array<keyof typeof cache>).forEach((k) => cache[k].shift());
		}
	}

	function handleDeviceChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		selectedDevice = target.value;
	}

	$: if (metrics.length > 0) {
		processMetrics(metrics);
	}
</script>

<div class="p-4" id="disk" role="tabpanel" aria-labelledby="disk-tab">
	<!-- Disk Summary Section -->
	<div class="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="flex items-center">
				<HardDrive size={20} class="mr-2 text-blue-600 dark:text-blue-400" />
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Total Disk Space</p>
					<p class="text-xl font-bold text-blue-600 dark:text-blue-400">
						{formatBytes(diskSummary.total)}
					</p>
				</div>
			</div>
		</div>

		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="flex items-center">
				<Database size={20} class="mr-2 text-red-600 dark:text-red-400" />
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Used Disk Space</p>
					<p class="text-xl font-bold text-red-600 dark:text-red-400">
						{formatBytes(diskSummary.used)}
					</p>
				</div>
			</div>
		</div>

		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="flex items-center">
				<Database size={20} class="mr-2 text-green-600 dark:text-green-400" />
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Free Disk Space</p>
					<p class="text-xl font-bold text-green-600 dark:text-green-400">
						{formatBytes(diskSummary.free)}
					</p>
				</div>
			</div>
		</div>

		<div
			class="rounded-lg border border-gray-100 bg-white p-4 shadow-sm dark:border-gray-700 dark:bg-gray-900"
		>
			<div class="flex items-center">
				<Activity size={20} class="mr-2 text-yellow-600 dark:text-yellow-400" />
				<div>
					<p class="text-sm text-gray-500 dark:text-gray-400">Disk Usage</p>
					<p class="text-xl font-bold text-yellow-600 dark:text-yellow-400">
						{diskSummary.percentage.toFixed(1)}%
					</p>
				</div>
			</div>
		</div>
	</div>

	<!-- Charts Row 1: Usage Overview -->
	<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Partition Usage</h4>
			<div use:chart={radialChartOptions}></div>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">
				📊 Mountpoint Overview
			</h4>
			<div class="divide-y divide-gray-200 text-sm dark:divide-gray-700">
				{#each mountpoints.slice(0, 8) as mp}
					<div class="grid grid-cols-2 gap-4 py-2">
						<span class="font-semibold text-blue-600 dark:text-blue-400">{mp.mountpoint}</span>
						<div class="text-right">
							<div class="mb-1 text-xs">{formatBytes(mp.used)} / {formatBytes(mp.total)}</div>
							<div class="h-2 w-full rounded bg-gray-200 dark:bg-gray-700">
								<div class="h-2 rounded bg-emerald-500" style="width: {mp.used_percent}%"></div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>

	<!-- I/O Section -->
	<div class="mb-6">
		<div class="mb-4 flex items-center justify-between">
			<h4 class="text-sm font-semibold text-gray-700 dark:text-gray-200">Disk I/O by Device</h4>
			<select
				bind:value={selectedDevice}
				on:change={handleDeviceChange}
				class="rounded border border-gray-300 bg-white px-3 py-1 text-sm text-gray-900 dark:border-gray-700 dark:bg-gray-900 dark:text-white"
			>
				{#each availableDevices as device}
					<option value={device}>
						{device}
						{#if mountpoints.find((mp) => mp.device === device)}
							({mountpoints.find((mp) => mp.device === device)?.mountpoint})
						{/if}
					</option>
				{/each}
			</select>
		</div>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
				<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">IOPS</h4>
				<div use:chart={iopsChartOptions}></div>
			</div>

			<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
				<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Throughput</h4>
				<div use:chart={throughputChartOptions}></div>
			</div>
		</div>
	</div>

	<!-- Charts Row 2: Activity and Inodes -->
	<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h4
				class="mb-4 border-b border-gray-200 pb-2 text-sm font-semibold text-gray-700 dark:border-gray-700 dark:text-gray-200"
			>
				📌 Most Used Mountpoints
			</h4>
			<div class="space-y-3">
				{#each mountpoints.slice(0, 5) as mp}
					<div class="grid grid-cols-2 gap-4">
						<span class="font-medium text-blue-600 dark:text-blue-400">{mp.mountpoint}</span>
						<div>
							<div class="mb-1 flex justify-between text-xs">
								<span>{mp.used_percent.toFixed(1)}%</span>
							</div>
							<div class="h-2 w-full rounded bg-gray-200 dark:bg-gray-700">
								<div class="h-2 rounded bg-blue-500" style="width: {mp.used_percent}%"></div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">🗂️ Inode Usage</h4>
			<div use:chart={inodeChartOptions}></div>
		</div>
	</div>

	<!-- Disk Usage Chart -->
	<div class="mb-6">
		<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
			<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">Disk Usage</h4>
			<div class="flex justify-center">
				<div use:chart={diskUsageChartOptions}></div>
			</div>
		</div>
	</div>

	<!-- Mountpoint Table -->
	<div class="rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<h4 class="mb-4 text-sm font-semibold text-gray-700 dark:text-gray-200">
			Disk Usage by Mountpoint
		</h4>
		<div class="overflow-x-auto">
			<table class="min-w-full text-sm">
				<thead class="bg-gray-50 text-left text-gray-600 dark:bg-gray-700 dark:text-gray-300">
					<tr>
						<th class="px-4 py-2">Mountpoint</th>
						<th class="px-4 py-2">FS Type</th>
						<th class="px-4 py-2">Total</th>
						<th class="px-4 py-2">Used</th>
						<th class="px-4 py-2">Free</th>
						<th class="px-4 py-2">% Used</th>
						<th class="px-4 py-2">Device</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
					{#each mountpoints as mp}
						<tr class="text-gray-700 dark:text-gray-300">
							<td class="px-4 py-2 font-medium text-blue-500 dark:text-blue-400">{mp.mountpoint}</td
							>
							<td class="px-4 py-2">{mp.fstype}</td>
							<td class="px-4 py-2">{formatBytes(mp.total)}</td>
							<td class="px-4 py-2">{formatBytes(mp.used)}</td>
							<td class="px-4 py-2">{formatBytes(mp.free)}</td>
							<td class="px-4 py-2">
								<span
									class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium
									{mp.used_percent > 90
										? 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400'
										: mp.used_percent > 75
											? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400'
											: 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400'}"
								>
									{mp.used_percent.toFixed(1)}%
								</span>
							</td>
							<td class="px-4 py-2">{mp.device}</td>
						</tr>
					{:else}
						<tr>
							<td colspan="7" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
								No disk data available
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>
