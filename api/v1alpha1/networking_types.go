package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Endpoints exposed by the application to be exposed on the internet.
type ApplicationEndpoint struct {
	// Port to expose this endpoint on.
	Port int32 `json:"port,omitempty"`

	// Domain to expose this endpoint on.
	// Leave empty if the application should not be exposed on the internet.
	Domain string `json:"domain,omitempty"`

	// Path to access on the domain to expose this endpoint on. By default it will
	// be exposed on '/' root path.
	//+optional
	//+kubebuilder:default="/"
	DomainPath string `json:"domain_path,omitempty"`
}

type ApplicationEndpoints []ApplicationEndpoint

func (e *ApplicationEndpoint) AsContainerPort() corev1.ContainerPort {
	return corev1.ContainerPort{
		Name:          fmt.Sprintf("port%d", e.Port),
		ContainerPort: e.Port,
	}
}

// AsContainerPorts converts the endpoints to container ports.
// It set ports based on unique port numbers to avoid issues with
// multiple endpoints with the same port number. This won't be a problem
// because on the ingress level, the actual ports and domains will be
// routed properly to the container's same ports
func (e *ApplicationEndpoints) AsContainerPorts() []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0, len(*e))
	portSet := map[string]corev1.ContainerPort{}

	for _, endpoint := range *e {
		containerPort := endpoint.AsContainerPort()
		portSet[containerPort.Name] = containerPort
	}

	for _, port := range portSet {
		ports = append(ports, port)
	}

	return ports
}

// AsServicePort converts the endpoint to a Kubernetes service port.
// Setting the target port and port to the Port specified on the
// application endpoint.
// This only support TCP protocol for now.
func (e *ApplicationEndpoint) AsServicePort() corev1.ServicePort {
	return corev1.ServicePort{
		Name:       fmt.Sprintf("port%d", e.Port),
		Port:       e.Port,
		TargetPort: intstr.FromInt(int(e.Port)),
		Protocol:   corev1.ProtocolTCP,
	}
}

func (e *ApplicationEndpoints) AsServicePorts() []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0, len(*e))
	portSet := map[string]corev1.ServicePort{}

	for _, endpoint := range *e {
		servicePort := endpoint.AsServicePort()
		portSet[servicePort.Name] = servicePort
	}

	for _, port := range portSet {
		ports = append(ports, port)
	}

	return ports
}
