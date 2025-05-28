package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog" // Novo: Logging estruturado
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Pacotes do Kubernetes client-go

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // Corrigido: Alias correto
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// Para métricas Prometheus
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// Para métricas mais detalhadas, você pode precisar do metrics server client
	// metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// --- Variáveis Globais e Configuração ---
var (
	clientset *kubernetes.Clientset
	logger    *slog.Logger
	// Token de autenticação esperado (lido da variável de ambiente EXPECTED_AUTH_TOKEN)
	expectedAuthToken string

	// Métricas Prometheus
	nodeCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_metrics_api_nodes_total",
			Help: "Total number of nodes in the cluster.",
		},
	)
	podCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_metrics_api_pods_total",
			Help: "Total number of pods in the cluster.",
		},
	)
	deploymentCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_metrics_api_deployments_total",
			Help: "Total number of deployments in the cluster.",
		},
	)
	serviceCountGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "k8s_metrics_api_services_total",
			Help: "Total number of services in the cluster.",
		},
	)
	nodeStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "k8s_metrics_api_node_status_ready",
			Help: "Status of nodes (1 if Ready, 0 if NotReady).",
		},
		[]string{"node_name", "status"}, // Labels para o nome do nó e seu status
	)
)

// ClusterMetrics define a estrutura para as métricas que queremos expor em JSON.
type ClusterMetrics struct {
	NodeCount       int        `json:"nodeCount"`
	PodCount        int        `json:"podCount"`
	DeploymentCount int        `json:"deploymentCount"`
	ServiceCount    int        `json:"serviceCount"`
	Nodes           []NodeInfo `json:"nodes"`
	Timestamp       time.Time  `json:"timestamp"`
}

// NodeInfo contém informações básicas sobre um nó para a resposta JSON.
type NodeInfo struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"` // Ex: Ready, NotReady
	AllocatableCPU    string            `json:"allocatableCpu"`
	AllocatableMemory string            `json:"allocatableMemory"`
	KubeletVersion    string            `json:"kubeletVersion"`
	OSImage           string            `json:"osImage"`
	Labels            map[string]string `json:"labels"`
}

// --- Inicialização ---

func initLogger() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger) // Define o logger padrão para o pacote slog
}

func initPrometheus() {
	prometheus.MustRegister(nodeCountGauge)
	prometheus.MustRegister(podCountGauge)
	prometheus.MustRegister(deploymentCountGauge)
	prometheus.MustRegister(serviceCountGauge)
	prometheus.MustRegister(nodeStatusGauge)
	logger.Info("Métricas Prometheus registradas.")
}

func initKubeClient() {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		logger.Warn("Não foi possível carregar a configuração InCluster, tentando kubeconfig local.", "error", err.Error())
		var kubeconfig *string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(opcional) caminho absoluto para o arquivo kubeconfig")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "caminho absoluto para o arquivo kubeconfig")
		}
		flag.Parse()

		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			logger.Error("Erro ao construir config a partir do kubeconfig.", "error", err.Error())
			os.Exit(1) // Sai se não conseguir configurar o cliente
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("Erro ao criar o clientset Kubernetes.", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("Clientset Kubernetes inicializado com sucesso.")
}

// --- Middlewares ---

