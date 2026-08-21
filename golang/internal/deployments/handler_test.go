package deployments

import (
	"encoding/json"
	"errors"
	"testing"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"net/http"
	"net/http/httptest"
)

func TestDeploymentHealthHandlerHealthy(t *testing.T) {
	desiredReplicas := int32(3)

	clientset := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "nginx",
				Namespace:  "default",
				Generation: 1,
			},
			Spec: v1.DeploymentSpec{
				Replicas: &desiredReplicas,
			},
			Status: v1.DeploymentStatus{
				ReadyReplicas:      3,
				ObservedGeneration: 1,
			},
		},
	)

	req := httptest.NewRequest("GET", "/deployments/health", nil)

	rec := httptest.NewRecorder()

	// Call the handler function
	h := DeploymentHealthHandler(clientset)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Expected Content-Type application/json, got %s", contentType)
	}

	var healthResponse DeploymentHealthResponse

	if err := json.NewDecoder(rec.Body).Decode(&healthResponse); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if healthResponse.Status != "healthy" {
		t.Fatalf("Expected status 'healthy', got '%s'", healthResponse.Status)
	}

	if len(healthResponse.UnhealthyDeployments) != 0 {
		t.Fatalf("Expected no unhealthy deployments, got %d", len(healthResponse.UnhealthyDeployments))
	}
}

func TestDeploymentHealthHandlerUnhealthy(t *testing.T) {
	replicas := int32(3)

	clientset := fake.NewSimpleClientset(
		&v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "nginx",
				Namespace:  "default",
				Generation: 1,
			},
			Spec: v1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: v1.DeploymentStatus{
				ReadyReplicas:      2,
				ObservedGeneration: 1,
			},
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/deployments/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler := DeploymentHealthHandler(clientset)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"Expectedstatus code 200, got %d", rec.Code)
	}

	var response DeploymentHealthResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "unhealthy" {
		t.Errorf(
			"Expected status 'unhealthy', got '%s'", response.Status)
	}

	if len(response.UnhealthyDeployments) != 1 {
		t.Fatalf(
			"Expected unhealthy deployments 1, got %d", len(response.UnhealthyDeployments))
	}

	deployment := response.UnhealthyDeployments[0]

	if deployment.Namespace != "default" {
		t.Errorf(
			"Expected namespace default, want %q", deployment.Namespace)
	}

	if deployment.Name != "nginx" {
		t.Errorf(
			"Expected name nginx, want %q", deployment.Name)
	}
}

func TestDeploymentHealthHandlerKubernetesError(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	expectedErr := errors.New("kubernetes API unavailable")

	clientset.PrependReactor(
		"list",
		"deployments",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/deployments/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler := DeploymentHealthHandler(clientset)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"Expected status code %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}
