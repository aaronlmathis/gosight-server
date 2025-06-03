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
Top Navbar Component
Fixed top navigation bar with logo and theme toggle
-->

<script lang="ts">
  import Button from '$lib/components/ui/button/button.svelte';
  import { Moon, Sun } from 'lucide-svelte';
  import { toast } from 'svelte-sonner';

  // State for theme
  let darkMode = $state(false);

  // Initialize theme from localStorage
  if (typeof window !== 'undefined') {
    const stored = localStorage.getItem('theme');
    darkMode = stored === 'dark' || (!stored && window.matchMedia('(prefers-color-scheme: dark)').matches);
    
    $effect(() => {
      if (darkMode) {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
    });
  }

  function toggleDarkMode() {
    darkMode = !darkMode;
    if (darkMode) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
    toast.success(`Switched to ${darkMode ? 'dark' : 'light'} mode`);
  }
</script>

<!-- Fixed Top Navbar -->
<div class="fixed top-0 left-0 right-0 z-50 flex items-center border-b bg-sidebar h-16">
  <!-- Logo Area - 16rem width, no border -->
  <div class="flex items-center px-4 py-4" style="width: 16rem;">
    <h1 class="text-4xl font-bold">GoSight</h1>
  </div>

  <!-- Spacer -->
  <div class="flex-1"></div>

  <!-- Right Section - Theme Toggle -->
  <div class="flex items-center px-4">
    <Button
      variant="ghost"
      size="sm"
      onclick={toggleDarkMode}
      class="gap-2"
    >
      {#if darkMode}
        <Sun class="h-4 w-4" />
        <span class="hidden sm:inline">Light</span>
      {:else}
        <Moon class="h-4 w-4" />
        <span class="hidden sm:inline">Dark</span>
      {/if}
    </Button>
  </div>
</div>
