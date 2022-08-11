package sumcheck

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"math/bits"
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
func (m messageCounter) NextFromBytes(_ []byte) fr.Element {
	m.increment()
	return m.Element
	/*res := make([]fr.Element, size)
	for i := 0; i < size; i++ {
		m.increment()
		res[i] = m.Element
	}
	return res*/
}

type singleMultilinClaim struct {
	g polynomial.MultiLin
}

type singleMultilinSubClaim struct {
	g polynomial.MultiLin
}

func (c singleMultilinClaim) VarsNum() int {
	return bits.TrailingZeros(uint(len(c.g)))
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

func (c singleMultilinClaim) Combine(fr.Element) (SubClaim, polynomial.Polynomial) {
	sub := singleMultilinSubClaim{c.g.Clone()}

	return sub, getSumsPoly(c.g)[1:]
}

func (c singleMultilinSubClaim) Next(r fr.Element) polynomial.Polynomial {
	c.g.Fold(r)
	return getSumsPoly(c.g)[1:]
}

type singleMultilinLazyClaim struct {
	g          polynomial.MultiLin
	claimedSum fr.Element
}

func (c singleMultilinLazyClaim) CombinedSum(combinationCoeffs fr.Element) fr.Element {
	return c.claimedSum
}

func (c singleMultilinLazyClaim) CombinedEval(combinationCoeffs fr.Element, r []fr.Element) fr.Element {
	return c.g.Evaluate(r)
}

func (c singleMultilinLazyClaim) Degree(i int) int {
	return 1
}

func (c singleMultilinLazyClaim) ClaimsNum() int {
	return 1
}

func (c singleMultilinLazyClaim) VarsNum() int {
	return bits.TrailingZeros(uint(len(c.g)))
}

func TestSumcheckDeterministicHashSingleClaimMultilin(t *testing.T) {
	var one, two, three, four, five, six, seven, eight fr.Element

	one.SetOne()
	two.Double(&one)
	three.Add(&two, &one)
	four.Double(&two)
	five.Add(&three, &two)
	six.Double(&three)
	seven.Add(&four, &three)
	eight.Double(&four)

	poly := polynomial.MultiLin{one, two, three, four}
	claim := singleMultilinClaim{g: poly}

	proof := Prove(claim, messageCounter{}, []byte{})

	lazyClaim := singleMultilinLazyClaim{g: poly, claimedSum: poly.Sum()}

	if !Verify(lazyClaim, proof, messageCounter{}, []byte{}) {
		t.Fail()
	}
}
