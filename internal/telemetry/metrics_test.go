package telemetry

import (
	"context"
	"testing"

	"github.com/aaronlmathis/gosight-server/internal/sys"
	"github.com/aaronlmathis/gosight-shared/model"
	colmetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// MockResourceDiscoverer implements the ResourceDiscoverer interface for testing
type MockResourceDiscoverer struct {
	CallCount           int
	LastPayloadReceived *model.MetricPayload
	PayloadToReturn     *model.MetricPayload
}

func (m *MockResourceDiscoverer) ProcessMetricPayload(payload *model.MetricPayload) *model.MetricPayload {
	m.CallCount++
	m.LastPayloadReceived = payload
	return m.PayloadToReturn
}

func (m *MockResourceDiscoverer) ProcessLogPayload(payload *model.LogPayload) *model.LogPayload {
	return payload
}

func (m *MockResourceDiscoverer) ProcessTracePayload(payload *model.TracePayload) *model.TracePayload {
	return payload
}

// TestMetricsHandler_Export_ResourceDiscoveryIntegration tests the complete pipeline
// from OTLP metrics ingestion through resource discovery enrichment to broadcast/storage
func TestMetricsHandler_Export_ResourceDiscoveryIntegration(t *testing.T) {
	// Create mock resource discoverer
	mockResourceDiscovery := &MockResourceDiscoverer{
		PayloadToReturn: &model.MetricPayload{
			Metrics: []model.Metric{
				{
					Name:  "test_metric",
					Value: 42.0,
				},
			},
			Meta: &model.Meta{
				ResourceID: "resource-123",
				Labels: map[string]string{
					"service.name": "test-service",
					"environment":  "production",
				},
			},
		},
	}

	// Create minimal system context for testing
	sysCtx := &sys.SystemContext{
		Ctx: context.Background(),
		Stores: &sys.StoreModule{
			Metrics: nil, // Null store is fine for this test
		},
		Tele: &sys.TelemetryModule{
			ResourceDiscovery: mockResourceDiscovery,
		},
	}

	// Create metrics handler
	handler := NewMetricsHandler(sysCtx)

	// Create test OTLP request with resource attributes
	req := &colmetricpb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricpb.ResourceMetrics{
			{
				Resource: &resourcepb.Resource{
					Attributes: []*commonpb.KeyValue{
						{
							Key:   "service.name",
							Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "test-service"}},
						},
					},
				},
				ScopeMetrics: []*metricpb.ScopeMetrics{
					{
						Metrics: []*metricpb.Metric{
							{
								Name: "test_metric",
								Data: &metricpb.Metric_Gauge{
									Gauge: &metricpb.Gauge{
										DataPoints: []*metricpb.NumberDataPoint{
											{
												Value: &metricpb.NumberDataPoint_AsDouble{AsDouble: 42.0},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Execute the Export method
	resp, err := handler.Export(context.Background(), req)

	// Verify the response
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify that resource discovery was called
	if mockResourceDiscovery.CallCount != 1 {
		t.Errorf("Expected ProcessMetricPayload to be called once, got: %d", mockResourceDiscovery.CallCount)
	}

	// Verify the payload was passed correctly
	if mockResourceDiscovery.LastPayloadReceived == nil {
		t.Fatal("Expected payload to be passed to resource discovery")
	}

	payload := mockResourceDiscovery.LastPayloadReceived
	if len(payload.Metrics) != 1 {
		t.Errorf("Expected 1 metric, got: %d", len(payload.Metrics))
	}

	if payload.Metrics[0].Name != "test_metric" {
		t.Errorf("Expected metric name 'test_metric', got: %s", payload.Metrics[0].Name)
	}

	if payload.Meta == nil || payload.Meta.Labels["service.name"] != "test-service" {
		t.Error("Expected service.name tag to be preserved")
	}
}
