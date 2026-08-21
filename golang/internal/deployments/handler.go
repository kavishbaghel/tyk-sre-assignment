package deployments

import (
	"encoding/json"
	"net/http"

	"k8s.io/client-go/kubernetes"
)

func DeploymentHealthHandler(clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		deploymentList, err := ListDeployments(ctx, clientset)
		if err != nil {
			http.Error(w, "Failed to get deployment list", http.StatusServiceUnavailable)
			return
		}

		deploymentHealth := GetDeploymentHealth(deploymentList)

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(deploymentHealth); err != nil {
			return
		}
	}
}
