package utils

import (
	"reflect"
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
				path: "test_files/media_test.ts",
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

func TestProbeInto(t *testing.T) {
	type args struct {
		path string
		res  interface{}
		args map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ProbeInto(tt.args.path, tt.args.res, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("ProbeInto() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProbeWithArgs(t *testing.T) {
	type args struct {
		path string
		args map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *FFMpegProbe
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProbeWithArgs(tt.args.path, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProbeWithArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProbeWithArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
