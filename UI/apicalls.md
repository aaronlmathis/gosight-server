# GoSight Frontend API Calls

This document provides a comprehensive list of all API endpoints called by the SvelteKit frontend when interacting with the Go backend. All API routes are prefixed with `/api/v1`.

## Frontend-Backend Architecture

The GoSight application uses a SvelteKit frontend that communicates with a Go backend:

```
┌───────────────────┐         ┌───────────────────┐
│                   │ HTTP/WS │                   │
│  SvelteKit        │◄────────►  Go Backend       │
│  Frontend         │ API     │  Server           │
│                   │         │                   │
└───────────────────┘         └───────────────────┘
```

### SvelteKit Frontend

The frontend is implemented using SvelteKit, with:

- TypeScript for type safety
- API client (`src/lib/api.ts`) for server communication
- WebSocket connections for real-time updates
- Tailwind CSS for styling
- ApexCharts for data visualization

### Go Backend

The backend is implemented in Go and provides:

- RESTful API endpoints (`/api/v1/*`)
- WebSocket endpoints for real-time data streaming (`/ws/*`)
- Authentication and authorization
- Business logic and data processing
- Data storage and retrieval

The frontend does not access data stores directly—all data access is mediated through the backend API.

## General API Response Format

Most API responses follow this standard format:

```json
{
	"status": "success",
	"data": [
		/* response data */
	],
	"metadata": {
		"total": 100,
		"pages": 5,
		"current_page": 1,
		"limit": 20
	}
}
```

Error responses follow this format:

```json
{
	"status": "error",
	"message": "Error description",
	"code": "ERROR_CODE"
}
```

## Pagination

Many endpoints that return lists of items support pagination with these query parameters:

| Parameter | Description                      | Default            |
| --------- | -------------------------------- | ------------------ |
| `limit`   | Number of items per page         | 20                 |
| `page`    | Page number (1-based)            | 1                  |
| `sort`    | Field to sort by                 | Varies by endpoint |
| `order`   | Sort direction (`asc` or `desc`) | `desc`             |

The response metadata includes pagination details:

- `total`: Total number of items
- `pages`: Total number of pages
- `current_page`: Current page number
- `limit`: Items per page

## Authentication

Most API endpoints require authentication. The SvelteKit frontend handles this using JWT tokens:

1. The frontend obtains a JWT token via the `/api/v1/auth/login` endpoint
2. The token is stored securely and included in subsequent requests
3. If a request returns a 401 Unauthorized error, the frontend will redirect to the login page

Authentication is managed using secure HTTP-only cookies that contain the JWT token.

## Common HTTP Status Codes

| Code | Description           | Meaning                                        |
| ---- | --------------------- | ---------------------------------------------- |
| 200  | OK                    | Request successful                             |
| 201  | Created               | Resource created successfully                  |
| 400  | Bad Request           | Invalid request parameters or body             |
| 401  | Unauthorized          | Authentication required or failed              |
| 403  | Forbidden             | User lacks permission for the requested action |
| 404  | Not Found             | Resource not found                             |
| 409  | Conflict              | Resource conflict (e.g., duplicate entity)     |
| 422  | Unprocessable Entity  | Validation errors                              |
| 429  | Too Many Requests     | Rate limit exceeded                            |
| 500  | Internal Server Error | Server error occurred                          |

## Error Handling

The API client (`src/lib/api.ts`) includes built-in error handling. All API errors are wrapped in a `GoSightApiError` class which includes:

- `message`: Human-readable error message
- `status`: HTTP status code

Example of proper error handling in components:

```typescript
import { api, GoSightApiError } from '$lib/api';

async function fetchData() {
	try {
		const response = await api.endpoints.getAll();
		return response.data;
	} catch (error) {
		if (error instanceof GoSightApiError) {
			// Handle specific status codes
			if (error.status === 401) {
				// Handle authentication error
				return { error: 'Please log in to continue' };
			} else if (error.status === 403) {
				// Handle permission error
				return { error: 'You do not have permission to access this resource' };
			} else if (error.status === 404) {
				// Handle not found error
				return { error: 'The requested resource was not found' };
			} else {
				// Handle other API errors
				return { error: error.message || 'An unknown error occurred' };
			}
		} else {
			// Handle unexpected errors
			console.error('Unexpected error:', error);
			return { error: 'An unexpected error occurred' };
		}
	}
}
```

