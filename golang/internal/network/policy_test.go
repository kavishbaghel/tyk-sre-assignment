package network

import (
	"reflect"
	"strings"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildIsolationPolicies(t *testing.T) {
	sourceSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "frontend",
		},
	}

	destinationSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "backend",
		},
	}

	config := IsolationConfig{
		Source: WorkloadSelector{
			Namespaces:    []string{"frontend"},
			LabelSelector: sourceSelector,
		},
		Destination: WorkloadSelector{
			Namespaces:    []string{"backend"},
			LabelSelector: destinationSelector,
		},
	}

	policies, err := BuildNetworkIsolationPolicies(config)
	if err != nil {
		t.Fatalf("BuildNetworkIsolationPolicies() error = %v", err)
	}

	if len(policies) != 2 {
		t.Fatalf("got %d policies, want 2", len(policies))
	}

	tests := []struct {
		name      string
		policy    networkingv1.NetworkPolicy
		namespace string
		selector  metav1.LabelSelector
	}{
		{
			name:      "source policy",
			policy:    policies[0],
			namespace: "frontend",
			selector:  sourceSelector,
		},
		{
			name:      "destination policy",
			policy:    policies[1],
			namespace: "backend",
			selector:  destinationSelector,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.policy.Namespace != tt.namespace {
				t.Errorf(
					"namespace = %q, want %q",
					tt.policy.Namespace,
					tt.namespace,
				)
			}

			if !reflect.DeepEqual(
				tt.policy.Spec.PodSelector,
				tt.selector,
			) {
				t.Errorf(
					"PodSelector = %#v, want %#v",
					tt.policy.Spec.PodSelector,
					tt.selector,
				)
			}

			if !reflect.DeepEqual(
				tt.policy.Spec.PolicyTypes,
				[]networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			) {
				t.Errorf(
					"PolicyTypes = %#v, want [Ingress]",
					tt.policy.Spec.PolicyTypes,
				)
			}

			if len(tt.policy.Spec.Ingress) != 0 {
				t.Errorf(
					"Ingress rules = %d, want 0",
					len(tt.policy.Spec.Ingress),
				)
			}
		})
	}
}

func TestPolicyNameBuilder(t *testing.T) {
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "backend",
		},
	}

	name1, err := policyNameBuilder("backend", selector)
	if err != nil {
		t.Fatalf("policyNameBuilder() error = %v", err)
	}

	name2, err := policyNameBuilder("backend", selector)
	if err != nil {
		t.Fatalf("policyNameBuilder() error = %v", err)
	}

	if name1 != name2 {
		t.Errorf(
			"expected deterministic name, got %q and %q",
			name1,
			name2,
		)
	}

	if !strings.HasPrefix(name1, isolationPolicyPrefix) {
		t.Errorf(
			"name = %q, does not have prefix %q",
			name1,
			isolationPolicyPrefix,
		)
	}

	hashPart := strings.TrimPrefix(name1, isolationPolicyPrefix)

	if len(hashPart) != 32 {
		t.Errorf(
			"hash length = %d, want 32",
			len(hashPart),
		)
	}
}

func TestPolicyNameBuilderDifferentSelectors(t *testing.T) {
	backendSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "backend",
		},
	}

	paymentsSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "payments",
		},
	}

	backendName, err := policyNameBuilder(
		"production",
		backendSelector,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paymentsName, err := policyNameBuilder(
		"production",
		paymentsSelector,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backendName == paymentsName {
		t.Errorf(
			"different selectors generated same policy name: %q",
			backendName,
		)
	}
}

func TestPolicyNameBuilderDifferentNamespaces(t *testing.T) {
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "backend",
		},
	}

	backendName, err := policyNameBuilder(
		"backend",
		selector,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	productionName, err := policyNameBuilder(
		"production",
		selector,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if backendName == productionName {
		t.Errorf(
			"different namespaces generated same policy name: %q",
			backendName,
		)
	}
}

func TestBuildIsolationPoliciesMultipleNamespaces(t *testing.T) {
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "frontend",
		},
	}

	config := IsolationConfig{
		Source: WorkloadSelector{
			Namespaces:    []string{"frontend", "frontend-staging"},
			LabelSelector: selector,
		},
		Destination: WorkloadSelector{
			Namespaces: []string{"backend", "backend-staging"},
			LabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "backend",
				},
			},
		},
	}

	policies, err := BuildNetworkIsolationPolicies(config)
	if err != nil {
		t.Fatalf("BuildNetworkIsolationPolicies() error = %v", err)
	}

	if len(policies) != 4 {
		t.Fatalf(
			"got %d policies, want 4",
			len(policies),
		)
	}

	expectedNamespaces := []string{
		"frontend",
		"frontend-staging",
		"backend",
		"backend-staging",
	}

	for i, namespace := range expectedNamespaces {
		if policies[i].Namespace != namespace {
			t.Errorf(
				"policy[%d].Namespace = %q, want %q",
				i,
				policies[i].Namespace,
				namespace,
			)
		}
	}
}

func TestBuildIsolationPoliciesInvalidConfig(t *testing.T) {
	config := IsolationConfig{
		Source: WorkloadSelector{
			Namespaces: []string{"frontend"},
			// Empty selector.
			LabelSelector: metav1.LabelSelector{},
		},
		Destination: WorkloadSelector{
			Namespaces: []string{"backend"},
			LabelSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "backend",
				},
			},
		},
	}

	policies, err := BuildNetworkIsolationPolicies(config)

	if err == nil {
		t.Fatal("BuildNetworkIsolationPolicies() expected error, got nil")
	}

	if policies != nil {
		t.Errorf(
			"policies = %#v, want nil",
			policies,
		)
	}
}
