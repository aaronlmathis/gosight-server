<script lang="ts">
	import type { Widget } from '$lib/types/dashboard';

	// Import all widget components
	import MetricCard from './widgets/MetricCard.svelte';
	import ChartWidget from './widgets/ChartWidget.svelte';
	import AlertsList from './widgets/AlertsList.svelte';
	import EventsList from './widgets/EventsList.svelte';
	import QuickLinks from './widgets/QuickLinks.svelte';
	import SystemOverview from './widgets/SystemOverview.svelte';
	import EndpointCount from './widgets/EndpointCount.svelte';
	import AlertCount from './widgets/AlertCount.svelte';

	export let widget: Widget;

	// Map widget types to components
	const widgetComponents = {
		'metric-card': MetricCard,
		metric: MetricCard,
		'cpu-usage': ChartWidget,
		'memory-usage': ChartWidget,
		'disk-usage': ChartWidget,
		'network-traffic': ChartWidget,
		'response-time': ChartWidget,
		'error-rate': ChartWidget,
		throughput: ChartWidget,
		chart: ChartWidget,
		'chart-line': ChartWidget,
		'chart-donut': ChartWidget,
		'chart-bar': ChartWidget,
		'alerts-list': AlertsList,
		alerts_list: AlertsList,
		'events-list': EventsList,
		events_list: EventsList,
		'recent-events': EventsList,
		'active-alerts': AlertsList,
		'quick-links': QuickLinks,
		quick_links: QuickLinks,
		notes: QuickLinks, // Reuse for now
		'status-overview': MetricCard, // Reuse metric card
		'system-status': SystemOverview,
		'uptime-monitor': MetricCard,
		'service-health': MetricCard,
		'custom-chart': ChartWidget,
		'network-stats': ChartWidget,
		'container-stats': ChartWidget,
		'endpoint-health': MetricCard,
		'alert-summary': AlertCount,
		'log-stream': QuickLinks, // Temporary mapping
		endpoint_count: EndpointCount,
		alert_count: AlertCount,
		system_overview: SystemOverview
	};

	$: WidgetComponent = widgetComponents[widget.type] || MetricCard;
</script>

<!-- Render the appropriate widget component -->
<svelte:component this={WidgetComponent} {widget} />
