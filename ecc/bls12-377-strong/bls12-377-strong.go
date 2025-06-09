// Package bls12377strong efficient elliptic curve, pairing and hash to curve implementation for bls12-377-strong.
//
// bls12-377-strong: A batch-SMT-friendly G2-strong and Gt-strong Barreto--Lynn--Scott curve with
//
//	embedding degree k=12
//	seed x₀=-0x816163471f000001
//	𝔽r: r=0x10b392a66821715b61cc16222373e51032deb308a6575eae0f87c68e3e000001 (x₀⁴-x₀²+1)
//	𝔽p: p=0x16c06316cabdde5b00fdb45ba6a5ba3484e34c52377be4886858d90a5426e595e6f3b3fd8f94772366dd48007aaaaab ((x₀-1)² ⋅ r(x₀)/3+x₀)
//	(E/𝔽p): Y²=X³+1
//	(Et/𝔽p²): Y² = X³+1/(u+1) (D-type twist)
//	r ∣ #E(Fp) and r ∣ #Et(𝔽p²)
//
// Extension fields tower:
//
//	𝔽p²[u] = 𝔽p/u²+1
//	𝔽p⁶[v] = 𝔽p²/v³-u-1
//	𝔽p¹²[w] = 𝔽p⁶/w²-v
//
// optimal Ate loop size:
//
//	x₀
//
// Security: estimated 126-bit level following [https://eprint.iacr.org/2019/885.pdf]
// (r is 253 bits and p¹² is 4517 bits)
//
// # Warning
//
// This code has not been audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package bls12377strong

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-377-strong/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377-strong/internal/fptower"
)

// ID bls377strong ID
const ID = ecc.BLS12_377_STRONG

// aCurveCoeff is the a coefficients of the curve Y²=X³+ax+b
var aCurveCoeff fp.Element
var bCurveCoeff fp.Element

// twist
var twist fptower.E2

// bTwistCurveCoeff b coeff of the twist (defined over 𝔽p²) curve
var bTwistCurveCoeff fptower.E2

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

var g1GenAff G1Affine
var g2GenAff G2Affine

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counter
var LoopCounter [64]int8

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

// ψ o π o ψ⁻¹, where ψ:E → E' is the degree 6 iso defined over 𝔽p¹²
var endo struct {
	u fptower.E2
	v fptower.E2
}

// seed x₀ of the curve
var xGen big.Int

// expose the tower -- github.com/consensys/gnark uses it in a gnark circuit

// 𝔽p²
type E2 = fptower.E2

// 𝔽p⁶
type E6 = fptower.E6

// 𝔽p¹²
type E12 = fptower.E12

func init() {
	aCurveCoeff.SetUint64(0)
	bCurveCoeff.SetUint64(1)

	// D-twist
	twist.A0.SetUint64(1)
	twist.A1.SetUint64(1)
	bTwistCurveCoeff.Inverse(&twist)

	g1Gen.X.SetString("194723246994625592748501204359320585322923083844478683949332596811528155666341934162568693156285023902900452305591")
	g1Gen.Y.SetString("77322027808267321961126786225422958251306715431373304983097765527554768217263151021098250419837783210459916891736")
	g1Gen.Z.SetOne()

	g2Gen.X.SetString("139538753679008269700608213550751017855790082481593304207133459071771768290323119353705107467678720859393494004443",
		"212606921292338572193784645699821078543480875538489502757941879749041236013855214569323373047433596997860692987551")
	g2Gen.Y.SetString("165045341471788782528469833498467433128175966388645554526946736931572906635582732330301458751633758306432610141660",
		"27925942218996312396346158838248654495124072346833716936439736614332572876990879292697744908026038527610236485443")
	g2Gen.Z.SetString("1",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("218861136708759775118463921633794847536991885687584169137356657270576592308864618581121708045962261427481020639910") // x₀^5-3x₀^4+3x₀^3-x₀+1
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("86915380360507006654423840688405741568", 10) //(x₀²-1)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.A1.SetString("70427388980041189209419192134717659880491069728330620094944516061531890781668663842312340635652")
	endo.v.A0.SetString("144394925554309751978230035908754812058838028871212916374968753301545726589628552395959492401580323225873322644219")
	endo.v.A1.SetString("144394925554309751978230035908754812058838028871212916374968753301545726589628552395959492401580323225873322644219")

	// -x₀
	xGen.SetString("9322841860747558913", 10)

	// 2-NAF decomposition of -x₀ little endian
	ecc.NafDecomposition(&xGen, LoopCounter[:])
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
