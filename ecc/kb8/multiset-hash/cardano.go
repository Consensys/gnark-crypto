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
	e8Omega       extensions.E8
	e8Beta        extensions.E8
	e8One         extensions.E8
	e8Two         extensions.E8
	e8Three       extensions.E8
	e8Four        extensions.E8
	e8TwentySeven extensions.E8
	e8NegThree    extensions.E8
)

func init() {
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
	expByKBE8Cbrt(&y, x)
	var check extensions.E8
	check.Square(&y).Mul(&check, &y)
	if !check.Equal(x) {
		return nil
	}
	return z.Set(&y)
}

func expByKBE8Cbrt(z, x *extensions.E8) *extensions.E8 {
	// expByKBCbrt computation is derived from the addition chain:
	//
	//	_10       = 2*1
	//	_100      = 2*_10
	//	_110      = _10 + _100
	//	_1000     = _10 + _110
	//	_1010     = _10 + _1000
	//	_1011     = 1 + _1010
	//	_1100     = 1 + _1011
	//	_10110    = _1010 + _1100
	//	_11100    = _110 + _10110
	//	_11110    = _10 + _11100
	//	_11111    = 1 + _11110
	//	_101011   = _1100 + _11111
	//	_1000111  = _11100 + _101011
	//	_1001011  = _100 + _1000111
	//	_1010011  = _1000 + _1001011
	//	_1010101  = _10 + _1010011
	//	_1011001  = _100 + _1010101
	//	_1110111  = _11110 + _1011001
	//	_1111001  = _10 + _1110111
	//	_10001111 = _10110 + _1111001
	//	_10010101 = _110 + _10001111
	//	_10011101 = _1000 + _10010101
	//	_10100101 = _1000 + _10011101
	//	_10101111 = _1010 + _10100101
	//	_10110111 = _1000 + _10101111
	//	_11000011 = _1100 + _10110111
	//	_11001011 = _1000 + _11000011
	//	_11001101 = _10 + _11001011
	//	_11001111 = _10 + _11001101
	//	_11010101 = _110 + _11001111
	//	_11011101 = _1000 + _11010101
	//	i49       = ((_11001101 + _11011101) << 7 + _10110111) << 8 + _1011001
	//	i80       = ((i49 << 2 + 1) << 16 + _10011101) << 11
	//	i100      = ((_1001011 + i80) << 9 + _1000111) << 8 + _1010011
	//	i128      = ((i100 << 11 + _11001011) << 9 + _11001111) << 6
	//	i154      = ((_11111 + i128) << 15 + _10100101) << 8 + _10010101
	//	i182      = ((i154 << 9 + _10101111) << 8 + _1111001) << 9
	//	i200      = ((_10010101 + i182) << 8 + _11011101) << 7 + _1110111
	//	i228      = ((i200 << 9 + _11001101) << 8 + _11010101) << 9
	//	i251      = ((_11000011 + i228) << 8 + _101011) << 12 + _11011101
	//	i273      = ((_110 + i251) << 8 + _10001111) << 11 + _11010101
	//	i296      = ((i273 << 8 + _1010101) << 8 + _1010101) << 5
	//	return      _1011 + i296
	//
	// Operations: 239 squares 58 multiplies
	//
	// Generated by github.com/mmcloughlin/addchain v0.4.0.

	var (
		t0  extensions.E8
		t1  extensions.E8
		t2  extensions.E8
		t3  extensions.E8
		t4  extensions.E8
		t5  extensions.E8
		t6  extensions.E8
		t7  extensions.E8
		t8  extensions.E8
		t9  extensions.E8
		t10 extensions.E8
		t11 extensions.E8
		t12 extensions.E8
		t13 extensions.E8
		t14 extensions.E8
		t15 extensions.E8
		t16 extensions.E8
		t17 extensions.E8
		t18 extensions.E8
		t19 extensions.E8
		t20 extensions.E8
		t21 extensions.E8
		t22 extensions.E8
		t23 extensions.E8
		t24 extensions.E8
	)

	t0.Square(x)
	t1.Square(&t0)
	t2.Mul(&t0, &t1)
	t3.Mul(&t0, &t2)
	t4.Mul(&t0, &t3)
	t5.Mul(x, &t4)
	t6.Mul(x, &t5)
	t7.Mul(&t4, &t6)
	t8.Mul(&t2, &t7)
	t9.Mul(&t0, &t8)
	t10.Mul(x, &t9)
	t11.Mul(&t6, &t10)
	t8.Mul(&t8, &t11)
	t12.Mul(&t1, &t8)
	t13.Mul(&t3, &t12)
	t14.Mul(&t0, &t13)
	t1.Mul(&t1, &t14)
	t9.Mul(&t9, &t1)
	t15.Mul(&t0, &t9)
	t7.Mul(&t7, &t15)
	t16.Mul(&t2, &t7)
	t17.Mul(&t3, &t16)
	t18.Mul(&t3, &t17)
	t4.Mul(&t4, &t18)
	t19.Mul(&t3, &t4)
	t6.Mul(&t6, &t19)
	t20.Mul(&t3, &t6)
	t21.Mul(&t0, &t20)
	t0.Mul(&t0, &t21)
	t22.Mul(&t2, &t0)
	t3.Mul(&t3, &t22)
	t23.Mul(&t21, &t3)
	for s := 0; s < 7; s++ {
		t23.Square(&t23)
	}
	t19.Mul(&t19, &t23)
	for s := 0; s < 8; s++ {
		t19.Square(&t19)
	}
	t1.Mul(&t1, &t19)
	for s := 0; s < 2; s++ {
		t1.Square(&t1)
	}
	t24.Mul(x, &t1)
	for s := 0; s < 16; s++ {
		t24.Square(&t24)
	}
	t17.Mul(&t17, &t24)
	for s := 0; s < 11; s++ {
		t17.Square(&t17)
	}
	t12.Mul(&t12, &t17)
	for s := 0; s < 9; s++ {
		t12.Square(&t12)
	}
	t8.Mul(&t8, &t12)
	for s := 0; s < 8; s++ {
		t8.Square(&t8)
	}
	t13.Mul(&t13, &t8)
	for s := 0; s < 11; s++ {
		t13.Square(&t13)
	}
	t20.Mul(&t20, &t13)
	for s := 0; s < 9; s++ {
		t20.Square(&t20)
	}
	t0.Mul(&t0, &t20)
	for s := 0; s < 6; s++ {
		t0.Square(&t0)
	}
	t10.Mul(&t10, &t0)
	for s := 0; s < 15; s++ {
		t10.Square(&t10)
	}
	t18.Mul(&t18, &t10)
	for s := 0; s < 8; s++ {
		t18.Square(&t18)
	}
	t18.Mul(&t16, &t18)
	for s := 0; s < 9; s++ {
		t18.Square(&t18)
	}
	t4.Mul(&t4, &t18)
	for s := 0; s < 8; s++ {
		t4.Square(&t4)
	}
	t15.Mul(&t15, &t4)
	for s := 0; s < 9; s++ {
		t15.Square(&t15)
	}
	t16.Mul(&t16, &t15)
	for s := 0; s < 8; s++ {
		t16.Square(&t16)
	}
	t16.Mul(&t3, &t16)
	for s := 0; s < 7; s++ {
		t16.Square(&t16)
	}
	t9.Mul(&t9, &t16)
	for s := 0; s < 9; s++ {
		t9.Square(&t9)
	}
	t21.Mul(&t21, &t9)
	for s := 0; s < 8; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t22, &t21)
	for s := 0; s < 9; s++ {
		t21.Square(&t21)
	}
	t6.Mul(&t6, &t21)
	for s := 0; s < 8; s++ {
		t6.Square(&t6)
	}
	t11.Mul(&t11, &t6)
	for s := 0; s < 12; s++ {
		t11.Square(&t11)
	}
	t3.Mul(&t3, &t11)
	t2.Mul(&t2, &t3)
	for s := 0; s < 8; s++ {
		t2.Square(&t2)
	}
	t7.Mul(&t7, &t2)
	for s := 0; s < 11; s++ {
		t7.Square(&t7)
	}
	t22.Mul(&t22, &t7)
	for s := 0; s < 8; s++ {
		t22.Square(&t22)
	}
	t22.Mul(&t14, &t22)
	for s := 0; s < 8; s++ {
		t22.Square(&t22)
	}
	t14.Mul(&t14, &t22)
	for s := 0; s < 5; s++ {
		t14.Square(&t14)
	}
	z.Mul(&t5, &t14)

	return z
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