// authMiddleware verifica o token de autenticação.
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			logger.Warn("Tentativa de acesso não autorizado: token não fornecido.", "remoteAddr", r.RemoteAddr, "path", r.URL.Path)
			http.Error(w, "Acesso não autorizado: token não fornecido", http.StatusUnauthorized)
			return
		}

		// Espera "Bearer <token>"
		parts := strings.Split(token, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.Warn("Tentativa de acesso não autorizado: formato de token inválido.", "remoteAddr", r.RemoteAddr, "path", r.URL.Path)
			http.Error(w, "Acesso não autorizado: formato de token inválido", http.StatusUnauthorized)
			return
		}
		actualToken := parts[1]

		if actualToken != expectedAuthToken {
			logger.Warn("Tentativa de acesso não autorizado: token inválido.", "remoteAddr", r.RemoteAddr, "path", r.URL.Path)
			http.Error(w, "Acesso não autorizado: token inválido", http.StatusUnauthorized)
			return
		}

		logger.Debug("Acesso autorizado.", "remoteAddr", r.RemoteAddr, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

// loggingMiddleware registra informações sobre cada requisição.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Cria um ResponseWriter customizado para capturar o status code
		rw := &responseWriterDelegator{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r) // Chama o próximo handler

		logger.Info("Requisição HTTP",
			"method", r.Method,
			"path", r.URL.Path,
			"remoteAddr", r.RemoteAddr,
			"userAgent", r.UserAgent(),
			"statusCode", rw.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}

// responseWriterDelegator é usado para capturar o status code da resposta.
type responseWriterDelegator struct {
	http.ResponseWriter
	statusCode int
}

func (rwd *responseWriterDelegator) WriteHeader(code int) {
	rwd.statusCode = code
	rwd.ResponseWriter.WriteHeader(code)
}

// --- Handlers HTTP ---

func metricsJSONHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	opLogger := logger.With("handler", "metricsJSONHandler")

	opLogger.Info("Coletando métricas do cluster para JSON...")

	// Coleta de informações dos Nós
	nodesList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		opLogger.Error("Erro ao listar nós.", "error", err)
		http.Error(w, fmt.Sprintf("Erro ao listar nós: %v", err), http.StatusInternalServerError)
		return
	}
	nodeCountGauge.Set(float64(len(nodesList.Items))) // Atualiza métrica Prometheus

	var nodeInfos []NodeInfo
	nodeStatusGauge.Reset() // Limpa os status antigos antes de definir os novos
	for _, node := range nodesList.Items {
		status := "NotReady"
		isReady := 0.0
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				status = "Ready"
				isReady = 1.0
				break
			}
		}
		nodeInfos = append(nodeInfos, NodeInfo{
			Name:              node.Name,
			Status:            status,
			AllocatableCPU:    node.Status.Allocatable.Cpu().String(),
			AllocatableMemory: node.Status.Allocatable.Memory().String(),
			KubeletVersion:    node.Status.NodeInfo.KubeletVersion,
			OSImage:           node.Status.NodeInfo.OSImage,
			Labels:            node.Labels,
		})
		nodeStatusGauge.WithLabelValues(node.Name, status).Set(isReady) // Atualiza métrica Prometheus
	}

	// Coleta de informações dos Pods
	podsList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		opLogger.Error("Erro ao listar pods.", "error", err)
		http.Error(w, fmt.Sprintf("Erro ao listar pods: %v", err), http.StatusInternalServerError)
		return
	}
	podCountGauge.Set(float64(len(podsList.Items))) // Atualiza métrica Prometheus

	// Coleta de informações dos Deployments
	deploymentsList, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		opLogger.Error("Erro ao listar deployments.", "error", err)
		http.Error(w, fmt.Sprintf("Erro ao listar deployments: %v", err), http.StatusInternalServerError)
		return
	}
	deploymentCountGauge.Set(float64(len(deploymentsList.Items))) // Atualiza métrica Prometheus

	// Coleta de informações dos Services
	servicesList, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		opLogger.Error("Erro ao listar services.", "error", err)
		http.Error(w, fmt.Sprintf("Erro ao listar services: %v", err), http.StatusInternalServerError)
		return
	}
	serviceCountGauge.Set(float64(len(servicesList.Items))) // Atualiza métrica Prometheus

	metrics := ClusterMetrics{
		NodeCount:       len(nodesList.Items),
		PodCount:        len(podsList.Items),
		DeploymentCount: len(deploymentsList.Items),
		ServiceCount:    len(servicesList.Items),
		Nodes:           nodeInfos,
		Timestamp:       time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		opLogger.Error("Erro ao codificar métricas para JSON.", "error", err)
		http.Error(w, fmt.Sprintf("Erro ao codificar métricas para JSON: %v", err), http.StatusInternalServerError)
	}
	opLogger.Info("Métricas JSON enviadas com sucesso.")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "timestamp": time.Now().UTC().Format(time.RFC3339)}); err != nil {
		logger.Error("Erro ao codificar resposta de health check.", "error", err)
		// Não podemos enviar http.Error aqui porque o header já foi escrito
	}
}

