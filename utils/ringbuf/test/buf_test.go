package ringbuf

import (
	"testing"

	"github.com/bldsoft/gost/utils/ringbuf"
	"github.com/stretchr/testify/require"
)

func TestRingBuf(t *testing.T) {
	type args struct {
		capacity  int
		overwrite bool
		actions   []interface{}
	}
	type push struct {
		items         []int
		expectedPushN int
		expectedLen   int
	}
	type pull struct {
		expectedRead []int
		expectedLen  int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty",
			args: args{
				capacity:  10,
				overwrite: false,
				actions: []interface{}{
					push{
						items:       nil,
						expectedLen: 0,
					},
				},
			},
		},
		{
			name: "read from empty",
			args: args{
				capacity:  10,
				overwrite: false,
				actions: []interface{}{
					pull{
						expectedRead: []int{0, 0, 0, 0, 0},
						expectedLen:  0,
					},
				},
			},
		},
		{
			name: "push half, pull all",
			args: args{
				capacity:  10,
				overwrite: false,
				actions: []interface{}{
					push{
						items:         []int{1, 2, 3, 4, 5},
						expectedPushN: 5,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{1, 2, 3, 4, 5},
						expectedLen:  0,
					},
				},
			},
		},
		{
			name: "push all, pull all",
			args: args{
				capacity:  5,
				overwrite: false,
				actions: []interface{}{
					push{
						items:         []int{1, 2, 3, 4, 5},
						expectedPushN: 5,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{1, 2, 3, 4, 5},
						expectedLen:  0,
					},
				},
			},
		},
		{
			name: "push extra",
			args: args{
				capacity:  5,
				overwrite: false,
				actions: []interface{}{
					push{
						items:         []int{1, 2, 3, 4, 5, 6, 7},
						expectedPushN: 5,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{1, 2, 3, 4, 5},
						expectedLen:  0,
					},
				},
			},
		},
		{
			name: "splitted pull",
			args: args{
				capacity:  5,
				overwrite: false,
				actions: []interface{}{
					push{
						items:         []int{1, 2, 3, 4, 5},
						expectedPushN: 5,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{1, 2},
						expectedLen:  3,
					},
					push{
						items:         []int{6, 7, 8},
						expectedPushN: 2,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{3, 4, 5, 6, 7, 0, 0},
						expectedLen:  0,
					},
				},
			},
		},
		{
			name: "overwrite",
			args: args{
				capacity:  5,
				overwrite: true,
				actions: []interface{}{
					push{
						items:         []int{1, 2, 3, 4, 5, 6, 7},
						expectedPushN: 7,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{3},
						expectedLen:  4,
					},
					push{
						items:         []int{8, 9},
						expectedPushN: 2,
						expectedLen:   5,
					},
					pull{
						expectedRead: []int{5, 6, 7, 8, 9, 0},
						expectedLen:  0,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ringBuf := ringbuf.New[int](tt.args.capacity).WithOverwrite(tt.args.overwrite)
			for _, action := range tt.args.actions {
				switch a := action.(type) {
				case push:
					require.Equal(t, a.expectedPushN, ringBuf.Push(a.items...))

					require.Equal(t, tt.args.capacity, ringBuf.Cap())
					require.Equal(t, a.expectedLen, ringBuf.Len())
					require.Equal(t, a.expectedLen == 0, ringBuf.Empty())
					require.Equal(t, a.expectedLen == tt.args.capacity, ringBuf.Full())
				case pull:
					for _, exp := range a.expectedRead {
						act, _ := ringBuf.Pull()
						require.Equal(t, exp, act)
					}

					require.Equal(t, tt.args.capacity, ringBuf.Cap())
					require.Equal(t, a.expectedLen, ringBuf.Len())
					require.Equal(t, a.expectedLen == 0, ringBuf.Empty())
					require.Equal(t, a.expectedLen == tt.args.capacity, ringBuf.Full())
				}
			}
		})
	}
}

func TestRingBufCopy(t *testing.T) {
	makeBuf := func(l, c int) *ringbuf.RingBuf[int] {
		res := ringbuf.New[int](c)
		for i := 1; i <= l; i++ {
			res.Push(i)
		}
		return res
	}
	tests := []struct {
		name     string
		buf      *ringbuf.RingBuf[int]
		expected []int
	}{
		{
			name:     "empty buf",
			buf:      ringbuf.New[int](10),
			expected: []int{0, 0, 0},
		},
		{
			name:     "half filled buf",
			buf:      makeBuf(5, 10),
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "full buf",
			buf:      makeBuf(5, 5),
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name: "splitted buf",
			buf: func() *ringbuf.RingBuf[int] {
				res := ringbuf.New[int](10)
				res.Push(1, 2, 3, 4, 5)
				res.Pull()
				res.Pull()
				res.Pull()
				res.Push(6)
				return res
			}(),
			expected: []int{4, 5, 6},
		},
		{
			name:     "small destination",
			buf:      makeBuf(10, 10),
			expected: []int{1, 2},
		},
		{
			name:     "zero destination",
			buf:      makeBuf(10, 10),
			expected: []int{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dst := make([]int, len(tt.expected))
			require.Equal(t, min(len(dst), tt.buf.Len()), tt.buf.Copy(dst))
			require.Equal(t, tt.expected, dst)
		})
	}
}
