package network

import (
	"context"
	"testing"

	"errors"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestApplyIsolationPoliciesCreatesPolicy(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	policy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "backend",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "backend",
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
		},
	}

	err := ApplyNetworkIsolationPolicy(
		context.Background(),
		clientset,
		[]networkingv1.NetworkPolicy{policy},
	)

	if err != nil {
		t.Fatalf("ApplyNetworkIsolationPolicy() error = %v", err)
	}

	created, err := clientset.
		NetworkingV1().
		NetworkPolicies("backend").
		Get(
			context.Background(),
			"test-policy",
			metav1.GetOptions{},
		)

	if err != nil {
		t.Fatalf("failed to get created policy: %v", err)
	}

	if created.Name != "test-policy" {
		t.Errorf("Name = %q, want %q", created.Name, "test-policy")
	}
}

func TestApplyIsolationPoliciesUpdatesPolicy(t *testing.T) {
	existing := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-policy",
			Namespace:       "backend",
			ResourceVersion: "1",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "old",
				},
			},
		},
	}

	clientset := fake.NewSimpleClientset(existing)

	desired := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "backend",
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "backend",
				},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{},
		},
	}

	err := ApplyNetworkIsolationPolicy(
		context.Background(),
		clientset,
		[]networkingv1.NetworkPolicy{desired},
	)

	if err != nil {
		t.Fatalf("ApplyNetworkIsolationPolicy() error = %v", err)
	}

	updated, err := clientset.
		NetworkingV1().
		NetworkPolicies("backend").
		Get(
			context.Background(),
			"test-policy",
			metav1.GetOptions{},
		)

	if err != nil {
		t.Fatalf("failed to get updated policy: %v", err)
	}

	if updated.Spec.PodSelector.MatchLabels["app"] != "backend" {
		t.Errorf(
			"PodSelector app = %q, want %q",
			updated.Spec.PodSelector.MatchLabels["app"],
			"backend",
		)
	}
}

func TestApplyIsolationPoliciesGetError(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	expectedErr := errors.New("kubernetes API unavailable")

	clientset.PrependReactor(
		"get",
		"networkpolicies",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		},
	)

	policy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-policy",
			Namespace: "backend",
		},
	}

	err := ApplyNetworkIsolationPolicy(
		context.Background(),
		clientset,
		[]networkingv1.NetworkPolicy{policy},
	)

	if err == nil {
		t.Fatal("ApplyNetworkIsolationPolicy() expected error, got nil")
	}
}
