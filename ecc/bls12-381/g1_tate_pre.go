package bls12381

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

const (
	g1TateGenX = "0xD82B23C3EE86C6B55930A7755FEB499A697AAE08D97E677F61EBF6894E57EC7434DA198FE1FBF0EF1C7004640A74203"
	g1TateGenY = "0x75868854578CF684F73F747280EF3F0A86CD94B3FB5954BC8B6FA4888BE7B2FB766E6DAF6F4F0AB9FE3E757B4BE8404"
)

var (
	g1TateGenOnce sync.Once
	g1TateGenAff  G1Affine

	g1TatePreTableOnce sync.Once
	g1TatePreTable     G1TatePreTable

	tateNafOnce   sync.Once
	tateNafDigits [tateNafMaxLen]int8
	tateNafLen    int

	tateExp1Once     sync.Once
	tateExp1Exponent big.Int
)

// G1TatePreTable holds the precomputation table and its associated auxiliary point.
type G1TatePreTable struct {
	Q   G1Affine
	Tab []fp.Element
}

const tateNafMaxLen = 65

// IsInSubGroupTatePre checks whether p is in the correct subgroup using the Tate test with precomputation.
func (p *G1Affine) IsInSubGroupTatePre(tab G1TatePreTable) bool {
	return G1IsValidTatePre(p, &tab.Q, tab.Tab)
}

// G1IsValidTatePre checks whether a is in the correct subgroup using the Tate test with precomputation.
func G1IsValidTatePre(a, q *G1Affine, tab []fp.Element) bool {
	if a.IsInfinity() {
		return false
	}
	if !a.IsOnCurve() {
		return false
	}
	return testTatePre(a, q, tab)
}

func testTatePre(p, q *G1Affine, tab []fp.Element) bool {
	var p2 G1Affine
	p2.X.Mul(&p.X, &thirdRootOneG1)
	p2.Y.Set(&p.Y)

	var n1, d1, n2, d2 fp.Element
	n1.Sub(&p.X, &q.X)
	n2.Sub(&p2.X, &q.X)
	d1.SetOne()
	d2.SetOne()
	tateMillerPre(tab, &n1, &d1, &n2, &d2, q, p, &p2)

	if n1.IsZero() || d1.IsZero() || n2.IsZero() || d2.IsZero() {
		return false
	}

	n1.Mul(&n1, &d2)
	n2.Mul(&n2, &d1)
	d1.Mul(&d1, &d2)
	d1.Inverse(&d1)
	n1.Mul(&n1, &d1)
	n2.Mul(&n2, &d1)
	tateExp1(&n1)
	tateExp2(&n2, &d2, &n2)
	return n2.Equal(&d2) && n1.IsOne()
}

