package gkr

import (
	fiatshamir "github.com/consensys/gnark-crypto/fiat-shamir"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational"
	"github.com/consensys/gnark-crypto/internal/generator/test_vector_utils/small_rational/test_vector_utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

var two = test_vector_utils.ToElement(2)
var three = test_vector_utils.ToElement(3)
var four = test_vector_utils.ToElement(4)

func TestSingleMulGateTwoInstances(t *testing.T) {
	testSingleMulGate(t, []small_rational.SmallRational{*four, *three}, []small_rational.SmallRational{*two, *three})
}

func testSingleMulGate(t *testing.T, inputAssignments ...[]small_rational.SmallRational) {

	c := make(Circuit, 3)
	c[2] = Wire{
		Gate:   mulGate{},
		Inputs: []*Wire{&c[0], &c[1]},
	}

	assignment := WireAssignment{&c[0]: inputAssignments[0], &c[1]: inputAssignments[1]}.Complete(c)

	proof, err := Prove(c, assignment, fiatshamir.WithHash(test_vector_utils.NewMessageCounter(1, 1)))
	assert.NoError(t, err)

	err = Verify(c, assignment, proof, fiatshamir.WithHash(test_vector_utils.NewMessageCounter(1, 1)))
	assert.NoError(t, err, "proof rejected")

	err = Verify(c, assignment, proof, fiatshamir.WithHash(test_vector_utils.NewMessageCounter(0, 1)))
	assert.NotNil(t, err, "bad proof accepted")
}

func TestSingleMulGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMulGate)
}

func testManyInstances(t *testing.T, numInput int, test func(*testing.T, ...[]small_rational.SmallRational)) {
	fullAssignments := make([][]small_rational.SmallRational, numInput)
	maxSize := 1 << 2

	t.Log("Entered test orchestrator, assigning and randomizing inputs")

	for i := range fullAssignments {
		fullAssignments[i] = make([]small_rational.SmallRational, maxSize)
		setRandom(fullAssignments[i])
	}

	inputAssignments := make([][]small_rational.SmallRational, numInput)
	for numEvals := maxSize; numEvals <= maxSize; numEvals *= 2 {
		for i, fullAssignment := range fullAssignments {
			inputAssignments[i] = fullAssignment[:numEvals]
		}

		t.Log("Selected inputs for test")
		test(t, inputAssignments...)
	}
}

func setRandom(slice []small_rational.SmallRational) {
	for i := range slice {
		slice[i].SetUint64(uint64(i))
	}
}

type mulGate struct{}

func (g mulGate) Evaluate(element ...small_rational.SmallRational) (result small_rational.SmallRational) {
	result.Mul(&element[0], &element[1])
	return
}

func (g mulGate) Degree() int {
	return 2
}
