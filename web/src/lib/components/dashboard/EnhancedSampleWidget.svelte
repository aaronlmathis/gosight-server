<!-- 
Copyright (C) 2025 Aaron Mathis
This file is part of GoSight Server.
-->

<script lang="ts">
  import * as Card from '$lib/components/ui/card';
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as HoverCard from '$lib/components/ui/hover-card';
  import * as Table from '$lib/components/ui/table';
  import Button from '$lib/components/ui/button/button.svelte';
  import Badge from '$lib/components/ui/badge/badge.svelte';
  import Progress from '$lib/components/ui/progress/progress.svelte';
  import Skeleton from '$lib/components/ui/skeleton/skeleton.svelte';
  import { 
    TrendingUp, 
    TrendingDown, 
    Activity, 
    ChevronDown, 
    AlertTriangle, 
    CheckCircle,
    XCircle,
    Clock,
    User,
    ArrowUpRight
  } from 'lucide-svelte';
  import type { Widget } from '$lib/types/dashboard';

  interface Props {
    widget: Widget;
  }

  let { widget }: Props = $props();

  // Define types for different data structures
  type MetricData = {
    value: string;
    change: string;
    trend: 'up';
    label: string;
  };

  type GaugeData = {
    value: number;
    max: number;
    label: string;
  };

  type ChartData = {
    data: { name: string; value: number; }[];
  };

  type TableData = {
    columns: string[];
    rows: string[][];
  };

  type ListData = {
    items: { id: number; title: string; time: string; type: string; }[];
  };

  type AlertsData = {
    items: { id: number; title: string; severity: string; time: string; }[];
  };

  type StatusData = {
    services: { name: string; status: string; uptime: string; }[];
  };

  type WidgetDataType = MetricData | GaugeData | ChartData | TableData | ListData | AlertsData | StatusData;

  // Sample data for different widget types
  const sampleData: Record<string, WidgetDataType> = {
    metric: {
      value: '42.3K',
      change: '+12.5%',
      trend: 'up' as const,
      label: 'Total Users'
    },
    'metric-card': {
      value: '42.3K',
      change: '+12.5%',
      trend: 'up' as const,
      label: 'Total Users'
    },
    gauge: {
      value: 73,
      max: 100,
      label: 'CPU Usage'
    },
    chart: {
      data: [
        { name: 'Mon', value: 400 },
        { name: 'Tue', value: 300 },
        { name: 'Wed', value: 500 },
        { name: 'Thu', value: 200 },
        { name: 'Fri', value: 600 }
      ]
    },
    table: {
      columns: ['Name', 'Status', 'Last Seen'],
      rows: [
        ['Server 1', 'Online', '2 min ago'],
        ['Server 2', 'Offline', '1 hour ago'],
        ['Server 3', 'Online', 'Just now']
      ]
    },
    list: {
      items: [
        { id: 1, title: 'User login', time: '2 min ago', type: 'info' },
        { id: 2, title: 'Database backup completed', time: '5 min ago', type: 'success' },
        { id: 3, title: 'High memory usage detected', time: '10 min ago', type: 'warning' }
      ]
    },
    alerts: {
      items: [
        { id: 1, title: 'High CPU usage on server-1', severity: 'warning', time: '5 min ago' },
        { id: 2, title: 'Database connection lost', severity: 'error', time: '10 min ago' },
        { id: 3, title: 'Backup completed successfully', severity: 'success', time: '1 hour ago' }
      ]
    },
    status: {
      services: [
        { name: 'API Gateway', status: 'healthy', uptime: '99.9%' },
        { name: 'Database', status: 'healthy', uptime: '99.8%' },
        { name: 'Cache Server', status: 'degraded', uptime: '95.2%' }
      ]
    },
    
    // Core Observability
    'service-map': {
      services: [
        { name: 'API Gateway', status: 'healthy', uptime: '99.9%' },
        { name: 'User Service', status: 'healthy', uptime: '99.7%' },
        { name: 'Payment Service', status: 'degraded', uptime: '97.2%' }
      ]
    },
    'sla-overview': {
      value: '99.9%',
      change: '+0.1%',
      trend: 'up' as const,
      label: 'SLA Uptime'
    },
    'health-check': {
      services: [
        { name: 'Frontend', status: 'healthy', uptime: '100%' },
        { name: 'API', status: 'healthy', uptime: '99.9%' },
        { name: 'Database', status: 'healthy', uptime: '99.8%' }
      ]
    },
    'distributed-trace': {
      data: [
        { name: 'Request', value: 45 },
        { name: 'Auth', value: 12 },
        { name: 'DB Query', value: 89 },
        { name: 'Response', value: 23 }
      ]
    },

    // Metrics & KPIs
    'golden-signals': {
      value: '125ms',
      change: '-5ms',
      trend: 'down' as const,
      label: 'P95 Latency'
    },
    'apdex-score': {
      value: 94,
      max: 100,
      label: 'Apdex Score'
    },
    'error-rate': {
      value: '0.1%',
      change: '-0.05%',
      trend: 'down' as const,
      label: 'Error Rate'
    },

    // Infrastructure
    'cpu': {
      value: 32,
      max: 100,
      label: 'CPU Usage'
    },
    'memory': {
      value: '4.2GB',
      change: '+0.3GB',
      trend: 'up' as const,
      label: 'Memory Usage'
    },
    'network-io': {
      value: '150Mbps',
      change: '+25Mbps',
      trend: 'up' as const,
      label: 'Network I/O'
    },
    'disk-io': {
      value: '2.1K IOPS',
      change: '+200',
      trend: 'up' as const,
      label: 'Disk I/O'
    },
    'node-health': {
      value: '8/10',
      change: '+1',
      trend: 'up' as const,
      label: 'Healthy Nodes'
    },

    // Application Performance
    'response-time': {
      value: '125ms',
      change: '-15ms',
      trend: 'down' as const,
      label: 'Response Time'
    },
    'throughput': {
      value: '1.2K/s',
      change: '+200/s',
      trend: 'up' as const,
      label: 'Throughput'
    },
    'database-metrics': {
      value: '45ms',
      change: '-5ms',
      trend: 'down' as const,
      label: 'Query Time'
    },
    'cache-metrics': {
      value: '89%',
      change: '+2%',
      trend: 'up' as const,
      label: 'Cache Hit Rate'
    },
    'queue-metrics': {
      value: '47',
      change: '-12',
      trend: 'down' as const,
      label: 'Queue Depth'
    },

    // Alerting & Incidents
    'alert-feed': {
      items: [
        { id: 1, title: 'High CPU usage on server-1', severity: 'warning', time: '5 min ago' },
        { id: 2, title: 'Database connection lost', severity: 'error', time: '10 min ago' },
        { id: 3, title: 'Memory threshold exceeded', severity: 'warning', time: '15 min ago' }
      ]
    },
    'incident-timeline': {
      items: [
        { id: 1, title: 'Database outage resolved', time: '2 hours ago', type: 'success' },
        { id: 2, title: 'API rate limiting activated', time: '4 hours ago', type: 'warning' },
        { id: 3, title: 'Scheduled maintenance completed', time: '6 hours ago', type: 'info' }
      ]
    },
    'alert-heatmap': {
      data: [
        { name: '00:00', value: 2 },
        { name: '06:00', value: 1 },
        { name: '12:00', value: 8 },
        { name: '18:00', value: 5 }
      ]
    },
    'oncall-status': {
      value: 'John Doe',
      change: 'Next: Jane Smith',
      trend: 'up' as const,
      label: 'On-Call Engineer'
    },
    'mttr-trends': {
      value: '12m',
      change: '-3m',
      trend: 'down' as const,
      label: 'Mean Time to Resolution'
    },
    'alert-list': {
      items: [
        { id: 1, title: 'CPU spike detected', severity: 'warning', time: '2 min ago' },
        { id: 2, title: 'Disk space low', severity: 'error', time: '5 min ago' }
      ]
    },

    // Logs & Events
    'log-stream': {
      items: [
        { id: 1, title: '[INFO] User authentication successful', time: '2 sec ago', type: 'info' },
        { id: 2, title: '[WARN] High memory usage detected', time: '5 sec ago', type: 'warning' },
        { id: 3, title: '[ERROR] Database connection timeout', time: '10 sec ago', type: 'error' }
      ]
    },
    'log-analytics': {
      data: [
        { name: 'INFO', value: 1240 },
        { name: 'WARN', value: 89 },
        { name: 'ERROR', value: 23 },
        { name: 'DEBUG', value: 456 }
      ]
    },
    'error-tracking': {
      items: [
        { id: 1, title: 'NullPointerException in UserService', severity: 'error', time: '1 min ago' },
        { id: 2, title: 'Database timeout in OrderProcessor', severity: 'error', time: '3 min ago' }
      ]
    },
    'audit-trail': {
      items: [
        { id: 1, title: 'Admin user created new account', time: '5 min ago', type: 'info' },
        { id: 2, title: 'Failed login attempt detected', time: '8 min ago', type: 'warning' }
      ]
    },
    'log-correlation': {
      data: [
        { name: 'Request ID: 123', value: 45 },
        { name: 'Trace ID: 456', value: 89 },
        { name: 'Span ID: 789', value: 12 }
      ]
    },
    'log-viewer': {
      items: [
        { id: 1, title: '2025-01-15 10:30:45 [INFO] Application started', time: '1 min ago', type: 'info' },
        { id: 2, title: '2025-01-15 10:31:02 [WARN] Cache miss for key: user:123', time: '2 min ago', type: 'warning' }
      ]
    },

    // Security & Compliance
    'security-events': {
      items: [
        { id: 1, title: 'Failed login attempt from 192.168.1.100', severity: 'warning', time: '3 min ago' },
        { id: 2, title: 'Suspicious file upload detected', severity: 'error', time: '10 min ago' }
      ]
    },
    'vulnerability-scan': {
      value: '3 High',
      change: '-1',
      trend: 'down' as const,
      label: 'Vulnerabilities'
    },
    'compliance-status': {
      value: '98%',
      change: '+1%',
      trend: 'up' as const,
      label: 'Compliance Score'
    },
    'access-patterns': {
      data: [
        { name: 'Normal', value: 890 },
        { name: 'Suspicious', value: 23 },
        { name: 'Blocked', value: 7 }
      ]
    },

    // Business Intelligence
    'user-analytics': {
      value: '12.4K',
      change: '+1.2K',
      trend: 'up' as const,
      label: 'Active Users'
    },
    'conversion-funnel': {
      data: [
        { name: 'Visitors', value: 10000 },
        { name: 'Sign-ups', value: 2500 },
        { name: 'Trials', value: 800 },
        { name: 'Purchases', value: 240 }
      ]
    },
    'revenue-metrics': {
      value: '$142K',
      change: '+$12K',
      trend: 'up' as const,
      label: 'Monthly Revenue'
    },
    'feature-adoption': {
      value: 67,
      max: 100,
      label: 'Feature Adoption'
    },
    'cost-analytics': {
      value: '$1.2K',
      change: '-$200',
      trend: 'down' as const,
      label: 'Infrastructure Costs'
    },

    // Charts & Visualization
    bar: {
      data: [
        { name: 'Jan', value: 400 },
        { name: 'Feb', value: 300 },
        { name: 'Mar', value: 600 },
        { name: 'Apr', value: 800 }
      ]
    },
    pie: {
      data: [
        { name: 'Desktop', value: 60 },
        { name: 'Mobile', value: 35 },
        { name: 'Tablet', value: 5 }
      ]
    },

    // Legacy/Compatibility
    'status-indicator': {
      services: [
        { name: 'All Systems', status: 'healthy', uptime: '99.9%' },
        { name: 'API', status: 'healthy', uptime: '99.8%' },
        { name: 'Database', status: 'healthy', uptime: '99.7%' }
      ]
    }
  };

  let data = $derived(sampleData[widget.type as keyof typeof sampleData] || sampleData.metric);
  
  let isCollapsed = $state(false);
  let isExpanded = $derived(!isCollapsed);
  let isLoading = $state(false);

  // Type guards
  function isMetricData(data: WidgetDataType): data is MetricData {
    return 'trend' in data && 'change' in data;
  }

  function isGaugeData(data: WidgetDataType): data is GaugeData {
    return 'max' in data && typeof (data as any).value === 'number';
  }

  function isChartData(data: WidgetDataType): data is ChartData {
    return 'data' in data;
  }

  function isTableData(data: WidgetDataType): data is TableData {
    return 'columns' in data && 'rows' in data;
  }

  function isListData(data: WidgetDataType): data is ListData {
    return 'items' in data && Array.isArray((data as any).items) && (data as any).items.length > 0 && 'type' in (data as any).items[0];
  }

  function isAlertsData(data: WidgetDataType): data is AlertsData {
    return 'items' in data && Array.isArray((data as any).items) && (data as any).items.length > 0 && 'severity' in (data as any).items[0];
  }

  function isStatusData(data: WidgetDataType): data is StatusData {
    return 'services' in data;
  }

  function hasLabel(data: WidgetDataType): boolean {
    return 'label' in data;
  }

  // Simulate loading for demonstration
  function refresh() {
    isLoading = true;
    setTimeout(() => {
      isLoading = false;
    }, 1000);
  }

  function getStatusIcon(status: string) {
    switch (status) {
      case 'healthy':
      case 'Online':
        return CheckCircle;
      case 'degraded':
      case 'Offline':
        return XCircle;
      default:
        return AlertTriangle;
    }
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'healthy':
      case 'Online':
      case 'success':
        return 'text-green-500';
      case 'degraded':
      case 'warning':
        return 'text-yellow-500';
      case 'error':
      case 'Offline':
        return 'text-red-500';
      default:
        return 'text-gray-500';
    }
  }
