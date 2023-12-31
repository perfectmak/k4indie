package v1alpha1

import (
	"errors"
	"strings"
)

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

var ErrInvalidRuntimeSize = errors.New("invalid runtime size")

type RuntimeImage string

func (r RuntimeImage) Tag() string {
	splits := strings.Split(string(r), ":")
	if len(splits) < 2 {
		return "unknown"
	}

	return splits[1]
}

func (r RuntimeImage) String() string {
	return string(r)
}
