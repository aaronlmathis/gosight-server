<script lang="ts">
	import DashboardManager from '$lib/components/dashboard/DashboardManager.svelte';
	import DashboardToolbar from '$lib/components/dashboard/DashboardToolbar.svelte';
	import { isEditMode } from '$lib/stores/dashboard';

	let showDashboardManager = false;
	let selectedDashboard = '';

	function handleSelectDashboard(event: CustomEvent<{ dashboardId: string }>) {
		selectedDashboard = event.detail.dashboardId;
		console.log('Dashboard selected:', selectedDashboard);
		showDashboardManager = false;
	}

	function openDashboardManager() {
		showDashboardManager = true;
	}
</script>

<div class="p-8">
	<h1 class="mb-4 text-2xl font-bold">Dashboard Components Test</h1>

	<div class="mb-4">
		<strong>Selected Dashboard:</strong>
		{selectedDashboard || 'None'}
		<br />
		<strong>Edit Mode:</strong>
		{$isEditMode ? 'ON' : 'OFF'}
	</div>

	<div class="space-y-4">
		<div>
			<h2 class="mb-2 text-lg font-semibold">Dashboard Toolbar Test</h2>
			<DashboardToolbar
				activeDashboardId="test-dashboard"
				on:addWidget={() => alert('Add Widget clicked')}
				on:createDashboard={openDashboardManager}
				on:deleteDashboard={(e) => alert(`Delete dashboard: ${e.detail.id}`)}
				on:duplicateDashboard={(e) => alert(`Duplicate dashboard: ${e.detail.id}`)}
				on:configureDashboard={(e) => alert(`Configure dashboard: ${e.detail.id}`)}
			/>
		</div>

		<div>
			<h2 class="mb-2 text-lg font-semibold">Dashboard Manager Test</h2>
			<button
				type="button"
				class="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
				on:click={openDashboardManager}
			>
				Open Dashboard Manager
			</button>
		</div>
	</div>
</div>

<DashboardManager
	bind:isOpen={showDashboardManager}
	currentDashboardId={selectedDashboard}
	on:selectDashboard={handleSelectDashboard}
	on:close={() => (showDashboardManager = false)}
/>
