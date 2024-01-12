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

func TestRingBuf_TMP(t *testing.T) {
	b := New[int](5)
	for i := range [5]int{} {
		b.Enqueue(i)
	} // [0 1 2 3 4]
	v, _ := b.Dequeue() // [1 2 3 4]
	assert.Equal(t, 0, *v)
	v, _ = b.Dequeue() // [2 3 4]
	assert.Equal(t, 1, *v)

	for _, v := range []int{11, 22, 33, 44} {
		b.Enqueue(v)
	} // [11, 22, 33, 44, 4]
	v, _ = b.Dequeue()
	assert.Equal(t, 4, *v) // [11, 22, 33, 44]
	v, _ = b.Dequeue()
	assert.Equal(t, 11, *v) // [22, 33, 44]
	v, _ = b.Dequeue()
	assert.Equal(t, 22, *v) // [33, 44]
	v, _ = b.Dequeue()
	assert.Equal(t, 33, *v) // [44]
	v, _ = b.Dequeue()
	assert.Equal(t, 44, *v) // []
	v, err := b.Dequeue()
	assert.Nil(t, v)
	assert.NotNil(t, err)

	// assert.Equal(t, []int{2222, 3, 4, 5, 1111}, b.ToSlice())
}
