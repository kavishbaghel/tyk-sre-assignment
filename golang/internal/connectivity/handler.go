package connectivity

import (
	"encoding/json"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

func NewConnectivityHandler(clientset kubernetes.Interface) *ConnectivityHandler {
	return &ConnectivityHandler{
		clientset: clientset,
	}
}

func (h *ConnectivityHandler) ToolConnectivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	discoveryClient := h.clientset.Discovery()
	_, err := discoveryClient.ServerVersion()
	if err != nil {
		response := connectivityResponse{
			Status:  "error",
			Message: "error connecting to Kubernetes API Server",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	response := connectivityResponse{
		Status:  "ok",
		Message: "Successfully connected to Kubernetes API Server",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
