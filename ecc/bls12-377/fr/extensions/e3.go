package extensions

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
)

// E3 is a degree-three finite field extension of fr3
type E3 struct {
	A0, A1, A2 fr.Element
}

// Equal returns true if z equals x, false otherwise
// note this is more efficient than calling "z == x"
func (z *E3) Equal(x *E3) bool {
	return z.A0.Equal(&x.A0) && z.A1.Equal(&x.A1) && z.A2.Equal(&x.A2)
}

// SetString sets a E3 elmt from string
func (z *E3) SetString(s1, s2, s3 string) *E3 {
	z.A0.SetString(s1)
	z.A1.SetString(s2)
	z.A2.SetString(s3)
	return z
}

// SetZero sets an E3 elmt to zero
func (z *E3) SetZero() *E3 {
	*z = E3{}
	return z
}

// Clone returns a copy of self
func (z *E3) Clone() *E3 {
	return &E3{
		A0: z.A0,
		A1: z.A1,
		A2: z.A2,
	}
}

// Set Sets a E3 elmt form another E3 elmt
func (z *E3) Set(x *E3) *E3 {
	*z = *x
	return z
}

// SetOne sets z to 1 in Montgomery form and returns z
func (z *E3) SetOne() *E3 {
	z.A0.SetOne()
	z.A1.SetZero()
	z.A2.SetZero()
	return z
}

// SetRandom sets z to a random elmt
func (z *E3) SetRandom() (*E3, error) {
	if _, err := z.A0.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A1.SetRandom(); err != nil {
		return nil, err
	}
	if _, err := z.A2.SetRandom(); err != nil {
		return nil, err
	}
	return z, nil
}

// IsZero returns true if z is zero, false otherwise
func (z *E3) IsZero() bool {
	return z.A0.IsZero() && z.A1.IsZero() && z.A2.IsZero()
}

// IsOne returns true if z is one, false otherwise
func (z *E3) IsOne() bool {
	return z.A0.IsOne() && z.A1.IsZero() && z.A2.IsZero()
}

// Neg negates the E3 number
func (z *E3) Neg(x *E3) *E3 {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
	z.A2.Neg(&x.A2)
	return z
}

// Add adds two elements of E3
func (z *E3) Add(x, y *E3) *E3 {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
	z.A2.Add(&x.A2, &y.A2)
	return z
}

// Sub subtracts two elements of E3
func (z *E3) Sub(x, y *E3) *E3 {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
	z.A2.Sub(&x.A2, &y.A2)
	return z
}

// Double doubles an element in E3
func (z *E3) Double(x *E3) *E3 {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
	z.A2.Double(&x.A2)
	return z
}

// String puts E3 elmt in string form
func (z *E3) String() string {
	return (z.A0.String() + "+(" + z.A1.String() + ")*u+(" + z.A2.String() + ")*u**2")
}

// MulByElement multiplies an element in E3 by an element in fr
func (z *E3) MulByElement(x *E3, y *fr.Element) *E3 {
	_y := *y
	z.A0.Mul(&x.A0, &_y)
	z.A1.Mul(&x.A1, &_y)
	z.A2.Mul(&x.A2, &_y)
	return z
}

// Mul sets z to the E3-product of x,y, returns z
func (z *E3) Mul(x, y *E3) *E3 {
	// Karatsuba method for cubic extensions
	// https://eprint.iacr.org/2006/471.pdf (section 4)
	var t0, t1, t2, c0, c1, c2, tmp fr.Element
	t0.Mul(&x.A0, &y.A0)
	t1.Mul(&x.A1, &y.A1)
	t2.Mul(&x.A2, &y.A2)

	c0.Add(&x.A1, &x.A2)
	tmp.Add(&y.A1, &y.A2)
	c0.Mul(&c0, &tmp).Sub(&c0, &t1).Sub(&t2, &c0)

	tmp.Add(&x.A0, &x.A2)
	c2.Add(&y.A0, &y.A2).Mul(&c2, &tmp).Sub(&c2, &t0).Sub(&c2, &t2)

	c1.Add(&x.A0, &x.A1)
	tmp.Add(&y.A0, &y.A1)
	c1.Mul(&c1, &tmp).Sub(&c1, &t0).Sub(&c1, &t1)

	z.A0.Add(&c0, &t0)
	z.A1.Sub(&c1, &t2)
	z.A2.Add(&c2, &t1)

	return z
}

