package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"k8s-metrics-api/internal/k8s"
	"k8s-metrics-api/internal/metrics"
)

// newTestClient creates a k8s.Client with fake clientset for testing
func newTestClient(objects ...runtime.Object) *k8s.Client {
	fakeClientset := fake.NewSimpleClientset(objects...)
	return &k8s.Client{
		Clientset: fakeClientset,
	}
}

func TestHealthCheckHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "health check returns ok",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"status", "ts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			handler := &Handler{log: logger}

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			w := httptest.NewRecorder()

			// Act
			handler.HealthCheckHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field)
			}

			assert.Equal(t, "ok", response["status"])

			// Verify timestamp is valid
			ts, ok := response["ts"].(string)
			require.True(t, ok)
			_, err = time.Parse(time.RFC3339, ts)
			assert.NoError(t, err)
		})
	}
}

func TestMetricsJSONHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupObjects   func() []runtime.Object
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "successful metrics collection",
			setupObjects: func() []runtime.Object {
				return []runtime.Object{
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{Name: "node1"},
						Status: corev1.NodeStatus{
							Conditions: []corev1.NodeCondition{
								{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
							},
						},
					},
					&corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
						Status:     corev1.PodStatus{Phase: corev1.PodRunning},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "container1"}},
						},
					},
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"},
					},
					&corev1.Namespace{
						ObjectMeta: metav1.ObjectMeta{Name: "default"},
					},
				}
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"nodeCount", "podCount", "deploymentCount", "serviceCount", "namespaceCount", "podPhases", "timestamp"},
		},
		{
			name: "empty cluster",
			setupObjects: func() []runtime.Object {
				return []runtime.Object{}
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"nodeCount", "podCount", "deploymentCount", "serviceCount", "namespaceCount", "podPhases", "timestamp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			objects := tt.setupObjects()
			k8sClient := newTestClient(objects...)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			promMetrics := metrics.NewPrometheusMetrics(logger)

			handler := New(k8sClient, promMetrics, logger)

			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			w := httptest.NewRecorder()

			// Act
			handler.MetricsJSONHandler(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var response ClusterMetrics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Verify all expected fields are present
				assert.GreaterOrEqual(t, response.NodeCount, 0)
				assert.GreaterOrEqual(t, response.PodCount, 0)
				assert.GreaterOrEqual(t, response.DeploymentCount, 0)
				assert.GreaterOrEqual(t, response.ServiceCount, 0)
				assert.GreaterOrEqual(t, response.NamespaceCount, 0)
				assert.NotNil(t, response.PodPhases)
				assert.WithinDuration(t, time.Now(), response.Timestamp, 5*time.Second)

				// Verify specific counts based on test data
				if len(objects) > 0 {
					assert.Equal(t, 1, response.NodeCount)
					assert.Equal(t, 1, response.PodCount)
					assert.Equal(t, 1, response.ServiceCount)
					assert.Equal(t, 1, response.NamespaceCount)
					assert.Contains(t, response.PodPhases, "Running")
					assert.Equal(t, 1, response.PodPhases["Running"])
				}
			}
		})
	}
}
