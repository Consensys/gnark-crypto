package bw6_761

import (
	"math/big"
	"sync"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bw6_761/fp"
	"github.com/consensys/gurvy/utils"
)

// E: y**2=x**3-1
// Etwist: y**2 = x**3+4

var bw6_761 Curve
var initOnce sync.Once

const ID = gurvy.BW6_761

// parameters for pippenger ScalarMulByGen
const sGen = 4
const bGen = sGen

type PairingResult = E6

// BW6_761 returns BW6_761 curve
func BW6_761() *Curve {
	initOnce.Do(initBW6_761)
	return &bw6_761
}

// Curve represents the BW6_761 curve and pre-computed constants
type Curve struct {
	B fp.Element // A, B coefficients of the curve x^3 = y^2 +AX+b

	g1Gen G1Jac // generator of torsion group G1Jac
	g2Gen G2Jac // generator of torsion group G2Jac

	g1Infinity G1Jac // infinity (in Jacobian coords)
	g2Infinity G2Jac

	// TODO store this number as a MAX_SIZE constant, or with build tags
	loopCounter [64]int8 // NAF decomposition of t-1, t is the trace of the Frobenius restricted on the r torsion group

	// precomputed values for ScalarMulByGen
	tGenG1 [((1 << bGen) - 1)]G1Jac
	tGenG2 [((1 << bGen) - 1)]G2Jac
}

func initBW6_761() {

	// A, B coeffs of the curve in Mont form
	bw6_761.B.SetUint64(4)

	// Setting G1Jac
	bw6_761.g1Gen.X.SetString("5492337019202608651620810666633622531924946248948182754748114963334556774714407693672822645637243083342924475378144397780999025266189779523629084326871556483802038026432771927197170911996417793635501066231650458516636932478125208")
	bw6_761.g1Gen.Y.SetString("4874298780810344118673004453041997030286302865034758641338313952140849332867290574388366379298818956144982860224857872858166812124104845663394852158352478303048122861831479086904887356602146134586313962565783961814162269209043907")
	bw6_761.g1Gen.Z.SetString("1")

	// Setting G2Jac
	bw6_761.g2Gen.X.SetString("5779457169892140542970811884673908634889239063901429247094594197042136765689827803062459420720318762253427359282239252479201196985966853806926626938528693270647807548111019296972244105103687281416386903420911111573334083829048020")
	bw6_761.g2Gen.Y.SetString("2945005085389580383802706904000483833228424888054664780252599806365093320701303614818391222418768857269542753796449953578553937529004880983494788715529986360817835802796138196037201453469654110552028363169895102423753717534586247")
	bw6_761.g2Gen.Z.SetString("1")

	// Setting the loop counter for Miller loop in NAF form
	// TODO Optimized Miller loop described in the paper
	// https://eprint.iacr.org/2020/351.pdf (Algorithm 5)
	// for now just use the fr modulus as described in Section 3.3
	// (fr < trace of Frobenius so this is faster than trace of Frobenius)
	T, _ := new(big.Int).SetString("258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177", 10)
	utils.NafDecomposition(T, bw6_761.loopCounter[:])

	// infinity point G1
	bw6_761.g1Infinity.X.SetOne()
	bw6_761.g1Infinity.Y.SetOne()

	// infinity point G2
	bw6_761.g2Infinity.X.SetOne()
	bw6_761.g2Infinity.Y.SetOne()

	// precomputed values for ScalarMulByGen
	bw6_761.tGenG1[0].Set(&bw6_761.g1Gen)
	for j := 1; j < len(bw6_761.tGenG1)-1; j = j + 2 {
		bw6_761.tGenG1[j].Set(&bw6_761.tGenG1[j/2]).Double()
		bw6_761.tGenG1[j+1].Set(&bw6_761.tGenG1[(j+1)/2]).Add(&bw6_761, &bw6_761.tGenG1[j/2])
	}
	bw6_761.tGenG2[0].Set(&bw6_761.g2Gen)
	for j := 1; j < len(bw6_761.tGenG2)-1; j = j + 2 {
		bw6_761.tGenG2[j].Set(&bw6_761.tGenG2[j/2]).Double()
		bw6_761.tGenG2[j+1].Set(&bw6_761.tGenG2[(j+1)/2]).Add(&bw6_761, &bw6_761.tGenG2[j/2])
	}
}
