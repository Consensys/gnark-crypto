package sumcheck

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/polynomial"
	"math/bits"
	"testing"
)

// This is a very bad fiat-shamir challenge generator
type messageCounter struct {
	state uint64
	step  uint64
}

func newMessageCounterGenerator(startState, step int) func() ArithmeticTranscript {
	return func() ArithmeticTranscript {
		return &messageCounter{state: uint64(startState), step: uint64(step)}
	}
}

func (m *messageCounter) incrementAndReturn() fr.Element {
	var res fr.Element
	res.SetUint64(m.state)
	m.state += m.step
	return res
}

func (m *messageCounter) NextFromElements(_ []fr.Element) fr.Element {
	return m.incrementAndReturn()
}
func (m *messageCounter) NextFromBytes(_ []byte) fr.Element {
	return m.incrementAndReturn()
}

type singleMultilinClaim struct {
	g polynomial.MultiLin
}

func (c singleMultilinClaim) ProveFinalEval(r []fr.Element) interface{} {
	return nil // verifier can compute the final eval itself
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

func sumForX1One(g polynomial.MultiLin) polynomial.Polynomial {
	sum := g[len(g)/2]
	for i := len(g)/2 + 1; i < len(g); i++ {
		sum.Add(&sum, &g[i])
	}
	return []fr.Element{sum}
}

func (c singleMultilinClaim) Combine(fr.Element) (SubClaim, polynomial.Polynomial) {
	sub := singleMultilinSubClaim{c.g.Clone()}

	return &sub, sumForX1One(c.g)
}

func (c *singleMultilinSubClaim) Next(r fr.Element) polynomial.Polynomial {
	c.g.Fold(r)
	return sumForX1One(c.g)
}

type singleMultilinLazyClaim struct {
	g          polynomial.MultiLin
	claimedSum fr.Element
}

func (c singleMultilinLazyClaim) VerifyFinalEval(r []fr.Element, combinationCoeff fr.Element, purportedValue fr.Element, proof interface{}) bool {
	val := c.g.Evaluate(r)
	return val.Equal(&purportedValue)
}

func (c singleMultilinLazyClaim) CombinedSum(combinationCoeffs fr.Element) fr.Element {
	return c.claimedSum
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

func testSumcheckSingleClaimMultilin(polyInt []uint64, hashGenerator func() ArithmeticTranscript) bool {
	poly := make(polynomial.MultiLin, len(polyInt))
	for i, n := range polyInt {
		poly[i].SetUint64(n)
	}

	claim := singleMultilinClaim{g: poly}

	proof := Prove(claim, hashGenerator(), []byte{})

	lazyClaim := singleMultilinLazyClaim{g: poly, claimedSum: poly.Sum()}

	return Verify(lazyClaim, proof, hashGenerator(), []byte{})
}

// For debugging TODO Remove
func printMsws(limit int) {
	var one, iElem fr.Element
	one.SetOne()

	for i := 1; i <= limit; i++ {
		iElem.Add(&iElem, &one)
		fmt.Printf("%d: %d\n", i, iElem[fr.Limbs-1])
	}
}

func TestSumcheckDeterministicHashSingleClaimMultilin(t *testing.T) {
	printMsws(36)

	polys := [][]uint64{
		{1, 2, 3, 4},             // 1 + 2X₁ + X₂
		{1, 2, 3, 4, 5, 6, 7, 8}, // 1 + 4X₁ + 2X₂ + X₃
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, // 1 + 8X₁ + 4X₂ + 2X₃ + X₄
	}

	const MaxStep = 4
	const MaxStart = 4
	hashGens := make([]func() ArithmeticTranscript, 0, MaxStart*MaxStep)

	for step := 0; step < MaxStep; step++ {
		for startState := 0; startState < MaxStart; startState++ {
			hashGens = append(hashGens, newMessageCounterGenerator(startState, step))
		}
	}

	for _, poly := range polys {
		for _, hashGen := range hashGens {
			if !testSumcheckSingleClaimMultilin(poly, hashGen) {
				t.Error(poly, hashGen())
				t.Fail()
			}
		}
	}
}
