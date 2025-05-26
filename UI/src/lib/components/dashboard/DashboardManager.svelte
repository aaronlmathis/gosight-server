<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Modal, Card, Input, Label, Select, Textarea, Badge, Spinner } from 'flowbite-svelte';
	import CompatButton from '$lib/components/CompatButton.svelte';
	import {
		Plus,
		Download,
		Upload,
		Copy,
		Trash2,
		Edit3,
		Save,
		X,
		FileText,
		Folder
	} from 'lucide-svelte';
	import { dashboardStore } from '$lib/stores/dashboard';
	import type { Dashboard, DashboardPreferences, Widget } from '$lib/types/dashboard';
	import { dataService } from '$lib/services/dataService';

	const dispatch = createEventDispatcher<{
		selectDashboard: { dashboardId: string };
		close: void;
	}>();

	export let isOpen = false;
	export let currentDashboardId = '';

	let activeTab: 'manage' | 'templates' | 'import-export' = 'manage';
	let showCreateModal = false;
	let showDeleteModal = false;
	let showRenameModal = false;
	let showImportModal = false;
	let showExportModal = false;

	// Form states
	let newDashboardName = '';
	let newDashboardDescription = '';
	let selectedTemplate = '';
	let dashboardToDelete: Dashboard | null = null;
	let dashboardToRename: Dashboard | null = null;
	let renameValue = '';
	let importData = '';
	let exportData = '';
	let loading = false;
	let error = '';

	// Dashboard templates
	const dashboardTemplates = [
		{
			id: 'system-monitoring',
			name: 'System Monitoring',
			description: 'Comprehensive system metrics and health monitoring',
			category: 'Infrastructure',
			widgets: [
				{
					type: 'system_overview',
					title: 'System Overview',
					position: { x: 0, y: 0, width: 6, height: 3 }
				},
				{ type: 'cpu-usage', title: 'CPU Usage', position: { x: 6, y: 0, width: 6, height: 3 } },
				{
					type: 'memory-usage',
					title: 'Memory Usage',
					position: { x: 0, y: 3, width: 6, height: 3 }
				},
				{ type: 'disk-usage', title: 'Disk Usage', position: { x: 6, y: 3, width: 6, height: 3 } },
				{
					type: 'alert_count',
					title: 'Active Alerts',
					position: { x: 0, y: 6, width: 4, height: 2 }
				},
				{
					type: 'endpoint_count',
					title: 'Endpoints',
					position: { x: 4, y: 6, width: 4, height: 2 }
				},
				{
					type: 'recent-events',
					title: 'Recent Events',
					position: { x: 8, y: 6, width: 4, height: 2 }
				}
			]
		},
		{
			id: 'security-overview',
			name: 'Security Overview',
			description: 'Security alerts, events, and threat monitoring',
			category: 'Security',
			widgets: [
				{
					type: 'alert_count',
					title: 'Security Alerts',
					position: { x: 0, y: 0, width: 6, height: 3 }
				},
				{
					type: 'alerts_list',
					title: 'Critical Alerts',
					position: { x: 6, y: 0, width: 6, height: 4 }
				},
				{
					type: 'events_list',
					title: 'Security Events',
					position: { x: 0, y: 3, width: 6, height: 4 }
				},
				{
					type: 'endpoint_count',
					title: 'Protected Endpoints',
					position: { x: 0, y: 7, width: 4, height: 1 }
				},
				{
					type: 'quick_links',
					title: 'Security Tools',
					position: { x: 4, y: 7, width: 8, height: 1 }
				}
			]
		},
		{
			id: 'application-performance',
			name: 'Application Performance',
			description: 'Application metrics, response times, and performance monitoring',
			category: 'Performance',
			widgets: [
				{
					type: 'response-time',
					title: 'Response Time',
					position: { x: 0, y: 0, width: 6, height: 3 }
				},
				{ type: 'throughput', title: 'Throughput', position: { x: 6, y: 0, width: 6, height: 3 } },
				{ type: 'error-rate', title: 'Error Rate', position: { x: 0, y: 3, width: 6, height: 3 } },
				{
					type: 'network-traffic',
					title: 'Network Traffic',
					position: { x: 6, y: 3, width: 6, height: 3 }
				},
				{
					type: 'alerts_list',
					title: 'Performance Alerts',
					position: { x: 0, y: 6, width: 12, height: 2 }
				}
			]
		},
		{
			id: 'executive-summary',
			name: 'Executive Summary',
			description: 'High-level overview for management and executives',
			category: 'Business',
			widgets: [
				{
					type: 'system_overview',
					title: 'System Health',
					position: { x: 0, y: 0, width: 4, height: 3 }
				},
				{
					type: 'alert_count',
					title: 'Alert Summary',
					position: { x: 4, y: 0, width: 4, height: 3 }
				},
				{
					type: 'endpoint_count',
					title: 'Infrastructure',
					position: { x: 8, y: 0, width: 4, height: 3 }
				},
				{
					type: 'uptime-monitor',
					title: 'Service Uptime',
					position: { x: 0, y: 3, width: 6, height: 2 }
				},
				{
					type: 'service-health',
					title: 'Service Health',
					position: { x: 6, y: 3, width: 6, height: 2 }
				},
				{
					type: 'quick_links',
					title: 'Key Reports',
					position: { x: 0, y: 5, width: 12, height: 3 }
				}
			]
		}
	];

	$: dashboards = $dashboardStore.dashboards;
	$: categories = [...new Set(dashboardTemplates.map((t) => t.category))];

	// Create new dashboard
	async function createDashboard() {
		if (!newDashboardName.trim()) {
			error = 'Dashboard name is required';
			return;
		}

		try {
			loading = true;
			error = '';

			const dashboardId = `dashboard-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
			let widgets: Partial<Widget>[] = [];

			// Apply template if selected
			if (selectedTemplate) {
				const template = dashboardTemplates.find((t) => t.id === selectedTemplate);
				if (template) {
					widgets = template.widgets.map((w) => ({
						type: w.type as any,
						title: w.title,
						position: w.position,
						config: {}
					}));
				}
			}

			// Create dashboard
			dashboardStore.addDashboard({
				name: newDashboardName,
				description: newDashboardDescription,
				isDefault: dashboards.length === 0,
				layout: {
					columns: 12,
					rowHeight: 60,
					margin: [16, 16],
					padding: [20, 20]
				},
				widgets: []
			});

			// Add widgets if from template
			for (const widget of widgets) {
				if (widget.type && widget.title && widget.position) {
					dashboardStore.addWidget(dashboardId, {
						type: widget.type,
						title: widget.title,
						position: widget.position,
						config: widget.config || {}
					});
				}
			}

			// Save configuration
			await dataService.saveDashboardConfig(dashboardId, {
				name: newDashboardName,
				description: newDashboardDescription,
				template: selectedTemplate || undefined
			});

			// Reset form
			newDashboardName = '';
			newDashboardDescription = '';
			selectedTemplate = '';
			showCreateModal = false;

			// Select new dashboard
			dispatch('selectDashboard', { dashboardId });
		} catch (err) {
			console.error('Failed to create dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to create dashboard';
		} finally {
			loading = false;
		}
	}

	// Delete dashboard
	async function deleteDashboard() {
		if (!dashboardToDelete) return;

		try {
			loading = true;
			error = '';

			// Remove from store
			dashboardStore.deleteDashboard(dashboardToDelete.id);

			// Clear backend config
			await dataService.clearDashboardConfig(dashboardToDelete.id);

			// Switch to first available dashboard if deleted was current
			if (dashboardToDelete?.id === currentDashboardId && dashboards.length > 1) {
				const remaining = dashboards.filter((d) => d.id !== dashboardToDelete!.id);
				if (remaining.length > 0) {
					dispatch('selectDashboard', { dashboardId: remaining[0].id });
				}
			}

			showDeleteModal = false;
			dashboardToDelete = null;
		} catch (err) {
			console.error('Failed to delete dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to delete dashboard';
		} finally {
			loading = false;
		}
	}

	// Rename dashboard
	async function renameDashboard() {
		if (!dashboardToRename || !renameValue.trim()) return;

		try {
			loading = true;
			error = '';

			// Update store
			dashboardStore.updateDashboard(dashboardToRename.id, {
				name: renameValue,
				updatedAt: new Date().toISOString()
			});

			// Save to backend
			await dataService.saveDashboardConfig(dashboardToRename.id, {
				name: renameValue
			});

			showRenameModal = false;
			dashboardToRename = null;
			renameValue = '';
		} catch (err) {
			console.error('Failed to rename dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to rename dashboard';
		} finally {
			loading = false;
		}
	}

	// Duplicate dashboard
	async function duplicateDashboard(dashboard: Dashboard) {
		try {
			loading = true;
			error = '';

			const newId = `dashboard-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
			const newName = `${dashboard.name} (Copy)`;

			// Create duplicate
			dashboardStore.addDashboard({
				name: newName,
				description: dashboard.description,
				isDefault: false,
				layout: dashboard.layout,
				widgets: dashboard.widgets.map((w) => ({
					...w,
					id: `widget-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
					createdAt: new Date().toISOString(),
					updatedAt: new Date().toISOString()
				}))
			});

			// Save to backend
			await dataService.saveDashboardConfig(newId, {
				name: newName,
				duplicatedFrom: dashboard.id
			});

			dispatch('selectDashboard', { dashboardId: newId });
		} catch (err) {
			console.error('Failed to duplicate dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to duplicate dashboard';
		} finally {
			loading = false;
		}
	}

	// Export dashboard
	function exportDashboard(dashboard: Dashboard) {
		const exportConfig = {
			version: '1.0',
			dashboard: {
				name: dashboard.name,
				layout: dashboard.layout,
				widgets: dashboard.widgets.map((w) => ({
					type: w.type,
					title: w.title,
					position: w.position,
					config: w.config
				}))
			},
			exportedAt: new Date().toISOString(),
			exportedBy: 'GoSight Dashboard Manager'
		};

		exportData = JSON.stringify(exportConfig, null, 2);
		showExportModal = true;
	}

	// Import dashboard
	async function importDashboard() {
		try {
			loading = true;
			error = '';

			const config = JSON.parse(importData);
			if (!config.dashboard || !config.dashboard.name) {
				throw new Error('Invalid dashboard configuration');
			}

			const dashboardId = `dashboard-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

			// Create dashboard
			dashboardStore.addDashboard({
				name: config.dashboard.name,
				description: config.dashboard.description,
				isDefault: false,
				layout: config.dashboard.layout || {
					columns: 12,
					rowHeight: 100,
					margin: [10, 10],
					padding: [10, 10]
				},
				widgets: []
			});

			// Add widgets
			for (const widgetConfig of config.dashboard.widgets || []) {
				dashboardStore.addWidget(dashboardId, {
					type: widgetConfig.type,
					title: widgetConfig.title,
					position: widgetConfig.position,
					config: widgetConfig.config || {}
				});
			}

			// Save to backend
			await dataService.saveDashboardConfig(dashboardId, {
				name: config.dashboard.name,
				imported: true,
				importedAt: new Date().toISOString()
			});

			importData = '';
			showImportModal = false;
			dispatch('selectDashboard', { dashboardId });
		} catch (err) {
			console.error('Failed to import dashboard:', err);
			error = err instanceof Error ? err.message : 'Failed to import dashboard configuration';
		} finally {
			loading = false;
		}
	}

	// Copy export data to clipboard
	async function copyExportData() {
		try {
			await navigator.clipboard.writeText(exportData);
		} catch (err) {
			console.error('Failed to copy to clipboard:', err);
		}
	}

	// Download export data as file
	function downloadExportData(dashboard: Dashboard) {
		const blob = new Blob([exportData], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${dashboard.name.replace(/[^a-z0-9]/gi, '_').toLowerCase()}_dashboard.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	// Open modals
	function openCreateModal() {
		newDashboardName = '';
		newDashboardDescription = '';
		selectedTemplate = '';
		error = '';
		showCreateModal = true;
	}

	function openDeleteModal(dashboard: Dashboard) {
		dashboardToDelete = dashboard;
		error = '';
		showDeleteModal = true;
	}

	function openRenameModal(dashboard: Dashboard) {
		dashboardToRename = dashboard;
		renameValue = dashboard.name;
		error = '';
		showRenameModal = true;
	}

	function openImportModal() {
		importData = '';
		error = '';
		showImportModal = true;
	}
</script>

<Modal bind:open={isOpen} size="xl" class="w-full max-w-6xl">
	<div class="p-6">
		<!-- Header -->
		<div class="mb-6 flex items-center justify-between">
			<h2 class="text-2xl font-bold text-gray-900 dark:text-white">Dashboard Manager</h2>
			<CompatButton color="light" size="sm" on:click={() => dispatch('close')}>
				<X class="h-4 w-4" />
			</CompatButton>
		</div>

		<!-- Error Display -->
		{#if error}
			<div class="mb-4 rounded-lg border border-red-200 bg-red-50 p-3">
				<p class="text-sm text-red-600">{error}</p>
			</div>
		{/if}

		<!-- Tabs -->
		<div class="mb-6 border-b border-gray-200 dark:border-gray-700">
			<nav class="-mb-px flex space-x-8">
				<button
					class="border-b-2 px-1 py-2 text-sm font-medium {activeTab === 'manage'
						? 'border-blue-500 text-blue-600'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
					on:click={() => (activeTab = 'manage')}
				>
					<Folder class="mr-2 inline h-4 w-4" />
					Manage Dashboards
				</button>
				<button
					class="border-b-2 px-1 py-2 text-sm font-medium {activeTab === 'templates'
						? 'border-blue-500 text-blue-600'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
					on:click={() => (activeTab = 'templates')}
				>
					<FileText class="mr-2 inline h-4 w-4" />
					Templates
				</button>
				<button
					class="border-b-2 px-1 py-2 text-sm font-medium {activeTab === 'import-export'
						? 'border-blue-500 text-blue-600'
						: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}"
					on:click={() => (activeTab = 'import-export')}
				>
					<Upload class="mr-2 inline h-4 w-4" />
					Import/Export
				</button>
			</nav>
		</div>

		<!-- Tab Content -->
		{#if activeTab === 'manage'}
			<!-- Manage Dashboards Tab -->
			<div class="space-y-4">
				<!-- Actions Bar -->
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-medium text-gray-900 dark:text-white">
						Your Dashboards ({dashboards.length})
					</h3>
					<CompatButton color="blue" size="sm" on:click={openCreateModal}>
						<Plus class="mr-2 h-4 w-4" />
						Create Dashboard
					</CompatButton>
				</div>

				<!-- Dashboards Grid -->
				<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each dashboards as dashboard (dashboard.id)}
						<Card
							class="relative {dashboard.id === currentDashboardId ? 'ring-2 ring-blue-500' : ''}"
						>
							<!-- Default Badge -->
							{#if dashboard.isDefault}
								<Badge color="blue" class="absolute top-2 right-2 text-xs">Default</Badge>
							{/if}

							<div class="p-4">
								<h4 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
									{dashboard.name}
								</h4>

								<div class="mb-4 text-sm text-gray-500 dark:text-gray-400">
									<p>{dashboard.widgets.length} widgets</p>
									<p>Updated {new Date(dashboard.updatedAt).toLocaleDateString()}</p>
								</div>

								<!-- Actions -->
								<div class="flex items-center justify-between">
									<div class="flex space-x-2">
										<CompatButton
											size="xs"
											color="blue"
											on:click={() => dispatch('selectDashboard', { dashboardId: dashboard.id })}
										>
											{dashboard.id === currentDashboardId ? 'Current' : 'Select'}
										</CompatButton>
										<CompatButton
											size="xs"
											color="light"
											on:click={() => openRenameModal(dashboard)}
										>
											<Edit3 class="h-3 w-3" />
										</CompatButton>
										<CompatButton
											size="xs"
											color="light"
											on:click={() => duplicateDashboard(dashboard)}
										>
											<Copy class="h-3 w-3" />
										</CompatButton>
									</div>

									<div class="flex space-x-2">
										<CompatButton
											size="xs"
											color="light"
											on:click={() => exportDashboard(dashboard)}
										>
											<Download class="h-3 w-3" />
										</CompatButton>
										{#if !dashboard.isDefault}
											<CompatButton
												size="xs"
												color="red"
												on:click={() => openDeleteModal(dashboard)}
											>
												<Trash2 class="h-3 w-3" />
											</CompatButton>
										{/if}
									</div>
								</div>
							</div>
						</Card>
					{/each}
				</div>
			</div>
		{:else if activeTab === 'templates'}
			<!-- Templates Tab -->
			<div class="space-y-4">
				<h3 class="text-lg font-medium text-gray-900 dark:text-white">Dashboard Templates</h3>

				<!-- Templates Grid -->
				<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
					{#each dashboardTemplates as template (template.id)}
						<Card class="p-4">
							<div class="mb-3 flex items-start justify-between">
								<div>
									<h4 class="text-lg font-semibold text-gray-900 dark:text-white">
										{template.name}
									</h4>
									<Badge color="gray" class="mt-1 text-xs">{template.category}</Badge>
								</div>
							</div>

							<p class="mb-4 text-sm text-gray-600 dark:text-gray-400">
								{template.description}
							</p>

							<div class="mb-4 text-xs text-gray-500 dark:text-gray-400">
								{template.widgets.length} widgets included:
								<span class="ml-1">
									{template.widgets
										.slice(0, 3)
										.map((w) => w.title)
										.join(', ')}
									{#if template.widgets.length > 3}...{/if}
								</span>
							</div>

							<CompatButton
								size="sm"
								color="blue"
								class="w-full"
								on:click={() => {
									selectedTemplate = template.id;
									newDashboardName = template.name;
									newDashboardDescription = template.description;
									showCreateModal = true;
								}}
							>
								Use Template
							</CompatButton>
						</Card>
					{/each}
				</div>
			</div>
		{:else if activeTab === 'import-export'}
			<!-- Import/Export Tab -->
			<div class="space-y-6">
				<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
					<!-- Import Section -->
					<Card class="p-4">
						<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Import Dashboard</h3>

						<div class="space-y-4">
							<div>
								<Label class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
									Dashboard Configuration (JSON)
								</Label>
								<Textarea
									bind:value={importData}
									placeholder="Paste your dashboard configuration JSON here..."
									rows={8}
									class="font-mono text-xs"
								/>
							</div>

							<CompatButton
								color="blue"
								class="w-full"
								disabled={!importData.trim() || loading}
								on:click={importDashboard}
							>
								{#if loading}
									<Spinner size="4" class="mr-2" />
								{/if}
								<Upload class="mr-2 h-4 w-4" />
								Import Dashboard
							</CompatButton>
						</div>
					</Card>

					<!-- Export Section -->
					<Card class="p-4">
						<h3 class="mb-4 text-lg font-medium text-gray-900 dark:text-white">Export Dashboard</h3>

						<div class="space-y-4">
							<p class="text-sm text-gray-600 dark:text-gray-400">
								Select a dashboard to export its configuration for backup or sharing.
							</p>

							<div class="space-y-2">
								{#each dashboards as dashboard (dashboard.id)}
									<div
										class="flex items-center justify-between rounded-lg border border-gray-200 p-3 dark:border-gray-700"
									>
										<div>
											<p class="font-medium text-gray-900 dark:text-white">{dashboard.name}</p>
											<p class="text-xs text-gray-500 dark:text-gray-400">
												{dashboard.widgets.length} widgets
											</p>
										</div>
										<CompatButton
											size="xs"
											color="light"
											on:click={() => exportDashboard(dashboard)}
										>
											<Download class="h-3 w-3" />
											Export
										</CompatButton>
									</div>
								{/each}
							</div>
						</div>
					</Card>
				</div>
			</div>
		{/if}
	</div>
</Modal>

<!-- Create Dashboard Modal -->
<Modal bind:open={showCreateModal} size="lg">
	<div class="p-6">
		<h3 class="mb-4 text-xl font-bold text-gray-900 dark:text-white">Create New Dashboard</h3>

		<div class="space-y-4">
			<div>
				<Label for="dashboard-name" class="mb-2">Dashboard Name</Label>
				<Input
					id="dashboard-name"
					bind:value={newDashboardName}
					placeholder="Enter dashboard name"
				/>
			</div>

			<div>
				<Label for="dashboard-description" class="mb-2">Description (Optional)</Label>
				<Textarea
					id="dashboard-description"
					bind:value={newDashboardDescription}
					placeholder="Brief description of this dashboard"
					rows={3}
				/>
			</div>

			<div>
				<Label for="template-select" class="mb-2">Start from Template (Optional)</Label>
				<Select id="template-select" bind:value={selectedTemplate}>
					<option value="">Blank Dashboard</option>
					{#each dashboardTemplates as template}
						<option value={template.id}>{template.name} - {template.description}</option>
					{/each}
				</Select>
			</div>
		</div>

		<div class="mt-6 flex justify-end space-x-3">
			<CompatButton color="light" on:click={() => (showCreateModal = false)}>Cancel</CompatButton>
			<CompatButton
				color="blue"
				disabled={!newDashboardName.trim() || loading}
				on:click={createDashboard}
			>
				{#if loading}
					<Spinner size="4" class="mr-2" />
				{/if}
				Create Dashboard
			</CompatButton>
		</div>
	</div>
</Modal>

<!-- Delete Confirmation Modal -->
<Modal bind:open={showDeleteModal} size="md">
	<div class="p-6">
		<h3 class="mb-4 text-xl font-bold text-gray-900 dark:text-white">Delete Dashboard</h3>
		<p class="mb-6 text-gray-600 dark:text-gray-400">
			Are you sure you want to delete "{dashboardToDelete?.name}"? This action cannot be undone.
		</p>

		<div class="flex justify-end space-x-3">
			<CompatButton color="light" on:click={() => (showDeleteModal = false)}>Cancel</CompatButton>
			<CompatButton color="red" disabled={loading} on:click={deleteDashboard}>
				{#if loading}
					<Spinner size="4" class="mr-2" />
				{/if}
				Delete Dashboard
			</CompatButton>
		</div>
	</div>
</Modal>

<!-- Rename Modal -->
<Modal bind:open={showRenameModal} size="md">
	<div class="p-6">
		<h3 class="mb-4 text-xl font-bold text-gray-900 dark:text-white">Rename Dashboard</h3>

		<div class="mb-6">
			<Label for="rename-input" class="mb-2">Dashboard Name</Label>
			<Input id="rename-input" bind:value={renameValue} placeholder="Enter new name" />
		</div>

		<div class="flex justify-end space-x-3">
			<CompatButton color="light" on:click={() => (showRenameModal = false)}>Cancel</CompatButton>
			<CompatButton
				color="blue"
				disabled={!renameValue.trim() || loading}
				on:click={renameDashboard}
			>
				{#if loading}
					<Spinner size="4" class="mr-2" />
				{/if}
				Rename
			</CompatButton>
		</div>
	</div>
</Modal>

<!-- Export Modal -->
<Modal bind:open={showExportModal} size="lg">
	<div class="p-6">
		<h3 class="mb-4 text-xl font-bold text-gray-900 dark:text-white">Export Dashboard</h3>

		<div class="space-y-4">
			<div>
				<Label class="mb-2">Configuration JSON</Label>
				<Textarea bind:value={exportData} readonly rows={12} class="font-mono text-xs" />
			</div>

			<div class="flex justify-between">
				<div class="space-x-2">
					<CompatButton color="light" size="sm" on:click={copyExportData}>
						<Copy class="mr-2 h-4 w-4" />
						Copy to Clipboard
					</CompatButton>
					<CompatButton
						color="light"
						size="sm"
						on:click={() => downloadExportData(dashboardToDelete!)}
					>
						<Download class="mr-2 h-4 w-4" />
						Download File
					</CompatButton>
				</div>
				<CompatButton color="blue" on:click={() => (showExportModal = false)}>Done</CompatButton>
			</div>
		</div>
	</div>
</Modal>
