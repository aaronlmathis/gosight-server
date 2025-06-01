<!-- 
Copyright (C) 2025 Aaron Mathis
This file is part of GoSight Server.
-->

<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import * as Drawer from '$lib/components/ui/drawer';
  import * as Tabs from '$lib/components/ui/tabs';
  import Button from '$lib/components/ui/button/button.svelte';
  import Input from '$lib/components/ui/input/input.svelte';
  import Badge from '$lib/components/ui/badge/badge.svelte';
  import { Plus, Search, TrendingUp, BarChart3, PieChart, Gauge, Cpu, Database, Activity, Clock } from 'lucide-svelte';
  import type { WidgetType } from '$lib/types/dashboard';
  import { dashboardStore, draggedWidget } from '$lib/stores/dashboard';

  const dispatch = createEventDispatcher();

  let drawerOpen = $state(false);
  let searchQuery = $state('');

  // Check if we're on mobile (simplified approach)
  let isMobile = $state(false);
  
  if (typeof window !== 'undefined') {
    isMobile = window.innerWidth < 768;
    window.addEventListener('resize', () => {
      isMobile = window.innerWidth < 768;
    });
  }

  interface WidgetItem {
    type: WidgetType;
    title: string;
    description: string;
    icon: any;
    preview: string;
    tags: string[];
  }

  interface WidgetCategory {
    label: string;
    icon: any;
    widgets: WidgetItem[];
  }

  const widgetCategories: Record<string, WidgetCategory> = {
    metrics: {
      label: 'Metrics',
      icon: Gauge,
      widgets: [
        { type: 'metric' as WidgetType, title: 'Metric Display', description: 'Single value with trend', icon: TrendingUp, preview: '42', tags: ['kpi', 'number'] },
        { type: 'gauge' as WidgetType, title: 'Gauge Chart', description: 'Circular progress indicator', icon: Gauge, preview: '75%', tags: ['progress', 'percentage'] }
      ]
    },
    charts: {
      label: 'Charts',
      icon: BarChart3,
      widgets: [
        { type: 'chart' as WidgetType, title: 'Line Chart', description: 'Time series data visualization', icon: TrendingUp, preview: '📈', tags: ['trend', 'time'] },
        { type: 'bar' as WidgetType, title: 'Bar Chart', description: 'Compare values across categories', icon: BarChart3, preview: '📊', tags: ['comparison', 'category'] },
        { type: 'pie' as WidgetType, title: 'Pie Chart', description: 'Show proportions', icon: PieChart, preview: '🥧', tags: ['proportion', 'percentage'] }
      ]
    },
    data: {
      label: 'Data',
      icon: Database,
      widgets: [
        { type: 'table' as WidgetType, title: 'Data Table', description: 'Tabular data with sorting', icon: Database, preview: '📋', tags: ['table', 'list'] },
        { type: 'list' as WidgetType, title: 'Activity List', description: 'Recent events and activities', icon: Activity, preview: '📝', tags: ['events', 'log'] }
      ]
    },
    monitoring: {
      label: 'Monitoring',
      icon: Activity,
      widgets: [
        { type: 'status' as WidgetType, title: 'Status Board', description: 'Service health overview', icon: Activity, preview: '🟢', tags: ['health', 'uptime'] },
        { type: 'alerts' as WidgetType, title: 'Alert List', description: 'Current alerts and warnings', icon: Activity, preview: '🚨', tags: ['alerts', 'warnings'] }
      ]
    },
    system: {
      label: 'System',
      icon: Cpu,
      widgets: [
        { type: 'cpu' as WidgetType, title: 'CPU Usage', description: 'Processor utilization', icon: Cpu, preview: '32%', tags: ['cpu', 'performance'] },
        { type: 'memory' as WidgetType, title: 'Memory Usage', description: 'RAM consumption', icon: Cpu, preview: '4.2GB', tags: ['memory', 'ram'] }
      ]
    }
  };

  let filteredCategories = $derived(
    Object.entries(widgetCategories).reduce((acc, [key, category]) => {
      const filteredWidgets = category.widgets.filter(widget => 
        searchQuery === '' || 
        widget.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        widget.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        widget.tags.some((tag: string) => tag.toLowerCase().includes(searchQuery.toLowerCase()))
      );
      
      if (filteredWidgets.length > 0) {
        (acc as any)[key] = { ...category, widgets: filteredWidgets };
      }
      return acc;
    }, {})
  );

  function addWidget(widgetType: WidgetType, title: string) {
    const position = dashboardStore.findEmptyPosition(4, 3);
    dashboardStore.addWidget({
      type: widgetType,
      title: title,
      config: {},
      position
    });
    drawerOpen = false;
  }

  function startDrag(widgetType: WidgetType) {
    draggedWidget.set(widgetType);
  }
