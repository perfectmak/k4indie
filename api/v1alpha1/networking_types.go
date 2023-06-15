package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type ApplicationEndpoint struct {
	// Port to expose this endpoint on.
	Port int32 `json:"port,omitempty"`

	// Domain to expose this endpoint on.
	// Leave empty if the application should not be exposed on the internet.
	Domain string `json:"domain,omitempty"`
}

type ApplicationEndpoints []ApplicationEndpoint

func (e *ApplicationEndpoint) AsContainerPort() corev1.ContainerPort {
	return corev1.ContainerPort{
		Name:          fmt.Sprintf("port%d", e.Port),
		ContainerPort: e.Port,
	}
}

func (e *ApplicationEndpoints) AsContainerPorts() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, len(*e))

	for i, endpoint := range *e {
		ports[i] = endpoint.AsContainerPort()
	}

	return ports
}
