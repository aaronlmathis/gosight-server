<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# rules

```go
import "github.com/aaronlmathis/gosight-server/internal/rules"
```

Package rules provides the core logic for evaluating alert rules and emitting events based on metric data.

## Index

- [type Evaluator](<#Evaluator>)
  - [func NewEvaluator\(store rulestore.RuleStore, alertMgr \*alerts.Manager\) \*Evaluator](<#NewEvaluator>)
  - [func \(e \*Evaluator\) EvaluateLogs\(ctx context.Context, logs \[\]model.LogEntry, meta \*model.Meta\)](<#Evaluator.EvaluateLogs>)
  - [func \(e \*Evaluator\) EvaluateMetric\(ctx context.Context, metrics \[\]model.Metric, meta \*model.Meta\)](<#Evaluator.EvaluateMetric>)


<a name="Evaluator"></a>
## type [Evaluator](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/rules/engine.go#L45-L50>)

Evaluator is responsible for evaluating alert rules against incoming metrics and logs. It uses a RuleStore to manage the rules and an AlertManager to handle the state of alerts. The Evaluator maintains a history of metrics for each rule and endpoint combination, allowing it to track the state of alerts over time. The firing map is used to track which rules are currently firing for each endpoint, preventing duplicate alerts from being emitted.

```go
type Evaluator struct {
    AlertMgr *alerts.Manager
    // contains filtered or unexported fields
}
```

<a name="NewEvaluator"></a>
### func [NewEvaluator](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/rules/engine.go#L59>)

```go
func NewEvaluator(store rulestore.RuleStore, alertMgr *alerts.Manager) *Evaluator
```



<a name="Evaluator.EvaluateLogs"></a>
### func \(\*Evaluator\) [EvaluateLogs](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/rules/engine.go#L137>)

```go
func (e *Evaluator) EvaluateLogs(ctx context.Context, logs []model.LogEntry, meta *model.Meta)
```

EvaluateLogs processes the given logs and metadata, checking them against active rules in the store. It emits events when rules are triggered based on the logs. The evaluation is done in the context of the provided context.Context. The logs are expected to be in the format of model.LogEntry, and the metadata is expected to be in the format of model.Meta. Logs are point\-in\-time events, so they are always evaluated immediately.

<a name="Evaluator.EvaluateMetric"></a>
### func \(\*Evaluator\) [EvaluateMetric](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/rules/engine.go#L74>)

```go
func (e *Evaluator) EvaluateMetric(ctx context.Context, metrics []model.Metric, meta *model.Meta)
```

EvaluateMetric processes the given metrics and metadata, checking them against active rules in the store. It emits events when rules are triggered based on the metrics. The evaluation is done in the context of the provided context.Context. The metrics are expected to be in the format of model.Metric, and the metadata is expected to be in the format of model.Meta.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
