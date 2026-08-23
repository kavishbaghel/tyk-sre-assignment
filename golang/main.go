package main

import (
	"flag"
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"

	"github.com/TykTechnologies/tyk-sre-assignment/internal/connectivity"
	"github.com/TykTechnologies/tyk-sre-assignment/internal/deployments"
	"github.com/TykTechnologies/tyk-sre-assignment/internal/network"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig, leave empty for in-cluster")
	listenAddr := flag.String("address", ":8080", "HTTP server listen address")

	flag.Parse()

	kConfig, err := buildKubernetesConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(kConfig)
	if err != nil {
		panic(err)
	}

	version, err := getKubernetesVersion(clientset)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to Kubernetes %s\n", version)

	if err := startServer(*listenAddr, clientset); err != nil {
		panic(err)
	}
}

// getKubernetesVersion returns a string GitVersion of the Kubernetes server defined by the clientset.
//
// If it can't connect an error will be returned, which makes it useful to check connectivity.
func getKubernetesVersion(clientset kubernetes.Interface) (string, error) {
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}

	return version.String(), nil
}

// startServer launches an HTTP server with defined handlers and blocks until it's terminated or fails with an error.
//
// Expects a listenAddr to bind to.
func startServer(listenAddr string, clientset kubernetes.Interface) error {

	http.HandleFunc("/healthz", healthHandler)
	http.HandleFunc("/deployments/health", deployments.DeploymentHealthHandler(clientset))
	http.HandleFunc("/network/isolate", network.NewHandler(clientset).IsolateWorkloads)
	http.HandleFunc("/network/revert", network.NewHandler(clientset).RevertIsolation)
	http.HandleFunc("/connectivity", connectivity.NewConnectivityHandler(clientset).ToolConnectivityHandler)

	fmt.Printf("Server listening on %s\n", listenAddr)

	return http.ListenAndServe(listenAddr, nil)
}

// healthHandler responds with the health status of the application.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("ok"))
	if err != nil {
		fmt.Println("failed writing to response")
	}
}

// buildKubernetesConfig builds a Kubernetes client configuration based on the provided kubeconfig path.
func buildKubernetesConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
}
