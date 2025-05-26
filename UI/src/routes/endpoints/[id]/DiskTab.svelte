<script lang="ts">
	import { onMount, tick } from 'svelte';
	export let metrics: any[];

	let usageChartEl!: HTMLElement;
	let radialChartEl!: HTMLElement;
	let diskCharts: any = {};
	import { formatBytes } from '$lib/utils';

	function initDiskCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;
		if (!diskCharts.usage) {
			const options = {
					chart: {
						type: 'donut',
						height: 400,
						toolbar: { show: false },
						background: 'transparent'
					},
					series: [0, 100],
					labels: ['Used', 'Free'],
					colors: ['#ef4444', '#10b981'],
					plotOptions: {
						pie: {
							donut: {
								size: '60%',
								labels: {
									show: true,
									name: { fontSize: '14px' },
									value: { fontSize: '20px', formatter: (val: number) => formatBytes(val) },
									total: { show: true, label: 'Total', formatter: () => 'Loading...' }
								}
							}
						}
					}
				},
				optionsRadial = {
					chart: {
						type: 'radialBar',
						height: 400,
						toolbar: { show: false },
						background: 'transparent'
					},
					series: [],
					labels: [],
					colors: ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'],
					plotOptions: {
						radialBar: {
							hollow: { size: '40%' },
							dataLabels: { total: { formatter: () => '0.0%' } }
						}
					}
				};

			diskCharts.usage = new window.ApexCharts(usageChartEl, options);
			diskCharts.usage.render();
			diskCharts.radial = new window.ApexCharts(radialChartEl, optionsRadial);
			diskCharts.radial.render();
		}
	}

	onMount(async () => {
		await tick();
		initDiskCharts();
	});

	$: if (metrics.length && diskCharts.usage && diskCharts.radial) {
		const usageData: Record<string, { used: number; total: number }> = {};
		metrics.forEach((m: any) => {
			const { name, dimensions, value } = m;
			const mp = dimensions?.mount_point || dimensions?.device || '/';
			if (!usageData[mp]) usageData[mp] = { used: 0, total: 0 };
			if (name.includes('usage') || name.includes('used')) usageData[mp].used = Number(value);
			if (name.includes('total')) usageData[mp].total = Number(value);
		});
		const series: number[] = [];
		const labels: string[] = [];
		Object.entries(usageData).forEach(([mp, { used, total }]) => {
			if (used > 0 && total > 0) {
				series.push(used);
				labels.push(mp.length > 12 ? mp.slice(0, 10) + '…' : mp);
			}
		});
		if (series.length) {
			const totalUsed = series.reduce((a, b) => a + b, 0);
			const freeSeries = series.map(
				(u, i) => usageData[labels[i]].total - usageData[labels[i]].used
			);
			diskCharts.usage.updateSeries([series[0], freeSeries[0]]);
			diskCharts.usage.updateOptions({
				plotOptions: {
					pie: {
						donut: {
							labels: {
								total: {
									formatter: (): string =>
										`${((series[0] / usageData[labels[0]].total) * 100).toFixed(1)}%`
								}
							}
						}
					}
				}
			});
			diskCharts.radial.updateSeries(
				series.map((u, i) => Math.round((u / usageData[labels[i]].total) * 100))
			);
			diskCharts.radial.updateOptions({
				labels,
				plotOptions: {
					radialBar: {
						dataLabels: {
							total: { formatter: (): string => `${(totalUsed / series.length).toFixed(1)}%` }
						}
					}
				}
			});
		}
	}
</script>

<div class="p-4" role="tabpanel" aria-labelledby="disk-tab">
	<h3 class="mb-4 text-lg font-medium">Disk</h3>
	<div class="mb-6 rounded-lg bg-white p-4 shadow dark:bg-gray-800">
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div bind:this={usageChartEl} class="h-64"></div>
			<div bind:this={radialChartEl} class="h-64"></div>
		</div>
	</div>
</div>
