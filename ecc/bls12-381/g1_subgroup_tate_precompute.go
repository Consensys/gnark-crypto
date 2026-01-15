package bls12381

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

type loopkupTable struct {
	q   G1Affine
	tab []fp.Element
}

var (
	precomputeTableOnce sync.Once
	precomputedTable    loopkupTable

	tateExp1Once     sync.Once
	tateExp1Exponent big.Int
)

// NAF digits used by Algorithms 3/4 (Section 4.2) for BLS12-381.
var naf = [65]int8{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, -1, 0, 1,
}

// precomputeTableDefault returns a cached precomputation table for torsionPoint.
func precomputeTableDefault() loopkupTable {
	precomputeTableOnce.Do(func() {
		precomputedTable = generateTable(&torsionPoint)
	})
	return precomputedTable
}

// generateTable precomputes the lookup table used by the Tate-based G1 membership test.
// Algorithm 3 (Section 4.2): lookup table generation for the shared Miller loop.
func generateTable(q *G1Affine) loopkupTable {
	i := 63
	j := 0
	if naf[0] < 0 {
		j = 1
	}

	tab := make([]fp.Element, 0, len(naf)*4)
	var t0, t1, qNeg G1Affine
	t0.Set(q)
	qNeg.Neg(q)

	var u0, u1 fp.Element

	for i >= j {
		if naf[i] == 0 && i > j {
			u0.Square(&t0.X)
			u1.Double(&u0)
			u0.Add(&u0, &u1)
			u1.Double(&t0.Y)
			u1.Inverse(&u1)
			u1.Neg(&u1)
			u0.Mul(&u0, &u1)
			tab = append(tab, u0)

			t0.Double(&t0)
			u0.Square(&t0.X)
			u1.Double(&u0)
			u0.Add(&u0, &u1)
			u1.Double(&t0.Y)
			u1.Inverse(&u1)
			tab = append(tab, t0.X, t0.Y)
			u0.Mul(&u0, &u1)
			tab = append(tab, u0)

			t0.Double(&t0)
			i--

			if naf[i] > 0 {
				u0.Sub(&t0.Y, &q.Y)
				u1.Sub(&t0.X, &q.X)
				u1.Inverse(&u1)
				u0.Mul(&u0, &u1)
				tab = append(tab, u0, t0.X)
				t0.Add(&t0, q)
			}
			if naf[i] < 0 {
				u0.Add(&t0.Y, &q.Y)
				u1.Sub(&q.X, &t0.X)
				u1.Inverse(&u1)
				u0.Mul(&u0, &u1)
				t0.Add(&t0, &qNeg)
				tab = append(tab, u0, t0.X)
			}
			i--
			continue
		}

		if naf[i] == 1 {
			tab = append(tab, t0.X, t0.Y)

			var lambda1, lambda2 fp.Element
			lambda1.Sub(&t0.Y, &q.Y)
			lambda2.Sub(&t0.X, &q.X)
			lambda2.Inverse(&lambda2)
			lambda1.Mul(&lambda1, &lambda2)

			t1.Add(&t0, q)
			lambda2.Sub(&t1.Y, &t0.Y)
			u0.Sub(&t1.X, &t0.X)
			u0.Inverse(&u0)
			lambda2.Mul(&lambda2, &u0)

			u0.Mul(&lambda1, &lambda2)
			lambda2.Add(&lambda1, &lambda2)
			u0.Add(&u0, &t0.X)
			u0.Add(&u0, &t1.X)
			tab = append(tab, u0, lambda2)

			t0.Add(&t1, &t0)
			i--
			continue
		}

		if naf[i] == -1 {
			tab = append(tab, t0.X, t0.Y)

			var lambda1, lambda2 fp.Element
			lambda1.Add(&t0.Y, &q.Y)
			lambda2.Sub(&t0.X, &q.X)
			lambda2.Inverse(&lambda2)
			lambda1.Mul(&lambda1, &lambda2)

			t1.Sub(&t0, q)
			lambda2.Sub(&t1.Y, &t0.Y)
			u0.Sub(&t1.X, &t0.X)
			u0.Inverse(&u0)
			lambda2.Mul(&lambda2, &u0)

			u0.Mul(&lambda1, &lambda2)
			lambda2.Add(&lambda1, &lambda2)
			u0.Add(&u0, &t0.X)
			u0.Add(&u0, &t1.X)
			tab = append(tab, u0, lambda2)

			t0.Add(&t1, &t0)
			i--
			continue
		}

		u0.Square(&t0.X)
		u1.Double(&u0)
		u0.Add(&u0, &u1)
		u1.Double(&t0.Y)
		u1.Inverse(&u1)
		u1.Neg(&u1)
		u0.Mul(&u0, &u1)
		tab = append(tab, u0)

		t0.Double(&t0)
		tab = append(tab, t0.X, t0.Y)
		i--
	}

	if naf[0] < 0 {
		tab = append(tab, t0.X)
	}

	return loopkupTable{
		q:   *q,
		tab: tab,
	}
}
