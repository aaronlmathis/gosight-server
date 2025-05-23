<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# metricindex

```go
import "github.com/aaronlmathis/gosight-server/internal/store/metricindex"
```

## Index

- [type MetricIndex](<#MetricIndex>)
  - [func NewMetricIndex\(\) \*MetricIndex](<#NewMetricIndex>)
  - [func \(idx \*MetricIndex\) Add\(namespace, sub, name string, dims map\[string\]string\)](<#MetricIndex.Add>)
  - [func \(idx \*MetricIndex\) FilterMetricNames\(filters map\[string\]string\) \[\]string](<#MetricIndex.FilterMetricNames>)
  - [func \(idx \*MetricIndex\) GetAllMetricNames\(\) \[\]string](<#MetricIndex.GetAllMetricNames>)
  - [func \(idx \*MetricIndex\) GetDimensions\(\) map\[string\]\[\]string](<#MetricIndex.GetDimensions>)
  - [func \(idx \*MetricIndex\) GetDimensionsForMetric\(fullMetric string\) \(\[\]string, error\)](<#MetricIndex.GetDimensionsForMetric>)
  - [func \(idx \*MetricIndex\) GetLabelValues\(label, contains string\) \[\]string](<#MetricIndex.GetLabelValues>)
  - [func \(idx \*MetricIndex\) GetMetricNames\(ns, sub string\) \[\]string](<#MetricIndex.GetMetricNames>)
  - [func \(idx \*MetricIndex\) GetNamespaces\(\) \[\]string](<#MetricIndex.GetNamespaces>)
  - [func \(idx \*MetricIndex\) GetSubNamespaces\(ns string\) \[\]string](<#MetricIndex.GetSubNamespaces>)
  - [func \(idx \*MetricIndex\) ListLabelValues\(key, contains string\) \[\]string](<#MetricIndex.ListLabelValues>)


<a name="MetricIndex"></a>
## type [MetricIndex](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L37-L45>)



```go
type MetricIndex struct {
    Namespaces       map[string]struct{}
    SubNamespaces    map[string]map[string]struct{}            // namespace → subnamespace
    MetricNames      map[string]map[string]map[string]struct{} // ns → sub → metric names
    Dimensions       map[string]map[string]struct{}            // dim key → value set
    MetricDimensions map[string]map[string]string              // metricFullName → dim key → value
    LabelValues      map[string]map[string]struct{}
    // contains filtered or unexported fields
}
```

<a name="NewMetricIndex"></a>
### func [NewMetricIndex](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L47>)

```go
func NewMetricIndex() *MetricIndex
```



<a name="MetricIndex.Add"></a>
### func \(\*MetricIndex\) [Add](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L58>)

```go
func (idx *MetricIndex) Add(namespace, sub, name string, dims map[string]string)
```



<a name="MetricIndex.FilterMetricNames"></a>
### func \(\*MetricIndex\) [FilterMetricNames](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L211>)

```go
func (idx *MetricIndex) FilterMetricNames(filters map[string]string) []string
```

FilterMetricNames returns all metric names that match given label filters

<a name="MetricIndex.GetAllMetricNames"></a>
### func \(\*MetricIndex\) [GetAllMetricNames](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L126>)

```go
func (idx *MetricIndex) GetAllMetricNames() []string
```



<a name="MetricIndex.GetDimensions"></a>
### func \(\*MetricIndex\) [GetDimensions](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L159>)

```go
func (idx *MetricIndex) GetDimensions() map[string][]string
```



<a name="MetricIndex.GetDimensionsForMetric"></a>
### func \(\*MetricIndex\) [GetDimensionsForMetric](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L171>)

```go
func (idx *MetricIndex) GetDimensionsForMetric(fullMetric string) ([]string, error)
```



<a name="MetricIndex.GetLabelValues"></a>
### func \(\*MetricIndex\) [GetLabelValues](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L188>)

```go
func (idx *MetricIndex) GetLabelValues(label, contains string) []string
```



<a name="MetricIndex.GetMetricNames"></a>
### func \(\*MetricIndex\) [GetMetricNames](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L145>)

```go
func (idx *MetricIndex) GetMetricNames(ns, sub string) []string
```



<a name="MetricIndex.GetNamespaces"></a>
### func \(\*MetricIndex\) [GetNamespaces](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L102>)

```go
func (idx *MetricIndex) GetNamespaces() []string
```



<a name="MetricIndex.GetSubNamespaces"></a>
### func \(\*MetricIndex\) [GetSubNamespaces](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L112>)

```go
func (idx *MetricIndex) GetSubNamespaces(ns string) []string
```



<a name="MetricIndex.ListLabelValues"></a>
### func \(\*MetricIndex\) [ListLabelValues](<https://github.com/aaronlmathis/gosight-server/blob/main/internal/store/metricindex/metricIndex.go#L253>)

```go
func (idx *MetricIndex) ListLabelValues(key, contains string) []string
```

ListLabelValues returns all known values for a given label key \(optionally filtered\)

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
