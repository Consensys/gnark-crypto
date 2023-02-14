package pedersenhash

import (
	"math/big"

	starkcurve "github.com/consensys/gnark-crypto/ecc/stark-curve"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

var (
	shiftPoint starkcurve.G1Jac
	p0         starkcurve.G1Jac
	p1         starkcurve.G1Jac
	p2         starkcurve.G1Jac
	p3         starkcurve.G1Jac
)

func init() {
	// The curve points come from the [reference implementation].
	//
	// [reference implementation]: https://github.com/starkware-libs/cairo-lang/blob/de741b92657f245a50caab99cfaef093152fd8be/src/starkware/crypto/signature/fast_pedersen_hash.py

	shiftPoint.X.SetString("2089986280348253421170679821480865132823066470938446095505822317253594081284")
	shiftPoint.Y.SetString("1713931329540660377023406109199410414810705867260802078187082345529207694986")
	shiftPoint.Z.SetOne()

	p0.X.SetString("996781205833008774514500082376783249102396023663454813447423147977397232763")
	p0.Y.SetString("1668503676786377725805489344771023921079126552019160156920634619255970485781")
	p0.Z.SetOne()

	p1.X.SetString("2251563274489750535117886426533222435294046428347329203627021249169616184184")
	p1.Y.SetString("1798716007562728905295480679789526322175868328062420237419143593021674992973")
	p1.Z.SetOne()

	p2.X.SetString("2138414695194151160943305727036575959195309218611738193261179310511854807447")
	p2.Y.SetString("113410276730064486255102093846540133784865286929052426931474106396135072156")
	p2.Z.SetOne()

	p3.X.SetString("2379962749567351885752724891227938183011949129833673362440656643086021394946")
	p3.Y.SetString("776496453633298175483985398648758586525933812536653089401905292063708816422")
	p3.Z.SetOne()
}

// PedersenArray implements [Pedersen array hashing].
//
// [Pedersen array hashing]: https://docs.starknet.io/documentation/develop/Hashing/hash-functions/#array_hashing
func PedersenArray(elems ...*fp.Element) *fp.Element {
	d := new(fp.Element)
	for _, e := range elems {
		d = Pedersen(d, e)
	}
	return Pedersen(d, new(fp.Element).SetUint64(uint64(len(elems))))
}

// Pedersen implements the [Pedersen hash] based on the [reference implementation].
//
// [Pedersen hash]: https://docs.starknet.io/documentation/develop/Hashing/hash-functions/#pedersen_hash
// [reference implementation]: https://github.com/starkware-libs/cairo-lang/blob/de741b92657f245a50caab99cfaef093152fd8be/src/starkware/crypto/signature/fast_pedersen_hash.py
func Pedersen(a *fp.Element, b *fp.Element) *fp.Element {

	result := new(starkcurve.G1Jac).Set(&shiftPoint)
	result.AddAssign(processElement(a, &p0, &p1))
	result.AddAssign(processElement(b, &p2, &p3))

	// recover the affine x coordinate
	var x fp.Element
	x.Inverse(&result.Z).
		Square(&x)
	x.Mul(&result.X, &x)

	return &x
}

func processElement(a *fp.Element, p1 *starkcurve.G1Jac, p2 *starkcurve.G1Jac) *starkcurve.G1Jac {
	var bigInt big.Int
	var aBytes [32]byte
	a.BigInt(&bigInt).FillBytes(aBytes[:])

	highPart := bigInt.SetUint64(uint64(aBytes[0])) // The top nibble (bits 249-252)
	lowPart := aBytes[1:]                           // Zero-out the top nibble (bits 249-252)

	m := new(starkcurve.G1Jac).ScalarMultiplication(p2, highPart)

	var n starkcurve.G1Jac
	n.ScalarMultiplication(p1, bigInt.SetBytes(lowPart))
	return m.AddAssign(&n)
}
