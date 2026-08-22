package connectivity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestToolConnectivityHandlerSuccess(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	handler := NewConnectivityHandler(clientset)

	req := httptest.NewRequest(
		http.MethodGet,
		"/connectivity",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ToolConnectivityHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusOK,
		)
	}

	expectedBody := "{\"status\":\"ok\",\"message\":\"Successfully connected to Kubernetes API Server\"}\n"

	if rec.Body.String() != expectedBody {
		t.Errorf(
			"response body = %q, want %q",
			rec.Body.String(),
			expectedBody,
		)
	}

	expectedContentType := "application/json"

	if contentType := rec.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf(
			"Content-Type = %q, want %q",
			contentType,
			expectedContentType,
		)
	}
}

func TestToolConnectivityHandlerMethodNotAllowed(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	handler := NewConnectivityHandler(clientset)

	req := httptest.NewRequest(
		http.MethodPost,
		"/connectivity",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ToolConnectivityHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status code = %d, want %d",
			rec.Code,
			http.StatusMethodNotAllowed,
		)
	}

	expectedBody := "Method not allowed\n"

	if rec.Body.String() != expectedBody {
		t.Errorf(
			"response body = %q, want %q",
			rec.Body.String(),
			expectedBody,
		)
	}
}
