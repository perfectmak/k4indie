package resolvers

import (
	"reflect"
	"testing"
)

func TestMergeDefaultLabels(t *testing.T) {
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "should merge default labels",
			args: args{
				labels: map[string]string{
					"app.kubernetes.io/instance": "test",
					"app.kubernetes.io/version":  "test",
				},
			},
			want: map[string]string{
				"app.kubernetes.io/instance":   "test",
				"app.kubernetes.io/version":    "test",
				"app.kubernetes.io/part-of":    "k4indie-operator",
				"app.kubernetes.io/created-by": "controller-manager",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeDefaultLabels(tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeDefaultLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
