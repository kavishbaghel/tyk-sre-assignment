package deployments

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListDeployments(ctx context.Context, clientset kubernetes.Interface) ([]v1.Deployment, error) {
	deploymentsList, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deploymentsList.Items, nil
}
