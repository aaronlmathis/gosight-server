<script lang="ts">
	import { onMount, tick } from 'svelte';
	import type { Endpoint, Metric, AlertRuleFormData } from '$lib/types';
	import { formatDate, formatDuration, getBadgeClass } from '$lib/utils';
	import { api } from '$lib/api';
	import {
		AlertTriangle,
		Settings,
		Play,
		Tag,
		Plus,
		Zap,
		RefreshCw,
		RotateCcw,
		Activity,
		Server,
		Globe,
		Monitor,
		Clock,
		Users,
		Layers,
		Cpu
	} from 'lucide-svelte';

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
	export let events: any[] = [];
	export let runCommand: (command: string) => void;

	// Modal state for alert rule creation
	let showCreateModal = false;

	// Form data for quick alert rule creation
	let formData: AlertRuleFormData = {
		name: '',
		description: '',
		severity: 'warning',
		enabled: true,
		endpoint_id: '',
		metric_name: '',
		operator: 'gt',
		threshold: 0,
		duration: 300
	};

	// Element bindings
	let mainCpuTableBody!: HTMLTableSectionElement;
	let mainMemTableBody!: HTMLTableSectionElement;

	// Current metrics values
	let cpuPercent = 0;
	let memoryPercent = 0;
	let swapPercent = 0;

	// Modal and alert rule functions
	function openCreateModal() {
		formData = {
			name: '',
			description: '',
			severity: 'warning',
			enabled: true,
			endpoint_id: endpoint.id,
			metric_name: '',
			operator: 'gt',
			threshold: 0,
			duration: 300
		};
		showCreateModal = true;
	}

	function closeCreateModal() {
		showCreateModal = false;
	}

	async function saveRule() {
		try {
			const alertRuleData = {
				name: formData.name,
				description: formData.description,
				message: `Alert for ${formData.metric_name || 'metric'}`,
				level: formData.severity,
				enabled: formData.enabled,
				type: 'metric',
				match: {
					endpoint_ids: [formData.endpoint_id],
					labels: {}
				},
				scope: {
					namespace: 'system',
					subnamespace: '',
					metric: formData.metric_name
				},
				expression: {
					datatype: 'value',
					operator: formData.operator,
					value: formData.threshold
				},
				actions: [],
				options: {
					cooldown: '30s',
					eval_interval: '10s',
					repeat_interval: '1m',
					notify_on_resolve: true
				}
			};

			await api.alerts.createRule(alertRuleData);
			closeCreateModal();
			// Show success message or refresh alerts if needed
		} catch (err) {
			console.error('Failed to create alert rule:', err);
			alert('Failed to create alert rule: ' + (err as Error).message);
		}
	}

	// Placeholder functions for future features
	function runPlaybook() {
		alert('Run Playbook functionality coming soon!');
	}

	function addTag() {
		alert('Add Tag functionality coming soon!');
	}

	// Extract current metric values for display
	function updateMetricValues(metrics: Metric[]) {
		if (!metrics || metrics.length === 0) return;

		// Find the latest metrics for each type
		const latestMetrics = metrics.reduce(
			(acc, metric) => {
				const key = metric.name;
				if (!acc[key] || new Date(metric.timestamp) > new Date(acc[key].timestamp)) {
					acc[key] = metric;
				}
				return acc;
			},
			{} as Record<string, Metric>
		);

		// Update percentage values
		cpuPercent = latestMetrics.cpu_usage?.value || latestMetrics.cpu_percent?.value || 0;
		memoryPercent = latestMetrics.memory_usage?.value || latestMetrics.memory_percent?.value || 0;
		swapPercent = latestMetrics.swap_usage?.value || latestMetrics.swap_percent?.value || 0;
	}

	// Reactive update: when new metrics arrive, update values
	$: if (metrics.length > 0) {
		updateMetricValues(metrics);
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
		// Chart initialization removed - overview tab now focuses on system info and processes
	});
</script>