func tateMillerPre(tab []fp.Element, n1, d1, n2, d2 *fp.Element, q, p, p2 *G1Affine) {
	s := tateNAFDigits()
	i := len(s) - 2
	j := 0
	k := 0
	if s[0] < 0 {
		j = 1
	}

	var f1, g1, f2, g2 fp.Element
	var u0, u1, u2, u3 fp.Element
	var v0, v1, v2 fp.Element

	u1.Sub(&p.Y, &q.Y)
	u2.Sub(&p.X, &q.X)
	u3.Sub(&p2.X, &q.X)

	for i >= j {
		if s[i] == 0 && i > j {
			u0.Sub(&p.Y, &tab[k+2])
			f1.Sub(&p.X, &tab[k+1])
			f2.Sub(&p2.X, &tab[k+1])
			g1.Mul(&f1, &tab[k])
			g1.Sub(&u0, &g1)
			g2.Mul(&f2, &tab[k])
			g2.Sub(&u0, &g2)
			f1.Mul(&f1, &tab[k+3])
			f1.Sub(&u0, &f1)
			f2.Mul(&f2, &tab[k+3])
			f2.Sub(&u0, &f2)

			n1.Square(n1)
			n1.Square(n1)
			n1.Mul(n1, &f1)
			d1.Square(d1)
			d1.Mul(d1, &g1)
			d1.Square(d1)

			n2.Square(n2)
			n2.Square(n2)
			n2.Mul(n2, &f2)
			d2.Square(d2)
			d2.Mul(d2, &g2)
			d2.Square(d2)

			k += 4
			i--

			if s[i] > 0 {
				f1.Mul(&tab[k], &u2)
				f1.Sub(&u1, &f1)
				f2.Mul(&tab[k], &u3)
				f2.Sub(&u1, &f2)
				g1.Sub(&p.X, &tab[k+1])
				g2.Sub(&p2.X, &tab[k+1])

				n1.Mul(n1, &f1)
				d1.Mul(d1, &g1)
				n2.Mul(n2, &f2)
				d2.Mul(d2, &g2)
				k += 2
			}

			if s[i] < 0 {
				f1.Sub(&p.X, &tab[k+1])
				f2.Sub(&p2.X, &tab[k+1])
				g1.Mul(&tab[k], &u2)
				g1.Sub(&u1, &g1)
				g2.Mul(&tab[k], &u3)
				g2.Sub(&u1, &g2)

				n1.Mul(n1, &f1)
				d1.Mul(d1, &g1)
				n2.Mul(n2, &f2)
				d2.Mul(d2, &g2)
				k += 2
			}
			i--
			continue
		}

		if s[i] == 1 {
			u0.Sub(&p.X, &tab[k])
			g1.Sub(&p2.X, &tab[k])
			g2.Add(&p.X, &tab[k+2])
			f1.Add(&p2.X, &tab[k+2])
			v0.Mul(&u0, &g2)
			v1.Mul(&g1, &f1)
			g2.Sub(&p.Y, &tab[k+1])
			v2.Mul(&g2, &tab[k+3])
			v0.Sub(&v0, &v2)
			v1.Sub(&v1, &v2)
			f1.Set(&v0)
			f2.Set(&v1)

			n1.Square(n1)
			n1.Mul(n1, &f1)
			d1.Mul(d1, &u0)
			d1.Square(d1)

			n2.Square(n2)
			n2.Mul(n2, &f2)
			d2.Mul(d2, &g1)
			d2.Square(d2)

			k += 4
			i--
			continue
		}

		if s[i] == -1 {
			u0.Sub(&p.X, &tab[k])
			g1.Sub(&p2.X, &tab[k])
			g2.Add(&p.X, &tab[k+2])
			f1.Add(&p2.X, &tab[k+2])
			v0.Mul(&u0, &g2)
			v1.Mul(&g1, &f1)
			g2.Sub(&p.Y, &tab[k+1])
			v2.Mul(&g2, &tab[k+3])
			v0.Sub(&v0, &v2)
			v1.Sub(&v1, &v2)
			f1.Set(&v0)
			f2.Set(&v1)

			n1.Square(n1)
			n1.Mul(n1, &f1)
			d1.Mul(d1, &u0)
			d1.Square(d1)
			d1.Mul(d1, &u2)

			n2.Square(n2)
			n2.Mul(n2, &f2)
			d2.Mul(d2, &g1)
			d2.Square(d2)
			d2.Mul(d2, &u3)

			k += 4
			i--
			continue
		}

		f1.Sub(&p.X, &tab[k+1])
		f2.Sub(&p2.X, &tab[k+1])
		u0.Sub(&p.Y, &tab[k+2])
		g1.Mul(&tab[k], &f1)
		g1.Sub(&u0, &g1)
		g2.Mul(&tab[k], &f2)
		g2.Sub(&u0, &g2)

		n1.Square(n1)
		n1.Mul(n1, &f1)
		d1.Square(d1)
		d1.Mul(d1, &g1)
		n2.Square(n2)
		n2.Mul(n2, &f2)
		d2.Square(d2)
		d2.Mul(d2, &g2)

		k += 3
		i--
	}

	if s[0] < 0 {
		g1.Sub(&p.X, &tab[k])
		g2.Sub(&p2.X, &tab[k])
		n1.Square(n1)
		d1.Square(d1)
		d1.Mul(d1, &g1)
		n2.Square(n2)
		d2.Square(d2)
		d2.Mul(d2, &g2)
	}
}