## Code Examples

### Fetching Data with the API Client

The SvelteKit frontend uses a unified API client defined in `src/lib/api.ts`:

```typescript
// Fetch alerts with pagination
async function loadAlerts(page = 1) {
	try {
		const response = await api.alerts.getAll({
			limit: 20,
			page: page,
			state: 'active',
			sort: 'timestamp',
			order: 'desc'
		});
		return response.data;
	} catch (error) {
		console.error('Failed to load alerts:', error);
		return [];
	}
}

// Get endpoint details
async function getEndpointDetails(id) {
	try {
		const response = await api.endpoints.get(id);
		return response.data;
	} catch (error) {
		console.error(`Failed to get endpoint ${id}:`, error);
		return null;
	}
}

// Send a command to an endpoint
async function executeCommand(endpointId, command, args = []) {
	try {
		const response = await api.commands.send(endpointId, {
			command: command,
			args: args
		});
		return response.data;
	} catch (error) {
		console.error('Command execution failed:', error);
		throw error;
	}
}
```

### Working with WebSockets

Using the WebSocket utilities defined in `src/lib/websocket.ts`:

```typescript
import { websocketManager } from '$lib/websocket';

// Subscribe to real-time alerts
const unsubscribeAlerts = websocketManager.subscribeToAlerts((alertData) => {
	console.log('New alert received:', alertData);
	// Update UI or state with the new alert
});

// Connect to specific endpoint's metrics
websocketManager.metrics.connect();

// Clean up subscriptions when component is destroyed
onDestroy(() => {
	unsubscribeAlerts();
	websocketManager.metrics.disconnect();
});
```

## Data Filtering

Many API endpoints support advanced filtering to narrow down results:

### Time-Based Filtering

For endpoints that return time-series data, use the following parameters:

- `start`: Start timestamp (ISO 8601 format, e.g., `2025-01-15T00:00:00Z`)
- `end`: End timestamp (ISO 8601 format)

Alternatively, some endpoints support relative time using:

- `range`: A human-readable time range (e.g., `1h`, `24h`, `7d`, `30d`)

### Resource-Specific Filtering

Endpoints typically support filtering by their resource attributes:

- Alerts: `level` (critical, warning, info), `state` (firing, resolved, acknowledged)
- Logs: `level` (debug, info, warn, error, critical), `contains` (text search)
- Metrics: `name`, `namespace`, `endpoint_id`
- Events: `type`, `category`, `source`, `target`

### Text Search

For free text search, use:

- `contains` parameter for specific resource endpoints
- `/search?q=query` endpoint for global search across all resources

## Glossary of Terms

| Term          | Description                                                |
| ------------- | ---------------------------------------------------------- |
| **Alert**     | A notification triggered by a rule condition being met     |
| **Endpoint**  | A monitored entity (host, container, network device)       |
| **Event**     | A discrete occurrence in the system (state change, action) |
| **Host**      | A physical or virtual machine running the GoSight agent    |
| **Container** | An isolated environment running within a host              |
| **Log**       | A recorded message from a system or application            |
| **Metric**    | A measurable value representing system performance         |
| **Namespace** | A categorization for metrics (e.g., system, container)     |
| **Rule**      | A condition that triggers an alert when met                |
| **Tag**       | A key-value label attached to endpoints for categorization |

## Alerts API

| Endpoint                        | Method | Description                                  | Parameters                                                                   |
| ------------------------------- | ------ | -------------------------------------------- | ---------------------------------------------------------------------------- |
| `/alerts`                       | GET    | Get a list of alerts with optional filtering | `limit`, `page`, `state`, `level`, `rule_id`, `sort`, `order`, `endpoint_id` |
| `/alerts`                       | POST   | Create a new alert                           | Alert object in request body                                                 |
| `/alerts/active`                | GET    | Get only active alerts                       | None                                                                         |
| `/alerts/rules`                 | GET    | Get all alert rules                          | None                                                                         |
| `/alerts/rules`                 | POST   | Create a new alert rule                      | Rule object in request body                                                  |
| `/alerts/rules/{ruleId}`        | PUT    | Update an existing alert rule                | Rule object in request body                                                  |
| `/alerts/rules/{ruleId}`        | DELETE | Delete an alert rule                         | None                                                                         |
| `/alerts/summary`               | GET    | Get alerts summary statistics                | None                                                                         |
| `/alerts/{alertId}/acknowledge` | POST   | Acknowledge a specific alert                 | None                                                                         |
| `/alerts/{alertId}/resolve`     | POST   | Resolve a specific alert                     | None                                                                         |
| `/alerts/{alertId}/context`     | GET    | Get context around a specific alert          | Optional `window` parameter                                                  |

