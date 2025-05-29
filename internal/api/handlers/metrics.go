// File: server/internal/api/handlers/metrics.go
// Description: This file contains the metrics and metadata handlers for the GoSight server.

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	"github.com/aaronlmathis/gosight-shared/utils"
	"github.com/gorilla/mux"
)

// MetricsHandler provides handlers for metrics and metadata API endpoints
type MetricsHandler struct {
	Sys *sys.SystemContext
}

// NewMetricsHandler creates a new MetricsHandler
func NewMetricsHandler(sys *sys.SystemContext) *MetricsHandler {
	return &MetricsHandler{
		Sys: sys,
	}
}

// GetNamespaces retrieves all namespaces.
// It returns a JSON object containing the namespaces.
// The URL format is: /api/v1/
func (h *MetricsHandler) GetNamespaces(w http.ResponseWriter, r *http.Request) {
	namespaces := h.Sys.Tele.Index.GetNamespaces()
	sort.Strings(namespaces)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(namespaces)
}

// GetSubNamespaces retrieves all sub-namespaces for a given namespace.
// It returns a JSON object containing the sub-namespaces.
// The URL format is: /api/v1/{namespace}
func (h *MetricsHandler) GetSubNamespaces(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := strings.ToLower(vars["namespace"])

	subNamespaces := h.Sys.Tele.Index.GetSubNamespaces(namespace)
	sort.Strings(subNamespaces)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(subNamespaces)
}

// GetMetricNames retrieves all metric names for a given namespace and subnamespace.
// It returns a JSON object containing the metric names.
// The URL format is: /api/v1/{namespace}/{sub}/metrics
func (h *MetricsHandler) GetMetricNames(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])

	metricNames := h.Sys.Tele.Index.GetMetricNames(namespace, sub)
	sort.Strings(metricNames)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metricNames)
}

// GetDimensions retrieves all available dimensions.
// It returns a JSON object containing the dimensions.
// The URL format is: /api/v1/dimensions
func (h *MetricsHandler) GetDimensions(w http.ResponseWriter, r *http.Request) {
	dimensions := h.Sys.Tele.Index.GetDimensions()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dimensions)
}

// GetMetricDimensions retrieves the dimensions for a specific metric.
// It accepts a namespace, subnamespace, and metric name as URL parameters.
// The response is a JSON object containing the dimensions.
// The URL format is: /api/v1/{namespace}/{sub}/{metric}/dimensions
func (h *MetricsHandler) GetMetricDimensions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	fullMetric := fmt.Sprintf("%s.%s.%s", namespace, sub, metric)
	//dimensions := h.Sys.Tele.Index.GetMetricDimensions(fullMetric)
	dimensions := h.Sys.Cache.Metrics.GetMetricDimensions(fullMetric)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dimensions)
}

// GetMetricData retrieves metric data for a specific metric.
// It accepts a namespace, subnamespace, and metric name as URL parameters.
// The response is a JSON object containing the metric data.
// The URL format is: /api/v1/{namespace}/{sub}/{metric}/data
func (h *MetricsHandler) GetMetricData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	fullMetric := fmt.Sprintf("%s.%s.%s", ns, sub, metric)

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	stepStr := r.URL.Query().Get("step")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid start time: %v", err)})
			return
		}
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			utils.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid end time: %v", err)})
			return
		}
	}

	if start.IsZero() || end.IsZero() {
		utils.JSON(w, http.StatusBadRequest, map[string]string{"error": "start and end time must be specified"})
		return
	}

	filters := h.parseQueryFilters(r)

	points, err := h.Sys.Stores.Metrics.QueryRange(fullMetric, start, end, stepStr, filters)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to query range data: %v", err)})
		return
	}

	utils.JSON(w, http.StatusOK, points)
}

// GetMetricLatest retrieves the latest value for a specific metric.
// It accepts a namespace, subnamespace, and metric name as URL parameters.
// The response is a JSON object containing the latest value and timestamp.
// The URL format is: /api/v1/{namespace}/{sub}/{metric}/latest
func (h *MetricsHandler) GetMetricLatest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ns := strings.ToLower(vars["namespace"])
	sub := strings.ToLower(vars["sub"])
	metric := strings.ToLower(vars["metric"])

	fullMetric := fmt.Sprintf("%s.%s.%s", ns, sub, metric)
	filters := h.parseQueryFilters(r)

	latest, err := h.Sys.Stores.Metrics.QueryInstant(fullMetric, filters)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to query latest data: %v", err)})
		return
	}

	utils.JSON(w, http.StatusOK, latest)
}

