package polynomial

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"testing"
)

func TestFoldBilinear(t *testing.T) {
	// f = c₀ + c₁ X₁ + c₂ X₂ + c₃ X₁ X₂
	var coefficients [4]fr.Element
	for i := 0; i < 4; i++ {
		if _, err := coefficients[i].SetRandom(); err != nil {
			t.Error(err)
		}
	}

	var r fr.Element
	if _, err := r.SetRandom(); err != nil {
		t.Error(err)
	}

	// interpolate at {0,1}²:
	m := make(MultiLin, 4)
	m[0] = coefficients[0]
	m[1].Add(&coefficients[0], &coefficients[2])
	m[2].Add(&coefficients[0], &coefficients[1])
	m[3].
		Add(&m[1], &coefficients[1]).
		Add(&m[3], &coefficients[3])

	m.Fold(r)

	// interpolate at {r}×{0,1}:
	var expected0, expected1 fr.Element
	expected0.
		Mul(&r, &coefficients[1]).
		Add(&expected0, &coefficients[0])

	expected1.
		Mul(&r, &coefficients[3]).
		Add(&expected1, &coefficients[2]).
		Add(&expected0, &expected1)

	if !m[0].Equal(&expected0) || !m[1].Equal(&expected1) {
		t.Fail()
	}
}
