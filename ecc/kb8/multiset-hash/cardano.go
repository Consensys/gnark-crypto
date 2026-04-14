package multisethash

import (
	"math/big"
	"slices"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// Cardano solver for the depressed cubic x^3 - 3x + c = 0 over kb8 Fp^8.
// This mirrors the structure of the secp256r1 Cardano solver in PR #831:
// repeated-root case, square-discriminant case over the base field, and a
// quadratic-extension fallback when the discriminant is not a square.

var (
	e8CbrtExponent  big.Int
	e16CbrtExponent big.Int
	e8Omega         extensions.E8
	e8Beta          extensions.E8
	e8One           extensions.E8
	e8Two           extensions.E8
	e8Three         extensions.E8
	e8Four          extensions.E8
	e8TwentySeven   extensions.E8
	e8NegThree      extensions.E8
)

func init() {
	var q8, q16 big.Int
	q8.Exp(koalabear.Modulus(), big.NewInt(8), nil)
	q16.Mul(&q8, &q8)

	// q8 ≡ 4 (mod 9), so cubic residues admit x^((2q8+1)/9) as a cube root.
	e8CbrtExponent.Mul(&q8, big.NewInt(2))
	e8CbrtExponent.Add(&e8CbrtExponent, big.NewInt(1))
	e8CbrtExponent.Div(&e8CbrtExponent, big.NewInt(9))

	// q16 ≡ 7 (mod 9), so cubic residues admit x^((q16+2)/9) as a cube root.
	e16CbrtExponent.Add(&q16, big.NewInt(2))
	e16CbrtExponent.Div(&e16CbrtExponent, big.NewInt(9))

	e8One.SetOne()
	e8Two.C0.B0.A0.SetUint64(2)
	e8Three.C0.B0.A0.SetUint64(3)
	e8Four.C0.B0.A0.SetUint64(4)
	e8TwentySeven.C0.B0.A0.SetUint64(27)
	e8NegThree.Set(&e8Three).Neg(&e8NegThree)

	e8Beta = findNonSquare()
	e8Omega = findPrimitiveCubeRoot()
}

func depressedCubicRoot(c extensions.E8) (extensions.E8, bool) {
	roots := cardanoRoots(c)
	if len(roots) == 0 {
		return extensions.E8{}, false
	}
	slices.SortFunc(roots, func(a, b extensions.E8) int {
		return a.Cmp(&b)
	})
	return roots[0], true
}

func cardanoRoots(c extensions.E8) []extensions.E8 {
	var a3, neg4a3, k27c2, delta extensions.E8
	a3.Square(&e8NegThree).Mul(&a3, &e8NegThree)
	neg4a3.Mul(&a3, &e8Four).Neg(&neg4a3)
	k27c2.Square(&c).Mul(&k27c2, &e8TwentySeven)
	delta.Sub(&neg4a3, &k27c2)

	var inv2, inv4, inv27 extensions.E8
	inv2.Inverse(&e8Two)
	inv4.Inverse(&e8Four)
	inv27.Inverse(&e8TwentySeven)

	var discD, a3Over27 extensions.E8
	discD.Square(&c).Mul(&discD, &inv4)
	a3Over27.Mul(&a3, &inv27)
	discD.Add(&discD, &a3Over27)

	var negCHalf extensions.E8
	negCHalf.Mul(&c, &inv2).Neg(&negCHalf)

	if delta.IsZero() {
		return repeatedRoots(c)
	}

	if delta.Legendre() == -1 {
		return cardanoRootsViaQuadraticExtension(negCHalf, discD)
	}

	return cardanoRootsBaseField(negCHalf, discD)
}

func repeatedRoots(c extensions.E8) []extensions.E8 {
	var invA, r0, r1, twoA extensions.E8
	invA.Inverse(&e8NegThree)
	r0.Mul(&c, &invA).Mul(&r0, &e8Three)
	twoA.Double(&e8NegThree)
	r1.Inverse(&twoA).Mul(&r1, &c).Mul(&r1, &e8Three).Neg(&r1)
	return dedupRoots([]extensions.E8{r0, r1})
}

func cardanoRootsBaseField(negCHalf, discD extensions.E8) []extensions.E8 {
	var d, w extensions.E8
	d.Sqrt(&discD)
	w.Add(&negCHalf, &d)
	if w.IsZero() {
		w.Sub(&negCHalf, &d)
	}

	var u extensions.E8
	if cbrtE8(&u, &w) == nil {
		return nil
	}

	var omega2 extensions.E8
	omega2.Square(&e8Omega)

	var invU, r0, r1, r2, t1, t2 extensions.E8
	invU.Inverse(&u)
	r0.Add(&u, &invU)
	t1.Mul(&e8Omega, &u)
	t2.Mul(&omega2, &invU)
	r1.Add(&t1, &t2)
	t1.Mul(&omega2, &u)
	t2.Mul(&e8Omega, &invU)
	r2.Add(&t1, &t2)

	return filterValidRoots(negCHalf, []extensions.E8{r0, r1, r2})
}

func cardanoRootsViaQuadraticExtension(negCHalf, discD extensions.E8) []extensions.E8 {
	var discOverBeta, sqrtDiscOverBeta extensions.E8
	discOverBeta.Div(&discD, &e8Beta)
	if discOverBeta.Legendre() != 1 {
		return nil
	}
	sqrtDiscOverBeta.Sqrt(&discOverBeta)

	w := e16{
		A0: negCHalf,
		A1: sqrtDiscOverBeta,
	}

	var u e16
	if u.Cbrt(&w) == nil {
		return nil
	}

	var omega2 extensions.E8
	omega2.Square(&e8Omega)
	zetas := [3]extensions.E8{e8One, e8Omega, omega2}

	for _, zeta := range zetas {
		var cand, inv, sum e16
		cand.MulByE8(&u, &zeta)
		inv.Inverse(&cand)
		sum.Add(&cand, &inv)
		if sum.A1.IsZero() && isDepressedCubicRoot(&sum.A0, &negCHalf) {
			return []extensions.E8{sum.A0}
		}
	}

	return nil
}

func filterValidRoots(negCHalf extensions.E8, roots []extensions.E8) []extensions.E8 {
	res := make([]extensions.E8, 0, len(roots))
	for _, root := range roots {
		if isDepressedCubicRoot(&root, &negCHalf) {
			res = append(res, root)
		}
	}
	return dedupRoots(res)
}

func dedupRoots(roots []extensions.E8) []extensions.E8 {
	if len(roots) == 0 {
		return nil
	}
	slices.SortFunc(roots, func(a, b extensions.E8) int {
		return a.Cmp(&b)
	})
	out := roots[:1]
	for i := 1; i < len(roots); i++ {
		if !roots[i].Equal(&out[len(out)-1]) {
			out = append(out, roots[i])
		}
	}
	return out
}

func isDepressedCubicRoot(x, negCHalf *extensions.E8) bool {
	var lhs, rhs extensions.E8
	lhs.Square(x).Mul(&lhs, x)
	rhs.Double(negCHalf).Neg(&rhs)
	lhs.Sub(&lhs, x).Sub(&lhs, x).Sub(&lhs, x).Add(&lhs, &rhs)
	return lhs.IsZero()
}

func cbrtE8(z, x *extensions.E8) *extensions.E8 {
	var y extensions.E8
	y.Exp(*x, &e8CbrtExponent)
	var check extensions.E8
	check.Square(&y).Mul(&check, &y)
	if !check.Equal(x) {
		return nil
	}
	return z.Set(&y)
}

func findPrimitiveCubeRoot() extensions.E8 {
	var exp big.Int
	exp.Exp(koalabear.Modulus(), big.NewInt(8), nil)
	exp.Sub(&exp, big.NewInt(1))
	exp.Div(&exp, big.NewInt(3))

	for _, candidate := range e8SearchCandidates() {
		var w extensions.E8
		w.Exp(candidate, &exp)
		if !w.IsOne() {
			return w
		}
	}
	panic("kb8 multiset hash: failed to find primitive cube root in Fp^8")
}

func findNonSquare() extensions.E8 {
	for _, candidate := range e8SearchCandidates() {
		if !candidate.IsZero() && candidate.Legendre() == -1 {
			return candidate
		}
	}
	panic("kb8 multiset hash: failed to find quadratic non-residue in Fp^8")
}

func e8SearchCandidates() []extensions.E8 {
	const searchSpace = 6560 // 3^8 - 1
	res := make([]extensions.E8, 0, searchSpace)
	for n := 1; n <= searchSpace; n++ {
		res = append(res, ternaryCandidate(n))
	}
	return res
}

func ternaryCandidate(n int) extensions.E8 {
	var x extensions.E8
	for i := 0; i < 8; i++ {
		v := uint64(n % 3)
		n /= 3
		if v != 0 {
			setE8Coeff(&x, i, v)
		}
	}
	return x
}

func setE8Coeff(x *extensions.E8, idx int, v uint64) {
	switch idx {
	case 0:
		x.C0.B0.A0.SetUint64(v)
	case 1:
		x.C0.B0.A1.SetUint64(v)
	case 2:
		x.C0.B1.A0.SetUint64(v)
	case 3:
		x.C0.B1.A1.SetUint64(v)
	case 4:
		x.C1.B0.A0.SetUint64(v)
	case 5:
		x.C1.B0.A1.SetUint64(v)
	case 6:
		x.C1.B1.A0.SetUint64(v)
	case 7:
		x.C1.B1.A1.SetUint64(v)
	default:
		panic("invalid E8 coefficient index")
	}
}

type e16 struct {
	A0, A1 extensions.E8
}

func (z *e16) Set(x *e16) *e16 {
	z.A0.Set(&x.A0)
	z.A1.Set(&x.A1)
	return z
}

func (z *e16) Add(x, y *e16) *e16 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	return z
}

