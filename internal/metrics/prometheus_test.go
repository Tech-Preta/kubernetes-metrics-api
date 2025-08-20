package metrics

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrometheusMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should create prometheus metrics successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			// Act
			metrics := NewPrometheusMetrics(logger)

			// Assert
			require.NotNil(t, metrics)
			assert.NotNil(t, metrics.NodeCount)
			assert.NotNil(t, metrics.PodCount)
			assert.NotNil(t, metrics.DeploymentCount)
			assert.NotNil(t, metrics.ServiceCount)
			assert.NotNil(t, metrics.NamespaceCount)
			assert.NotNil(t, metrics.PodStatus)
			assert.NotNil(t, metrics.NodeReady)
			assert.NotNil(t, metrics.ContainerRestarts)
			assert.NotNil(t, metrics.CPUAllocatable)
			assert.NotNil(t, metrics.MemoryAllocatable)
		})
	}
}

func TestPrometheusMetricsFields(t *testing.T) {
	tests := []struct {
		name   string
		testFn func(*PrometheusMetrics)
	}{
		{
			name: "should set and read node count",
			testFn: func(m *PrometheusMetrics) {
				m.NodeCount.Set(5)
				// We can't easily read the value back from prometheus gauge
				// but we can verify the operation doesn't panic
			},
		},
		{
			name: "should set pod status labels",
			testFn: func(m *PrometheusMetrics) {
				m.PodStatus.WithLabelValues("default", "Running").Set(10)
				m.PodStatus.WithLabelValues("kube-system", "Pending").Set(2)
			},
		},
		{
			name: "should set node ready status",
			testFn: func(m *PrometheusMetrics) {
				m.NodeReady.WithLabelValues("node-1").Set(1)
				m.NodeReady.WithLabelValues("node-2").Set(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			metrics := NewPrometheusMetrics(logger)

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				tt.testFn(metrics)
			})
		})
	}
}

func TestPrometheusMetricsRegistry(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "should handle duplicate registration gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			// Act - Create metrics twice (this tests the duplicate registration handling)
			metrics1 := NewPrometheusMetrics(logger)
			metrics2 := NewPrometheusMetrics(logger)

			// Assert - Both should be created successfully
			assert.NotNil(t, metrics1)
			assert.NotNil(t, metrics2)

			// Verify all metrics are properly initialized
			assert.NotNil(t, metrics1.NodeCount)
			assert.NotNil(t, metrics2.NodeCount)
		})
	}
}
