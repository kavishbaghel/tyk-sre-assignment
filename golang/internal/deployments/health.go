package deployments

import v1 "k8s.io/api/apps/v1"

func IsDeploymentHealthy(deployment *v1.Deployment) bool {
	if deployment == nil {
		return false
	}

	if deployment.Spec.Replicas == nil {
		return false
	}

	desiredReplicas := *deployment.Spec.Replicas
	readyReplicas := deployment.Status.ReadyReplicas

	observedGeneration := deployment.Status.ObservedGeneration
	generation := deployment.Generation

	return desiredReplicas <= readyReplicas && observedGeneration >= generation
}

func GetDeploymentHealth(deployments []v1.Deployment) DeploymentHealthResponse {
	healthResponse := DeploymentHealthResponse{
		Status:               "healthy",
		UnhealthyDeployments: []UnhealthyDeployment{},
	}

	for i := range deployments {
		deployment := &deployments[i]

		if IsDeploymentHealthy(deployment) {
			continue
		}

		healthResponse.Status = "unhealthy"

		reason := "Deployment is not healthy"

		if deployment.Spec.Replicas == nil {
			reason = "Deployment has no replicas defined"
		} else if deployment.Status.ObservedGeneration < deployment.Generation {
			reason = "Latest deployment generation has not been observed by the controller"
		} else if deployment.Status.AvailableReplicas < *deployment.Spec.Replicas {
			reason = "Deployment has fewer available replicas than desired"
		}

		healthResponse.UnhealthyDeployments = append(healthResponse.UnhealthyDeployments, UnhealthyDeployment{
			Name:      deployment.Name,
			Namespace: deployment.Namespace,
			Reason:    reason,
		})
	}

	return healthResponse
}
