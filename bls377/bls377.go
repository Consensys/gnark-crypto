package bls377

import (
	"math/big"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bls377/fp"
	"github.com/consensys/gurvy/bls377/fr"
	"github.com/consensys/gurvy/bls377/internal/fptower"
	"github.com/consensys/gurvy/utils"
)

// E: y**2=x**3+1
// Etwist: y**2 = x**3+u**-1
// Tower: Fp->Fp2, u**2=5 -> Fp12, v**6=u
// Generator (BLS12 family): x=9586122913090633729
// optimal Ate loop: trace(frob)-1=x
// trace of pi: x+1
// Fp: p=258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177
// Fr: r=8444461749428370424248824938781546531375899335154063827935233455917409239041 (x**4-x**2+1)

// ID bls377 ID
const ID = gurvy.BLS377

// bCurveCoeff b coeff of the curve
var bCurveCoeff fp.Element

// bTwistCurveCoeff b coeff of the twist (defined over Fp2) curve
var bTwistCurveCoeff fptower.E2

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

var g1GenAff G1Affine
var g2GenAff G2Affine

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counter (=trace-1 = x in BLS family)
var loopCounter [64]int8

// Parameters useful for the GLV scalar multiplication. The third roots define the
//  endomorphisms phi1 and phi2 for <G1Affine> and <G2Affine>. lambda is such that <r, phi-lambda> lies above
// <r> in the ring Z[phi]. More concretely it's the associated eigenvalue
// of phi1 (resp phi2) restricted to <G1Affine> (resp <G2Affine>)
// cf https://www.cosic.esat.kuleuven.be/nessie/reports/phase2/GLV.pdf
var thirdRootOneG1 fp.Element
var thirdRootOneG2 fp.Element
var lambdaGLV big.Int

// glvBasis stores R-linearly independant vectors (a,b), (c,d)
// in ker((u,v)->u+vlambda[r]), and their determinant
var glvBasis utils.Lattice

// psi o pi o psi**-1, where psi:E->E' is the degree 6 iso defined over Fp12
var endo struct {
	u fptower.E2
	v fptower.E2
}

// generator of the curve
var xGen big.Int

// expose the tower -- github.com/consensys/gnark uses it in a gnark circuit

// E2 is a degree two finite field extension of fp.Element
type E2 = fptower.E2

// E6 is a degree three finite field extension of fp2
type E6 = fptower.E6

// E12 is a degree two finite field extension of fp6
type E12 = fptower.E12

func init() {

	bCurveCoeff.SetUint64(1)
	bTwistCurveCoeff.A1.SetUint64(1)
	bTwistCurveCoeff.Inverse(&bTwistCurveCoeff)

	g1Gen.X.SetString("68333130937826953018162399284085925021577172705782285525244777453303237942212457240213897533859360921141590695983")
	g1Gen.Y.SetString("243386584320553125968203959498080829207604143167922579970841210259134422887279629198736754149500839244552761526603")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("129200027147742761118726589615458929865665635908074731940673005072449785691019374448547048953080140429883331266310",
		"218164455698855406745723400799886985937129266327098023241324696183914328661520330195732120783615155502387891913936")
	g2Gen.Y.SetString("178797786102020318006939402153521323286173305074858025240458924050651930669327663166574060567346617543016897467207",
		"246194676937700783734853490842104812127151341609821057456393698060154678349106147660301543343243364716364400889778")
	g2Gen.Z.SetString("1",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("80949648264912719408558363140637477264845294720710499478137287262712535938301461879813459410945")
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("91893752504881257701523279626832445440", 10) //(x**2-1)
	_r := fr.Modulus()
	utils.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.A0.SetString("80949648264912719408558363140637477264845294720710499478137287262712535938301461879813459410946")
	endo.v.A0.SetString("216465761340224619389371505802605247630151569547285782856803747159100223055385581585702401816380679166954762214499")

	// binary decomposition of 15132376222941642752 little endian
	loopCounter = [64]int8{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1}

	xGen.SetString("9586122913090633729", 10)

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	return
}