func (z *e16) expByKBCbrt(x *e16) *e16 {
	// expByKBCbrt computation is derived from the addition chain:
	//
	//	_10      = 2*1
	//	_11      = 1 + _10
	//	_101     = _10 + _11
	//	_110     = 1 + _101
	//	_111     = 1 + _110
	//	_1001    = _10 + _111
	//	_1011    = _10 + _1001
	//	_1101    = _10 + _1011
	//	_1111    = _10 + _1101
	//	_10001   = _10 + _1111
	//	_10101   = _110 + _1111
	//	_10111   = _10 + _10101
	//	_11001   = _10 + _10111
	//	_11011   = _10 + _11001
	//	_100001  = _110 + _11011
	//	_100011  = _10 + _100001
	//	_100101  = _10 + _100011
	//	_100111  = _10 + _100101
	//	_101001  = _10 + _100111
	//	_101011  = _10 + _101001
	//	_101101  = _10 + _101011
	//	_101111  = _10 + _101101
	//	_110001  = _10 + _101111
	//	_110011  = _10 + _110001
	//	_110101  = _10 + _110011
	//	_110111  = _10 + _110101
	//	_111001  = _10 + _110111
	//	_111011  = _10 + _111001
	//	_1100100 = _101001 + _111011
	//	_1111111 = _11011 + _1100100
	//	i53      = ((_1100100 << 5 + _1011) << 8 + _1111111) << 8
	//	i75      = ((_11001 + i53) << 12 + _101001) << 7 + _110111
	//	i97      = ((i75 << 6 + _101111) << 7 + _110001) << 7
	//	i112     = (2*(_110101 + i97) + 1) << 11 + _111001
	//	i137     = ((i112 << 8 + _100001) << 5 + _111) << 10
	//	i153     = ((_111011 + i137) << 6 + _111001) << 7 + _111011
	//	i175     = ((i153 << 8 + _110101) << 5 + _10001) << 7
	//	i197     = ((_111011 + i175) << 11 + _101111) << 8 + _110101
	//	i217     = ((2*i197 + 1) << 9 + _1111111) << 8
	//	i232     = ((_100101 + i217) << 7 + _100101) << 5 + _10001
	//	i254     = ((i232 << 5 + _1101) << 6 + _1011) << 9
	//	i269     = ((_100001 + i254) << 5 + _11001) << 7 + _101
	//	i298     = ((i269 << 10 + _101001) << 7 + _101001) << 10
	//	i315     = ((_11011 + i298) << 7 + _101011) << 7 + _110101
	//	i335     = ((2*i315 + 1) << 11 + _100101) << 6
	//	i355     = ((_101111 + i335) << 6 + _101) << 11 + _110011
	//	i382     = ((i355 << 7 + _101101) << 10 + _100111) << 8
	//	i397     = ((_10101 + i382) << 7 + _111001) << 5 + _11011
	//	i423     = ((i397 << 5 + _1111) << 11 + _1101) << 8
	//	i438     = ((_111011 + i423) << 4 + _1001) << 8 + _111001
	//	i461     = ((i438 << 7 + _111011) << 6 + _10111) << 8
	//	i477     = ((_1111111 + i461) << 7 + _101001) << 6 + _100011
	//	i501     = ((i477 << 5 + _10111) << 7 + _1001) << 10
	//	i513     = ((_110011 + i501) << 4 + _1001) << 5 + _11
	//	i539     = ((i513 << 11 + _1111) << 7 + _111) << 6
	//	i558     = ((_111 + i539) << 7 + _1111) << 9 + _110101
	//	i578     = ((i558 << 6 + _10101) << 6 + _10101) << 6
	//	return     (_10101 + i578) << 5 + _1011
	//
	// Operations: 487 squares 98 multiplies
	//
	// Generated by github.com/mmcloughlin/addchain v0.4.0.

	var (
		t0  e16
		t1  e16
		t2  e16
		t3  e16
		t4  e16
		t5  e16
		t6  e16
		t7  e16
		t8  e16
		t9  e16
		t10 e16
		t11 e16
		t12 e16
		t13 e16
		t14 e16
		t15 e16
		t16 e16
		t17 e16
		t18 e16
		t19 e16
		t20 e16
		t21 e16
		t22 e16
		t23 e16
		t24 e16
		t25 e16
		t26 e16
		t27 e16
		t28 e16
	)

	t0.Square(x)
	t1.Mul(x, &t0)
	t2.Mul(&t0, &t1)
	t3.Mul(x, &t2)
	t4.Mul(x, &t3)
	t5.Mul(&t0, &t4)
	t6.Mul(&t0, &t5)
	t7.Mul(&t0, &t6)
	t8.Mul(&t0, &t7)
	t9.Mul(&t0, &t8)
	t10.Mul(&t3, &t8)
	t11.Mul(&t0, &t10)
	t12.Mul(&t0, &t11)
	t13.Mul(&t0, &t12)
	t3.Mul(&t3, &t13)
	t14.Mul(&t0, &t3)
	t15.Mul(&t0, &t14)
	t16.Mul(&t0, &t15)
	t17.Mul(&t0, &t16)
	t18.Mul(&t0, &t17)
	t19.Mul(&t0, &t18)
	t20.Mul(&t0, &t19)
	t21.Mul(&t0, &t20)
	t22.Mul(&t0, &t21)
	t23.Mul(&t0, &t22)
	t24.Mul(&t0, &t23)
	t25.Mul(&t0, &t24)
	t0.Mul(&t0, &t25)
	t26.Mul(&t17, &t0)
	t27.Mul(&t13, &t26)
	for s := 0; s < 5; s++ {
		t26.Square(&t26)
	}
	t26.Mul(&t6, &t26)
	for s := 0; s < 8; s++ {
		t26.Square(&t26)
	}
	t26.Mul(&t27, &t26)
	for s := 0; s < 8; s++ {
		t26.Square(&t26)
	}
	t26.Mul(&t12, &t26)
	for s := 0; s < 12; s++ {
		t26.Square(&t26)
	}
	t26.Mul(&t17, &t26)
	for s := 0; s < 7; s++ {
		t26.Square(&t26)
	}
	t24.Mul(&t24, &t26)
	for s := 0; s < 6; s++ {
		t24.Square(&t24)
	}
	t24.Mul(&t20, &t24)
	for s := 0; s < 7; s++ {
		t24.Square(&t24)
	}
	t21.Mul(&t21, &t24)
	for s := 0; s < 7; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t23, &t21)
	t21.Square(&t21)
	t21.Mul(x, &t21)
	for s := 0; s < 11; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t25, &t21)
	for s := 0; s < 8; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t3, &t21)
	for s := 0; s < 5; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t4, &t21)
	for s := 0; s < 10; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t0, &t21)
	for s := 0; s < 6; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t25, &t21)
	for s := 0; s < 7; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t0, &t21)
	for s := 0; s < 8; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t23, &t21)
	for s := 0; s < 5; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t9, &t21)
	for s := 0; s < 7; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t0, &t21)
	for s := 0; s < 11; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t20, &t21)
	for s := 0; s < 8; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t23, &t21)
	t21.Square(&t21)
	t21.Mul(x, &t21)
	for s := 0; s < 9; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t27, &t21)
	for s := 0; s < 8; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t15, &t21)
	for s := 0; s < 7; s++ {
		t21.Square(&t21)
	}
	t21.Mul(&t15, &t21)
	for s := 0; s < 5; s++ {
		t21.Square(&t21)
	}
	t9.Mul(&t9, &t21)
	for s := 0; s < 5; s++ {
		t9.Square(&t9)
	}
	t9.Mul(&t7, &t9)
	for s := 0; s < 6; s++ {
		t9.Square(&t9)
	}
	t9.Mul(&t6, &t9)
	for s := 0; s < 9; s++ {
		t9.Square(&t9)
	}
	t3.Mul(&t3, &t9)
	for s := 0; s < 5; s++ {
		t3.Square(&t3)
	}
	t12.Mul(&t12, &t3)
	for s := 0; s < 7; s++ {
		t12.Square(&t12)
	}
	t12.Mul(&t2, &t12)
	for s := 0; s < 10; s++ {
		t12.Square(&t12)
	}
	t12.Mul(&t17, &t12)
	for s := 0; s < 7; s++ {
		t12.Square(&t12)
	}
	t12.Mul(&t17, &t12)
	for s := 0; s < 10; s++ {
		t12.Square(&t12)
	}
	t12.Mul(&t13, &t12)
	for s := 0; s < 7; s++ {
		t12.Square(&t12)
	}
	t18.Mul(&t18, &t12)
	for s := 0; s < 7; s++ {
		t18.Square(&t18)
	}
	t18.Mul(&t23, &t18)
	t18.Square(&t18)
	t28.Mul(x, &t18)
	for s := 0; s < 11; s++ {
		t28.Square(&t28)
	}
	t15.Mul(&t15, &t28)
	for s := 0; s < 6; s++ {
		t15.Square(&t15)
	}
	t20.Mul(&t20, &t15)
	for s := 0; s < 6; s++ {
		t20.Square(&t20)
	}
	t2.Mul(&t2, &t20)
	for s := 0; s < 11; s++ {
		t2.Square(&t2)
	}
	t2.Mul(&t22, &t2)
	for s := 0; s < 7; s++ {
		t2.Square(&t2)
	}
	t19.Mul(&t19, &t2)
	for s := 0; s < 10; s++ {
		t19.Square(&t19)
	}
	t16.Mul(&t16, &t19)
	for s := 0; s < 8; s++ {
		t16.Square(&t16)
	}
	t16.Mul(&t10, &t16)
	for s := 0; s < 7; s++ {
		t16.Square(&t16)
	}
	t16.Mul(&t25, &t16)
	for s := 0; s < 5; s++ {
		t16.Square(&t16)
	}
	t13.Mul(&t13, &t16)
	for s := 0; s < 5; s++ {
		t13.Square(&t13)
	}
	t13.Mul(&t8, &t13)
	for s := 0; s < 11; s++ {
		t13.Square(&t13)
	}
	t7.Mul(&t7, &t13)
	for s := 0; s < 8; s++ {
		t7.Square(&t7)
	}
	t7.Mul(&t0, &t7)
	for s := 0; s < 4; s++ {
		t7.Square(&t7)
	}
	t7.Mul(&t5, &t7)
	for s := 0; s < 8; s++ {
		t7.Square(&t7)
	}
	t25.Mul(&t25, &t7)
	for s := 0; s < 7; s++ {
		t25.Square(&t25)
	}
	t0.Mul(&t0, &t25)
	for s := 0; s < 6; s++ {
		t0.Square(&t0)
	}
	t0.Mul(&t11, &t0)
	for s := 0; s < 8; s++ {
		t0.Square(&t0)
	}
	t27.Mul(&t27, &t0)
	for s := 0; s < 7; s++ {
		t27.Square(&t27)
	}
	t17.Mul(&t17, &t27)
	for s := 0; s < 6; s++ {
		t17.Square(&t17)
	}
	t14.Mul(&t14, &t17)
	for s := 0; s < 5; s++ {
		t14.Square(&t14)
	}
	t11.Mul(&t11, &t14)
	for s := 0; s < 7; s++ {
		t11.Square(&t11)
	}
	t11.Mul(&t5, &t11)
	for s := 0; s < 10; s++ {
		t11.Square(&t11)
	}
	t22.Mul(&t22, &t11)
	for s := 0; s < 4; s++ {
		t22.Square(&t22)
	}
	t5.Mul(&t5, &t22)
	for s := 0; s < 5; s++ {
		t5.Square(&t5)
	}
	t1.Mul(&t1, &t5)
	for s := 0; s < 11; s++ {
		t1.Square(&t1)
	}
	t1.Mul(&t8, &t1)
	for s := 0; s < 7; s++ {
		t1.Square(&t1)
	}
	t1.Mul(&t4, &t1)
	for s := 0; s < 6; s++ {
		t1.Square(&t1)
	}
	t4.Mul(&t4, &t1)
	for s := 0; s < 7; s++ {
		t4.Square(&t4)
	}
	t8.Mul(&t8, &t4)
	for s := 0; s < 9; s++ {
		t8.Square(&t8)
	}
	t23.Mul(&t23, &t8)
	for s := 0; s < 6; s++ {
		t23.Square(&t23)
	}
	t23.Mul(&t10, &t23)
	for s := 0; s < 6; s++ {
		t23.Square(&t23)
	}
	t23.Mul(&t10, &t23)
	for s := 0; s < 6; s++ {
		t23.Square(&t23)
	}
	t10.Mul(&t10, &t23)
	for s := 0; s < 5; s++ {
		t10.Square(&t10)
	}
	z.Mul(&t6, &t10)

	return z
}

func (z *e16) Cbrt(x *e16) *e16 {
	var y e16
	y.expByKBCbrt(x)
	var check e16
	check.Square(&y).Mul(&check, &y)
	if !check.A0.Equal(&x.A0) || !check.A1.Equal(&x.A1) {
		return nil
	}
	return z.Set(&y)
}
