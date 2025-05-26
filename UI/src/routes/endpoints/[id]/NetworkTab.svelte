<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { chart } from 'svelte-apexcharts';

	export let metrics: any[];

	// Chart data
	let networkData = {
		upload: [] as Array<[number, number]>,
		download: [] as Array<[number, number]>
	};

	// Reactive chart options
	$: isDark = typeof window !== 'undefined' && document.documentElement.classList.contains('dark');
	$: textColor = isDark ? '#d1d5db' : '#374151';
	$: gridColor = isDark ? '#374151' : '#e5e7eb';
	$: theme = isDark ? 'dark' : 'light';

	$: trafficChartOptions = {
		chart: {
			type: 'area',
			height: 320,
			toolbar: { show: false },
			animations: { enabled: true },
			background: 'transparent'
		},
		series: [
			{ name: 'Upload', data: networkData.upload },
			{ name: 'Download', data: networkData.download }
		],
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
				formatter: (val: number) => formatBytes(val),
				style: { colors: textColor }
			},
			title: {
				text: 'Bytes/sec',
				style: { color: textColor }
			}
		},
		colors: ['#3b82f6', '#10b981'],
		tooltip: {
			x: { format: 'HH:mm:ss' },
			y: { formatter: (val: number) => `${formatBytes(val)}/s` }
		},
		legend: {
			labels: { colors: textColor }
		},
		grid: { borderColor: gridColor },
		theme: { mode: theme }
	};

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}

	function isNetworkUpload(m: any): boolean {
		const name = m.name.toLowerCase();
		const ns = (m.namespace || '').toLowerCase();
		return (
			name === 'network_upload' ||
			name === 'net_upload' ||
			name === 'tx_bytes' ||
			(ns.includes('network') && (name.includes('upload') || name.includes('tx')))
		);
	}

	function isNetworkDownload(m: any): boolean {
		const name = m.name.toLowerCase();
		const ns = (m.namespace || '').toLowerCase();
		return (
			name === 'network_download' ||
			name === 'net_download' ||
			name === 'rx_bytes' ||
			(ns.includes('network') && (name.includes('download') || name.includes('rx')))
		);
	}

	function processNetworkMetrics(metrics: any[]) {
		const uploadMetrics = metrics.filter(isNetworkUpload);
		const downloadMetrics = metrics.filter(isNetworkDownload);

		networkData = {
			upload: uploadMetrics.map((m) => [new Date(m.timestamp).getTime(), m.value || 0]),
			download: downloadMetrics.map((m) => [new Date(m.timestamp).getTime(), m.value || 0])
		};
	}

	// Reactive metrics processing
	$: if (metrics.length > 0) {
		processNetworkMetrics(metrics);
	}

	function initNetworkCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;
		if (!networkCharts.traffic) {
			const options = {
				chart: {
					type: 'area',
					height: 320,
					toolbar: { show: false },
					animations: { enabled: true },
					background: 'transparent'
				},
				series: [
					{ name: 'Upload (Mbps)', data: [] },
					{ name: 'Download (Mbps)', data: [] }
				],
				xaxis: { type: 'datetime', labels: { format: 'HH:mm:ss' } },
				yaxis: { min: 0, labels: { formatter: (v: number) => `${v.toFixed(1)} Mbps` } },
				colors: ['#3b82f6', '#10b981'],
				stroke: { curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.7, opacityTo: 0.1, stops: [0, 100] }
				},
				tooltip: {
					theme: 'dark',
					x: { format: 'HH:mm:ss' },
					y: { formatter: (v: number) => `${v.toFixed(2)} Mbps` }
				}
			};
			networkCharts.traffic = new window.ApexCharts(trafficChartEl, options);
			networkCharts.traffic.render();
		}
	}

	$: if (metrics.length && networkCharts.traffic) {
		const uploadSeries: [number, number][] = [];
		const downloadSeries: [number, number][] = [];
		metrics.forEach((m: any) => {
			const ts = new Date(m.timestamp).getTime();
			const val =
				parseFloat(m.value) > 1e6 ? parseFloat(m.value) / (1024 * 1024) : parseFloat(m.value);
			if (isNetworkUpload(m)) uploadSeries.push([ts, val]);
			else if (isNetworkDownload(m)) downloadSeries.push([ts, val]);
		});
		try {
			networkCharts.traffic.updateSeries(
				[
					{ name: 'Upload (Mbps)', data: uploadSeries },
					{ name: 'Download (Mbps)', data: downloadSeries }
				],
				false
			);
		} catch {}
	}
</script>

<div class="p-4" role="tabpanel" aria-labelledby="network-tab">
	<h3 class="mb-4 text-lg font-medium">Network</h3>
	<div class="mb-6 rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<div bind:this={trafficChartEl} class="h-80"></div>
	</div>
</div>
