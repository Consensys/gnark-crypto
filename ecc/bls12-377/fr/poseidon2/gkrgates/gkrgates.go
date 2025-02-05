// Package gkrgates implements the Poseidon2 permutation gate for GKR
//
// This implementation is based on the [poseidon2] package, but exposes the
// primitives as gates for inclusion in GKR circuits.

// TODO(@Tabaie @ThomasPiellard) generify once Poseidon2 parameters are known for all curves
package gkrgates

import (
	"fmt"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/gkr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
)

// The GKR gates needed for proving Poseidon2 permutations

// extKeySBoxGate applies the external matrix mul, then adds the round key, then applies the sBox
// because of its symmetry, we don't need to define distinct x1 and x2 versions of it
type extKeySBoxGate struct {
	roundKey fr.Element
}

func (g *extKeySBoxGate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}

	x[0].
		Double(&x[0]).
		Add(&x[0], &x[1]).
		Add(&x[0], &g.roundKey)
	return sBox2(x[0])
}

func (g *extKeySBoxGate) Degree() int {
	return poseidon2.DegreeSBox()
}

// for x1, the partial round gates are identical to full round gates
// for x2, the partial round gates are just a linear combination
// TODO @Tabaie eliminate the x2 partial round gates and have the x1 gates depend on i - rf/2 or so previous x1's

// extGate2 applies the external matrix mul, outputting the second element of the result
type extGate2 struct{}

func (extGate2) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	x[1].
		Double(&x[1]).
		Add(&x[1], &x[0])
	return x[1]
}

func (g extGate2) Degree() int {
	return 1
}

// intGate2 applies the internal matrix mul, returning the second element
type intGate2 struct {
}

func (g intGate2) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	x[0].Add(&x[0], &x[1])
	x[1].
		Double(&x[1]).
		Add(&x[1], &x[0])
	return x[1]
}

func (g intGate2) Degree() int {
	return 1
}

// intKeySBoxGateFr applies the second row of internal matrix mul, then adds the round key, then applies the sBox
type intKeySBoxGate2 struct {
	roundKey fr.Element
}

func (g *intKeySBoxGate2) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	x[0].Add(&x[0], &x[1])
	x[1].
		Double(&x[1]).
		Add(&x[1], &x[0]).
		Add(&x[1], &g.roundKey)

	return sBox2(x[1])
}

func (g *intKeySBoxGate2) Degree() int {
	return poseidon2.DegreeSBox()
}

type extGate struct{}

func (g extGate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	x[0].
		Double(&x[0]).
		Add(&x[0], &x[1])
	return x[0]
}

func (g extGate) Degree() int {
	return 1
}

// sBox2 is Hash.sBox for t=2
func sBox2(x fr.Element) fr.Element {
	var y fr.Element
	y.Square(&x).Square(&y).Square(&y).Square(&y).Mul(&x, &y)
	return y
}

var initOnce sync.Once

// RegisterGkrGates registers the Poseidon2 permutation gates for GKR
func RegisterGkrGates() {
	initOnce.Do(
		func() {
			p := poseidon2.NewDefaultParameters()
			halfRf := p.NbFullRounds / 2

			gateNameX := func(i int) string {
				return fmt.Sprintf("x-round=%d%s", i, p.String())
			}
			gateNameY := func(i int) string {
				return fmt.Sprintf("y-round=%d%s", i, p.String())
			}

			fullRound := func(i int) {
				gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
					roundKey: p.RoundKeys[i][0],
				}

				gkr.Gates[gateNameY(i)] = &extKeySBoxGate{
					roundKey: p.RoundKeys[i][1],
				}
			}

			for i := range halfRf {
				fullRound(i)
			}

			{ // i = halfRf: first partial round
				i := halfRf
				gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
					roundKey: p.RoundKeys[i][0],
				}

				gkr.Gates[gateNameY(i)] = extGate2{}
			}

			for i := halfRf + 1; i < halfRf+p.NbPartialRounds; i++ {
				gkr.Gates[gateNameX(i)] = &extKeySBoxGate{ // for x1, intKeySBox is identical to extKeySBox
					roundKey: p.RoundKeys[i][0],
				}

				gkr.Gates[gateNameY(i)] = intGate2{}

			}

			{
				i := halfRf + p.NbPartialRounds
				gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
					roundKey: p.RoundKeys[i][0],
				}

				gkr.Gates[gateNameY(i)] = &intKeySBoxGate2{
					roundKey: p.RoundKeys[i][1],
				}
			}

			for i := halfRf + p.NbPartialRounds + 1; i < p.NbPartialRounds+p.NbFullRounds; i++ {
				fullRound(i)
			}

			gkr.Gates[gateNameY(p.NbPartialRounds+p.NbFullRounds)] = extGate{}
		},
	)
}
