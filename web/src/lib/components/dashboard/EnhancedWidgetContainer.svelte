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
  import { type Snippet } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import { quintOut } from 'svelte/easing';
  import * as Tooltip from '$lib/components/ui/tooltip';
  import * as ContextMenu from '$lib/components/ui/context-menu';
  import * as Dialog from '$lib/components/ui/dialog';
  import Button from '$lib/components/ui/button/button.svelte';
  import { Settings, Copy, Trash2 } from 'lucide-svelte';
  import type { Widget, WidgetPosition } from '$lib/types/dashboard';
  import { isEditMode, draggedWidget, dashboardStore } from '$lib/stores/dashboard';

  interface Props {
    widget: Widget;
    children?: Snippet;
    onmove?: (event: { widget: Widget; position: WidgetPosition }) => void;
    onresize?: (event: { widget: Widget; size: Pick<WidgetPosition, 'width' | 'height'> }) => void;
    onremove?: (event: { widget: Widget }) => void;
    onconfigure?: (event: { widget: Widget }) => void;
  }

  let { widget, children, onmove, onresize, onremove, onconfigure }: Props = $props();

  let showConfigDialog = $state(false);
  let isDragging = $state(false);

  // CRITICAL FIX: Use the widget prop directly for positioning
  let gridColumn = $derived(`${widget.position.x + 1} / span ${widget.position.width}`);
  let gridRow = $derived(`${widget.position.y + 1} / span ${widget.position.height}`);

  // Debug logging
  $effect(() => {
    console.log(`Widget ${widget.id} - Position:`, widget.position);
    console.log(`Widget ${widget.id} - Grid styles:`, { gridColumn, gridRow });
  });

  function handleDragStart(event: DragEvent) {
    if (!$isEditMode) return;

    isDragging = true;
    console.log('Drag started for widget:', widget);
    draggedWidget.set(widget);

    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', widget.id);
    }
  }

  function handleDragEnd() {
    console.log('Drag ended for widget:', widget.id);
    isDragging = false;
  }

  function handleConfigure() {
    onconfigure?.({ widget });
    showConfigDialog = true;
  }

  function handleDuplicate() {
    dashboardStore.addWidget({
      type: widget.type,
      title: `${widget.title} Copy`,
      position: { 
        x: widget.position.x + 1, 
        y: widget.position.y,
        width: widget.position.width,
        height: widget.position.height
      },
      config: widget.config || {}
    });
  }

  function handleRemove() {
    onremove?.({ widget });
  }

  function handleClose() {
    showConfigDialog = false;
  }
</script>

<ContextMenu.Root>
  <ContextMenu.Trigger class="block h-full">
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
      {#if $isEditMode}
        <Tooltip.Root>
          <Tooltip.Trigger>
            <Button
              variant="outline"
              size="sm"
              class="absolute -top-2 -right-2 z-10 h-6 w-6 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
              onclick={handleConfigure}
            >
              <Settings class="h-3 w-3" />
            </Button>
          </Tooltip.Trigger>
          <Tooltip.Content>
            <p>Configure Widget</p>
          </Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger>
            <div
              class="absolute top-2 left-2 z-10 h-6 w-6 cursor-grab active:cursor-grabbing opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-xs font-bold text-muted-foreground hover:text-foreground"
              title="Drag to move"
            >
              ⋮⋮
            </div>
          </Tooltip.Trigger>
          <Tooltip.Content>
            <p>Drag to move</p>
          </Tooltip.Content>
        </Tooltip.Root>
      {/if}

      {@render children?.()}
    </div>
  </ContextMenu.Trigger>

  <ContextMenu.Content class="w-64">
    <ContextMenu.Item onclick={handleConfigure}>
      <Settings class="mr-2 h-4 w-4" />
      Configure
    </ContextMenu.Item>
    <ContextMenu.Item onclick={handleDuplicate}>
      <Copy class="mr-2 h-4 w-4" />
      Duplicate
    </ContextMenu.Item>
    <ContextMenu.Item onclick={handleRemove} class="text-destructive focus:text-destructive">
      <Trash2 class="mr-2 h-4 w-4" />
      Remove
    </ContextMenu.Item>
  </ContextMenu.Content>
</ContextMenu.Root>

<Dialog.Root bind:open={showConfigDialog}>
  <Dialog.Content class="sm:max-w-[425px]">
    <Dialog.Header>
      <Dialog.Title>Configure {widget.title}</Dialog.Title>
      <Dialog.Description>
        Customize the settings for this widget.
      </Dialog.Description>
    </Dialog.Header>
    <div class="grid gap-4 py-4">
      <div class="text-sm text-muted-foreground">
        Widget configuration options would go here.
      </div>
    </div>
    <Dialog.Footer>
      <Button variant="outline" onclick={handleClose}>
        Cancel
      </Button>
      <Button onclick={handleClose}>
        Save changes
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style>
  .widget-container {
    min-height: 120px;
    min-width: 200px;
  }

  .widget-container:hover {
    z-index: 5;
  }
</style>
