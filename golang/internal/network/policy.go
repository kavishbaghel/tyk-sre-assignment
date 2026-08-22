package network

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const isolationPolicyPrefix = "tyk-isolation-policy-"

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

func networkPolicyBuilder(namespace string, selector metav1.LabelSelector) networkingv1.NetworkPolicy {
	policyName, err := policyNameBuilder(namespace, selector)
	if err != nil {
		panic(fmt.Errorf("Failed to build policy name: %w", err))
	}
	return networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: namespace,
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