// G1TateGen returns the auxiliary point used by the Tate-based G1 membership test.
func G1TateGen() G1Affine {
	g1TateGenOnce.Do(func() {
		if _, err := g1TateGenAff.X.SetString(g1TateGenX); err != nil {
			panic(err)
		}
		if _, err := g1TateGenAff.Y.SetString(g1TateGenY); err != nil {
			panic(err)
		}
	})
	return g1TateGenAff
}

// G1TatePreTableDefault returns a cached precomputation table for G1TateGen.
func G1TatePreTableDefault() G1TatePreTable {
	g1TatePreTableOnce.Do(func() {
		q := G1TateGen()
		g1TatePreTable = G1MillerTab(&q)
	})
	return g1TatePreTable
}

// G1MillerTab precomputes the lookup table used by the Tate-based G1 membership test.
func G1MillerTab(q *G1Affine) G1TatePreTable {
	s := tateNAFDigits()
	i := len(s) - 2
	j := 0
	if s[0] < 0 {
		j = 1
	}

	tab := make([]fp.Element, 0, len(s)*4)
	var t0, t1, qNeg G1Affine
	t0.Set(q)
	qNeg.Neg(q)

	var u0, u1 fp.Element

	for i >= j {
		if s[i] == 0 && i > j {
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

			if s[i] > 0 {
				u0.Sub(&t0.Y, &q.Y)
				u1.Sub(&t0.X, &q.X)
				u1.Inverse(&u1)
				u0.Mul(&u0, &u1)
				tab = append(tab, u0, t0.X)
				t0.Add(&t0, q)
			}
			if s[i] < 0 {
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

		if s[i] == 1 {
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

		if s[i] == -1 {
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

	if s[0] < 0 {
		tab = append(tab, t0.X)
	}

	return G1TatePreTable{
		Q:   *q,
		Tab: tab,
	}
}

func tateExp1(x *fp.Element) {
	// Operations: 311 squares 70 multiplies
	//
	// Generated by github.com/mmcloughlin/addchain v0.4.0.

	// Allocate Temporaries.
	var z fp.Element
	var (
		t0  = new(fp.Element)
		t1  = new(fp.Element)
		t2  = new(fp.Element)
		t3  = new(fp.Element)
		t4  = new(fp.Element)
		t5  = new(fp.Element)
		t6  = new(fp.Element)
		t7  = new(fp.Element)
		t8  = new(fp.Element)
		t9  = new(fp.Element)
		t10 = new(fp.Element)
		t11 = new(fp.Element)
		t12 = new(fp.Element)
		t13 = new(fp.Element)
		t14 = new(fp.Element)
		t15 = new(fp.Element)
		t16 = new(fp.Element)
		t17 = new(fp.Element)
		t18 = new(fp.Element)
		t19 = new(fp.Element)
		t20 = new(fp.Element)
		t21 = new(fp.Element)
		t22 = new(fp.Element)
		t23 = new(fp.Element)
		t24 = new(fp.Element)
		t25 = new(fp.Element)
		t26 = new(fp.Element)
	)

	// Step 1: t2 = x^0x2
	t2.Square(x)

	// Step 2: t3 = x^0x4
	t3.Square(t2)

	// Step 3: t5 = x^0x6
	t5.Mul(t2, t3)

	// Step 4: t22 = x^0x7
	t22.Mul(x, t5)

	// Step 5: t8 = x^0xa
	t8.Mul(t3, t5)

	// Step 6: t24 = x^0xe
	t24.Mul(t3, t8)

	// Step 7: t1 = x^0x10
	t1.Mul(t2, t24)

	// Step 8: z = x^0x15
	z.Mul(t22, t24)

	// Step 9: t4 = x^0x18
	t4.Mul(t8, t24)

	// Step 10: t0 = x^0x1a
	t0.Mul(t2, t4)

	// Step 11: t16 = x^0x1f
	t16.Mul(t22, t4)

	// Step 12: t15 = x^0x22
	t15.Mul(t8, t4)

	// Step 13: t21 = x^0x23
	t21.Mul(x, t15)

	// Step 14: t20 = x^0x3d
	t20.Mul(t0, t21)

	// Step 15: t0 = x^0x55
	t0.Mul(t4, t20)

	// Step 16: t6 = x^0x6d
	t6.Mul(t4, t0)

	// Step 17: t23 = x^0x6f
	t23.Mul(t2, t6)

	// Step 18: t11 = x^0x7f
	t11.Mul(t1, t23)

	// Step 19: t19 = x^0x81
	t19.Mul(t2, t11)

	// Step 20: t17 = x^0x8b
	t17.Mul(t8, t19)

	// Step 21: t10 = x^0x8d
	t10.Mul(t2, t17)

	// Step 22: t7 = x^0x95
	t7.Mul(t8, t17)

	// Step 23: t4 = x^0x97
	t4.Mul(t2, t7)

	// Step 24: t25 = x^0x99
	t25.Mul(t2, t4)

	// Step 25: t9 = x^0x9b
	t9.Mul(t2, t25)

	// Step 26: t12 = x^0xa5
	t12.Mul(t8, t9)

	// Step 27: t1 = x^0xab
	t1.Mul(t5, t12)

	// Step 28: t14 = x^0xb5
	t14.Mul(t8, t1)

	// Step 29: t13 = x^0xbb
	t13.Mul(t5, t14)

	// Step 30: t8 = x^0xbd
	t8.Mul(t2, t13)

	// Step 31: t2 = x^0xbf
	t2.Mul(t2, t8)

	// Step 32: t5 = x^0xc5
	t5.Mul(t5, t2)

	// Step 33: t18 = x^0xe7
	t18.Mul(t15, t5)

	// Step 34: t15 = x^0xeb
	t15.Mul(t3, t18)

	// Step 35: t24 = x^0xf9
	t24.Mul(t24, t15)

	// Step 36: t3 = x^0xfd
	t3.Mul(t3, t24)

	// Step 44: t26 = x^0xfd00
	t26.Square(t3)
	for s := 1; s < 8; s++ {
		t26.Square(t26)
	}

	// Step 45: t26 = x^0xfd99
	t26.Mul(t25, t26)

	// Step 56: t26 = x^0x7ecc800
	for s := 0; s < 11; s++ {
		t26.Square(t26)
	}

	// Step 57: t25 = x^0x7ecc899
	t25.Mul(t25, t26)

	// Step 66: t25 = x^0xfd9913200
	for s := 0; s < 9; s++ {
		t25.Square(t25)
	}

	// Step 67: t25 = x^0xfd99132a5
	t25.Mul(t12, t25)

	// Step 77: t25 = x^0x3f6644ca9400
	for s := 0; s < 10; s++ {
		t25.Square(t25)
	}

	// Step 78: t24 = x^0x3f6644ca94f9
	t24.Mul(t24, t25)

	// Step 85: t24 = x^0x1fb322654a7c80
	for s := 0; s < 7; s++ {
		t24.Square(t24)
	}

	// Step 86: t23 = x^0x1fb322654a7cef
	t23.Mul(t23, t24)

	// Step 90: t23 = x^0x1fb322654a7cef0
	for s := 0; s < 4; s++ {
		t23.Square(t23)
	}

	// Step 91: t22 = x^0x1fb322654a7cef7
	t22.Mul(t22, t23)

	// Step 102: t22 = x^0xfd99132a53e77b800
	for s := 0; s < 11; s++ {
		t22.Square(t22)
	}

	// Step 103: t21 = x^0xfd99132a53e77b823
	t21.Mul(t21, t22)

	// Step 114: t21 = x^0x7ecc899529f3bdc11800
	for s := 0; s < 11; s++ {
		t21.Square(t21)
	}

	// Step 115: t21 = x^0x7ecc899529f3bdc118bd
	t21.Mul(t8, t21)

	// Step 121: t21 = x^0x1fb322654a7cef70462f40
	for s := 0; s < 6; s++ {
		t21.Square(t21)
	}

	// Step 122: t20 = x^0x1fb322654a7cef70462f7d
	t20.Mul(t20, t21)

	// Step 132: t20 = x^0x7ecc899529f3bdc118bdf400
	for s := 0; s < 10; s++ {
		t20.Square(t20)
	}

	// Step 133: t19 = x^0x7ecc899529f3bdc118bdf481
	t19.Mul(t19, t20)

	// Step 142: t19 = x^0xfd99132a53e77b82317be90200
	for s := 0; s < 9; s++ {
		t19.Square(t19)
	}

	// Step 143: t18 = x^0xfd99132a53e77b82317be902e7
	t18.Mul(t18, t19)

	// Step 151: t18 = x^0xfd99132a53e77b82317be902e700
	for s := 0; s < 8; s++ {
		t18.Square(t18)
	}

	// Step 152: t17 = x^0xfd99132a53e77b82317be902e78b
	t17.Mul(t17, t18)

	// Step 157: t17 = x^0x1fb322654a7cef70462f7d205cf160
	for s := 0; s < 5; s++ {
		t17.Square(t17)
	}

	// Step 158: t16 = x^0x1fb322654a7cef70462f7d205cf17f
	t16.Mul(t16, t17)

	// Step 169: t16 = x^0xfd99132a53e77b82317be902e78bf800
	for s := 0; s < 11; s++ {
		t16.Square(t16)
	}

	// Step 170: t15 = x^0xfd99132a53e77b82317be902e78bf8eb
	t15.Mul(t15, t16)

	// Step 179: t15 = x^0x1fb322654a7cef70462f7d205cf17f1d600
	for s := 0; s < 9; s++ {
		t15.Square(t15)
	}

	// Step 180: t14 = x^0x1fb322654a7cef70462f7d205cf17f1d6b5
	t14.Mul(t14, t15)

	// Step 190: t14 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad400
	for s := 0; s < 10; s++ {
		t14.Square(t14)
	}

	// Step 191: t13 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad4bb
	t13.Mul(t13, t14)

	// Step 201: t13 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52ec00
	for s := 0; s < 10; s++ {
		t13.Square(t13)
	}

	// Step 202: t12 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5
	t12.Mul(t12, t13)

	// Step 209: t12 = x^0xfd99132a53e77b82317be902e78bf8eb5a9765280
	for s := 0; s < 7; s++ {
		t12.Square(t12)
	}

	// Step 210: t11 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff
	t11.Mul(t11, t12)

	// Step 219: t11 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe00
	for s := 0; s < 9; s++ {
		t11.Square(t11)
	}

	// Step 220: t10 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d
	t10.Mul(t10, t11)

	// Step 228: t10 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d00
	for s := 0; s < 8; s++ {
		t10.Square(t10)
	}

	// Step 229: t9 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9b
	t9.Mul(t9, t10)

	// Step 237: t9 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9b00
	for s := 0; s < 8; s++ {
		t9.Square(t9)
	}

	// Step 238: t8 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd
	t8.Mul(t8, t9)

	// Step 239: t8 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377a
	t8.Square(t8)

	// Step 240: t8 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b
	t8.Mul(x, t8)

	// Step 255: t8 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd8000
	for s := 0; s < 15; s++ {
		t8.Square(t8)
	}

	// Step 256: t7 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd8095
	t7.Mul(t7, t8)

	// Step 265: t7 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a00
	for s := 0; s < 9; s++ {
		t7.Square(t7)
	}

	// Step 266: t6 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a6d
	t6.Mul(t6, t7)

	// Step 275: t6 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad4bb297fa366ef60254da00
	for s := 0; s < 9; s++ {
		t6.Square(t6)
	}

	// Step 276: t6 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad4bb297fa366ef60254daab
	t6.Mul(t1, t6)

	// Step 285: t6 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b55600
	for s := 0; s < 9; s++ {
		t6.Square(t6)
	}

	// Step 286: t5 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c5
	t5.Mul(t5, t6)

	// Step 295: t5 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a00
	for s := 0; s < 9; s++ {
		t5.Square(t5)
	}

	// Step 296: t4 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a97
	t4.Mul(t4, t5)

	// Step 306: t4 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad4bb297fa366ef60254daab62a5c00
	for s := 0; s < 10; s++ {
		t4.Square(t4)
	}

	// Step 307: t3 = x^0x7ecc899529f3bdc118bdf48173c5fc75ad4bb297fa366ef60254daab62a5cfd
	t3.Mul(t3, t4)

	// Step 313: t3 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973f40
	for s := 0; s < 6; s++ {
		t3.Square(t3)
	}

	// Step 314: t2 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff
	t2.Mul(t2, t3)

	// Step 325: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff800
	for s := 0; s < 11; s++ {
		t2.Square(t2)
	}

	// Step 326: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff855
	t2.Mul(t0, t2)

	// Step 334: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff85500
	for s := 0; s < 8; s++ {
		t2.Square(t2)
	}

	// Step 335: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff85555
	t2.Mul(t0, t2)

	// Step 343: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff8555500
	for s := 0; s < 8; s++ {
		t2.Square(t2)
	}

	// Step 344: t2 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff8555555
	t2.Mul(t0, t2)

	// Step 353: t2 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff0aaaaaa00
	for s := 0; s < 9; s++ {
		t2.Square(t2)
	}

	// Step 354: t2 = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff0aaaaaa55
	t2.Mul(t0, t2)

	// Step 363: t2 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a6d55b152e7ffe1555554aa00
	for s := 0; s < 9; s++ {
		t2.Square(t2)
	}

	// Step 364: t1 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a6d55b152e7ffe1555554aaab
	t1.Mul(t1, t2)

	// Step 372: t1 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a6d55b152e7ffe1555554aaab00
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 373: t0 = x^0x3f6644ca94f9dee08c5efa40b9e2fe3ad6a5d94bfd1b377b012a6d55b152e7ffe1555554aaab55
	t0.Mul(t0, t1)

	// Step 379: t0 = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff85555552aaad540
	for s := 0; s < 6; s++ {
		t0.Square(t0)
	}

	// Step 380: z = x^0xfd99132a53e77b82317be902e78bf8eb5a97652ff46cddec04a9b556c54b9fff85555552aaad555
	z.Mul(&z, t0)

	// Step 381: z = x^0x1fb322654a7cef70462f7d205cf17f1d6b52eca5fe8d9bbd809536aad8a973fff0aaaaaa5555aaaa
	z.Square(&z)

	x.Set(&z)
}

func tateExp2(a, c, b *fp.Element) {
	var u0, u1, u2, u3 fp.Element

	u0.Square(b)
	fpExpZ(&u1, b)
	fpExpZ(&u2, &u1)
	fpExpZ(&u3, &u2)
	u0.Mul(&u0, &u2)
	u0.Mul(&u0, &u3)
	fpExpZ(&u3, &u3)
	u1.Mul(&u1, &u3)
	fpExpZ(&u3, &u3)
	u1.Mul(&u1, &u3)

	a.Set(&u0)
	c.Set(&u1)
}

func fpExpZ(z, x *fp.Element) {
	var u0 fp.Element
	u0.Square(x)
	u0.Mul(&u0, x)
	u0.Square(&u0)
	u0.Square(&u0)
	u0.Mul(&u0, x)
	u0.Square(&u0)
	u0.Square(&u0)
	u0.Square(&u0)
	u0.Mul(&u0, x)
	for i := 0; i < 9; i++ {
		u0.Square(&u0)
	}
	u0.Mul(&u0, x)
	for i := 0; i < 32; i++ {
		u0.Square(&u0)
	}
	u0.Mul(&u0, x)
	for i := 0; i < 16; i++ {
		u0.Square(&u0)
	}

	z.Set(&u0)
}

func tateNAFDigits() []int8 {
	tateNafOnce.Do(func() {
		tateNafLen = nafDigitsFixed(&tateNafDigits, &xGen)
	})
	return tateNafDigits[:tateNafLen]
}

func nafDigitsFixed(out *[tateNafMaxLen]int8, n *big.Int) int {
	k := new(big.Int).Set(n)
	one := big.NewInt(1)
	three := big.NewInt(3)

	i := 0
	for k.Sign() > 0 {
		if k.Bit(0) == 1 {
			mod4 := new(big.Int).And(k, three).Int64()
			z := int8(1)
			if mod4 == 3 {
				z = -1
			}
			out[i] = z
			if z > 0 {
				k.Sub(k, one)
			} else {
				k.Add(k, one)
			}
		} else {
			out[i] = 0
		}
		i++
		k.Rsh(k, 1)
	}
	if i == 0 {
		out[0] = 0
		i = 1
	}
	return i
}
