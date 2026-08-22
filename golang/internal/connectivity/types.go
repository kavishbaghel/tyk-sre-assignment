package connectivity

import (
	"k8s.io/client-go/kubernetes"
)

type ConnectivityHandler struct {
	clientset kubernetes.Interface
}

type connectivityResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
