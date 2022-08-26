package sumcheck

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// This is an implementation of Fiat-Shamir optimized for in-circuit verifiers.
// i.e. the hashes used operate on and return field elements.

type ArithmeticTranscript interface {
	Update(...interface{})
	Next(...interface{}) fr.Element
	NextN(int, ...interface{}) []fr.Element
}

// This is a very bad fiat-shamir challenge generator
type MessageCounter struct {
	state   uint64
	step    uint64
	updated bool
}

func (m *MessageCounter) Update(i ...interface{}) {
	m.state += m.step
	m.updated = true
}

func (m *MessageCounter) Next(i ...interface{}) (challenge fr.Element) {
	if !m.updated || len(i) != 0 {
		m.Update(i)
	}
	fmt.Println("hash returning", m.state)
	challenge.SetUint64(m.state)
	m.updated = false
	return
}

func (m *MessageCounter) NextN(N int, i ...interface{}) (challenges []fr.Element) {
	challenges = make([]fr.Element, N)
	for n := 0; n < N; n++ {
		challenges[n] = m.Next(i)
		i = []interface{}{}
	}
	return
}

func NewMessageCounter(startState, step int) ArithmeticTranscript {
	transcript := &MessageCounter{state: uint64(startState), step: uint64(step)}
	transcript.Update([]byte{})
	return transcript
}

func NewMessageCounterGenerator(startState, step int) func() ArithmeticTranscript {
	return func() ArithmeticTranscript {
		return NewMessageCounter(startState, step)
	}
}
