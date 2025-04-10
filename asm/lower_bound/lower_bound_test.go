package lower_bound

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func LowerBound(slice []int64, value int64) int64

func TestLowerBound(t *testing.T) {
	type testCases struct {
		name   string
		arg    []int64
		value  int64
		result int64
	}

	tableTests := []testCases{
		{
			name:   "exact match",
			arg:    []int64{1, 2, 3, 4},
			value:  3,
			result: 2,
		},
		{
			name:   "empty",
			arg:    []int64{},
			value:  100,
			result: -1,
		},
		{
			name:   "none",
			arg:    []int64{10, 20, 30},
			value:  5,
			result: -1,
		},
		{
			name:   "last",
			arg:    []int64{5, 6, 7},
			value:  11,
			result: 2,
		},
		{
			name:   "one",
			arg:    []int64{-1},
			value:  -1,
			result: 0,
		},
		{
			name:   "one",
			arg:    []int64{-1},
			value:  1,
			result: 0,
		},
		{
			name:   "first",
			arg:    []int64{5, 10, 15},
			value:  7,
			result: 0,
		},
		{
			name:   "lower match",
			arg:    []int64{1, 2, 6, 8},
			value:  7,
			result: 2,
		},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.result, LowerBound(tt.arg, tt.value))
		})
	}

	t.Run("32-overflow", func(t *testing.T) {
		s := make([]int64, 10)

		s[len(s)-5] = 1 << 33
		s[len(s)-4] = 1 << 34
		s[len(s)-3] = 1 << 35
		s[len(s)-2] = 1 << 36
		s[len(s)-1] = 1 << 37

		require.EqualValues(t, len(s)-3, LowerBound(s, 1<<35))
	})

	t.Run("performance test", func(t *testing.T) {
		const index = 432000

		solution := testing.Benchmark(func(b *testing.B) {
			s := make([]int64, 1_000_000)

			for i := 0; i < len(s); i++ {
				s[i] = int64(i)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ind := LowerBound(s, int64(index))
				require.Equal(t, int64(index), ind)
			}
		})

		check := testing.Benchmark(func(b *testing.B) {
			s := make([]int64, 1_000_000)

			for i := 0; i < len(s); i++ {
				s[i] = int64(i)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				ind, ok := slices.BinarySearch(s, index)

				require.True(t, ok)
				require.Equal(t, index, ind)
			}
		})

		require.Less(t, (float64(solution.NsPerOp())+10e-9)/(float64(check.NsPerOp())+10e-9), 3.0)
	})
}
