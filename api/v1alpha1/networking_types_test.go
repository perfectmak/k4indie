package v1alpha1

import (
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
				{
					Port: 8080,
				},
				{
					Port:   8080,
					Domain: "kudi.ai",
				},
			},
			want: []corev1.ContainerPort{
				{
					Name:          "port80",
					ContainerPort: 80,
				},
				{
					Name:          "port8080",
					ContainerPort: 8080,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sortContainers(tt.e.AsContainerPorts())
			expected := sortContainers(tt.want)

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(
					"ApplicationEndpoints.AsContainerPorts() = %v, want %v",
					actual,
					expected,
				)
			}
		})
	}
}

func TestApplicationEndpoints_AsServicePorts(t *testing.T) {
	tests := []struct {
		name string
		e    *ApplicationEndpoints
		want []corev1.ServicePort
	}{
		{
			name: "empty",
			e:    &ApplicationEndpoints{},
			want: []corev1.ServicePort{},
		},
		{
			name: "single",
			e: &ApplicationEndpoints{
				{
					Port:   80,
					Domain: "example.com",
				},
				{
					Port: 8080,
				},
				{
					Port:   8080,
					Domain: "kudi.ai",
				},
			},
			want: []corev1.ServicePort{
				{
					Name:       "port80",
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
				{
					Name:       "port8080",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sortServicePorts(tt.e.AsServicePorts())
			expected := sortServicePorts(tt.want)

			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(
					"ApplicationEndpoints.AsServicePorts() = %v, want %v",
					actual,
					expected,
				)
			}
		})
	}
}

func sortContainers(ports []corev1.ContainerPort) []corev1.ContainerPort {
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].ContainerPort < ports[j].ContainerPort
	})

	return ports
}

func sortServicePorts(ports []corev1.ServicePort) []corev1.ServicePort {
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	return ports
}
