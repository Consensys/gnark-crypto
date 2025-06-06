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
	tate = *expByp11(&tate)

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
var exp1 big.Int

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
}

// expByp11 uses a short addition chain to compute x^p11 where p11=(p-1)/11 .
func expByp11(x *fp.Element) *fp.Element {
	// Operations: 372 squares 77 multiplies
	//
	// Generated by github.com/mmcloughlin/addchain v0.4.0.

	// Allocate Temporaries.
	var z = new(fp.Element)
	var (
		t0  = new(fp.Element)
		t1  = new(fp.Element)
		t2  = new(fp.Element)
		t3  = new(fp.Element)
		t4  = new(fp.Element)
		t5  = new(fp.Element)
		t6  = new(fp.Element)
		t7  = new(fp.Element)
		t8  = new(fp.Element)
		t9  = new(fp.Element)
		t10 = new(fp.Element)
		t11 = new(fp.Element)
		t12 = new(fp.Element)
		t13 = new(fp.Element)
		t14 = new(fp.Element)
		t15 = new(fp.Element)
		t16 = new(fp.Element)
		t17 = new(fp.Element)
		t18 = new(fp.Element)
		t19 = new(fp.Element)
		t20 = new(fp.Element)
		t21 = new(fp.Element)
		t22 = new(fp.Element)
		t23 = new(fp.Element)
		t24 = new(fp.Element)
		t25 = new(fp.Element)
		t26 = new(fp.Element)
		t27 = new(fp.Element)
		t28 = new(fp.Element)
		t29 = new(fp.Element)
		t30 = new(fp.Element)
	)

	// Step 1: t0 = x^0x2
	t0.Square(x)

	// Step 2: t9 = x^0x4
	t9.Square(t0)

	// Step 3: t1 = x^0x8
	t1.Square(t9)

	// Step 4: t16 = x^0xc
	t16.Mul(t9, t1)

	// Step 5: t15 = x^0xe
	t15.Mul(t0, t16)

	// Step 6: t5 = x^0x12
	t5.Mul(t9, t15)

	// Step 7: t29 = x^0x13
	t29.Mul(x, t5)

	// Step 8: t2 = x^0x17
	t2.Mul(t9, t29)

	// Step 9: t11 = x^0x1a
	t11.Mul(t1, t5)

	// Step 10: t8 = x^0x1b
	t8.Mul(x, t11)

	// Step 11: t6 = x^0x1d
	t6.Mul(t0, t8)

	// Step 12: t18 = x^0x1e
	t18.Mul(x, t6)

	// Step 13: z = x^0x1f
	z.Mul(x, t18)

	// Step 14: t24 = x^0x27
	t24.Mul(t1, z)

	// Step 15: t12 = x^0x29
	t12.Mul(t0, t24)

	// Step 16: t28 = x^0x43
	t28.Mul(t11, t12)

	// Step 17: t21 = x^0x4f
	t21.Mul(t16, t28)

	// Step 18: t14 = x^0x51
	t14.Mul(t0, t21)

	// Step 19: t4 = x^0x55
	t4.Mul(t9, t14)

	// Step 20: t7 = x^0x59
	t7.Mul(t9, t4)

	// Step 21: t27 = x^0x5b
	t27.Mul(t0, t7)

	// Step 22: t3 = x^0x5d
	t3.Mul(t0, t27)

	// Step 23: t10 = x^0x65
	t10.Mul(t1, t3)

	// Step 24: t23 = x^0x69
	t23.Mul(t9, t10)

	// Step 25: t17 = x^0x6b
	t17.Mul(t0, t23)

	// Step 26: t5 = x^0x7d
	t5.Mul(t5, t17)

	// Step 27: t26 = x^0x81
	t26.Mul(t9, t5)

	// Step 28: t25 = x^0x83
	t25.Mul(t0, t26)

	// Step 29: t9 = x^0x89
	t9.Mul(t1, t26)

	// Step 30: t19 = x^0x91
	t19.Mul(t1, t9)

	// Step 31: t1 = x^0x97
	t1.Mul(t15, t9)

	// Step 32: t20 = x^0xa1
	t20.Mul(t18, t25)

	// Step 33: t22 = x^0xa7
	t22.Mul(t18, t9)

	// Step 34: t13 = x^0xb1
	t13.Mul(t11, t1)

	// Step 35: t11 = x^0xc5
	t11.Mul(t18, t22)

	// Step 36: t15 = x^0xd3
	t15.Mul(t15, t11)

	// Step 37: t18 = x^0xf1
	t18.Mul(t18, t15)

	// Step 38: t16 = x^0xfd
	t16.Mul(t16, t18)

	// Step 39: t0 = x^0xff
	t0.Mul(t0, t16)

	// Step 40: t30 = x^0x12e
	t30.Mul(t27, t15)

	// Step 45: t30 = x^0x25c0
	for s := 0; s < 5; s++ {
		t30.Square(t30)
	}

	// Step 46: t29 = x^0x25d3
	t29.Mul(t29, t30)

	// Step 59: t29 = x^0x4ba6000
	for s := 0; s < 13; s++ {
		t29.Square(t29)
	}

	// Step 60: t29 = x^0x4ba6059
	t29.Mul(t7, t29)

	// Step 69: t29 = x^0x974c0b200
	for s := 0; s < 9; s++ {
		t29.Square(t29)
	}

	// Step 70: t28 = x^0x974c0b243
	t28.Mul(t28, t29)

	// Step 76: t28 = x^0x25d302c90c0
	for s := 0; s < 6; s++ {
		t28.Square(t28)
	}

	// Step 77: t28 = x^0x25d302c90dd
	t28.Mul(t6, t28)

	// Step 88: t28 = x^0x12e9816486e800
	for s := 0; s < 11; s++ {
		t28.Square(t28)
	}

	// Step 89: t28 = x^0x12e9816486e8a7
	t28.Mul(t22, t28)

	// Step 96: t28 = x^0x974c0b243745380
	for s := 0; s < 7; s++ {
		t28.Square(t28)
	}

	// Step 97: t27 = x^0x974c0b2437453db
	t27.Mul(t27, t28)

	// Step 110: t27 = x^0x12e9816486e8a7b6000
	for s := 0; s < 13; s++ {
		t27.Square(t27)
	}

	// Step 111: t26 = x^0x12e9816486e8a7b6081
	t26.Mul(t26, t27)

	// Step 120: t26 = x^0x25d302c90dd14f6c10200
	for s := 0; s < 9; s++ {
		t26.Square(t26)
	}

	// Step 121: t25 = x^0x25d302c90dd14f6c10283
	t25.Mul(t25, t26)

	// Step 127: t25 = x^0x974c0b2437453db040a0c0
	for s := 0; s < 6; s++ {
		t25.Square(t25)
	}

	// Step 128: t24 = x^0x974c0b2437453db040a0e7
	t24.Mul(t24, t25)

	// Step 139: t24 = x^0x4ba605921ba29ed8205073800
	for s := 0; s < 11; s++ {
		t24.Square(t24)
	}

	// Step 140: t23 = x^0x4ba605921ba29ed8205073869
	t23.Mul(t23, t24)

	// Step 149: t23 = x^0x974c0b2437453db040a0e70d200
	for s := 0; s < 9; s++ {
		t23.Square(t23)
	}

	// Step 150: t22 = x^0x974c0b2437453db040a0e70d2a7
	t22.Mul(t22, t23)

	// Step 159: t22 = x^0x12e9816486e8a7b608141ce1a54e00
	for s := 0; s < 9; s++ {
		t22.Square(t22)
	}

	// Step 160: t21 = x^0x12e9816486e8a7b608141ce1a54e4f
	t21.Mul(t21, t22)

	// Step 170: t21 = x^0x4ba605921ba29ed82050738695393c00
	for s := 0; s < 10; s++ {
		t21.Square(t21)
	}

	// Step 171: t20 = x^0x4ba605921ba29ed82050738695393ca1
	t20.Mul(t20, t21)

	// Step 181: t20 = x^0x12e9816486e8a7b608141ce1a54e4f28400
	for s := 0; s < 10; s++ {
		t20.Square(t20)
	}

	// Step 182: t19 = x^0x12e9816486e8a7b608141ce1a54e4f28491
	t19.Mul(t19, t20)

	// Step 194: t19 = x^0x12e9816486e8a7b608141ce1a54e4f28491000
	for s := 0; s < 12; s++ {
		t19.Square(t19)
	}

	// Step 195: t18 = x^0x12e9816486e8a7b608141ce1a54e4f284910f1
	t18.Mul(t18, t19)

	// Step 205: t18 = x^0x4ba605921ba29ed82050738695393ca12443c400
	for s := 0; s < 10; s++ {
		t18.Square(t18)
	}

	// Step 206: t17 = x^0x4ba605921ba29ed82050738695393ca12443c46b
	t17.Mul(t17, t18)

	// Step 215: t17 = x^0x974c0b2437453db040a0e70d2a727942488788d600
	for s := 0; s < 9; s++ {
		t17.Square(t17)
	}

	// Step 216: t16 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd
	t16.Mul(t16, t17)

	// Step 226: t16 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf400
	for s := 0; s < 10; s++ {
		t16.Square(t16)
	}

	// Step 227: t15 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d3
	t15.Mul(t15, t16)

	// Step 236: t15 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a600
	for s := 0; s < 9; s++ {
		t15.Square(t15)
	}

	// Step 237: t14 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a651
	t14.Mul(t14, t15)

	// Step 246: t14 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca200
	for s := 0; s < 9; s++ {
		t14.Square(t14)
	}

	// Step 247: t13 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b1
	t13.Mul(t13, t14)

	// Step 255: t13 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b100
	for s := 0; s < 8; s++ {
		t13.Square(t13)
	}

	// Step 256: t12 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b129
	t12.Mul(t12, t13)

	// Step 266: t12 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a400
	for s := 0; s < 10; s++ {
		t12.Square(t12)
	}

	// Step 267: t12 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b
	t12.Mul(t8, t12)

	// Step 278: t12 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d800
	for s := 0; s < 11; s++ {
		t12.Square(t12)
	}

	// Step 279: t11 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c5
	t11.Mul(t11, t12)

	// Step 287: t11 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c500
	for s := 0; s < 8; s++ {
		t11.Square(t11)
	}

	// Step 288: t10 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c565
	t10.Mul(t10, t11)

	// Step 298: t10 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159400
	for s := 0; s < 10; s++ {
		t10.Square(t10)
	}

	// Step 299: t9 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489
	t9.Mul(t9, t10)

	// Step 304: t9 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b29120
	for s := 0; s < 5; s++ {
		t9.Square(t9)
	}

	// Step 305: t8 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b2913b
	t8.Mul(t8, t9)

	// Step 320: t8 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d8000
	for s := 0; s < 15; s++ {
		t8.Square(t8)
	}

	// Step 321: t7 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d8059
	t7.Mul(t7, t8)

	// Step 328: t7 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c80
	for s := 0; s < 7; s++ {
		t7.Square(t7)
	}

	// Step 329: t6 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d
	t6.Mul(t6, t7)

	// Step 339: t6 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b2913b00b27400
	for s := 0; s < 10; s++ {
		t6.Square(t6)
	}

	// Step 340: t6 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b2913b00b2745d
	t6.Mul(t3, t6)

	// Step 351: t6 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e800
	for s := 0; s < 11; s++ {
		t6.Square(t6)
	}

	// Step 352: t5 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e87d
	t5.Mul(t5, t6)

	// Step 360: t5 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e87d00
	for s := 0; s < 8; s++ {
		t5.Square(t5)
	}

	// Step 361: t4 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e87d55
	t4.Mul(t4, t5)

	// Step 371: t4 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f55400
	for s := 0; s < 10; s++ {
		t4.Square(t4)
	}

	// Step 372: t3 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f5545d
	t3.Mul(t3, t4)

	// Step 380: t3 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f5545d00
	for s := 0; s < 8; s++ {
		t3.Square(t3)
	}

	// Step 381: t2 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f5545d17
	t2.Mul(t2, t3)

	// Step 392: t2 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b2913b00b2745d0faaa2e8b800
	for s := 0; s < 11; s++ {
		t2.Square(t2)
	}

	// Step 393: t1 = x^0x974c0b2437453db040a0e70d2a727942488788d6fd34ca2b12906c62b2913b00b2745d0faaa2e8b897
	t1.Mul(t1, t2)

	// Step 403: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25c00
	for s := 0; s < 10; s++ {
		t1.Square(t1)
	}

	// Step 404: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cff
	t1.Mul(t0, t1)

	// Step 412: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cff00
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 413: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cffff
	t1.Mul(t0, t1)

	// Step 421: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cffff00
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 422: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cffffff
	t1.Mul(t0, t1)

	// Step 430: t1 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cffffff00
	for s := 0; s < 8; s++ {
		t1.Square(t1)
	}

	// Step 431: t0 = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cffffffff
	t0.Mul(t0, t1)

	// Step 436: t0 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e87d551745c4b9fffffffe0
	for s := 0; s < 5; s++ {
		t0.Square(t0)
	}

	// Step 437: t0 = x^0x4ba605921ba29ed82050738695393ca12443c46b7e9a65158948363159489d80593a2e87d551745c4b9fffffffff
	t0.Mul(z, t0)

	// Step 447: t0 = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f5545d1712e7ffffffffc00
	for s := 0; s < 10; s++ {
		t0.Square(t0)
	}

	// Step 448: z = x^0x12e9816486e8a7b608141ce1a54e4f284910f11adfa6994562520d8c56522760164e8ba1f5545d1712e7ffffffffc1f
	z.Mul(z, t0)

	// Step 449: z = x^0x25d302c90dd14f6c102839c34a9c9e509221e235bf4d328ac4a41b18aca44ec02c9d1743eaa8ba2e25cfffffffff83e
	z.Square(z)

	return z
}
