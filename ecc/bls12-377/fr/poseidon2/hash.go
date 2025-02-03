package poseidon2

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/gkr"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
	"hash"
)

// NewPoseidon2 returns a Poseidon2 hasher
// TODO @Tabaie @ThomasPiellard Generify once Poseidon2 parameters are known for all curves
func NewPoseidon2() gnarkHash.StateStorer {
	return gnarkHash.NewMerkleDamgardHasher(
		&Hash{params: params()}, make([]byte, fr.Bytes))
}

const (
	seed = "Poseidon2 hash for BLS12_377 with t=2, rF=6, rP=26, d=17"
	d    = 17
)

func params() parameters {
	return parameters{
		t:         2,
		rF:        6,
		rP:        26,
		roundKeys: InitRC(seed, 6, 26, 2),
	}
}

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_BLS12_377, func() hash.Hash {
		return NewPoseidon2()
	})
}

// The GKR gates needed for proving Poseidon2 permutations
// TODO @Tabaie @ThomasPiellard generify once Poseidon2 parameters are known for all curves

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
	return powerFr(x[0], d)
}

func (g *extKeySBoxGate) Degree() int {
	return d
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

	return powerFr(x[1], 17)
}

func (g *intKeySBoxGate2) Degree() int {
	return d
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

func powerFr(x fr.Element, n int) fr.Element {
	tmp := x
	switch n {
	case 3:
		x.Square(&x).Mul(&tmp, &x)
	case 5:
		x.Square(&x).Square(&x).Mul(&x, &tmp)
	case 7:
		x.Square(&x).Mul(&x, &tmp).Square(&x).Mul(&x, &tmp)
	case 17:
		x.Square(&x).Square(&x).Square(&x).Square(&x).Mul(&x, &tmp)
	case -1:
		x.Inverse(&x)
	default:
		panic("unknown sBox degree")
	}
	return x
}

func DefineGkrGates() {
	p := params()
	halfRf := p.rF / 2

	gateNameX := func(i int) string {
		return fmt.Sprintf("x-round=%d%s", i, seed)
	}
	gateNameY := func(i int) string {
		return fmt.Sprintf("y-round=%d%s", i, seed)
	}

	fullRound := func(i int) {
		gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
			roundKey: p.roundKeys[i][0],
		}

		gkr.Gates[gateNameY(i)] = &extKeySBoxGate{
			roundKey: p.roundKeys[i][1],
		}
	}

	for i := range halfRf {
		fullRound(i)
	}

	{ // i = halfRf: first partial round
		i := halfRf
		gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
			roundKey: p.roundKeys[i][0],
		}

		gkr.Gates[gateNameY(i)] = extGate2{}
	}

	for i := halfRf + 1; i < halfRf+p.rP; i++ {
		gkr.Gates[gateNameX(i)] = &extKeySBoxGate{ // for x1, intKeySBox is identical to extKeySBox
			roundKey: p.roundKeys[i][0],
		}

		gkr.Gates[gateNameY(i)] = intGate2{}

	}

	{
		i := halfRf + p.rP
		gkr.Gates[gateNameX(i)] = &extKeySBoxGate{
			roundKey: p.roundKeys[i][0],
		}

		gkr.Gates[gateNameY(i)] = &intKeySBoxGate2{
			roundKey: p.roundKeys[i][1],
		}
	}

	for i := halfRf + p.rP + 1; i < p.rP+p.rF; i++ {
		fullRound(i)
	}

	gkr.Gates[gateNameY(p.rP+p.rF)] = extGate{}
}
