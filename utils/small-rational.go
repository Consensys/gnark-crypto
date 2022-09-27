package utils

import (
	"math/big"
	"math/bits"
)

type SmallRational struct {
	numerator   int64
	denominator int64
}

func (z *SmallRational) simplify() {
	// do nothing for now
	// factoring 64bit numbers can be practical
}

func (z *SmallRational) Add(x, y *SmallRational) *SmallRational {
	//TODO: Exploit cases where one denom divides the other
	*z = SmallRational{x.numerator*y.denominator + y.numerator*x.denominator, x.denominator * y.denominator}
	return z
}

func (z *SmallRational) IsZero() bool {
	return z.numerator == 0
}

func (z *SmallRational) Inverse(x *SmallRational) *SmallRational {
	if x.IsZero() {
		*z = *x
	} else {
		*z = SmallRational{x.denominator, x.numerator}
	}
	return z
}

func (z *SmallRational) Neg(x *SmallRational) *SmallRational {
	*z = SmallRational{-x.numerator, x.denominator}
	return z
}

func (z *SmallRational) Double(x *SmallRational) *SmallRational {
	if x.denominator%2 == 0 {
		*z = SmallRational{x.numerator, x.denominator / 2}
	} else {
		*z = SmallRational{x.numerator * 2, x.denominator / 2}
	}
	return z
}

func (z *SmallRational) sign() int {
	if z.IsZero() {
		return 0
	}
	if z.numerator > 0 {
		if z.denominator > 0 {
			return 1
		}
		return -1
	}
	if z.denominator > 0 {
		return -1
	}
	return 1
}

func (z *SmallRational) abs() (abs SmallRational) {
	abs = *z
	if abs.numerator < 0 {
		abs.numerator = -abs.numerator
	}
	if abs.denominator < 0 {
		abs.denominator = -abs.denominator
	}
	return abs
}

func (z *SmallRational) Equal(x *SmallRational) bool {
	return z.Cmp(x) == 0
}

func (z *SmallRational) Sub(x, y *SmallRational) *SmallRational {
	var yNeg SmallRational
	yNeg.Neg(y)
	z.Add(x, &yNeg)
	return z
}

func (z *SmallRational) ToBigIntRegular(*big.Int) big.Int {
	panic("Not implemented")
}

// TODO: Test this
func (z *SmallRational) Cmp(x *SmallRational) int {
	zSign, xSign := z.sign(), x.sign()

	if zSign > xSign {
		return 1
	}
	if zSign < xSign {
		return -1
	}

	xAbs, zAbs := x.abs(), z.abs()

	cross0Hi, cross0Lo := bits.Mul64(uint64(xAbs.numerator), uint64(zAbs.denominator))
	cross1Hi, cross1Lo := bits.Mul64(uint64(zAbs.numerator), uint64(xAbs.denominator))

	if cross1Hi > cross0Hi {
		return zSign
	}
	if cross1Hi < cross0Hi {
		return -zSign
	}
	if cross1Lo > cross0Lo {
		return zSign
	}
	if cross1Lo < cross0Lo {
		return -zSign
	}
	return 0
}

func BatchInvert(a []SmallRational) []SmallRational {
	res := make([]SmallRational, len(a))
	for i := range a {
		res[i].Inverse(&a[i])
	}
	return res
}

func (z *SmallRational) Mul(x, y *SmallRational) *SmallRational {
	*z = SmallRational{x.numerator * y.numerator, x.denominator * y.denominator}
	return z
}

func (z *SmallRational) SetOne() *SmallRational {
	z.numerator = 1
	z.denominator = 1
	return z
}

func (z *SmallRational) SetInt64(i int64) *SmallRational {
	z.numerator = i
	z.denominator = 1
	return z
}
