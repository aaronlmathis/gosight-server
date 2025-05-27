<!-- Test page to verify gear icon modal functionality -->
<script lang="ts">
	import { onMount } from 'svelte';
	import DashboardGrid from '$lib/components/dashboard/DashboardGrid.svelte';
	import type { Widget, Dashboard } from '$lib/types/dashboard';
	import { isEditMode, dashboardStore } from '$lib/stores/dashboard';

	// Set edit mode to true for testing
	isEditMode.set(true);

	// Create some test widgets to verify gear icon functionality
	const testWidgets: Widget[] = [
		{
			id: 'test-widget-1',
			type: 'chart',
			title: 'Test Chart Widget',
			position: { x: 0, y: 0, width: 6, height: 4 },
			config: {
				refreshInterval: 5000,
				showTitle: true
			},
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString()
		},
		{
			id: 'test-widget-2', 
			type: 'metric-card',
			title: 'Test Metric Widget',
			position: { x: 6, y: 0, width: 6, height: 4 },
			config: {
				refreshInterval: 10000,
				showTitle: true,
				metricType: 'cpu'
			},
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString()
		},
		{
			id: 'test-widget-3',
			type: 'system-status',
			title: 'Test System Status Widget', 
			position: { x: 0, y: 4, width: 4, height: 3 },
			config: {
				refreshInterval: 3000,
				showTitle: true
			},
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString()
		}
	];

	// Create a test dashboard and add it to the store
	let testDashboardId = 'test-dashboard';

	onMount(() => {
		// Add test dashboard to store
		const newDashboard = dashboardStore.addDashboard({
			name: 'Gear Icon Test Dashboard',
			description: 'Test dashboard for gear icon functionality',
			isDefault: false,
			widgets: testWidgets,
			layout: {
				columns: 12,
				rowHeight: 60,
				margin: [16, 16],
				padding: [20, 20]
			}
		});
		
		// Store the dashboard ID for use in the grid
		testDashboardId = newDashboard.id;
	});

	function handleWidgetUpdate(event: CustomEvent<Widget>) {
		console.log('Widget updated:', event.detail);
	}

	function handleWidgetDelete(event: CustomEvent<{ id: string }>) {
		console.log('Widget delete requested:', event.detail.id);
	}

	function handleWidgetConfigure(event: CustomEvent<{ widget: Widget }>) {
		console.log('Widget configure requested:', event.detail.widget);
	}
</script>

<div class="p-8">
	<h1 class="mb-6 text-2xl font-bold">Gear Icon Test Page</h1>
	
	<div class="mb-4 space-y-2">
		<p><strong>Edit Mode:</strong> {$isEditMode ? 'ON' : 'OFF'}</p>
		<p><strong>Instructions:</strong> The dashboard below is in edit mode. Click the gear icons in the top-right corner of each widget to test the configuration modal.</p>
		<p><strong>Expected Behavior:</strong> Clicking a gear icon should open the widget configuration modal.</p>
	</div>

	<div class="border-2 border-dashed border-gray-300 rounded-lg p-4">
		<h2 class="mb-4 text-lg font-semibold">Test Dashboard Grid</h2>
		
		<DashboardGrid
			dashboardId={testDashboardId}
			gridCols={12}
			gridRows={8}
			on:widgetUpdate={handleWidgetUpdate}
			on:widgetDelete={handleWidgetDelete}  
			on:widgetConfigure={handleWidgetConfigure}
		/>
	</div>

	<div class="mt-6 text-sm text-gray-600">
		<p><strong>Debug Info:</strong></p>
		<p>• Check browser console for debug logs when clicking gear icons</p>
		<p>• Each widget should show a gear icon in the top-right when hovering</p>
		<p>• Modal should open and close properly when gear icons are clicked</p>
	</div>
</div>
