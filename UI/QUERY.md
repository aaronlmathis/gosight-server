GoSight Metrics Querying System
Based on my analysis of the codebase, here's how metrics are queried in GoSight:

1. Data Collection Architecture
Agents collect metrics from endpoints (hosts, containers) and send them to the server
Metrics structure: namespace.subnamespace.metric_name with optional dimensions/tags
Example: system.cpu.usage_percent, container.podman.mem_usage_bytes
2. Storage Backend
VictoriaMetrics is used as the time-series database backend
Prometheus-compatible format for storage and querying
MetricCache provides in-memory indexing for fast namespace/metric discovery
3. API Endpoint Structure
The frontend queries metrics through these REST API endpoints:

4. Chart Widget Query Flow
For the dashboard chart widgets, here's the complete query flow:

Configuration Phase:
Widget Configuration Modal loads available namespaces via /api/v1/
User selects namespace → loads subnamespaces via /api/v1/{namespace}
User selects subnamespace → loads metrics via /api/v1/{namespace}/{subnamespace}
User selects metric → loads dimensions/tags via /api/v1/{namespace}/{subnamespace}/{metric}/dimensions
User selects required tags → metric is added to selectedMetrics array
Data Querying Phase:
Historical Data: Uses /api/v1/{namespace}/{subnamespace}/{metric}/data with time range parameters
Real-time Updates: WebSocket subscriptions for live metric updates
Tag Filtering: Metrics are filtered by selected tags at query time
5. Example Query in Action
When a chart widget queries a metric like system.cpu.usage_percent with tags {host: "server1", scope: "total"}:

6. Tag/Dimension System
Dimensions: Automatically added by collectors (e.g., host, device, container_name)
Tags: User-defined labels for filtering and grouping
Mandatory Tag Selection: In the new design, users must select at least one tag when adding metrics to ensure proper data scoping
7. Backend Query Translation
The backend translates these API calls to VictoriaMetrics PromQL queries:

This architecture provides a clean separation between the metric discovery (cached in-memory) and time-series data retrieval (VictoriaMetrics), enabling both fast UI interactions and efficient time-series queries.