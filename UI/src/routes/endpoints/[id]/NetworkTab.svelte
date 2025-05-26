<script lang="ts">
	import { onMount, tick } from 'svelte';
	export let metrics: any[];
	let trafficChartEl!: HTMLElement;
	let networkCharts: any = {};

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

	onMount(async () => {
		await tick();
		initNetworkCharts();
	});

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
