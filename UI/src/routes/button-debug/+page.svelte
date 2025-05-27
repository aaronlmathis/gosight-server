<script lang="ts">
	import CompatButton from '$lib/components/CompatButton.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { Plus, Edit3, Eye } from 'lucide-svelte';

	let showModal = false;
	let clickCount = 0;
	let lastClicked = 'None';

	function testFunction(buttonName: string) {
		clickCount++;
		lastClicked = buttonName;
		console.log(`${buttonName} clicked - Count: ${clickCount}`);
		alert(`${buttonName} clicked - Count: ${clickCount}`);
	}

	function openModal() {
		console.log('Opening modal');
		showModal = true;
	}

	function closeModal() {
		console.log('Closing modal');
		showModal = false;
	}

	// Mock template object
	const mockTemplate = {
		id: 'test-template',
		name: 'Test Template',
		description: 'A test template'
	};
</script>

<svelte:head>
	<title>Button Functionality Test</title>
</svelte:head>

<div class="container mx-auto p-8">
	<h1 class="mb-6 text-3xl font-bold">Button Functionality Test</h1>

	<div class="mb-6 rounded border border-blue-200 bg-blue-50 p-4">
		<p class="text-lg">
			<strong>Last Clicked:</strong> <span class="font-mono">{lastClicked}</span> |
			<strong>Total Clicks:</strong> <span class="font-mono">{clickCount}</span>
		</p>
	</div>

	<!-- Test buttons outside modal -->
	<div class="space-y-6">
		<div class="rounded-lg border p-6">
			<h2 class="mb-4 text-xl font-semibold">Test 1: Buttons Outside Modal</h2>
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<h3 class="font-medium">CompatButton Components:</h3>
					<CompatButton color="blue" size="sm" on:click={() => testFunction('CompatButton Blue')}>
						<Plus class="mr-2 h-4 w-4" />
						CompatButton Blue
					</CompatButton>
					<CompatButton color="green" size="sm" on:click={() => testFunction('CompatButton Green')}>
						<Edit3 class="mr-2 h-4 w-4" />
						CompatButton Green
					</CompatButton>
				</div>
				<div class="space-y-2">
					<h3 class="font-medium">Native HTML Buttons:</h3>
					<button
						type="button"
						class="inline-flex items-center rounded bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700"
						on:click={() => testFunction('Native Blue')}
					>
						<Plus class="mr-2 h-4 w-4" />
						Native Blue
					</button>
					<button
						type="button"
						class="inline-flex items-center rounded bg-green-600 px-3 py-2 text-sm font-medium text-white hover:bg-green-700"
						on:click={() => testFunction('Native Green')}
					>
						<Edit3 class="mr-2 h-4 w-4" />
						Native Green
					</button>
				</div>
			</div>
		</div>

		<div class="rounded-lg border p-6">
			<h2 class="mb-4 text-xl font-semibold">Test 2: Modal Trigger</h2>
			<CompatButton color="purple" on:click={openModal}>Open Modal for Testing</CompatButton>
		</div>
	</div>
</div>

<!-- Modal with buttons inside -->
<Modal bind:show={showModal} title="Modal Button Test" size="lg">
	<div class="space-y-6 p-6">
		<div class="rounded border p-4">
			<h3 class="mb-3 text-lg font-semibold">Buttons Inside Modal</h3>
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
				<div class="space-y-2">
					<h4 class="font-medium">CompatButton Components:</h4>
					<CompatButton
						color="blue"
						size="sm"
						class="w-full"
						on:click={() => testFunction('Modal CompatButton Blue')}
					>
						<Plus class="mr-2 h-4 w-4" />
						Modal CompatButton Blue
					</CompatButton>
					<CompatButton
						color="red"
						size="sm"
						class="w-full"
						on:click={() => testFunction('Modal CompatButton Red')}
					>
						<Eye class="mr-2 h-4 w-4" />
						Modal CompatButton Red
					</CompatButton>
				</div>
				<div class="space-y-2">
					<h4 class="font-medium">Native HTML Buttons:</h4>
					<button
						type="button"
						class="inline-flex w-full items-center justify-center rounded bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700"
						on:click={() => testFunction('Modal Native Blue')}
					>
						<Plus class="mr-2 h-4 w-4" />
						Modal Native Blue
					</button>
					<button
						type="button"
						class="inline-flex w-full items-center justify-center rounded bg-red-600 px-3 py-2 text-sm font-medium text-white hover:bg-red-700"
						on:click={() => testFunction('Modal Native Red')}
					>
						<Eye class="mr-2 h-4 w-4" />
						Modal Native Red
					</button>
				</div>
			</div>
		</div>

		<div class="rounded border p-4">
			<h3 class="mb-3 text-lg font-semibold">Template-style Button Test</h3>
			<div class="rounded-lg border bg-white p-4">
				<h4 class="mb-2 font-medium">{mockTemplate.name}</h4>
				<p class="mb-3 text-sm text-gray-600">{mockTemplate.description}</p>

				<CompatButton
					color="blue"
					size="sm"
					class="w-full"
					on:click={() => testFunction(`Template: ${mockTemplate.name}`)}
				>
					Use Template (CompatButton)
				</CompatButton>

				<button
					type="button"
					class="mt-2 w-full rounded-md border border-transparent bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700"
					on:click={() => testFunction(`Template Native: ${mockTemplate.name}`)}
				>
					Use Template (Native)
				</button>
			</div>
		</div>

		<div class="flex justify-end space-x-2 pt-4">
			<CompatButton color="gray" on:click={closeModal}>Close Modal</CompatButton>
		</div>
	</div>
</Modal>
