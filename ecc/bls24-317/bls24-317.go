package bls24317

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fp"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/internal/fptower"
)

// E: y**2=x**3+4
// Etwist: y**2 = x**3+4*v
// Tower: Fp->Fp2, u**2=-1 -> Fp4, v**2=u+1 -> Fp8, w**2=v -> Fp24, i**3=w
// Generator (BLS24 family): x=0xd9018000 (32 bits)
// optimal Ate loop: trace(frob)-1=x
// trace of pi: x+1
// Fp: p=0x1058ca226f60892cf28fc5a0b7f9d039169a61e684c73446d6f339e43424bf7e8d512e565dab2aab (317 bits)
// Fr: r=0x443f917ea68dafc2d0b097f28d83cd491cd1e79196bf0e7af000000000000001 (255 bits)

// ID bls317 ID
const ID = ecc.BLS24_317

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

// glvBasis stores R-linearly independent vectors (a,b), (c,d)
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

	bCurveCoeff.SetUint64(4)
	// M-twist
	twist.B1.SetOne()
	bTwistCurveCoeff.MulByElement(&twist, &bCurveCoeff)

	twoInv.SetOne().Double(&twoInv).Inverse(&twoInv)

	// E(1,y)*c
	g1Gen.X.SetString("26261810162995192444253184251590159762050205376519976412461726336843100448942248976252388876791")
	g1Gen.Y.SetString("26146603602820658047261036676090398397874822703333117264049387703172159980214065566219085800243")
	g1Gen.Z.SetString("1")

	// E'(1,y)*c'
	g2Gen.X.B0.SetString("28498404142312365002533744693556861244212064443103687717510540998257508853975496760832205123607",
		"104881342316154169720140745551267577558255475983798552134082689646705436288255501236462500135051")
	g2Gen.X.B1.SetString("134208762611471838850128095341317427866582025424914361408168906642550705688378271974920859507485",
		"47807860684290705153036437491997319116342330273104493957877398921782737166446662055996604784294")
	g2Gen.Y.B0.SetString("91516448788529060702418635560646746547369142933278847722177434542449427480796649633689953798948",
		"13448671391015186163413673966297442264556781166352891049005282051703895543542296449974630011689")
	g2Gen.Y.B1.SetString("1980905665816458576882252418967038151483710575831277397652951146268622037800272983431026055487",
		"134363379072057086809745572347104070037544575425956896869689256737197090432635401300100624083192")
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

	thirdRootOneG1.SetString("112388585831426139305998878408983604164339968939599860577886592073045019257058155724801")
	thirdRootOneG2.Square(&thirdRootOneG1)
	lambdaGLV.SetString("30869589236456844204538189757527902584770424025911415822847175497150445387776", 10) // x^8
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &lambdaGLV, &glvBasis)

	endo.u.B0.A0.SetString("100835231576138384070271140557450756773581004948002542492497192760544145876107391019725843007951")
	endo.u.B0.A1.SetString("100835231576138384070271140557450756773581004948002542492497192760544145876107391019725843007951")
	endo.v.B1.A0.SetString("65063930028143676778466901566890018271632055221368035552739808236464024322431728149960968101")
	endo.v.B1.A1.SetString("65063930028143676778466901566890018271632055221368035552739808236464024322431728149960968101")

	// binary decomposition of xGen little endian
	optimaAteLoop, _ := new(big.Int).SetString("3640754176", 10)
	ecc.NafDecomposition(optimaAteLoop, loopCounter[:])

	xGen.SetString("3640754176", 10)
}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g2Jac G2Jac, g1Aff G1Affine, g2Aff G2Affine) {
	g1Aff = g1GenAff
	g2Aff = g2GenAff
	g1Jac = g1Gen
	g2Jac = g2Gen
	return
}
