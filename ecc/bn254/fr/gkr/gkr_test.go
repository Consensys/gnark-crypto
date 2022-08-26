package gkr

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
	"testing"
)

func TestNoGateTwoInstances(t *testing.T) {
	// Testing a single instance is not possible because the sumcheck implementation doesn't cover the trivial 0-variate case
	testNoGate(t, []fr.Element{four, three})
}

func TestNoGate(t *testing.T) {
	testManyInstances(t, 1, testNoGate)
}

func TestSingleMulGateTwoInstances(t *testing.T) {
	testSingleMulGate(t, []fr.Element{four, three}, []fr.Element{two, three})
}

func TestSingleMulGate(t *testing.T) {
	testManyInstances(t, 2, testSingleMulGate)
}

func TestSingleInputTwoIdentityGatesGateTwoInstances(t *testing.T) {

	testSingleInputTwoIdentityGates(t, []fr.Element{two, three})
}

func TestSingleInputTwoIdentityGatesGate(t *testing.T) {

	testManyInstances(t, 2, testSingleInputTwoIdentityGates)
}

func TestRecreateSumcheckErrorFromSingleInputTwoIdentityGatesGateTwoInstances(t *testing.T) {
	circuit := Circuit{{Wire{
		Gate:       nil,
		Inputs:     []*Wire{},
		NumOutputs: 2,
	}}}

	wire := &circuit[0][0]

	assignment := WireAssignment{&circuit[0][0]: []fr.Element{two, three}}

	claimsManagerGen := func() *claimsManager {
		manager := newClaimsManager(circuit, assignment)
		manager.add(wire, []fr.Element{three}, five)
		manager.add(wire, []fr.Element{four}, six)
		return &manager
	}

	transcriptGen := sumcheck.NewMessageCounterGenerator(4, 1)

	proof := sumcheck.Prove(claimsManagerGen().getClaim(wire), transcriptGen())
	sumcheck.Verify(claimsManagerGen().getLazyClaim(wire), proof, transcriptGen())
}

// Complete the circuit evaluation from input values
func (a WireAssignment) Complete(c Circuit) WireAssignment {
	numEvaluations := len(a[&c[len(c)-1][0]])

	for i := len(c) - 2; i >= 0; i-- { //there can only be input wires in the bottommost layer
		layer := c[i]
		for j := 0; j < len(layer); j++ {
			wire := &layer[j]

			if !wire.IsInput() {
				evals := make([]fr.Element, numEvaluations)
				ins := make([]fr.Element, len(wire.Inputs))
				for k := 0; k < numEvaluations; k++ {
					for inI, in := range wire.Inputs {
						ins[inI] = a[in][k]
					}
					evals[k] = wire.Gate.Evaluate(ins...)
				}
				a[wire] = evals
			}
		}
	}
	return a
}

var one, two, three, four, five, six fr.Element

func init() {
	one.SetOne()
	two.Double(&one)
	three.Add(&two, &one)
	four.Double(&two)
	five.Add(&three, &two)
	six.Double(&three)
}

func testManyInstances(t *testing.T, numInput int, test func(*testing.T, ...[]fr.Element)) {
	fullAssignments := make([][]fr.Element, numInput)
	maxSize := 16777216

	for i := range fullAssignments {
		fullAssignments[i] = make([]fr.Element, maxSize)
		setRandom(fullAssignments[i])
	}

	inputAssignments := make([][]fr.Element, numInput)
	for numEvals := 2; numEvals <= maxSize; numEvals *= 2 {
		for i, fullAssignment := range fullAssignments {
			inputAssignments[i] = fullAssignment[:numEvals]
		}

		test(t, inputAssignments...)
	}
}

func testNoGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := Circuit{
		{
			{
				Inputs:     []*Wire{},
				NumOutputs: 1,
				Gate:       nil,
			},
		},
	}

	assignment := WireAssignment{&c[0][0]: inputAssignments[0]}

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	// Even though a hash is called here, the proof is empty

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}
}

func testSingleMulGate(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
		{
			Inputs:     []*Wire{},
			NumOutputs: 1,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{{
		Inputs:     []*Wire{&c[1][0], &c[1][1]},
		NumOutputs: 1,
		Gate:       mulGate{},
	}}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0], &c[1][1]: inputAssignments[1]}.Complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(1, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Bad proof accepted")
	}
}

func testSingleInputTwoIdentityGates(t *testing.T, inputAssignments ...[]fr.Element) {
	c := make(Circuit, 2)

	c[1] = CircuitLayer{
		{
			Inputs:     []*Wire{},
			NumOutputs: 2,
			Gate:       nil,
		},
	}

	c[0] = CircuitLayer{
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       identityGate{},
		},
		{
			Inputs:     []*Wire{&c[1][0]},
			NumOutputs: 1,
			Gate:       identityGate{},
		},
	}

	assignment := WireAssignment{&c[1][0]: inputAssignments[0]}.Complete(c)

	proof := Prove(c, assignment, sumcheck.NewMessageCounter(0, 1))

	if !Verify(c, assignment, proof, sumcheck.NewMessageCounter(0, 1)) {
		t.Error("Proof rejected")
	}

	if Verify(c, assignment, proof, sumcheck.NewMessageCounter(1, 1)) {
		t.Error("Bad proof accepted")
	}
}

type mulGate struct{}

func (m mulGate) Evaluate(element ...fr.Element) (result fr.Element) {
	result.Mul(&element[0], &element[1])
	return
}

func (m mulGate) Degree() int {
	return 2
}

func setRandom(slice []fr.Element) {
	for i := range slice {
		slice[i].SetRandom()
	}
}
