package bw761

import (
	"math/big"
	"sync"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bw761/fp"
	"github.com/consensys/gurvy/utils"
)

// E: y**2=x**3-1
// Etwist: y**2 = x**3+4
// field ext modulus: x**6+4

var bw761 Curve
var initOnce sync.Once

const ID = gurvy.BW761

// parameters for pippenger ScalarMulByGen
const sGen = 4
const bGen = sGen

type PairingResult = E6

// BW761 returns BW761 curve
func BW761() *Curve {
	initOnce.Do(initBW761)
	return &bw761
}

// Curve represents the BW761 curve and pre-computed constants
type Curve struct {
	B fp.Element // A, B coefficients of the curve x^3 = y^2 +AX+b

	g1Gen G1Jac // generator of torsion group G1Jac
	g2Gen G2Jac // generator of torsion group G2Jac

	g1Infinity G1Jac // infinity (in Jacobian coords)
	g2Infinity G2Jac

	// Miller loop counters in NAF form
	// TODO For the love of god, please clean this up
	loopCounter1 [64]int8
	loopCounter2 [127]int8

	// precomputed values for ScalarMulByGen
	tGenG1 [((1 << bGen) - 1)]G1Jac
	tGenG2 [((1 << bGen) - 1)]G2Jac
}

func initBW761() {

	// A, B coeffs of the curve in Mont form
	bw761.B.SetUint64(4)

	// Setting G1Jac
	bw761.g1Gen.X.SetString("5492337019202608651620810666633622531924946248948182754748114963334556774714407693672822645637243083342924475378144397780999025266189779523629084326871556483802038026432771927197170911996417793635501066231650458516636932478125208")
	bw761.g1Gen.Y.SetString("4874298780810344118673004453041997030286302865034758641338313952140849332867290574388366379298818956144982860224857872858166812124104845663394852158352478303048122861831479086904887356602146134586313962565783961814162269209043907")
	bw761.g1Gen.Z.SetString("1")

	// Setting G2Jac
	bw761.g2Gen.X.SetString("5779457169892140542970811884673908634889239063901429247094594197042136765689827803062459420720318762253427359282239252479201196985966853806926626938528693270647807548111019296972244105103687281416386903420911111573334083829048020")
	bw761.g2Gen.Y.SetString("2945005085389580383802706904000483833228424888054664780252599806365093320701303614818391222418768857269542753796449953578553937529004880983494788715529986360817835802796138196037201453469654110552028363169895102423753717534586247")
	bw761.g2Gen.Z.SetString("1")

	// Setting the loop counters for Miller loop in NAF form
	// https://eprint.iacr.org/2020/351.pdf (Algorithm 5)
	T, _ := new(big.Int).SetString("9586122913090633729", 10)
	utils.NafDecomposition(T, bw761.loopCounter1[:])
	T2, _ := new(big.Int).SetString("91893752504881257691937156713741811711", 10)
	utils.NafDecomposition(T2, bw761.loopCounter2[:])

	// infinity point G1
	bw761.g1Infinity.X.SetOne()
	bw761.g1Infinity.Y.SetOne()

	// infinity point G2
	bw761.g2Infinity.X.SetOne()
	bw761.g2Infinity.Y.SetOne()

	// precomputed values for ScalarMulByGen
	bw761.tGenG1[0].Set(&bw761.g1Gen)
	for j := 1; j < len(bw761.tGenG1)-1; j = j + 2 {
		bw761.tGenG1[j].Set(&bw761.tGenG1[j/2]).DoubleAssign()
		bw761.tGenG1[j+1].Set(&bw761.tGenG1[(j+1)/2]).AddAssign(&bw761, &bw761.tGenG1[j/2])
	}
	bw761.tGenG2[0].Set(&bw761.g2Gen)
	for j := 1; j < len(bw761.tGenG2)-1; j = j + 2 {
		bw761.tGenG2[j].Set(&bw761.tGenG2[j/2]).DoubleAssign()
		bw761.tGenG2[j+1].Set(&bw761.tGenG2[(j+1)/2]).AddAssign(&bw761, &bw761.tGenG2[j/2])
	}
}
