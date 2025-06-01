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
  import Button from '$lib/components/ui/button/button.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import * as AlertDialog from '$lib/components/ui/alert-dialog';
  import Input from '$lib/components/ui/input/input.svelte';
  import { Plus, X, Edit2 } from 'lucide-svelte';
  import { toast } from 'svelte-sonner';

  interface Props {
    children: import('svelte').Snippet<[{ dashboard: Dashboard }]>;
  }

  let { children }: Props = $props();

  // Dashboard state
  interface Dashboard {
    id: string;
    name: string;
    isActive: boolean;
  }

  let dashboards = $state<Dashboard[]>([
    { id: 'main', name: 'Main Dashboard', isActive: true },
    { id: 'performance', name: 'Performance', isActive: false },
    { id: 'security', name: 'Security', isActive: false }
  ]);

  let activeDashboardId = $state('main');
  let showCreateDialog = $state(false);
  let showDeleteDialog = $state(false);
  let deleteDashboardId = $state<string | null>(null);
  let newDashboardName = $state('');

  // Reactive derived values
  let activeDashboard = $derived(dashboards.find(d => d.id === activeDashboardId));

  function switchToDashboard(dashboardId: string) {
    activeDashboardId = dashboardId;
    // Update active state
    dashboards = dashboards.map(d => ({
      ...d,
      isActive: d.id === dashboardId
    }));
  }

  function createDashboard() {
    if (!newDashboardName.trim()) {
      toast.error('Dashboard name is required');
      return;
    }

    const id = newDashboardName.toLowerCase().replace(/\s+/g, '-');
    const newDashboard: Dashboard = {
      id,
      name: newDashboardName.trim(),
      isActive: false
    };

    dashboards = [...dashboards, newDashboard];
    switchToDashboard(id);
    
    newDashboardName = '';
    showCreateDialog = false;
    toast.success(`Dashboard "${newDashboard.name}" created`);
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

    const dashboardToDelete = dashboards.find(d => d.id === deleteDashboardId);
    if (!dashboardToDelete) return;

    // If deleting active dashboard, switch to first remaining
    if (deleteDashboardId === activeDashboardId) {
      const remaining = dashboards.filter(d => d.id !== deleteDashboardId);
      if (remaining.length > 0) {
        activeDashboardId = remaining[0].id;
      }
    }

    dashboards = dashboards.filter(d => d.id !== deleteDashboardId);
    
    toast.success(`Dashboard "${dashboardToDelete.name}" deleted`);
    showDeleteDialog = false;
    deleteDashboardId = null;
  }
</script>

<Tabs.Root bind:value={activeDashboardId} class="w-full">
  <!-- Tab List with Add Button -->
  <div class="flex items-center justify-center gap-2 border-b px-4">
    <Tabs.List class="h-10 flex-shrink-0">
      {#each dashboards as dashboard (dashboard.id)}
        <Tabs.Trigger 
          value={dashboard.id}
          class="relative group"
          onclick={() => switchToDashboard(dashboard.id)}
        >
          {dashboard.name}
          
          <!-- Delete button for non-active tabs when there's more than one -->
          {#if dashboards.length > 1}
            <span
              class="ml-2 inline-flex h-4 w-4 items-center justify-center rounded opacity-0 group-hover:opacity-100 hover:bg-destructive/10 hover:text-destructive cursor-pointer"
              onclick={(e) => {
                e.stopPropagation();
                confirmDeleteDashboard(dashboard.id);
              }}
              role="button"
              tabindex="0"
            >
              <X class="h-3 w-3" />
            </span>
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

  <!-- Tab Content -->
  {#each dashboards as dashboard (dashboard.id)}
    <Tabs.Content value={dashboard.id} class="mt-0">
      {@render children({ dashboard })}
    </Tabs.Content>
  {/each}
</Tabs.Root>

<!-- Create Dashboard Dialog -->
<Dialog.Root bind:open={showCreateDialog}>
  <Dialog.Content class="sm:max-w-md">
    <Dialog.Header>
      <Dialog.Title>Create New Dashboard</Dialog.Title>
      <Dialog.Description>
        Enter a name for your new dashboard.
      </Dialog.Description>
    </Dialog.Header>
    
    <div class="grid gap-4 py-4">
      <div class="grid gap-2">
        <label for="dashboard-name" class="text-sm font-medium">Dashboard Name</label>
        <Input
          id="dashboard-name"
          bind:value={newDashboardName}
          placeholder="e.g. Network Monitoring"
          onkeydown={(e) => {
            if (e.key === 'Enter') {
              createDashboard();
            }
          }}
        />
      </div>
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

<!-- Delete Dashboard Confirmation -->
<AlertDialog.Root bind:open={showDeleteDialog}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete Dashboard</AlertDialog.Title>
      <AlertDialog.Description>
        Are you sure you want to delete this dashboard? This action cannot be undone.
      </AlertDialog.Description>
    </AlertDialog.Header>
    
    <AlertDialog.Footer>
      <AlertDialog.Cancel onclick={() => showDeleteDialog = false}>
        Cancel
      </AlertDialog.Cancel>
      <AlertDialog.Action onclick={deleteDashboard} class="bg-destructive text-destructive-foreground hover:bg-destructive/90">
        Delete
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
