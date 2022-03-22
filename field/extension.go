package field

import "math/big"

//Extension is a simple radical extension, obtained by adjoining ⁿ√α to Fp
type Extension struct {
	Base   *Field  //Fp
	Size   big.Int //q
	Degree int     //n such that q = pⁿ TODO: Make uint8 so forced to be positive and small
	RootOf int64   //α
}

func NewTower(base *Field, degree uint8, rootOf int64) Extension {
	ret := Extension{
		Degree: int(degree),
		RootOf: rootOf,
		Base:   base,
	}
	ret.Size.Exp(base.ModulusBig, big.NewInt(int64(degree)), nil)
	return ret
}

func (f *Extension) FromInt64(i []int64) []big.Int {
	z := make([]big.Int, f.Degree)
	for n := 0; n < len(i) && n < int(f.Degree); n++ {
		z[n].SetInt64(i[n])
	}
	return z
}

func (f *Extension) Neg(x []big.Int) []big.Int {
	z := make([]big.Int, len(x))
	for n := 0; n < len(x); n++ {
		z[n].Neg(&x[n])
	}
	return z
}

func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func (f *Extension) Mul(x []big.Int, y []big.Int) []big.Int {
	z := make([]big.Int, f.Degree)
	maxP := len(x) + len(y) - 2
	alpha := big.NewInt(f.RootOf)

	for p := maxP; p >= 0; p-- {

		var rp big.Int

		for m := max(p-(len(y)-1), 0); m < len(x) && m <= p; m++ {
			n := p - m
			var prod big.Int
			prod.Mul(&x[m], &y[n])
			rp.Add(&rp, &prod)
		}

		rI := p % int(f.Degree) //reduced index

		z[rI].Add(&z[rI], &rp).Mod(&z[rI], f.Base.ModulusBig)

		if p >= int(f.Degree) {
			z[rI].Mul(&z[rI], alpha)
		}
	}

	return z
}

func (f *Extension) MulScalar(c *big.Int, x []big.Int) []big.Int {
	z := make([]big.Int, len(x))
	for i := 0; i < len(x); i++ {
		f.Base.Mul(&z[i], c, &x[i])
	}
	return z
}

func (f *Extension) Halve(z []big.Int) {
	for i := 0; i < len(z); i++ {
		if z[i].Bit(0) != 0 {
			z[i].Add(&z[i], f.Base.ModulusBig)
		}
		z[i].Rsh(&z[i], 1)
	}
}

func (f *Extension) reduce(z []big.Int) {
	for i := 0; i < len(z); i++ {
		z[i].Mod(&z[i], f.Base.ModulusBig)
	}
}

// Sqrt returning √ x, or nil if x is not qr.
func (f *Extension) Sqrt(x []big.Int) []big.Int {

	z := make([]big.Int, f.Degree)
	switch f.Degree {
	case 1:
		if z[0].ModSqrt(&x[0], f.Base.ModulusBig) == nil {
			return nil
		}
	case 2:
		// z = z₀ + z₁ i

		if x[0].BitLen() == 0 {
			z[1].ModInverse(big.NewInt(f.RootOf), f.Base.ModulusBig).Mul(&z[1], &x[1])
		}

		var discriminant big.Int
		z[0].Mul(&x[0], &x[0])
		z[1].Mul(&x[1], &x[1]).Mul(&z[1], big.NewInt(-f.RootOf))
		z[0].Sub(&z[0], &z[1])
		if discriminant.ModSqrt(&z[0], f.Base.ModulusBig) == nil {
			return nil
		}
		z[0].Add(&x[0], &discriminant)
		f.Base.halve(&z[0], &z[0])
		if z[0].ModSqrt(&z[0], f.Base.ModulusBig) == nil {
			z[0].Sub(&z[0], &discriminant)
			if z[0].ModSqrt(&z[0], f.Base.ModulusBig) == nil {
				return nil
			}
		}
		z[1].Lsh(&z[0], 1).ModInverse(&z[1], f.Base.ModulusBig).Mul(&z[1], &x[1])

	default:
		panic("only degrees 1 and 2 are supported")
	}

	f.reduce(z)
	return z
}

func (f *Extension) ToMont(x []big.Int) []big.Int {
	z := make([]big.Int, len(x))
	for i := 0; i < len(x); i++ {
		f.Base.ToMont(&z[i], &x[i])
	}
	return z
}

func (f *Extension) Equal(x []big.Int, y []big.Int) bool {
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

func (f *Extension) norm(z *big.Int, x []big.Int) *Extension {
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

func (f *Extension) Inverse(x []big.Int) []big.Int {
	z := make([]big.Int, f.Degree)
	switch f.Degree {
	case 1:
		z[0].ModInverse(&x[0], f.Base.ModulusBig)
	case 2:
		var normInv big.Int
		f.norm(&normInv, x)
		normInv.ModInverse(&normInv, f.Base.ModulusBig)
		z[0].Mul(&x[0], &normInv)

		z[1].Neg(&x[1]).Mul(&z[1], &normInv)
	}
	return z
}

func (f *Extension) Exp(x []big.Int, exp *big.Int) []big.Int {

	if exp.BitLen() == 0 {
		return f.FromInt64([]int64{1})
	}

	z := x

	for i := exp.BitLen() - 2; i >= 0; i-- {
		z = f.Mul(z, z)
		if exp.Bit(i) == 1 {
			z = f.Mul(z, x)
		}
	}

	return z
}

func (f *Extension) StringSliceToMont(hex []string) []big.Int {
	if len(hex) > int(f.Degree) {
		panic("too many monomials")
	}

	res := make([]big.Int, f.Degree)

	for i := 0; i < len(res); i++ {
		res[i] = f.Base.StringToMont(hex[i])
	}

	return res
}

func (f *Extension) StringToIntSliceSlice(hex [][]string) [][]big.Int {

	res := make([][]big.Int, len(hex))

	for i, hex := range hex {
		res[i] = f.StringSliceToMont(hex)
	}

	return res
}
