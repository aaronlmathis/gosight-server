<script lang="ts">
	import { onMount } from 'svelte';
	import {
		resources,
		resourceCounts,
		loadResources,
		updateResourceTags
	} from '$lib/stores/resourceStore';
	import type { Resource } from '$lib/types/resource';

	let selectedResource: Resource | null = null;
	let editingTags = false;
	let newTags: Record<string, string> = {};
	let tagEntries: Array<{ key: string; value: string }> = [];

	onMount(() => {
		loadResources();
		// Refresh every 30 seconds
		const interval = setInterval(() => loadResources(), 30000);
		return () => clearInterval(interval);
	});

	function selectResource(resource: Resource) {
		selectedResource = resource;
		newTags = { ...resource.tags };
		editingTags = false;
	}

	function startEditingTags() {
		editingTags = true;
		// Convert tags object to array for editing
		tagEntries = Object.entries(newTags).map(([key, value]) => ({ key, value }));
	}

	async function saveTags() {
		if (selectedResource) {
			// Convert tag entries back to object
			const tagsObject: Record<string, string> = {};
			tagEntries.forEach((entry) => {
				if (entry.key.trim() && entry.value.trim()) {
					tagsObject[entry.key.trim()] = entry.value.trim();
				}
			});

			await updateResourceTags(selectedResource.id, tagsObject);
			editingTags = false;
			selectedResource.tags = { ...tagsObject };
			newTags = { ...tagsObject };
		}
	}

	function cancelEditTags() {
		editingTags = false;
		newTags = selectedResource ? { ...selectedResource.tags } : {};
		tagEntries = Object.entries(newTags).map(([key, value]) => ({ key, value }));
	}

	function addTag() {
		tagEntries = [...tagEntries, { key: '', value: '' }];
	}

	function removeTag(index: number) {
		tagEntries = tagEntries.filter((_, i) => i !== index);
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'online':
				return 'text-green-600';
			case 'offline':
				return 'text-red-600';
			case 'idle':
				return 'text-yellow-600';
			default:
				return 'text-gray-600';
		}
	}
</script>

<div class="p-6">
	<div class="mb-6">
		<h1 class="mb-4 text-2xl font-bold text-gray-900">Resources</h1>

		<!-- Resource Summary -->
		<div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-4">
			{#each Array.from($resourceCounts.entries()) as [kind, count]}
				<div class="rounded-lg bg-white p-4 shadow">
					<div class="text-sm font-medium text-gray-500 uppercase">{kind}</div>
					<div class="text-2xl font-bold text-gray-900">{count}</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
		<!-- Resource List -->
		<div class="lg:col-span-2">
			<div class="rounded-lg bg-white shadow">
				<div class="px-4 py-5 sm:p-6">
					<h3 class="mb-4 text-lg font-medium text-gray-900">All Resources</h3>
					<div class="space-y-2">
						{#each $resources as resource}
							<div
								class="cursor-pointer rounded-lg border p-3 hover:bg-gray-50 {selectedResource?.id ===
								resource.id
									? 'border-blue-500 bg-blue-50'
									: 'border-gray-200'}"
								on:click={() => selectResource(resource)}
							>
								<div class="flex items-start justify-between">
									<div>
										<div class="font-medium text-gray-900">
											{resource.display_name || resource.name}
										</div>
										<div class="text-sm text-gray-500">
											{resource.kind} • {resource.environment || 'unknown'}
										</div>
									</div>
									<span
										class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStatusColor(
											resource.status
										)}"
									>
										{resource.status}
									</span>
								</div>
								{#if resource.group}
									<div class="mt-1 text-xs text-gray-400">Group: {resource.group}</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- Resource Details -->
		<div class="lg:col-span-1">
			{#if selectedResource}
				<div class="rounded-lg bg-white shadow">
					<div class="px-4 py-5 sm:p-6">
						<h3 class="mb-4 text-lg font-medium text-gray-900">Resource Details</h3>

						<dl class="grid grid-cols-1 gap-x-4 gap-y-3 text-sm">
							<div>
								<dt class="font-medium text-gray-500">Name</dt>
								<dd class="text-gray-900">
									{selectedResource.display_name || selectedResource.name}
								</dd>
							</div>
							<div>
								<dt class="font-medium text-gray-500">Kind</dt>
								<dd class="text-gray-900">{selectedResource.kind}</dd>
							</div>
							<div>
								<dt class="font-medium text-gray-500">Status</dt>
								<dd class={getStatusColor(selectedResource.status)}>{selectedResource.status}</dd>
							</div>
							<div>
								<dt class="font-medium text-gray-500">Last Seen</dt>
								<dd class="text-gray-900">
									{new Date(selectedResource.last_seen).toLocaleString()}
								</dd>
							</div>
							{#if selectedResource.ip_address}
								<div>
									<dt class="font-medium text-gray-500">IP Address</dt>
									<dd class="text-gray-900">{selectedResource.ip_address}</dd>
								</div>
							{/if}
							{#if selectedResource.os}
								<div>
									<dt class="font-medium text-gray-500">OS</dt>
									<dd class="text-gray-900">{selectedResource.os}</dd>
								</div>
							{/if}
						</dl>

						<!-- Labels -->
						<div class="mt-6">
							<h4 class="mb-2 font-medium text-gray-500">Labels</h4>
							<div class="space-y-1">
								{#each Object.entries(selectedResource.labels) as [key, value]}
									<div class="rounded bg-gray-100 px-2 py-1 text-xs">
										<span class="font-medium">{key}:</span>
										{value}
									</div>
								{/each}
							</div>
						</div>

						<!-- Tags -->
						<div class="mt-6">
							<div class="mb-2 flex items-center justify-between">
								<h4 class="font-medium text-gray-500">Tags</h4>
								{#if !editingTags}
									<button
										class="text-xs text-blue-600 hover:text-blue-800"
										on:click={startEditingTags}
									>
										Edit
									</button>
								{/if}
							</div>

							{#if editingTags}
								<div class="space-y-2">
									{#each tagEntries as entry, i}
										<div class="flex space-x-2">
											<input
												type="text"
												placeholder="Key"
												bind:value={entry.key}
												class="flex-1 rounded border border-gray-300 px-2 py-1 text-xs"
											/>
											<input
												type="text"
												placeholder="Value"
												bind:value={entry.value}
												class="flex-1 rounded border border-gray-300 px-2 py-1 text-xs"
											/>
											<button
												class="text-xs text-red-600 hover:text-red-800"
												on:click={() => removeTag(i)}
											>
												×
											</button>
										</div>
									{/each}
									<div class="flex space-x-2">
										<button class="text-xs text-blue-600 hover:text-blue-800" on:click={addTag}>
											+ Add Tag
										</button>
									</div>
									<div class="mt-3 flex space-x-2">
										<button
											class="rounded bg-blue-600 px-3 py-1 text-xs text-white hover:bg-blue-700"
											on:click={saveTags}
										>
											Save
										</button>
										<button
											class="rounded bg-gray-300 px-3 py-1 text-xs text-gray-700 hover:bg-gray-400"
											on:click={cancelEditTags}
										>
											Cancel
										</button>
									</div>
								</div>
							{:else}
								<div class="space-y-1">
									{#each Object.entries(selectedResource.tags) as [key, value]}
										<div class="rounded bg-blue-100 px-2 py-1 text-xs">
											<span class="font-medium">{key}:</span>
											{value}
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				</div>
			{:else}
				<div class="rounded-lg bg-white shadow">
					<div class="px-4 py-5 sm:p-6">
						<p class="text-gray-500">Select a resource to view details</p>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>
