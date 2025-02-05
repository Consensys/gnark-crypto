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

// arya
// The GKR gates needed for proving Poseidon2 permutations
// TODO @Tabaie @ThomasPiellard generify once Poseidon2 parameters are known for all curves

// extKeyGate applies the external matrix mul, then adds the round key, then applies the sBox
// because of its symmetry, we don't need to define distinct x1 and x2 versions of it
type extKeyGate struct {
	roundKey fr.Element
}

func (g *extKeyGate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}

	x[0].
		Double(&x[0]).
		Add(&x[0], &x[1]).
		Add(&x[0], &g.roundKey)
	return x[0]
}

func (g *extKeyGate) Degree() int {
	return 1
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
type intKeyGate2 struct {
	roundKey fr.Element
}

func (g *intKeyGate2) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 inputs")
	}
	x[0].Add(&x[0], &x[1])
	x[1].
		Double(&x[1]).
		Add(&x[1], &x[0]).
		Add(&x[1], &g.roundKey)

	return x[1]
}

func (g *intKeyGate2) Degree() int {
	return 1
}

type pow4Gate struct{}

func (g pow4Gate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 1 {
		panic("expected 1 input")
	}
	x[0].Square(&x[0]).Square(&x[0])
	return x[0]
}

func (g pow4Gate) Degree() int {
	return 4
}

type pow4TimesGate struct{}

type pow2Gate struct{}

func (g pow2Gate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 1 {
		panic("expected 1 input")
	}
	x[0].Square(&x[0])
	return x[0]
}

func (g pow2Gate) Degree() int {
	return 2
}

type pow2TimesGate struct{}

func (g pow2TimesGate) Degree() int {
	return 3
}

func (g pow2TimesGate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 2 input")
	}
	x[0].Square(&x[0]).Mul(&x[0], &x[1])
	return x[0]
}

func (g pow4TimesGate) Evaluate(x ...fr.Element) fr.Element {
	if len(x) != 2 {
		panic("expected 1 input")
	}
	x[0].Square(&x[0]).Square(&x[0]).Mul(&x[0], &x[1])
	return x[0]
}

func (g pow4TimesGate) Degree() int {
	return 5
}

func varIndex(varName string) int {
	switch varName {
	case "x":
		return 0
	case "y":
		return 1
	default:
		panic("unexpected varName")
	}
}

var initOnce sync.Once

// RegisterGkrGates registers the Poseidon2 permutation gates for GKR
func RegisterGkrGates() {
	initOnce.Do(
		func() {
			p := poseidon2.NewDefaultParameters()
			halfRf := p.NbFullRounds / 2
			params := p.String()

			gkr.Gates["pow2"] = pow2Gate{}
			gkr.Gates["pow4"] = pow4Gate{}
			gkr.Gates["pow2Times"] = pow2TimesGate{}
			gkr.Gates["pow4Times"] = pow4TimesGate{}

			gateNameLinear := func(varName string, i int) string {
				return fmt.Sprintf("%s-l-op-round=%d;%s", varName, i, params)
			}

			gateNameIntegrated := func(varName string, i int) string {
				return fmt.Sprintf("%s-i-op-round=%d;%s", varName, i, params)
			}

			extKeySBox := func(round int, varName string) {
				vI := varIndex(varName)
				gkr.Gates[gateNameIntegrated(varName, round)] = &extKeySBoxGate{ // in case we use an integrated S-box
					roundKey: p.RoundKeys[round][vI],
				}
				gkr.Gates[gateNameLinear(varName, round)] = &extKeyGate{ // in case we use a separate S-box
					roundKey: p.RoundKeys[round][vI],
				}
			}

			intKeySBox2 := func(round int) {
				gkr.Gates[gateNameLinear("y", round)] = &intKeyGate2{
					roundKey: p.RoundKeys[round][1],
				}
				gkr.Gates[gateNameIntegrated("y", round)] = &intKeySBoxGate2{
					roundKey: p.RoundKeys[round][1],
				}
			}

			fullRound := func(i int) {
				extKeySBox(i, "x")
				extKeySBox(i, "y")
			}

			for i := range halfRf {
				fullRound(i)
			}

			{ // i = halfRf: first partial round
				extKeySBox(halfRf, "x")
				gkr.Gates[gateNameLinear("y", halfRf)] = extGate2{}
			}

			for i := halfRf + 1; i < halfRf+p.NbPartialRounds; i++ {
				extKeySBox(i, "x") // for x1, intKeySBox is identical to extKeySBox
				gkr.Gates[gateNameLinear("y", i)] = intGate2{}
			}

			{
				i := halfRf + p.NbPartialRounds
				extKeySBox(i, "x")
				intKeySBox2(i)
			}

			for i := halfRf + p.NbPartialRounds + 1; i < p.NbPartialRounds+p.NbFullRounds; i++ {
				fullRound(i)
			}

			gkr.Gates[gateNameLinear("y", p.NbPartialRounds+p.NbFullRounds)] = extGate{}
		},
	)
}
