package utils

import "math/bits"

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
	zSign := z.sign()
	if zSign != x.sign() {
		return false
	}
	xAbs, zAbs := x.abs(), z.abs()

	cross0Hi, cross0Lo := bits.Mul64(uint64(xAbs.numerator), uint64(zAbs.denominator))
	cross1Hi, cross1Lo := bits.Mul64(uint64(zAbs.numerator), uint64(xAbs.denominator))

	return cross1Hi == cross0Hi && cross1Lo == cross0Lo
}

func (z *SmallRational) Sub(x, y *SmallRational) *SmallRational {
	var yNeg SmallRational
	yNeg.Neg(y)
	z.Add(x, &yNeg)
	return z
}