<div class="p-4" id="overview" role="tabpanel" aria-labelledby="overview-tab">
	<!-- System Info and Metrics Section -->
	<div class="mb-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
		<!-- Info Cards -->
		<div class="space-y-6 lg:col-span-2">
			<!-- Basic Info -->
			<div
				class="rounded-lg border border-gray-100 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-800"
			>
				<div class="mb-6 flex items-center space-x-2">
					<Server size={20} class="text-blue-500" />
					<h3 class="text-lg font-semibold text-gray-900 dark:text-white">System Information</h3>
				</div>

				<!-- Primary System Info Grid -->
				<div class="mb-6 grid grid-cols-1 gap-6 sm:grid-cols-2">
					<!-- Hostname Card -->
					<div
						class="rounded-lg border border-blue-200 bg-gradient-to-br from-blue-50 to-blue-100 p-4 dark:border-blue-700/50 dark:from-blue-900/20 dark:to-blue-800/20"
					>
						<div class="flex items-center space-x-3">
							<div class="flex-shrink-0">
								<Monitor size={18} class="text-blue-600 dark:text-blue-400" />
							</div>
							<div class="min-w-0 flex-1">
								<dt
									class="text-xs font-medium tracking-wide text-blue-700 uppercase dark:text-blue-300"
								>
									Hostname
								</dt>
								<dd class="truncate text-lg font-semibold text-blue-900 dark:text-blue-100">
									{endpoint.hostname || 'N/A'}
								</dd>
							</div>
						</div>
					</div>

					<!-- IP Address Card -->
					<div
						class="rounded-lg border border-green-200 bg-gradient-to-br from-green-50 to-green-100 p-4 dark:border-green-700/50 dark:from-green-900/20 dark:to-green-800/20"
					>
						<div class="flex items-center space-x-3">
							<div class="flex-shrink-0">
								<Globe size={18} class="text-green-600 dark:text-green-400" />
							</div>
							<div class="min-w-0 flex-1">
								<dt
									class="text-xs font-medium tracking-wide text-green-700 uppercase dark:text-green-300"
								>
									IP Address
								</dt>
								<dd class="truncate text-lg font-semibold text-green-900 dark:text-green-100">
									{endpoint.ip}
								</dd>
							</div>
						</div>
					</div>
				</div>

				<!-- Secondary System Info -->
				<div class="space-y-4">
					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Cpu size={16} class="flex-shrink-0 text-purple-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">
									Operating System
								</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{endpoint.os || 'N/A'}
								</dd>
							</div>
						</div>

						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Settings size={16} class="flex-shrink-0 text-orange-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">Agent Version</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{endpoint.agent_version || 'N/A'}
								</dd>
							</div>
						</div>

						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Clock size={16} class="flex-shrink-0 text-indigo-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">Last Seen</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{endpoint.last_seen ? formatDate(endpoint.last_seen) : 'N/A'}
								</dd>
							</div>
						</div>

						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Activity size={16} class="flex-shrink-0 text-cyan-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">Uptime</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{hostInfo.uptime || (endpoint.uptime ? formatDuration(endpoint.uptime) : 'N/A')}
								</dd>
							</div>
						</div>

						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Layers size={16} class="flex-shrink-0 text-pink-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">
									Running Processes
								</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{hostInfo.procs || 'N/A'}
								</dd>
							</div>
						</div>

						<div class="flex items-center space-x-3 rounded-lg bg-gray-50 p-3 dark:bg-gray-700/50">
							<Users size={16} class="flex-shrink-0 text-emerald-500" />
							<div class="min-w-0 flex-1">
								<dt class="text-sm font-medium text-gray-600 dark:text-gray-300">
									Users Logged In
								</dt>
								<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
									{hostInfo.users_loggedin || 'N/A'}
								</dd>
							</div>
						</div>
					</div>

					<!-- Additional host info from summary -->
					{#if hostInfo.summary && Object.keys(hostInfo.summary).length > 0}
						<div class="mt-6 border-t border-gray-200 pt-4 dark:border-gray-600">
							<h4 class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">
								Additional Information
							</h4>
							<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
								{#each Object.entries(hostInfo.summary) as [key, value]}
									<div
										class="flex items-center space-x-3 rounded bg-gray-50 p-2 dark:bg-gray-700/30"
									>
										<div class="h-2 w-2 flex-shrink-0 rounded-full bg-gray-400"></div>
										<div class="min-w-0 flex-1">
											<dt
												class="text-xs font-medium tracking-wide text-gray-500 uppercase dark:text-gray-400"
											>
												{key.charAt(0).toUpperCase() + key.slice(1).replace(/_/g, ' ')}
											</dt>
											<dd class="truncate text-sm font-medium text-gray-900 dark:text-white">
												{value || 'N/A'}
											</dd>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>

			<!-- Quick Stats Summary -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Current Status</h3>
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
					<div class="text-center">
						<p class="text-2xl font-bold text-blue-600 dark:text-blue-400">
							{cpuPercent.toFixed(1)}%
						</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">CPU Usage</p>
					</div>
					<div class="text-center">
						<p class="text-2xl font-bold text-green-600 dark:text-green-400">
							{memoryPercent.toFixed(1)}%
						</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">Memory Usage</p>
					</div>
					<div class="text-center">
						<p class="text-2xl font-bold text-yellow-500 dark:text-yellow-400">
							{swapPercent.toFixed(1)}%
						</p>
						<p class="text-sm text-gray-500 dark:text-gray-400">Swap Usage</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Sidebar -->
		<div class="space-y-6">
			<!-- Quick Actions -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Quick Actions</h3>
				<div class="space-y-2">
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('restart')}
					>
						<RotateCcw size={16} class="text-blue-500" />
						<span>Restart Service</span>
					</button>
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('status')}
					>
						<Activity size={16} class="text-green-500" />
						<span>Check Status</span>
					</button>
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={() => runCommand('update')}
					>
						<RefreshCw size={16} class="text-purple-500" />
						<span>Update Agent</span>
					</button>
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={openCreateModal}
					>
						<Plus size={16} class="text-orange-500" />
						<span>Create Alert Rule</span>
					</button>
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={runPlaybook}
					>
						<Play size={16} class="text-indigo-500" />
						<span>Run Playbook</span>
					</button>
					<button
						class="flex w-full items-center space-x-3 rounded px-3 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
						on:click={addTag}
					>
						<Tag size={16} class="text-cyan-500" />
						<span>Add Tag</span>
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

			<!-- Recent Events -->
			<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
				<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Recent Events</h3>
				<div class="space-y-3">
					{#each Array.isArray(events) ? events.slice(0, 5) : [] as event}
						<div
							class="flex items-center space-x-3 rounded-lg p-2 {event.type === 'endpoint_down'
								? 'bg-red-50 dark:bg-red-900/20'
								: event.type === 'endpoint_up'
									? 'bg-green-50 dark:bg-green-900/20'
									: event.type === 'alert_triggered'
										? 'bg-orange-50 dark:bg-orange-900/20'
										: 'bg-blue-50 dark:bg-blue-900/20'}"
						>
							<Zap
								size={16}
								class={event.type === 'endpoint_down'
									? 'text-red-500'
									: event.type === 'endpoint_up'
										? 'text-green-500'
										: event.type === 'alert_triggered'
											? 'text-orange-500'
											: 'text-blue-500'}
							/>
							<div class="min-w-0 flex-1">
								<p class="truncate text-xs font-medium text-gray-900 dark:text-white">
									{event.message || event.title || event.type}
								</p>
								<p class="text-xs text-gray-500 dark:text-gray-400">
									{formatDate(event.timestamp || event.created_at || new Date())}
								</p>
							</div>
						</div>
					{:else}
						<p class="text-sm text-gray-500 dark:text-gray-400">No recent events</p>
					{/each}
				</div>
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

