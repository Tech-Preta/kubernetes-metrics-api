package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"k8s-metrics-api/docs"
	"k8s-metrics-api/internal/config"
	"k8s-metrics-api/internal/handlers"
	"k8s-metrics-api/internal/k8s"
	"k8s-metrics-api/internal/metrics"
	"k8s-metrics-api/internal/middleware"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		os.Exit(1)
	}
	cfg.Logger.Info("Iniciando a API de Métricas Kubernetes...")

	k8sClient, err := k8s.NewClient(cfg.Logger)
	if err != nil {
		cfg.Logger.Error("Erro ao inicializar cliente Kubernetes", "error", err)
		os.Exit(1)
	}

	promMetrics := metrics.NewPrometheusMetrics(cfg.Logger)
	h := handlers.New(k8sClient, promMetrics, cfg.Logger)

	authMw := middleware.AuthMiddleware(cfg.ExpectedAuthToken, cfg.Logger)
	logMw := middleware.LoggingMiddleware(cfg.Logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", authMw(h.MetricsJSONHandler))
	mux.Handle("/prometheus", authMw(func(w http.ResponseWriter, r *http.Request) { promhttp.Handler().ServeHTTP(w, r) }))
	mux.HandleFunc("/healthz", h.HealthCheckHandler)

	// Servir swagger.yaml estático
	mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		b, err := fs.ReadFile(docs.SwaggerFS, "swagger.yaml")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = w.Write(b)
	})

	// Página HTML simples com Swagger UI via CDN
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		// Tema padrão do Swagger UI (removido fundo escuro customizado). Oculta apenas a barra superior para mais área útil.
		_, _ = w.Write([]byte(`<!DOCTYPE html><html lang="pt-br"><head><meta charset="utf-8"/><title>K8s Metrics API Docs</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
<style>.topbar{display:none}</style></head><body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>window.onload=()=>{fetch('/swagger.yaml',{cache:'no-store'}).then(()=>{window.ui=SwaggerUIBundle({url:'/swagger.yaml',dom_id:'#swagger-ui',deepLinking:true,tryItOutEnabled:true});});};</script>
</body></html>`))
	})

	server := &http.Server{Addr: ":" + cfg.Port, Handler: logMw(mux)}

	cfg.Logger.Info("Servidor de métricas Kubernetes escutando...", "port", cfg.Port)
	cfg.Logger.Info("Endpoints disponíveis:",
		"metricsJSON", fmt.Sprintf("http://localhost:%s/metrics (protegido)", cfg.Port),
		"prometheusMetrics", fmt.Sprintf("http://localhost:%s/prometheus (protegido)", cfg.Port),
		"healthCheck", fmt.Sprintf("http://localhost:%s/healthz", cfg.Port),
		"swaggerSpec", fmt.Sprintf("http://localhost:%s/swagger.yaml", cfg.Port),
		"swaggerUI", fmt.Sprintf("http://localhost:%s/docs", cfg.Port),
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		cfg.Logger.Error("Erro ao iniciar o servidor HTTP", "error", err)
		os.Exit(1)
	}
}
