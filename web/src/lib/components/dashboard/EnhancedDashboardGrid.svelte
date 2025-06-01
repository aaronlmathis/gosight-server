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
  import { dashboardStore, draggedWidget, isEditMode, activeDashboard } from '$lib/stores/dashboard';
  import type { WidgetPosition, Widget } from '$lib/types/dashboard';
  import EnhancedWidgetContainer from './EnhancedWidgetContainer.svelte';
  import EnhancedSampleWidget from './EnhancedSampleWidget.svelte';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Plus } from 'lucide-svelte';

  const { gridCols = 12, gridRows = 8 } = $props();

  let gridElement: HTMLElement;
  let isDragOver = $state(false);
  let dropPreview: WidgetPosition | null = $state(null);
  
  // Derived state from active dashboard
  let widgets = $derived($activeDashboard.widgets);
  let editMode = $derived($isEditMode);
  let maxRows = $derived(Math.max(gridRows, ...widgets.map(w => w.position.y + w.position.height)));

  // Debug logging
  $effect(() => {
    console.log('Dashboard Grid State Updated:', {
      dashboardId: $activeDashboard.id,
      dashboardName: $activeDashboard.name,
      widgets: widgets.length,
      editMode,
      maxRows,
      widgetDetails: widgets.map(w => ({ 
        id: w.id, 
        title: w.title, 
        position: w.position
      }))
    });
  });

  // Grid styling with conditional grid lines - REMOVE DASHED BORDER
  let gridStyle = $derived(`
    display: grid;
    grid-template-columns: repeat(${gridCols}, 1fr);
    grid-template-rows: repeat(${maxRows}, minmax(140px, auto));
    gap: 1rem;
    padding: 1rem;
    min-height: calc(100vh - 200px);
    position: relative;
    ${editMode ? `
      background-image: 
        linear-gradient(to right, rgba(148, 163, 184, 0.3) 1px, transparent 1px),
        linear-gradient(to bottom, rgba(148, 163, 184, 0.3) 1px, transparent 1px);
      background-size: 20px 20px;
      background-position: 0 0;
    ` : ''}
  `);

  function getGridPosition(event: DragEvent): WidgetPosition | null {
    if (!gridElement) return null;

    const rect = gridElement.getBoundingClientRect();
    const x = event.clientX - rect.left - 16; // Account for padding
    const y = event.clientY - rect.top - 16;

    const cellWidth = (rect.width - 32) / gridCols; // Account for padding and gap
    const cellHeight = (rect.height - 32) / maxRows;

    const gridX = Math.floor(x / cellWidth);
    const gridY = Math.floor(y / cellHeight);

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
    event.dataTransfer!.dropEffect = 'move';
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

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    isDragOver = false;

    console.log('Drop event:', { 
      draggedWidget: $draggedWidget, 
      dropPreview, 
      editMode 
    });

    if (!editMode || !$draggedWidget || !dropPreview) {
      console.log('Drop cancelled:', { editMode, draggedWidget: $draggedWidget, dropPreview });
      dropPreview = null;
      draggedWidget.set(null);
      return;
    }

    const position = dropPreview;
    dropPreview = null;

    console.log('Processing drop:', { draggedWidget: $draggedWidget, position });

    try {
      // Handle existing widget moves
      if (typeof $draggedWidget === 'object') {
        const newPosition = {
          x: position.x,
          y: position.y,
          width: $draggedWidget.position.width,
          height: $draggedWidget.position.height
        };
        console.log('Moving widget:', $draggedWidget.id, 'to position:', newPosition);
        dashboardStore.moveWidget($draggedWidget.id, newPosition);
      }
      // Handle new widget from palette
      else if (typeof $draggedWidget === 'string') {
        console.log('Adding new widget:', $draggedWidget, 'at position:', position);
        dashboardStore.addWidget({
          type: $draggedWidget as any,
          title: `New ${$draggedWidget}`,
          position,
          config: {}
        });
      }
    } catch (error) {
      console.error('Error during drop:', error);
    } finally {
      draggedWidget.set(null);
    }
  }

  function handleDragLeave(event: DragEvent) {
    if (event.currentTarget === event.target) {
      isDragOver = false;
      dropPreview = null;
    }
  }

  // Widget event handlers
  function handleWidgetMove(event: { widget: Widget; position: WidgetPosition }) {
    dashboardStore.moveWidget(event.widget.id, event.position);
  }

  function handleWidgetResize(event: { widget: Widget; size: Pick<WidgetPosition, 'width' | 'height'> }) {
    const newPosition = {
      ...event.widget.position,
      ...event.size
    };
    dashboardStore.moveWidget(event.widget.id, newPosition);
  }

  function handleWidgetRemove(event: { widget: Widget }) {
    dashboardStore.removeWidget(event.widget.id);
  }

  function handleWidgetConfigure(event: { widget: Widget }) {
    console.log('Configure widget:', event.widget);
    // Widget configuration logic would go here
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
    editMode && "grid-lines-visible"
  )}
  style={gridStyle}
  ondragover={handleDragOver}
  ondrop={handleDrop}
  ondragleave={handleDragLeave}
  role="application"
  aria-label="Dashboard widget grid"
>
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
  {#if dropPreview && editMode}
    <div
      class="drop-preview absolute rounded-lg border-2 border-dashed border-primary bg-primary/10 pointer-events-none z-10"
      style={`
        grid-column: ${dropPreview.x + 1} / span ${dropPreview.width};
        grid-row: ${dropPreview.y + 1} / span ${dropPreview.height};
      `}
    >
      <div class="h-full w-full flex items-center justify-center text-primary font-medium">
        Drop Here
      </div>
    </div>
  {/if}

  <!-- Add Widget Button (when in edit mode and no widgets) -->
  {#if editMode && widgets.length === 0}
    <div class="col-span-full flex items-center justify-center py-12">
      <Card.Root class="w-80">
        <Card.Content class="flex flex-col items-center justify-center p-6">
          <Plus class="h-12 w-12 text-muted-foreground mb-4" />
          <h3 class="text-lg font-semibold mb-2">No widgets yet</h3>
          <p class="text-sm text-muted-foreground text-center mb-4">
            Add your first widget to get started with your dashboard.
          </p>
          <Button onclick={handleAddWidget}>
            <Plus class="mr-2 h-4 w-4" />
            Add Widget
          </Button>
        </Card.Content>
      </Card.Root>
    </div>
  {/if}
</div>

<style>
  .dashboard-grid {
    transition: all 0.3s ease;
  }

  .dashboard-grid.grid-lines-visible {
    background-attachment: local;
  }

  .drop-preview {
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 0.3;
    }
    50% {
      opacity: 0.7;
    }
  }
</style>