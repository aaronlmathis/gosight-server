<!-- 
Copyright (C) 2025 Aaron Mathis
This file is part of GoSight Server.

GoSight Server is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight Server is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight Server.  If not, see <https://www.gnu.org/licenses/>.
-->

<!--
Dashboard Grid Component
Main grid layout component that handles drag-and-drop positioning of widgets.
Provides visual feedback during drag operations and manages widget positioning.
-->

<script lang="ts">
  import { dashboardStore, draggedWidget, isEditMode, activeDashboard } from '$lib/stores/dashboard';
  import type { WidgetPosition, Widget } from '$lib/types/dashboard';
  import WidgetContainer from './WidgetContainer.svelte';
  import SampleWidget from './SampleWidget.svelte';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import PlusIcon from '@lucide/svelte/icons/plus';

  export let gridCols = 12;
  export let gridRows = 8;

  let gridElement: HTMLElement;
  let isDragOver = false;
  let dropPreview: WidgetPosition | null = null;

  $: widgets = $activeDashboard?.widgets || [];
  $: maxRows = Math.max(gridRows, ...widgets.map((w: Widget) => w.position.y + w.position.height));

  $: gridStyle = `
    display: grid;
    grid-template-columns: repeat(${gridCols}, 1fr);
    grid-template-rows: repeat(${maxRows}, minmax(120px, 1fr));
    gap: 1rem;
    padding: 1rem;
    min-height: calc(100vh - 300px);
  `;

  function getGridPosition(event: DragEvent): WidgetPosition | null {
    if (!gridElement) return null;

    const rect = gridElement.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    // Account for padding and gaps
    const cellWidth = (rect.width - 32 - (gridCols - 1) * 16) / gridCols; // 32px total padding, 16px gaps
    const cellHeight = 120 + 16; // 120px min height + 16px gap

    const gridX = Math.floor((x - 16) / cellWidth); // 16px left padding
    const gridY = Math.floor((y - 16) / cellHeight); // 16px top padding

    if (gridX >= 0 && gridX < gridCols && gridY >= 0 && gridY < maxRows) {
      return { x: gridX, y: gridY, width: 2, height: 2 };
    }

    return null;
  }
  function isPositionOccupied(position: WidgetPosition, excludeWidgetId?: string): boolean {
    return widgets.some((widget: Widget) => {
      if (excludeWidgetId && widget.id === excludeWidgetId) return false;
      
      const wPos = widget.position;
      return !(
        position.x >= wPos.x + wPos.width ||
        position.x + position.width <= wPos.x ||
        position.y >= wPos.y + wPos.height ||
        position.y + position.height <= wPos.y
      );
    });
  }
  function handleDragOver(event: DragEvent) {
    if (!$isEditMode || !$draggedWidget) return;

    event.preventDefault();
    isDragOver = true;

    const position = getGridPosition(event);
    if (position && $draggedWidget && typeof $draggedWidget === 'object' && 'position' in $draggedWidget) {
      position.width = $draggedWidget.position.width;
      position.height = $draggedWidget.position.height;
      
      // Ensure position is within bounds
      position.x = Math.max(0, Math.min(gridCols - position.width, position.x));
      position.y = Math.max(0, position.y);
      
      // Check if position would overlap with other widgets (excluding the dragged widget)
      const wouldOverlap = isPositionOccupied(position, $draggedWidget.id);
      
      dropPreview = wouldOverlap ? null : position;
    }
  }

  function handleDragLeave(event: DragEvent) {
    // Only hide preview if leaving the grid entirely
    const rect = gridElement?.getBoundingClientRect();
    if (rect) {
      const x = event.clientX;
      const y = event.clientY;
      const outsideGrid = x < rect.left || x > rect.right || y < rect.top || y > rect.bottom;
      
      if (outsideGrid) {
        isDragOver = false;
        dropPreview = null;
      }
    }
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    isDragOver = false;    if (!$draggedWidget || !dropPreview || typeof $draggedWidget !== 'object' || !('id' in $draggedWidget)) {
      dropPreview = null;
      return;
    }

    dashboardStore.moveWidget($draggedWidget.id, {
      x: dropPreview.x,
      y: dropPreview.y,
      width: dropPreview.width,
      height: dropPreview.height
    });
    
    dropPreview = null;
  }

  function handleWidgetMove(event: CustomEvent) {
    const { widget, position } = event.detail;
    dashboardStore.moveWidget(widget.id, position);
  }
  function handleWidgetResize(event: CustomEvent) {
    const { widget, size } = event.detail;
    dashboardStore.resizeWidget(widget.id, size);
  }

  function handleWidgetRemove(event: CustomEvent) {
    dashboardStore.removeWidget(event.detail.widget.id);
  }

  function handleWidgetConfigure(event: CustomEvent) {
    // TODO: Open widget configuration modal
    console.log('Configure widget:', event.detail.widget);
  }

  function handleAddWidget() {
    const position = dashboardStore.findEmptyPosition(3, 2);
    dashboardStore.addWidget({
      type: 'metric-card',
      title: `Widget ${widgets.length + 1}`,
      position,
      config: {}
    });
  }
