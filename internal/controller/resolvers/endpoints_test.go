package resolvers

import (
	"reflect"
	"testing"

	"github.com/perfectmak/k4indie/api/v1alpha1"
)

func TestEndpointsWithDomains(t *testing.T) {
	type args struct {
		endpoints *v1alpha1.ApplicationEndpoints
	}
	tests := []struct {
		name string
		args args
		want []v1alpha1.ApplicationEndpoint
	}{
		{
			name: "should return empty list if no endpoints",
			args: args{
				endpoints: &v1alpha1.ApplicationEndpoints{},
			},
			want: []v1alpha1.ApplicationEndpoint{},
		},
		{
			name: "should return empty list if no endpoints with domains",
			args: args{
				endpoints: &v1alpha1.ApplicationEndpoints{
					{
						Port:   80,
						Domain: "",
					},
				},
			},
			want: []v1alpha1.ApplicationEndpoint{},
		},
		{
			name: "should return endpoints with domains",
			args: args{
				endpoints: &v1alpha1.ApplicationEndpoints{
					{
						Port:   80,
						Domain: "example.com",
					},
					{
						Port:   8080,
						Domain: "",
					},
				},
			},
			want: []v1alpha1.ApplicationEndpoint{
				{
					Port:   80,
					Domain: "example.com",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EndpointsWithDomains(tt.args.endpoints); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EndpointsWithDomains() = %v, want %v", got, tt.want)
			}
		})
	}
}
