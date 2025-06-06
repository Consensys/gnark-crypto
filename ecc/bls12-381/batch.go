package bls12381

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// IsInSubGroupBatchNaive checks if a batch of points P_i are in G1.
// This is a naive method that checks each point individually.
func IsInSubGroupBatchNaive(points []G1Affine) bool {
	for i := range points {
		if !points[i].IsInSubGroup() {
			return false
		}
	}
	return true
}

// IsInSubGroupBatch checks if a batch of points P_i are in G1.
// It generates random scalars s_i in the range [0, bound) and performs
// n=rounds multi-scalar-multiplication ∑[s_i]P_i of sizes N=len(points)
func IsInSubGroupBatch(points []G1Affine, bound *big.Int, rounds int) bool {

	// 1. Check points are on E[r*e']
	for i := range points {
		// 1.1. Tate_{3,P3}(Q) = (y-2)^((p-1)/3) == 1, with P3 = (0,2).
		if !isFirstTateOne(points[i]) {
			return false
		}
		// 1.2. Tate_{11,P11}(Q) == 1
		if !isSecondTateOne(points[i]) {
			return false
		}
	}

	// 2. Check Sj are on E[r]
	for i := 0; i < rounds; i++ {
		b, err := rand.Int(rand.Reader, bound)
		if err != nil {
			panic(err)
		}
		randoms := make([]fr.Element, len(points))
		for j := range randoms {
			randoms[j].SetBigInt(b)
		}
		var sum G1Jac
		sum.MultiExp(points[:], randoms[:], ecc.MultiExpConfig{})
		if !sum.IsInSubGroup() {
			return false
		}
	}
	return true
}

// isFirstTateOne checks that Tate_{3,P3}(Q) = (y-2)^((p-1)/3) == 1
// where P3 = (0,2) a point of order 3 on the curve.
func isFirstTateOne(point G1Affine) bool {
	var tate, two fp.Element
	two.SetInt64(2)
	tate.Sub(&point.Y, &two).Exp(tate, &exp1)
	return tate.IsOne()
}

// isSecondTateOne checks that Tate_{11,P11}(Q) == 1
// where P11 = (x,y) a point of order 11 on the curve.
// x = 0x1147c19050b3c4b663a4ca29c4859eeb1ac05a91659009602e7443347ad659e9f838f4ed07337c4c6d3a48d612b4bb92
// y = 0x8d7c25237c7dcea6ea0c6c37053882c59cc0ee424b3545bb25116d53e383574063149edb438b959dd169d0e01b2d3bc
func isSecondTateOne(point G1Affine) bool {

	// f_{11,P} = (l_{P,P}^4 * (l_{4P,P} * l_{2P,2P})^2 * l_{5P,5P} * v_P) /
	// 			  (v_{2P} * (v_{5P} * v_{4P})^2 * v_{10P})

	var num, denom, tate, f1, f2 fp.Element

	// l_{P,P}^4
	f1.Mul(&point.X, &lines[0].a).Add(&f1, &point.Y).Add(&f1, &lines[0].b)
	num.Square(&f1).Square(&num)
	// (l_{4P,P} * l_{2P,2P})^2
	f1.Mul(&point.X, &lines[1].a).Add(&f1, &point.Y).Add(&f1, &lines[1].b)
	f2.Mul(&point.X, &lines[2].a).Add(&f2, &point.Y).Add(&f2, &lines[2].b)
	f1.Mul(&f1, &f2).Square(&f1)
	num.Mul(&num, &f1)
	// l_{5P,5P}
	f1.Mul(&point.X, &lines[3].a).Add(&f1, &point.Y).Add(&f1, &lines[3].b)
	num.Mul(&num, &f1)
	// v_P
	f1.Add(&point.X, &verticals[0].a)
	num.Mul(&num, &f1)

	// v_{2P}^4
	f1.Add(&point.X, &verticals[1].a)
	denom.Square(&f1).Square(&denom)
	// (v_{5P} * v_{4P})^2
	f1.Add(&point.X, &verticals[2].a)
	f2.Add(&point.X, &verticals[3].a)
	f1.Mul(&f1, &f2).Square(&f1)
	denom.Mul(&denom, &f1)
	// v_{10P}
	f1.Add(&point.X, &verticals[4].a)
	denom.Mul(&denom, &f1)

	// denom^{-1} = denom^{10} in (./p)_{11}
	f1.Square(&denom)
	f2.Square(&f1).Square(&f2)
	denom.Mul(&f1, &f2)

	// tate = num * denom^{-1}
	tate.Mul(&num, &denom)

	// tate^((p-1)/11)
	tate.Exp(tate, &exp2)

	return tate.IsOne()
}

// --------------------
type line struct {
	a, b fp.Element
}

type vertical struct {
	a fp.Element
}

var lines [4]line
var verticals [5]vertical
var exp1, exp2 big.Int

func init() {
	// P = (
	// 	   0x1147c19050b3c4b663a4ca29c4859eeb1ac05a91659009602e7443347ad659e9f838f4ed07337c4c6d3a48d612b4bb92,
	// 	   0x8d7c25237c7dcea6ea0c6c37053882c59cc0ee424b3545bb25116d53e383574063149edb438b959dd169d0e01b2d3bc,
	// )
	// l_{P,P}
	lines[0].a.SetString("789121243908217914986864598066009119517422372378730756745337365117675966292344742905526727060437918078278447811549")
	lines[0].b.SetString("3612936748462981376847170683871485802736210445703285432212303692032207881168735791458346615228397209773215450322724")
	// l_{4P,P}
	lines[1].a.SetString("3044955060911815760317430904013995440345887181956594873416907983204589894974213429837152871343960265135164233337361")
	lines[1].b.SetString("2923263761390984622063929787982700865830737712220983008621080895356443028600442613029223424168507822973127367399724")
	// l_{2P,2P}
	lines[2].a.SetString("3411000602890276893231702527782215157235594872160146738324812479245553530339046596170681084273733812777231937319695")
	lines[2].b.SetString("3799684642228555457262516666742549614090265470273353723005096827797599755046665230249426064988247375982743408812613")
	// l_{5P,5P}
	lines[3].a.SetString("1899248040746765214347655688391736275055198297344521124529117722142686150237741687305287681132657130556244969856414")
	lines[3].b.SetString("2220821630890179640850044703399006017189993769556623826482523541536896844660108417748957554483159722637182947913708")
	// v_{P}
	verticals[0].a.SetString("1342728378592133176225836625560744379350333988650090061881201644643929491989325398013742949610187243243594448563993")
	// v_{2P}
	verticals[1].a.SetString("3593196851125462192837623759409677287782506485357690247608362212285083916026308461561859647188864289044500677504754")
	// v_{5P}
	verticals[2].a.SetString("402067627672051250698017802728272265604843594840627748942542238023604973355831187895721091551929220859648553932084")
	// v_{4P}
	verticals[3].a.SetString("2918694819079567036475700436273278909235182889931955058334034890737743469722865331326312274449648412759333253175181")
	// v_{10P}
	verticals[4].a.SetString("1342728378592133176225836625560744379350333988650090061881201644643929491989325398013742949610187243243594448563993")

	// (p-1)/3
	exp1.SetString("1334136518407222464472596608578634718852294273313002628444019378708010550163612621480895876376338554679298090853262", 10)
	// (p-1)/11
	exp2.SetString("363855414111060672128889984157809468777898438176273444121096194193093786408257987676607966284455969457990388414526", 10)
}
