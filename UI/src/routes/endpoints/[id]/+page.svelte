<script lang="ts">
	import { page } from '$app/stores';
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import { websocketManager } from '$lib/websocket';
	import type { Endpoint, Metric, Event, LogEntry, Alert } from '$lib/types';
	import {
		formatDate,
		formatBytes,
		formatDuration,
		getStatusBadgeClass,
		getLevelBadgeClass
	} from '$lib/utils';
	import {
		ChevronLeft,
		Activity,
		AlertTriangle,
		Database,
		FileText,
		Settings,
		Cpu,
		HardDrive,
		Wifi,
		Monitor,
		Terminal,
		ScrollText
	} from 'lucide-svelte';

	let endpoint: Endpoint | null = null;
	let metrics: Metric[] = [];
	let events: Event[] = [];
	let logs: LogEntry[] = [];
	let alerts: Alert[] = [];
	let processes: any[] = []; // Add processes array
	let processHistory: Array<{ timestamp: number; processes: any[] }> = []; // Process history for tooltips
	let latestCpuPercent = 0; // Latest CPU percentage for tooltips
	let latestMemUsedPercent = 0; // Latest memory percentage for tooltips
	let loading = true;
	let error = '';
	let activeTab = 'overview';

	// CPU Info state
	let cpuInfo = {
		model: '--',
		vendor: '--',
		cores: '--', 
		threads: '--',
		baseClock: '--',
		cache: '--',
		family: '--',
		stepping: '--',
		physical: '--'
	};

	// CPU Time Counters state
	let cpuTimeCounters = {
		user: '--',
		system: '--',
		idle: '--',
		nice: '--',
		iowait: '--',
		irq: '--',
		softirq: '--',
		steal: '--',
		guest: '--',
		guest_nice: '--'
	};

	// Per-Core Usage state
	let perCoreData: Record<string, { usage?: number; clock?: number }> = {};

	const endpointId = $page.params.id;

	// Chart instances
	let overviewCharts: any = {};
	let computeCharts: any = {};
	let networkCharts: any = {};
	let diskCharts: any = {};

	let unsubscribeMetrics: (() => void) | null = null;
	let unsubscribeEvents: (() => void) | null = null;
	let unsubscribeLogs: (() => void) | null = null;
	let unsubscribeAlerts: (() => void) | null = null;
	let unsubscribeProcesses: (() => void) | null = null; // Add processes unsubscribe

	onMount(async () => {
		await loadEndpointData();
		setupRealTimeUpdates();
		// Initialize Overview charts since it's the default active tab
		setTimeout(() => {
			initOverviewCharts();
		}, 100);
	});

	onDestroy(() => {
		if (unsubscribeMetrics) unsubscribeMetrics();
		if (unsubscribeEvents) unsubscribeEvents();
		if (unsubscribeLogs) unsubscribeLogs();
		if (unsubscribeAlerts) unsubscribeAlerts();
		if (unsubscribeProcesses) unsubscribeProcesses(); // Add processes cleanup

		// Disconnect websockets
		websocketManager.disconnect();

		// Cleanup all charts
		Object.values(overviewCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(computeCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(networkCharts).forEach((chart: any) => chart?.destroy?.());
		Object.values(diskCharts).forEach((chart: any) => chart?.destroy?.());
	});
	async function loadEndpointData() {
		try {
			loading = true;
			// Only load endpoint information, not historical data
			// Charts will start empty and populate with live websocket data
			const endpointRes = await api.getEndpoint(endpointId);
			endpoint = endpointRes.data || endpointRes;

			// Initialize empty arrays - charts will populate with live data only
			metrics = [];
			events = [];
			logs = [];
			alerts = [];
		} catch (err) {
			console.error('Error loading endpoint data:', err);
			error = err instanceof Error ? err.message : 'Failed to load endpoint data';
			// Initialize empty arrays on error
			endpoint = null;
			metrics = [];
			events = [];
			logs = [];
			alerts = [];
		} finally {
			loading = false;
		}
	}

	function switchTab(tabId: string) {
		activeTab = tabId;
		// Initialize charts for the specific tab
		setTimeout(() => {
			if (tabId === 'overview') initOverviewCharts();
			else if (tabId === 'compute') initComputeCharts();
			else if (tabId === 'network') initNetworkCharts();
			else if (tabId === 'disk') initDiskCharts();
			else if (tabId === 'console') initConsole();
		}, 100);
	}

	// Helper function to identify metric types
	function isMetricType(
		metric: any,
		type: 'cpu' | 'memory' | 'swap' | 'network_upload' | 'network_download'
	): boolean {
		const metricName = metric.name?.toLowerCase() || '';
		const namespace = metric.namespace?.toLowerCase() || '';
		const subNamespace = metric.subnamespace?.toLowerCase() || '';
		const fullMetricName = `${namespace}.${subNamespace}.${metricName}`;

		switch (type) {
			case 'cpu':
				return (
					metricName === 'cpu_usage' ||
					metricName === 'cpu_percent' ||
					(fullMetricName.includes('cpu') &&
						(metricName.includes('usage') || metricName.includes('percent')))
				);
			case 'memory':
				return (
					metricName === 'memory_usage' ||
					metricName === 'memory_percent' ||
					metricName === 'mem_percent' ||
					(fullMetricName.includes('memory') &&
						(metricName.includes('usage') || metricName.includes('percent')))
				);
			case 'swap':
				return (
					metricName === 'swap_usage' ||
					metricName === 'swap_percent' ||
					(fullMetricName.includes('swap') &&
						(metricName.includes('usage') || metricName.includes('percent')))
				);
			case 'network_upload':
				return (
					metricName === 'network_upload' ||
					metricName === 'net_upload' ||
					metricName === 'tx_bytes' ||
					(fullMetricName.includes('network') &&
						(metricName.includes('upload') || metricName.includes('tx')))
				);
			case 'network_download':
				return (
					metricName === 'network_download' ||
					metricName === 'net_download' ||
					metricName === 'rx_bytes' ||
					(fullMetricName.includes('network') &&
						(metricName.includes('download') || metricName.includes('rx')))
				);
			default:
				return false;
		}
	}

	function initOverviewCharts() {
		console.log('Initializing overview charts...');
		if (typeof window === 'undefined' || !window.ApexCharts) {
			console.log('ApexCharts not available, skipping chart initialization');
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
				swap_mini: [],
				compute_cpu: [],
				compute_memory: [],
				compute_swap: [],
				network_upload: [],
				network_download: [],
				disk_usage: {},
				disk_io_read: [],
				disk_io_write: []
			};
		}

		// Main Performance Metrics Chart
		if (!overviewCharts.main) {
			console.log('Creating main performance chart...');

			// Check if container exists
			const container = document.querySelector('#metrics-chart');
			if (!container) {
				console.error('Main chart container #metrics-chart not found');
				return;
			}

			// Charts start empty and populate with live websocket data only
			const initialCpuData: any[] = [];
			const initialMemoryData: any[] = [];

			// Store initial empty data
			(window as any).chartData.cpu = initialCpuData;
			(window as any).chartData.memory = initialMemoryData;

			const mainOptions = {
				chart: {
					type: 'area',
					height: 300,
					zoom: {
						type: 'x',
						enabled: true,
						autoScaleYaxis: true
					},
					toolbar: {
						autoSelected: 'zoom',
						show: false
					},
					animations: {
						enabled: true,
						easing: 'easeinout',
						speed: 400
					}
				},
				series: [
					{
						name: 'CPU Usage %',
						data: initialCpuData
					},
					{
						name: 'Memory Usage %',
						data: initialMemoryData
					}
				],
				stroke: {
					curve: 'smooth',
					width: 2
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.4,
						opacityTo: 0,
						stops: [0, 90, 100]
					}
				},
				dataLabels: {
					enabled: false
				},
				markers: {
					size: 0
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
						formatter: (val: number) => val.toFixed(1) + '%'
					},
					title: {
						text: 'Usage (%)'
					},
					min: 0,
					max: 100
				},
				colors: ['#3b82f6', '#10b981'],
				legend: {
					show: true,
					position: 'bottom'
				},
				tooltip: {
					shared: true,
					intersect: false,
					x: { format: 'MMM dd HH:mm' },
					y: { formatter: (val: number) => val.toFixed(1) + '%' }
				},
				grid: {
					borderColor: '#e0e0e0',
					strokeDashArray: 4
				}
			};

			try {
				overviewCharts.main = new window.ApexCharts(container, mainOptions);
				overviewCharts.main.render();
				console.log('Main performance chart created and rendered');
			} catch (err) {
				console.error('Error creating main chart:', err);
			}
		} else {
			console.log('Main chart already exists');
		}

		// Mini CPU Chart with process tooltip
		if (!overviewCharts.cpu) {
			const cpuContainer = document.querySelector('#miniCpuChart');
			if (!cpuContainer) {
				console.error('CPU mini chart container #miniCpuChart not found');
				return;
			}

			const cpuOptions = {
				chart: {
					type: 'area',
					height: 280,
					zoom: { enabled: false },
					toolbar: { show: false },
					animations: { enabled: true, easing: 'easeinout', speed: 500 }
				},
				series: [{ name: 'CPU Usage %', data: [] }],
				stroke: { show: true, curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				colors: ['#3b82f6'],
				xaxis: {
					type: 'datetime',
					labels: { format: 'HH:mm:ss' },
					axisBorder: { show: false },
					axisTicks: { show: false }
				},
				yaxis: {
					labels: { formatter: (val: number) => `${val.toFixed(1)}%` }
				},
				dataLabels: {
					enabled: false
				},
				grid: { borderColor: '#e0e0e0', strokeDashArray: 4 },
				tooltip: {
					custom: function ({ series, seriesIndex, dataPointIndex, w }: any) {
						const value = series[seriesIndex][dataPointIndex];
						const processTooltip = getTopProcessesTooltip('cpu');
						return `
                            <div class="
							  bg-white/50            <!-- 50% white background -->
								dark:bg-gray-800/50    <!-- 50% dark background -->
								border
								border-gray-200/50     <!-- 50% light border -->
								dark:border-gray-700/50<!-- 50% dark border -->
								rounded-lg
								p-3
								shadow-lg
								backdrop-blur-sm       <!-- optional: add a slight backdrop blur behind it -->
							">
                                <div class="text-sm font-semibold mb-2">CPU Usage: ${value.toFixed(1)}%</div>
                                ${processTooltip}
                            </div>
                        `;
					}
				}
			};

			try {
				overviewCharts.cpu = new window.ApexCharts(cpuContainer, cpuOptions);
				overviewCharts.cpu.render();
				console.log('Mini CPU chart created');
			} catch (err) {
				console.error('Error creating mini CPU chart:', err);
			}
		}

		// Mini Memory Chart with process tooltip
		if (!overviewCharts.memory) {
			const memContainer = document.querySelector('#miniMemoryChart');
			if (!memContainer) {
				console.error('Memory mini chart container #miniMemoryChart not found');
				return;
			}

			const memOptions = {
				chart: {
					type: 'area',
					height: 280,
					zoom: { enabled: false },
					toolbar: { show: false },
					animations: { enabled: true, easing: 'easeinout', speed: 500 }
				},
				series: [{ name: 'Memory Usage %', data: [] }],
				stroke: { show: true, curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				colors: ['#10b981'],
				xaxis: {
					type: 'datetime',
					labels: { format: 'HH:mm:ss' },
					axisBorder: { show: false },
					axisTicks: { show: false }
				},
				yaxis: {
					labels: { formatter: (val: number) => `${val.toFixed(1)}%` }
				},
				dataLabels: {
					enabled: false
				},
				grid: { borderColor: '#e0e0e0', strokeDashArray: 4 },
				tooltip: {
					custom: function ({ series, seriesIndex, dataPointIndex, w }: any) {
						const value = series[seriesIndex][dataPointIndex];
						const processTooltip = getTopProcessesTooltip('memory');
						return `
                            <div class="
								bg-white/50            <!-- 50% white background -->
								dark:bg-gray-800/50    <!-- 50% dark background -->
								border
								border-gray-200/50     <!-- 50% light border -->
								dark:border-gray-700/50<!-- 50% dark border -->
								rounded-lg
								p-3
								shadow-lg
								backdrop-blur-sm       <!-- optional: add a slight backdrop blur behind it -->
							">
                                <div class="text-sm font-semibold mb-2">Memory Usage: ${value.toFixed(1)}%</div>
                                ${processTooltip}
                            </div>
                        `;
					}
				}
			};

			try {
				overviewCharts.memory = new window.ApexCharts(memContainer, memOptions);
				overviewCharts.memory.render();
				console.log('Mini Memory chart created');
			} catch (err) {
				console.error('Error creating mini Memory chart:', err);
			}
		}

		// Mini Swap Chart
		if (!overviewCharts.swap) {
			const swapContainer = document.querySelector('#miniSwapChart');
			if (!swapContainer) {
				console.error('Swap mini chart container #miniSwapChart not found');
				return;
			}

			const swapOptions = {
				chart: {
					type: 'area',
					height: 280,
					zoom: { enabled: false },
					toolbar: { show: false },
					animations: { enabled: true, easing: 'easeinout', speed: 500 }
				},
				series: [{ name: 'Swap Usage %', data: [] }],
				stroke: { show: true, curve: 'smooth', width: 2 },
				fill: {
					type: 'gradient',
					gradient: { shadeIntensity: 1, opacityFrom: 0.4, opacityTo: 0, stops: [0, 90, 100] }
				},
				colors: ['#f87171'],
				xaxis: {
					type: 'datetime',
					labels: { format: 'HH:mm:ss' },
					axisBorder: { show: false },
					axisTicks: { show: false }
				},
				yaxis: {
					labels: { formatter: (val: number) => `${val.toFixed(1)}%` }
				},
				dataLabels: {
					enabled: false
				},
				grid: { borderColor: '#e0e0e0', strokeDashArray: 4 },
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => `${val.toFixed(1)}%` }
				}
			};

			try {
				overviewCharts.swap = new window.ApexCharts(swapContainer, swapOptions);
				overviewCharts.swap.render();
				console.log('Mini Swap chart created');
			} catch (err) {
				console.error('Error creating mini Swap chart:', err);
			}
		}

		console.log('Overview charts initialization complete');
	}

	function initComputeCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// CPU Usage Chart
		if (!computeCharts.cpuUsage) {
			const cpuUsageOptions = {
				chart: {
					type: 'area',
					height: 250,
					zoom: {
						type: 'x',
						enabled: true,
						autoScaleYaxis: true
					},
					toolbar: {
						autoSelected: 'zoom',
						show: false
					},
					animations: {
						enabled: true
					}
				},
				stroke: {
					curve: 'smooth',
					width: 2
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.4,
						opacityTo: 0,
						stops: [0, 90, 100]
					}
				},
				dataLabels: {
					enabled: false
				},
				markers: {
					size: 0
				},
				series: [{ name: 'CPU Usage %', data: [] }],
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
						formatter: (val: number) => val.toFixed(1) + '%'
					},
					title: {
						text: 'CPU Usage (%)'
					},
					min: 0,
					max: 100
				},
				colors: ['#3b82f6'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					custom: generateProcessTooltip(false)
				}
			};
			computeCharts.cpuUsage = new window.ApexCharts(
				document.querySelector('#cpuUsageChart'),
				cpuUsageOptions
			);
			computeCharts.cpuUsage.render();
		}

		// CPU Load Chart
		if (!computeCharts.cpuLoad) {
			const cpuLoadOptions = {
				chart: {
					type: 'area',
					height: 280,
					toolbar: { show: false },
					animations: {
						enabled: true,
						easing: 'easeinout',
						speed: 400
					}
				},
				stroke: {
					curve: 'smooth',
					width: 3
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.5,
						opacityTo: 0.2,
						stops: [0, 90, 100]
					}
				},
				dataLabels: {
					enabled: false
				},
				markers: {
					size: 0
				},
				series: [
					{ name: '1m', data: [] },
					{ name: '5m', data: [] },
					{ name: '15m', data: [] }
				],
				xaxis: {
					type: 'datetime',
					labels: {
						format: 'HH:mm:ss'
					}
				},
				yaxis: {
					min: 0,
					max: 4,
					tickAmount: 4,
					labels: {
						formatter: (val: number) => val.toFixed(2)
					},
					title: {
						text: 'Load Avg'
					}
				},
				colors: ['#3b82f6', '#10b981', '#f59e0b'],
				legend: {
					position: 'bottom',
					fontSize: '12px'
				},
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: {
						formatter: (val: number) => val.toFixed(2)
					}
				},
				annotations: {
					yaxis: [
						{
							y: 1.0,
							borderColor: '#facc15',
							label: {
								text: 'Warn ≥ 1.0',
								style: { background: '#facc15', color: '#000' }
							}
						},
						{
							y: 1.5,
							borderColor: '#f87171',
							label: {
								text: 'High ≥ 1.5',
								style: { background: '#f87171', color: '#fff' }
							}
						}
					]
				}
			};
			computeCharts.cpuLoad = new window.ApexCharts(
				document.querySelector('#cpuLoadChart'),
				cpuLoadOptions
			);
			computeCharts.cpuLoad.render();
		}

		// Memory Usage Chart
		if (!computeCharts.memoryUsage) {
			const memoryUsageOptions = {
				chart: {
					type: 'area',
					height: 250,
					zoom: {
						type: 'x',
						enabled: true,
						autoScaleYaxis: true
					},
					toolbar: {
						autoSelected: 'zoom',
						show: false
					},
					animations: {
						enabled: true
					}
				},
				stroke: {
					curve: 'smooth',
					width: 2
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.4,
						opacityTo: 0,
						stops: [0, 90, 100]
					}
				},
				dataLabels: {
					enabled: false
				},
				markers: {
					size: 0
				},
				series: [{ name: 'Memory Usage %', data: [] }],
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
						formatter: (val: number) => val.toFixed(1) + '%'
					},
					title: {
						text: 'Memory Usage (%)'
					},
					min: 0,
					max: 100
				},
				colors: ['#10b981'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					custom: generateProcessTooltip(true)
				}
			};
			computeCharts.memoryUsage = new window.ApexCharts(
				document.querySelector('#memoryUsageChart'),
				memoryUsageOptions
			);
			computeCharts.memoryUsage.render();
		}

		// Swap Usage Chart
		if (!computeCharts.swapUsage) {
			const swapUsageOptions = {
				chart: {
					type: 'area',
					height: 250,
					zoom: {
						type: 'x',
						enabled: true,
						autoScaleYaxis: true
					},
					toolbar: {
						autoSelected: 'zoom',
						show: false
					},
					animations: {
						enabled: true
					}
				},
				stroke: {
					curve: 'smooth',
					width: 2
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.4,
						opacityTo: 0,
						stops: [0, 90, 100]
					}
				},
				dataLabels: {
					enabled: false
				},
				markers: {
					size: 0
				},
				series: [{ name: 'Swap Usage %', data: [] }],
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
						formatter: (val: number) => val.toFixed(1) + '%'
					},
					title: {
						text: 'Swap Usage (%)'
					},
					min: 0,
					max: 100
				},
				colors: ['#f87171'],
				tooltip: {
					x: { format: 'HH:mm:ss' },
					y: { formatter: (val: number) => val.toFixed(1) + '%' }
				}
			};
			computeCharts.swapUsage = new window.ApexCharts(
				document.querySelector('#swapUsageChart'),
				swapUsageOptions
			);
			computeCharts.swapUsage.render();
		}
	}

	function initNetworkCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// Network Traffic Chart
		if (!networkCharts.traffic) {
			const trafficOptions = {
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
				xaxis: {
					type: 'datetime',
					labels: {
						format: 'HH:mm:ss',
						style: {
							colors: '#6b7280',
							fontSize: '12px'
						}
					}
				},
				yaxis: {
					min: 0,
					forceNiceScale: true,
					labels: {
						formatter: (val: number) => `${val.toFixed(1)} Mbps`,
						style: {
							colors: '#6b7280',
							fontSize: '12px'
						}
					},
					title: {
						text: 'Bandwidth (Mbps)',
						style: {
							color: '#6b7280',
							fontSize: '12px'
						}
					}
				},
				colors: ['#3b82f6', '#10b981'],
				stroke: {
					curve: 'smooth',
					width: 2
				},
				fill: {
					type: 'gradient',
					gradient: {
						shadeIntensity: 1,
						opacityFrom: 0.7,
						opacityTo: 0.1,
						stops: [0, 100]
					}
				},
				tooltip: {
					theme: 'dark',
					x: {
						format: 'HH:mm:ss'
					},
					y: {
						formatter: (val: number) => `${val.toFixed(2)} Mbps`
					}
				},
				grid: {
					borderColor: '#374151',
					strokeDashArray: 4
				},
				legend: {
					position: 'top',
					horizontalAlign: 'left',
					labels: {
						colors: '#6b7280'
					}
				}
			};
			networkCharts.traffic = new window.ApexCharts(
				document.querySelector('#networkTrafficChart'),
				trafficOptions
			);
			networkCharts.traffic.render();
		}
	}

	function initDiskCharts() {
		if (typeof window === 'undefined' || !window.ApexCharts) return;

		// Disk Usage Donut Chart
		if (!diskCharts.usage) {
			const diskUsageOptions = {
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
								name: {
									show: true,
									fontSize: '14px',
									color: '#6b7280'
								},
								value: {
									show: true,
									fontSize: '20px',
									fontWeight: 600,
									color: '#374151',
									formatter: (val: any) => formatBytes(parseFloat(val))
								},
								total: {
									show: true,
									label: 'Total Space',
									fontSize: '14px',
									color: '#6b7280',
									formatter: () => 'Loading...'
								}
							}
						}
					}
				},
				tooltip: {
					theme: 'dark',
					y: {
						formatter: (val: any, opts: any) => {
							const total = opts.w.config.series.reduce((a: number, b: number) => a + b, 0);
							const pct = total > 0 ? (val / total) * 100 : 0;
							return `${formatBytes(val)} (${pct.toFixed(1)}%)`;
						}
					}
				},
				legend: {
					position: 'bottom',
					horizontalAlign: 'center',
					labels: {
						colors: '#6b7280'
					}
				},
				dataLabels: {
					enabled: true,
					style: {
						colors: ['#fff']
					}
				}
			};
			diskCharts.usage = new window.ApexCharts(
				document.querySelector('#diskUsageDonutChart'),
				diskUsageOptions
			);
			diskCharts.usage.render();
		}

		// Disk Usage Radial Chart for multiple mountpoints
		if (!diskCharts.radial) {
			const diskRadialOptions = {
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
						offsetY: 0,
						hollow: {
							size: '40%',
							background: 'transparent'
						},
						track: {
							background: '#374151'
						},
						dataLabels: {
							name: {
								show: true,
								fontSize: '14px',
								color: '#6b7280'
							},
							value: {
								show: true,
								fontSize: '20px',
								fontWeight: 600,
								color: '#374151',
								formatter: (val: number) => `${val}%`
							},
							total: {
								show: true,
								label: 'Average',
								fontSize: '14px',
								color: '#6b7280',
								formatter: () => '0.0%'
							}
						}
					}
				},
				legend: {
					show: true,
					position: 'bottom',
					horizontalAlign: 'center',
					labels: {
						colors: '#6b7280'
					}
				}
			};
			diskCharts.radial = new window.ApexCharts(
				document.querySelector('#diskRadialChart'),
				diskRadialOptions
			);
			diskCharts.radial.render();
		}
	}

	function setupRealTimeUpdates() {
		console.log('Setting up real-time updates for endpoint:', endpointId);

		// Connect websockets with endpoint filtering (like the original implementation)
		websocketManager.connect(endpointId);

		// Subscribe to metrics updates
		unsubscribeMetrics = websocketManager.subscribeToMetrics((metricsPayload) => {
			//console.log('Received metric update:', metricsPayload);
			if (metricsPayload && metricsPayload.endpoint_id === endpointId) {
				// The payload contains an array of metrics in the 'metrics' property
				if (Array.isArray(metricsPayload.metrics)) {
					// Add individual metrics to our metrics array for display
					const existingMetrics = Array.isArray(metrics) ? metrics : [];
					const newMetrics = metricsPayload.metrics.map((m: any) => ({
						...m,
						endpoint_id: metricsPayload.endpoint_id,
						timestamp: metricsPayload.timestamp
					}));
					metrics = [...newMetrics, ...existingMetrics.slice(0, 99)];
					updateCharts(metricsPayload);
				}
			}
		});

		// Subscribe to events updates
		unsubscribeEvents = websocketManager.subscribeToEvents((latestEvent) => {
			//console.log('Received event update:', latestEvent);
			if (latestEvent && latestEvent.endpoint_id === endpointId) {
				// Ensure events is an array before using slice
				const existingEvents = Array.isArray(events) ? events : [];
				events = [latestEvent, ...existingEvents.slice(0, 49)];
			}
		});

		// Subscribe to logs updates
		unsubscribeLogs = websocketManager.subscribeToLogs((latestLog) => {
			//console.log('Received log update:', latestLog);
			if (latestLog && latestLog.endpoint_id === endpointId) {
				// Ensure logs is an array before using slice
				const existingLogs = Array.isArray(logs) ? logs : [];
				logs = [latestLog, ...existingLogs.slice(0, 49)];
			}
		});

		// Subscribe to alerts updates
		unsubscribeAlerts = websocketManager.subscribeToAlerts((latestAlert) => {
			//console.log('Received alert update:', latestAlert);
			if (latestAlert && latestAlert.endpoint_id === endpointId) {
				// Ensure alerts is an array before using slice
				const existingAlerts = Array.isArray(alerts) ? alerts : [];
				alerts = [latestAlert, ...existingAlerts.slice(0, 19)];
			}
		});

		// Subscribe to processes updates
		unsubscribeProcesses = websocketManager.subscribeToProcesses((processData) => {
			console.log('Received process update:', processData);
			if (processData && processData.endpoint_id === endpointId) {
				if (Array.isArray(processData.processes)) {
					processes = processData.processes;

					// Add to process history for tooltips (similar to vanilla JS)
					const ts = new Date(processData.timestamp).getTime();
					processHistory.push({ timestamp: ts, processes: processData.processes });

					// Keep only last 30 minutes of process history
					const cutoff = Date.now() - 30 * 60 * 1000;
					while (processHistory.length > 0 && processHistory[0].timestamp < cutoff) {
						processHistory.shift();
					}

					updateProcessTables();
				}
			}
		});
	}

	// Function to update process tables
	function updateProcessTables() {
		if (!Array.isArray(processes) || processes.length === 0) return;

		// Sort by CPU and get top 5
		const topCpu = [...processes]
			.sort((a, b) => (parseFloat(b.cpu_percent) || 0) - (parseFloat(a.cpu_percent) || 0))
			.slice(0, 5);

		// Sort by Memory and get top 5
		const topMemory = [...processes]
			.sort((a, b) => (parseFloat(b.memory_percent) || 0) - (parseFloat(a.memory_percent) || 0))
			.slice(0, 5);

		// Update CPU table
		const cpuTable = document.querySelector('#cpu-table tbody');
		if (cpuTable) {
			cpuTable.innerHTML = '';
			topCpu.forEach((proc) => {
				const row = document.createElement('tr');
				row.innerHTML = `
                    <td class="px-3 py-2 text-center font-mono text-xs">${proc.pid || '--'}</td>
                    <td class="px-3 py-2 text-xs">${proc.username || '--'}</td>
                    <td class="px-3 py-2 text-right font-semibold text-xs">${(parseFloat(proc.cpu_percent) || 0).toFixed(1)}%</td>
                    <td class="px-3 py-2 truncate font-mono text-xs" title="${proc.name || '--'}">${(proc.name || '--').slice(0, 20)}</td>
                `;
				cpuTable.appendChild(row);
			});
		}

		// Update Memory table
		const memTable = document.querySelector('#mem-table tbody');
		if (memTable) {
			memTable.innerHTML = '';
			topMemory.forEach((proc) => {
				const row = document.createElement('tr');
				row.innerHTML = `
                    <td class="px-3 py-2 text-center font-mono text-xs">${proc.pid || '--'}</td>
                    <td class="px-3 py-2 text-xs">${proc.username || '--'}</td>
                    <td class="px-3 py-2 text-right font-semibold text-xs">${(parseFloat(proc.memory_percent) || 0).toFixed(1)}%</td>
                    <td class="px-3 py-2 truncate font-mono text-xs" title="${proc.name || '--'}">${(proc.name || '--').slice(0, 20)}</td>
                `;
				memTable.appendChild(row);
			});
		}
	}

	// Function to get top processes for tooltips (enhanced version similar to vanilla JS)
	function getTopProcessesTooltip(type: 'cpu' | 'memory', limit: number = 5): string {
		if (!Array.isArray(processes) || processes.length === 0) {
			return `<div class="text-xs text-gray-500">No process data available</div>`;
		}

		const sortKey = type === 'cpu' ? 'cpu_percent' : 'memory_percent';
		const topProcs = [...processes]
			.sort((a, b) => (parseFloat(b[sortKey]) || 0) - (parseFloat(a[sortKey]) || 0))
			.slice(0, limit);

		const title = type === 'cpu' ? 'Top CPU Processes' : 'Top Memory Processes';

		// Create table-based tooltip similar to vanilla JS
		let html = `
			<div class="text-xs font-semibold mb-2 flex justify-between items-center">
				<span>${title}</span>
				<span>Total: ${getCurrentUsage(type).toFixed(1)}%</span>
			</div>
			<table style="width:100%; border-collapse:collapse; font-size:11px;">
				<thead>
					<tr style="color:#9ca3af;">
						<th style="padding:4px 6px; text-align:left;">PID</th>
						<th style="padding:4px 6px; text-align:left;">Command</th>
						<th style="padding:4px 6px; text-align:right;">Usage</th>
					</tr>
				</thead>
				<tbody>`;

		topProcs.forEach((proc) => {
			const value = (parseFloat(proc[sortKey]) || 0).toFixed(1);
			const command = truncateCommand(proc.name || proc.executable || '(?)', 25);
			const color = type === 'cpu' ? '#3b82f6' : '#10b981';

			html += `
				<tr style="border-bottom:1px solid #e5e7eb;">
					<td style="padding:4px 6px; color:#6b7280;">${proc.pid || '?'}</td>
					<td style="padding:4px 6px; max-width:120px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" title="${proc.name || '?'}">${command}</td>
					<td style="padding:4px 6px; text-align:right; font-weight:500; color:${color};">${value}%</td>
				</tr>`;
		});

		html += `</tbody></table>`;
		return html;
	}

	// Enhanced tooltip function that uses process history (like vanilla JS)
	function generateProcessTooltip(isMem: boolean) {
		return function ({ series, seriesIndex, dataPointIndex, w }: any) {
			const point = w.config.series[seriesIndex].data[dataPointIndex];
			if (!point || !point.x) return '';

			const hoverTime = point.x;
			const snapshot = findClosestSnapshot(hoverTime);
			if (!snapshot || !snapshot.processes) return 'No process data';

			const labelKey = isMem ? 'memory_percent' : 'cpu_percent';
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
				<div style="background:rgba(255,255,255,0.95); border:1px solid #e5e7eb; border-radius:8px; padding:8px; box-shadow:0 4px 6px -1px rgba(0,0,0,0.1);">
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

	function updateCharts(metricsPayload: any) {
		if (!metricsPayload || !Array.isArray(metricsPayload.metrics)) {
			console.log('No metrics array found in payload');
			return;
		}

		const timestamp = new Date(metricsPayload.timestamp).getTime();
		//console.log('Updating charts with timestamp:', timestamp);
		//console.log('Available chart instances:', {
		//	main: !!overviewCharts.main,
		//	cpu: !!overviewCharts.cpu,
		//	memory: !!overviewCharts.memory,
		//		swap: !!overviewCharts.swap
		//});

		// Store current data for charts (using global state)
		if (!(window as any).chartData) {
			(window as any).chartData = {
				cpu: [],
				memory: [],
				swap: [],
				cpu_mini: [],
				memory_mini: [],
				swap_mini: [],
				compute_cpu: [],
				network_upload: [],
				network_download: []
			};
		}

		// Process each metric in the array
		metricsPayload.metrics.forEach((metric: any) => {
			const metricValue = parseFloat(metric.value);
			const metricName = metric.name;
			const namespace = metric.namespace?.toLowerCase() || '';
			const subNamespace = metric.subnamespace?.toLowerCase() || '';
			const dimensions = metric.dimensions || {};

			//console.log(`Processing metric: ${namespace}.${subNamespace}.${metricName} = ${metricValue}`);

			// Create a full metric identifier for better matching
			const fullMetricName = `${namespace}.${subNamespace}.${metricName}`;

			// Update percentage labels in the UI based on various possible metric names
			const isCpuMetric =
				metricName === 'usage_percent' && subNamespace === 'cpu' && dimensions?.scope === 'total';

			const isMemoryMetric = metricName === 'used_percent' && subNamespace === 'memory';

			const isSwapMetric =
				['swap_used_percent', 'swap_total', 'swap_free', 'swap_used'].includes(metricName) &&
				subNamespace === 'memory';

			const isNetworkUploadMetric =
				metricName === 'network_upload' ||
				metricName === 'net_upload' ||
				metricName === 'tx_bytes' ||
				(fullMetricName.includes('network') &&
					(metricName.includes('upload') || metricName.includes('tx')));
			const isNetworkDownloadMetric =
				metricName === 'network_download' ||
				metricName === 'net_download' ||
				metricName === 'rx_bytes' ||
				(fullMetricName.includes('network') &&
					(metricName.includes('download') || metricName.includes('rx')));

			if (isCpuMetric) {
				const label = document.getElementById('cpu-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
				latestCpuPercent = metricValue; // Update latest CPU percentage
			}
			if (isMemoryMetric) {
				const label = document.getElementById('mem-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
				latestMemUsedPercent = metricValue; // Update latest memory percentage
			}
			if (isSwapMetric) {
				const label = document.getElementById('swap-percent-label');
				if (label) label.textContent = `${metricValue.toFixed(1)}%`;
			}

			// Update main performance metrics chart in overview
			if (overviewCharts.main && (isCpuMetric || isMemoryMetric)) {
				//console.log('Updating main chart for metric:', fullMetricName);

				if (isCpuMetric) {
					// Store CPU data
					(window as any).chartData.cpu.push({ x: timestamp, y: metricValue });
					(window as any).chartData.cpu = (window as any).chartData.cpu.slice(-50); // Keep last 50 points
				}
				if (isMemoryMetric) {
					// Store Memory data
					(window as any).chartData.memory.push({ x: timestamp, y: metricValue });
					(window as any).chartData.memory = (window as any).chartData.memory.slice(-50); // Keep last 50 points
				}

				// Always update both series to maintain chart integrity
				try {
					overviewCharts.main.updateSeries(
						[
							{
								name: 'CPU Usage %',
								data: (window as any).chartData.cpu
							},
							{
								name: 'Memory Usage %',
								data: (window as any).chartData.memory
							}
						],
						false
					); // false = don't redraw each update
					/*
					console.log(
						'Updated main chart series with',
						(window as any).chartData.cpu.length,
						'CPU points and',
						(window as any).chartData.memory.length,
						'memory points'
					);
					*/
				} catch (err) {
					console.error('Error updating main chart:', err);
				}
			}

			// Update overview mini charts using updateSeries (more reliable than appendData)
			if (isCpuMetric && overviewCharts.cpu) {
				//console.log('Updating mini CPU chart');
				try {
					(window as any).chartData.cpu_mini = (window as any).chartData.cpu_mini || [];
					(window as any).chartData.cpu_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.cpu_mini = (window as any).chartData.cpu_mini.slice(-20); // Keep last 20 points

					overviewCharts.cpu.updateSeries(
						[
							{
								name: 'CPU',
								data: (window as any).chartData.cpu_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini CPU chart:', err);
				}
			}
			if (isMemoryMetric && overviewCharts.memory) {
				//console.log('Updating mini Memory chart');
				try {
					(window as any).chartData.memory_mini = (window as any).chartData.memory_mini || [];
					(window as any).chartData.memory_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.memory_mini = (window as any).chartData.memory_mini.slice(-20); // Keep last 20 points

					overviewCharts.memory.updateSeries(
						[
							{
								name: 'Memory',
								data: (window as any).chartData.memory_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini Memory chart:', err);
				}
			}
			if (isSwapMetric && overviewCharts.swap) {
				//console.log('Updating mini Swap chart');
				try {
					(window as any).chartData.swap_mini = (window as any).chartData.swap_mini || [];
					(window as any).chartData.swap_mini.push({ x: timestamp, y: metricValue });
					(window as any).chartData.swap_mini = (window as any).chartData.swap_mini.slice(-20); // Keep last 20 points

					overviewCharts.swap.updateSeries(
						[
							{
								name: 'Swap',
								data: (window as any).chartData.swap_mini
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating mini Swap chart:', err);
				}
			}

			// Update compute charts
			if (isCpuMetric && computeCharts.cpuUsage) {
				try {
					(window as any).chartData.compute_cpu = (window as any).chartData.compute_cpu || [];
					(window as any).chartData.compute_cpu.push({ x: timestamp, y: metricValue });
					(window as any).chartData.compute_cpu = (window as any).chartData.compute_cpu.slice(-50);

					computeCharts.cpuUsage.updateSeries(
						[
							{
								name: 'CPU Usage %',
								data: (window as any).chartData.compute_cpu
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating compute CPU chart:', err);
				}
			}

			if (isMemoryMetric && computeCharts.memoryUsage) {
				try {
					(window as any).chartData.compute_memory = (window as any).chartData.compute_memory || [];
					(window as any).chartData.compute_memory.push({ x: timestamp, y: metricValue });
					(window as any).chartData.compute_memory = (window as any).chartData.compute_memory.slice(
						-50
					);

					computeCharts.memoryUsage.updateSeries(
						[
							{
								name: 'Memory Usage %',
								data: (window as any).chartData.compute_memory
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating compute Memory chart:', err);
				}
			}

			if (isSwapMetric && computeCharts.swapUsage) {
				try {
					(window as any).chartData.compute_swap = (window as any).chartData.compute_swap || [];
					(window as any).chartData.compute_swap.push({ x: timestamp, y: metricValue });
					(window as any).chartData.compute_swap = (window as any).chartData.compute_swap.slice(
						-50
					);

					computeCharts.swapUsage.updateSeries(
						[
							{
								name: 'Swap Usage %',
								data: (window as any).chartData.compute_swap
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating compute Swap chart:', err);
				}
			}

			// Update network charts
			if (isNetworkUploadMetric && networkCharts.traffic) {
				try {
					(window as any).chartData.network_upload = (window as any).chartData.network_upload || [];
					// Convert bytes to Mbps if needed (assuming value is in bytes/sec)
					const mbpsValue = metricValue > 1000000 ? metricValue / (1024 * 1024) : metricValue;
					(window as any).chartData.network_upload.push({ x: timestamp, y: mbpsValue });
					(window as any).chartData.network_upload = (window as any).chartData.network_upload.slice(
						-50
					);

					networkCharts.traffic.updateSeries(
						[
							{
								name: 'Upload (Mbps)',
								data: (window as any).chartData.network_upload
							},
							{
								name: 'Download (Mbps)',
								data: (window as any).chartData.network_download || []
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating network upload chart:', err);
				}
			}

			if (isNetworkDownloadMetric && networkCharts.traffic) {
				try {
					(window as any).chartData.network_download =
						(window as any).chartData.network_download || [];
					// Convert bytes to Mbps if needed (assuming value is in bytes/sec)
					const mbpsValue = metricValue > 1000000 ? metricValue / (1024 * 1024) : metricValue;
					(window as any).chartData.network_download.push({ x: timestamp, y: mbpsValue });
					(window as any).chartData.network_download = (
						window as any
					).chartData.network_download.slice(-50);

					networkCharts.traffic.updateSeries(
						[
							{
								name: 'Upload (Mbps)',
								data: (window as any).chartData.network_upload || []
							},
							{
								name: 'Download (Mbps)',
								data: (window as any).chartData.network_download
							}
						],
						false
					);
				} catch (err) {
					console.error('Error updating network download chart:', err);
				}
			}

			// Update disk metrics for usage charts
			const isDiskUsageMetric =
				metricName === 'disk_usage' ||
				metricName === 'disk_used' ||
				metricName === 'used_bytes' ||
				(namespace === 'system' &&
					subNamespace === 'disk' &&
					(metricName.includes('usage') || metricName.includes('used')));

			const isDiskTotalMetric =
				metricName === 'disk_total' ||
				metricName === 'total_bytes' ||
				(namespace === 'system' && subNamespace === 'disk' && metricName.includes('total'));

			const mountPoint = metric.dimensions?.mount_point || metric.dimensions?.device || '/';

			if (isDiskUsageMetric || isDiskTotalMetric) {
				console.log(`Processing disk metric: ${metricName} for ${mountPoint} = ${metricValue}`);

				if (!(window as any).chartData.disk_usage[mountPoint]) {
					(window as any).chartData.disk_usage[mountPoint] = { used: 0, total: 0 };
				}

				if (isDiskUsageMetric) {
					(window as any).chartData.disk_usage[mountPoint].used = metricValue;
				}
				if (isDiskTotalMetric) {
					(window as any).chartData.disk_usage[mountPoint].total = metricValue;
				}

				// Update donut chart if we have both used and total
				const diskData = (window as any).chartData.disk_usage[mountPoint];
				if (diskData.used > 0 && diskData.total > 0 && diskCharts.usage) {
					try {
						const usedBytes = diskData.used;
						const totalBytes = diskData.total;
						const freeBytes = totalBytes - usedBytes;
						const usagePercent = (usedBytes / totalBytes) * 100;

						// Update stat cards
						const totalElement = document.getElementById('disk-total');
						const usedElement = document.getElementById('disk-used');
						const freeElement = document.getElementById('disk-free');
						const percentElement = document.getElementById('disk-percent');

						if (totalElement) totalElement.textContent = formatBytes(totalBytes);
						if (usedElement) usedElement.textContent = formatBytes(usedBytes);
						if (freeElement) freeElement.textContent = formatBytes(freeBytes);
						if (percentElement) percentElement.textContent = usagePercent.toFixed(1);

						// Update donut chart
						diskCharts.usage.updateSeries([usedBytes, freeBytes]);
						diskCharts.usage.updateOptions({
							plotOptions: {
								pie: {
									donut: {
										labels: {
											total: {
												formatter: () => `${usagePercent.toFixed(1)}%`
											}
										}
									}
								}
							}
						});
					} catch (err) {
						console.error('Error updating disk usage chart:', err);
					}
				}

				// Update radial chart for multiple mountpoints
				if (diskCharts.radial) {
					try {
						const allMountPoints = Object.keys((window as any).chartData.disk_usage);
						const series: number[] = [];
						const labels: string[] = [];

						allMountPoints.forEach((mount) => {
							const data = (window as any).chartData.disk_usage[mount];
							if (data.used > 0 && data.total > 0) {
								const percent = (data.used / data.total) * 100;
								series.push(Math.round(percent));
								labels.push(mount.length > 12 ? mount.slice(0, 10) + '…' : mount);
							}
						});

						if (series.length > 0) {
							const avgPercent = (
								series.reduce((a: number, b: number) => a + b, 0) / series.length
							).toFixed(1);

							diskCharts.radial.updateSeries(series);
							diskCharts.radial.updateOptions({
								labels: labels,
								plotOptions: {
									radialBar: {
										dataLabels: {
											total: {
												formatter: () => `${avgPercent}%`
											}
										}
									}
								}
							});
						}
					} catch (err) {
						console.error('Error updating disk radial chart:', err);
					}
				}
			}
		});

		// Force chart redraws with proper ApexCharts methods
		setTimeout(() => {
			try {
				// Use updateOptions to force a refresh instead of redrawPaths
				if (overviewCharts.main) {
					overviewCharts.main.updateOptions({}, true, true); // redrawPaths: true, animate: true
				}
				if (overviewCharts.cpu) {
					overviewCharts.cpu.updateOptions({}, true, true);
				}
				if (overviewCharts.memory) {
					overviewCharts.memory.updateOptions({}, true, true);
				}
				if (overviewCharts.swap) {
					overviewCharts.swap.updateOptions({}, true, true);
				}
				if (computeCharts.cpuUsage) {
					computeCharts.cpuUsage.updateOptions({}, true, true);
				}
			} catch (err) {
				console.error('Error refreshing charts:', err);
			}
		}, 50);
	}

	async function runCommand(command: string) {
		try {
			await api.sendCommand(endpointId, { command, args: [] });
		} catch (err) {
			console.error('Failed to run command:', err);
		}
	}

	function getBadgeClass(value: string): string {
		// Try status class first, then level class
		const statusClass = getStatusBadgeClass(value);
		if (statusClass !== 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200') {
			return statusClass;
		}
		return getLevelBadgeClass(value);
	}

	// Initialize console when tab is switched
	function initConsole() {
		setTimeout(() => {
			const input = document.getElementById('console-command') as HTMLInputElement;
			const responsesEl = document.getElementById('console-responses');

			if (!input || !responsesEl || input.dataset.bound) return;
			input.dataset.bound = 'true';

			input.addEventListener('keydown', async (e) => {
				if (e.key !== 'Enter') return;

				const cmd = input.value.trim();
				if (!cmd) return;

				// Echo command
				const echo = document.createElement('div');
				echo.innerHTML = `<span class="text-blue-400">user</span>@<span class="text-purple-400">${endpoint?.hostname || 'host'}</span>:<span class="text-red-400">~</span>$ <span class="text-green-400">${cmd}</span>`;
				responsesEl.appendChild(echo);

				// Show placeholder while waiting
				const pending = document.createElement('div');
				pending.className = 'text-gray-500 whitespace-pre-wrap';
				pending.textContent = '[executing...]';
				responsesEl.appendChild(pending);

				input.value = '';

				try {
					await runCommand(cmd);
					// Response will arrive via websocket
				} catch (err) {
					pending.className = 'text-red-400';
					pending.textContent = '❌ ' + (err instanceof Error ? err.message : 'Command failed');
				}

				// Scroll to bottom
				const output = document.getElementById('console-output');
				if (output) output.scrollTop = output.scrollHeight;
			});
		}, 100);
	}
</script>

<svelte:head>
	<title>Endpoint {endpoint?.name || endpointId} - GoSight</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<!-- Header -->
	<div class="bg-white shadow dark:bg-gray-800">
		<div class="px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 items-center justify-between">
				<div class="flex items-center">
					<a href="/endpoints" class="mr-4 rounded-md p-2 hover:bg-gray-100 dark:hover:bg-gray-700">
						<ChevronLeft size={20} />
					</a>
					<div>
						<h1 class="text-xl font-semibold text-gray-900 dark:text-white">
							{endpoint?.hostname || 'Loading...'}
						</h1>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							{endpoint?.ip || ''}
						</p>
					</div>
				</div>
				<div class="flex items-center space-x-2">
					<span
						class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getBadgeClass(
							endpoint?.status || 'unknown'
						)}"
					>
						{endpoint?.status || 'Unknown'}
					</span>
					{#if endpoint?.os}
						<span class="text-sm text-gray-500 dark:text-gray-400">{endpoint.os}</span>
					{/if}
				</div>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="flex h-64 items-center justify-center">
			<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div class="p-6">
			<div
				class="rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
			>
				<p class="text-red-800 dark:text-red-200">{error}</p>
			</div>
		</div>
	{:else if endpoint}
		<!-- Tabs -->
		<div class="border-b border-gray-200 dark:border-gray-700">
			<nav class="-mb-px flex space-x-8 px-4 sm:px-6 lg:px-8" id="dashboardTabs" role="tablist">
				{#each [{ id: 'overview', label: 'Overview', icon: Activity }, { id: 'compute', label: 'Compute', icon: Cpu }, { id: 'disk', label: 'Disk', icon: HardDrive }, { id: 'network', label: 'Network', icon: Wifi }, { id: 'activity', label: 'Activity', icon: Monitor }, { id: 'logs', label: 'Logs', icon: ScrollText }, { id: 'console', label: 'Console', icon: Terminal }] as tab}
					<button
						class="border-b-2 px-1 py-4 text-sm font-medium {activeTab === tab.id
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'}"
						on:click={() => switchTab(tab.id)}
						role="tab"
						aria-controls={tab.id}
						aria-selected={activeTab === tab.id}
					>
						<div class="flex items-center">
							<svelte:component this={tab.icon} size={16} class="mr-2" />
							{tab.label}
						</div>
					</button>
				{/each}
			</nav>
		</div>

		<!-- Content -->
		<div class="bg-gray-50 dark:bg-gray-800" id="dashboardTabContent">
			<!-- OVERVIEW TAB -->
			{#if activeTab === 'overview'}
				<div class="p-4" id="overview" role="tabpanel" aria-labelledby="overview-tab">
					<!-- System Info and Metrics Section -->
					<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
						<!-- Info Cards -->
						<div class="space-y-6 lg:col-span-2">
							<!-- Basic Info -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									System Information
								</h3>
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
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
											Operating System
										</dt>
										<dd class="text-sm text-gray-900 dark:text-white">{endpoint.os || 'N/A'}</dd>
									</div>
									<div>
										<dt class="text-sm font-medium text-gray-500 dark:text-gray-400">
											Agent Version
										</dt>
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
											{endpoint.uptime ? formatDuration(endpoint.uptime) : 'N/A'}
										</dd>
									</div>
								</dl>
							</div>

							<!-- Metrics Chart -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Performance Metrics
								</h3>
								<div id="metrics-chart"></div>
							</div>
						</div>

						<!-- Sidebar -->
						<div class="space-y-6">
							<!-- Quick Actions -->
							<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Quick Actions
								</h3>
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
								<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
									Recent Alerts
								</h3>
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
									id="cpu-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniCpuChart" class="h-20 w-full"></div>
							</div>
						</div>

						<div
							class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 dark:text-gray-400">Memory Used</p>
								<p
									class="text-2xl font-bold text-green-600 dark:text-green-400"
									id="mem-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniMemoryChart" class="h-20 w-full"></div>
							</div>
						</div>

						<div
							class="flex h-full flex-col justify-between rounded-lg border border-gray-100 bg-white p-4 shadow-sm hover:shadow-md sm:p-6 dark:border-gray-700 dark:bg-gray-800"
						>
							<div class="flex items-center justify-between">
								<p class="text-sm text-gray-500 dark:text-gray-400">Swap Used</p>
								<p
									class="text-2xl font-bold text-yellow-500 dark:text-yellow-400"
									id="swap-percent-label"
								>
									--%
								</p>
							</div>
							<p class="mb-1 text-xs text-gray-400 dark:text-gray-500">percent</p>
							<div class="mt-2">
								<div id="miniSwapChart" class="h-20 w-full"></div>
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
								id="log-stream"
								class="h-full space-y-2 overflow-y-auto rounded-md border border-gray-200 bg-gray-50 p-3 font-mono text-sm break-words whitespace-pre-wrap text-gray-800 shadow-inner dark:border-gray-600 dark:bg-gray-800 dark:text-gray-200"
							>
								{#each Array.isArray(logs) ? logs.slice(0, 10) : [] as log}
									<div class="text-xs">
										<span class="text-gray-500 dark:text-gray-400"
											>[{formatDate(log.timestamp)}]</span
										>
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
								<table
									id="cpu-table"
									class="w-full text-left text-sm text-gray-700 dark:text-gray-200"
								>
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
									<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
										<tr
											><td colspan="4" class="px-3 py-4 text-center text-gray-500"
												>Loading processes...</td
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
								<table
									id="mem-table"
									class="w-full text-left text-sm text-gray-700 dark:text-gray-200"
								>
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
									<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
										<tr
											><td colspan="4" class="px-3 py-4 text-center text-gray-500"
												>Loading processes...</td
											></tr
										>
									</tbody>
								</table>
							</div>
						</div>
					</div>
				</div>

				<!-- COMPUTE TAB -->
			{:else if activeTab === 'compute'}
				<div class="p-4" id="compute" role="tabpanel" aria-labelledby="compute-tab">
					<!-- Compute Resources Section -->
					<div class="mb-6 rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
						<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Compute Resources</h2>
						<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
							<!-- CPU Usage Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									CPU Usage Over Time
								</h3>
								<div id="cpuUsageChart" class="mt-2 h-32"></div>
							</div>

							<!-- CPU Load Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									CPU Load Average
								</h3>
								<div id="cpuLoadChart" class="mt-2 h-32"></div>
							</div>

							<!-- Memory Usage Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									Memory Usage Over Time
								</h3>
								<div id="memoryUsageChart" class="mt-2 h-32"></div>
							</div>

							<!-- Swap Usage Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									Swap Usage Over Time
								</h3>
								<div id="swapUsageChart" class="mt-2 h-32"></div>
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
								<tbody class="bg-white dark:bg-gray-900">
									<!-- Process rows will be populated by JavaScript -->
									<tr>
										<td
											colspan="5"
											class="px-3 py-2 text-center text-xs text-gray-500 dark:text-gray-400"
										>
											Loading processes...
										</td>
									</tr>
								</tbody>
							</table>
						</div>
					</div>
				</div>

				<!-- DISK TAB -->
			{:else if activeTab === 'disk'}
				<div class="p-4" id="disk" role="tabpanel" aria-labelledby="disk-tab">
					<!-- Disk Usage Section -->
					<div class="mb-6 rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
						<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Disk Usage</h2>
						<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
							<!-- Disk Usage Donut Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									Disk Usage Overview
								</h3>
								<div id="diskUsageDonutChart" class="mt-2 h-32"></div>
							</div>

							<!-- Disk Usage Radial Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									Disk Usage by Mount Point
								</h3>
								<div id="diskRadialChart" class="mt-2 h-32"></div>
							</div>
						</div>
					</div>

					<!-- Disk I/O Section -->
					<div class="mb-6 rounded-lg bg-white p-4 shadow-md dark:bg-gray-800">
						<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Disk I/O</h2>
						<div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
							<!-- Disk Read/Write Chart -->
							<div class="rounded-lg bg-gray-50 p-4 dark:bg-gray-900">
								<h3 class="text-xs font-semibold text-gray-900 dark:text-white">
									Disk Read/Write Over Time
								</h3>
								<div id="diskIoChart" class="mt-2 h-32"></div>
							</div>
						</div>
					</div>
				</div>

				<!-- NETWORK TAB -->
			{:else if activeTab === 'network'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="network"
					role="tabpanel"
					aria-labelledby="network-tab"
				>
					<div class="space-y-6">
						<!-- Network Interface Stats -->
						<div
							class="rounded-lg border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-900"
						>
							<h3 class="mb-4 text-lg font-semibold text-gray-800 dark:text-white">
								Network Traffic
							</h3>
							<div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-3">
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Current Upload</p>
									<p class="text-xl font-bold text-red-600 dark:text-red-400" id="current-tx">
										-- Mbps
									</p>
								</div>
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">
										Current Download
									</p>
									<p class="text-xl font-bold text-blue-600 dark:text-blue-400" id="current-rx">
										-- Mbps
									</p>
								</div>
								<div class="text-center">
									<p class="text-sm font-medium text-gray-500 dark:text-gray-400">Peak Bandwidth</p>
									<p class="text-sm text-gray-600 dark:text-gray-300" id="peak-bandwidth">
										↑ -- / ↓ --
									</p>
								</div>
							</div>
							<div id="networkTrafficChart" class="h-80 w-full"></div>
						</div>
					</div>
				</div>

				<!-- ACTIVITY TAB -->
			{:else if activeTab === 'activity'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="activity"
					role="tabpanel"
					aria-labelledby="activity-tab"
				>
					<div class="rounded-lg bg-white shadow dark:bg-gray-800">
						<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">Recent Events</h3>
						</div>
						<div class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each Array.isArray(events) ? events : [] as event}
								<div class="px-6 py-4">
									<div class="flex items-center justify-between">
										<div>
											<h4 class="text-sm font-medium text-gray-900 dark:text-white">
												{event.category || event.type}
											</h4>
											<p class="text-sm text-gray-500 dark:text-gray-400">
												{event.message}
											</p>
										</div>
										<div class="text-right">
											<p class="text-xs text-gray-500 dark:text-gray-400">
												{formatDate(event.timestamp)}
											</p>
										</div>
									</div>
								</div>
							{:else}
								<div class="px-6 py-8 text-center">
									<p class="text-gray-500 dark:text-gray-400">No events found</p>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<!-- LOGS TAB -->
			{:else if activeTab === 'logs'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="logs"
					role="tabpanel"
					aria-labelledby="logs-tab"
				>
					<div class="rounded-lg bg-white shadow dark:bg-gray-800">
						<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
							<h3 class="text-lg font-medium text-gray-900 dark:text-white">System Logs</h3>
						</div>
						<div class="divide-y divide-gray-200 dark:divide-gray-700">
							{#each Array.isArray(logs) ? logs : [] as log}
								<div class="px-6 py-3">
									<div class="flex items-start space-x-3">
										<span
											class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {getBadgeClass(
												log.level
											)}"
										>
											{log.level}
										</span>
										<div class="min-w-0 flex-1">
											<p class="text-sm break-words text-gray-900 dark:text-white">{log.message}</p>
											<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
												{formatDate(log.timestamp)}
											</p>
										</div>
									</div>
								</div>
							{:else}
								<div class="px-6 py-8 text-center">
									<p class="text-gray-500 dark:text-gray-400">No logs found</p>
								</div>
							{/each}
						</div>
					</div>
				</div>

				<!-- CONSOLE TAB -->
			{:else if activeTab === 'console'}
				<div
					class="rounded-lg bg-gray-50 p-4 dark:bg-gray-800"
					id="console"
					role="tabpanel"
					aria-labelledby="console-tab"
				>
					<div class="rounded-lg bg-black p-4 font-mono text-sm text-green-400 shadow-lg">
						<div class="mb-4">
							<div id="console-prompt" class="mb-2">
								<span class="text-blue-400">user</span>@<span class="text-purple-400"
									>{endpoint?.hostname || 'host'}</span
								>:<span class="text-red-400">~</span><span class="text-white">$</span><span
									class="blink-cursor"
								></span>
							</div>
						</div>
						<div id="console-output" class="mb-4 h-96 overflow-y-auto">
							<div id="console-responses" class="space-y-1">
								<div class="text-gray-400">
									Welcome to {endpoint?.hostname || 'remote'} console. Type commands below:
								</div>
							</div>
						</div>
						<div class="flex items-center">
							<span class="text-blue-400">user</span>@<span class="text-purple-400"
								>{endpoint?.hostname || 'host'}</span
							>:<span class="text-red-400">~</span><span class="text-white">$ </span>
							<input
								id="console-command"
								class="ml-1 flex-1 border-none bg-transparent text-green-400 outline-none"
								type="text"
								placeholder="Enter command..."
								autocomplete="off"
							/>
						</div>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
