<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Card, Badge, Input } from 'flowbite-svelte';
	import Modal from '$lib/components/Modal.svelte';
	import { WIDGET_TEMPLATES } from '$lib/widgets/templates';
	import type { WidgetTemplate, WidgetPosition } from '$lib/types/dashboard';
	import { Search, Plus } from 'lucide-svelte';

	export let isOpen = false;
	export let gridSize: { width: number; height: number };

	const dispatch = createEventDispatcher<{
		addWidget: { template: WidgetTemplate; position: WidgetPosition };
		close: void;
	}>();

	let searchTerm = '';
	let selectedCategory = 'All';

	$: categories = ['All', ...new Set(WIDGET_TEMPLATES.map((t) => t.category))];
	$: filteredTemplates = WIDGET_TEMPLATES.filter((template) => {
		const matchesSearch =
			template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
			template.description.toLowerCase().includes(searchTerm.toLowerCase());
		const matchesCategory = selectedCategory === 'All' || template.category === selectedCategory;
		return matchesSearch && matchesCategory;
	});

	function findAvailablePosition(width: number, height: number): WidgetPosition | null {
		// Try to find an available spot in the grid
		for (let y = 0; y <= gridSize.height - height; y++) {
			for (let x = 0; x <= gridSize.width - width; x++) {
				// This is a simplified placement - in real implementation,
				// we'd check for actual widget overlaps
				return { x, y, width, height };
			}
		}
		return null;
	}

	function handleAddWidget(template: WidgetTemplate) {
		const position = findAvailablePosition(template.defaultSize.width, template.defaultSize.height);

		if (position) {
			dispatch('addWidget', { template, position });
			isOpen = false;
		} else {
			alert('No space available for this widget size');
		}
	}

	// Icon mapping for widget types
	const iconMap: Record<string, string> = {
		'metric-card': '📊',
		'cpu-usage': '🖥️',
		'memory-usage': '💾',
		'disk-usage': '💿',
		'network-traffic': '🌐',
		'response-time': '⏱️',
		'error-rate': '⚠️',
		throughput: '📈',
		'alerts-list': '🚨',
		'recent-events': '📰',
		'quick-links': '🔗',
		notes: '📝',
		'status-overview': '✅',
		'uptime-monitor': '⏰',
		'service-health': '❤️',
		'custom-chart': '📉'
	};
</script>

<Modal bind:show={isOpen} size="lg" title="Add Widget" on:close={() => dispatch('close')}>
	<!-- Search and Filters -->
	<div class="mb-6 flex gap-4">
		<div class="relative flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform text-gray-400" />
			<Input bind:value={searchTerm} placeholder="Search widgets..." class="pl-10" />
		</div>
		<select
			bind:value={selectedCategory}
			class="rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:ring-2 focus:ring-blue-500"
		>
			{#each categories as category}
				<option value={category}>{category}</option>
			{/each}
		</select>
	</div>

	<!-- Widget Templates Grid -->
	<div class="grid max-h-96 grid-cols-1 gap-4 overflow-y-auto md:grid-cols-2 lg:grid-cols-3">
		{#each filteredTemplates as template}
			<div
				class="cursor-pointer rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-gray-700 dark:bg-gray-800"
				role="button"
				tabindex="0"
				on:click={() => handleAddWidget(template)}
				on:keydown={(e) => e.key === 'Enter' && handleAddWidget(template)}
			>
				<div class="flex h-full flex-col">
					<!-- Icon and Title -->
					<div class="mb-2 flex items-center gap-3">
						<div class="text-2xl">
							{iconMap[template.type] || '📊'}
						</div>
						<div class="flex-1">
							<h4 class="text-sm font-medium text-gray-900">{template.name}</h4>
							<Badge color="gray" class="mt-1 text-xs">{template.category}</Badge>
						</div>
					</div>

					<!-- Description -->
					<p class="mb-3 flex-1 text-xs text-gray-600">
						{template.description}
					</p>

					<!-- Size Info -->
					<div class="flex items-center justify-between text-xs text-gray-500">
						<span>Size: {template.defaultSize.width}×{template.defaultSize.height}</span>
						<button
							type="button"
							class="inline-flex items-center rounded bg-blue-600 px-2 py-1 text-xs font-medium text-white hover:bg-blue-700 focus:ring-4 focus:ring-blue-300 focus:outline-none dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
						>
							<Plus class="mr-1 h-3 w-3" />
							Add
						</button>
					</div>
				</div>
			</div>
		{/each}
	</div>

	{#if filteredTemplates.length === 0}
		<div class="py-8 text-center text-gray-500">
			<div class="mb-2 text-4xl">🔍</div>
			<p>No widgets found matching your search.</p>
		</div>
	{/if}
</Modal>
