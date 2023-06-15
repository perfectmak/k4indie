package v1alpha1

// +kubebuilder:validation:Enum=basic;basic-2x;standard-2x;performance
type RuntimeSize string

var (
	// 256MB RAM, 256m vCPU
	BasicMachineType RuntimeSize = "basic"
	// 256MB RAM, 500m vCPU
	Basic2xMachineType RuntimeSize = "basic-2x"
	// 512MB RAM, 500m vCPU
	StandardMachineType RuntimeSize = "standard"
	// 512MB RAM, 1vCPU
	Standard2xMachineType RuntimeSize = "standard-2x"
	// 1GB RAM, 2vCPU
	PerformanceMachineType RuntimeSize = "performance"
)

// Map of all valid runtime sizes.
var RuntimeSizes = map[RuntimeSize]struct{}{
	BasicMachineType:       {},
	Basic2xMachineType:     {},
	StandardMachineType:    {},
	Standard2xMachineType:  {},
	PerformanceMachineType: {},
}
