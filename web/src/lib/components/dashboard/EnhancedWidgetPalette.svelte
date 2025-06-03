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
  import { Plus, Search, TrendingUp, TrendingDown, BarChart3, PieChart, Gauge, Cpu, Database, Activity, Clock, AlertTriangle, List } from 'lucide-svelte';
  import type { WidgetType } from '$lib/types/dashboard';
  import { dashboardStore, draggedWidget } from '$lib/stores/dashboardStore';
  import { WIDGET_CONFIGS } from '$lib/configs/widget-sizing';

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
    observability: {
      label: 'Core Observability',
      icon: Activity,
      widgets: [
        { type: 'service-map' as WidgetType, title: 'Service Map', description: 'Microservices topology and dependencies', icon: Activity, preview: '🗺️', tags: ['services', 'topology', 'dependencies', 'microservices'] },
        { type: 'sla-overview' as WidgetType, title: 'SLA Dashboard', description: 'Service level objectives and error budgets', icon: TrendingUp, preview: '99.9%', tags: ['sla', 'slo', 'reliability', 'uptime'] },
        { type: 'health-check' as WidgetType, title: 'Health Check Matrix', description: 'Multi-service health status overview', icon: Activity, preview: '🟢', tags: ['health', 'status', 'services', 'monitoring'] },
        { type: 'distributed-trace' as WidgetType, title: 'Trace Timeline', description: 'Distributed tracing visualization', icon: Clock, preview: '🔗', tags: ['tracing', 'latency', 'distributed', 'performance'] }
      ]
    },
    metrics: {
      label: 'Metrics & KPIs',
      icon: Gauge,
      widgets: [
        { type: 'golden-signals' as WidgetType, title: 'Golden Signals', description: 'Latency, traffic, errors, and saturation', icon: Gauge, preview: '📊', tags: ['golden-signals', 'sre', 'latency', 'errors'] },
        { type: 'metric' as WidgetType, title: 'Business KPI', description: 'Custom business metrics and indicators', icon: TrendingUp, preview: '42K↗', tags: ['kpi', 'business', 'metrics', 'trend'] },
        { type: 'apdex-score' as WidgetType, title: 'Apdex Score', description: 'Application performance index tracking', icon: Gauge, preview: '0.94', tags: ['apdex', 'performance', 'user-satisfaction', 'sre'] },
        { type: 'error-rate' as WidgetType, title: 'Error Rate', description: 'Service error rates and trends', icon: AlertTriangle, preview: '0.1%', tags: ['errors', 'reliability', 'quality', 'monitoring'] }
      ]
    },
    infrastructure: {
      label: 'Infrastructure',
      icon: Cpu,
      widgets: [
        { type: 'cpu' as WidgetType, title: 'CPU Utilization', description: 'System and container CPU usage', icon: Cpu, preview: '32%', tags: ['cpu', 'performance', 'infrastructure', 'utilization'] },
        { type: 'memory' as WidgetType, title: 'Memory Usage', description: 'RAM and memory pool monitoring', icon: Database, preview: '4.2GB', tags: ['memory', 'ram', 'infrastructure', 'capacity'] },
        { type: 'network-io' as WidgetType, title: 'Network I/O', description: 'Network throughput and bandwidth usage', icon: Activity, preview: '150Mbps', tags: ['network', 'bandwidth', 'throughput', 'infrastructure'] },
        { type: 'disk-io' as WidgetType, title: 'Disk I/O', description: 'Storage performance and IOPS monitoring', icon: Database, preview: '2.1K IOPS', tags: ['disk', 'storage', 'iops', 'performance'] },
        { type: 'node-health' as WidgetType, title: 'Node Health', description: 'Kubernetes/cluster node status', icon: Cpu, preview: '8/10', tags: ['kubernetes', 'nodes', 'cluster', 'infrastructure'] }
      ]
    },
    applications: {
      label: 'Application Performance',
      icon: BarChart3,
      widgets: [
        { type: 'response-time' as WidgetType, title: 'Response Time', description: 'API and service response latencies', icon: Clock, preview: '125ms', tags: ['latency', 'response-time', 'api', 'performance'] },
        { type: 'throughput' as WidgetType, title: 'Throughput', description: 'Requests per second and transaction volume', icon: TrendingUp, preview: '1.2K/s', tags: ['throughput', 'rps', 'traffic', 'load'] },
        { type: 'database-metrics' as WidgetType, title: 'Database Performance', description: 'Query time, connections, and DB health', icon: Database, preview: '45ms', tags: ['database', 'queries', 'connections', 'performance'] },
        { type: 'cache-metrics' as WidgetType, title: 'Cache Performance', description: 'Hit rates, miss rates, and cache efficiency', icon: Database, preview: '89%', tags: ['cache', 'redis', 'hit-rate', 'performance'] },
        { type: 'queue-metrics' as WidgetType, title: 'Queue Metrics', description: 'Message queue depth and processing rates', icon: List, preview: '47', tags: ['queue', 'messaging', 'backlog', 'processing'] }
      ]
    },
    alerts: {
      label: 'Alerting & Incidents',
      icon: AlertTriangle,
      widgets: [
        { type: 'alert-feed' as WidgetType, title: 'Live Alert Feed', description: 'Real-time alert stream with severity filtering', icon: AlertTriangle, preview: '🚨', tags: ['alerts', 'real-time', 'incidents', 'monitoring'] },
        { type: 'incident-timeline' as WidgetType, title: 'Incident Timeline', description: 'Ongoing and recent incident tracking', icon: Clock, preview: '⏰', tags: ['incidents', 'timeline', 'mttr', 'sre'] },
        { type: 'alert-heatmap' as WidgetType, title: 'Alert Heatmap', description: 'Alert frequency patterns over time', icon: BarChart3, preview: '🔥', tags: ['alerts', 'patterns', 'frequency', 'analysis'] },
        { type: 'oncall-status' as WidgetType, title: 'On-Call Status', description: 'Current on-call engineer and escalation', icon: Activity, preview: '👤', tags: ['oncall', 'escalation', 'contact', 'incident-response'] },
        { type: 'mttr-trends' as WidgetType, title: 'MTTR Trends', description: 'Mean time to resolution analytics', icon: TrendingDown, preview: '12m', tags: ['mttr', 'resolution', 'incidents', 'trends'] }
      ]
    },
    logs: {
      label: 'Logs & Events',
      icon: List,
      widgets: [
        { type: 'log-stream' as WidgetType, title: 'Live Log Stream', description: 'Real-time log tail with filtering', icon: List, preview: '📄', tags: ['logs', 'real-time', 'streaming', 'debugging'] },
        { type: 'log-analytics' as WidgetType, title: 'Log Analytics', description: 'Log volume, patterns, and error analysis', icon: BarChart3, preview: '📊', tags: ['logs', 'analytics', 'patterns', 'errors'] },
        { type: 'error-tracking' as WidgetType, title: 'Error Tracking', description: 'Application errors and stack traces', icon: AlertTriangle, preview: '🐛', tags: ['errors', 'exceptions', 'debugging', 'stack-traces'] },
        { type: 'audit-trail' as WidgetType, title: 'Audit Trail', description: 'Security and compliance event logging', icon: List, preview: '🔍', tags: ['audit', 'security', 'compliance', 'events'] },
        { type: 'log-correlation' as WidgetType, title: 'Log Correlation', description: 'Cross-service log correlation and tracing', icon: Activity, preview: '🔗', tags: ['correlation', 'tracing', 'debugging', 'multi-service'] }
      ]
    },
    security: {
      label: 'Security & Compliance',
      icon: AlertTriangle,
      widgets: [
        { type: 'security-events' as WidgetType, title: 'Security Events', description: 'Authentication failures and security alerts', icon: AlertTriangle, preview: '🔒', tags: ['security', 'authentication', 'threats', 'compliance'] },
        { type: 'vulnerability-scan' as WidgetType, title: 'Vulnerability Status', description: 'Security scan results and CVE tracking', icon: AlertTriangle, preview: '⚠️', tags: ['vulnerabilities', 'cve', 'security', 'scanning'] },
        { type: 'compliance-status' as WidgetType, title: 'Compliance Dashboard', description: 'Regulatory compliance monitoring', icon: Activity, preview: '✅', tags: ['compliance', 'regulations', 'audit', 'governance'] },
        { type: 'access-patterns' as WidgetType, title: 'Access Patterns', description: 'User access behavior and anomalies', icon: List, preview: '👥', tags: ['access', 'behavior', 'anomalies', 'security'] }
      ]
    },
    business: {
      label: 'Business Intelligence',
      icon: TrendingUp,
      widgets: [
        { type: 'user-analytics' as WidgetType, title: 'User Analytics', description: 'Active users, sessions, and engagement', icon: TrendingUp, preview: '12.4K', tags: ['users', 'sessions', 'engagement', 'business'] },
        { type: 'conversion-funnel' as WidgetType, title: 'Conversion Funnel', description: 'User journey and conversion tracking', icon: BarChart3, preview: '📊', tags: ['conversion', 'funnel', 'journey', 'business'] },
        { type: 'revenue-metrics' as WidgetType, title: 'Revenue Metrics', description: 'Financial KPIs and revenue tracking', icon: TrendingUp, preview: '$142K', tags: ['revenue', 'financial', 'kpi', 'business'] },
        { type: 'feature-adoption' as WidgetType, title: 'Feature Adoption', description: 'Feature usage and adoption rates', icon: Gauge, preview: '67%', tags: ['features', 'adoption', 'usage', 'product'] },
        { type: 'cost-analytics' as WidgetType, title: 'Cost Analytics', description: 'Infrastructure and operational costs', icon: Database, preview: '$1.2K', tags: ['costs', 'infrastructure', 'optimization', 'finops'] }
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
    // Use the new smart sizing function that automatically calculates size based on widget type and screen size
    dashboardStore.addWidgetWithSmartSizing(widgetType, title, {});
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
      <Tabs.Root value="observability" class="w-full">
        <Tabs.List class="grid grid-cols-2 w-full mb-4">
          {#each Object.entries(filteredCategories).slice(0, 4) as [key, category]}
            {@const CategoryIcon = (category as WidgetCategory).icon}
            <Tabs.Trigger value={key} class="text-xs flex flex-col items-center gap-1 p-2">
              <CategoryIcon class="h-4 w-4" />
              <span class="truncate">{(category as WidgetCategory).label}</span>
            </Tabs.Trigger>
          {/each}
        </Tabs.List>
        
        <!-- Secondary categories row -->
        <Tabs.List class="grid grid-cols-2 w-full mb-4">
          {#each Object.entries(filteredCategories).slice(4, 8) as [key, category]}
            {@const CategoryIcon = (category as WidgetCategory).icon}
            <Tabs.Trigger value={key} class="text-xs flex flex-col items-center gap-1 p-2">
              <CategoryIcon class="h-4 w-4" />
              <span class="truncate">{(category as WidgetCategory).label}</span>
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
