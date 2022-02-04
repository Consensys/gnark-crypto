package field

import "math/big"

type Tower struct {
	Base   *Field  //Fp
	Size   big.Int //q
	Degree uint8   //n such that q = pⁿ
	RootOf int64   //Number
}

func NewTower(base *Field, degree uint8, rootOf int64) Tower {
	ret := Tower{
		Degree: degree,
		RootOf: rootOf,
		Base:   base,
	}
	ret.Size.Exp(base.ModulusBig, big.NewInt(int64(degree)), nil)
	return ret
}

func (f *Tower) SetInt64(z *[]big.Int, i []int64) *Tower {
	*z = make([]big.Int, f.Degree)
	for n := 0; n < len(i) && n < int(f.Degree); n++ {
		(*z)[n].SetInt64(i[n])
	}
	return f
}

func (f *Tower) Neg(z *[]big.Int, x []big.Int) *Tower {
	r := make([]big.Int, len(x))
	for n := 0; n < len(x); n++ {
		r[n].Neg(&x[n])
	}
	*z = r
	return f
}

func (f *Tower) Mul(z *[]big.Int, x []big.Int, y []big.Int) *Tower {
	r := make([]big.Int, f.Degree)
	c := big.NewInt(1)
	maxP := len(x) + len(y) - 2

	for p := 0; p <= maxP; p++ {

		var rp big.Int

		for m := p - len(y) - 1; m < len(x); m++ {
			n := p - m
			var prod big.Int
			prod.Mul(&x[m], &y[n])
			rp.Add(&rp, &prod).Mod(&rp, f.Base.ModulusBig)
		}

		rp.Mul(&rp, c)
		rPtr := &r[p%int(f.Degree)]
		rPtr.Add(rPtr, &rp)

		if p >= maxP-int(f.Degree) {
			rPtr.Mod(rPtr, f.Base.ModulusBig)
		}

		if uint8(p+1)%f.Degree == 0 {
			c.Mul(c, big.NewInt(f.RootOf))
		}
	}

	*z = r

	return f
}

func (f *Tower) MulScalar(z *[]big.Int, c *big.Int, x []big.Int) *Tower {
	res := make([]big.Int, len(x))
	for i := 0; i < len(x); i++ {
		f.Base.Mul(&res[i], c, &x[i])
	}
	*z = res
	return f
}

func (f *Tower) Halve(z *[]big.Int) *Tower {
	for i := 0; i < len(*z); i++ {
		if (*z)[i].Bit(0) != 0 {
			(*z)[i].Add(&(*z)[i], f.Base.ModulusBig)
		}
		(*z)[i].Rsh(&(*z)[i], 1)
	}
	return f
}

func (f *Tower) reduce(z []big.Int) {
	for _, x := range z {
		x.Mod(&x, f.Base.ModulusBig)
	}
}

// Sqrt z <- √ x, returning whether x is qr. If not, z is unchanged.
func (f *Tower) Sqrt(z *[]big.Int, x []big.Int) bool {

	r := make([]big.Int, f.Degree)
	switch f.Degree {
	case 1:
		if r[0].ModSqrt(&x[0], f.Base.ModulusBig) == nil {
			return false
		}
	case 2:
		// r = r₀ + r₁ i

		if x[0].BitLen() == 0 {
			r[1].ModInverse(big.NewInt(f.RootOf), f.Base.ModulusBig).Mul(&r[1], &x[1])
		}

		var discriminant big.Int
		r[0].Mul(&x[0], &x[0])
		r[1].Mul(&x[1], &x[1]).Mul(&r[1], big.NewInt(-f.RootOf))
		r[0].Sub(&r[0], &r[1])
		if discriminant.ModSqrt(&r[0], f.Base.ModulusBig) == nil {
			return false
		}
		r[0].Add(&x[0], &discriminant)
		f.Base.halve(&r[0], &r[0])
		if r[0].ModSqrt(&r[0], f.Base.ModulusBig) == nil {
			r[0].Sub(&r[0], &discriminant)
			if r[0].ModSqrt(&r[0], f.Base.ModulusBig) == nil {
				return false
			}
		}
		r[1].Lsh(&r[0], 1).ModInverse(&r[1], f.Base.ModulusBig).Mul(&r[1], &x[1])

	default:
		panic("only degrees 1 and 2 are supported")
	}

	f.reduce(r)
	*z = r
	return true
}

func (f *Tower) ToMont(z *[]big.Int, x []big.Int) *Tower {
	r := make([]big.Int, len(x))
	for i := 0; i < len(x); i++ {
		f.Base.ToMont(&r[i], &x[i])
	}
	*z = r
	return f
}

func (f *Tower) Equal(x []big.Int, y []big.Int) bool {
	if len(x) != len(y) {
		return false
	}
	for i := 0; i < len(x); i++ {
		var diff big.Int
		if diff.Sub(&x[i], &y[i]).Mod(&diff, f.Base.ModulusBig).BitLen() != 0 {
			return false
		}
	}
	return true
}

func (f *Tower) norm(z *big.Int, x []big.Int) *Tower {
	if f.Degree != 2 {
		panic("only degree 2 supported")
	}
	var x0Sq big.Int

	x0Sq.Mul(&x[0], &x[0])

	res := big.NewInt(-f.RootOf)
	res.Mul(res, &x[1]).Mul(res, &x[1]).Add(res, &x0Sq)

	z.Set(res)

	return f
}

func (f *Tower) Inverse(z *[]big.Int, x []big.Int) *Tower {
	r := make([]big.Int, f.Degree)
	switch f.Degree {
	case 1:
		r[0].ModInverse(&x[0], f.Base.ModulusBig)
	case 2:
		var normInv big.Int
		f.norm(&normInv, x)
		normInv.ModInverse(&normInv, f.Base.ModulusBig)
		r[0].Mul(&x[0], &normInv)

		r[1].Neg(&x[1]).Mul(&r[1], &normInv)
	}
	*z = r
	return f
}

func (f *Tower) Exp(z *[]big.Int, x []big.Int, exp *big.Int) *Tower {

	if exp.BitLen() == 0 {
		f.SetInt64(z, []int64{1})
		return f
	}

	res := x

	for i := exp.BitLen() - 2; i >= 0; i-- {
		f.Mul(&res, res, res)
		if exp.Bit(i) == 1 {
			f.Mul(&res, res, x)
		}
	}

	*z = res
	return f
}

func (f *Tower) HexSliceToMont(hex []string) []big.Int {
	if len(hex) > int(f.Degree) {
		panic("too many monomials")
	}

	res := make([]big.Int, f.Degree)

	for i := 0; i < len(res); i++ {
		res[i] = f.Base.HexToMont(hex[i])
	}

	return res
}

func (f *Tower) HexToIntSliceSlice(hex [][]string) [][]big.Int {

	res := make([][]big.Int, len(hex))

	for i, hex := range hex {
		res[i] = f.HexSliceToMont(hex)
	}

	return res
}
