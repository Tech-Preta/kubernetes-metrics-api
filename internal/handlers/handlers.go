package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s-metrics-api/internal/k8s"
	"k8s-metrics-api/internal/metrics"
)

// Handler agrega dependências.
type Handler struct {
	k8s *k8s.Client
	m   *metrics.PrometheusMetrics
	log *slog.Logger
}

// New cria Handler.
func New(k8sClient *k8s.Client, m *metrics.PrometheusMetrics, logger *slog.Logger) *Handler {
	return &Handler{k8s: k8sClient, m: m, log: logger}
}

// ClusterMetrics resposta JSON.
type ClusterMetrics struct {
	NodeCount       int            `json:"nodeCount"`
	PodCount        int            `json:"podCount"`
	DeploymentCount int            `json:"deploymentCount"`
	ServiceCount    int            `json:"serviceCount"`
	NamespaceCount  int            `json:"namespaceCount"`
	PodPhases       map[string]int `json:"podPhases"`
	Timestamp       time.Time      `json:"timestamp"`
}

// MetricsJSONHandler coleta e retorna métricas.
func (h *Handler) MetricsJSONHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	cs := h.k8s.Clientset

	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.m.NodeCount.Set(float64(len(nodes.Items)))
	h.m.NodeReady.Reset()
	for _, n := range nodes.Items {
		ready := false
		for _, c := range n.Status.Conditions {
			if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
		h.m.NodeReady.WithLabelValues(n.Name).Set(boolToFloat(ready))
		h.m.CPUAllocatable.WithLabelValues(n.Name).Set(float64(n.Status.Allocatable.Cpu().MilliValue()) / 1000)
		h.m.MemoryAllocatable.WithLabelValues(n.Name).Set(float64(n.Status.Allocatable.Memory().Value()))
	}

	pods, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.m.PodCount.Set(float64(len(pods.Items)))
	h.m.PodStatus.Reset()
	h.m.ContainerRestarts.Reset()
	podPhases := map[string]int{}
	cpuReq := map[string]float64{}
	memReq := map[string]float64{}
	cpuLim := map[string]float64{}
	memLim := map[string]float64{}
	for _, p := range pods.Items {
		phase := string(p.Status.Phase)
		podPhases[phase]++
		h.m.PodStatus.WithLabelValues(p.Namespace, phase).Inc()
		for _, cs := range p.Status.ContainerStatuses {
			h.m.ContainerRestarts.WithLabelValues(p.Namespace, p.Name, cs.Name).Set(float64(cs.RestartCount))
		}
		for _, c := range p.Spec.Containers {
			if q, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuReq[p.Namespace] += float64(q.MilliValue()) / 1000
			}
			if q, ok := c.Resources.Requests[corev1.ResourceMemory]; ok {
				memReq[p.Namespace] += float64(q.Value())
			}
			if q, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
				cpuLim[p.Namespace] += float64(q.MilliValue()) / 1000
			}
			if q, ok := c.Resources.Limits[corev1.ResourceMemory]; ok {
				memLim[p.Namespace] += float64(q.Value())
			}
		}
	}
	h.m.CPURequests.Reset()
	h.m.MemoryRequests.Reset()
	h.m.CPULimits.Reset()
	h.m.MemoryLimits.Reset()
	for ns, v := range cpuReq {
		h.m.CPURequests.WithLabelValues(ns).Set(v)
	}
	for ns, v := range memReq {
		h.m.MemoryRequests.WithLabelValues(ns).Set(v)
	}
	for ns, v := range cpuLim {
		h.m.CPULimits.WithLabelValues(ns).Set(v)
	}
	for ns, v := range memLim {
		h.m.MemoryLimits.WithLabelValues(ns).Set(v)
	}

	namespaces, err := cs.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.m.NamespaceCount.Set(float64(len(namespaces.Items)))

	deployments, err := cs.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.m.DeploymentCount.Set(float64(len(deployments.Items)))
	h.m.DeploymentDesired.Reset()
	h.m.DeploymentAvailable.Reset()
	for _, d := range deployments.Items {
		h.m.DeploymentDesired.WithLabelValues(d.Namespace, d.Name).Set(float64(*d.Spec.Replicas))
		h.m.DeploymentAvailable.WithLabelValues(d.Namespace, d.Name).Set(float64(d.Status.AvailableReplicas))
	}

	services, err := cs.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.m.ServiceCount.Set(float64(len(services.Items)))

	resp := ClusterMetrics{NodeCount: len(nodes.Items), PodCount: len(pods.Items), DeploymentCount: len(deployments.Items), ServiceCount: len(services.Items), NamespaceCount: len(namespaces.Items), PodPhases: podPhases, Timestamp: time.Now().UTC()}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// HealthCheckHandler simples.
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "ts": time.Now().UTC().Format(time.RFC3339)})
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
