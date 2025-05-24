<!-- Alert Rules Management Page -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { formatDate } from '$lib/utils';
	import {
		Button,
		Badge,
		Spinner,
		Input,
		Select,
		Modal,
		Card,
		Timeline,
		TimelineItem,
		Radio,
		Textarea
	} from 'flowbite-svelte';
	import {
		PlusOutline,
		EditOutline,
		TrashBinOutline,
		CheckOutline,
		ClockOutline
	} from 'flowbite-svelte-icons';
	import type { AlertRule, AlertRuleFormData, Endpoint } from '$lib/types';

	let alertRules: AlertRule[] = [];
	let loading = true;
	let error = '';
	let showCreateModal = false;
	let showEditModal = false;
	let showStepperModal = false;
	let editingRule: AlertRule | null = null;

	// Stepper state
	let currentStep = 0;
	let stepTitles = ['Type', 'Scope', 'Condition', 'Actions', 'Review'];

	// Form data using the proper FormData interface
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

	// Complex alert data structure for stepper
	let alertData: any = {
		type: 'metric',
		name: '',
		description: '',
		match: {
			endpoint_ids: [],
			labels: {}
		},
		scope: {
			namespace: '',
			subnamespace: '',
			metric: ''
		},
		expression: {
			datatype: 'value',
			operator: 'gt',
			value: 0
		},
		level: 'warning',
		actions: [],
		options: {
			cooldown: '30s',
			eval_interval: '10s',
			repeat_interval: '1m',
			notify_on_resolve: true
		},
		enabled: true,
		availableDimensions: []
	};

	let endpoints: Endpoint[] = [];
	let namespaces: string[] = [];
	let subnamespaces: string[] = [];
	let metrics: string[] = [];
	let dimensions: string[] = [];
	let filters: Array<{ key: string; value: string }> = [];

	onMount(async () => {
		await loadAlertRules();
		await loadEndpoints();
		await loadNamespaces();
	});

	async function loadAlertRules() {
		try {
			loading = true;
			error = '';
			alertRules = await api.alerts.getRules();
		} catch (err) {
			error = 'Failed to load alert rules: ' + (err as Error).message;
		} finally {
			loading = false;
		}
	}

	async function loadEndpoints() {
		try {
			endpoints = await api.endpoints.getAll();
		} catch (err) {
			console.error('Failed to load endpoints:', err);
		}
	}

	async function loadNamespaces() {
		try {
			const response = await api.request('/');
			namespaces = Array.isArray(response) ? response : [];
		} catch (err) {
			console.error('Failed to load namespaces:', err);
		}
	}

	async function loadSubnamespaces(namespace: string) {
		try {
			const response = await api.request(`/${namespace}`);
			subnamespaces = Array.isArray(response) ? response : [];
		} catch (err) {
			console.error('Failed to load subnamespaces:', err);
		}
	}

	async function loadMetrics(namespace: string, subnamespace: string) {
		try {
			const response = await api.request(`/${namespace}/${subnamespace}`);
			metrics = Array.isArray(response) ? response : [];
		} catch (err) {
			console.error('Failed to load metrics:', err);
		}
	}

	async function loadDimensions(namespace: string, subnamespace: string, metric: string) {
		try {
			const response = await api.request(`/${namespace}/${subnamespace}/${metric}/dimensions`);
			dimensions = Array.isArray(response) ? response : [];
			alertData.availableDimensions = dimensions;
		} catch (err) {
			console.error('Failed to load dimensions:', err);
		}
	}

	// Convert complex AlertRule to simple form data
	function alertRuleToFormData(rule: AlertRule): AlertRuleFormData {
		return {
			name: rule.name,
			description: rule.description || '',
			severity: rule.level,
			enabled: rule.enabled,
			endpoint_id: rule.match.endpoint_ids?.[0] || '',
			metric_name: rule.scope.metric || '',
			operator: mapOperatorToFormOperator(rule.expression.operator),
			threshold: typeof rule.expression.value === 'number' ? rule.expression.value : 0,
			duration: parseInt(rule.options.eval_interval || '300')
		};
	}

	// Convert simple form data to complex AlertRule structure
	function formDataToAlertRule(formData: AlertRuleFormData): Partial<AlertRule> {
		return {
			name: formData.name,
			description: formData.description,
			message: `Alert for ${formData.metric_name}`,
			level: formData.severity,
			enabled: formData.enabled,
			type: 'metric',
			match: {
				endpoint_ids: formData.endpoint_id ? [formData.endpoint_id] : [],
				category: 'monitoring'
			},
			scope: {
				metric: formData.metric_name
			},
			expression: {
				operator: mapFormOperatorToOperator(formData.operator),
				value: formData.threshold,
				datatype: 'numeric'
			},
			actions: [],
			options: {
				eval_interval: formData.duration.toString(),
				cooldown: '300',
				notify_on_resolve: true
			}
		};
	}

	// Map backend operators to form operators
	function mapOperatorToFormOperator(operator: string): 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte' {
		const mapping: Record<string, 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte'> = {
			'>': 'gt',
			'<': 'lt',
			'=': 'eq',
			'!=': 'ne',
			'>=': 'gte',
			'<=': 'lte'
		};
		return mapping[operator] || 'gt';
	}

	// Map form operators to backend operators
	function mapFormOperatorToOperator(operator: 'gt' | 'lt' | 'eq' | 'ne' | 'gte' | 'lte'): string {
		const mapping: Record<string, string> = {
			gt: '>',
			lt: '<',
			eq: '=',
			ne: '!=',
			gte: '>=',
			lte: '<='
		};
		return mapping[operator] || '>';
	}

	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'critical':
				return 'text-red-600 bg-red-100';
			case 'warning':
				return 'text-orange-600 bg-orange-100';
			case 'info':
				return 'text-blue-600 bg-blue-100';
			default:
				return 'text-gray-600 bg-gray-100';
		}
	}

	function openCreateModal() {
		formData = {
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
		showCreateModal = true;
	}

	function openStepperModal() {
		resetStepper();
		showStepperModal = true;
	}

	function openEditModal(rule: AlertRule) {
		editingRule = rule;
		formData = alertRuleToFormData(rule);
		showEditModal = true;
	}

	function closeModals() {
		showCreateModal = false;
		showEditModal = false;
		showStepperModal = false;
		editingRule = null;
	}

	async function saveRule() {
		try {
			const alertRuleData = formDataToAlertRule(formData);
			if (editingRule) {
				await api.alerts.updateRule(editingRule.id, alertRuleData);
			} else {
				await api.alerts.createRule(alertRuleData);
			}
			closeModals();
			await loadAlertRules();
		} catch (err) {
			console.error('Failed to save alert rule:', err);
		}
	}

	async function deleteRule(ruleId: string) {
		if (!confirm('Are you sure you want to delete this alert rule?')) return;

		try {
			await api.alerts.deleteRule(ruleId);
			await loadAlertRules();
		} catch (err) {
			console.error('Failed to delete alert rule:', err);
		}
	}

	async function toggleRule(rule: AlertRule) {
		try {
			const updatedRule = { ...rule, enabled: !rule.enabled };
			await api.alerts.updateRule(rule.id, updatedRule);
			await loadAlertRules();
		} catch (err) {
			console.error('Failed to toggle alert rule:', err);
		}
	}

	// Stepper functions
	function nextStep() {
		if (currentStep < stepTitles.length - 1) {
			// Validation before moving from step 1 (Scope) to step 2 (Condition)
			if (currentStep === 1 && alertData.type === 'metric') {
				const ns = alertData.scope.namespace;
				const subns = alertData.scope.subnamespace;
				const metric = alertData.scope.metric;

				if (!ns || !subns || !metric) {
					alert('Please complete Namespace, SubNamespace, and Metric before proceeding.');
					return;
				}
			}
			currentStep++;
		}
	}

	function prevStep() {
		if (currentStep > 0) {
			currentStep--;
		}
	}

	function resetStepper() {
		currentStep = 0;
		alertData = {
			type: 'metric',
			name: '',
			description: '',
			match: {
				endpoint_ids: [],
				labels: {}
			},
			scope: {
				namespace: '',
				subnamespace: '',
				metric: ''
			},
			expression: {
				datatype: 'value',
				operator: 'gt',
				value: 0
			},
			level: 'warning',
			actions: [],
			options: {
				cooldown: '30s',
				eval_interval: '10s',
				repeat_interval: '1m',
				notify_on_resolve: true
			},
			availableDimensions: []
		};
		filters = [];
	}

	async function submitAlert() {
		try {
			// Convert stepper data to API format
			const alertRuleData = {
				name: alertData.name,
				description: alertData.description,
				message: `Alert for ${alertData.scope.metric}`,
				level: alertData.level,
				enabled: true,
				type: alertData.type,
				match: {
					endpoint_ids: alertData.match.endpoint_ids,
					labels: alertData.match.labels
				},
				scope: alertData.scope,
				expression: alertData.expression,
				actions: alertData.actions,
				options: alertData.options
			};

			await api.alerts.createRule(alertRuleData);
			closeModals();
			await loadAlertRules();
		} catch (err) {
			console.error('Failed to create alert rule:', err);
			alert('Failed to create alert rule: ' + (err as Error).message);
		}
	}

	function addFilter() {
		filters = [...filters, { key: '', value: '' }];
	}

	function removeFilter(index: number) {
		filters = filters.filter((_, i) => i !== index);
	}

	// Handle namespace changes
	async function handleNamespaceChange(namespace: string) {
		alertData.scope.namespace = namespace;
		alertData.scope.subnamespace = '';
		alertData.scope.metric = '';
		subnamespaces = [];
		metrics = [];
		if (namespace) {
			await loadSubnamespaces(namespace);
		}
	}

	// Handle subnamespace changes
	async function handleSubnamespaceChange(subnamespace: string) {
		alertData.scope.subnamespace = subnamespace;
		alertData.scope.metric = '';
		metrics = [];
		if (alertData.scope.namespace && subnamespace) {
			await loadMetrics(alertData.scope.namespace, subnamespace);
		}
	}

	// Handle metric changes
	async function handleMetricChange(metric: string) {
		alertData.scope.metric = metric;
		if (alertData.scope.namespace && alertData.scope.subnamespace && metric) {
			await loadDimensions(alertData.scope.namespace, alertData.scope.subnamespace, metric);
		}
	}
</script>

<svelte:head>
	<title>Alert Rules - GoSight</title>
</svelte:head>

<div class="p-6">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Alert Rules</h1>
			<p class="text-gray-600 dark:text-gray-400">Configure and manage alert rules</p>
		</div>
		<div class="flex gap-2">
			<button
				on:click={openCreateModal}
				class="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
			>
				<i class="fas fa-plus mr-2"></i>
				Quick Rule
			</button>
			<button
				on:click={openStepperModal}
				class="rounded-lg bg-green-600 px-4 py-2 text-white transition-colors hover:bg-green-700"
			>
				<i class="fas fa-cog mr-2"></i>
				Advanced Rule
			</button>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
		</div>
	{:else if error}
		<div
			class="rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20"
		>
			<div class="flex">
				<i class="fas fa-exclamation-triangle mt-0.5 mr-3 text-red-500"></i>
				<div>
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
					<p class="mt-1 text-sm text-red-600 dark:text-red-300">{error}</p>
				</div>
			</div>
		</div>
	{:else}
		<!-- Rules Grid -->
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each alertRules as rule (rule.id)}
				<div class="rounded-lg bg-white p-6 shadow dark:bg-gray-800">
					<div class="mb-4 flex items-start justify-between">
						<div class="flex-1">
							<h3 class="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
								{rule.name}
							</h3>
							<p class="text-sm text-gray-600 dark:text-gray-400">
								{rule.description || 'No description'}
							</p>
						</div>
						<div class="flex items-center gap-2">
							<span
								class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getSeverityColor(
									rule.level
								)}"
							>
								{rule.level}
							</span>
							<button
								on:click={() => toggleRule(rule)}
								class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:ring-2 focus:ring-blue-600 focus:ring-offset-2 focus:outline-none {rule.enabled
									? 'bg-blue-600'
									: 'bg-gray-200'}"
							>
								<span class="sr-only">Enable rule</span>
								<span
									class="pointer-events-none relative inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out {rule.enabled
										? 'translate-x-5'
										: 'translate-x-0'}"
								>
									<span
										class="absolute inset-0 flex h-full w-full items-center justify-center transition-opacity duration-200 ease-in-out {rule.enabled
											? 'opacity-0'
											: 'opacity-100'}"
									>
										<svg class="h-3 w-3 text-gray-400" fill="none" viewBox="0 0 12 12">
											<path
												d="M4 8l2-2m0 0l2-2M6 6L4 4m2 2l2 2"
												stroke="currentColor"
												stroke-width="2"
												stroke-linecap="round"
												stroke-linejoin="round"
											/>
										</svg>
									</span>
									<span
										class="absolute inset-0 flex h-full w-full items-center justify-center transition-opacity duration-200 ease-in-out {rule.enabled
											? 'opacity-100'
											: 'opacity-0'}"
									>
										<svg class="h-3 w-3 text-blue-600" fill="currentColor" viewBox="0 0 12 12">
											<path
												d="M3.707 5.293a1 1 0 00-1.414 1.414l1.414-1.414zM5 8l-.707.707a1 1 0 001.414 0L5 8zm4.707-3.293a1 1 0 00-1.414-1.414l1.414 1.414zm-7.414 2l2 2 1.414-1.414-2-2-1.414 1.414zm3.414 2l4-4-1.414-1.414-4 4 1.414 1.414z"
											/>
										</svg>
									</span>
								</span>
							</button>
						</div>
					</div>

					<div class="space-y-2 text-sm">
						{#if rule.scope.metric}
							<div class="flex justify-between">
								<span class="text-gray-600 dark:text-gray-400">Metric:</span>
								<span class="text-gray-900 dark:text-white">{rule.scope.metric}</span>
							</div>
						{/if}
						{#if rule.expression.value !== undefined}
							<div class="flex justify-between">
								<span class="text-gray-600 dark:text-gray-400">Threshold:</span>
								<span class="text-gray-900 dark:text-white"
									>{rule.expression.operator} {rule.expression.value}</span
								>
							</div>
						{/if}
						{#if rule.options.eval_interval}
							<div class="flex justify-between">
								<span class="text-gray-600 dark:text-gray-400">Duration:</span>
								<span class="text-gray-900 dark:text-white">{rule.options.eval_interval}s</span>
							</div>
						{/if}
						{#if rule.match.endpoint_ids && rule.match.endpoint_ids.length > 0}
							<div class="flex justify-between">
								<span class="text-gray-600 dark:text-gray-400">Endpoints:</span>
								<span class="text-gray-900 dark:text-white">{rule.match.endpoint_ids.length}</span>
							</div>
						{/if}
					</div>

					<div class="mt-4 flex justify-end gap-2">
						<button
							on:click={() => openEditModal(rule)}
							class="px-3 py-1 text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
						>
							<i class="fas fa-edit mr-1"></i>
							Edit
						</button>
						<button
							on:click={() => deleteRule(rule.id)}
							class="px-3 py-1 text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300"
						>
							<i class="fas fa-trash mr-1"></i>
							Delete
						</button>
					</div>
				</div>
			{/each}
		</div>

		{#if alertRules.length === 0}
			<div class="py-12 text-center">
				<i class="fas fa-bell-slash mb-4 text-4xl text-gray-400"></i>
				<h3 class="mb-2 text-lg font-medium text-gray-900 dark:text-white">No Alert Rules</h3>
				<p class="mb-4 text-gray-600 dark:text-gray-400">
					Get started by creating your first alert rule.
				</p>
				<button
					on:click={openCreateModal}
					class="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
				>
					<i class="fas fa-plus mr-2"></i>
					Quick Rule
				</button>
			</div>
		{/if}
	{/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal || showEditModal}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<button
				class="bg-opacity-50 fixed inset-0 cursor-pointer bg-black transition-opacity"
				on:click={closeModals}
				on:keydown={(e) => e.key === 'Escape' && closeModals()}
				aria-label="Close modal"
			></button>

			<div class="relative w-full max-w-lg rounded-lg bg-white shadow-xl dark:bg-gray-800">
				<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">
						{editingRule ? 'Edit Alert Rule' : 'Create Alert Rule'}
					</h3>
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
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
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
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						></textarea>
					</div>

					<div class="grid grid-cols-2 gap-4">
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
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							>
								<option value="info">Info</option>
								<option value="warning">Warning</option>
								<option value="critical">Critical</option>
							</select>
						</div>

						<div>
							<label
								for="rule-endpoint"
								class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
							>
								Endpoint
							</label>
							<select
								id="rule-endpoint"
								bind:value={formData.endpoint_id}
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							>
								<option value="">All Endpoints</option>
								{#each endpoints as endpoint}
									<option value={endpoint.id}>{endpoint.name}</option>
								{/each}
							</select>
						</div>
					</div>

					<div>
						<label
							for="rule-metric"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Metric Name
						</label>
						<input
							id="rule-metric"
							type="text"
							bind:value={formData.metric_name}
							placeholder="e.g., cpu_usage, memory_usage, response_time"
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
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
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							>
								<option value="gt">Greater than</option>
								<option value="lt">Less than</option>
								<option value="eq">Equal to</option>
								<option value="gte">Greater than or equal</option>
								<option value="lte">Less than or equal</option>
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
								step="0.01"
								class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
							/>
						</div>
					</div>

					<div>
						<label
							for="rule-duration"
							class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300"
						>
							Duration (seconds)
						</label>
						<input
							id="rule-duration"
							type="number"
							bind:value={formData.duration}
							min="0"
							class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
						/>
						<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
							How long the condition must persist before triggering an alert
						</p>
					</div>

					<div class="flex items-center">
						<input
							type="checkbox"
							id="enabled"
							bind:checked={formData.enabled}
							class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
						/>
						<label for="enabled" class="ml-2 block text-sm text-gray-900 dark:text-white">
							Enable this rule
						</label>
					</div>

					<div class="flex justify-end gap-3 pt-4">
						<button
							type="button"
							on:click={closeModals}
							class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600"
						>
							Cancel
						</button>
						<button
							type="submit"
							class="rounded-lg border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
						>
							{editingRule ? 'Update' : 'Create'} Rule
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<!-- Stepper Modal -->
{#if showStepperModal}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<button
				class="bg-opacity-50 fixed inset-0 cursor-pointer bg-black transition-opacity"
				on:click={closeModals}
				on:keydown={(e) => e.key === 'Escape' && closeModals()}
				aria-label="Close modal"
			></button>

			<div class="relative w-full max-w-3xl rounded-lg bg-white shadow-xl dark:bg-gray-800">
				<div class="border-b border-gray-200 px-6 py-4 dark:border-gray-700">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">
						Create Alert Rule - Step {currentStep + 1}: {stepTitles[currentStep]}
					</h3>
				</div>

				<div class="px-6 py-4">
					{#if currentStep === 0}
						<!-- Step 1: Type -->
						<div>
							<div class="mb-4">
								<label
									for="alert-type"
									class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
								>
									Alert Type
								</label>
								<select
									id="alert-type"
									bind:value={alertData.type}
									class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								>
									<option value="metric">Metric Alert</option>
									<option value="log">Log Alert</option>
								</select>
							</div>

							<div class="mb-4">
								<label
									for="alert-name"
									class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
								>
									Alert Name
								</label>
								<input
									id="alert-name"
									type="text"
									bind:value={alertData.name}
									placeholder="Enter alert name"
									class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								/>
							</div>

							<div>
								<label
									for="alert-description"
									class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
								>
									Description
								</label>
								<textarea
									id="alert-description"
									bind:value={alertData.description}
									placeholder="Enter alert description"
									rows={2}
									class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								></textarea>
							</div>
						</div>
					{:else if currentStep === 1}
						<!-- Step 2: Scope -->
						<div>
							<h4 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Metric Scope</h4>
							<div class="grid grid-cols-2 gap-4">
								<div>
									<label
										for="namespace-select"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Namespace
									</label>
									<select
										id="namespace-select"
										bind:value={alertData.scope.namespace}
										on:change={() => handleNamespaceChange(alertData.scope.namespace)}
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									>
										<option value="">Select namespace</option>
										{#each namespaces as ns}
											<option value={ns}>{ns}</option>
										{/each}
									</select>
								</div>

								<div>
									<label
										for="subnamespace-select"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Subnamespace
									</label>
									<select
										id="subnamespace-select"
										bind:value={alertData.scope.subnamespace}
										on:change={() => handleSubnamespaceChange(alertData.scope.subnamespace)}
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									>
										<option value="">Select subnamespace</option>
										{#each subnamespaces as subns}
											<option value={subns}>{subns}</option>
										{/each}
									</select>
								</div>
							</div>

							<div class="mt-4">
								<label
									for="metric-select"
									class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
								>
									Metric
								</label>
								<select
									id="metric-select"
									bind:value={alertData.scope.metric}
									on:change={() => handleMetricChange(alertData.scope.metric)}
									class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								>
									<option value="">Select metric</option>
									{#each metrics as metric}
										<option value={metric}>{metric}</option>
									{/each}
								</select>
							</div>

							{#if dimensions.length > 0}
								<div class="mt-4">
									<h5 class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
										Filters (Optional)
									</h5>
									{#each filters as filter, index}
										<div class="mb-2 flex gap-2">
											<select
												bind:value={filter.key}
												class="flex-1 rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											>
												<option value="">Select dimension</option>
												{#each dimensions as dim}
													<option value={dim}>{dim}</option>
												{/each}
											</select>
											<input
												type="text"
												bind:value={filter.value}
												placeholder="Value"
												class="flex-1 rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
											/>
											<button
												type="button"
												on:click={() => removeFilter(index)}
												aria-label="Remove filter"
												class="rounded-lg bg-red-600 px-3 py-2 text-white transition-colors hover:bg-red-700"
											>
												<i class="fas fa-trash"></i>
											</button>
										</div>
									{/each}
									<button
										type="button"
										on:click={addFilter}
										class="rounded-lg bg-blue-600 px-3 py-2 text-white transition-colors hover:bg-blue-700"
									>
										<i class="fas fa-plus mr-2"></i>
										Add Filter
									</button>
								</div>
							{/if}
						</div>
					{:else if currentStep === 2}
						<!-- Step 3: Condition -->
						<div>
							<h4 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">
								Alert Condition
							</h4>
							<div class="grid grid-cols-2 gap-4">
								<div>
									<label
										for="operator-select"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Operator
									</label>
									<select
										id="operator-select"
										bind:value={alertData.expression.operator}
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									>
										<option value="gt">Greater than (&gt;)</option>
										<option value="lt">Less than (&lt;)</option>
										<option value="eq">Equal to (=)</option>
										<option value="ne">Not equal to (≠)</option>
										<option value="gte">Greater than or equal (≥)</option>
										<option value="lte">Less than or equal (≤)</option>
									</select>
								</div>

								<div>
									<label
										for="threshold-value"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Threshold Value
									</label>
									<input
										id="threshold-value"
										type="number"
										bind:value={alertData.expression.value}
										placeholder="Enter threshold value"
										step="0.01"
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									/>
								</div>
							</div>

							<div class="mt-4 grid grid-cols-2 gap-4">
								<div>
									<label
										for="eval-interval"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Evaluation Interval
									</label>
									<input
										id="eval-interval"
										type="text"
										bind:value={alertData.options.eval_interval}
										placeholder="e.g., 10s, 1m"
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									/>
									<p class="mt-1 text-xs text-gray-500">How often to check the condition</p>
								</div>

								<div>
									<label
										for="cooldown"
										class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
									>
										Cooldown Period
									</label>
									<input
										id="cooldown"
										type="text"
										bind:value={alertData.options.cooldown}
										placeholder="e.g., 30s, 5m"
										class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
									/>
									<p class="mt-1 text-xs text-gray-500">Time to wait before re-triggering</p>
								</div>
							</div>

							<div class="mt-4">
								<label
									for="severity-select"
									class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300"
								>
									Severity Level
								</label>
								<select
									id="severity-select"
									bind:value={alertData.level}
									class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-gray-900 focus:border-transparent focus:ring-2 focus:ring-blue-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white"
								>
									<option value="info">Info</option>
									<option value="warning">Warning</option>
									<option value="critical">Critical</option>
								</select>
							</div>
						</div>
					{:else if currentStep === 3}
						<!-- Step 4: Actions -->
						<div>
							<h4 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Alert Actions</h4>
							<div class="mb-4">
								<label class="flex items-center">
									<input
										type="checkbox"
										bind:checked={alertData.options.notify_on_resolve}
										class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
									<span class="ml-2 text-sm text-gray-900 dark:text-white">
										Notify when alert resolves
									</span>
								</label>
							</div>

							<div class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
								<p class="text-sm text-gray-600 dark:text-gray-400">
									Alert actions such as email notifications, webhooks, and integrations will be
									available in a future update.
								</p>
								<button
									type="button"
									disabled
									class="mt-2 cursor-not-allowed rounded-lg bg-gray-300 px-4 py-2 text-gray-500"
								>
									<i class="fas fa-plus mr-2"></i>
									Add Action (Coming Soon)
								</button>
							</div>
						</div>
					{:else if currentStep === 4}
						<!-- Step 5: Review -->
						<div>
							<h4 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
								Review Your Alert Rule
							</h4>
							<div class="space-y-3 rounded-lg border border-gray-200 p-4 dark:border-gray-700">
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Name:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.name || 'Unnamed Alert'}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Description:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.description || 'No description'}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Type:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.type}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Metric:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.scope.namespace}/{alertData.scope.subnamespace}/{alertData.scope
											.metric}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Condition:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.expression.operator}
										{alertData.expression.value}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Severity:</span>
									<span class="font-medium text-gray-900 capitalize dark:text-white">
										{alertData.level}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Evaluation Interval:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.options.eval_interval}
									</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600 dark:text-gray-400">Cooldown:</span>
									<span class="font-medium text-gray-900 dark:text-white">
										{alertData.options.cooldown}
									</span>
								</div>
							</div>

							<div class="mt-6">
								<button
									type="button"
									on:click={submitAlert}
									class="w-full rounded-lg bg-green-600 px-4 py-2 text-white transition-colors hover:bg-green-700 focus:ring-2 focus:ring-green-500 focus:ring-offset-2 focus:outline-none"
								>
									<i class="fas fa-check mr-2"></i>
									Create Alert Rule
								</button>
							</div>
						</div>
					{/if}
				</div>

				<div class="flex justify-between border-t border-gray-200 px-6 py-4 dark:border-gray-700">
					<button
						on:click={prevStep}
						class="rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:bg-gray-700 dark:text-gray-300"
						disabled={currentStep === 0}
					>
						<i class="fas fa-arrow-left mr-2"></i>
						Back
					</button>
					<button
						on:click={nextStep}
						class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
						disabled={currentStep === stepTitles.length - 1}
					>
						<i class="fas fa-arrow-right mr-2"></i>
						Next
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