## Endpoints API

| Endpoint                | Method | Description                               | Parameters                   |
| ----------------------- | ------ | ----------------------------------------- | ---------------------------- |
| `/endpoints`            | GET    | Get all endpoints with optional filtering | `type`, `status`, `hostname` |
| `/endpoints/{id}`       | GET    | Get a specific endpoint by ID             | None                         |
| `/endpoints/hosts`      | GET    | Get only host endpoints                   | Various filter parameters    |
| `/endpoints/containers` | GET    | Get only container endpoints              | Various filter parameters    |

## Events API

| Endpoint         | Method | Description                        | Parameters                                                                                                                                   |
| ---------------- | ------ | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `/events`        | GET    | Get events with optional filtering | `limit`, `level`, `type`, `category`, `scope`, `target`, `source`, `contains`, `start`, `end`, `hostID`, `endpointID`, `endpoint_id`, `sort` |
| `/events/recent` | GET    | Get recent events                  | `limit` (defaults to 10)                                                                                                                     |

## Logs API

| Endpoint       | Method | Description                      | Parameters                                                                                          |
| -------------- | ------ | -------------------------------- | --------------------------------------------------------------------------------------------------- |
| `/logs`        | GET    | Get logs with optional filtering | `limit`, `page`, `level`, `contains`, `start`, `end`, `hostID`, `endpointID`, `endpoint_id`, `sort` |
| `/logs/latest` | GET    | Get most recent logs             | `limit` (defaults to 50)                                                                            |

## Metrics API

| Endpoint                                                             | Method | Description                            | Parameters                                                   |
| -------------------------------------------------------------------- | ------ | -------------------------------------- | ------------------------------------------------------------ |
| `/metrics`                                                           | GET    | Get metrics with optional filtering    | `limit`, `endpoint_id`, `endpointID`, `start`, `end`, `name` |
| `/metrics/system`                                                    | GET    | Get system overview metrics            | None                                                         |
| `/metrics/namespaces`                                                | GET    | Get all metric namespaces              | None                                                         |
| `/metrics/namespaces/{namespace}`                                    | GET    | Get sub-namespaces for a namespace     | None                                                         |
| `/metrics/namespaces/{namespace}/{subNamespace}`                     | GET    | Get metric names for a sub-namespace   | None                                                         |
| `/metrics/namespaces/{namespace}/{subNamespace}/{metric}/dimensions` | GET    | Get dimensions for a specific metric   | None                                                         |
| `/metrics/namespaces/{namespace}/{subNamespace}/{metric}/data`       | GET    | Get data for a specific metric         | Various time series parameters                               |
| `/metrics/namespaces/{namespace}/{subNamespace}/{metric}/latest`     | GET    | Get latest value for a specific metric | None                                                         |
| `/metrics/query`                                                     | GET    | Query metrics with custom parameters   | Various query parameters                                     |

## Reports API

| Endpoint           | Method | Description                        | Parameters                                            |
| ------------------ | ------ | ---------------------------------- | ----------------------------------------------------- |
| `/reports/summary` | GET    | Get system summary report          | `range` (time range)                                  |
| `/reports/alerts`  | GET    | Get alerts report                  | `range`, optional `endpoints` array                   |
| `/reports/metrics` | GET    | Get metrics report                 | `range`, optional `endpoints` array                   |
| `/reports/events`  | GET    | Get events report                  | `range`, optional `endpoints` array                   |
| `/reports/export`  | GET    | Export a report in various formats | `type`, `format`, `range`, optional `endpoints` array |

## Commands API

