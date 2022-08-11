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

func (m *messageCounter) incrementAndReturn() fr.Element {
	var res fr.Element
	res.SetUint64(m.state)
	fmt.Println("Hash = ", m.state)
	m.state += m.step
	return res
}

func (m *messageCounter) NextFromElements(_ []fr.Element) fr.Element {
	return m.incrementAndReturn()
}
func (m *messageCounter) NextFromBytes(_ []byte) fr.Element {
	return m.incrementAndReturn()
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

func sumForX1One(g polynomial.MultiLin) polynomial.Polynomial {
	sum := g[len(g)/2]
	for i := len(g)/2 + 1; i < len(g); i++ {
		sum.Add(&sum, &g[i])
	}
	return []fr.Element{sum}
}

func (c singleMultilinClaim) Combine(fr.Element) (SubClaim, polynomial.Polynomial) {
	sub := singleMultilinSubClaim{c.g.Clone()}

	return sub, sumForX1One(c.g)
}

func (c singleMultilinSubClaim) Next(r fr.Element) polynomial.Polynomial {
	fmt.Println("Prover next called")
	c.g.Fold(r)
	return sumForX1One(c.g)
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

func testSumcheckDeterministicHashSingleClaimMultilin(polyInt []uint64, hash messageCounter) bool {
	poly := make(polynomial.MultiLin, len(polyInt))
	for i, n := range polyInt {
		poly[i].SetUint64(n)
	}

	claim := singleMultilinClaim{g: poly}

	workingHash := hash
	proof := Prove(claim, &workingHash, []byte{})

	fmt.Println("Verify")

	lazyClaim := singleMultilinLazyClaim{g: poly, claimedSum: poly.Sum()}

	workingHash = hash
	return Verify(lazyClaim, proof, &workingHash, []byte{})
}

func printMsws(limit int) {
	var one, iElem fr.Element
	one.SetOne()

	for i := 1; i <= limit; i++ {
		iElem.Add(&iElem, &one)
		fmt.Printf("%d: %d\n", i, iElem[fr.Limbs-1])
	}
}

func TestSumcheckDeterministicHashSingleClaimMultilin(t *testing.T) {
	printMsws(10)

	poly := []uint64{1, 2, 3, 4} // 1 + 2X₁ + X₂
	if !testSumcheckDeterministicHashSingleClaimMultilin(poly, messageCounter{1, 0}) {
		t.Fail()
	}
}
