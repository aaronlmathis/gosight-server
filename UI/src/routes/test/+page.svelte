<script lang="ts">
	import DashboardManager from '$lib/components/dashboard/DashboardManager.svelte';
	import CompatButton from '$lib/components/CompatButton.svelte';

	let showManager = false;

	function openManager() {
		console.log('Opening dashboard manager...');
		console.log('showManager before:', showManager);
		showManager = true;
		console.log('showManager after:', showManager);
	}

	function handleClose() {
		console.log('Closing dashboard manager...');
		console.log('showManager before:', showManager);
		showManager = false;
		console.log('showManager after:', showManager);
	}

	// Debug the reactive statement
	$: console.log('showManager reactive:', showManager);
</script>

<svelte:head>
	<title>Dashboard Manager Test - GoSight</title>
</svelte:head>

<div class="p-8">
	<h1 class="mb-4 text-2xl font-bold">Dashboard Manager Test</h1>

	<CompatButton color="blue" on:click={openManager}>
		Open Dashboard Manager (Test Button - Should work!)
	</CompatButton>

	<p class="mt-2 text-sm">Manager state: {showManager ? 'OPEN' : 'CLOSED'}</p>

	<div class="mt-4 rounded bg-gray-100 p-4">
		<h2 class="mb-2 text-lg font-semibold">Test Instructions:</h2>
		<ol class="list-inside list-decimal space-y-1 text-sm">
			<li>Click "Open Dashboard Manager" above - it should open the modal</li>
			<li>Try clicking "Create Dashboard" button inside the modal</li>
			<li>Try clicking any dashboard action buttons</li>
			<li>Check browser console for logs</li>
		</ol>
	</div>

	<DashboardManager
		bind:isOpen={showManager}
		currentDashboardId="test-dashboard"
		on:close={handleClose}
		on:selectDashboard={(e) => console.log('Selected dashboard:', e.detail)}
	/>
</div>
