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
Enhanced Dashboard Test Page
Test page for the enhanced dashboard components with shadcn/ui styling
-->

<script lang="ts">
  import EnhancedDashboardGrid from '$lib/components/dashboard/EnhancedDashboardGrid.svelte';
  import EnhancedDashboardToolbar from '$lib/components/dashboard/EnhancedDashboardToolbar.svelte';
  import EnhancedWidgetPalette from '$lib/components/dashboard/EnhancedWidgetPalette.svelte';
  import { isEditMode } from '$lib/stores/dashboard';

  // Removed hardcoded widget initialization to allow persistent drag/drop
</script>

<svelte:head>
  <title>Enhanced Dashboard - GoSight</title>
  <meta name="description" content="Enhanced dashboard with modern shadcn/ui components" />
</svelte:head>

<div class="h-screen flex flex-col bg-background">
  <!-- Header -->
  <header class="border-b bg-card">
    <div class="container mx-auto px-4 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold">Enhanced Dashboard</h1>
          <p class="text-sm text-muted-foreground">Modern dashboard with shadcn/ui components</p>
        </div>
        <div class="flex items-center gap-4">
          {#if $isEditMode}
            <EnhancedWidgetPalette />
          {/if}
          <EnhancedDashboardToolbar />
        </div>
      </div>
    </div>
  </header>

  <!-- Dashboard Content -->
  <main class="flex-1 overflow-hidden">
    <div class="h-full p-4">
      <EnhancedDashboardGrid />
    </div>
  </main>
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }
</style>
