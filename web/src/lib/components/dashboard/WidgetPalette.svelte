<!-- 
Copyright (C) 2025  import { dashboardStore, activeDashboard } from '$lib/stores/dashboard';
  import type { WidgetTemplate, WidgetType } from '$lib/types/dashboard';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
  import TableIcon from '@lucide/svelte/icons/table';
  import ActivityIcon from '@lucide/svelte/icons/activity';
  import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
  import FileTextIcon from '@lucide/svelte/icons/file-text';
  import GaugeIcon from '@lucide/svelte/icons/gauge';
  import PlusIcon from '@lucide/svelte/icons/plus';
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
Widget Palette Component
Provides a sidebar or dropdown menu for adding new widgets to the dashboard.
Includes predefined widget templates with icons and descriptions.
-->

<script lang="ts">
  import { dashboardStore, activeDashboard, isEditMode } from '$lib/stores/dashboard';
  import type { WidgetTemplate, WidgetType } from '$lib/types/dashboard';
  import { cn } from '$lib/utils';
  import * as Card from '$lib/components/ui/card';
  import { Button } from '$lib/components/ui/button';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
  import TableIcon from '@lucide/svelte/icons/table';
  import ActivityIcon from '@lucide/svelte/icons/activity';
  import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
  import FileTextIcon from '@lucide/svelte/icons/file-text';
  import GaugeIcon from '@lucide/svelte/icons/gauge';
  import PlusIcon from '@lucide/svelte/icons/plus';

  export let variant: 'dropdown' | 'sidebar' = 'dropdown';

  const widgetTemplates: WidgetTemplate[] = [
    {
      type: 'metric-card',
      name: 'Metric Card',
      description: 'Display a single metric with value and description',
      defaultSize: { width: 3, height: 2 },
      icon: 'gauge'
    },
    {
      type: 'chart',
      name: 'Chart',
      description: 'Visualize data with various chart types',
      defaultSize: { width: 6, height: 4 },
      icon: 'bar-chart'
    },
    {
      type: 'table',
      name: 'Data Table',
      description: 'Display tabular data with sorting and filtering',
      defaultSize: { width: 8, height: 4 },
      icon: 'table'
    },
    {
      type: 'log-viewer',
      name: 'Log Viewer',
      description: 'View and search through log entries',
      defaultSize: { width: 12, height: 6 },
      icon: 'file-text'
    },
    {
      type: 'alert-list',
      name: 'Alert List',
      description: 'Show active alerts and notifications',
      defaultSize: { width: 4, height: 4 },
      icon: 'alert-triangle'
    },
    {
      type: 'status-indicator',
      name: 'Status Indicator',
      description: 'Show system or service status',
      defaultSize: { width: 2, height: 2 },
      icon: 'activity'
    }
  ];

  function getIcon(iconName: string) {
    switch (iconName) {
      case 'gauge': return GaugeIcon;
      case 'bar-chart': return BarChart3Icon;
      case 'table': return TableIcon;
      case 'file-text': return FileTextIcon;
      case 'alert-triangle': return AlertTriangleIcon;
      case 'activity': return ActivityIcon;
      default: return GaugeIcon;
    }
  }

  function addWidget(template: WidgetTemplate) {
    const position = dashboardStore.findEmptyPosition(
      template.defaultSize.width,
      template.defaultSize.height
    );
    
    const widgetCount = $activeDashboard?.widgets?.length || 0;
    
    dashboardStore.addWidget({
      type: template.type,
      title: `${template.name} ${widgetCount + 1}`,
      position: {
        ...position,
        width: template.defaultSize.width,
        height: template.defaultSize.height
      },
      config: {}
    });
  }
</script>

{#if variant === 'dropdown'}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button {...props} variant="default" class="gap-2">
          <PlusIcon class="h-4 w-4" />
          Add Widget
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-80" align="end">
      <DropdownMenu.Label>Choose a Widget Type</DropdownMenu.Label>
      <DropdownMenu.Separator />
      
      <div class="grid gap-1 p-1">        {#each widgetTemplates as template}
          {@const IconComponent = getIcon(template.icon)}
          <DropdownMenu.Item 
            class="p-3 cursor-pointer focus:bg-accent"
            onclick={() => addWidget(template)}
          >
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0 mt-0.5">
                <IconComponent class="h-5 w-5 text-primary" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm">{template.name}</div>
                <div class="text-xs text-muted-foreground mt-0.5 leading-relaxed">
                  {template.description}
                </div>
                <div class="text-xs text-muted-foreground mt-1">
                  Size: {template.defaultSize.width}×{template.defaultSize.height}
                </div>
              </div>
            </div>
          </DropdownMenu.Item>
        {/each}
      </div>
    </DropdownMenu.Content>
  </DropdownMenu.Root>

{:else if variant === 'sidebar'}
  <Card.Root class="w-80 h-full">
    <Card.Header>
      <Card.Title class="text-lg">Widget Palette</Card.Title>
      <Card.Description>
        Drag widgets to the dashboard or click to add
      </Card.Description>
    </Card.Header>
    
    <Card.Content class="space-y-2 overflow-y-auto">
      {#each widgetTemplates as template}
        {@const IconComponent = getIcon(template.icon)}        <Card.Root 
          class="cursor-pointer hover:shadow-md transition-all duration-200 hover:scale-[1.02]"
          onclick={() => addWidget(template)}
        >
          <Card.Content class="p-4">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <div class="p-2 rounded-lg bg-primary/10">
                  <IconComponent class="h-5 w-5 text-primary" />
                </div>
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm mb-1">{template.name}</div>
                <div class="text-xs text-muted-foreground leading-relaxed mb-2">
                  {template.description}
                </div>
                <div class="text-xs text-muted-foreground">
                  Default size: {template.defaultSize.width}×{template.defaultSize.height}
                </div>
              </div>
            </div>
          </Card.Content>
        </Card.Root>
      {/each}
    </Card.Content>
  </Card.Root>
{/if}