</script>

<div
  bind:this={gridElement}  class={cn(
    "dashboard-grid relative",
    isDragOver && "bg-accent/10"
  )}
  style={gridStyle}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  aria-label="Dashboard widget grid"
>
  <!-- Existing Widgets -->
  {#each widgets as widget (widget.id)}
    <WidgetContainer 
      {widget} 
      on:move={handleWidgetMove}
      on:resize={handleWidgetResize}
      on:remove={handleWidgetRemove}
      on:configure={handleWidgetConfigure}
    >
      <SampleWidget {widget} />
    </WidgetContainer>
  {/each}

  <!-- Drop Preview -->
  {#if isDragOver && dropPreview}
    <div
      class="drop-preview rounded-xl border-2 border-dashed border-primary bg-primary/10 flex items-center justify-center text-primary text-sm font-medium"
      style="
        grid-column: {dropPreview.x + 1} / span {dropPreview.width};
        grid-row: {dropPreview.y + 1} / span {dropPreview.height};
      "
    >
      Drop here
    </div>
  {/if}

  <!-- Add Widget Button (Edit Mode) -->
  {#if $isEditMode}    <Card.Root class="dashboard-add-widget border-2 border-dashed border-muted-foreground/25 hover:border-primary/50 hover:bg-accent/50 transition-all duration-200 cursor-pointer" style="grid-column: 1 / span 3; grid-row: {maxRows + 1};">
      <Card.Content class="flex items-center justify-center h-full p-6" onclick={handleAddWidget}>
        <div class="text-center">
          <PlusIcon class="h-8 w-8 mx-auto mb-2 text-muted-foreground" />
          <div class="text-sm font-medium text-muted-foreground">Add Widget</div>
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  <!-- Empty State -->
  {#if widgets.length === 0}
    <div class="col-span-full row-span-full flex items-center justify-center" style="grid-row: 1 / span {Math.max(4, gridRows)};">
      <Card.Root class="max-w-md">
        <Card.Content class="text-center p-8">
          <div class="text-4xl mb-4">📊</div>
          <Card.Title class="mb-2">Welcome to your Dashboard</Card.Title>
          <Card.Description class="mb-4">
            Create your first dashboard by adding widgets. 
            {#if !$isEditMode}
              Enable edit mode to get started.
            {:else}
              Click "Add Widget" to begin.
            {/if}
          </Card.Description>          {#if $isEditMode}
            <button
              class="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              onclick={handleAddWidget}
            >
              <PlusIcon class="h-4 w-4 mr-2" />
              Add Your First Widget
            </button>
          {/if}
        </Card.Content>
      </Card.Root>
    </div>
  {/if}
</div>

<style>
  .dashboard-grid {
    transition: background-color 0.2s ease;
  }
  
  .drop-preview {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  .dashboard-add-widget:hover {
    transform: translateY(-1px);
  }
</style>
