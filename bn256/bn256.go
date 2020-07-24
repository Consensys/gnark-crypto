package bn256

import (
	"math/big"

	"github.com/consensys/gurvy"
	"github.com/consensys/gurvy/bn256/fp"
	"github.com/consensys/gurvy/utils"
)

// ID bn256 ID
var ID = gurvy.BN256

// B b coeff of the curve
var B fp.Element

// generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
var g1Gen G1Jac
var g2Gen G2Jac

// point at infinity
var g1Infinity G1Jac
var g2Infinity G2Jac

// optimal Ate loop counter (=trace-1 = x in BLS family)
var loopCounter [66]int8

// parameters for pippenger ScalarMulByGen
// TODO get rid of this, keep only double and add, and the multi exp
const sGen = 4
const bGen = sGen

var tGenG1 [((1 << bGen) - 1)]G1Jac
var tGenG2 [((1 << bGen) - 1)]G2Jac

func init() {

	B.SetUint64(1)

	g1Gen.X.SetString("20567171726433170376993012834626974355708098753738075953327671604980729474588")
	g1Gen.Y.SetString("14259118686601658563517637559143782061303537174604067025175876803301021346267")
	g1Gen.Z.SetString("1")

	g2Gen.X.SetString("14433365730775072582213482468844163390964025019096075555058505630999708262443",
		"3683446723006852480794963570030936618743148392137679437247363531986401769417")
	g2Gen.Y.SetString("21253271987667943455369004300257637004831224612428754877033343975009216128128",
		"12495620673937637012904672587588023149812491484245871073230980321212840773339")
	g2Gen.Z.SetString("1",
		"0")

	// binary decomposition of 15132376222941642752 little endian
	optimaAteLoop, _ := new(big.Int).SetString("29793968203157093288", 10)
	utils.NafDecomposition(optimaAteLoop, loopCounter[:])

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	g2Infinity.X.SetOne()
	g2Infinity.Y.SetOne()

	tGenG1[0].Set(&g1Gen)
	for j := 1; j < len(tGenG1)-1; j = j + 2 {
		tGenG1[j].Set(&tGenG1[j/2]).DoubleAssign()
		tGenG1[j+1].Set(&tGenG1[(j+1)/2]).AddAssign(&tGenG1[j/2])
	}
	tGenG2[0].Set(&g2Gen)
	for j := 1; j < len(tGenG2)-1; j = j + 2 {
		tGenG2[j].Set(&tGenG2[j/2]).DoubleAssign()
		tGenG2[j+1].Set(&tGenG2[(j+1)/2]).AddAssign(&tGenG2[j/2])
	}
}

// // TODO go:generate go run ../internal/generator.go -out . -package bn256 -t 4965661367192848881 -p 21888242871839275222246405745257275088696311157297823662689037894645226208583 -r 21888242871839275222246405745257275088548364400416034343698204186575808495617 -fp2 -1 -fp6 9,1

// // E: y**2=x**3+3
// // Etwist: y**2 = x**3+3*(u+9)**-1

// var bn256 Curve
// var initOnce sync.Once

// // ID bn256 ID
// const ID = gurvy.BN256

// // parameters for pippenger ScalarMulByGen
// const sGen = 4
// const bGen = sGen

// // PairingResult target group of the pairing
// type PairingResult = E12

// // BN256 returns BN256 curve
// func BN256() *Curve {
// 	initOnce.Do(initBN256)
// 	return &bn256
// }

// // Curve represents the BLS381 curve and pre-computed constants
// type Curve struct {

// 	// A, B coefficients of the curve x^3 = y^2 +AX+b
// 	B fp.Element

// 	// generators of the r-torsion subgroup, g1 in ker(pi), g2 in ker(p-q)
// 	g1Gen      G1Jac
// 	g2Gen      G2Jac
// 	g1Infinity G1Jac
// 	g2Infinity G2Jac

// 	// NAF decomposition of t-1, t is the trace of the Frobenius
// 	loopCounter [66]int8

// 	// precomputed values for ScalarMulByGen
// 	tGenG1 [((1 << bGen) - 1)]G1Jac
// 	tGenG2 [((1 << bGen) - 1)]G2Jac
// }

// func initBN256() {

// 	// A, B coeffs of the curve in Mont form
// 	bn256.B.SetUint64(3)

// 	// Setting G1Jac
// 	bn256.g1Gen.X.SetString("20567171726433170376993012834626974355708098753738075953327671604980729474588")
// 	bn256.g1Gen.Y.SetString("14259118686601658563517637559143782061303537174604067025175876803301021346267")
// 	bn256.g1Gen.Z.SetString("1")

// 	// Setting G2Jac
// 	bn256.g2Gen.X.SetString("14433365730775072582213482468844163390964025019096075555058505630999708262443",
// 		"3683446723006852480794963570030936618743148392137679437247363531986401769417")
// 	bn256.g2Gen.Y.SetString("21253271987667943455369004300257637004831224612428754877033343975009216128128",
// 		"12495620673937637012904672587588023149812491484245871073230980321212840773339")
// 	bn256.g2Gen.Z.SetString("1",
// 		"0")

// 	// Setting the loop counter for Miller loop in NAF form (6*t+2)
// 	T, _ := new(big.Int).SetString("29793968203157093288", 10)
// 	utils.NafDecomposition(T, bn256.loopCounter[:])

// 	// infinity point G1
// 	bn256.g1Infinity.X.SetOne()
// 	bn256.g1Infinity.Y.SetOne()

// 	// infinity point G2
// 	bn256.g2Infinity.X.SetOne()
// 	bn256.g2Infinity.Y.SetOne()

// 	// precomputed values for ScalarMulByGen
// 	bn256.tGenG1[0].Set(&bn256.g1Gen)
// 	for j := 1; j < len(bn256.tGenG1)-1; j = j + 2 {
// 		bn256.tGenG1[j].Set(&bn256.tGenG1[j/2]).DoubleAssign()
// 		bn256.tGenG1[j+1].Set(&bn256.tGenG1[(j+1)/2]).AddAssign(&bn256, &bn256.tGenG1[j/2])
// 	}
// 	bn256.tGenG2[0].Set(&bn256.g2Gen)
// 	for j := 1; j < len(bn256.tGenG2)-1; j = j + 2 {
// 		bn256.tGenG2[j].Set(&bn256.tGenG2[j/2]).DoubleAssign()
// 		bn256.tGenG2[j+1].Set(&bn256.tGenG2[(j+1)/2]).AddAssign(&bn256, &bn256.tGenG2[j/2])
// 	}
// }
