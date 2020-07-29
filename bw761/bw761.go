package bw761

import (
	"math/big"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bw761/fp"
	"github.com/consensys/gurvy/utils"
)

// https://eprint.iacr.org/2020/351.pdf

// E: y**2=x**3-1
// Etwist: y**2 = x**3+4
// Tower: Fp->Fp6, u**6=-4
// Generator (same as BLS377): x=9586122913090633729
// optimal Ate loops: x+1, x**2-x-1
// Fp: p=6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299
// Fr: r=258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177

// ID bls377 ID
var ID = gurvy.BW761

// B b coeff of the curve
var B fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counters
// Miller loop 1: f(P), div(f) = (x+1)(Q)-([x+1]Q)-x(O)
// Miller loop 2: f(P), div(f) = (x**3-x**2-x)(Q) -([x**3-x**2-x]Q)-(x**3-x**2-x-1)(O)
var loopCounter1 [64]int8
var loopCounter2 [127]int8

// parameters for pippenger ScalarMulByGen
// TODO get rid of this, keep only double and add, and the multi exp
const sGen = 4
const bGen = sGen

var tGenG1 [((1 << bGen) - 1)]G1Jac
var tGenG2 [((1 << bGen) - 1)]G2Jac

func init() {

	B.SetOne().Neg(&B)

	g1Gen.X.SetString("5492337019202608651620810666633622531924946248948182754748114963334556774714407693672822645637243083342924475378144397780999025266189779523629084326871556483802038026432771927197170911996417793635501066231650458516636932478125208")
	g1Gen.Y.SetString("4874298780810344118673004453041997030286302865034758641338313952140849332867290574388366379298818956144982860224857872858166812124104845663394852158352478303048122861831479086904887356602146134586313962565783961814162269209043907")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("5779457169892140542970811884673908634889239063901429247094594197042136765689827803062459420720318762253427359282239252479201196985966853806926626938528693270647807548111019296972244105103687281416386903420911111573334083829048020")
	g2Gen.Y.SetString("2945005085389580383802706904000483833228424888054664780252599806365093320701303614818391222418768857269542753796449953578553937529004880983494788715529986360817835802796138196037201453469654110552028363169895102423753717534586247")
	g2Gen.Z.SetString("1")

	//binary decomposition of 9586122913090633729, little endian
	loopCounter1 = [64]int8{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1}

	T, _ := new(big.Int).SetString("91893752504881257691937156713741811711", 10)
	utils.NafDecomposition(T, loopCounter2[:])
	// fmt.Print("[")
	// for _, v := range loopCounter2 {
	// 	fmt.Printf("%d,", v)
	// }
	// fmt.Print("]\n")

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	tGenG1[0].Set(&g1Gen)
	for j := 1; j < len(tGenG1)-1; j = j + 2 {
		tGenG1[j].Set(&tGenG1[j/2]).DoubleAssign()
		tGenG1[j+1].Set(&tGenG1[(j+1)/2]).AddAssign(&tGenG1[j/2])
	}
	tGenG2[0].Set(&g2Gen)
	for j := 1; j < len(tGenG2)-1; j = j + 2 {
		tGenG2[j].Set(&tGenG2[j/2]).DoubleAssign()
		tGenG2[j+1].Set(&tGenG2[(j+1)/2]).AddAssign(&tGenG2[j/2])
	}
}

// // E: y**2=x**3-1
// // Etwist: y**2 = x**3+4
// // field ext modulus: x**6+4

// var bw761 Curve
// var initOnce sync.Once

// const ID = gurvy.BW761

// // parameters for pippenger ScalarMulByGen
// const sGen = 4
// const bGen = sGen

// type PairingResult = E6

// // BW761 returns BW761 curve
// func BW761() *Curve {
// 	initOnce.Do(initBW761)
// 	return &bw761
// }

// // Curve represents the BW761 curve and pre-computed constants
// type Curve struct {
// 	B fp.Element // A, B coefficients of the curve x^3 = y^2 +AX+b

// 	g1Gen G1Jac // generator of torsion group G1Jac
// 	g2Gen G2Jac // generator of torsion group G2Jac

// 	g1Infinity G1Jac // infinity (in Jacobian coords)
// 	g2Infinity G2Jac

// 	// Miller loop counters in NAF form
// 	// TODO For the love of god, please clean this up
// 	loopCounter1 [64]int8
// 	loopCounter2 [127]int8

// 	// precomputed values for ScalarMulByGen
// 	tGenG1 [((1 << bGen) - 1)]G1Jac
// 	tGenG2 [((1 << bGen) - 1)]G2Jac
// }

// func initBW761() {

// 	// A, B coeffs of the curve in Mont form
// 	bw761.B.SetUint64(4)

// 	// Setting G1Jac
// 	bw761.g1Gen.X.SetString("5492337019202608651620810666633622531924946248948182754748114963334556774714407693672822645637243083342924475378144397780999025266189779523629084326871556483802038026432771927197170911996417793635501066231650458516636932478125208")
// 	bw761.g1Gen.Y.SetString("4874298780810344118673004453041997030286302865034758641338313952140849332867290574388366379298818956144982860224857872858166812124104845663394852158352478303048122861831479086904887356602146134586313962565783961814162269209043907")
// 	bw761.g1Gen.Z.SetString("1")

// 	// Setting G2Jac
// 	bw761.g2Gen.X.SetString("5779457169892140542970811884673908634889239063901429247094594197042136765689827803062459420720318762253427359282239252479201196985966853806926626938528693270647807548111019296972244105103687281416386903420911111573334083829048020")
// 	bw761.g2Gen.Y.SetString("2945005085389580383802706904000483833228424888054664780252599806365093320701303614818391222418768857269542753796449953578553937529004880983494788715529986360817835802796138196037201453469654110552028363169895102423753717534586247")
// 	bw761.g2Gen.Z.SetString("1")

// 	// Setting the loop counters for Miller loop in NAF form
// 	// https://eprint.iacr.org/2020/351.pdf (Algorithm 5)
// 	T, _ := new(big.Int).SetString("9586122913090633729", 10)
// 	utils.NafDecomposition(T, bw761.loopCounter1[:])

// 	T2, _ := new(big.Int).SetString("91893752504881257691937156713741811711", 10)
// 	utils.NafDecomposition(T2, bw761.loopCounter2[:])

// 	// infinity point G1
// 	bw761.g1Infinity.X.SetOne()
// 	bw761.g1Infinity.Y.SetOne()

// 	// infinity point G2
// 	bw761.g2Infinity.X.SetOne()
// 	bw761.g2Infinity.Y.SetOne()

// 	// precomputed values for ScalarMulByGen
// 	bw761.tGenG1[0].Set(&bw761.g1Gen)
// 	for j := 1; j < len(bw761.tGenG1)-1; j = j + 2 {
// 		bw761.tGenG1[j].Set(&bw761.tGenG1[j/2]).DoubleAssign()
// 		bw761.tGenG1[j+1].Set(&bw761.tGenG1[(j+1)/2]).AddAssign(&bw761, &bw761.tGenG1[j/2])
// 	}
// 	bw761.tGenG2[0].Set(&bw761.g2Gen)
// 	for j := 1; j < len(bw761.tGenG2)-1; j = j + 2 {
// 		bw761.tGenG2[j].Set(&bw761.tGenG2[j/2]).DoubleAssign()
// 		bw761.tGenG2[j+1].Set(&bw761.tGenG2[(j+1)/2]).AddAssign(&bw761, &bw761.tGenG2[j/2])
// 	}
// }
