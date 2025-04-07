package kubernetesClient

import (
	"flag"
	"os"
	"path/filepath"

	"imesh.ai/ratelimit-cm-watcher/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig *string
)

func getClientConfig(development bool) (*rest.Config, error) {
	if development {
		logger.L.Debug("K8s client type as in cluster")
		return rest.InClusterConfig()
	}

	if kubeconfig == nil {
		kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "path to the kubeconfig file")
	}

	logger.L.Debug("K8s client type as in Local")

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func GetClient() *kubernetes.Clientset {
	config, err := getClientConfig(false)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
