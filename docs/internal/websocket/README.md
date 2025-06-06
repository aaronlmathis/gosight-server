<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# websocket

```go
import "github.com/aaronlmathis/gosight-server/internal/websocket"
```

server/internal/http/websocket/commandhub.go

## Index

- [type AlertsHub](<#AlertsHub>)
  - [func NewAlertsHub\(metaTracker \*metastore.MetaTracker\) \*AlertsHub](<#NewAlertsHub>)
  - [func \(h \*AlertsHub\) Broadcast\(payload model.AlertInstance\)](<#AlertsHub.Broadcast>)
  - [func \(h \*AlertsHub\) Run\(ctx context.Context\)](<#AlertsHub.Run>)
  - [func \(h \*AlertsHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#AlertsHub.ServeWS>)
- [type Client](<#Client>)
  - [func \(c \*Client\) Close\(\)](<#Client.Close>)
- [type CommandHub](<#CommandHub>)
  - [func NewCommandHub\(metaTracker \*metastore.MetaTracker\) \*CommandHub](<#NewCommandHub>)
  - [func \(h \*CommandHub\) Broadcast\(result \*model.CommandResult\)](<#CommandHub.Broadcast>)
  - [func \(h \*CommandHub\) Run\(ctx context.Context\)](<#CommandHub.Run>)
  - [func \(h \*CommandHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#CommandHub.ServeWS>)
- [type EventsHub](<#EventsHub>)
  - [func NewEventsHub\(metaTracker \*metastore.MetaTracker\) \*EventsHub](<#NewEventsHub>)
  - [func \(h \*EventsHub\) Broadcast\(payload model.EventEntry\)](<#EventsHub.Broadcast>)
  - [func \(h \*EventsHub\) Run\(ctx context.Context\)](<#EventsHub.Run>)
  - [func \(h \*EventsHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#EventsHub.ServeWS>)
- [type HubManager](<#HubManager>)
  - [func NewHubManager\(metaTracker \*metastore.MetaTracker\) \*HubManager](<#NewHubManager>)
  - [func \(h \*HubManager\) StartAll\(ctx context.Context\)](<#HubManager.StartAll>)
- [type LogHub](<#LogHub>)
  - [func NewLogHub\(metaTracker \*metastore.MetaTracker\) \*LogHub](<#NewLogHub>)
  - [func \(h \*LogHub\) Broadcast\(payload model.LogPayload\)](<#LogHub.Broadcast>)
  - [func \(h \*LogHub\) Run\(ctx context.Context\)](<#LogHub.Run>)
  - [func \(h \*LogHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#LogHub.ServeWS>)
- [type MetricHub](<#MetricHub>)
  - [func NewMetricHub\(metaTracker \*metastore.MetaTracker\) \*MetricHub](<#NewMetricHub>)
  - [func \(h \*MetricHub\) Broadcast\(payload model.MetricPayload\)](<#MetricHub.Broadcast>)
  - [func \(h \*MetricHub\) Run\(ctx context.Context\)](<#MetricHub.Run>)
  - [func \(h \*MetricHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#MetricHub.ServeWS>)
- [type ProcessHub](<#ProcessHub>)
  - [func NewProcessHub\(metaTracker \*metastore.MetaTracker\) \*ProcessHub](<#NewProcessHub>)
  - [func \(h \*ProcessHub\) Broadcast\(payload model.ProcessPayload\)](<#ProcessHub.Broadcast>)
  - [func \(h \*ProcessHub\) Run\(ctx context.Context\)](<#ProcessHub.Run>)
  - [func \(h \*ProcessHub\) ServeWS\(w http.ResponseWriter, r \*http.Request\)](<#ProcessHub.ServeWS>)


<a name="AlertsHub"></a>
## type [AlertsHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/alerthub.go#L41-L46>)



```go
type AlertsHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewAlertsHub"></a>
### func [NewAlertsHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/alerthub.go#L49>)

```go
func NewAlertsHub(metaTracker *metastore.MetaTracker) *AlertsHub
```

NewAlertsHub creates a new AlertsHub instance.

<a name="AlertsHub.Broadcast"></a>
### func \(\*AlertsHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/alerthub.go#L149>)

```go
func (h *AlertsHub) Broadcast(payload model.AlertInstance)
```



<a name="AlertsHub.Run"></a>
### func \(\*AlertsHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/alerthub.go#L57>)

```go
func (h *AlertsHub) Run(ctx context.Context)
```



<a name="AlertsHub.ServeWS"></a>
### func \(\*AlertsHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/alerthub.go#L92>)

```go
func (h *AlertsHub) ServeWS(w http.ResponseWriter, r *http.Request)
```



<a name="Client"></a>
## type [Client](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/hub.go#L39-L47>)



```go
type Client struct {
    Conn       *websocket.Conn
    EndpointID string
    AgentID    string
    HostID     string
    Send       chan []byte
    ID         string
    // contains filtered or unexported fields
}
```

<a name="Client.Close"></a>
### func \(\*Client\) [Close](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/hub.go#L106>)

```go
func (c *Client) Close()
```



<a name="CommandHub"></a>
## type [CommandHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/command.go#L39-L44>)



```go
type CommandHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewCommandHub"></a>
### func [NewCommandHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/command.go#L46>)

```go
func NewCommandHub(metaTracker *metastore.MetaTracker) *CommandHub
```



<a name="CommandHub.Broadcast"></a>
### func \(\*CommandHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/command.go#L142>)

```go
func (h *CommandHub) Broadcast(result *model.CommandResult)
```



<a name="CommandHub.Run"></a>
### func \(\*CommandHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/command.go#L54>)

```go
func (h *CommandHub) Run(ctx context.Context)
```



<a name="CommandHub.ServeWS"></a>
### func \(\*CommandHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/command.go#L89>)

```go
func (h *CommandHub) ServeWS(w http.ResponseWriter, r *http.Request)
```



<a name="EventsHub"></a>
## type [EventsHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/eventhub.go#L41-L46>)



```go
type EventsHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewEventsHub"></a>
### func [NewEventsHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/eventhub.go#L48>)

```go
func NewEventsHub(metaTracker *metastore.MetaTracker) *EventsHub
```



<a name="EventsHub.Broadcast"></a>
### func \(\*EventsHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/eventhub.go#L157>)

```go
func (h *EventsHub) Broadcast(payload model.EventEntry)
```



<a name="EventsHub.Run"></a>
### func \(\*EventsHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/eventhub.go#L56>)

```go
func (h *EventsHub) Run(ctx context.Context)
```



<a name="EventsHub.ServeWS"></a>
### func \(\*EventsHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/eventhub.go#L97>)

```go
func (h *EventsHub) ServeWS(w http.ResponseWriter, r *http.Request)
```



<a name="HubManager"></a>
## type [HubManager](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/hub.go#L49-L56>)



```go
type HubManager struct {
    Metrics   *MetricHub
    Logs      *LogHub
    Alerts    *AlertsHub
    Events    *EventsHub
    Commands  *CommandHub
    Processes *ProcessHub
}
```

<a name="NewHubManager"></a>
### func [NewHubManager](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/hub.go#L59>)

```go
func NewHubManager(metaTracker *metastore.MetaTracker) *HubManager
```

NewHubManager creates a new HubManager with initialized hubs.

<a name="HubManager.StartAll"></a>
### func \(\*HubManager\) [StartAll](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/hub.go#L71>)

```go
func (h *HubManager) StartAll(ctx context.Context)
```

StartAll starts all hubs in separate goroutines.

<a name="LogHub"></a>
## type [LogHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/loghub.go#L42-L47>)



```go
type LogHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewLogHub"></a>
### func [NewLogHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/loghub.go#L49>)

```go
func NewLogHub(metaTracker *metastore.MetaTracker) *LogHub
```



<a name="LogHub.Broadcast"></a>
### func \(\*LogHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/loghub.go#L158>)

```go
func (h *LogHub) Broadcast(payload model.LogPayload)
```



<a name="LogHub.Run"></a>
### func \(\*LogHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/loghub.go#L57>)

```go
func (h *LogHub) Run(ctx context.Context)
```



<a name="LogHub.ServeWS"></a>
### func \(\*LogHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/loghub.go#L99>)

```go
func (h *LogHub) ServeWS(w http.ResponseWriter, r *http.Request)
```



<a name="MetricHub"></a>
## type [MetricHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/metrichub.go#L42-L47>)



```go
type MetricHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewMetricHub"></a>
### func [NewMetricHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/metrichub.go#L49>)

```go
func NewMetricHub(metaTracker *metastore.MetaTracker) *MetricHub
```



<a name="MetricHub.Broadcast"></a>
### func \(\*MetricHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/metrichub.go#L160>)

```go
func (h *MetricHub) Broadcast(payload model.MetricPayload)
```

Broadcast sends a pre\-serialized metric payload to all MetricHub clients.

<a name="MetricHub.Run"></a>
### func \(\*MetricHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/metrichub.go#L58>)

```go
func (h *MetricHub) Run(ctx context.Context)
```

Run starts the MetricHub's broadcast loop.

<a name="MetricHub.ServeWS"></a>
### func \(\*MetricHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/metrichub.go#L100>)

```go
func (h *MetricHub) ServeWS(w http.ResponseWriter, r *http.Request)
```

ServeWS upgrades HTTP to WebSocket and registers client to MetricHub.

<a name="ProcessHub"></a>
## type [ProcessHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/processhub.go#L41-L46>)



```go
type ProcessHub struct {
    // contains filtered or unexported fields
}
```

<a name="NewProcessHub"></a>
### func [NewProcessHub](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/processhub.go#L48>)

```go
func NewProcessHub(metaTracker *metastore.MetaTracker) *ProcessHub
```



<a name="ProcessHub.Broadcast"></a>
### func \(\*ProcessHub\) [Broadcast](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/processhub.go#L157>)

```go
func (h *ProcessHub) Broadcast(payload model.ProcessPayload)
```

Broadcast sends a pre\-serialized metric payload to all ProcessHub clients.

<a name="ProcessHub.Run"></a>
### func \(\*ProcessHub\) [Run](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/processhub.go#L57>)

```go
func (h *ProcessHub) Run(ctx context.Context)
```

Run starts the ProcessHub's broadcast loop.

<a name="ProcessHub.ServeWS"></a>
### func \(\*ProcessHub\) [ServeWS](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/websocket/processhub.go#L98>)

```go
func (h *ProcessHub) ServeWS(w http.ResponseWriter, r *http.Request)
```

ServeWS upgrades HTTP to WebSocket and registers client to ProcessHub.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
