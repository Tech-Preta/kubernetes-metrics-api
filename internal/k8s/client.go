package k8s

import (
	"flag"
	"log/slog"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wrap do clientset Kubernetes.
type Client struct {
	Clientset kubernetes.Interface
}

// NewClient cria client in-cluster ou via kubeconfig.
func NewClient(logger *slog.Logger) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Debug("Falling back para kubeconfig local", "error", err)
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{Clientset: cs}, nil
}
