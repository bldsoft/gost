package utils

import (
	"testing"
)

func TestProbe(t *testing.T) {
	type tmp struct {
		Duration float64 `json:"duration"`
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    tmp
		wantErr bool
	}{
		{
			name: "",
			args: args{
				path: "media_test.ts",
			},
			want: tmp{
				Duration: 6.013333,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Probe(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if val, _ := got.Duration(); val != tt.want.Duration {
				t.Errorf("Probe() duration = %v, want.Duration %v", val, tt.want.Duration)
				return
			}
		})
	}
}