func (z *e16) Mul(x, y *e16) *e16 {
	var a, b, c, d extensions.E8
	a.Mul(&x.A0, &y.A0)
	b.Mul(&x.A1, &y.A1).Mul(&b, &e8Beta)
	c.Add(&x.A0, &x.A1)
	d.Add(&y.A0, &y.A1)
	c.Mul(&c, &d).Sub(&c, &a)
	var rawB extensions.E8
	rawB.Mul(&x.A1, &y.A1)
	c.Sub(&c, &rawB)
	z.A0.Add(&a, &b)
	z.A1.Set(&c)
	return z
}

func (z *e16) Square(x *e16) *e16 {
	return z.Mul(x, x)
}

func (z *e16) Inverse(x *e16) *e16 {
	var t0, t1, denom extensions.E8
	t0.Square(&x.A0)
	t1.Square(&x.A1).Mul(&t1, &e8Beta)
	denom.Sub(&t0, &t1).Inverse(&denom)
	z.A0.Mul(&x.A0, &denom)
	z.A1.Mul(&x.A1, &denom).Neg(&z.A1)
	return z
}

func (z *e16) MulByE8(x *e16, y *extensions.E8) *e16 {
	z.A0.Mul(&x.A0, y)
	z.A1.Mul(&x.A1, y)
	return z
}

func (z *e16) Exp(x e16, k *big.Int) *e16 {
	z.A0.SetOne()
	z.A1.SetZero()
	for _, b := range k.Bytes() {
		for bit := 7; bit >= 0; bit-- {
			z.Square(z)
			if (b>>bit)&1 == 1 {
				z.Mul(z, &x)
			}
		}
	}
	return z
}

func (z *e16) Cbrt(x *e16) *e16 {
	var y e16
	y.Exp(*x, &e16CbrtExponent)
	var check e16
	check.Square(&y).Mul(&check, &y)
	if !check.A0.Equal(&x.A0) || !check.A1.Equal(&x.A1) {
		return nil
	}
	return z.Set(&y)
}
