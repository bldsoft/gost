package errgroup

import (
	"errors"
	"testing"
)

func TestGroup(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")
	type args struct {
		data []func() error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "errors",
			args: args{
				data: []func() error{
					func() error {
						return err1
					},
					func() error {
						return err2
					},
					func() error {
						return err1
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var eg Group
			for _, f := range tt.args.data {
				func(f func() error) {
					eg.Go(f)
				}(f)
			}
			err := eg.Wait()
			if (err != nil) != tt.wantErr {
				t.Errorf("errgroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
