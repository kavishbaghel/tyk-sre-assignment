package network

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestIsolateWorkloadsHandler(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	handler := NewHandler(clientset)

	requestBody := `{
		"source": {
			"namespaces": ["frontend"],
			"labelSelector": {
				"matchLabels": {
					"app": "frontend"
				}
			}
		},
		"destination": {
			"namespaces": ["backend"],
			"labelSelector": {
				"matchLabels": {
					"app": "backend"
				}
			}
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/network/isolate",
		strings.NewReader(requestBody),
	)

	rec := httptest.NewRecorder()

	handler.IsolateWorkloads(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}

	var response IsolationResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "isolated" {
		t.Errorf(
			"status = %q, want %q",
			response.Status,
			"isolated",
		)
	}

	if response.PoliciesApplied != 2 {
		t.Errorf(
			"policiesApplied = %d, want 2",
			response.PoliciesApplied,
		)
	}
}

func TestIsolateWorkloadsHandlerInvalidJSON(t *testing.T) {
	handler := NewHandler(fake.NewSimpleClientset())

	req := httptest.NewRequest(
		http.MethodPost,
		"/network/isolate",
		strings.NewReader(`{"invalid"`),
	)

	rec := httptest.NewRecorder()

	handler.IsolateWorkloads(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestIsolateWorkloadsHandlerInvalidConfig(t *testing.T) {
	handler := NewHandler(fake.NewSimpleClientset())

	requestBody := `{
		"source": {
			"namespaces": ["frontend"],
			"labelSelector": {}
		},
		"destination": {
			"namespaces": ["backend"],
			"labelSelector": {
				"matchLabels": {
					"app": "backend"
				}
			}
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/network/isolate",
		strings.NewReader(requestBody),
	)

	rec := httptest.NewRecorder()

	handler.IsolateWorkloads(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestIsolateWorkloadsHandlerKubernetesError(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	clientset.PrependReactor(
		"get",
		"networkpolicies",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("kubernetes API unavailable")
		},
	)

	handler := NewHandler(clientset)

	requestBody := `{
		"source": {
			"namespaces": ["frontend"],
			"labelSelector": {
				"matchLabels": {
					"app": "frontend"
				}
			}
		},
		"destination": {
			"namespaces": ["backend"],
			"labelSelector": {
				"matchLabels": {
					"app": "backend"
				}
			}
		}
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/network/isolate",
		strings.NewReader(requestBody),
	)

	rec := httptest.NewRecorder()

	handler.IsolateWorkloads(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}
}

func TestRevertIsolation(t *testing.T) {
	config := IsolationConfig{
		Source: WorkloadSelector{
			Namespaces: []string{"source"},
			LabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "frontend",
				},
			},
		},
		Destination: WorkloadSelector{
			Namespaces: []string{"destination"},
			LabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "backend",
				},
			},
		},
	}

	// Build the policies that would have been created by isolation.
	policies, err := BuildNetworkIsolationPolicies(config)
	if err != nil {
		t.Fatalf("BuildNetworkIsolationPolicies() error = %v", err)
	}

	// Put the policies into the fake Kubernetes cluster.
	clientset := fake.NewSimpleClientset()

	for i := range policies {
		_, err := clientset.NetworkingV1().NetworkPolicies(policies[i].Namespace).Create(
			context.Background(),
			&policies[i],
			metav1.CreateOptions{},
		)

		if err != nil {
			t.Fatalf("failed to create test policy %s/%s: %v", policies[i].Namespace, policies[i].Name, err)
		}
	}

	handler := NewHandler(clientset)

	body, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/network/revert", bytes.NewReader(body))

	rec := httptest.NewRecorder()

	handler.RevertIsolation(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var response RevertIsolationResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "isolation removed" {
		t.Errorf("response status = %q, want %q", response.Status, "isolation removed")
	}

	wantPoliciesRemoved := len(config.Source.Namespaces) + len(config.Destination.Namespaces)

	if response.PoliciesRemoved != wantPoliciesRemoved {
		t.Errorf("policies removed = %d, want %d", response.PoliciesRemoved, wantPoliciesRemoved)
	}

	if response.Message != "Network isolation policies reverted successfully" {
		t.Errorf("response message = %q, want %q", response.Message, "Network isolation policies reverted successfully")
	}

	// Verify that the policies were actually deleted.
	for _, policy := range policies {
		_, err := clientset.NetworkingV1().NetworkPolicies(policy.Namespace).Get(
			context.Background(),
			policy.Name,
			metav1.GetOptions{},
		)

		if !apierrors.IsNotFound(err) {
			t.Errorf("policy %s/%s still exists, get error = %v", policy.Namespace, policy.Name, err)
		}
	}
}
