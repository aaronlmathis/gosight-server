<script lang="ts">
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

	function useTemplate(template: any) {
		console.log('useTemplate called with:', template);
		clicked = `Template: ${template.name}`;
		clickCount++;
		alert(`Using template: ${template.name}`);
	}

	function testClick() {
		console.log('Test click function called');
		clicked = 'Test click';
		clickCount++;
		alert('Test click works!');
	}
</script>

<div class="p-8">
	<h1 class="mb-4 text-2xl font-bold">Button Test Page</h1>

	<div class="mb-4">
		<strong>Status:</strong>
		{clicked} (clicked {clickCount} times)
	</div>

	<div class="space-y-4">
		<div>
			<h2 class="mb-2 text-lg font-semibold">Test Button (Known Working)</h2>
			<button
				type="button"
				class="rounded bg-green-600 px-4 py-2 text-white hover:bg-green-700"
				on:click={testClick}
			>
				Test Click
			</button>
		</div>

		<div>
			<h2 class="mb-2 text-lg font-semibold">Template Buttons (Data Attribute Method)</h2>
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

		<div>
			<h2 class="mb-2 text-lg font-semibold">Template Buttons (Arrow Function Method)</h2>
			<div class="grid grid-cols-1 gap-2 md:grid-cols-3">
				{#each testTemplates as template}
					<button
						type="button"
						class="w-full rounded-md bg-purple-600 px-3 py-2 text-sm font-medium text-white hover:bg-purple-700 focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 focus:outline-none"
						on:click={() => useTemplate(template)}
					>
						Use {template.name} (Arrow)
					</button>
				{/each}
			</div>
		</div>
	</div>
</div>
