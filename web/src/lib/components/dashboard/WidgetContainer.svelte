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
Widget Container Component
Wraps individual dashboard widgets with drag handles, resize controls, and edit actions.
Uses shadcn/ui Card components for consistent styling.
-->

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Widget } from '$lib/types/dashboard';
  import { draggedWidget, isEditMode, selectedWidget } from '$lib/stores/dashboard';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import GripVerticalIcon from '@lucide/svelte/icons/grip-vertical';
  import XIcon from '@lucide/svelte/icons/x';
  import MaximizeIcon from '@lucide/svelte/icons/maximize';

  export let widget: Widget;

  const dispatch = createEventDispatcher<{
    move: { widget: Widget; position: { x: number; y: number } };
    resize: { widget: Widget; size: { width: number; height: number } };
    remove: { widget: Widget };
    configure: { widget: Widget };
  }>();

  let isDragging = false;
  let isResizing = false;
  let dragStartPos = { x: 0, y: 0 };
  let resizeStartSize = { width: 0, height: 0 };

  $: isSelected = $selectedWidget === widget.id;
  
  $: containerStyle = `
    grid-column: ${widget.position.x + 1} / span ${widget.position.width};
    grid-row: ${widget.position.y + 1} / span ${widget.position.height};
    min-height: ${widget.position.height * 120}px;
  `;

  function handleDragStart(event: DragEvent) {
    if (!$isEditMode) {
      event.preventDefault();
      return;
    }
    
    isDragging = true;
    draggedWidget.set(widget);
    selectedWidget.set(widget.id);
    
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', widget.id);
      
      // Create a custom drag image
      const dragImage = document.createElement('div');
      dragImage.className = 'bg-primary/20 border-2 border-dashed border-primary rounded-lg p-4 text-primary';
      dragImage.style.width = '200px';
      dragImage.style.height = '100px';
      dragImage.textContent = widget.title;
      document.body.appendChild(dragImage);
      event.dataTransfer.setDragImage(dragImage, 100, 50);
      
      // Clean up drag image after a short delay
      setTimeout(() => {
        document.body.removeChild(dragImage);
      }, 0);
    }
  }

  function handleDragEnd() {
    isDragging = false;
    draggedWidget.set(null);
  }

  function handleConfigure(event: MouseEvent) {
    event.stopPropagation();
    dispatch('configure', { widget });
  }

  function handleRemove(event: MouseEvent) {
    event.stopPropagation();
    dispatch('remove', { widget });
  }

  function handleSelect() {
    if ($isEditMode) {
      selectedWidget.set(widget.id);
    }
  }

  // Resize functionality
  function handleResizeStart(event: MouseEvent) {
    if (!$isEditMode) return;
    
    event.preventDefault();
    event.stopPropagation();
    
    isResizing = true;
    resizeStartSize = { width: widget.position.width, height: widget.position.height };
    selectedWidget.set(widget.id);
    
    const handleMouseMove = (e: MouseEvent) => {
      if (!isResizing) return;
      
      // Calculate new size based on mouse movement
      const deltaX = Math.round((e.clientX - event.clientX) / 100); // Approximate grid cell size
      const deltaY = Math.round((e.clientY - event.clientY) / 120); // Row height
      
      const newWidth = Math.max(1, resizeStartSize.width + deltaX);
      const newHeight = Math.max(1, resizeStartSize.height + deltaY);
      
      dispatch('resize', { 
        widget, 
        size: { width: newWidth, height: newHeight } 
      });
    };
    
    const handleMouseUp = () => {
      isResizing = false;
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
    
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }
</script>

<div
  class={cn(
    "widget-container relative transition-all duration-200 cursor-pointer",
    isDragging && "scale-105 opacity-80 z-50",
    isSelected && $isEditMode && "ring-2 ring-primary",
    !$isEditMode && "hover:shadow-md"
  )}
  style={containerStyle}
  draggable={$isEditMode}
  on:dragstart={handleDragStart}
  on:dragend={handleDragEnd}
  on:click={handleSelect}
  role="button"
  tabindex="0"
>
  <Card.Root class={cn(
    "h-full w-full transition-all duration-200 overflow-hidden",
    isDragging && "shadow-2xl border-primary",
    isSelected && $isEditMode && "border-primary",
    !$isEditMode && "hover:shadow-md"
  )}>
    
    <!-- Widget Header (Edit Mode Only) -->
    {#if $isEditMode}
      <Card.Header class="flex-row items-center justify-between space-y-0 pb-2 px-3 py-2 bg-muted/30">
        <div class="flex items-center space-x-2 min-w-0 flex-1">
          <Button 
            variant="ghost" 
            size="icon"
            class="cursor-move text-muted-foreground hover:text-foreground h-6 w-6 flex-shrink-0"
          >
            <GripVerticalIcon class="h-4 w-4" />
          </Button>
          <Card.Title class="text-sm font-medium truncate">{widget.title}</Card.Title>
        </div>
        
        <div class="flex items-center space-x-1 flex-shrink-0">
          <Button
            variant="ghost"
            size="icon"
            class="h-6 w-6 text-muted-foreground hover:text-foreground"
            on:click={handleConfigure}
          >
            <MaximizeIcon class="h-3 w-3" />
          </Button>
          
          <Button
            variant="ghost"
            size="icon"
            class="h-6 w-6 text-muted-foreground hover:text-destructive"
            on:click={handleRemove}
          >
            <XIcon class="h-3 w-3" />
          </Button>
        </div>
      </Card.Header>
    {/if}

    <!-- Widget Content -->
    <Card.Content class={cn(
      "flex-1 h-full overflow-hidden",
      $isEditMode ? "p-3" : "p-4"
    )}>
      <slot />
    </Card.Content>

    <!-- Resize Handle (Edit Mode Only) -->
    {#if $isEditMode && isSelected}
      <button
        class="absolute bottom-0 right-0 w-4 h-4 bg-primary/20 hover:bg-primary/40 cursor-se-resize border-l border-t border-primary/40 transition-colors"
        on:mousedown={handleResizeStart}
        aria-label="Resize widget"
      >
        <!-- Resize grip lines -->
        <div class="absolute bottom-1 right-1 w-2 h-2">
          <div class="absolute bottom-0 right-0 w-full h-px bg-primary/60"></div>
          <div class="absolute bottom-0 right-0 w-px h-full bg-primary/60"></div>
        </div>
      </button>
    {/if}
  </Card.Root>
</div>
