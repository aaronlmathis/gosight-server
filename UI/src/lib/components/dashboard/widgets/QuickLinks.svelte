<script lang="ts">
	import type { Widget, QuickLink, WidgetData } from '$lib/types/dashboard';
	import { Card } from 'flowbite-svelte';
	import { ExternalLink, Plus, Edit3, Trash2, Link as LinkIcon } from 'lucide-svelte';
	import { isEditMode } from '$lib/stores/dashboard';
	import { onMount } from 'svelte';
	import { dataService } from '$lib/services/dataService';

	export let widget: Widget;

	let links: QuickLink[] = [];
	let loading = true;
	let error = '';

	// Load quick links data
	async function loadLinksData() {
		try {
			loading = true;
			error = '';

			const data = await dataService.getWidgetData(widget);

			if (data.status === 'error') {
				error = data.error || 'Failed to load links';
				// Fallback to default links
				links = getDefaultLinks();
				return;
			}

			if (data.links && Array.isArray(data.links)) {
				links = data.links;
			} else {
				// Use default links if none configured
				links = getDefaultLinks();
			}
		} catch (err) {
			console.error('Failed to load links:', err);
			error = err instanceof Error ? err.message : 'Failed to load links';
			links = getDefaultLinks();
		} finally {
			loading = false;
		}
	}

	// Get default links based on GoSight context
	function getDefaultLinks(): QuickLink[] {
		return (
			widget.config?.links || [
				{
					id: '1',
					title: 'System Overview',
					url: '/dashboard/system',
					description: 'View system metrics and health'
				},
				{
					id: '2',
					title: 'Alerts Management',
					url: '/alerts',
					description: 'Manage alerts and notifications'
				},
				{
					id: '3',
					title: 'Container Monitoring',
					url: '/containers',
					description: 'Monitor container status and logs'
				},
				{
					id: '4',
					title: 'Network Diagnostics',
					url: '/network',
					description: 'Network analysis and troubleshooting'
				},
				{
					id: '5',
					title: 'Event Timeline',
					url: '/events',
					description: 'View system events and activities'
				}
			]
		);
	}

	function addLink() {
		const title = prompt('Link title:');
		const url = prompt('URL:');
		const description = prompt('Description (optional):');

		if (title && url) {
			const newLink: QuickLink = {
				id: `link-${Date.now()}`,
				title,
				url,
				description: description || ''
			};
			links = [...links, newLink];
			// TODO: Save to widget config via dataService
		}
	}

	function editLink(index: number) {
		const link = links[index];
		const title = prompt('Link title:', link.title);
		const url = prompt('URL:', link.url);
		const description = prompt('Description:', link.description);

		if (title && url) {
			links[index] = {
				id: link.id, // Preserve the existing ID
				title,
				url,
				description: description || ''
			};
			links = links; // Trigger reactivity
			// TODO: Save to widget config via dataService
		}
	}

	function removeLink(index: number) {
		if (confirm('Remove this link?')) {
			links = links.filter((_, i) => i !== index);
			// TODO: Save to widget config via dataService
		}
	}

	onMount(async () => {
		await loadLinksData();
	});
</script>

<Card class="relative h-full">
	<div class="flex h-full flex-col">
		<!-- Header -->
		{#if widget.config?.showTitle !== false}
			<div
				class="mb-3 flex items-center justify-between text-sm font-medium text-gray-900 dark:text-gray-100"
			>
				<span>{widget.title}</span>
				<div class="flex items-center gap-2">
					{#if loading}
						<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-500"></div>
					{/if}
					{#if $isEditMode}
						<button
							on:click={addLink}
							class="p-1 text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
							title="Add link"
						>
							<Plus class="h-4 w-4" />
						</button>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Error State -->
		{#if error}
			<div class="flex flex-1 items-center justify-center">
				<div class="text-center">
					<div class="mb-2 text-sm text-red-500 dark:text-red-400">
						<LinkIcon class="mx-auto mb-2 h-8 w-8" />
						Failed to load links
					</div>
					<div class="text-xs text-gray-500 dark:text-gray-400">{error}</div>
				</div>
			</div>
		{:else if loading}
			<!-- Loading State -->
			<div class="flex flex-1 items-center justify-center">
				<div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-500"></div>
			</div>
		{:else if links.length === 0}
			<!-- Empty State -->
			<div class="flex flex-1 items-center justify-center">
				<div class="text-center">
					<LinkIcon class="mx-auto mb-2 h-8 w-8 text-gray-400" />
					<div class="mb-2 text-sm text-gray-500 dark:text-gray-400">No quick links</div>
					{#if $isEditMode}
						<button
							on:click={addLink}
							class="text-xs font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300"
						>
							Add your first link
						</button>
					{/if}
				</div>
			</div>
		{:else}
			<!-- Links List -->
			<div class="flex-1 space-y-2 overflow-y-auto">
				{#each links as link, index}
					<div class="group relative">
						<a
							href={link.url}
							class="flex items-center gap-3 rounded-lg border border-gray-200 bg-white p-3 transition-all hover:border-blue-300 hover:bg-gray-50 hover:shadow-sm dark:border-gray-700 dark:bg-gray-800 dark:hover:border-blue-600 dark:hover:bg-gray-700"
						>
							<div class="flex-1">
								<div class="text-sm font-medium text-gray-900 dark:text-gray-100">
									{link.title}
								</div>
								{#if link.description}
									<div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
										{link.description}
									</div>
								{/if}
							</div>
							<ExternalLink
								class="h-4 w-4 text-gray-400 group-hover:text-blue-600 dark:text-gray-500 dark:group-hover:text-blue-400"
							/>
						</a>

						{#if $isEditMode}
							<div
								class="absolute top-2 right-2 opacity-0 transition-opacity group-hover:opacity-100"
							>
								<div
									class="flex gap-1 rounded border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-800"
								>
									<button
										on:click|preventDefault={() => editLink(index)}
										class="p-1 text-gray-400 hover:text-blue-600 dark:text-gray-500 dark:hover:text-blue-400"
										title="Edit link"
									>
										<Edit3 class="h-3 w-3" />
									</button>
									<button
										on:click|preventDefault={() => removeLink(index)}
										class="p-1 text-gray-400 hover:text-red-600 dark:text-gray-500 dark:hover:text-red-400"
										title="Remove link"
									>
										<Trash2 class="h-3 w-3" />
									</button>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
</Card>
