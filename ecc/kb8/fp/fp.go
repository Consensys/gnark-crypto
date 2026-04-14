package fp

import (
	"math/big"

	kb "github.com/consensys/gnark-crypto/field/koalabear"
)

type Element = kb.Element
type Vector = kb.Vector

const (
	Bits  = kb.Bits
	Bytes = kb.Bytes
	Limbs = kb.Limbs
)

var BigEndian = kb.BigEndian

func Modulus() *big.Int {
	return kb.Modulus()
}

func One() Element {
	return kb.One()
}

func BatchInvert(a []Element) []Element {
	return kb.BatchInvert(a)
}

func Generator(m uint64) (Element, error) {
	return kb.Generator(m)
}

func Butterfly(a, b *Element) {
	kb.Butterfly(a, b)
}

func MulBy3(x *Element) {
	kb.MulBy3(x)
}
