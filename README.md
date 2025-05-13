![Go](https://img.shields.io/badge/built%20with-Go-blue) ![License](https://img.shields.io/github/license/aaronlmathis/gosight-server) ![Status](https://img.shields.io/badge/status-active-brightgreen)

# GoSight Server

GoSight Server is a high-performance observability backend built in Go. It receives telemetry from GoSight Agents, stores metrics and logs, processes alerts, and serves a dynamic web UI for visualization and investigation.

## Features

- TLS/mTLS-secured gRPC endpoints
- VictoriaMetrics integration for metric storage
- PostgreSQL or file-backed log/event storage
- Rule-based alert evaluation + dispatch
- Live WebSocket telemetry streaming
- REST API for metrics, logs, endpoints, and alerts
- Dynamic Flowbite+Tailwind UI with ApexCharts

## Architecture

- Receives telemetry via gRPC stream
- Stores metrics in VictoriaMetrics
- Handles logs, alerts, commands, and events
- Web dashboard renders views from real-time and historical data

## Build

\`\`\`bash
go build -o gosight-server ./cmd
\`\`\`

## Running

\`\`\`bash
./gosight-server --config ./config.yaml
\`\`\`

See sample config in \`./server/config/\`.

## Key Components

- \`api/\` – gRPC and REST endpoints
- \`telemetry/\` – stream ingestion, rule engine, tag index
- \`web/\` – HTML templates, static assets, and dashboard
- \`store/\` – metricstore, logstore, eventstore
- \`tracker/\` – in-memory agent and container registry

## License

GPL-3.0-or-later
