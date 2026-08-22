package network

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateIsolationConfig(config IsolationConfig) error {
	if len(config.Source.Namespaces) == 0 {
		return fmt.Errorf("source namespaces cannot be empty")
	}

	if isLabelSelectorEmpty(config.Source.LabelSelector) {
		return fmt.Errorf("source label selector cannot be empty")
	}

	if len(config.Destination.Namespaces) == 0 {
		return fmt.Errorf("destination namespaces cannot be empty")
	}

	if isLabelSelectorEmpty(config.Destination.LabelSelector) {
		return fmt.Errorf("destination label selector cannot be empty")
	}

	for _, namespace := range config.Source.Namespaces {
		if namespace == "" {
			return fmt.Errorf("source namespace cannot be empty")
		}
	}

	for _, namespace := range config.Destination.Namespaces {
		if namespace == "" {
			return fmt.Errorf("destination namespace cannot be empty")
		}
	}

	return nil
}

func isLabelSelectorEmpty(selector metav1.LabelSelector) bool {
	return len(selector.MatchLabels) == 0 &&
		len(selector.MatchExpressions) == 0
}
