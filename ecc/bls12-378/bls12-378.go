package bls12378

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/internal/fptower"
)

// E: y**2=x**3+1
// Etwist: y**2 = x**3+u
// Tower: Fp->Fp2, u**2=-5 -> Fp12, v**6=u
// Generator (BLS12 family): x=11045256207009841153
// optimal Ate loop: trace(frob)-1=x
// trace of pi: x+1
// Fp: p=605248206075306171733248481581800960739847691770924913753520744034740935903401304776283802348837311170974282940417
// Fr: r=14883435066912132899950318861128167269793560281114003360875131245101026639873

// ID bls378 ID
const ID = ecc.BLS12_378

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

	bCurveCoeff.SetUint64(1)
	bTwistCurveCoeff.A1.SetUint64(1) // M-twist

	// E(3,y) * cofactor
	g1Gen.X.SetString("302027100877540500544138164010696035562809807233645104772290911818386302983750063098216015456036850656714568735197")
	g1Gen.Y.SetString("232851047397483214541821965369374725182070455016459237170823497053622811786333462699984177726412751508198874482530")
	g1Gen.Z.SetString("1")

	// E'(1,y) * cofactor'
	g2Gen.X.SetString("470810816643554779222760025249941413452299198622737082648784137654933833261310635469274149014014206108405592809732",
		"317092959336227428400228502739777439718827088477410533227996105067347670094088101088421556743730925535231685964487")
	g2Gen.Y.SetString("248853758964950314624408411876149087897475217517523838449839260719963153199419627931373025216041741725848318074460",
		"389162134924826972299508957175841717907876177152103852864177212390074067430801162403069988146334006672491106545644")
	g2Gen.Z.SetString("1",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("164391353554439166353793911729193406645071739502673898176639736370075683438438023898983435337729")
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("121997684678489422961514670190292369408", 10) //(x**2-1)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.A0.SetString("164391353554439166353793911729193406645071739502673898176639736370075683438438023898983435337730")
	endo.v.A0.SetString("595603361117066405543541008735167904222384847192046901135681663787023479658010166685728902742824780272831835669219")

	// binary decomposition of 11045256207009841153 little endian
	loopCounter = [64]int8{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1}

	xGen.SetString("11045256207009841153", 10)

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g2Jac G2Jac, g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	g1Jac = g1Gen
	g2Jac = g2Gen
	return
}
