package metrics

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics guarda referências às métricas registradas.
type PrometheusMetrics struct {
	NodeCount           prometheus.Gauge
	PodCount            prometheus.Gauge
	DeploymentCount     prometheus.Gauge
	ServiceCount        prometheus.Gauge
	NamespaceCount      prometheus.Gauge
	NodeReady           *prometheus.GaugeVec
	PodStatus           *prometheus.GaugeVec
	DeploymentDesired   *prometheus.GaugeVec
	DeploymentAvailable *prometheus.GaugeVec
	ContainerRestarts   *prometheus.GaugeVec
	CPUAllocatable      *prometheus.GaugeVec
	MemoryAllocatable   *prometheus.GaugeVec
	CPURequests         *prometheus.GaugeVec
	MemoryRequests      *prometheus.GaugeVec
	CPULimits           *prometheus.GaugeVec
	MemoryLimits        *prometheus.GaugeVec
}

// NewPrometheusMetrics cria e registra métricas.
func NewPrometheusMetrics(logger *slog.Logger) *PrometheusMetrics {
	m := &PrometheusMetrics{
		NodeCount:           prometheus.NewGauge(prometheus.GaugeOpts{Name: "k8s_nodes_total", Help: "Total de nós"}),
		PodCount:            prometheus.NewGauge(prometheus.GaugeOpts{Name: "k8s_pods_total", Help: "Total de pods"}),
		DeploymentCount:     prometheus.NewGauge(prometheus.GaugeOpts{Name: "k8s_deployments_total", Help: "Total de deployments"}),
		ServiceCount:        prometheus.NewGauge(prometheus.GaugeOpts{Name: "k8s_services_total", Help: "Total de services"}),
		NamespaceCount:      prometheus.NewGauge(prometheus.GaugeOpts{Name: "k8s_namespaces_total", Help: "Total de namespaces"}),
		NodeReady:           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_node_status_ready", Help: "1 se Ready"}, []string{"node"}),
		PodStatus:           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_pod_status_phase", Help: "Status por fase"}, []string{"namespace", "phase"}),
		DeploymentDesired:   prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_deployment_replicas_desired", Help: "Replicas desejadas"}, []string{"namespace", "deployment"}),
		DeploymentAvailable: prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_deployment_replicas_available", Help: "Replicas disponíveis"}, []string{"namespace", "deployment"}),
		ContainerRestarts:   prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_container_restarts_total", Help: "Restart count"}, []string{"namespace", "pod", "container"}),
		CPUAllocatable:      prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_node_cpu_allocatable_cores", Help: "CPU allocatable"}, []string{"node"}),
		MemoryAllocatable:   prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_node_memory_allocatable_bytes", Help: "Memória allocatable"}, []string{"node"}),
		CPURequests:         prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_namespace_cpu_requests_cores", Help: "Soma CPU requests"}, []string{"namespace"}),
		MemoryRequests:      prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_namespace_memory_requests_bytes", Help: "Soma memória requests"}, []string{"namespace"}),
		CPULimits:           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_namespace_cpu_limits_cores", Help: "Soma CPU limits"}, []string{"namespace"}),
		MemoryLimits:        prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "k8s_namespace_memory_limits_bytes", Help: "Soma memória limits"}, []string{"namespace"}),
	}

	collectors := []prometheus.Collector{
		m.NodeCount, m.PodCount, m.DeploymentCount, m.ServiceCount, m.NamespaceCount,
		m.NodeReady, m.PodStatus, m.DeploymentDesired, m.DeploymentAvailable,
		m.ContainerRestarts, m.CPUAllocatable, m.MemoryAllocatable,
		m.CPURequests, m.MemoryRequests, m.CPULimits, m.MemoryLimits,
	}
	for _, c := range collectors {
		_ = prometheus.Register(c) // ignora AlreadyRegistered
	}
	logger.Info("Métricas Prometheus registradas")
	return m
}
