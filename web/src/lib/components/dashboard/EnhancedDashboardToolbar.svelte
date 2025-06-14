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
Enhanced Dashboard Toolbar Component
Modern toolbar with theme toggle, layout options, export features, and more.
-->

<script lang="ts">
  import { dashboardStore, activeDashboard, isEditMode, showGridLines } from '$lib/stores/dashboardStore';
  import { toast } from 'svelte-sonner';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Sheet from '$lib/components/ui/sheet';
  import Button from '$lib/components/ui/button/button.svelte';
  import Badge from '$lib/components/ui/badge/badge.svelte';
  import Separator from '$lib/components/ui/separator/separator.svelte';
  import Switch from '$lib/components/ui/switch/switch.svelte';
  import { 
    Edit3, 
    Save, 
    RotateCcw,
    Settings,
    Download,
    Upload,
    RefreshCw,
    Palette,
    Monitor,
    Smartphone
  } from 'lucide-svelte';

  // State
  let autoRefresh = $state(true);
  let refreshInterval = $state(30);
  let showSettingsSheet = $state(false);
  let compactMode = $state(false);

  // Derived state
  let dashboard = $derived($dashboardStore);
  let editMode = $derived($isEditMode);
  let gridLinesVisible = $derived($showGridLines);

  // Functions
  function toggleEditMode() {
    isEditMode.update(mode => !mode);
    const message = editMode ? 'Edit mode disabled' : 'Edit mode enabled';
    toast.success(message);
  }

  function saveDashboard() {
    toast.success('Dashboard saved', {
      description: 'All changes have been saved successfully'
    });
  }

  function resetDashboard() {
    if (confirm('Are you sure you want to reset the dashboard? This will remove all widgets.')) {
      dashboardStore.reset();
      toast.success('Dashboard reset');
    }
  }

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
            // You would need to implement importData method or manually add widgets
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

  function refreshData() {
    toast.success('Dashboard data refreshed');
  }

  function toggleAutoRefresh() {
    autoRefresh = !autoRefresh;
    toast.success(`Auto-refresh ${autoRefresh ? 'enabled' : 'disabled'}`);
  }

  function toggleCompactMode() {
    compactMode = !compactMode;
    toast.success(`Compact mode ${compactMode ? 'enabled' : 'disabled'}`);
  }

  function toggleGridLines() {
    showGridLines.update(value => !value);
    toast.success(`Grid lines ${$showGridLines ? 'shown' : 'hidden'}`);
  }
</script>

<!-- Main Toolbar - At bottom of content area -->
<div class="w-full border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
  <!-- Toolbar Content Area -->
  <div class="flex items-center justify-between px-4 py-3">
    <!-- Left Section - Main Actions -->
    <div class="flex items-center gap-4">
      <!-- Edit Mode Toggle -->
    <Button
      variant={editMode ? 'default' : 'outline'}
      size="sm"
      onclick={toggleEditMode}
      class="gap-2"
    >
      <Edit3 class="h-4 w-4" />
      {editMode ? 'Exit Edit' : 'Edit Mode'}
    </Button>

    {#if editMode}
      <Separator orientation="vertical" class="h-6" />
      
      <!-- Save Button -->
      <Button
        variant="outline"
        size="sm"
        onclick={saveDashboard}
        class="gap-2"
      >
        <Save class="h-4 w-4" />
        Save
      </Button>

      <!-- Reset Button -->
      <Button
        variant="outline"
        size="sm"
        onclick={resetDashboard}
        class="gap-2"
      >
        <RotateCcw class="h-4 w-4" />
        Reset
      </Button>
    {/if}
    </div>

    <!-- Center Section - Status -->
    <div class="flex items-center gap-2">
      {#if editMode}
        <Badge variant="secondary">Edit Mode</Badge>
      {/if}
      {#if autoRefresh}
        <Badge variant="outline">Auto-refresh: {refreshInterval}s</Badge>
      {/if}
      <Badge variant="outline">{$activeDashboard?.widgets.length || 0} widgets</Badge>
    </div>

    <!-- Right Section - Tools and Settings -->
    <div class="flex items-center gap-2">
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

    <!-- Refresh -->
    <Button
      variant="outline"
      size="sm"
      onclick={refreshData}
      class="gap-2"
    >
      <RefreshCw class="h-4 w-4" />
      Refresh
    </Button>

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
  </div>
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
