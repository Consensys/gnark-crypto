package bls377

import (
	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bls377/fp"
)

// E: y**2=x**3+1
// Etwist: y**2 = x**3+u**-1
// Tower: Fp->Fp2, u**2=5 -> Fp12, v**6=u
// Generator (BLS12 family): x=9586122913090633729
// Fp: p=258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177
// Fr: r=8444461749428370424248824938781546531375899335154063827935233455917409239041

// ID bls377 ID
var ID = gurvy.BLS377

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

	B.SetUint64(1)

	g1Gen.X.SetString("68333130937826953018162399284085925021577172705782285525244777453303237942212457240213897533859360921141590695983")
	g1Gen.Y.SetString("243386584320553125968203959498080829207604143167922579970841210259134422887279629198736754149500839244552761526603")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("129200027147742761118726589615458929865665635908074731940673005072449785691019374448547048953080140429883331266310",
		"218164455698855406745723400799886985937129266327098023241324696183914328661520330195732120783615155502387891913936")
	g2Gen.Y.SetString("178797786102020318006939402153521323286173305074858025240458924050651930669327663166574060567346617543016897467207",
		"246194676937700783734853490842104812127151341609821057456393698060154678349106147660301543343243364716364400889778")
	g2Gen.Z.SetString("1",
		"0")

	// binary decomposition of 15132376222941642752 little endian
	loopCounter = [64]int8{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1}

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
