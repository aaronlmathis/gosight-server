<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Widget, WidgetPosition } from '$lib/types/dashboard';
	import { draggedWidget, isEditMode } from '$lib/stores/dashboard';
	import { Card } from 'flowbite-svelte';
	import { GripVertical, X, Settings, Copy } from 'lucide-svelte';

	export let widget: Widget;
	export let gridSize: { width: number; height: number };
	export let isDragging = false;
	export let isSelected = false;

	const dispatch = createEventDispatcher<{
		move: { widget: Widget; position: WidgetPosition };
		resize: { widget: Widget; position: WidgetPosition };
		remove: { widget: Widget };
		configure: { widget: Widget };
		duplicate: { widget: Widget };
		select: { widget: Widget };
	}>();

	let isDragOver = false;
	let dragStartPos = { x: 0, y: 0 };
	let resizing = false;
	let resizeHandle = '';

	$: containerStyle = `
		grid-column: ${widget.position.x + 1} / span ${widget.position.width};
		grid-row: ${widget.position.y + 1} / span ${widget.position.height};
		transform: ${isDragging ? 'scale(1.02)' : 'scale(1)'};
		opacity: ${isDragging ? 0.8 : 1};
		z-index: ${isDragging ? 1000 : isSelected ? 10 : 1};
		transition: ${isDragging ? 'none' : 'all 0.2s ease'};
	`;

	function handleDragStart(event: DragEvent) {
		if (!$isEditMode) return;

		isDragging = true;
		draggedWidget.set(widget);

		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move';
			event.dataTransfer.setData('text/plain', widget.id);
		}

		const rect = (event.target as HTMLElement).getBoundingClientRect();
		dragStartPos = {
			x: event.clientX - rect.left,
			y: event.clientY - rect.top
		};
	}

	function handleDragEnd(event: DragEvent) {
		isDragging = false;
		draggedWidget.set(null);
	}

	function handleDragOver(event: DragEvent) {
		if (!$isEditMode) return;
		event.preventDefault();
		isDragOver = true;
	}

	function handleDragLeave(event: DragEvent) {
		isDragOver = false;
	}

	function handleDrop(event: DragEvent) {
		if (!$isEditMode) return;
		event.preventDefault();
		isDragOver = false;

		const draggedId = event.dataTransfer?.getData('text/plain');
		if (draggedId && draggedId !== widget.id) {
			// Handle widget reordering logic here
		}
	}

	function handleClick(event: Event) {
		// Don't handle click if it's on a button
		if (event.target instanceof HTMLElement) {
			if (event.target.tagName === 'BUTTON' || event.target.closest('button')) {
				console.log('Click on button detected, ignoring container click');
				return;
			}
		}
		console.log('Container clicked, dispatching select');
		dispatch('select', { widget });
	}

	function handleRemove(event: Event) {
		event.stopPropagation();
		dispatch('remove', { widget });
	}

	function handleConfigure(event: Event) {
		event.stopPropagation();
		event.preventDefault();
		dispatch('configure', { widget });
	}

	function handleDuplicate(event: Event) {
		event.stopPropagation();
		dispatch('duplicate', { widget });
	}

	// Resize handlers
	function handleResizeStart(event: MouseEvent, handle: string) {
		if (!$isEditMode) return;
		event.preventDefault();
		event.stopPropagation();

		resizing = true;
		resizeHandle = handle;

		const onMouseMove = (e: MouseEvent) => handleResizeMove(e);
		const onMouseUp = () => handleResizeEnd();

		document.addEventListener('mousemove', onMouseMove);
		document.addEventListener('mouseup', onMouseUp);
	}

	function handleResizeMove(event: MouseEvent) {
		if (!resizing) return;

		// Calculate new size based on mouse position and resize handle
		const container = (event.target as HTMLElement).closest('.widget-container');
		if (!container) return;

		const rect = container.getBoundingClientRect();
		const newWidth = Math.max(1, Math.round((event.clientX - rect.left) / gridSize.width));
		const newHeight = Math.max(1, Math.round((event.clientY - rect.top) / gridSize.height));

		const newPosition: WidgetPosition = {
			...widget.position,
			width: newWidth,
			height: newHeight
		};

		dispatch('resize', { widget, position: newPosition });
	}

	function handleResizeEnd() {
		resizing = false;
		resizeHandle = '';
		document.removeEventListener('mousemove', handleResizeMove);
		document.removeEventListener('mouseup', handleResizeEnd);
	}
</script>

<div
	class="widget-container relative"
	style={containerStyle}
	draggable={$isEditMode}
	on:dragstart={handleDragStart}
	on:dragend={handleDragEnd}
	on:dragover={handleDragOver}
	on:dragleave={handleDragLeave}
	on:drop={handleDrop}
	on:click={handleClick}
	role="button"
	tabindex="0"
	on:keydown={(e) => e.key === 'Enter' && handleClick()}
>
	<Card
		class="h-full w-full border-2 transition-all duration-200 {isSelected
			? 'border-blue-500 ring-2 ring-blue-200 dark:ring-blue-800'
			: isDragOver
				? 'border-blue-300 ring-1 ring-blue-100'
				: 'border-gray-200 dark:border-gray-700'} {isDragging
			? 'shadow-2xl'
			: 'shadow-sm hover:shadow-md'}"
	>
		<!-- Widget Header -->
		{#if $isEditMode || widget.config.showTitle}
			<div
				class="flex items-center justify-between border-b border-gray-200 p-3 dark:border-gray-700"
			>
				<div class="flex items-center space-x-2">
					{#if $isEditMode}
						<div class="cursor-move text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
							<GripVertical size={16} />
						</div>
					{/if}
					<h3 class="truncate text-sm font-medium text-gray-900 dark:text-gray-100">
						{widget.title}
					</h3>
				</div>

				{#if $isEditMode}
					<div class="flex items-center space-x-1">
						<button
							on:click={handleDuplicate}
							class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
							title="Duplicate widget"
						>
							<Copy size={14} />
						</button>
						<button
							on:click={handleConfigure}
							class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
							title="Configure widget"
						>
							<Settings size={14} />
						</button>
						<button
							on:click={handleRemove}
							class="rounded p-1 text-gray-400 hover:bg-red-100 hover:text-red-600 dark:hover:bg-red-900 dark:hover:text-red-300"
							title="Remove widget"
						>
							<X size={14} />
						</button>
					</div>
				{/if}
			</div>
		{/if}

		<!-- Widget Content -->
		<div class="p-4">
			<slot />
		</div>

		<!-- Resize Handles (only in edit mode) -->
		{#if $isEditMode && isSelected}
			<div
				class="absolute right-0 bottom-0 h-3 w-3 cursor-se-resize bg-blue-500 opacity-60 hover:opacity-100"
				on:mousedown={(e) => handleResizeStart(e, 'se')}
				role="button"
				tabindex="0"
			></div>
		{/if}
	</Card>
</div>

<style>
	.widget-container {
		min-height: 120px;
		min-width: 200px;
	}

	.widget-container:hover {
		z-index: 5;
	}
</style>