</script>

{#if isLoading}
  <div class="space-y-3 p-4">
    <Skeleton class="h-4 w-3/4" />
    <Skeleton class="h-8 w-1/2" />
    <Skeleton class="h-4 w-full" />
  </div>
{:else}
  <div class="p-4 h-full">
    {#if isMetricData(data)}
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium text-muted-foreground">{data.label}</h3>
          {#if data.trend === 'up'}
            <TrendingUp class="h-4 w-4 text-green-500" />
          {:else}
            <TrendingDown class="h-4 w-4 text-red-500" />
          {/if}
        </div>
        <div class="text-2xl font-bold">{data.value}</div>
        <div class="flex items-center text-sm">
          <span class="text-green-500">{data.change}</span>
          <span class="text-muted-foreground ml-1">from last period</span>
        </div>
      </div>

    {:else if isGaugeData(data)}
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-medium">{data.label}</h3>
          <span class="text-2xl font-bold">{data.value}%</span>
        </div>
        <Progress value={data.value} class="h-2" />
        <div class="text-xs text-muted-foreground">
          {data.value}% of {data.max}% capacity
        </div>
      </div>

    {:else if isTableData(data)}
      <div class="space-y-3">
        <h3 class="text-sm font-medium">Server Status</h3>
        <Table.Root>
          <Table.Header>
            <Table.Row>
              {#each data.columns as column}
                <Table.Head class="text-xs">{column}</Table.Head>
              {/each}
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each data.rows as row}
              <Table.Row>
                {#each row as cell, i}
                  <Table.Cell class="text-xs">
                    {#if i === 1}
                      {#if cell}
                        {@const StatusIcon = getStatusIcon(cell)}
                        <div class="flex items-center gap-1">
                          <StatusIcon class="h-3 w-3 {getStatusColor(cell)}" />
                          <span>{cell}</span>
                        </div>
                      {/if}
                    {:else}
                      {cell}
                    {/if}
                  </Table.Cell>
                {/each}
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      </div>

    {:else if isListData(data)}
      <div class="space-y-3">
        <h3 class="text-sm font-medium">Recent Activity</h3>
        <div class="space-y-2">
          {#each data.items as item}
            <div class="flex items-start gap-2 p-2 rounded-lg hover:bg-accent transition-colors">
              <Activity class="h-4 w-4 mt-0.5 {getStatusColor(item.type)}" />
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{item.title}</p>
                <p class="text-xs text-muted-foreground">{item.time}</p>
              </div>
            </div>
          {/each}
        </div>
      </div>

    {:else if isAlertsData(data)}
      <div class="space-y-3">
        <Collapsible.Root bind:open={isExpanded}>
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-medium">Active Alerts</h3>
            <Collapsible.Trigger>
              <Button variant="ghost" size="sm" class="h-6 w-6 p-0">
                <ChevronDown class="h-4 w-4 transition-transform {isCollapsed ? '' : 'rotate-180'}" />
              </Button>
            </Collapsible.Trigger>
          </div>
          
          <Collapsible.Content>
            <div class="space-y-2 mt-2">
              {#each data.items as alert}
                <HoverCard.Root>
                  <HoverCard.Trigger>
                    <div class="flex items-center justify-between p-2 rounded border cursor-pointer">
                      <div class="flex items-center gap-2">
                        <AlertTriangle class="h-4 w-4 {getStatusColor(alert.severity)}" />
                        <span class="text-sm truncate">{alert.title}</span>
                      </div>
                      <Badge variant="outline" class="text-xs">
                        {alert.severity}
                      </Badge>
                    </div>
                  </HoverCard.Trigger>
                  <HoverCard.Content class="w-80">
                    <div class="space-y-2">
                      <h4 class="text-sm font-semibold">{alert.title}</h4>
                      <p class="text-sm text-muted-foreground">
                        Alert triggered {alert.time}
                      </p>
                      <div class="flex items-center gap-2">
                        <Badge variant="outline">{alert.severity}</Badge>
                        <span class="text-xs text-muted-foreground">View details</span>
                        <ArrowUpRight class="h-3 w-3" />
                      </div>
                    </div>
                  </HoverCard.Content>
                </HoverCard.Root>
              {/each}
            </div>
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

    {:else if isStatusData(data)}
      <div class="space-y-3">
        <h3 class="text-sm font-medium">Service Health</h3>
        <div class="space-y-2">
          {#each data.services as service}
            {@const StatusIcon = getStatusIcon(service.status)}
            <div class="flex items-center justify-between p-2 rounded-lg border">
              <div class="flex items-center gap-2">
                <StatusIcon class="h-4 w-4 {getStatusColor(service.status)}" />
                <span class="text-sm">{service.name}</span>
              </div>
              <div class="text-right">
                <div class="text-xs text-muted-foreground">{service.uptime}</div>
                <div class="text-xs font-medium capitalize">{service.status}</div>
              </div>
            </div>
          {/each}
        </div>
      </div>

    {:else}
      <!-- Default widget content -->
      <div class="flex items-center justify-center h-full text-muted-foreground">
        <div class="text-center">
          <Activity class="h-8 w-8 mx-auto mb-2 opacity-50" />
          <p class="text-sm">Widget content for {widget.type}</p>
        </div>
      </div>
    {/if}
  </div>
{/if}
