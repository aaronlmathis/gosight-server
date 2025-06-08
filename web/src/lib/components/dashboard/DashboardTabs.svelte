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
Dashboard Tabs Component
Manages multiple dashboards with tab interface
-->

<script lang="ts">
  import * as Tabs from '$lib/components/ui/tabs';
  import * as Dialog from '$lib/components/ui/dialog';
  import Button from '$lib/components/ui/button/button.svelte';
  import Input from '$lib/components/ui/input/input.svelte';
  import { Plus, X } from 'lucide-svelte';
  import { dashboardStore } from '$lib/stores/dashboardStore';
  import { toast } from 'svelte-sonner';
  import type { Snippet } from 'svelte';

  interface Props {
    children: Snippet<[{ dashboard: any }]>;
  }

  let { children }: Props = $props();

  let showCreateDialog = $state(false);
  let showDeleteDialog = $state(false);
  let deleteDashboardId = $state<string | null>(null);
  let newDashboardName = $state('');

  // Get reactive dashboard data
  let dashboards = $derived($dashboardStore.dashboards);
  let activeDashboardId = $derived($dashboardStore.activeDashboardId);

  function createDashboard() {
    if (!newDashboardName.trim()) {
      toast.error('Dashboard name is required');
      return;
    }

    dashboardStore.createDashboard(newDashboardName.trim());
    newDashboardName = '';
    showCreateDialog = false;
  }

  function confirmDeleteDashboard(dashboardId: string) {
    if (dashboards.length <= 1) {
      toast.error('Cannot delete the last dashboard');
      return;
    }
    
    deleteDashboardId = dashboardId;
    showDeleteDialog = true;
  }

  function deleteDashboard() {
    if (!deleteDashboardId) return;
    
    dashboardStore.deleteDashboard(deleteDashboardId);
    showDeleteDialog = false;
    deleteDashboardId = null;
  }
</script>

<Tabs.Root value={activeDashboardId} class="w-full">
  <!-- Tab List with Add Button -->
  <div class="flex items-center justify-center gap-2 border-b px-4">
    <Tabs.List class="h-10 flex-shrink-0">
      {#each dashboards as dashboard (dashboard.id)}
        <Tabs.Trigger 
          value={dashboard.id}
          class="relative group"
          onclick={() => dashboardStore.setActiveDashboard(dashboard.id)}
        >
          {dashboard.name}
          
          <!-- Delete button for non-active tabs when there's more than one -->
          {#if dashboards.length > 1}
          <button
            type="button"
            class="ml-2 inline-flex h-4 w-4 items-center justify-center rounded opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive cursor-pointer appearance-none border-none bg-transparent p-0"
            onclick={(e) => {
              e.stopPropagation();
              confirmDeleteDashboard(dashboard.id);
            }}
            aria-label="Delete dashboard"
          >
            <X class="h-3 w-3" />
          </button>
          {/if}
        </Tabs.Trigger>
      {/each}
    </Tabs.List>
    
    <!-- Add Dashboard Button -->
    <Button
      variant="ghost"
      size="sm"
      onclick={() => showCreateDialog = true}
      class="h-8 w-8 p-0 flex-shrink-0"
    >
      <Plus class="h-4 w-4" />
    </Button>
  </div>
  <!-- Tab Content - Each dashboard gets its own content -->
  {#each dashboards as dashboard (dashboard.id)}
    <Tabs.Content value={dashboard.id} class="space-y-4">
      {@render children({ dashboard })}
    </Tabs.Content>
  {/each}
</Tabs.Root>

<!-- Create Dashboard Dialog -->
<Dialog.Root bind:open={showCreateDialog}>
  <Dialog.Content class="sm:max-w-[425px]">
    <Dialog.Header>
      <Dialog.Title>Create New Dashboard</Dialog.Title>
      <Dialog.Description>
        Enter a name for your new dashboard.
      </Dialog.Description>
    </Dialog.Header>
    <div class="grid gap-4 py-4">
      <Input
        bind:value={newDashboardName}
        placeholder="Dashboard name"
        onkeydown={(e) => e.key === 'Enter' && createDashboard()}
      />
    </div>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => showCreateDialog = false}>
        Cancel
      </Button>
      <Button onclick={createDashboard}>
        Create Dashboard
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<!-- Delete Confirmation Dialog -->
<Dialog.Root bind:open={showDeleteDialog}>
  <Dialog.Content class="sm:max-w-[425px]">
    <Dialog.Header>
      <Dialog.Title>Delete Dashboard</Dialog.Title>
      <Dialog.Description>
        Are you sure you want to delete this dashboard? This action cannot be undone.
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer>
      <Button variant="outline" onclick={() => showDeleteDialog = false}>
        Cancel
      </Button>
      <Button variant="destructive" onclick={deleteDashboard}>
        Delete Dashboard
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
