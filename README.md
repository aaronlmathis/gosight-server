![Go](https://img.shields.io/badge/built%20with-Go-blue) ![License](https://img.shields.io/github/license/aaronlmathis/gosight-server) ![Status](https://img.shields.io/badge/status-active-brightgreen) [![Go Report Card](https://goreportcard.com/badge/github.com/aaronlmathis/gosight-server)](https://goreportcard.com/report/github.com/aaronlmathis/gosight-server)
# GoSight Server

GoSight Server is a high-performance observability backend built in Go. It receives telemetry from GoSight Agents, OpenTelemetry Endpoints (GRPC) and SDK, and Syslog, stores metrics and logs, processes alerts, and serves a dynamic web UI for visualization and investigation.

## Documentation
[Documentation](docs/)

## Features

- OpenTelemetry Endpoint for ingesting Metrics, Logs, and Traces
- TLS/mTLS-secured gRPC endpoints
- RFC 3164 / RFC 5424 / CEF compatible SysLog Server for ingesting logs
- VictoriaMetrics integration for metric and log storage (Time-Series + Compressed JSON for logs)
- PostgreSQL or file-backed log/event storage (AWS S3 Connector Planned)
- Rule-based alert evaluation + dispatch
- Live WebSocket telemetry streaming
- REST API for metrics, logs, endpoints, and alerts
- Reactive Sveltekit+Flowbite+Tailwind UI with ApexCharts and customizable dashboards

## Architecture

- Receives telemetry via gRPC stream
- Stores metrics in VictoriaMetrics
- Handles logs, alerts, commands, and events
- Web dashboard renders views from real-time and historical data

## Build

```bash
go build -o gosight-server ./cmd
```

## Running

```bash
./gosight-server --config ./config.yaml
```

See sample config in `./server/config/`.

## Key Components / Directory Overview

- `internal/alerts/` – Alert models and rule evaluation engine
- `internal/auth/` – JWT-based authentication and session management
- `internal/bootstrap/` – Server startup logic and configuration loading
- `internal/bufferengine/` – Queue/buffer system for telemetry processing
- `internal/cache/` – In-memory caching layer for metadata and sessions
- `internal/config/` – Configuration file parsing and defaults
- `internal/contextutil/` – Request-scoped context helpers
- `internal/dispatcher/` – Alert action dispatcher for routes (webhook, script)
- `internal/events/` – Event tracking and structured broadcasting
- `internal/grpc/` – gRPC service registration and listener setup
- `internal/http/` – HTTP server handlers and routing (REST and UI)
- `internal/rules/` – Rule parsing, condition logic, and evaluation context
- `internal/runner/` – Metric and log task execution pipeline
- `internal/store/` – Top-level store interface and wrappers:
  - `alertstore/`
  - `datastore/`
  - `eventstore/`
  - `logstore/`
  - `metastore/`
  - `metricindex/`
  - `metricstore/`
  - `routestore/`
  - `rulestore/`
  - `userstore/`
- `internal/syncmanager/` – Live sync and periodic persistence
- `internal/sys/` – System-level information access
- `internal/syslog/` – SysLog server for ingesting logs
- `internal/telemetry/` – Metric and log ingestion + transformation
- `internal/testutils/` – Mocks and utilities for testing
- `internal/tracker/` – In-memory agent/container tracker
- `internal/usermodel/` – User, role, and permission models
- `internal/websocket/` – WebSocket hubs and live stream broadcasting

## License

GPL-3.0-or-later
