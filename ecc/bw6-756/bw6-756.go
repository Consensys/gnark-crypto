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

package bw6756

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
)

// E: y**2=x**3+1
// Etwist: y**2 = x**3+33
// Tower: Fp->Fp6, u**6=33
// Generator (same as BLS378): x=11045256207009841153
// optimal Ate loops: x+1, x**2-x-1
// Fp: p=366325390957376286590726555727219947825377821289246188278797409783441745356050456327989347160777465284190855125642086860525706497928518803244008749360363712553766506755227344593404398783886857865261088226271336335268413437902849
// Fr: r=605248206075306171733248481581800960739847691770924913753520744034740935903401304776283802348837311170974282940417

// ID BW6_756 ID
const ID = ecc.BW6_756

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
var loopCounter0 [191]int8
var loopCounter1 [191]int8

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

	bCurveCoeff.SetOne()
	bTwistCurveCoeff.MulByNonResidue(&bCurveCoeff)

	// E(3,y) * cofactor
	g1Gen.X.SetString("286035407532233812057489253822435660910062665263942803649298092690795938518721117964189338863504082781482751182899097859005716378386344565362972291164604792882058761734674709131229927253172681714645554597102571818586966737895501")
	g1Gen.Y.SetString("250540671634276190125882738767359258920233951524378923555904955920886135268516617166458911260101792169356480449980342047600821278990712908224386045486820019065641642853528653616206514851361917670279865872746658429844440125628329")
	g1Gen.Z.SetString("1")

	// E(1,y) * cofactor
	g2Gen.X.SetString("270164867145533700243149075881223225204067215320977230235816769808318087164726583740674261721395147407122688542569094772405350936550575160051166652281373572919753182191250641388443572739372443497834910784618354592418817138212395")
	g2Gen.Y.SetString("296695446824796322573519291690935001172593568823998954880196613542512471119971074118215403545906873458039024520146929054366200365532511334310660691775675887531695313103875249166779149013653038059140912965769351316868363001510735")
	g2Gen.Z.SetString("1")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	// xGen+1
	loopCounter0 = [191]int8{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, -1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// xGen^3-xGen^2-xGen
	T, _ := new(big.Int).SetString("1347495683935914696108087318582641220368021451587784278015", 10)
	ecc.NafDecomposition(T, loopCounter1[:])

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG2.SetString("99497571833115712246976573293861816254377473715694998268521440373748988342600853091641405554217584221455319677515385376103078837731420131015700054219263015095146628991433981753068027965212839748934246550470657")
	thirdRootOneG1.Square(&thirdRootOneG2)
	lambdaGLV.SetString("164391353554439166353793911729193406645071739502673898176639736370075683438438023898983435337729", 10) // (x**5-3*x**4+3*x**3-x+1)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

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
