package errgroup

import (
	"errors"
	"slices"
	"strings"
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
			name: "ok",
			args: args{
				data: []func() error{
					func() error {
						return nil
					},
					func() error {
						return nil
					},
					func() error {
						return nil
					},
				},
			},
			wantErr: false,
		},
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
		{
			name: "mixed",
			args: args{
				data: []func() error{
					func() error {
						return nil
					},
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
			if err == nil {
				return
			}
			errArr := strings.Split(err.Error(), "\n")
			if uniqErr := slices.Compact(errArr); len(errArr) != len(uniqErr) {
				t.Errorf("non unique errors: %v, %v", errArr, uniqErr)
			}
		})
	}
}
