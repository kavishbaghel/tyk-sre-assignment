package network

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ApplyNetworkIsolationPolicy(ctx context.Context, clientset kubernetes.Interface, policies []networkingv1.NetworkPolicy) error {
	for i := range policies {
		policy := &policies[i]
		policyClient := clientset.NetworkingV1().NetworkPolicies(policy.Namespace)

		existingPolicy, err := policyClient.Get(ctx, policy.Name, metav1.GetOptions{})
		if err != nil {
			// Create the policy if it don't exist
			if apierrors.IsNotFound(err) {
				_, err = policyClient.Create(ctx, policy, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("failed to create network policy %s/%s: %w", policy.Namespace, policy.Name, err)
				}
				continue
			}
			return fmt.Errorf("failed to get network policy %s/%s: %w", policy.Namespace, policy.Name, err)
		}

		policy.ResourceVersion = existingPolicy.ResourceVersion

		_, err = policyClient.Update(ctx, policy, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update network policy %s/%s: %w", policy.Namespace, policy.Name, err)
		}
	}
	return nil
}
