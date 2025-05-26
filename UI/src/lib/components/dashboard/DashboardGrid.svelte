<script lang="ts">
	import { onMount } from 'svelte';
	import { dashboardStore, draggedWidget, isEditMode } from '$lib/stores/dashboard';
	import type { Widget, WidgetPosition, DropResult } from '$lib/types/dashboard';
	import WidgetContainer from './WidgetContainer.svelte';
	import WidgetRenderer from './WidgetRenderer.svelte';
	import WidgetConfigModal from './widgets/WidgetConfigModal.svelte';

	export let dashboardId: string;
	export let gridCols = 12;
	export let gridRows = 8;

	let gridElement: HTMLElement;
	let isDragOver = false;
	let dropPreview: WidgetPosition | null = null;
	let configModal = {
		isOpen: false,
		widget: null as Widget | null
	};

	$: dashboard = $dashboardStore.dashboards.find((d) => d.id === dashboardId);
	$: widgets = dashboard?.widgets || [];

	// Grid styling
	$: gridStyle = `
		display: grid;
		grid-template-columns: repeat(${gridCols}, 1fr);
		grid-template-rows: repeat(${gridRows}, minmax(120px, 1fr));
		gap: 1rem;
		padding: 1rem;
		min-height: calc(100vh - 200px);
		position: relative;
	`;

	function getGridPosition(event: DragEvent): WidgetPosition | null {
		if (!gridElement) return null;

		const rect = gridElement.getBoundingClientRect();
		const x = event.clientX - rect.left;
		const y = event.clientY - rect.top;

		const cellWidth = (rect.width - (gridCols - 1) * 16) / gridCols; // 16px gap
		const cellHeight = (rect.height - (gridRows - 1) * 16) / gridRows;

		const gridX = Math.floor(x / (cellWidth + 16));
		const gridY = Math.floor(y / (cellHeight + 16));

		if (gridX >= 0 && gridX < gridCols && gridY >= 0 && gridY < gridRows) {
			return { x: gridX, y: gridY, width: 2, height: 2 }; // Default size
		}

		return null;
	}

	function canPlaceWidget(position: WidgetPosition, excludeWidgetId?: string): boolean {
		const endX = position.x + position.width;
		const endY = position.y + position.height;

		// Check bounds
		if (endX > gridCols || endY > gridRows) return false;

		// Check for overlaps
		for (const widget of widgets) {
			if (excludeWidgetId && widget.id === excludeWidgetId) continue;

			const widgetEndX = widget.position.x + widget.position.width;
			const widgetEndY = widget.position.y + widget.position.height;

			if (
				!(
					position.x >= widgetEndX ||
					endX <= widget.position.x ||
					position.y >= widgetEndY ||
					endY <= widget.position.y
				)
			) {
				return false;
			}
		}

		return true;
	}

	function handleDragOver(event: DragEvent) {
		if (!$isEditMode) return;

		event.preventDefault();
		isDragOver = true;

		const position = getGridPosition(event);
		if (position && $draggedWidget) {
			// Use dragged widget's size
			position.width = $draggedWidget.position.width;
			position.height = $draggedWidget.position.height;

			if (canPlaceWidget(position, $draggedWidget.id)) {
				dropPreview = position;
			} else {
				dropPreview = null;
			}
		}
	}

	function handleDragLeave(event: DragEvent) {
		// Only clear if leaving the grid entirely
		if (!gridElement.contains(event.relatedTarget as Node)) {
			isDragOver = false;
			dropPreview = null;
		}
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		isDragOver = false;

		if (!$draggedWidget || !dropPreview) {
			dropPreview = null;
			return;
		}

		const position = dropPreview;
		dropPreview = null;

		// Move widget to new position
		dashboardStore.moveWidget(dashboardId, $draggedWidget.id, position);
		draggedWidget.set(null);
	}

	function handleWidgetMove(event: CustomEvent<{ widget: Widget; position: WidgetPosition }>) {
		const { widget, position } = event.detail;
		if (canPlaceWidget(position, widget.id)) {
			dashboardStore.moveWidget(dashboardId, widget.id, position);
		}
	}

	function handleWidgetResize(event: CustomEvent<{ widget: Widget; position: WidgetPosition }>) {
		const { widget, position } = event.detail;
		if (canPlaceWidget(position, widget.id)) {
			dashboardStore.resizeWidget(dashboardId, widget.id, position);
		}
	}

	function handleWidgetRemove(event: CustomEvent<{ widget: Widget }>) {
		dashboardStore.removeWidget(dashboardId, event.detail.widget.id);
	}

	function handleWidgetConfigure(event: CustomEvent<{ widget: Widget }>) {
		configModal.widget = event.detail.widget;
		configModal.isOpen = true;
	}

	function handleConfigSave(event: CustomEvent<{ config: any }>) {
		if (configModal.widget) {
			dashboardStore.updateWidget(dashboardId, configModal.widget.id, {
				config: event.detail.config,
				updatedAt: new Date().toISOString()
			});
		}
		configModal.isOpen = false;
		configModal.widget = null;
	}

	function handleConfigCancel() {
		configModal.isOpen = false;
		configModal.widget = null;
	}

	function handleWidgetDuplicate(event: CustomEvent<{ widget: Widget }>) {
		const { widget } = event.detail;

		// Find available position
		let newPosition = { ...widget.position, x: widget.position.x + widget.position.width };
		if (newPosition.x + newPosition.width > gridCols) {
			newPosition.x = 0;
			newPosition.y = widget.position.y + widget.position.height;
		}

		if (canPlaceWidget(newPosition)) {
			dashboardStore.duplicateWidget(dashboardId, widget.id, newPosition);
		}
	}

	onMount(() => {
		// Load dashboard data
		dashboardStore.load();
	});
