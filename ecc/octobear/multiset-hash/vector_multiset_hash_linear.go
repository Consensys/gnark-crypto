package multisethash

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/octobear"
	"github.com/consensys/gnark-crypto/field/koalabear/extensions"
)

// Linear-separator vector ECMSH (Section 4.3 "linear separator").
//
// The digest is a vector of N = 23 ECMSH accumulators. Coordinate i uses
// the y-increment relation with per-coordinate slot s_i = msg + i*M:
//
//	y_i(msg, k) = T * (msg + i*M) + k,   k in [0, T)
//
// With T = 128 and M = 2^18, the encoded ordinates satisfy
// y_i < N*M*T = 23 * 2^18 * 128 < p/2, so the image is inverse-free.
const (
	linearN = 23
	linearT = 128
	linearM = 1 << 18
)

var errLinearMsgOutOfRange = fmt.Errorf("octobear vector multiset hash: linear message must be < 2^18 (= %d)", linearM)

// LinearAccumulator holds the N affine accumulator points for the
// linear-separator vector ECMSH.
type LinearAccumulator struct {
	sum [linearN]octobear.G1Affine
}

// NewLinearAccumulator returns a zero (all-infinity) LinearAccumulator.
func NewLinearAccumulator() LinearAccumulator {
	var a LinearAccumulator
	for i := range a.sum {
		a.sum[i].SetInfinity()
	}
	return a
}

// Insert maps msg to N curve points using the linear domain separator and
// adds each point to the corresponding accumulator coordinate.
func (a *LinearAccumulator) Insert(msg uint32) error {
	pts, _, err := MapLinear(msg)
	if err != nil {
		return err
	}
	for i := range a.sum {
		a.sum[i].Add(&a.sum[i], &pts[i])
	}
	return nil
}

// Remove maps msg to N curve points and subtracts each from the
// corresponding accumulator coordinate.
func (a *LinearAccumulator) Remove(msg uint32) error {
	pts, _, err := MapLinear(msg)
	if err != nil {
		return err
	}
	var neg octobear.G1Affine
	for i := range a.sum {
		neg.Neg(&pts[i])
		a.sum[i].Add(&a.sum[i], &neg)
	}
	return nil
}

// Digest returns the current vector of accumulator points.
func (a *LinearAccumulator) Digest() [linearN]octobear.G1Affine {
	return a.sum
}

// Reset clears the accumulator to the all-infinity state.
func (a *LinearAccumulator) Reset() {
	for i := range a.sum {
		a.sum[i].SetInfinity()
	}
}

// HashLinear returns the linear-separator vector ECMSH of msgs.
func HashLinear(msgs []uint32) ([linearN]octobear.G1Affine, error) {
	acc := NewLinearAccumulator()
	for _, msg := range msgs {
		if err := acc.Insert(msg); err != nil {
			return [linearN]octobear.G1Affine{}, err
		}
	}
	return acc.Digest(), nil
}

// MapLinear deterministically maps msg to N curve points using the linear
// domain separator y_i(msg, k) = T*(msg + i*M) + k. It returns the N points
// and the offsets k_i in [0, T) that produced them.
func MapLinear(msg uint32) ([linearN]octobear.G1Affine, [linearN]uint8, error) {
	var (
		pts     [linearN]octobear.G1Affine
		offsets [linearN]uint8
	)
	if uint64(msg) >= linearM {
		return pts, offsets, errLinearMsgOutOfRange
	}
	_, b := octobear.CurveCoefficients()
	for i := 0; i < linearN; i++ {
		baseY := (uint64(msg) + uint64(i)*linearM) * linearT
		p, k, err := mapAtBase(baseY, linearT, &b)
		if err != nil {
			return pts, offsets, err
		}
		pts[i] = p
		offsets[i] = k
	}
	return pts, offsets, nil
}

// mapAtBase scans k in [0, bound) and returns the first curve point
// whose ordinate is y = baseY + k in the base subfield. baseY + bound
// must remain strictly below p/2 to keep the image inverse-free.
func mapAtBase(baseY uint64, bound uint64, b *extensions.E8) (octobear.G1Affine, uint8, error) {
	for k := uint64(0); k < bound; k++ {
		var y, c, ySquared extensions.E8
		y.C0.B0.A0.SetUint64(baseY + k)

		ySquared.Square(&y)
		c.Sub(b, &ySquared)

		x, ok := depressedCubicRoot(c)
		if !ok {
			continue
		}

		p := octobear.G1Affine{X: x, Y: y}
		if p.IsOnCurve() && p.IsInSubGroup() {
			return p, uint8(k), nil
		}
	}
	return octobear.G1Affine{}, 0, errMapFailure
}
