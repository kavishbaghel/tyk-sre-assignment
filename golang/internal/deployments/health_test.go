package deployments

import (
	"testing"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsDeploymentHealthy(t *testing.T) {
	desiredReplicas := int32(3)

	testData := []struct {
		name       string
		deployment *v1.Deployment
		expected   bool
	}{
		{
			name:       "Nil deployment",
			deployment: nil,
			expected:   false,
		},
		{
			name: "Deployment with unspecified replicas",
			deployment: &v1.Deployment{
				Spec: v1.DeploymentSpec{
					Replicas: nil,
				},
			},
			expected: false,
		},
		{
			name: "Deployment with desired replicas equal to ready replicas",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: v1.DeploymentSpec{
					Replicas: &desiredReplicas,
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      desiredReplicas,
					ObservedGeneration: 1,
				},
			},
			expected: true,
		},
		{
			name: "Deployment with ready replicas more than desired replicas",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: v1.DeploymentSpec{
					Replicas: &desiredReplicas,
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      4,
					ObservedGeneration: 1,
				},
			},
			expected: true,
		},
		{
			name: "Deployment with desired replicas greater than ready replicas",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: v1.DeploymentSpec{
					Replicas: &desiredReplicas,
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      2,
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "Deployment with stale observed generation",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 2,
				},
				Spec: v1.DeploymentSpec{
					Replicas: &desiredReplicas,
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      3,
					ObservedGeneration: 1,
				},
			},
			expected: false,
		},
		{
			name: "Deployment with observed generation ahead of current generation",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: v1.DeploymentSpec{
					Replicas: &desiredReplicas,
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      3,
					ObservedGeneration: 2,
				},
			},
			expected: true,
		},
		{
			name: "Deployment with zero desired replicas",
			deployment: &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 1,
				},
				Spec: v1.DeploymentSpec{
					Replicas: new(int32), // Zero replicas
				},
				Status: v1.DeploymentStatus{
					ReadyReplicas:      0,
					ObservedGeneration: 1,
				},
			},
			expected: true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDeploymentHealthy(tt.deployment)
			if got != tt.expected {
				t.Errorf(
					"IsDeploymentHealthy() = %v, want %v",
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestGetDeploymentHealth(t *testing.T) {
	replicas3 := int32(3)

	tests := []struct {
		name        string
		deployments []v1.Deployment
		want        DeploymentHealthResponse
	}{
		{
			name:        "empty deployment list",
			deployments: []v1.Deployment{},
			want: DeploymentHealthResponse{
				Status:               "healthy",
				UnhealthyDeployments: []UnhealthyDeployment{},
			},
		},
		{
			name: "all deployments are healthy",
			deployments: []v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "nginx",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      3,
						ObservedGeneration: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "api",
						Namespace:  "production",
						Generation: 2,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      3,
						ObservedGeneration: 2,
					},
				},
			},
			want: DeploymentHealthResponse{
				Status:               "healthy",
				UnhealthyDeployments: []UnhealthyDeployment{},
			},
		},
		{
			name: "deployment has insufficient ready replicas",
			deployments: []v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "nginx",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      2,
						ObservedGeneration: 1,
					},
				},
			},
			want: DeploymentHealthResponse{
				Status: "unhealthy",
				UnhealthyDeployments: []UnhealthyDeployment{
					{
						Namespace: "default",
						Name:      "nginx",
						Reason:    "Deployment has fewer available replicas than desired",
					},
				},
			},
		},
		{
			name: "deployment generation has not been observed",
			deployments: []v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "api",
						Namespace:  "production",
						Generation: 2,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      3,
						ObservedGeneration: 1,
					},
				},
			},
			want: DeploymentHealthResponse{
				Status: "unhealthy",
				UnhealthyDeployments: []UnhealthyDeployment{
					{
						Namespace: "production",
						Name:      "api",
						Reason:    "Latest deployment generation has not been observed by the controller",
					},
				},
			},
		},
		{
			name: "deployment replicas are unspecified",
			deployments: []v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "worker",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: v1.DeploymentSpec{
						Replicas: nil,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      3,
						ObservedGeneration: 1,
					},
				},
			},
			want: DeploymentHealthResponse{
				Status: "unhealthy",
				UnhealthyDeployments: []UnhealthyDeployment{
					{
						Namespace: "default",
						Name:      "worker",
						Reason:    "Deployment has no replicas defined",
					},
				},
			},
		},
		{
			name: "multiple deployments are unhealthy",
			deployments: []v1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "frontend",
						Namespace:  "default",
						Generation: 1,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      1,
						ObservedGeneration: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "backend",
						Namespace:  "production",
						Generation: 3,
					},
					Spec: v1.DeploymentSpec{
						Replicas: &replicas3,
					},
					Status: v1.DeploymentStatus{
						ReadyReplicas:      3,
						ObservedGeneration: 2,
					},
				},
			},
			want: DeploymentHealthResponse{
				Status: "unhealthy",
				UnhealthyDeployments: []UnhealthyDeployment{
					{
						Namespace: "default",
						Name:      "frontend",
						Reason:    "Deployment has fewer available replicas than desired",
					},
					{
						Namespace: "production",
						Name:      "backend",
						Reason:    "Latest deployment generation has not been observed by the controller",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDeploymentHealth(tt.deployments)

			if got.Status != tt.want.Status {
				t.Errorf(
					"GetDeploymentHealth() status = %q, want %q",
					got.Status,
					tt.want.Status,
				)
			}

			if len(got.UnhealthyDeployments) != len(tt.want.UnhealthyDeployments) {
				t.Fatalf(
					"GetDeploymentHealth() returned %d unhealthy deployments, want %d",
					len(got.UnhealthyDeployments),
					len(tt.want.UnhealthyDeployments),
				)
			}

			for i := range tt.want.UnhealthyDeployments {
				gotDeployment := got.UnhealthyDeployments[i]
				wantDeployment := tt.want.UnhealthyDeployments[i]

				if gotDeployment.Namespace != wantDeployment.Namespace {
					t.Errorf(
						"deployment[%d].Namespace = %q, want %q",
						i,
						gotDeployment.Namespace,
						wantDeployment.Namespace,
					)
				}

				if gotDeployment.Name != wantDeployment.Name {
					t.Errorf(
						"deployment[%d].Name = %q, want %q",
						i,
						gotDeployment.Name,
						wantDeployment.Name,
					)
				}

				if gotDeployment.Reason != wantDeployment.Reason {
					t.Errorf(
						"deployment[%d].Reason = %q, want %q",
						i,
						gotDeployment.Reason,
						wantDeployment.Reason,
					)
				}
			}
		})
	}
}
