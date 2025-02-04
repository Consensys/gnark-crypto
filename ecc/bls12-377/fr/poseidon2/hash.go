package poseidon2

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/gkr"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark-crypto/utils"
	"hash"
	"strconv"
	"strings"
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
	return sBox2(x[0])
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

	return sBox2(x[1])
}

func (g *intKeyGate2) Degree() int {
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

// sBox2 is Hash.sBox for t=2
func sBox2(x fr.Element) fr.Element {
	var y fr.Element
	y.Square(&x).Square(&y).Square(&y).Square(&y).Mul(&x, &y)
	return y
}

func gateName(prefix string, i ...int) string {
	return fmt.Sprintf("%s-round=%s;%s", prefix, strings.Join(utils.Map(i, strconv.Itoa), "-"), seed)
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

func DefineGkrGates() {
	p := params()
	halfRf := p.rF / 2

	sBox := func(round int, varName string) {
		gkr.Gates[gateName(varName, round, 1)] = pow4Gate{}

		gkr.Gates[gateName(varName, round, 2)] = pow4TimesGate{}
	}

	extKeySBox := func(round int, varName string) {
		gkr.Gates[gateName(varName, round, 0)] = &extKeyGate{
			roundKey: p.roundKeys[round][varIndex(varName)],
		}
		sBox(round, varName)
	}

	intKeySBox2 := func(round int) {
		gate := gateName("y", round, 0)
		gkr.Gates[gate] = &intKeyGate2{
			roundKey: p.roundKeys[round][1],
		}
		sBox(round, "y")
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
		gkr.Gates[gateName("y", halfRf)] = extGate2{}
	}

	for i := halfRf + 1; i < halfRf+p.rP; i++ {
		extKeySBox(i, "x") // for x1, intKeySBox is identical to extKeySBox
		gkr.Gates[gateName("y", i)] = intGate2{}
	}

	{
		i := halfRf + p.rP
		extKeySBox(i, "x")
		intKeySBox2(i)
	}

	for i := halfRf + p.rP + 1; i < p.rP+p.rF; i++ {
		fullRound(i)
	}

	gkr.Gates[gateName("y", p.rP+p.rF)] = extGate{}
}
