package alert

import (
	"cmp"
	"context"
	"slices"
	"sort"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueueScheduledTask(t *testing.T) {
	type producedTaks struct {
		id              int
		issuedIn        time.Duration
		startIn         time.Duration
		expectedStartIn time.Duration
	}

	tests := []struct {
		name         string
		producedTaks []producedTaks
	}{
		{
			name:         "empty",
			producedTaks: []producedTaks{},
		},
		{
			name: "monotonic",
			producedTaks: []producedTaks{
				{id: 1, issuedIn: 0, startIn: 10 * time.Second},
				{id: 2, issuedIn: 0, startIn: 20 * time.Second},
				{id: 3, issuedIn: 0, startIn: 30 * time.Second},
			},
		}, {
			name: "issued later, earlier start",
			producedTaks: []producedTaks{
				{id: 3, issuedIn: 0, startIn: 30 * time.Second},
				{id: 2, issuedIn: 0, startIn: 20 * time.Second},
				{id: 1, issuedIn: 0, startIn: 10 * time.Second},
			},
		}, {
			name: "issued later, earlier start",
			producedTaks: []producedTaks{
				{id: 1, issuedIn: 0, startIn: time.Second * 2},
				{id: 2, issuedIn: time.Second, startIn: time.Second},
			},
		}, {
			name: "negative start",
			producedTaks: []producedTaks{
				{id: 1, issuedIn: 0, startIn: 10 * time.Second},
				{id: 2, issuedIn: 50 * time.Second, startIn: -10 * time.Second, expectedStartIn: 50 * time.Second},
				{id: 3, issuedIn: 0, startIn: 60 * time.Second},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				queue := newQueue[int]()

				expected := slices.Clone(test.producedTaks)

				sort.Slice(expected, func(i, j int) bool {
					expectedStartIn := cmp.Or(expected[i].expectedStartIn, expected[i].startIn)
					expectedStartIn2 := cmp.Or(expected[j].expectedStartIn, expected[j].startIn)
					return expectedStartIn < expectedStartIn2
				})

				now := time.Now()
				go func() {
					sort.Slice(test.producedTaks, func(i, j int) bool {
						return test.producedTaks[i].issuedIn < test.producedTaks[j].issuedIn
					})
					for _, task := range test.producedTaks {
						time.Sleep(time.Until(now.Add(task.issuedIn)))
						queue.Push(task.id, now.Add(task.startIn))
					}
					queue.Close()
				}()

				for i, task := range queue.SyncSeq2(context.Background()) {
					require.Equal(t, expected[i].id, task)
					expectedStartIn := cmp.Or(expected[i].expectedStartIn, expected[i].startIn)
					deviation := time.Since(now.Add(expectedStartIn))
					deviation = max(deviation, -deviation) // abs
					require.Less(t, deviation, 10*time.Millisecond, "task %d", task)
				}
			})
		})
	}
}

func TestQueueClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newQueue[int]()
		expected := make([]int, 0, 10)
		for i := range cap(expected) {
			expected = append(expected, i)
			queue.Push(i, time.Now().Add(time.Duration(i)*time.Second))
		}
		queue.Close()

		actual := slices.Collect(queue.SyncSeq(context.Background()))
		require.Equal(t, expected, actual)
	})
}

func TestQueueContextDone(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newQueue[int]()
		now := time.Now()
		queue.Push(1, now.Add(time.Second))
		queue.Push(2, now.Add(2*time.Second))
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		queue.Push(4, now.Add(4*time.Second))
		queue.Push(5, now.Add(5*time.Second))
		actual := slices.Collect(queue.SyncSeq(ctx))
		require.Equal(t, []int{1, 2}, actual)
	})
}

func TestQueueRemoveFunc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		queue := newQueue[int]()
		queue.Push(1, time.Now().Add(time.Second))
		queue.Push(2, time.Now().Add(2*time.Second))
		queue.Push(3, time.Now().Add(3*time.Second))
		queue.Push(4, time.Now().Add(4*time.Second))

		go func() {
			time.Sleep(1500 * time.Millisecond)
			for _, v := range []int{2, 4} {
				queue.RemoveFirstFunc(func(i int) bool {
					return i == v
				})
			}
			queue.Close()
		}()

		actual := slices.Collect(queue.SyncSeq(context.Background()))
		require.Equal(t, []int{1, 3}, actual)
	})
}
