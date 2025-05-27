<!-- 
	This is a test page for debugging button component issues.
	Created during development to isolate CompatButton vs native button problems.
	Can be removed in production.
-->
<script lang="ts">
	import CompatButton from '$lib/components/CompatButton.svelte';
	import Modal from '$lib/components/Modal.svelte';

	let showModal = false;
	let message = 'No clicks yet';

	function testClick() {
		message = 'Button clicked at ' + new Date().toLocaleTimeString();
		console.log('Test click function called');
	}

	function openModal() {
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}
</script>

<svelte:head>
	<title>Simple Button Test</title>
</svelte:head>

<div class="container mx-auto p-8">
	<h1 class="mb-6 text-2xl font-bold">Simple Button Test</h1>

	<div class="space-y-4">
		<p class="text-lg">
			Current message: <span class="rounded bg-gray-100 px-2 py-1 font-mono">{message}</span>
		</p>

		<!-- Test 1: Regular button outside modal -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 1: Button outside modal</h2>
			<CompatButton color="blue" on:click={testClick}>Click me (outside modal)</CompatButton>
		</div>

		<!-- Test 2: Button that opens modal -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 2: Button that opens modal</h2>
			<CompatButton color="green" on:click={openModal}>Open Modal</CompatButton>
		</div>

		<!-- Test 3: Native HTML button -->
		<div class="rounded border p-4">
			<h2 class="mb-2 text-lg font-semibold">Test 3: Native HTML button</h2>
			<button class="rounded bg-red-500 px-4 py-2 text-white hover:bg-red-600" on:click={testClick}>
				Native button
			</button>
		</div>
	</div>
</div>

<!-- Modal with button inside -->
<Modal bind:show={showModal} title="Modal Test">
	<div class="space-y-4 p-6">
		<p>This is a modal with buttons inside.</p>

		<div class="space-y-2">
			<CompatButton color="blue" on:click={testClick}>CompatButton inside modal</CompatButton>

			<button
				class="block rounded bg-purple-500 px-4 py-2 text-white hover:bg-purple-600"
				on:click={testClick}
			>
				Native button inside modal
			</button>
		</div>

		<div class="flex justify-end space-x-2 pt-4">
			<CompatButton color="gray" on:click={closeModal}>Close</CompatButton>
		</div>
	</div>
</Modal>
