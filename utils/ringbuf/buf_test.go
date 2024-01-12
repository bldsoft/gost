package ringbuf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingBuf_Enqueue(t *testing.T) {
	type args struct {
		from []int
		size int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "same",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7, 8}, 8},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "smaller",
			args: args{[]int{1, 2, 3, 4, 5}, 6},
			want: []int{1, 2, 3, 4, 5, 0},
		},
		{
			name: "bigger",
			args: args{[]int{1, 2, 3, 4, 5}, 4},
			want: []int{5, 2, 3, 4},
		},
		{
			name: "x2 + 1",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7}, 3},
			want: []int{7, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New[int](tt.args.size)
			for _, v := range tt.args.from {
				b.Enqueue(v)
			}
			assert.Equal(t, tt.want, b.data)
		})
	}
}

func TestRingBuf_ToSlice(t *testing.T) {
	type args struct {
		from []int
		size int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "same",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7, 8}, 8},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "smaller",
			args: args{[]int{1, 2, 3, 4, 5}, 6},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "bigger",
			args: args{[]int{1, 2, 3, 4, 5}, 4},
			want: []int{5, 2, 3, 4},
		},
		{
			name: "x2 + 1",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7}, 3},
			want: []int{7, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New[int](tt.args.size)
			for _, v := range tt.args.from {
				b.Enqueue(v)
			}
			assert.Equal(t, tt.want, b.ToSlice())
		})
	}
}

func TestRingBuf_ToSliceAfterClean(t *testing.T) {
	type args struct {
		from []int
		size int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "same",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7, 8}, 8},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name: "smaller",
			args: args{[]int{1, 2, 3, 4, 5}, 6},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name: "bigger",
			args: args{[]int{1, 2, 3, 4, 5}, 4},
			want: []int{5, 2, 3, 4},
		},
		{
			name: "x2 + 1",
			args: args{[]int{1, 2, 3, 4, 5, 6, 7}, 3},
			want: []int{7, 5, 6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New[int](tt.args.size)
			for i := range [10]int{} {
				b.Enqueue(i)
			}
			b.Clear()

			for _, v := range tt.args.from {
				b.Enqueue(v)
			}

			assert.Equal(t, tt.want, b.ToSlice())
		})
	}
}
