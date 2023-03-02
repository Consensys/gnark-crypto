// Package bw6756 efficient elliptic curve, pairing and hash to curve implementation for bw6-756.
//
// bw6-756: A Brezing--Weng curve (2-chain with bls12-378)
//
//	embedding degree k=6
//	seed x₀=11045256207009841153.
//	𝔽p: p=366325390957376286590726555727219947825377821289246188278797409783441745356050456327989347160777465284190855125642086860525706497928518803244008749360363712553766506755227344593404398783886857865261088226271336335268413437902849
//	𝔽r: r=605248206075306171733248481581800960739847691770924913753520744034740935903401304776283802348837311170974282940417
//	(E/𝔽p): Y²=X³+1
//	(Eₜ/𝔽p): Y² = X³+33 (M-type twist)
//	r ∣ #E(Fp) and r ∣ #Eₜ(𝔽p)
//
// Extension fields tower:
//
//	𝔽p³[u] = 𝔽p/u³-33
//	𝔽p⁶[v] = 𝔽p²/v²-u
//
// optimal Ate loops:
//
//	x₀+1, x₀²-x₀-1
//
// Security: estimated 126-bit level following [https://eprint.iacr.org/2019/885.pdf]
// (r is 378 bits and p⁶ is 4536 bits)
//
// # Warning
//
// This code has not been audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package bw6756

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
)

// ID BW6_756 ID
const ID = ecc.BW6_756

// aCurveCoeff is the a coefficients of the curve Y²=X³+ax+b
var aCurveCoeff fp.Element

// bCurveCoeff b coeff of the curve Y²=X³+b
var bCurveCoeff fp.Element

// bTwistCurveCoeff b coeff of the twist (defined over 𝔽p) curve
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
// endomorphisms ϕ₁ and ϕ₂ for <G1Affine> and <G2Affine>. lambda is such that <r, ϕ-λ> lies above
// <r> in the ring Z[ϕ]. More concretely it's the associated eigenvalue
// of ϕ₁ (resp ϕ₂) restricted to <G1Affine> (resp <G2Affine>)
// see https://www.cosic.esat.kuleuven.be/nessie/reports/phase2/GLV.pdf
var thirdRootOneG1 fp.Element
var thirdRootOneG2 fp.Element
var lambdaGLV big.Int

// glvBasis stores R-linearly independent vectors (a,b), (c,d)
// in ker((u,v) → u+vλ[r]), and their determinant
var glvBasis ecc.Lattice

// generator of the curve
var xGen big.Int

func init() {
	aCurveCoeff.SetUint64(0)
	bCurveCoeff.SetOne()
	bTwistCurveCoeff.MulByNonResidue(&bCurveCoeff)

	// E(3,y) * cofactor
	g1Gen.X.SetString("286035407532233812057489253822435660910062665263942803649298092690795938518721117964189338863504082781482751182899097859005716378386344565362972291164604792882058761734674709131229927253172681714645554597102571818586966737895501")
	g1Gen.Y.SetString("250540671634276190125882738767359258920233951524378923555904955920886135268516617166458911260101792169356480449980342047600821278990712908224386045486820019065641642853528653616206514851361917670279865872746658429844440125628329")
	g1Gen.Z.SetOne()

	// E(1,y) * cofactor
	g2Gen.X.SetString("270164867145533700243149075881223225204067215320977230235816769808318087164726583740674261721395147407122688542569094772405350936550575160051166652281373572919753182191250641388443572739372443497834910784618354592418817138212395")
	g2Gen.Y.SetString("296695446824796322573519291690935001172593568823998954880196613542512471119971074118215403545906873458039024520146929054366200365532511334310660691775675887531695313103875249166779149013653038059140912965769351316868363001510735")
	g2Gen.Z.SetOne()

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	// x₀+1
	loopCounter0 = [191]int8{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 1, 0, 0, -1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	// x₀³-x₀²-x₀
	T, _ := new(big.Int).SetString("1347495683935914696108087318582641220368021451587784278015", 10)
	ecc.NafDecomposition(T, loopCounter1[:])

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG2.SetString("99497571833115712246976573293861816254377473715694998268521440373748988342600853091641405554217584221455319677515385376103078837731420131015700054219263015095146628991433981753068027965212839748934246550470657")
	thirdRootOneG1.Square(&thirdRootOneG2)
	lambdaGLV.SetString("164391353554439166353793911729193406645071739502673898176639736370075683438438023898983435337729", 10) // (x⁵-3x⁴+3x³-x+1)
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

// CurveCoefficients returns the a, b coefficients of the curve equation.
func CurveCoefficients() (a, b fp.Element) {
	return aCurveCoeff, bCurveCoeff
}
