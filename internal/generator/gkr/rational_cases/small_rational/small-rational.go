package small_rational

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
)

type SmallRational struct {
	Numerator   int64
	Denominator int64 // By convention, Denominator == 0 also indicates zero
}

var smallPrimes = []int64{2, 3, 5, 7, 11, 13}

func (z *SmallRational) simplify() {
	// factoring 64bit numbers can be practical, TODO: Sophisticated algorithm?

	if z.Numerator == 0 || z.Denominator == 0 {
		return
	}

	for _, p := range smallPrimes {
		for z.Numerator%p == 0 && z.Denominator%p == 0 {
			z.Numerator /= p
			z.Denominator /= p
		}
	}

}

func (z *SmallRational) Add(x, y *SmallRational) *SmallRational {
	if x.Denominator == 0 {
		*z = *y
	} else if y.Denominator == 0 {
		*z = *x
	} else {
		//TODO: Exploit cases where one denom divides the other
		*z = SmallRational{x.Numerator*y.Denominator + y.Numerator*x.Denominator, x.Denominator * y.Denominator}
		z.simplify()
	}

	return z
}

func (z *SmallRational) IsZero() bool {
	return z.Numerator == 0 || z.Denominator == 0
}

func (z *SmallRational) Inverse(x *SmallRational) *SmallRational {
	if x.IsZero() {
		*z = *x
	} else {
		*z = SmallRational{x.Denominator, x.Numerator}
	}
	return z
}

func (z *SmallRational) Neg(x *SmallRational) *SmallRational {
	*z = SmallRational{-x.Numerator, x.Denominator}
	return z
}

func (z *SmallRational) Double(x *SmallRational) *SmallRational {
	if x.Denominator%2 == 0 {
		*z = SmallRational{x.Numerator, x.Denominator / 2}
	} else {
		*z = SmallRational{x.Numerator * 2, x.Denominator}
	}
	return z
}

func (z *SmallRational) sign() int {
	if z.IsZero() {
		return 0
	}
	if z.Numerator > 0 {
		if z.Denominator > 0 {
			return 1
		}
		return -1
	}
	if z.Denominator > 0 {
		return -1
	}
	return 1
}

func (z *SmallRational) abs() (abs SmallRational) {
	abs = *z
	if abs.Numerator < 0 {
		abs.Numerator = -abs.Numerator
	}
	if abs.Denominator < 0 {
		abs.Denominator = -abs.Denominator
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

	cross0Hi, cross0Lo := bits.Mul64(uint64(xAbs.Numerator), uint64(zAbs.Denominator))
	cross1Hi, cross1Lo := bits.Mul64(uint64(zAbs.Numerator), uint64(xAbs.Denominator))

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
	*z = SmallRational{x.Numerator * y.Numerator, x.Denominator * y.Denominator}
	z.simplify()
	return z
}

func (z *SmallRational) SetOne() *SmallRational {
	return z.SetInt64(1)
}

func (z *SmallRational) SetZero() *SmallRational {
	return z.SetInt64(0)
}

func (z *SmallRational) SetInt64(i int64) *SmallRational {
	z.Numerator = i
	z.Denominator = 1
	return z
}

func (z *SmallRational) SetRandom() (*SmallRational, error) {
	/*bytes := make([]byte, 2*64/8)
	n, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	if n != len(bytes) {
		return nil, fmt.Errorf("%d bytes read instead of %d", n, len(bytes))
	}

	// TODO: Verify that in case of overflow casting gives a negative
	z.Numerator = int64(binary.BigEndian.Uint64(bytes[:64/8]))
	z.Denominator = int64(binary.BigEndian.Uint64(bytes[64/8:]))

	return z, nil*/

	bytes := make([]byte, 1)
	n, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	if n != len(bytes) {
		return nil, fmt.Errorf("%d bytes read instead of %d", n, len(bytes))
	}

	z.Numerator = int64(bytes[0]%16) - 8
	z.Denominator = int64((bytes[0]) / 16)

	z.simplify()

	return z, nil
}

func (z *SmallRational) SetUint64(i uint64) {
	z.Numerator = int64(i)
	z.Denominator = 1
}

func (z *SmallRational) IsOne() bool {
	return z.Numerator == z.Denominator && z.Denominator != 0
}

func (z *SmallRational) Text(base int) string {

	if z.Denominator == 0 {
		return "0"
	}

	if z.Denominator < 0 {
		z.Numerator = -z.Numerator
		z.Denominator = -z.Denominator
	}

	if z.Numerator%z.Denominator == 0 {
		z.Numerator /= z.Denominator
		z.Denominator = 1
	}

	numerator := strconv.FormatInt(z.Numerator, base)

	if z.Denominator == 1 {
		return numerator
	}

	return numerator + "/" + strconv.FormatInt(z.Denominator, base)
}

func (z *SmallRational) Set(x *SmallRational) *SmallRational {
	z.Numerator = x.Numerator
	z.Denominator = x.Denominator
	return z
}

func Modulus() *big.Int {
	res := big.NewInt(1)
	res.Lsh(res, 64)
	return res
}
