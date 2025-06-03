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
Enhanced Widget Container Component
Modern widget wrapper with context menus, tooltips, animations, and advanced UX.
Uses all available shadcn/ui components for a sleek experience.
-->

<script lang="ts">
  import type { Widget, WidgetPosition } from '$lib/types/dashboard';
  import { isEditMode, draggedWidget } from '$lib/stores/dashboardStore';
  import { fade, scale } from 'svelte/transition';
  import { quintOut } from 'svelte/easing';
  import { Settings, X, Move, GripVertical } from 'lucide-svelte';
  import Button from '$lib/components/ui/button/button.svelte';

  const { 
    widget, 
    children,
    onmove = () => {},
    onresize = () => {},
    onremove = () => {},
    onconfigure = () => {}
  }: {
    widget: Widget;
    children?: any;
    onmove?: (event: { widget: Widget; position: WidgetPosition }) => void;
    onresize?: (event: { widget: Widget; size: Pick<WidgetPosition, 'width' | 'height'> }) => void;
    onremove?: (event: { widget: Widget }) => void;
    onconfigure?: (event: { widget: Widget }) => void;
  } = $props();

  let isDragging = $state(false);

  // Grid positioning styles - FIXED for CSS Grid
  let gridColumn = $derived(`${widget.position.x + 1} / span ${widget.position.width}`);
  let gridRow = $derived(`${widget.position.y + 1} / span ${widget.position.height}`);

  // Debug logging for position changes
  $effect(() => {
    console.log(`Widget ${widget.id} - Position updated:`, widget.position);
    console.log(`Widget ${widget.id} - Grid styles updated:`, { gridColumn, gridRow });
  });

  function handleDragStart(event: DragEvent) {
    if (!$isEditMode) {
      event.preventDefault();
      return;
    }

    console.log('🚀 Drag started for widget:', widget.id);
    isDragging = true;
    draggedWidget.set(widget);
    
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', widget.id);
    }
  }

  function handleDragEnd(event: DragEvent) {
    console.log('🏁 Drag ended for widget:', widget.id);
    isDragging = false;
  }

  function handleRemove() {
    onremove({ widget });
  }

  function handleConfigure() {
    onconfigure({ widget });
  }
</script>

<!-- FIXED: Direct CSS Grid child with proper styling -->
<div 
  class="widget-container group relative h-full rounded-lg border bg-card shadow-sm transition-all hover:shadow-md"
  class:ring-2={$isEditMode}
  class:ring-primary={$isEditMode}
  class:opacity-50={isDragging}
  class:scale-95={isDragging}
  style:grid-column={gridColumn}
  style:grid-row={gridRow}
  draggable={$isEditMode}
  ondragstart={handleDragStart}
  ondragend={handleDragEnd}
  role="button"
  aria-label="Draggable widget: {widget.title}"
  tabindex={$isEditMode ? 0 : -1}
  in:scale={{ duration: 300, easing: quintOut }}
  out:fade={{ duration: 200 }}
>
  <!-- Drag Handle - Restored -->
  {#if $isEditMode}
    <div class="absolute top-2 left-2 opacity-0 group-hover:opacity-100 transition-opacity cursor-move z-10">
      <GripVertical class="h-4 w-4 text-muted-foreground" />
    </div>
  {/if}

  <!-- Edit Mode Controls - Restored -->
  {#if $isEditMode}
    <div class="absolute -top-2 -right-2 flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
      <Button
        variant="secondary"
        size="sm"
        class="h-6 w-6 p-0 bg-background border shadow-sm"
        onclick={handleConfigure}
        title="Configure widget"
      >
        <Settings class="h-3 w-3" />
      </Button>
      <Button
        variant="destructive"
        size="sm"
        class="h-6 w-6 p-0"
        onclick={handleRemove}
        title="Remove widget"
      >
        <X class="h-3 w-3" />
      </Button>
    </div>
  {/if}

  <!-- Widget Content - Restored original structure -->
  <div class="flex h-full w-full flex-col overflow-hidden">
    <div class="flex-1 p-4">
      {@render children?.()}
    </div>
  </div>
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
