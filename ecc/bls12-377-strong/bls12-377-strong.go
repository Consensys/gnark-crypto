// Package bls12377strong efficient elliptic curve, pairing and hash to curve implementation for bls12-377-strong.
//
// bls12-377-strong: A batch SMT-friendly G2- and Gt- strong Barreto--Lynn--Scott curve with
//
//	embedding degree k=12
//	seed x₀=0x8000000000001163
//	𝔽r: r=0x10000000000008b18000000001c5726d400000290ff0a113000164f811c92089 (x₀⁴-x₀²+1)
//	𝔽p: p=0x15555555555566b80000000005e74378555556670530ac9c001be13dfc2115482e6dce9a57ae2a19cd84202412da3ef ((x₀-1)² ⋅ r(x₀)/3+x₀)
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
	twist.A0.SetUint64(2)
	twist.A1.SetUint64(1)
	bTwistCurveCoeff.Inverse(&twist)

	// cofactor * E(1,y)
	g1Gen.X.SetString("119032536028858284813403985676422819177240772841936675489113680890095013633468443632425375131153391389909041648542")
	g1Gen.Y.SetString("14862348291137921244379417782305839502634586502229003834413058582981182560007046793890137272115570289170871870739")
	g1Gen.Z.SetOne()

	g2Gen.X.SetString("8660812305207864393786249665246289109629756126921984067125112455703885598458019670541094414862926629318245834440",
		"154914380712861925719861234784216244827365812881695855125911815059330365567862755626026689722957051461992272065777")
	g2Gen.Y.SetString("80817493613498193224839977547396778116805799629774740679425781872036787703947037111844861876477342223808230950121",
		"45123149462580565523243805369980317722834844743326218962015350984834197151176442171800908962622986501611728517375")
	g2Gen.Z.SetString("1",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("66749594872528601112692535115452694730463020851273681873215020777094334903430823628450258725295") // x₀^5-3x₀^4+3x₀^3-x₀+1
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("85070591730234697972301523939176107080", 10) //(x₀²-1)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.A0.SetString("88899026867714575188751882972325772064086217814419839293293529345724897417332022887401511902040845984717680389861")
	endo.u.A1.SetString("176618080069251975695932762281816448616069709450241741294490249889740395629078583406915317197028257142967053061951")
	endo.v.A0.SetString("137764902161523724326024409884254444150371201229781842718963172888657649467106698224478747992403404446630035153285")
	endo.v.A1.SetString("70311022050158941927701660579722878126128079702520884352918338625171444218636495686162913580612066007627520089883")

	// binary decomposition of x₀ little endian
	LoopCounter = [64]int8{1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}

	// x₀
	xGen.SetString("9223372036854780259", 10)

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
