package deployments

import (
	"context"
	"errors"
	"testing"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestListDeployments(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx",
				Namespace: "default",
			},
		},
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "redis",
				Namespace: "default",
			},
		},
	)

	ctx := context.Background()
	deployments, err := ListDeployments(ctx, clientset)
	if err != nil {
		t.Fatalf("ListDeployments returned an error: %v", err)
	}

	if len(deployments) != 2 {
		t.Fatalf("Expected 2 deployments, got %d", len(deployments))
	}

	if !containsDeployment(deployments, "default", "nginx") {
		t.Errorf("Expected deployment 'nginx' in namespace 'default' not found")
	}

	if !containsDeployment(deployments, "default", "redis") {
		t.Errorf("Expected deployment 'redis' in namespace 'default' not found")
	}

}

func TestListDeploymentsEmpty(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	deployments, err := ListDeployments(
		context.Background(),
		clientset,
	)
	if err != nil {
		t.Fatalf("ListDeployments() returned an error: %v", err)
	}

	if len(deployments) != 0 {
		t.Fatalf(
			"ListDeployments() returned %d deployments, want 0",
			len(deployments),
		)
	}
}

func TestListDeploymentsError(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	expectedErr := errors.New("kubernetes API unavailable")

	clientset.PrependReactor(
		"list",
		"deployments",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		},
	)

	_, err := ListDeployments(
		context.Background(),
		clientset,
	)

	if err == nil {
		t.Fatal("ListDeployments() expected an error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf(
			"ListDeployments() error = %v, want %v",
			err,
			expectedErr,
		)
	}
}

func containsDeployment(deployments []v1.Deployment, namespace string, name string) bool {
	for _, deployment := range deployments {
		if deployment.Namespace == namespace &&
			deployment.Name == name {
			return true
		}
	}

	return false
}