// --- Função Principal ---

func main() {
	// Inicializa o logger primeiro
	initLogger()
	logger.Info("Iniciando a API de Métricas Kubernetes...")

	// Configura o token de autenticação a partir da variável de ambiente
	expectedAuthToken = os.Getenv("EXPECTED_AUTH_TOKEN")
	if expectedAuthToken == "" {
		logger.Error("Variável de ambiente EXPECTED_AUTH_TOKEN não definida. Defina um token de autenticação.")
		os.Exit(1)
	}
	// Remove qualquer quebra de linha ou espaço em branco do token
	expectedAuthToken = strings.TrimSpace(expectedAuthToken)
	logger.Info("Token de autenticação configurado com sucesso.")

	// Inicializa o cliente Kubernetes
	initKubeClient()

	// Inicializa e registra as métricas Prometheus
	initPrometheus()

	// Define os handlers HTTP
	mux := http.NewServeMux()

	// Endpoint de métricas JSON (protegido)
	mux.HandleFunc("/metrics", authMiddleware(metricsJSONHandler))

	// Endpoint de métricas Prometheus (protegido)
	// O promhttp.Handler() já é um http.Handler, então o adaptamos para http.HandlerFunc se necessário ou usamos http.Handle
	// Para aplicar middleware a um http.Handler, podemos envolvê-lo.
	prometheusHandler := promhttp.Handler()
	mux.Handle("/metrics-prometheus", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Atualiza as métricas K8s antes de servir as métricas Prometheus
		// Isso garante que os gauges Prometheus estejam com os valores mais recentes.
		// Em um cenário de alta performance, você pode querer atualizar isso em um loop separado.
		go func() { // Executa em uma goroutine para não bloquear o handler
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout para coleta
			defer cancel()

			nodesList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err == nil {
				nodeCountGauge.Set(float64(len(nodesList.Items)))
				nodeStatusGauge.Reset()
				for _, node := range nodesList.Items {
					status := "NotReady"
					isReady := 0.0
					for _, cond := range node.Status.Conditions {
						if cond.Type == "Ready" && cond.Status == "True" {
							status = "Ready"
							isReady = 1.0
							break
						}
					}
					nodeStatusGauge.WithLabelValues(node.Name, status).Set(isReady)
				}
			} else {
				logger.Error("Erro ao atualizar contagem de nós para Prometheus.", "error", err)
			}

			podsList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
			if err == nil {
				podCountGauge.Set(float64(len(podsList.Items)))
			} else {
				logger.Error("Erro ao atualizar contagem de pods para Prometheus.", "error", err)
			}

			deploymentsList, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
			if err == nil {
				deploymentCountGauge.Set(float64(len(deploymentsList.Items)))
			} else {
				logger.Error("Erro ao atualizar contagem de deployments para Prometheus.", "error", err)
			}

			servicesList, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
			if err == nil {
				serviceCountGauge.Set(float64(len(servicesList.Items)))
			} else {
				logger.Error("Erro ao atualizar contagem de services para Prometheus.", "error", err)
			}
			logger.Debug("Métricas K8s atualizadas para o handler Prometheus.")
		}()

		prometheusHandler.ServeHTTP(w, r)
	}))

	// Endpoint de health check (não protegido)
	mux.HandleFunc("/healthz", healthCheckHandler)

	// Configura a porta do servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Adiciona o middleware de logging a todas as rotas
	loggedMux := loggingMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggedMux, // Usa o mux com logging
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	logger.Info("Servidor de métricas Kubernetes escutando...", "port", port)
	logger.Info("Endpoints disponíveis:",
		"metricsJSON", fmt.Sprintf("http://localhost:%s/metrics (protegido)", port),
		"metricsPrometheus", fmt.Sprintf("http://localhost:%s/metrics-prometheus (protegido)", port),
		"healthCheck", fmt.Sprintf("http://localhost:%s/healthz", port),
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Erro ao iniciar o servidor HTTP.", "error", err)
		os.Exit(1)
	}
}
