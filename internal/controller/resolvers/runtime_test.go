package resolvers

import (
	"reflect"
	"testing"

	"github.com/perfectmak/k4indie/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetResourcesForRuntimeSize(t *testing.T) {
	type args struct {
		size v1alpha1.RuntimeSize
	}
	tests := []struct {
		name    string
		args    args
		want    corev1.ResourceRequirements
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				size: v1alpha1.BasicMachineType,
			},
			want: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("256m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("256m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
			wantErr: false,
		},
		{
			name: "incorrect size",
			args: args{
				size: "incorrect",
			},
			want:    corev1.ResourceRequirements{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResourcesForRuntimeSize(tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourcesForRuntimeSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResourcesForRuntimeSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
