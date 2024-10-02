package errgroup

import (
	"fmt"
	"strings"
	"testing"

	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/stretchr/testify/require"
)

func getErrGroup(errsCount int, shouldPanic bool) *errgroup.Group {
	eg := &errgroup.Group{}

	for i := 1; i <= errsCount; i++ {
		eg.Go(func() error {
			return fmt.Errorf("error %d", i)
		})
	}

	if shouldPanic {
		eg.Go(func() error {
			panic("panic")
		})
	}

	return eg
}

func TestErrGroup(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *errgroup.Group
		expectErr   bool
		errCount    int
		expectPanic bool
	}{
		{
			name: "no errors",
			setup: func() *errgroup.Group {
				return getErrGroup(0, false)
			},
			expectErr:   false,
			errCount:    0,
			expectPanic: false,
		},

		{
			name: "one error",
			setup: func() *errgroup.Group {
				return getErrGroup(1, false)
			},
			expectErr:   true,
			errCount:    1,
			expectPanic: false,
		},

		{
			name: "two errors",
			setup: func() *errgroup.Group {
				return getErrGroup(2, false)
			},
			expectErr:   true,
			errCount:    2,
			expectPanic: false,
		},

		{
			name: "panic",
			setup: func() *errgroup.Group {
				return getErrGroup(0, true)
			},
			expectErr:   false,
			errCount:    0,
			expectPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					require.True(t, tc.expectPanic)
				}
			}()

			eg := tc.setup()
			err := eg.Wait()

			if tc.expectErr {
				require.NotNil(t, err)
				errCount := len(strings.Split(err.Error(), "\n"))
				require.Equal(t, tc.errCount, errCount)
			} else {
				require.Nil(t, err)
			}

			require.False(t, tc.expectPanic)
		})
	}
}
