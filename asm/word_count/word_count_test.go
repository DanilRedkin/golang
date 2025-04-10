package word_count

import (
	"fmt"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func WordCount(data []rune) int32

func TestWordCount(t *testing.T) {
	t.Parallel()

	type testCases struct {
		name   string
		arg    []rune
		result int32
	}

	tableTests := []testCases{
		{
			name:   "zero",
			arg:    []rune{},
			result: 0,
		},
		{
			name:   "one",
			arg:    []rune("word"),
			result: 1,
		},
		{
			name:   "even alternation",
			arg:    []rune(" go     optimizations       course      here    "),
			result: 4,
		},
		{
			name:   "odd alternation",
			arg:    []rune("go optimizations course here"),
			result: 4,
		},
		{
			name:   "less then max latin",
			arg:    []rune{'\t', '\n', 'g', '\v', '\f', 'o', '\r', ' ', 0x85, 'c', 0xA0},
			result: 3,
		},
		{
			name:   "more than max latin",
			arg:    []rune{'1', rune(unicode.White_Space.R16[2].Lo), '2'},
			result: 2,
		},
	}

	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.result, WordCount(tt.arg))
		})
	}

	t.Run("TestWordCountRanges", func(t *testing.T) {
		t.Parallel()

		targetRange := unicode.White_Space.R16
		template := "%s Hello, %s %s bonus %s task %s"

		for _, e := range targetRange {
			for i := e.Lo; i <= e.Hi; i += e.Stride {
				r := string(rune(i))
				s := fmt.Sprintf(template, r, r, r, r, r)
				require.EqualValues(t, 3, WordCount([]rune(s)))
			}
		}
	})
}
