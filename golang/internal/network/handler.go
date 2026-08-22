package network

import (
	"encoding/json"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

func NewHandler(clientset kubernetes.Interface) *Handler {
	return &Handler{
		clientset: clientset,
	}
}

func (h *Handler) IsolateWorkloads(w http.ResponseWriter, r *http.Request) {
	var config IsolationConfig
	// Parse the request body into the IsolationConfig struct
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the isolation configuration
	if err := ValidateIsolationConfig(config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build the network policies based on the isolation configuration
	policies, err := BuildNetworkIsolationPolicies(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply the network policies to the Kubernetes cluster
	if err := ApplyNetworkIsolationPolicy(r.Context(), h.clientset, policies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := IsolationResponse{
		Status:          "isolated",
		PoliciesApplied: len(policies),
		Message:         "Network isolation policies applied successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
