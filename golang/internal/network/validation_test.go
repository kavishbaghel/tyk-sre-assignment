package network

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateIsolationConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  IsolationConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty source namespaces",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source label selector",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty destination namespaces",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty destination label selector",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty source namespace in list",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace", ""},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty destination namespace in list",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"source-namespace"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "source-app",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"destination-namespace", ""},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "destination-app",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid configuration with multiple namespaces",
			config: IsolationConfig{
				Source: WorkloadSelector{
					Namespaces: []string{"frontend", "frontend-staging"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "frontend",
						},
					},
				},
				Destination: WorkloadSelector{
					Namespaces: []string{"backend", "backend-staging"},
					LabelSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "backend",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIsolationConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIsolationConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
