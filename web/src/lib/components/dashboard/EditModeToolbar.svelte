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
Edit Mode Toolbar Component
Horizontal toolbar that appears during edit mode
-->

<script lang="ts">
  import Button from '$lib/components/ui/button/button.svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Sheet from '$lib/components/ui/sheet';
  import Switch from '$lib/components/ui/switch/switch.svelte';
  import Separator from '$lib/components/ui/separator/separator.svelte';
  import WidgetPalette from './EnhancedWidgetPalette.svelte';
  import { dashboardStore, showGridLines } from '$lib/stores/dashboardStore';
  import { toast } from 'svelte-sonner';
  import { 
    Download, 
    Upload, 
    Settings, 
    Monitor, 
    Smartphone,
    Palette
  } from 'lucide-svelte';

  // State
  let showSettingsSheet = $state(false);
  let compactMode = $state(false);
  let autoRefresh = $state(true);
  let refreshInterval = $state(30);

  // Reactive state
  let gridLinesVisible = $derived(Boolean($showGridLines));

  // Functions
  function exportDashboard() {
    const data = JSON.stringify($dashboardStore, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `dashboard-${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    
    URL.revokeObjectURL(url);
    toast.success('Dashboard exported');
  }

  function importDashboard() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';
    
    input.onchange = () => {
      const file = input.files?.[0];
      if (!file) return;
      
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const data = JSON.parse(e.target?.result as string);
          if (data.widgets && Array.isArray(data.widgets)) {
            dashboardStore.reset();
            toast.success('Dashboard imported successfully');
          } else {
            toast.error('Invalid dashboard file format');
          }
        } catch (error) {
          toast.error('Failed to import dashboard file');
        }
      };
      reader.readAsText(file);
    };
    
    input.click();
  }

  function toggleCompactMode() {
    compactMode = !compactMode;
    toast.success(`Compact mode ${compactMode ? 'enabled' : 'disabled'}`);
  }

  function toggleGridLines() {
    showGridLines.update((value: boolean) => !value);
    toast.success(`Grid lines ${$showGridLines ? 'shown' : 'hidden'}`);
  }

  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;
    toast.success(`Auto-refresh ${autoRefresh ? 'enabled' : 'disabled'}`);
  }
</script>

<!-- Edit Mode Toolbar -->
<div class="flex items-center justify-end gap-2 py-2">
  <!-- Widget Palette -->
  <WidgetPalette />

  <!-- Export/Import -->
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      <Button variant="outline" size="sm" class="gap-2">
        <Download class="h-4 w-4" />
        Export
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content>
      <DropdownMenu.Item onclick={exportDashboard}>
        <Download class="h-4 w-4 mr-2" />
        Export Configuration
      </DropdownMenu.Item>
      <DropdownMenu.Item onclick={importDashboard}>
        <Upload class="h-4 w-4 mr-2" />
        Import Configuration
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Root>

  <!-- View Options -->
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      <Button variant="outline" size="sm" class="gap-2">
        <Monitor class="h-4 w-4" />
        View
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content>
      <DropdownMenu.CheckboxItem bind:checked={compactMode} onclick={toggleCompactMode}>
        <Smartphone class="h-4 w-4 mr-2" />
        Compact Mode
      </DropdownMenu.CheckboxItem>
      <DropdownMenu.CheckboxItem bind:checked={gridLinesVisible} onclick={toggleGridLines}>
        <Palette class="h-4 w-4 mr-2" />
        Show Grid Lines
      </DropdownMenu.CheckboxItem>
    </DropdownMenu.Content>
  </DropdownMenu.Root>

  <!-- Settings -->
  <Button
    variant="outline"
    size="sm"
    onclick={() => showSettingsSheet = true}
    class="gap-2"
  >
    <Settings class="h-4 w-4" />
    Settings
  </Button>
</div>

<!-- Settings Sheet -->
<Sheet.Root bind:open={showSettingsSheet}>
  <Sheet.Content side="right" class="w-96">
    <Sheet.Header>
      <Sheet.Title>Dashboard Settings</Sheet.Title>
      <Sheet.Description>
        Configure your dashboard preferences and behavior.
      </Sheet.Description>
    </Sheet.Header>
    
    <div class="space-y-6 py-6">
      <!-- Auto Refresh Settings -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium">Auto Refresh</h4>
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Enable auto refresh</span>
          <Switch bind:checked={autoRefresh} onCheckedChange={toggleAutoRefresh} />
        </div>
        {#if autoRefresh}
          <div class="space-y-2">
            <label for="refresh-interval" class="text-sm text-muted-foreground">Refresh interval (seconds)</label>
            <input
              id="refresh-interval"
              type="number"
              bind:value={refreshInterval}
              min="5"
              max="300"
              step="5"
              class="w-full px-3 py-2 border rounded-md bg-background"
            />
          </div>
        {/if}
      </div>

      <Separator />

      <!-- Layout Settings -->
      <div class="space-y-3">
        <h4 class="text-sm font-medium">Layout</h4>
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Compact mode</span>
          <Switch bind:checked={compactMode} onCheckedChange={toggleCompactMode} />
        </div>
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Show grid lines</span>
          <Switch bind:checked={gridLinesVisible} onCheckedChange={toggleGridLines} />
        </div>
      </div>
    </div>

    <Sheet.Footer>
      <Button onclick={() => showSettingsSheet = false}>Close</Button>
    </Sheet.Footer>
  </Sheet.Content>
</Sheet.Root>
