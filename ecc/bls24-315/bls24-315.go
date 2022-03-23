package bls24315

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/internal/fptower"
)

// E: y**2=x**3+1
// Etwist: y**2 = x**3+v**-1
// Tower: Fp->Fp2, u**2=13 -> Fp4, v**2=u -> Fp8, w**2=v -> Fp24, i**3=w
// Generator (BLS24 family): x=-0xbfcfffff (32 bits)
// optimal Ate loop: trace(frob)-1=x
// trace of pi: x+1
// Fp: p=0x4c23a02b586d650d3f7498be97c5eafdec1d01aa27a1ae0421ee5da52bde5026fe802ff40300001 (315 bits)
// Fr: r=0x196deac24a9da12b25fc7ec9cf927a98c8c480ece644e36419d0c5fd00c00001 (253 bits)

// ID bls315 ID
const ID = ecc.BLS24_315

// bCurveCoeff b coeff of the curve
var bCurveCoeff fp.Element

// twist
var twist fptower.E4

// bTwistCurveCoeff b coeff of the twist (defined over Fp4) curve
var bTwistCurveCoeff fptower.E4

// twoInv 1/2 mod p (needed for DoubleStep in Miller loop)
var twoInv fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

var g1GenAff G1Affine
var g2GenAff G2Affine

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counter (=trace-1 = x in BLS24 family)
var loopCounter [33]int8

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

// psi o pi o psi**-1, where psi:E->E' is the degree 6 iso defined over Fp24
var endo struct {
	u fptower.E4
	v fptower.E4
}

// generator of the curve
var xGen big.Int

// expose the tower -- github.com/consensys/gnark uses it in a gnark circuit

// E2 is a degree two finite field extension of fp.Element
type E2 = fptower.E2

// E4 is a degree two finite field extension of fp2
type E4 = fptower.E4

// E12 is a degree three finite field extension of fp4
type E12 = fptower.E12

// E24 is a degree two finite field extension of fp6
type E24 = fptower.E24

func init() {

	bCurveCoeff.SetUint64(1)
	twist.B1.SetOne()
	bTwistCurveCoeff.Inverse(&twist)

	twoInv.SetOne().Double(&twoInv).Inverse(&twoInv)

	// E(1,y)*c
	g1Gen.X.SetString("34223510504517033132712852754388476272837911830964394866541204856091481856889569724484362330263")
	g1Gen.Y.SetString("24215295174889464585413596429561903295150472552154479431771837786124301185073987899223459122783")
	g1Gen.Z.SetString("1")

	// E'(5,y)*c'
	g2Gen.X.B0.SetString("24614737899199071964341749845083777103809664018538138889239909664991294445469052467064654073699",
		"17049297748993841127032249156255993089778266476087413538366212660716380683149731996715975282972")
	g2Gen.X.B1.SetString("11950668649125904104557740112865942804623051114821811669564995102755430514441092495782202668342",
		"3603055379462539802413979855826194299714805833759849528529386570240639115620788686893505938793")
	g2Gen.Y.B0.SetString("31740092748246070457677943092194030978994615503726570180895475408200863271773078192139722193079",
		"30261413948955264769241509843031153941332801192447678605718183215275065425758214858190865971597")
	g2Gen.Y.B1.SetString("14195825602561496219090410113749222574308144851497375443809100117082380611212823440674391088885",
		"2391152940984805871402135750194189812615420966694899795235607856168224901793030297133493038211")
	g2Gen.Z.B0.SetString("1",
		"0")
	g2Gen.Z.B1.SetString("0",
		"0")

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()
	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	thirdRootOneG1.SetString("39705142672498995661671850106945620852186608752525090699191017895721506694646055668218723303426") // -x^9+3x^8-4x^7+4x^6-3x^5+2x^3-2x^2+x-2 ((-1+tx/yx)/2 mod px)
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("11502027791375260645628074404575422496066855707288983427913398978447461580801", 10) // x^8
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.B0.A0.SetString("17432737665785421589107433512831558061649422754130449334965277047994983947893909429238815314776")
	endo.v.B0.A0.SetString("13266452002786802757645810648664867986567631927642464177452792960815113608167203350720036682455")

	// binary decomposition of 3218079743 little endian
	optimaAteLoop, _ := new(big.Int).SetString("3218079743", 10)
	ecc.NafDecomposition(optimaAteLoop, loopCounter[:])

	xGen.SetString("3218079743", 10)
}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g2Jac G2Jac, g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	g1Jac = g1Gen
	g2Jac = g2Gen
	return
}