</script>

<!-- Desktop Dropdown Menu -->
{#if !isMobile}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      <Button variant="outline" size="sm" class="gap-2">
        <Plus class="h-4 w-4" />
        Add Widget
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-80 max-h-96 overflow-y-auto">
      <DropdownMenu.Label>Widget Categories</DropdownMenu.Label>
      <DropdownMenu.Separator />
      
      {#each Object.entries(widgetCategories) as [key, category]}
        <DropdownMenu.Sub>
          <DropdownMenu.SubTrigger>
            {@const CategoryIcon = category.icon}
            <CategoryIcon class="h-4 w-4 mr-2" />
            {category.label}
          </DropdownMenu.SubTrigger>
          <DropdownMenu.SubContent class="w-64">
            {#each category.widgets as widget}
              <DropdownMenu.Item 
                onclick={() => addWidget(widget.type, widget.title)}
                class="flex flex-col items-start p-3 cursor-pointer"
              >
                {@const WidgetIcon = widget.icon}
                <div class="flex items-center gap-2 w-full">
                  <WidgetIcon class="h-4 w-4" />
                  <span class="font-medium">{widget.title}</span>
                  <span class="text-lg ml-auto">{widget.preview}</span>
                </div>
                <p class="text-xs text-muted-foreground mt-1">{widget.description}</p>
                <div class="flex gap-1 mt-1">
                  {#each widget.tags.slice(0, 2) as tag}
                    <Badge variant="outline" class="text-xs">{tag}</Badge>
                  {/each}
                </div>
              </DropdownMenu.Item>
            {/each}
          </DropdownMenu.SubContent>
        </DropdownMenu.Sub>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
  <!-- Mobile Drawer Button -->
  <Button variant="outline" size="sm" onclick={() => drawerOpen = true} class="gap-2">
    <Plus class="h-4 w-4" />
    Add Widget
  </Button>
{/if}

<!-- Mobile Drawer -->
<Drawer.Root bind:open={drawerOpen}>
  <Drawer.Content class="max-h-[80vh]">
    <Drawer.Header>
      <Drawer.Title>Add Widget</Drawer.Title>
      <Drawer.Description>
        Choose a widget to add to your dashboard
      </Drawer.Description>
    </Drawer.Header>
    
    <div class="p-4 space-y-4 overflow-y-auto">
      <!-- Search -->
      <div class="relative">
        <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          bind:value={searchQuery}
          placeholder="Search widgets..."
          class="pl-10"
        />
      </div>

      <!-- Categories -->
      <Tabs.Root value="metrics" class="w-full">
        <Tabs.List class="grid grid-cols-3 w-full">
          {#each Object.entries(filteredCategories).slice(0, 3) as [key, category]}
            {@const CategoryIcon = (category as WidgetCategory).icon}
            <Tabs.Trigger value={key} class="text-xs">
              <CategoryIcon class="h-3 w-3 mr-1" />
              {(category as WidgetCategory).label}
            </Tabs.Trigger>
          {/each}
        </Tabs.List>
        
        {#each Object.entries(filteredCategories) as [key, category]}
          <Tabs.Content value={key} class="space-y-2 mt-4">
            {#each (category as WidgetCategory).widgets as widget}
              {@const WidgetIcon = widget.icon}
              <div
                class="cursor-pointer hover:bg-accent transition-colors p-3 rounded-lg border bg-card"
                onclick={() => addWidget(widget.type, widget.title)}
                onkeydown={(e) => e.key === 'Enter' && addWidget(widget.type, widget.title)}
                role="button"
                tabindex="0"
                draggable="true"
                ondragstart={() => startDrag(widget.type)}
              >
                <div class="flex items-center gap-3">
                  <WidgetIcon class="h-5 w-5" />
                  <div class="flex-1">
                    <h4 class="font-medium">{widget.title}</h4>
                    <p class="text-sm text-muted-foreground">{widget.description}</p>
                    <div class="flex gap-1 mt-1">
                      {#each widget.tags as tag}
                        <Badge variant="outline" class="text-xs">{tag}</Badge>
                      {/each}
                    </div>
                  </div>
                  <div class="text-2xl">{widget.preview}</div>
                </div>
              </div>
            {/each}
          </Tabs.Content>
        {/each}
      </Tabs.Root>
    </div>
  </Drawer.Content>
</Drawer.Root>
