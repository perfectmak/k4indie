package resolvers

import (
	"github.com/perfectmak/k4indie/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getCpuAndMemoryForRuntimeSize(size v1alpha1.RuntimeSize) (cpu, memory string, err error) {
	switch size {
	case v1alpha1.BasicMachineType:
		cpu = "256m"
		memory = "256Mi"
	case v1alpha1.Basic2xMachineType:
		cpu = "500m"
		memory = "256Mi"
	case v1alpha1.StandardMachineType:
		cpu = "500m"
		memory = "512Mi"
	case v1alpha1.Standard2xMachineType:
		cpu = "1"
		memory = "1Gi"
	case v1alpha1.PerformanceMachineType:
		cpu = "2"
		memory = "2Gi"
	default:
		err = v1alpha1.ErrInvalidRuntimeSize
	}
	return
}

func GetResourcesForRuntimeSize(size v1alpha1.RuntimeSize) (corev1.ResourceRequirements, error) {
	cpu, memory, err := getCpuAndMemoryForRuntimeSize(size)
	if err != nil {
		return corev1.ResourceRequirements{}, err
	}

	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpu),
			corev1.ResourceMemory: resource.MustParse(memory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpu),
			corev1.ResourceMemory: resource.MustParse(memory),
		},
	}, nil
}
