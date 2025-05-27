<script lang="ts">
	import Modal from '$lib/components/Modal.svelte';

	let showModal = false;
	let clicked = '';
	let clickCount = 0;

	const testTemplates = [
		{ id: '1', name: 'Template One' },
		{ id: '2', name: 'Template Two' },
		{ id: '3', name: 'Template Three' }
	];

	function handleTemplateClick(event: MouseEvent) {
		const button = event.currentTarget as HTMLElement;
		const templateId = button.dataset.templateId;
		console.log('handleTemplateClick called with templateId:', templateId);
		clicked = `Template ID: ${templateId}`;
		clickCount++;
		alert(`Template clicked: ${templateId}`);
	}

	function openModal() {
		showModal = true;
	}

	function testClick() {
		console.log('Test click function called');
		clicked = 'Test click in modal';
		clickCount++;
		alert('Test click works in modal!');
	}
</script>

<div class="p-8">
	<h1 class="mb-4 text-2xl font-bold">Modal Button Test</h1>

	<div class="mb-4">
		<strong>Status:</strong>
		{clicked} (clicked {clickCount} times)
	</div>

	<button
		type="button"
		class="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
		on:click={openModal}
	>
		Open Modal Test
	</button>
</div>

<Modal bind:show={showModal} title="Button Test Modal" size="lg">
	<div class="space-y-4">
		<div>
			<h2 class="mb-2 text-lg font-semibold">Test Button (Known Working)</h2>
			<button
				type="button"
				class="rounded bg-green-600 px-4 py-2 text-white hover:bg-green-700"
				on:click={testClick}
			>
				Test Click in Modal
			</button>
		</div>

		<div>
			<h2 class="mb-2 text-lg font-semibold">Template Buttons in Modal</h2>
			<div class="grid grid-cols-1 gap-2 md:grid-cols-3">
				{#each testTemplates as template}
					<button
						type="button"
						class="w-full rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none"
						data-template-id={template.id}
						on:click={handleTemplateClick}
					>
						Use {template.name}
					</button>
				{/each}
			</div>
		</div>
	</div>
</Modal>
