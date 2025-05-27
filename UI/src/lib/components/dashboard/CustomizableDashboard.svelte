<script lang="ts">
	import { onMount } from 'svelte';
	import { dashboardStore, isEditMode, currentDashboard } from '$lib/stores/dashboard';
	import type { WidgetTemplate, WidgetPosition } from '$lib/types/dashboard';
	import DashboardToolbar from './DashboardToolbar.svelte';
	import DashboardGrid from './DashboardGrid.svelte';
	import WidgetPicker from './WidgetPicker.svelte';
	import DashboardManager from './DashboardManager.svelte';

	let activeDashboardId = '';
	let showWidgetPicker = false;
	let showDashboardManager = false;
	let gridSize = { width: 12, height: 8 };

	// Subscribe to current dashboard store
	$: activeDashboardId = $currentDashboard;

	$: if ($dashboardStore.dashboards.length > 0 && !activeDashboardId) {
		activeDashboardId = $dashboardStore.dashboards[0].id;
		currentDashboard.set(activeDashboardId);
	}

	function handleAddWidget() {
		showWidgetPicker = true;
	}

	function handleWidgetSelected(
		event: CustomEvent<{ template: WidgetTemplate; position: WidgetPosition }>
	) {
		const { template, position } = event.detail;

		const newWidget = {
			id: `widget-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
			type: template.type,
			title: template.name,
			position,
			config: { ...template.defaultConfig }
		};

		dashboardStore.addWidget(activeDashboardId, newWidget);
		showWidgetPicker = false;
	}

	function handleCreateDashboard() {
		const name = prompt('Dashboard name:') || `Dashboard ${$dashboardStore.dashboards.length + 1}`;
		const description = prompt('Description (optional):') || '';

		const newDashboardData = {
			name,
			isDefault: false,
			widgets: [],
			layout: {
				columns: 12,
				rowHeight: 60,
				margin: [16, 16] as [number, number],
				padding: [20, 20] as [number, number]
			}
		};

		const dashboard = dashboardStore.addDashboard(newDashboardData);
		activeDashboardId = dashboard.id;
	}

	function handleDeleteDashboard(event: CustomEvent<{ id: string }>) {
		dashboardStore.deleteDashboard(event.detail.id);

		// Switch to first available dashboard
		if ($dashboardStore.dashboards.length > 0) {
			activeDashboardId = $dashboardStore.dashboards[0].id;
		}
	}

	function handleDuplicateDashboard(event: CustomEvent<{ id: string }>) {
		const original = $dashboardStore.dashboards.find((d) => d.id === event.detail.id);
		if (!original) return;

		const copyData = {
			name: `${original.name} (Copy)`,
			isDefault: false,
			widgets: original.widgets.map((widget) => ({
				...widget,
				id: `widget-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
			})),
			layout: { ...original.layout }
		};

		const copy = dashboardStore.addDashboard(copyData);
		activeDashboardId = copy.id;
	}

	function handleConfigureDashboard(event: CustomEvent<{ id: string }>) {
		showDashboardManager = true;
	}

	function handleDashboardSelected(event: CustomEvent<{ dashboardId: string }>) {
		dashboardStore.setActiveDashboard(event.detail.dashboardId);
		showDashboardManager = false;
	}

	function handleDashboardManagerClose() {
		showDashboardManager = false;
	}

	onMount(() => {
		// Initialize dashboard store
		dashboardStore.load();

		// Create default dashboard if none exist
		if ($dashboardStore.dashboards.length === 0) {
			const defaultDashboardData = {
				name: 'My Dashboard',
				isDefault: true,
				widgets: [],
				layout: {
					columns: 12,
					rowHeight: 60,
					margin: [16, 16] as [number, number],
					padding: [20, 20] as [number, number]
				}
			};

			const dashboard = dashboardStore.addDashboard(defaultDashboardData);
			currentDashboard.set(dashboard.id);
		} else {
			// Set the first dashboard as active if none is set
			if (
				!$currentDashboard ||
				!$dashboardStore.dashboards.find((d) => d.id === $currentDashboard)
			) {
				currentDashboard.set($dashboardStore.dashboards[0].id);
			}
		}
	});
</script>

<div class="flex h-screen flex-col bg-gray-50 dark:bg-gray-900">
	<!-- Toolbar -->
	<DashboardToolbar
		{activeDashboardId}
		on:addWidget={handleAddWidget}
		on:createDashboard={handleCreateDashboard}
		on:deleteDashboard={handleDeleteDashboard}
		on:duplicateDashboard={handleDuplicateDashboard}
		on:configureDashboard={handleConfigureDashboard}
	/>

	<!-- Main Dashboard Content -->
	<div class="flex-1 overflow-hidden">
		{#if activeDashboardId}
			<DashboardGrid
				dashboardId={activeDashboardId}
				gridCols={gridSize.width}
				gridRows={gridSize.height}
			/>
		{:else}
			<div class="flex h-full items-center justify-center">
				<div class="text-center">
					<div class="mb-4 text-6xl">📊</div>
					<h2 class="mb-2 text-2xl font-semibold text-gray-900 dark:text-white">
						Welcome to GoSight
					</h2>
					<p class="mb-4 text-gray-600 dark:text-gray-400">
						Create your first dashboard to get started
					</p>
					<button
						on:click={handleCreateDashboard}
						class="rounded-lg bg-blue-600 px-4 py-2 text-white transition-colors hover:bg-blue-700"
					>
						Create Dashboard
					</button>
				</div>
			</div>
		{/if}
	</div>

	<!-- Widget Picker Modal -->
	<WidgetPicker
		bind:isOpen={showWidgetPicker}
		{gridSize}
		on:addWidget={handleWidgetSelected}
		on:close={() => (showWidgetPicker = false)}
	/>

	<!-- Dashboard Manager Modal -->
	<DashboardManager
		bind:isOpen={showDashboardManager}
		currentDashboardId={activeDashboardId}
		on:selectDashboard={handleDashboardSelected}
		on:close={handleDashboardManagerClose}
	/>
</div>

<!-- Edit Mode Indicator -->
{#if $isEditMode}
	<div
		class="fixed right-4 bottom-4 flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-white shadow-lg"
	>
		<div class="h-2 w-2 animate-pulse rounded-full bg-white"></div>
		Edit Mode Active - Click gear icons to configure widgets
	</div>
{:else}
	<div
		class="fixed right-4 bottom-4 flex items-center gap-2 rounded-lg bg-gray-600 px-4 py-2 text-white shadow-lg"
	>
		<div class="h-2 w-2 rounded-full bg-white"></div>
		View Mode - Enable edit mode to configure widgets
	</div>
{/if}
