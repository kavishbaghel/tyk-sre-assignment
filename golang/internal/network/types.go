package network

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type WorkloadSelector struct {
	Namespaces    []string             `json:"namespaces"`
	LabelSelector metav1.LabelSelector `json:"labelSelector"`
}

type IsolationConfig struct {
	Source      WorkloadSelector `json:"source"`
	Destination WorkloadSelector `json:"destination"`
}

type IsolationResponse struct {
	Status          string `json:"status"`
	PoliciesApplied int    `json:"policiesApplied"`
	Message         string `json:"message,omitempty"`
}

type RevertIsolationResponse struct {
	Status          string `json:"status"`
	PoliciesRemoved int    `json:"policiesRemoved"`
	Message         string `json:"message,omitempty"`
}

type Handler struct {
	clientset kubernetes.Interface
}
