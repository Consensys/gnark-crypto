package gkr

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/sumcheck"
	"testing"
)

func TestSingleIdentityGateTwoInstances(t *testing.T) {
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
