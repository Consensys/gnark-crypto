package bls1239

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-39/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-39/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-39/internal/fptower"
)

// E: y**2=x**3+2
// Etwist: y**2 = x**3+2/(u+1)
// Tower: Fp->Fp2, u**2=3 -> Fp12, v**6=u+1
// Generator (BLS12 family): x=100
// optimal Ate loop: trace(frob)-1=x
// trace of pi: x+1
// Fp: p=326667333367
// Fr: r=99990001
// ID bls39 ID
const ID = ecc.BLS12_39

// bCurveCoeff b coeff of the curve
var bCurveCoeff fp.Element

// twist
var twist fptower.E2

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
var loopCounter [7]int8

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
var glvBasis ecc.Lattice

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

	bCurveCoeff.SetUint64(2)
	twist.A0.SetUint64(1)
	twist.A1.SetUint64(1)
	bTwistCurveCoeff.Inverse(&twist).Double(&bTwistCurveCoeff)

	g1Gen.X.SetString("76374581475")
	g1Gen.Y.SetString("135768504117")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("170522782386",
		"184493119176")
	g2Gen.Y.SetString("113781902987",
		"323607052549")
	g2Gen.Z.SetString("1",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("9702999901")
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("9999", 10)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.A0.SetString("159768345029")
	endo.u.A1.SetString("182009477101")
	endo.v.A0.SetString("293515655025")
	endo.v.A1.SetString("228828781692")

	// binary decomposition of 100 little endian
	loopCounter = [7]int8{0, 0, 1, 0, 0, 1, 1}

	xGen.SetString("100", 10)

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g2Jac G2Jac, g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	g1Jac = g1Gen
	g2Jac = g2Gen
	return
}
