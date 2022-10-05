package small_rational

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type SmallRational struct {
	Numerator   big.Int
	Denominator big.Int // By convention, Denominator == 0 also indicates zero
}

var smallPrimes = []*big.Int{
	big.NewInt(2), big.NewInt(3), big.NewInt(5),
	big.NewInt(7), big.NewInt(11), big.NewInt(13),
}

func bigDivides(p, a *big.Int) bool {
	var remainder big.Int
	remainder.Mod(a, p)
	return remainder.BitLen() == 0
}

func (z *SmallRational) simplify() {
	// factoring 64bit numbers can be practical, TODO: Sophisticated algorithm?

	if z.Numerator.BitLen() == 0 || z.Denominator.BitLen() == 0 {
		return
	}

	for _, p := range smallPrimes {
		for bigDivides(p, &z.Numerator) && bigDivides(p, &z.Denominator) {
			z.Numerator.Div(&z.Numerator, p)
			z.Denominator.Div(&z.Denominator, p)
		}
	}

}
func (z *SmallRational) Square(x *SmallRational) *SmallRational {
	z.Numerator.Mul(&x.Numerator, &x.Numerator)
	z.Denominator.Mul(&x.Denominator, &x.Denominator)

	return z
}

func (z *SmallRational) String() string {
	return z.Text(10)
}

func (z *SmallRational) Add(x, y *SmallRational) *SmallRational {
	if x.Denominator.BitLen() == 0 {
		*z = *y
	} else if y.Denominator.BitLen() == 0 {
		*z = *x
	} else {
		//TODO: Exploit cases where one denom divides the other
		var numDen, denNum big.Int
		numDen.Mul(&x.Numerator, &y.Denominator)
		denNum.Mul(&x.Denominator, &y.Numerator)
		z.Numerator.Add(&denNum, &numDen)
		z.Denominator.Mul(&x.Denominator, &y.Denominator)
		z.simplify()
	}

	return z
}

func (z *SmallRational) IsZero() bool {
	return z.Numerator.BitLen() == 0 || z.Denominator.BitLen() == 0
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
	z.Numerator.Neg(&x.Numerator)
	z.Denominator = x.Denominator
	return z
}

func (z *SmallRational) Double(x *SmallRational) *SmallRational {
	if x.Denominator.Bit(0) == 0 {
		z.Numerator = x.Numerator
		z.Denominator.Rsh(&x.Denominator, 1)
	}
	return z
}

func (z *SmallRational) sign() int {
	return z.Numerator.Sign() * z.Denominator.Sign()
}

/*
func (z *SmallRational) abs() (abs SmallRational) {
	abs = *z
	if abs.Numerator < 0 {
		abs.Numerator = -abs.Numerator
	}
	if abs.Denominator < 0 {
		abs.Denominator = -abs.Denominator
	}
	return abs
}*/

func (z *SmallRational) MarshalJSON() ([]byte, error) {
	return []byte(z.String()), nil
}

func (z *SmallRational) UnmarshalJson(data []byte) error {
	_, err := z.Set(string(data))
	return err
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

func (z *SmallRational) Cmp(x *SmallRational) int {
	zSign, xSign := z.sign(), x.sign()

	if zSign > xSign {
		return 1
	}
	if zSign < xSign {
		return -1
	}

	var Z, X big.Int
	Z.Mul(&z.Numerator, &x.Denominator)
	X.Mul(&x.Numerator, &z.Denominator)

	Z.Abs(&Z)
	X.Abs(&X)

	return Z.Cmp(&X) * zSign

}

func BatchInvert(a []SmallRational) []SmallRational {
	res := make([]SmallRational, len(a))
	for i := range a {
		res[i].Inverse(&a[i])
	}
	return res
}

func (z *SmallRational) Mul(x, y *SmallRational) *SmallRational {
	z.Numerator.Mul(&x.Numerator, &y.Numerator)
	z.Denominator.Mul(&x.Denominator, &y.Denominator)
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
	z.Numerator.SetInt64(i)
	z.Denominator.SetInt64(1)
	return z
}

func (z *SmallRational) SetRandom() (*SmallRational, error) {

	bytes := make([]byte, 1)
	n, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	if n != len(bytes) {
		return nil, fmt.Errorf("%d bytes read instead of %d", n, len(bytes))
	}

	z.Numerator.SetInt64(int64(bytes[0]%16) - 8)
	z.Denominator.SetInt64(int64((bytes[0]) / 16))

	z.simplify()

	return z, nil
}

func (z *SmallRational) SetUint64(i uint64) {
	z.Numerator.SetUint64(i)
	z.Denominator.SetUint64(1)
}

func (z *SmallRational) IsOne() bool {
	return z.Numerator.Cmp(&z.Denominator) == 0 && z.Denominator.BitLen() != 0
}

func (z *SmallRational) Text(base int) string {

	if z.Denominator.BitLen() == 0 {
		return "0"
	}

	if z.Denominator.Sign() < 0 {
		z.Numerator.Neg(&z.Numerator)
		z.Denominator.Neg(&z.Denominator)
	}

	if bigDivides(&z.Denominator, &z.Numerator) {
		z.Numerator.Div(&z.Numerator, &z.Denominator)
		z.Denominator.SetInt64(1)
	}

	numerator := z.Numerator.Text(base)

	if z.Denominator.IsInt64() && z.Denominator.Int64() == 1 {
		return numerator
	}

	return numerator + "/" + z.Denominator.Text(base)
}

func (z *SmallRational) Set(x interface{}) (*SmallRational, error) {

	switch v := x.(type) {
	case *SmallRational:
		z.Numerator = v.Numerator
		z.Denominator = v.Denominator
	case SmallRational:
		z.Numerator = v.Numerator
		z.Denominator = v.Denominator
	case int:
		z.SetInt64(int64(v))
	case float64:
		asInt := int64(v)
		if float64(asInt) != v {
			return nil, fmt.Errorf("cannot currently parse float")
		}
		z.SetInt64(asInt)
	case string:
		sep := strings.Split(v, "/")
		switch len(sep) {
		case 1:
			if asInt, err := strconv.Atoi(sep[0]); err == nil {
				z.SetInt64(int64(asInt))
			} else {
				return nil, err
			}
		case 2:
			var err error
			var num, denom int
			num, err = strconv.Atoi(sep[0])
			if err != nil {
				return nil, err
			}
			denom, err = strconv.Atoi(sep[1])
			if err != nil {
				return nil, err
			}
			z.Numerator.SetInt64(int64(num))
			z.Denominator.SetInt64(int64(denom))
		default:
			return nil, fmt.Errorf("cannot parse \"%s\"", v)
		}
	default:
		return nil, fmt.Errorf("cannot parse %T", x)
	}

	return z, nil
}

func Modulus() *big.Int {
	res := big.NewInt(1)
	res.Lsh(res, 64)
	return res
}
