package network

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
