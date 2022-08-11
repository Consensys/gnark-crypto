package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"testing"
)

// This is a very bad fiat-shamir challenge generator
type messageCounter struct {
	fr.Element
}

func (m messageCounter) increment() {
	var one fr.Element
	one.SetOne()
	m.Add(&m.Element, &one)
}
func (m messageCounter) NextFromElements(_ []fr.Element) fr.Element {
	m.increment()
	return m.Element
}
func (m messageCounter) NextFromBytes(_ []byte, size int) []fr.Element {
	res := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		m.increment()
		res[i] = m.Element
	}
	return res
}

type singleMultilinClaim struct {
	varsNum int
	g       polynomial.MultiLin
}

type singleMultilinSubClaim struct {
	g polynomial.MultiLin
}

func (c singleMultilinClaim) VarsNum() int {
	return c.varsNum
}

func (c singleMultilinClaim) ClaimsNum() int {
	return 1
}

func getSumsPoly(g polynomial.MultiLin) polynomial.Polynomial {
	g1 := make(polynomial.Polynomial, 2)

	var r fr.Element
	gR := g.Clone()
	gR.Fold(r)
	g1[0] = gR.Sum()

	r.SetOne()
	gR = g.Clone()
	gR.Fold(r)
	g1[1] = gR.Sum()

	return g1
}

func (c singleMultilinClaim) Combine([]fr.Element) (SubClaim, polynomial.Polynomial) {
	sub := singleMultilinSubClaim{c.g.Clone()}

	return sub, getSumsPoly(c.g)
}

func (c singleMultilinSubClaim) Next(r fr.Element) polynomial.Polynomial {
	c.g.Fold(r)
	return getSumsPoly(c.g)
}

func TestSumcheckDeterministicHashSingleClaimMultilin(t *testing.T) {

}
