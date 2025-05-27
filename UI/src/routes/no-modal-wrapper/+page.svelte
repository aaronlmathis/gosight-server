<script>
	import NoModalDashboardManager from './+page.svelte';

	let showManager = false;
</script>

<div class="p-8">
	<h1 class="mb-6 text-2xl font-bold">No-Modal Dashboard Manager Test</h1>

	<button class="rounded bg-blue-600 px-4 py-2 text-white" on:click={() => (showManager = true)}>
		Open Dashboard Manager (No Modal Component)
	</button>

	<!-- Import the component that doesn't use the Modal wrapper -->
	<svelte:component
		this={NoModalDashboardManager}
		bind:isOpen={showManager}
		on:close={() => (showManager = false)}
	/>
</div>
