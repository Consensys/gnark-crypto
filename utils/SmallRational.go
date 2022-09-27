package utils

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