</script>

<div
	bind:this={gridElement}
	class="dashboard-grid relative"
	style={gridStyle}
	on:dragover={handleDragOver}
	on:dragleave={handleDragLeave}
	on:drop={handleDrop}
	role="grid"
	tabindex="0"
>
	<!-- Widgets -->
	{#each widgets as widget (widget.id)}
		<WidgetContainer
			{widget}
			gridSize={{ width: gridCols, height: gridRows }}
			on:move={handleWidgetMove}
			on:resize={handleWidgetResize}
			on:remove={handleWidgetRemove}
			on:configure={handleWidgetConfigure}
			on:duplicate={handleWidgetDuplicate}
		>
			<WidgetRenderer {widget} />
		</WidgetContainer>
	{/each}

	<!-- Drop Preview -->
	{#if isDragOver && dropPreview}
		<div
			class="drop-preview rounded-lg border-2 border-dashed border-blue-400 bg-blue-50 opacity-50"
			style="
				grid-column: {dropPreview.x + 1} / span {dropPreview.width};
				grid-row: {dropPreview.y + 1} / span {dropPreview.height};
			"
		></div>
	{/if}

	<!-- Empty State -->
	{#if widgets.length === 0}
		<div class="col-span-full row-span-full flex items-center justify-center">
			<div class="text-center text-gray-500">
				<div class="mb-2 text-xl">📊</div>
				<h3 class="mb-2 text-lg font-medium">No widgets yet</h3>
				<p class="text-sm">
					{#if $isEditMode}
						Add widgets from the widget library to get started
					{:else}
						Enable edit mode to add widgets to your dashboard
					{/if}
				</p>
			</div>
		</div>
	{/if}
</div>

<!-- Widget Configuration Modal -->
<WidgetConfigModal
	bind:isOpen={configModal.isOpen}
	widget={configModal.widget}
	on:save={handleConfigSave}
	on:cancel={handleConfigCancel}
/>

<style>
	.dashboard-grid {
		background-image:
			linear-gradient(rgba(0, 0, 0, 0.05) 1px, transparent 1px),
			linear-gradient(90deg, rgba(0, 0, 0, 0.05) 1px, transparent 1px);
		background-size: 20px 20px;
	}

	.drop-preview {
		pointer-events: none;
		z-index: 1000;
	}
</style>