<!-- Create Alert Rule Modal -->
{#if showCreateModal}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<button
				class="bg-opacity-50 fixed inset-0 cursor-pointer bg-black transition-opacity"
				on:click={closeCreateModal}
				on:keydown={(e) => e.key === 'Escape' && closeCreateModal()}
				aria-label="Close modal"
			></button>

			<div class="relative w-full max-w-lg rounded-lg bg-white shadow-xl dark:bg-gray-800">
				<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">Create Alert Rule</h3>
				</div>

				<form on:submit|preventDefault={saveRule} class="space-y-4 px-6 py-4">
					<div>
						<label
							for="rule-name"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Name
						</label>
						<input
							id="rule-name"
							type="text"
							bind:value={formData.name}
							required
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
							placeholder="Enter rule name"
						/>
					</div>

					<div>
						<label
							for="rule-description"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Description
						</label>
						<textarea
							id="rule-description"
							bind:value={formData.description}
							rows="2"
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
							placeholder="Enter rule description"
						></textarea>
					</div>

					<div>
						<label
							for="rule-metric"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Metric
						</label>
						<select
							id="rule-metric"
							bind:value={formData.metric_name}
							required
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
						>
							<option value="">Select metric</option>
							<option value="cpu_usage">CPU Usage</option>
							<option value="memory_usage">Memory Usage</option>
							<option value="disk_usage">Disk Usage</option>
							<option value="network_usage">Network Usage</option>
						</select>
					</div>

					<div class="grid grid-cols-2 gap-4">
						<div>
							<label
								for="rule-operator"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								Operator
							</label>
							<select
								id="rule-operator"
								bind:value={formData.operator}
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
							>
								<option value="gt">Greater than</option>
								<option value="gte">Greater than or equal</option>
								<option value="lt">Less than</option>
								<option value="lte">Less than or equal</option>
								<option value="eq">Equal</option>
								<option value="ne">Not equal</option>
							</select>
						</div>

						<div>
							<label
								for="rule-threshold"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								Threshold
							</label>
							<input
								id="rule-threshold"
								type="number"
								bind:value={formData.threshold}
								required
								step="0.1"
								class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
								placeholder="0"
							/>
						</div>
					</div>

					<div>
						<label
							for="rule-severity"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Severity
						</label>
						<select
							id="rule-severity"
							bind:value={formData.severity}
							class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-2 focus:ring-blue-200 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:focus:border-blue-400"
						>
							<option value="info">Info</option>
							<option value="warning">Warning</option>
							<option value="critical">Critical</option>
						</select>
					</div>

					<div class="flex items-center">
						<input
							id="rule-enabled"
							type="checkbox"
							bind:checked={formData.enabled}
							class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-2 focus:ring-blue-500"
						/>
						<label
							for="rule-enabled"
							class="ml-2 text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Enable rule
						</label>
					</div>

					<div class="flex justify-end space-x-3 pt-4">
						<button
							type="button"
							on:click={closeCreateModal}
							class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="rounded-lg border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
						>
							Create Rule
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
