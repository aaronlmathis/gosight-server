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
Enhanced Dashboard Grid Component
Modern grid layout with grid lines support, drag-and-drop positioning of widgets.
Uses shadcn/ui styling and proper Svelte 5 patterns.
-->

<script lang="ts">
  import { dashboardStore, draggedWidget, isEditMode, showGridLines } from '$lib/stores/dashboard';
  import type { WidgetPosition, Widget } from '$lib/types/dashboard';
  import EnhancedWidgetContainer from './EnhancedWidgetContainer.svelte';
  import EnhancedSampleWidget from './EnhancedSampleWidget.svelte';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import { Plus } from 'lucide-svelte';

  const { gridCols = 12, gridRows = 8 } = $props();

  let gridElement: HTMLElement;
  let isDragOver = $state(false);
  let dropPreview: WidgetPosition | null = $state(null);

  // Derived state
  let widgets = $derived($dashboardStore.widgets);
  let editMode = $derived($isEditMode);
  let gridLinesVisible = $derived($showGridLines);
  let maxRows = $derived(Math.max(gridRows, ...widgets.map(w => w.position.y + w.position.height)));

  // Debug logging
  $effect(() => {
    console.log('Dashboard Grid State:', {
      widgets: widgets.length,
      editMode,
      gridLinesVisible,
      maxRows,
      widgetDetails: widgets.map(w => ({ 
        id: w.id, 
        title: w.title, 
        position: w.position 
      }))
    });
  });

  // Auto-enable grid lines when edit mode is enabled
  $effect(() => {
    if (editMode && !gridLinesVisible) {
      showGridLines.set(true);
    }
  });

  // Grid styling with conditional grid lines
  let gridStyle = $derived(`
    display: grid;
    grid-template-columns: repeat(${gridCols}, 1fr);
    grid-template-rows: repeat(${maxRows}, minmax(140px, auto));
    gap: 1rem;
    padding: 1rem;
    min-height: calc(100vh - 200px);
    position: relative;
    ${gridLinesVisible ? `
      background-image: 
        linear-gradient(to right, var(--color-border) 1px, transparent 1px),
        linear-gradient(to bottom, var(--color-border) 1px, transparent 1px);
      background-size: 20px 20px;
      background-position: 0 0;
    ` : ''}
    ${editMode ? `
      border: 2px dashed var(--color-border);
      border-radius: 0.5rem;
      background-color: var(--color-muted);
    ` : ''}
  `);

  // Show grid cell boundaries when dragging for visual debugging
  let showCellBoundaries = $derived(editMode && isDragOver);

  function getGridPosition(event: DragEvent): WidgetPosition | null {
    if (!gridElement) return null;

    const rect = gridElement.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    // Use a simpler approach that works better with CSS Grid
    // We'll calculate based on percentages of the container
    const containerWidth = rect.width;
    const containerHeight = rect.height;
    
    // Account for padding (1rem = 16px)
    const padding = 16;
    const gap = 16;
    
    const usableWidth = containerWidth - (2 * padding);
    const usableHeight = containerHeight - (2 * padding);
    
    const cellWidth = (usableWidth - (gridCols - 1) * gap) / gridCols;
    const cellHeight = (usableHeight - (maxRows - 1) * gap) / maxRows;
    
    // Calculate position relative to the usable area
    const relativeX = x - padding;
    const relativeY = y - padding;
    
    const gridX = Math.floor(relativeX / (cellWidth + gap));
    const gridY = Math.floor(relativeY / (cellHeight + gap));

    // Debug logging (more concise)
    console.log('Drop position:', { gridX, gridY, mouseX: x, mouseY: y });

    if (gridX >= 0 && gridX < gridCols && gridY >= 0) {
      return { x: gridX, y: gridY, width: 1, height: 1 };
    }

    return null;
  }

  function isPositionOccupied(position: WidgetPosition, excludeWidgetId?: string): boolean {
    return widgets.some(widget => {
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
    if (!editMode) return;

    event.preventDefault();
    isDragOver = true;

    const position = getGridPosition(event);
    if (position && $draggedWidget) {
      // For existing widgets, use their current size
      if (typeof $draggedWidget === 'object') {
        position.width = $draggedWidget.position.width;
        position.height = $draggedWidget.position.height;
      } else {
        // For new widgets from palette, use default size
        position.width = 3;
        position.height = 2;
      }

      // Ensure widget fits within grid bounds
      position.x = Math.max(0, Math.min(gridCols - position.width, position.x));
      position.y = Math.max(0, Math.min(maxRows - position.height, position.y));

      // Only show preview if position is not occupied
      if (!isPositionOccupied(position, typeof $draggedWidget === 'object' ? $draggedWidget.id : undefined)) {
        dropPreview = position;
      } else {
        dropPreview = null;
      }
    } else {
      dropPreview = null;
    }
  }

  function handleDragLeave(event: DragEvent) {
    // Only hide preview if we're leaving the grid entirely
    if (!gridElement?.contains(event.relatedTarget as Node)) {
      isDragOver = false;
      dropPreview = null;
    }
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    isDragOver = false;

    if (!editMode || !$draggedWidget || !dropPreview) {
      dropPreview = null;
      return;
    }

    const position = dropPreview;
    dropPreview = null;

    console.log('Dropping widget:', { draggedWidget: $draggedWidget, position });

    // Handle existing widget moves
    if (typeof $draggedWidget === 'object') {
      // Only update x and y, preserve original width and height
      const newPosition = {
        x: position.x,
        y: position.y,
        width: $draggedWidget.position.width,
        height: $draggedWidget.position.height
      };
      console.log('Moving widget to:', newPosition);
      dashboardStore.moveWidget($draggedWidget.id, newPosition);
    }
    // Handle new widget from palette
    else if (typeof $draggedWidget === 'string') {
      dashboardStore.addWidget({
        type: $draggedWidget as any,
        title: `New ${$draggedWidget}`,
        position,
        config: {}
      });
    }

    draggedWidget.set(null);
  }

  function handleWidgetMove(event: { widget: Widget; position: WidgetPosition }) {
    const { widget, position } = event;
    dashboardStore.moveWidget(widget.id, position);
  }

  function handleWidgetResize(event: { widget: Widget; size: Pick<WidgetPosition, 'width' | 'height'> }) {
    const { widget, size } = event;
    dashboardStore.resizeWidget(widget.id, size);
  }

  function handleWidgetRemove(event: { widget: Widget }) {
    dashboardStore.removeWidget(event.widget.id);
  }

  function handleWidgetConfigure(event: { widget: Widget }) {
    // TODO: Open widget configuration modal
    console.log('Configure widget:', event.widget);
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
  bind:this={gridElement}
  class={cn(
    "dashboard-grid relative",
    isDragOver && "bg-accent/10",
    gridLinesVisible && "grid-lines-visible"
  )}
  style={gridStyle}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
  role="application"
  aria-label="Dashboard widget grid"
>
  <!-- Debug Grid Cell Boundaries (only when dragging) -->
  {#if showCellBoundaries}
    {#each Array(gridCols * maxRows) as _, i}
      {@const col = i % gridCols}
      {@const row = Math.floor(i / gridCols)}
      <div
        class="absolute border border-red-300/50 pointer-events-none"
        style="
          grid-column: {col + 1};
          grid-row: {row + 1};
          background: rgba(255, 0, 0, 0.05);
          z-index: 999;
        "
      ></div>
    {/each}
  {/if}

  <!-- Existing Widgets -->
  {#each widgets as widget (widget.id)}
    <EnhancedWidgetContainer 
      {widget} 
      onmove={handleWidgetMove}
      onresize={handleWidgetResize}
      onremove={handleWidgetRemove}
      onconfigure={handleWidgetConfigure}
    >
      {#snippet children()}
        <EnhancedSampleWidget {widget} />
      {/snippet}
    </EnhancedWidgetContainer>
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

  <!-- Empty State -->
  {#if widgets.length === 0}
    <div
      class="col-span-full row-span-4 flex flex-col items-center justify-center text-center p-8"
    >
      <Card.Root class="w-full max-w-md">
        <Card.Content class="p-8">
          <div class="flex flex-col items-center gap-4">
            <div class="rounded-full bg-primary/10 p-3">
              <Plus class="h-8 w-8 text-primary" />
            </div>
            <div class="space-y-2">
              <h3 class="text-lg font-semibold">No widgets yet</h3>
              <p class="text-sm text-muted-foreground">
                {editMode 
                  ? 'Use the widget palette in the toolbar to add your first widget, or drag widgets from the "Add Widget" menu'
                  : 'Switch to edit mode to start adding widgets to your dashboard'
                }
              </p>
            </div>
            {#if editMode}
              <button
                onclick={handleAddWidget}
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-10 px-4 py-2 gap-2"
              >
                <Plus class="h-4 w-4" />
                Add Sample Widget
              </button>
            {/if}
          </div>
        </Card.Content>
      </Card.Root>
    </div>
  {/if}

  <!-- Edit Mode Helper Text -->
  {#if editMode && widgets.length === 0}
    <div class="col-span-full row-span-1 flex items-center justify-center">
      <div class="text-center p-4 bg-primary/5 rounded-lg border border-primary/20">
        <p class="text-sm text-primary font-medium">
          Grid lines visible • Click "Add Widget" or drag widgets here
        </p>
      </div>
    </div>
  {/if}
</div>

<style>
  .grid-lines-visible {
    position: relative;
  }

  .grid-lines-visible::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    z-index: 1;
  }

  .drop-preview {
    z-index: 10;
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 0.7;
    }
    50% {
      opacity: 1;
    }
  }
</style>