package deployments

type UnhealthyDeployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Reason    string `json:"reason"`
}

type DeploymentHealthResponse struct {
	Status               string                `json:"status"`
	UnhealthyDeployments []UnhealthyDeployment `json:"unhealthy_deployments"`
}
