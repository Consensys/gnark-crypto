package multisethash

import (
	"errors"

	"github.com/consensys/gnark-crypto/ecc/kb8"
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

var (
	errMapFailure = errors.New("kb8 multiset hash: failed to map message after 256 offsets")
	mapPerm       = poseidon2.NewPermutation(16, 6, 21)
)

// Accumulator stores an additive multiset hash state in affine coordinates.
// Updates are carried out in Jacobian coordinates and normalized back to affine.
type Accumulator struct {
	sum kb8.G1Affine
}

// NewAccumulator returns a zero accumulator.
func NewAccumulator() Accumulator {
	var a Accumulator
	a.sum.SetInfinity()
	return a
}

// Insert maps msg to kb8 and adds it to the accumulator.
func (a *Accumulator) Insert(msg *extensions.E8) error {
	p, _, err := Map(msg)
	if err != nil {
		return err
	}
	var sumJac kb8.G1Jac
	sumJac.FromAffine(&a.sum).AddMixed(&p)
	a.sum.FromJacobian(&sumJac)
	return nil
}

// Remove maps msg to kb8 and subtracts it from the accumulator.
func (a *Accumulator) Remove(msg *extensions.E8) error {
	p, _, err := Map(msg)
	if err != nil {
		return err
	}
	p.Neg(&p)
	var sumJac kb8.G1Jac
	sumJac.FromAffine(&a.sum).AddMixed(&p)
	a.sum.FromJacobian(&sumJac)
	return nil
}

// Digest returns the current accumulator state in affine coordinates.
func (a *Accumulator) Digest() kb8.G1Affine {
	return a.sum
}

// Reset clears the accumulator.
func (a *Accumulator) Reset() {
	a.sum.SetInfinity()
}

// Hash returns the multiset hash of msgs.
func Hash(msgs []extensions.E8) (kb8.G1Affine, error) {
	acc := NewAccumulator()
	for i := range msgs {
		if err := acc.Insert(&msgs[i]); err != nil {
			return kb8.G1Affine{}, err
		}
	}
	return acc.Digest(), nil
}

// Map deterministically maps msg to a point on kb8 using KoalaBear Poseidon2.
// It returns the mapped point and the offset in [0, 255] that produced it.
func Map(msg *extensions.E8) (kb8.G1Affine, uint8, error) {
	var state [16]koalabear.Element
	messageToState(msg, state[:8])

	a, b := kb8.CurveCoefficients()
	for offset := uint16(0); offset < 256; offset++ {
		state[8].SetUint64(uint64(offset))
		clearStateSuffix(state[9:])

		if err := mapPerm.Permutation(state[:]); err != nil {
			return kb8.G1Affine{}, 0, err
		}

		x := stateToE8(state[:8])
		var rhs, y, tmp extensions.E8
		rhs.Square(&x).Mul(&rhs, &x)
		tmp.Mul(&x, &a)
		rhs.Add(&rhs, &tmp).Add(&rhs, &b)
		if rhs.Legendre() != 1 {
			continue
		}
		y.Sqrt(&rhs)
		if y.LexicographicallyLargest() {
			y.Neg(&y)
		}

		p := kb8.G1Affine{X: x, Y: y}
		if p.IsOnCurve() && p.IsInSubGroup() {
			return p, uint8(offset), nil
		}
	}

	return kb8.G1Affine{}, 0, errMapFailure
}

func clearStateSuffix(s []koalabear.Element) {
	for i := range s {
		s[i].SetZero()
	}
}

func messageToState(msg *extensions.E8, out []koalabear.Element) {
	out[0] = msg.C0.B0.A0
	out[1] = msg.C0.B0.A1
	out[2] = msg.C0.B1.A0
	out[3] = msg.C0.B1.A1
	out[4] = msg.C1.B0.A0
	out[5] = msg.C1.B0.A1
	out[6] = msg.C1.B1.A0
	out[7] = msg.C1.B1.A1
}

func stateToE8(in []koalabear.Element) extensions.E8 {
	var x extensions.E8
	x.C0.B0.A0 = in[0]
	x.C0.B0.A1 = in[1]
	x.C0.B1.A0 = in[2]
	x.C0.B1.A1 = in[3]
	x.C1.B0.A0 = in[4]
	x.C1.B0.A1 = in[5]
	x.C1.B1.A0 = in[6]
	x.C1.B1.A1 = in[7]
	return x
}
