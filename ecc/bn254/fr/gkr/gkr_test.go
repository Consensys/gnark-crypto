package gkr

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
	"testing"
)

func TestNoGateTwoInstances(t *testing.T) {
	// Testing a single instance is not possible because the sumcheck implementation doesn't cover the trivial 0-variate case

	c := Circuit{
		{
			{
				Inputs:     []*Wire{},
				NumOutputs: 1,
				Gate:       nil,
			},
		},
	}

	var four, three fr.Element
	three.SetUint64(3)
	four.SetUint64(4)
	assignment := WireAssignment{&c[0][0]: []fr.Element{four, three}}

	proof := Prove(c, assignment, &messageCounter{state: 1, step: 1}, []byte{})

	if !Verify(c, assignment, proof, &messageCounter{state: 1, step: 1}, []byte{}) {
		t.Error("Proof rejected")
	}
}

func TestSingleMulGateTwoInstances(t *testing.T) {
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

	var two, three, four fr.Element
	two.SetUint64(2)
	three.SetUint64(3)
	four.SetUint64(4)
	assignment := WireAssignment{&c[1][0]: []fr.Element{four, three}, &c[1][1]: []fr.Element{two, three}}.Complete(c)

	proof := Prove(c, assignment, &messageCounter{state: 1, step: 1}, []byte{})

	if !Verify(c, assignment, proof, &messageCounter{state: 1, step: 1}, []byte{}) {
		t.Error("Proof rejected")
	}
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

// This is a very bad fiat-shamir challenge generator, and copy-pasted from sumcheck
type messageCounter struct {
	state uint64
	step  uint64
}

func newMessageCounterGenerator(startState, step int) func() sumcheck.ArithmeticTranscript {
	return func() sumcheck.ArithmeticTranscript {
		return &messageCounter{state: uint64(startState), step: uint64(step)}
	}
}

func (m *messageCounter) incrementAndReturn() fr.Element {
	var res fr.Element
	res.SetUint64(m.state)
	fmt.Println("Hash returning", m.state)
	m.state += m.step
	return res
}

func (m *messageCounter) NextFromElements(_ []fr.Element) fr.Element {
	return m.incrementAndReturn()
}
func (m *messageCounter) NextFromBytes(_ []byte) fr.Element {
	return m.incrementAndReturn()
}

type mulGate struct{}

func (m mulGate) Evaluate(element ...fr.Element) (result fr.Element) {
	result.Mul(&element[0], &element[1])
	return
}

func (m mulGate) Degree() int {
	return 2
}
