package small_rational

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmp(t *testing.T) {

	cases := make([]SmallRational, 36)

	for i := int64(0); i < 9; i++ {
		if i%2 == 0 {
			cases[4*i].Numerator.SetInt64((i - 4) / 2)
			cases[4*i].Denominator.SetInt64(1)
		} else {
			cases[4*i].Numerator.SetInt64(i - 4)
			cases[4*i].Denominator.SetInt64(2)
		}

		cases[4*i+1].Numerator.Neg(&cases[4*i].Numerator)
		cases[4*i+1].Denominator.Neg(&cases[4*i].Denominator)

		cases[4*i+2].Numerator.Lsh(&cases[4*i].Numerator, 1)
		cases[4*i+2].Denominator.Lsh(&cases[4*i].Denominator, 1)

		cases[4*i+3].Numerator.Neg(&cases[4*i+2].Numerator)
		cases[4*i+3].Denominator.Neg(&cases[4*i+2].Denominator)
	}

	for i := range cases {
		for j := range cases {
			I, J := i/4, j/4
			var expectedCmp int
			cmp := cases[i].Cmp(&cases[j])
			if I < J {
				expectedCmp = -1
			} else if I == J {
				expectedCmp = 0
			} else {
				expectedCmp = 1
			}
			assert.Equal(t, expectedCmp, cmp, "comparing index %d, index %d", i, j)
		}
	}

	zeroIndex := len(cases) / 8
	var weirdZero SmallRational
	for i := range cases {
		I := i / 4
		var expectedCmp int
		cmp := cases[i].Cmp(&weirdZero)
		cmpNeg := weirdZero.Cmp(&cases[i])
		if I < zeroIndex {
			expectedCmp = -1
		} else if I == zeroIndex {
			expectedCmp = 0
		} else {
			expectedCmp = 1
		}

		assert.Equal(t, expectedCmp, cmp, "comparing index %d, 0/0", i)
		assert.Equal(t, -expectedCmp, cmpNeg, "comparing 0/0, index %d", i)
	}
}
