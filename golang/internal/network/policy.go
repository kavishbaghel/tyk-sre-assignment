package network

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	isolationPolicyPrefix = "tyk-isolation-policy-"
	managedByLabelKey     = "app.kubernetes.io/managed-by"
	managedByLabelValue   = "tyk-sre-app"

	isolationLabelKey   = "tyk-sre-app/isolation"
	isolationLabelValue = "true"
)

func BuildNetworkIsolationPolicies(config IsolationConfig) ([]networkingv1.NetworkPolicy, error) {
	if err := ValidateIsolationConfig(config); err != nil {
		return nil, err
	}

	policies := make([]networkingv1.NetworkPolicy, 0,
		len(config.Source.Namespaces)+len(config.Destination.Namespaces),
	)

	for _, namespace := range config.Source.Namespaces {
		policies = append(policies, networkPolicyBuilder(namespace, config.Source.LabelSelector))
	}
	for _, namespace := range config.Destination.Namespaces {
		policies = append(policies, networkPolicyBuilder(namespace, config.Destination.LabelSelector))
	}

	return policies, nil
}

func DeleteNetworkIsolationPolicies(ctx context.Context, clientset kubernetes.Interface, config IsolationConfig) error {
	if err := ValidateIsolationConfig(config); err != nil {
		return err
	}

	policies, err := BuildNetworkIsolationPolicies(config)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		policyClient := clientset.NetworkingV1().NetworkPolicies(policy.Namespace)

		existingPolicy, err := policyClient.Get(ctx, policy.Name, metav1.GetOptions{})

		if err != nil {
			// Revert is idempotent. If the policy is already gone,
			// there is nothing left to do.
			if apierrors.IsNotFound(err) {
				continue
			}

			return fmt.Errorf("failed to get network policy %s/%s: %w", policy.Namespace, policy.Name, err)
		}

		// Only delete policies created by this application.
		if existingPolicy.Labels[managedByLabelKey] != managedByLabelValue ||
			existingPolicy.Labels[isolationLabelKey] != isolationLabelValue {
			return fmt.Errorf("network policy %s/%s is not managed by this application", policy.Namespace, policy.Name)
		}

		if err := policyClient.Delete(ctx, policy.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("failed to delete network policy %s/%s: %w", policy.Namespace, policy.Name, err)
		}
	}

	return nil
}

func networkPolicyBuilder(namespace string, selector metav1.LabelSelector) networkingv1.NetworkPolicy {
	policyName, err := policyNameBuilder(namespace, selector)
	if err != nil {
		panic(fmt.Errorf("Failed to build policy name: %w", err))
	}
	return networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: namespace,
			Labels: map[string]string{
				managedByLabelKey: managedByLabelValue,
				isolationLabelKey: isolationLabelValue,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: selector,
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}
}

func policyNameBuilder(namespace string, selector metav1.LabelSelector) (string, error) {
	id := struct {
		Namespace string               `json:"namespace"`
		Selector  metav1.LabelSelector `json:"selector"`
	}{
		Namespace: namespace,
		Selector:  selector,
	}

	data, err := json.Marshal(id)
	if err != nil {
		return "", fmt.Errorf("Failed to serialize policy id: %w", err)
	}

	policyHash := md5.Sum(data)
	policyString := hex.EncodeToString(policyHash[:])
	return isolationPolicyPrefix + policyString, nil
}
