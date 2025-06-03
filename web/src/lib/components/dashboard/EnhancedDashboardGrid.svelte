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
  import { dashboardStore, draggedWidget, isEditMode, activeDashboard } from '$lib/stores/dashboardStore';
  import type { WidgetPosition, Widget } from '$lib/types/dashboard';
  import EnhancedWidgetContainer from './EnhancedWidgetContainer.svelte';
  import EnhancedSampleWidget from './EnhancedSampleWidget.svelte';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Plus } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { getCurrentBreakpoint, getWidgetSize, validateWidgetSize } from '$lib/configs/widget-sizing';

  const { gridCols = 12, gridRows = 8 } = $props();

  let gridElement: HTMLElement;
  let isDragOver = $state(false);
  let dropPreview: WidgetPosition | null = $state(null);
  let screenSize = $state('desktop');
  let currentGridCols = $state(gridCols);

  // Professional breakpoint system
  const BREAKPOINTS = {
    mobile: { width: 640, cols: 2 },
    tablet: { width: 1024, cols: 6 },
    desktop: { width: 1440, cols: 12 },
    ultrawide: { width: 2560, cols: 16 }
  };

  const WIDGET_CONSTRAINTS = {
    minWidth: 1,
    minHeight: 1,
    maxWidth: 8,
    maxHeight: 6,
    defaultWidth: 3,
    defaultHeight: 2
  };

  // Responsive grid calculation
  function updateGridCols() {
    const width = window.innerWidth;
    const oldScreenSize = screenSize;
    
    if (width < BREAKPOINTS.mobile.width) {
      screenSize = 'mobile';
      currentGridCols = BREAKPOINTS.mobile.cols;
    } else if (width < BREAKPOINTS.tablet.width) {
      screenSize = 'tablet';
      currentGridCols = BREAKPOINTS.tablet.cols;
    } else if (width < BREAKPOINTS.ultrawide.width) {
      screenSize = 'desktop';
      currentGridCols = BREAKPOINTS.desktop.cols;
    } else {
      screenSize = 'ultrawide';
      currentGridCols = BREAKPOINTS.ultrawide.cols;
    }

    console.log(`📱 Screen size: ${screenSize}, Grid cols: ${currentGridCols}`);
    
    // If screen size changed, adapt widget sizes
    if (oldScreenSize !== screenSize && $activeDashboard) {
      adaptWidgetsToBreakpoint();
    }
  }

  // Adapt all widgets to the current breakpoint
  function adaptWidgetsToBreakpoint() {
    if (!$activeDashboard) return;
    
    console.log(`🔄 Adapting widgets to ${screenSize} breakpoint`);
    
    $activeDashboard.widgets.forEach(widget => {
      const breakpoint = getCurrentBreakpoint();
      const idealSize = getWidgetSize(widget.type, breakpoint);
      
      // Only update if the size is significantly different and valid
      if (
        (Math.abs(widget.position.width - idealSize.width) > 1 || 
         Math.abs(widget.position.height - idealSize.height) > 1) &&
        validateWidgetSize(widget.type, idealSize)
      ) {
        const newPosition = {
          ...widget.position,
          width: idealSize.width,
          height: idealSize.height,
          // Ensure it still fits in the grid
          x: Math.min(widget.position.x, currentGridCols - idealSize.width)
        };
        
        console.log(`📏 Resizing ${widget.title} from ${widget.position.width}x${widget.position.height} to ${idealSize.width}x${idealSize.height}`);
        dashboardStore.moveWidget(widget.id, newPosition);
      }
    });
  }

  // Responsive widget position adjustment
  function adjustWidgetForScreenSize(widget: Widget): WidgetPosition {
    const pos = { ...widget.position };
    
    // Ensure widget fits in current grid
    pos.width = Math.min(pos.width, currentGridCols);
    pos.x = Math.min(pos.x, currentGridCols - pos.width);
    
    // Mobile-specific adjustments
    if (screenSize === 'mobile') {
      // On mobile, most widgets should be full-width
      pos.width = Math.min(pos.width, 2);
      pos.x = Math.min(pos.x, currentGridCols - pos.width);
    }
    
    return pos;
  }

  // Derived state from active dashboard
  let widgets = $derived($activeDashboard.widgets.map(w => ({
    ...w,
    position: adjustWidgetForScreenSize(w)
  })));
  
  let maxRows = $derived(Math.max(gridRows, ...widgets.map(w => w.position.y + w.position.height)));

  // SINGLE responsive grid styling declaration
  let gridStyle = $derived(`
    display: grid;
    grid-template-columns: repeat(${currentGridCols}, 1fr);
    grid-template-rows: repeat(${maxRows}, 140px);
    gap: ${screenSize === 'mobile' ? '0.5rem' : '1rem'};
    padding: ${screenSize === 'mobile' ? '0.5rem' : '1rem'};
    min-height: calc(100vh - 300px);
    width: 100%;
    box-sizing: border-box;
    ${$isEditMode ? `
      background-image: 
        linear-gradient(to right, rgba(148, 163, 184, 0.3) 1px, transparent 1px),
        linear-gradient(to bottom, rgba(148, 163, 184, 0.3) 1px, transparent 1px);
      background-size: calc(100% / ${currentGridCols}) ${screenSize === 'mobile' ? '121px' : '141px'};
      background-position: ${screenSize === 'mobile' ? '0.5rem 0.5rem' : '1rem 1rem'};
    ` : ''}
  `);

  // Constraint validation for new positions
  function validatePosition(position: WidgetPosition): WidgetPosition {
    return {
      x: Math.max(0, Math.min(currentGridCols - position.width, position.x)),
      y: Math.max(0, position.y),
      width: Math.max(WIDGET_CONSTRAINTS.minWidth, Math.min(WIDGET_CONSTRAINTS.maxWidth, Math.min(currentGridCols, position.width))),
      height: Math.max(WIDGET_CONSTRAINTS.minHeight, Math.min(WIDGET_CONSTRAINTS.maxHeight, position.height))
    };
  }

  function getGridPosition(event: DragEvent): WidgetPosition | null {
    if (!gridElement) return null;

    const rect = gridElement.getBoundingClientRect();
    const padding = screenSize === 'mobile' ? 8 : 16;
    const gap = screenSize === 'mobile' ? 8 : 16;
    
    const x = event.clientX - rect.left - padding;
    const y = event.clientY - rect.top - padding;

    const cellWidth = (rect.width - (padding * 2) - (currentGridCols - 1) * gap) / currentGridCols;
    const cellHeight = screenSize === 'mobile' ? 120 + gap : 140 + gap;

    const gridX = Math.floor(x / (cellWidth + gap));
    const gridY = Math.floor(y / cellHeight);

    if (gridX >= 0 && gridX < currentGridCols && gridY >= 0) {
      return validatePosition({ 
        x: gridX, 
        y: gridY, 
        width: WIDGET_CONSTRAINTS.defaultWidth, 
        height: WIDGET_CONSTRAINTS.defaultHeight 
      });
    }

    return null;
  }

  // Handle responsive changes
  onMount(() => {
    updateGridCols();
    
    const handleResize = () => updateGridCols();
    window.addEventListener('resize', handleResize);
    
    return () => window.removeEventListener('resize', handleResize);
  });

  // Debug logging
  $effect(() => {
    console.log('Dashboard Grid State Updated:', {
      dashboardId: $activeDashboard.id,
      dashboardName: $activeDashboard.name,
      widgets: widgets.length,
      editMode: $isEditMode,
      maxRows,
      screenSize,
      currentGridCols,
      widgetDetails: widgets.map(w => ({ 
        id: w.id, 
        title: w.title, 
        position: w.position
      }))
    });
  });

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
    if (!$isEditMode) return;

    event.preventDefault();
    event.dataTransfer!.dropEffect = 'move';
    isDragOver = true;

    const position = getGridPosition(event);
    if (position && $draggedWidget) {
      // For existing widgets, use their current size but validate
      if (typeof $draggedWidget === 'object') {
        position.width = $draggedWidget.position.width;
        position.height = $draggedWidget.position.height;
        // Validate the position
        const validatedPosition = validatePosition(position);
        position.width = validatedPosition.width;
        position.height = validatedPosition.height;
        position.x = validatedPosition.x;
        position.y = validatedPosition.y;
      }

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

    if (!$isEditMode || !$draggedWidget || !dropPreview) {
      dropPreview = null;
      draggedWidget.set(null);
      return;
    }

    const position = validatePosition(dropPreview);
    dropPreview = null;

    if (typeof $draggedWidget === 'object') {
      const newPosition = {
        x: position.x,
        y: position.y,
        width: $draggedWidget.position.width,
        height: $draggedWidget.position.height
      };
      dashboardStore.moveWidget($draggedWidget.id, validatePosition(newPosition));
      draggedWidget.set(null);
    } else if (typeof $draggedWidget === 'string') {
      dashboardStore.addWidget({
        type: $draggedWidget as any,
        title: `New ${$draggedWidget}`,
        position,
        config: {}
      });
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
    dashboardStore.moveWidget(event.widget.id, validatePosition(event.position));
  }

  function handleWidgetResize(event: { widget: Widget; size: Pick<WidgetPosition, 'width' | 'height'> }) {
    const newPosition = validatePosition({
      ...event.widget.position,
      ...event.size
    });
    dashboardStore.moveWidget(event.widget.id, newPosition);
  }

  function handleWidgetRemove(event: { widget: Widget }) {
    dashboardStore.removeWidget(event.widget.id);
  }

  function handleWidgetConfigure(event: { widget: Widget }) {
    console.log('Configure widget:', event.widget);
  }

  function handleAddWidget() {
    const position = dashboardStore.findEmptyPosition(
      WIDGET_CONSTRAINTS.defaultWidth, 
      WIDGET_CONSTRAINTS.defaultHeight
    );
    dashboardStore.addWidget({
      type: 'metric-card',
      title: `Widget ${widgets.length + 1}`,
      position: validatePosition(position),
      config: {}
    });
  }
</script>

<div
  bind:this={gridElement}
  class={cn(
    "dashboard-grid relative",
    isDragOver && "bg-accent/10",
    $isEditMode && "grid-lines-visible",
    screenSize === 'mobile' && "mobile-grid"
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
      <EnhancedSampleWidget {widget} />
    </EnhancedWidgetContainer>
  {/each}

  <!-- Drop Preview -->
  {#if dropPreview && $isEditMode}
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
  {#if $isEditMode && widgets.length === 0}
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

  .dashboard-grid.mobile-grid {
    /* Mobile-specific styling */
    min-height: auto;
  }

  .dashboard-grid.grid-lines-visible {
    background-attachment: local;
  }

  .drop-preview {
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 0.7; }
  }
</style>