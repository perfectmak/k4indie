package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestApplicationEndpoints_AsContainerPorts(t *testing.T) {
	tests := []struct {
		name string
		e    *ApplicationEndpoints
		want []corev1.ContainerPort
	}{
		{
			name: "empty",
			e:    &ApplicationEndpoints{},
			want: []corev1.ContainerPort{},
		},
		{
			name: "single",
			e: &ApplicationEndpoints{
				{
					Port:   80,
					Domain: "example.com",
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          "port80",
					ContainerPort: 80,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.AsContainerPorts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ApplicationEndpoints.AsContainerPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}
