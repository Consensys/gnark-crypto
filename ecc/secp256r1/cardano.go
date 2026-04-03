package secp256r1

// Cardano solver for the depressed cubic x³ − 3x + c = 0 over secp256r1 Fp.
// Requires q ≡ 3 mod 4 (for Fp2 sqrt) and q ≡ 4 mod 9 (for Fp cbrt).

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
	fp2 "github.com/consensys/gnark-crypto/ecc/secp256r1/internal/fptower"
)

var omegaFq fp.Element // primitive cube root of unity in Fp

func init() {
	q := fp.Modulus()
	exp := new(big.Int).Sub(q, big.NewInt(1))
	exp.Div(exp, big.NewInt(3))
	var one fp.Element
	one.SetOne()
	for i := int64(2); ; i++ {
		var g, w fp.Element
		g.SetInt64(i)
		w.Exp(g, exp)
		if !w.Equal(&one) {
			omegaFq = w
			break
		}
	}
}

// CardanoRoots returns all roots in Fp of x³ − 3x + c = 0
// using Cardano's formula.
func CardanoRoots(c fp.Element) []fp.Element {
	var a fp.Element
	a.SetInt64(-3)

	var zero fp.Element

	// Δ = −4a³ − 27c²
	var a3, neg4a3, k27c2, delta fp.Element
	a3.Square(&a).Mul(&a3, &a)
	neg4a3.Mul(&a3, new(fp.Element).SetInt64(4)).Neg(&neg4a3)
	k27c2.Square(&c).Mul(&k27c2, new(fp.Element).SetInt64(27))
	delta.Sub(&neg4a3, &k27c2)

	// disc_D = c²/4 + a³/27
	var inv4, inv27, discD fp.Element
	inv4.SetInt64(4)
	inv4.Inverse(&inv4)
	inv27.SetInt64(27)
	inv27.Inverse(&inv27)
	discD.Square(&c).Mul(&discD, &inv4)
	var a3over27 fp.Element
	a3over27.Mul(&a3, &inv27)
	discD.Add(&discD, &a3over27)

	// −c/2
	var inv2, negCHalf fp.Element
	inv2.SetInt64(2)
	inv2.Inverse(&inv2)
	negCHalf.Mul(&c, &inv2).Neg(&negCHalf)

	om := omegaFq
	var om2 fp.Element
	om2.Square(&om)
	var one fp.Element
	one.SetOne()
	zetas := [3]fp.Element{one, om, om2}

	// Case 1: Δ = 0 (repeated root)
	if delta.Equal(&zero) {
		var invA, r0, r1 fp.Element
		invA.Inverse(&a)
		r0.Mul(&c, &invA).Mul(&r0, new(fp.Element).SetInt64(3))
		var twoA fp.Element
		twoA.Double(&a)
		r1.Inverse(&twoA).Mul(&r1, &c).Mul(&r1, new(fp.Element).SetInt64(3)).Neg(&r1)
		return []fp.Element{r0, r1}
	}

	// Case 2: Δ non-square → one real root via Fp2
	if delta.Legendre() == -1 {
		var discDE2, D fp2.E2
		discDE2.A0 = discD
		D.Sqrt(&discDE2)

		w := fp2.E2{A0: negCHalf, A1: D.A1}
		if w.IsZero() {
			w.A1.Neg(&D.A1)
		}

		var u fp2.E2
		u.Cbrt(&w)

		for _, zeta := range zetas {
			var cand fp2.E2
			cand.MulByElement(&u, &zeta)
			var inv fp2.E2
			inv.Inverse(&cand)
			var rRe, rIm fp.Element
			rRe.Add(&cand.A0, &inv.A0)
			rIm.Add(&cand.A1, &inv.A1)
			if rIm.Equal(&zero) {
				return []fp.Element{rRe}
			}
		}
		return []fp.Element{}
	}

	// Case 3: Δ square → 0 or 3 roots in Fp
	var DFq, wFq fp.Element
	DFq.Sqrt(&discD)
	wFq.Add(&negCHalf, &DFq)
	if wFq.Equal(&zero) {
		wFq.Sub(&negCHalf, &DFq)
	}

	var uFq fp.Element
	if uFq.Cbrt(&wFq) == nil {
		return []fp.Element{}
	}

	var invU, r0, r1, r2, t1, t2 fp.Element
	invU.Inverse(&uFq)
	r0.Add(&uFq, &invU)
	t1.Mul(&om, &uFq)
	t2.Mul(&om2, &invU)
	r1.Add(&t1, &t2)
	t1.Mul(&om2, &uFq)
	t2.Mul(&om, &invU)
	r2.Add(&t1, &t2)
	return []fp.Element{r0, r1, r2}
}
