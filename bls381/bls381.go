package bls381

import (
	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bls381/fp"
)

// E: y**2=x**3+4
// Etwist: y**2 = x**3+4*(u+1)
// Tower: Fp->Fp2, u**2=-1 -> Fp12, v**6=u+1
// Generator (BLS12 family): x=15132376222941642752
// Fp: p=4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787
// Fr: r=52435875175126190479447740508185965837690552500527637822603658699938581184513

// var bls381 Curve
// var initOnce sync.Once

// ID bls381 ID
var ID = gurvy.BLS381

// B b coeff of the curve
var B fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counter (=trace-1 = x in BLS family)
var loopCounter [64]int8

// parameters for pippenger ScalarMulByGen
// TODO get rid of this, keep only double and add, and the multi exp
const sGen = 4
const bGen = sGen

var tGenG1 [((1 << bGen) - 1)]G1Jac
var tGenG2 [((1 << bGen) - 1)]G2Jac

func init() {

	B.SetUint64(4)

	g1Gen.X.SetString("2407661716269791519325591009883849385849641130669941829988413640673772478386903154468379397813974815295049686961384")
	g1Gen.Y.SetString("821462058248938975967615814494474302717441302457255475448080663619194518120412959273482223614332657512049995916067")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("3914881020997020027725320596272602335133880006033342744016315347583472833929664105802124952724390025419912690116411",
		"277275454976865553761595788585036366131740173742845697399904006633521909118147462773311856983264184840438626176168")
	g2Gen.Y.SetString("253800087101532902362860387055050889666401414686580130872654083467859828854605749525591159464755920666929166876282",
		"1710145663789443622734372402738721070158916073226464929008132596760920130516982819361355832232719175024697380252309")
	g2Gen.Z.SetString("1",
		"0")

	// binary decomposition of 15132376222941642752 little endian
	loopCounter = [64]int8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1, 1}

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
