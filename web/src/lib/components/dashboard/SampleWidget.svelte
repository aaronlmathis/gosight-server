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
Sample Widget Components
Provides sample implementations for different widget types to demonstrate
the dashboard system functionality.
-->

<script lang="ts">
  import type { Widget } from '$lib/types/dashboard';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
  import TrendingDownIcon from '@lucide/svelte/icons/trending-down';
  import ActivityIcon from '@lucide/svelte/icons/activity';
  import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
  import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
  import ClockIcon from '@lucide/svelte/icons/clock';

  export let widget: Widget;

  // Sample data generation
  function generateMetricValue() {
    return Math.floor(Math.random() * 1000) + 1;
  }

  function generateTrend() {
    return Math.random() > 0.5 ? 'up' : 'down';
  }

  function generatePercentage() {
    return (Math.random() * 20 - 10).toFixed(1); // -10% to +10%
  }

  $: metricValue = generateMetricValue();
  $: trend = generateTrend();
  $: percentage = generatePercentage();
</script>

{#if widget.type === 'metric-card'}
  <div class="h-full flex flex-col justify-center items-center text-center space-y-2">
    <div class="text-3xl font-bold text-primary">{metricValue}</div>
    <div class="text-sm font-medium text-foreground">
      {widget.config.metric || 'Sample Metric'}
    </div>
    <div class="flex items-center gap-1 text-xs text-muted-foreground">
      {#if trend === 'up'}
        <TrendingUpIcon class="h-3 w-3 text-green-500" />
        <span class="text-green-500">+{percentage}%</span>
      {:else}
        <TrendingDownIcon class="h-3 w-3 text-red-500" />
        <span class="text-red-500">{percentage}%</span>
      {/if}
      <span class="ml-1">vs last period</span>
    </div>
    <div class="text-xs text-muted-foreground">
      Updated: {new Date().toLocaleTimeString()}
    </div>
  </div>

{:else if widget.type === 'chart'}
  <div class="h-full flex flex-col">
    <div class="text-sm font-medium mb-3">
      {widget.config.title || 'Performance Chart'}
    </div>
    <div class="flex-1 bg-gradient-to-br from-primary/10 to-primary/5 rounded-lg border-2 border-dashed border-primary/20 flex items-center justify-center relative overflow-hidden">
      <!-- Simulated chart bars -->
      <div class="absolute bottom-0 left-0 right-0 flex items-end justify-around h-3/4 px-4">
        {#each Array(8) as _, i}
          <div 
            class="bg-primary/60 rounded-t-sm w-6"
            style="height: {Math.random() * 100}%"
          ></div>
        {/each}
      </div>
      <div class="text-xs text-muted-foreground relative z-10 bg-background/80 px-2 py-1 rounded">
        Chart visualization
      </div>
    </div>
  </div>

{:else if widget.type === 'table'}
  <div class="h-full flex flex-col">
    <div class="text-sm font-medium mb-3">
      {widget.config.title || 'Data Table'}
    </div>
    <div class="flex-1 overflow-auto">
      <div class="space-y-1">
        <!-- Header -->
        <div class="grid grid-cols-3 gap-2 text-xs font-medium text-muted-foreground pb-1 border-b">
          <div>Resource</div>
          <div>Status</div>
          <div>Usage</div>
        </div>
        <!-- Sample rows -->
        {#each ['CPU', 'Memory', 'Disk', 'Network'] as resource, i}
          <div class="grid grid-cols-3 gap-2 text-xs py-1">
            <div class="font-medium">{resource}</div>
            <div>
              {#if i % 3 === 0}
                <Badge variant="default" class="h-4 text-xs">Active</Badge>
              {:else if i % 3 === 1}
                <Badge variant="secondary" class="h-4 text-xs">Warning</Badge>
              {:else}
                <Badge variant="destructive" class="h-4 text-xs">Error</Badge>
              {/if}
            </div>
            <div class="text-muted-foreground">{Math.floor(Math.random() * 100)}%</div>
          </div>
        {/each}
      </div>
    </div>
  </div>

{:else if widget.type === 'log-viewer'}
  <div class="h-full flex flex-col">
    <div class="text-sm font-medium mb-3">
      {widget.config.title || 'Recent Logs'}
    </div>
    <div class="flex-1 overflow-auto bg-muted/30 rounded border p-2 font-mono text-xs space-y-1">
      {#each ['INFO', 'WARN', 'ERROR', 'INFO', 'DEBUG'] as level, i}
        <div class="flex gap-2">
          <span class="text-muted-foreground">{new Date().toTimeString().split(' ')[0]}</span>
          <span class={cn(
            "font-medium",
            level === 'ERROR' && "text-red-500",
            level === 'WARN' && "text-yellow-500",
            level === 'INFO' && "text-blue-500",
            level === 'DEBUG' && "text-gray-500"
          )}>[{level}]</span>
          <span class="text-foreground">Sample log message {i + 1}</span>
        </div>
      {/each}
    </div>
  </div>

{:else if widget.type === 'alert-list'}
  <div class="h-full flex flex-col">
    <div class="text-sm font-medium mb-3">
      {widget.config.title || 'Active Alerts'}
    </div>
    <div class="flex-1 overflow-auto space-y-2">
      {#each [
        { severity: 'critical', message: 'High CPU usage detected' },
        { severity: 'warning', message: 'Disk space running low' },
        { severity: 'info', message: 'Scheduled maintenance' }
      ] as alert}
        <div class="flex items-start gap-2 p-2 rounded-lg bg-muted/50">
          {#if alert.severity === 'critical'}
            <AlertTriangleIcon class="h-4 w-4 text-red-500 mt-0.5 flex-shrink-0" />
          {:else if alert.severity === 'warning'}
            <AlertTriangleIcon class="h-4 w-4 text-yellow-500 mt-0.5 flex-shrink-0" />
          {:else}
            <CheckCircleIcon class="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
          {/if}
          <div class="flex-1 min-w-0">
            <div class="text-xs font-medium">{alert.message}</div>
            <div class="text-xs text-muted-foreground flex items-center gap-1 mt-1">
              <ClockIcon class="h-3 w-3" />
              {new Date().toLocaleTimeString()}
            </div>
          </div>
        </div>
      {/each}
    </div>
  </div>

{:else if widget.type === 'status-indicator'}
  <div class="h-full flex flex-col justify-center items-center text-center space-y-3">
    <div class="relative">
      <ActivityIcon class="h-8 w-8 text-green-500" />
      <div class="absolute -top-1 -right-1 w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
    </div>
    <div class="text-sm font-medium">System Online</div>
    <div class="text-xs text-muted-foreground">
      {widget.config.service || 'All services operational'}
    </div>
    <Badge variant="default" class="bg-green-500 hover:bg-green-500">
      Healthy
    </Badge>
  </div>

{:else}
  <!-- Default widget content -->
  <div class="h-full flex flex-col justify-center items-center text-center space-y-2">
    <div class="text-lg font-semibold text-muted-foreground">{widget.type}</div>
    <div class="text-sm text-muted-foreground">Widget content</div>
    <div class="text-xs text-muted-foreground">
      Configure this widget to customize its display
    </div>
  </div>
{/if}
