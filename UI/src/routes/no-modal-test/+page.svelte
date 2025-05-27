<script>
	import NoModalDashboardManager from './+page.svelte';
	import { dashboardStore } from '$lib/stores/dashboard';
	import CompatButton from '$lib/components/CompatButton.svelte';
	import { Plus } from 'lucide-svelte';

	let showManager = true; // Start with it open for easier testing

	function testFunction() {
		console.log('Test function called successfully!');
		alert('Test function works!');
	}

	function openCreateModal() {
		console.log('openCreateModal called from wrapper');
		alert('Create Modal function called from wrapper!');
	}
</script>

<div class="p-8">
	<h1 class="mb-6 text-2xl font-bold">No-Modal Dashboard Manager Test</h1>

	<div class="mb-6 space-y-4">
		<h2 class="text-lg font-medium">Direct Button Tests (Outside any Modal)</h2>
		<div class="flex gap-2">
			<button
				class="rounded bg-red-500 px-4 py-2 text-white"
				on:click={() => alert('Simple button works!')}
			>
				Simple Button
			</button>

			<button class="rounded bg-green-500 px-4 py-2 text-white" on:click={testFunction}>
				Named Function Button
			</button>

			<CompatButton color="blue" size="sm" on:click={openCreateModal}>
				<Plus class="mr-2 h-4 w-4" />
				CompatButton Test
			</CompatButton>
		</div>
	</div>

	<button
		class="mb-4 rounded bg-blue-600 px-4 py-2 text-white"
		on:click={() => (showManager = !showManager)}
	>
		{showManager ? 'Hide' : 'Show'} Dashboard Manager (No Modal Component)
	</button>

	<!-- Dashboard store info -->
	<div class="mb-4 rounded bg-gray-100 p-4">
		<p>Dashboard store loaded: {$dashboardStore ? 'Yes' : 'No'}</p>
		<p>Dashboards count: {$dashboardStore?.dashboards?.length || 0}</p>
	</div>
</div>