// MulAssign sets z to the E3-product of z,y, returns z
func (z *E3) MulAssign(x *E3) *E3 {
	return z.Mul(z, x)
}

// Square sets z to the E3-product of x,x, returns z
func (z *E3) Square(x *E3) *E3 {

	// Algorithm 16 from https://eprint.iacr.org/2010/354.pdf
	var c4, c5, c1, c2, c3, c6 fr.Element

	c6.Double(&x.A1)
	c4.Mul(&x.A0, &c6) // x.A0 * xA1 * 2
	c5.Square(&x.A2)
	c1.Sub(&c4, &c5)
	c2.Sub(&c4, &c5)

	c3.Square(&x.A0)
	c4.Sub(&x.A0, &x.A1).Add(&c4, &x.A2)
	c5.Mul(&c6, &x.A2) // x.A1 * xA2 * 2
	c4.Square(&c4)
	c4.Add(&c4, &c5).Sub(&c4, &c3)

	z.A0.Sub(&c3, &c5)
	z.A1 = c1
	z.A2.Add(&c2, &c4)

	return z
}

// MulByNonResidue mul x by (0,1,0)
func (z *E3) MulByNonResidue(x *E3) *E3 {
	z.A2, z.A1, z.A0 = x.A1, x.A0, x.A2
	z.A0.Neg(&z.A0)
	return z
}

// Inverse an element in E3
//
// if x == 0, sets and returns z = x
func (z *E3) Inverse(x *E3) *E3 {
	// Algorithm 17 from https://eprint.iacr.org/2010/354.pdf
	// step 9 is wrong in the paper it's t1-t4
	var t0, t1, t2, t3, t4, t5, t6, c0, c1, c2, d1, d2 fr.Element
	t0.Square(&x.A0)
	t1.Square(&x.A1)
	t2.Square(&x.A2)
	t3.Mul(&x.A0, &x.A1)
	t4.Mul(&x.A0, &x.A2)
	t5.Mul(&x.A1, &x.A2)
	c0.Add(&t5, &t0)
	c1.Neg(&t2).Sub(&c1, &t3)
	c2.Sub(&t1, &t4)
	t6.Mul(&x.A0, &c0)
	d1.Mul(&x.A2, &c1)
	d2.Mul(&x.A1, &c2)
	d1.Add(&d1, &d2)
	t6.Sub(&t6, &d1)
	t6.Inverse(&t6)
	z.A0.Mul(&c0, &t6)
	z.A1.Mul(&c1, &t6)
	z.A2.Mul(&c2, &t6)

	return z
}

// BatchInvertE3 returns a new slice with every element in a inverted.
// It uses Montgomery batch inversion trick.
//
// if a[i] == 0, returns result[i] = a[i]
func BatchInvertE3(a []E3) []E3 {
	res := make([]E3, len(a))
	if len(a) == 0 {
		return res
	}

	zeroes := make([]bool, len(a))
	var accumulator E3
	accumulator.SetOne()

	for i := 0; i < len(a); i++ {
		if a[i].IsZero() {
			zeroes[i] = true
			continue
		}
		res[i].Set(&accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	accumulator.Inverse(&accumulator)

	for i := len(a) - 1; i >= 0; i-- {
		if zeroes[i] {
			continue
		}
		res[i].Mul(&res[i], &accumulator)
		accumulator.Mul(&accumulator, &a[i])
	}

	return res
}