// HandleAPIQuery handles flexible label-based queries without requiring a metric name.
// Supports optional time range via start= and end= query params.
// It also supports sorting and limiting the results.
// The query parameters are:
// - metric: the metric name(s) to query
// - start: the start time for the query (RFC3339 format)
// - end: the end time for the query (RFC3339 format)
// - step: the step interval for the query (default is 15s)
// - limit: the maximum number of results to return
// - sort: the sort order for the results (asc or desc)
// - tags: additional filters for the query (key=value pairs)
// The response is a JSON object containing the query results.
func (h *MetricsHandler) HandleAPIQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	metricNames := query["metric"]

	// Optional time range
	startStr := query.Get("start")
	endStr := query.Get("end")

	stepStr := query.Get("step")
	if stepStr == "" {
		stepStr = "15s" // default to 15s if not provided
	}

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "invalid 'start' format (RFC3339)", http.StatusBadRequest)
			return
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "invalid 'end' format (RFC3339)", http.StatusBadRequest)
			return
		}
	}

	limitStr := query.Get("limit")
	sortOrder := query.Get("sort") // "asc" or "desc"

	var limit int
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "invalid 'limit' value", http.StatusBadRequest)
			return
		}
	}

	filters := make(map[string]string)
	for key, vals := range query {
		if len(vals) == 0 {
			continue
		}
		switch key {
		case "metric", "start", "end", "limit", "sort":
			continue
		case "tags":
			tagParts := strings.Split(vals[0], ",")
			for _, part := range tagParts {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					filters[kv[0]] = kv[1]
				}
			}
		default:
			filters[key] = vals[0]
		}
	}

	if len(filters) == 0 && len(metricNames) == 0 {
		http.Error(w, "must specify at least one filter or a metric name", http.StatusBadRequest)
		return
	}

	var result any

	switch {
	case len(metricNames) > 0 && !start.IsZero() && !end.IsZero():
		result, err = h.Sys.Stores.Metrics.QueryMultiRange(metricNames, start, end, stepStr, filters)

	case len(metricNames) > 0:
		result, err = h.Sys.Stores.Metrics.QueryMultiInstant(metricNames, filters)
	case len(metricNames) == 0:
		// Power mode — return matching metrics across all known names
		names := h.Sys.Tele.Index.FilterMetricNames(filters)
		if len(names) == 0 {
			http.Error(w, "no metrics matched filters", http.StatusNotFound)
			return
		}

		if !start.IsZero() && !end.IsZero() {
			result, err = h.Sys.Stores.Metrics.QueryMultiRange(names, start, end, stepStr, filters)
		} else {
			result, err = h.Sys.Stores.Metrics.QueryMultiInstant(names, filters)
		}
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("query failed: %v", err), http.StatusInternalServerError)
		return
	}

	// don't let a null result be returned to api
	if result == nil {
		result = []model.MetricRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	if sortOrder != "" || limit > 0 {
		result = h.applySortAndLimit(result, sortOrder, limit)
	}
	_ = json.NewEncoder(w).Encode(result)
}

// HandleExportQuery handles flexible label-based queries without requiring a metric name.
// Supports optional time range via start= and end= query params.
func (h *MetricsHandler) HandleExportQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Extract label filters
	labels := make([]string, 0)
	for k, vals := range q {
		if len(vals) > 0 && k != "start" && k != "end" {
			labels = append(labels, fmt.Sprintf(`%s="%s"`, k, vals[0]))
		}
	}

	// Build match[] expression
	sort.Strings(labels)
	matchExpr := fmt.Sprintf("{%s}", strings.Join(labels, ","))

	params := url.Values{}
	params.Add("match[]", matchExpr)

	// Optional time range
	if start := q.Get("start"); start != "" {
		params.Add("start", start)
	}
	if end := q.Get("end"); end != "" {
		params.Add("end", end)
	}
	if _, ok := q["start"]; !ok {
		params.Add("start", fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix()))
	}
	if _, ok := q["end"]; !ok {
		params.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	}

	// Format: json or prom line format
	if format := q.Get("format"); format != "" {
		params.Add("format", format)
	}

	// Build final URL
	exportURL := fmt.Sprintf("%s?%s", h.Sys.Cfg.MetricStore.URL+"/api/v1/export", params.Encode())

	resp, err := http.Get(exportURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Export query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
	}
}

// Utility functions

// applySortAndLimit sorts and limits the data based on the provided sort order and limit.
// It assumes the data is a slice of model.MetricRow.
// If the data is not of this type, it returns the original data.
func (h *MetricsHandler) applySortAndLimit(data any, sortKey string, limit int) any {
	rows, ok := data.([]model.MetricRow)
	if !ok {
		return data
	}

	if sortKey == "asc" {
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Value < rows[j].Value
		})
	} else if sortKey == "desc" {
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Value > rows[j].Value
		})
	}

	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}

	return rows
}

// parseQueryFilters parses the query parameters from the request and returns a map of filters.
// It ignores the "start", "end", "latest", and "step" parameters.
// It also handles multiple values for the same key by creating a regex pattern.
func (h *MetricsHandler) parseQueryFilters(r *http.Request) map[string]string {
	filters := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}
		// Skip time-related and control parameters
		if key == "start" || key == "end" || key == "latest" || key == "step" {
			continue
		}
		// For multiple values, create a regex pattern
		if len(values) > 1 {
			filters[key] = fmt.Sprintf("(%s)", strings.Join(values, "|"))
		} else {
			filters[key] = values[0]
		}
	}
	return filters
}