| Endpoint                 | Method | Description                           | Parameters                                               |
| ------------------------ | ------ | ------------------------------------- | -------------------------------------------------------- |
| `/commands/{endpointId}` | POST   | Send a command to a specific endpoint | Command object with `command` and `args` in request body |

## Auth API

| Endpoint             | Method | Description                    | Parameters                                                                                               |
| -------------------- | ------ | ------------------------------ | -------------------------------------------------------------------------------------------------------- |
| `/auth/login`        | POST   | User login                     | Credentials with `username` and `password` in request body                                               |
| `/auth/register`     | POST   | Register a new user            | User data with `username`, `email`, `password`, `confirmPassword`, optional `first_name` and `last_name` |
| `/auth/logout`       | POST   | User logout                    | None                                                                                                     |
| `/auth/me`           | GET    | Get current authenticated user | None                                                                                                     |
| `/users/profile`     | PUT    | Update user profile            | Profile data in request body                                                                             |
| `/users/password`    | PUT    | Update user password           | Password data with `current_password`, `new_password`, `confirm_password` in request body                |
| `/users/preferences` | GET    | Get user preferences           | None                                                                                                     |
| `/users/preferences` | PUT    | Update user preferences        | Preferences object in request body                                                                       |

## Network Devices API

| Endpoint                       | Method | Description                      | Parameters                    |
| ------------------------------ | ------ | -------------------------------- | ----------------------------- |
| `/network-devices`             | GET    | Get all network devices          | None                          |
| `/network-devices`             | POST   | Create a new network device      | Device object in request body |
| `/network-devices/{id}`        | PUT    | Update a network device          | Device object in request body |
| `/network-devices/{id}`        | DELETE | Delete a network device          | None                          |
| `/network-devices/{id}/toggle` | POST   | Toggle a network device's status | None                          |

## Tags API

| Endpoint                   | Method | Description                            | Parameters                  |
| -------------------------- | ------ | -------------------------------------- | --------------------------- |
| `/tags/keys`               | GET    | Get all available tag keys             | None                        |
| `/tags/values`             | GET    | Get all available tag values           | None                        |
| `/tags/{endpointId}`       | GET    | Get tags for a specific endpoint       | None                        |
| `/tags/{endpointId}`       | POST   | Set tags for a specific endpoint       | Tags object in request body |
| `/tags/{endpointId}`       | PATCH  | Update tags for a specific endpoint    | Tags object in request body |
| `/tags/{endpointId}/{key}` | DELETE | Delete a specific tag from an endpoint | None                        |

## Search API

| Endpoint  | Method | Description                              | Parameters         |
| --------- | ------ | ---------------------------------------- | ------------------ |
| `/search` | GET    | Perform a global search across resources | `q` (search query) |

## WebSocket Connections

In addition to REST API calls, the frontend maintains WebSocket connections for real-time updates:

### Available WebSocket Endpoints

| WebSocket Endpoint | Description                            | Query Parameters                            |
| ------------------ | -------------------------------------- | ------------------------------------------- |
| `/ws/alerts`       | Real-time alerts as they are generated | None                                        |
| `/ws/events`       | Real-time events as they occur         | Optional `endpointID` to filter by endpoint |
| `/ws/logs`         | Real-time log entries                  | Optional `endpointID` to filter by endpoint |
| `/ws/metrics`      | Real-time metric updates               | Optional `endpointID` to filter by endpoint |
| `/ws/command`      | Real-time command execution results    | Optional `endpointID` to filter by endpoint |
| `/ws/process`      | Real-time process information          | Optional `endpointID` to filter by endpoint |

### WebSocket Message Format

Messages received from WebSocket endpoints are typically JSON-formatted and include:

- Timestamp
- Type-specific fields (vary by endpoint)
- Associated metadata

Example alert message:

```json
{
	"id": "alert-123",
	"rule_id": "rule-456",
	"level": "critical",
	"message": "CPU usage exceeds threshold",
	"timestamp": "2023-04-15T14:22:36Z",
	"endpoint_id": "endpoint-789",
	"state": "firing",
	"details": {
		"value": 95.2,
		"threshold": 90
	}
}
```

These WebSocket connections allow the dashboard to update in real-time without polling the server.
