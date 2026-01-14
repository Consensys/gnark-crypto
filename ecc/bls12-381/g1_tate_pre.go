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

func tateExp1(z *fp.Element) {
	tateExp1Once.Do(func() {
		pMinus1 := fp.Modulus()
		pMinus1.Sub(pMinus1, big.NewInt(1))
		denom := new(big.Int).Set(&xGen)
		denom.Add(denom, big.NewInt(1))
		pMinus1.Div(pMinus1, denom)
		tateExp1Exponent.Set(pMinus1)
	})

	var out fp.Element
	out.Exp(*z, &tateExp1Exponent)
	z.Set(&out)
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
