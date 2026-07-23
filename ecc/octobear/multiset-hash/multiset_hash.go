package multisethash

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/octobear"
)

const tweakBound = 256

var errMapFailure = errors.New("octobear multiset hash: failed to map message in tweak window")

// Accumulator stores an additive multiset hash state in affine coordinates.
type Accumulator struct {
	sum octobear.G1Affine
}

// NewAccumulator returns a zero accumulator.
func NewAccumulator() Accumulator {
	var a Accumulator
	a.sum.SetInfinity()
	return a
}

// Insert maps msg to octobear and adds it to the accumulator.
func (a *Accumulator) Insert(msg uint16) error {
	p, _, err := Map(msg)
	if err != nil {
		return err
	}
	a.sum.Add(&a.sum, &p)
	return nil
}

// Remove maps msg to octobear and subtracts it from the accumulator.
func (a *Accumulator) Remove(msg uint16) error {
	p, _, err := Map(msg)
	if err != nil {
		return err
	}
	p.Neg(&p)
	a.sum.Add(&a.sum, &p)
	return nil
}

// Digest returns the current accumulator state in affine coordinates.
func (a *Accumulator) Digest() octobear.G1Affine {
	return a.sum
}

// Reset clears the accumulator.
func (a *Accumulator) Reset() {
	a.sum.SetInfinity()
}

// Hash returns the multiset hash of msgs.
func Hash(msgs []uint16) (octobear.G1Affine, error) {
	acc := NewAccumulator()
	for _, msg := range msgs {
		if err := acc.Insert(msg); err != nil {
			return octobear.G1Affine{}, err
		}
	}
	return acc.Digest(), nil
}

// Map deterministically maps msg to a point on octobear using the y-increment method.
// It returns the mapped point and the first offset k in [0, 255] such that
// y = msg*256 + k yields a point (x, y) on octobear.
func Map(msg uint16) (octobear.G1Affine, uint8, error) {
	_, b := octobear.CurveCoefficients()
	return mapAtBase(uint64(msg)*tweakBound, tweakBound, &b)
}
