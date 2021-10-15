// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bw6633

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
)

// E: y**2=x**3+4
// Etwist: y**2 = x**3+8
// Tower: Fp->Fp6, u**6=2
// Generator (same as BLS24-315): x=-3218079743
// Fp: p=20494478644167774678813387386538961497669590920908778075528754551012016751717791778743535050360001387419576570244406805463255765034468441182772056330021723098661967429339971741066259394985997
// Fr: r=39705142709513438335025689890408969744933502416914749335064285505637884093126342347073617133569

// ID BW6_633 ID
const ID = ecc.BW6_633

// bCurveCoeff b coeff of the curve
var bCurveCoeff fp.Element

// bTwistCurveCoeff b coeff of the twist (defined over Fp) curve
var bTwistCurveCoeff fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

var g1GenAff G1Affine
var g2GenAff G2Affine

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counters
var loopCounter0 [159]int8
var loopCounter1 [159]int8

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

// generator of the curve
var xGen big.Int

func init() {

	bCurveCoeff.SetUint64(4)
	bTwistCurveCoeff.SetUint64(8) // M-twist

	// E1(2,y)*cofactor
	g1Gen.X.SetString("14087405796052437206213362229855313116771222912153372774869400386285407949123477431442535997951698710614498307938219633856996133201713506830167161540335446217605918678317160130862890417553415")
	g1Gen.Y.SetString("5208886161111258314476333487866604447704068601830026647530443033297117148121067806438008469463787158470000157308702133756065259580313172904438248825389121766442385979570644351664733475122746")
	g1Gen.Z.SetString("1")

	// E2(2,y))*cofactor
	g2Gen.X.SetString("13658793733252505713431834233072715040674666715141692574468286839081203251180283741830175712695426047062165811313478642863696265647598838732554425602399576125615559121457137320131899043374497")
	g2Gen.Y.SetString("599560264833409786573595720823495699033661029721475252751314180543773745554433461106678360045466656230822473390866244089461950086268801746497554519984580043036179195728559548424763890207250")
	g2Gen.Z.SetString("1")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	// binary decomposition of u+1 (negative)
	loopCounter0 = [159]int8{0, -1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, -1, 0, 0, 0, 0, 0, 0, 0, -1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// xGen^5-xGen^4-xGen (negative)
	T, _ := new(big.Int).SetString("345131030376204096837580131803633448876874137601", 10)
	ecc.NafDecomposition(T, loopCounter1[:])

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("4098895725012429242072311240482566844345873033931481129362557724405008256668293241245050359832461015092695507587185678086043587575438449040313411246717257958467499181450742260777082884928318") // (45-10*x+151*x^2-187*x^3+171*x^4-49*x^5-110*x^6+430*x^7-696*x^8+702*x^9-528*x^10+201*x^11+144*x^12-274*x^13+181*x^14-34*x^15-63*x^16+92*x^17-56*x^18+13*x^19)/15
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("39705142672498995661671850106945620852186608752525090699191017895721506694646055668218723303426", 10) // 1-x+2*x^2-2*x^3+3*x^5-4*x^6+4*x^7-3*x^8+x^9
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	xGen.SetString("3218079743", 10) // negative

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g2Jac G2Jac, g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	g1Jac = g1Gen
	g2Jac = g2Gen
	return
}
