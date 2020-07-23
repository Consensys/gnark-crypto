package bls381

import (
	"sync"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bls381/fp"
)

// TODO go:generate go run ../internal/generator.go -out . -package bls381 -t 15132376222941642752 -tNeg -p 4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787 -r 52435875175126190479447740508185965837690552500527637822603658699938581184513 -fp2 -1 -fp6 1,1

// E: y**2=x**3+4
// Etwist: y**2 = x**3+4*(u+1)

var bls381 Curve
var initOnce sync.Once

// ID bls381 ID
const ID = gurvy.BLS381

// parameters for pippenger ScalarMulByGen
const sGen = 4
const bGen = sGen

// PairingResult target group of the pairing
type PairingResult = E12

// BLS381 returns BLS381 curve
func BLS381() *Curve {
	initOnce.Do(initBLS381)
	return &bls381
}

// Curve represents the BLS381 curve and pre-computed constants
type Curve struct {

	// A, B coefficients of the curve x^3 = y^2 +AX+b
	B fp.Element

	// generators of the r-torsion subgroup, g1 in ker(pi), g2 in ker(p-q)
	g1Gen      G1Jac
	g2Gen      G2Jac
	g1Infinity G1Jac
	g2Infinity G2Jac

	// NAF decomposition takes 65 trits for bls381 but only 64 trits for bls377
	loopCounter [64]int8

	// precomputed values for ScalarMulByGen
	tGenG1 [((1 << bGen) - 1)]G1Jac
	tGenG2 [((1 << bGen) - 1)]G2Jac
}

func initBLS381() {

	// A, B coeffs of the curve in Mont form
	bls381.B.SetUint64(4)

	// Setting G1Jac
	bls381.g1Gen.X.SetString("2407661716269791519325591009883849385849641130669941829988413640673772478386903154468379397813974815295049686961384")
	bls381.g1Gen.Y.SetString("821462058248938975967615814494474302717441302457255475448080663619194518120412959273482223614332657512049995916067")
	bls381.g1Gen.Z.SetString("1")

	// Setting G2Jac
	bls381.g2Gen.X.SetString("3914881020997020027725320596272602335133880006033342744016315347583472833929664105802124952724390025419912690116411",
		"277275454976865553761595788585036366131740173742845697399904006633521909118147462773311856983264184840438626176168")
	bls381.g2Gen.Y.SetString("253800087101532902362860387055050889666401414686580130872654083467859828854605749525591159464755920666929166876282",
		"1710145663789443622734372402738721070158916073226464929008132596760920130516982819361355832232719175024697380252309")
	bls381.g2Gen.Z.SetString("1",
		"0")

	// 15132376222941642752 (trace of the Frobenius, it's the shortest Miller loop for BLS family)
	bls381.loopCounter = [64]int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 1}

	// infinity point G1
	bls381.g1Infinity.X.SetOne()
	bls381.g1Infinity.Y.SetOne()

	// infinity point G2
	bls381.g2Infinity.X.SetOne()
	bls381.g2Infinity.Y.SetOne()

	// precomputed values for ScalarMulByGen
	bls381.tGenG1[0].Set(&bls381.g1Gen)
	for j := 1; j < len(bls381.tGenG1)-1; j = j + 2 {
		bls381.tGenG1[j].Set(&bls381.tGenG1[j/2]).DoubleAssign()
		bls381.tGenG1[j+1].Set(&bls381.tGenG1[(j+1)/2]).AddAssign(&bls381, &bls381.tGenG1[j/2])
	}
	bls381.tGenG2[0].Set(&bls381.g2Gen)
	for j := 1; j < len(bls381.tGenG2)-1; j = j + 2 {
		bls381.tGenG2[j].Set(&bls381.tGenG2[j/2]).DoubleAssign()
		bls381.tGenG2[j+1].Set(&bls381.tGenG2[(j+1)/2]).AddAssign(&bls381, &bls381.tGenG2[j/2])
	}
}
