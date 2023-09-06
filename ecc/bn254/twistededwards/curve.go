// Copyright 2020 Consensys Software Inc.
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

// Code generated by consensys/gnark-crypto DO NOT EDIT

package twistededwards

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element
	Cofactor fr.Element
	Order    big.Int
	Base     PointAffine
}

// GetEdwardsCurve returns the twisted Edwards curve on bn254/Fr
func GetEdwardsCurve() CurveParams {
	initOnce.Do(initCurveParams)
	// copy to keep Order private
	var res CurveParams

	res.A.Set(&curveParams.A)
	res.D.Set(&curveParams.D)
	res.Cofactor.Set(&curveParams.Cofactor)
	res.Order.Set(&curveParams.Order)
	res.Base.Set(&curveParams.Base)

	return res
}

var (
	initOnce    sync.Once
	curveParams CurveParams
)

func initCurveParams() {
	curveParams.A.SetString("-1")
	curveParams.D.SetString("12181644023421730124874158521699555681764249180949974110617291017600649128846")
	curveParams.Cofactor.SetString("8")
	curveParams.Order.SetString("2736030358979909402780800718157159386076813972158567259200215660948447373041", 10)

	curveParams.Base.X.SetString("9671717474070082183213120605117400219616337014328744928644933853176787189663")
	curveParams.Base.Y.SetString("16950150798460657717958625567821834550301663161624707787222815936182638968203")
}

// mulByA multiplies fr.Element by curveParams.A
func mulByA(x *fr.Element) {
	x.Neg(x)
}

// mulByD multiplies fr.Element by curveParams.D
func mulByD(x *fr.Element) {
	// Generated by github.com/mmcloughlin/addchain v0.4.0.
	// Operations: 248 doublings 58 additions

	var z, t0, t1, t2, t3, t4, t5, t6, t7, t8, t9, t10, t11, t12, t13, t14, t15 fr.Element

	t1.Double(x)
	t14.Add(x, &t1)
	t8.Add(&t1, &t14)
	z.Add(&t1, &t8)
	t13.Add(&t1, &z)
	t12.Add(&t1, &t13)
	t2.Add(&t1, &t12)
	t6.Add(&t1, &t2)
	t11.Add(&t1, &t6)
	t7.Add(&t1, &t11)
	t5.Add(&t1, &t7)
	t0.Add(&t1, &t5)
	t4.Add(&t1, &t0)
	t9.Add(&t1, &t4)
	t3.Add(&t1, &t9)
	t10.Add(&t0, &t3)
	t1.Add(&t12, &t10)
	t15.Double(&t10)
	t10.Add(&t0, &t15)
	for s := 0; s < 3; s++ {
		t15.Double(&t15)
	}
	t15.Add(&t3, &t15)
	for s := 0; s < 4; s++ {
		t15.Double(&t15)
	}
	t15.Add(&t2, &t15)
	for s := 0; s < 3; s++ {
		t15.Double(&t15)
	}
	t15.Add(x, &t15)
	for s := 0; s < 8; s++ {
		t15.Double(&t15)
	}
	t15.Add(&t6, &t15)
	for s := 0; s < 8; s++ {
		t15.Double(&t15)
	}
	t15.Add(&t5, &t15)
	for s := 0; s < 4; s++ {
		t15.Double(&t15)
	}
	t15.Add(&t6, &t15)
	for s := 0; s < 3; s++ {
		t15.Double(&t15)
	}
	t15.Add(x, &t15)
	for s := 0; s < 6; s++ {
		t15.Double(&t15)
	}
	t14.Add(&t14, &t15)
	for s := 0; s < 7; s++ {
		t14.Double(&t14)
	}
	t13.Add(&t13, &t14)
	for s := 0; s < 5; s++ {
		t13.Double(&t13)
	}
	t13.Add(&t2, &t13)
	for s := 0; s < 7; s++ {
		t13.Double(&t13)
	}
	t13.Add(&t7, &t13)
	for s := 0; s < 5; s++ {
		t13.Double(&t13)
	}
	t13.Add(&t9, &t13)
	for s := 0; s < 3; s++ {
		t13.Double(&t13)
	}
	t13.Add(&z, &t13)
	for s := 0; s < 8; s++ {
		t13.Double(&t13)
	}
	t13.Add(&z, &t13)
	for s := 0; s < 6; s++ {
		t13.Double(&t13)
	}
	t12.Add(&t12, &t13)
	for s := 0; s < 6; s++ {
		t12.Double(&t12)
	}
	t12.Add(&t0, &t12)
	for s := 0; s < 5; s++ {
		t12.Double(&t12)
	}
	t12.Add(&t7, &t12)
	for s := 0; s < 9; s++ {
		t12.Double(&t12)
	}
	t12.Add(&t10, &t12)
	for s := 0; s < 6; s++ {
		t12.Double(&t12)
	}
	t11.Add(&t11, &t12)
	for s := 0; s < 10; s++ {
		t11.Double(&t11)
	}
	t10.Add(&t10, &t11)
	for s := 0; s < 11; s++ {
		t10.Double(&t10)
	}
	t10.Add(&z, &t10)
	for s := 0; s < 7; s++ {
		t10.Double(&t10)
	}
	t9.Add(&t9, &t10)
	for s := 0; s < 5; s++ {
		t9.Double(&t9)
	}
	t8.Add(&t8, &t9)
	for s := 0; s < 7; s++ {
		t8.Double(&t8)
	}
	t8.Add(&t5, &t8)
	for s := 0; s < 5; s++ {
		t8.Double(&t8)
	}
	t7.Add(&t7, &t8)
	for s := 0; s < 4; s++ {
		t7.Double(&t7)
	}
	t7.Add(&t2, &t7)
	for s := 0; s < 6; s++ {
		t7.Double(&t7)
	}
	t7.Add(&t3, &t7)
	for s := 0; s < 5; s++ {
		t7.Double(&t7)
	}
	t7.Add(&t0, &t7)
	for s := 0; s < 5; s++ {
		t7.Double(&t7)
	}
	t7.Add(&t4, &t7)
	for s := 0; s < 9; s++ {
		t7.Double(&t7)
	}
	t7.Add(&t4, &t7)
	for s := 0; s < 4; s++ {
		t7.Double(&t7)
	}
	t6.Add(&t6, &t7)
	for s := 0; s < 2; s++ {
		t6.Double(&t6)
	}
	t6.Add(x, &t6)
	for s := 0; s < 10; s++ {
		t6.Double(&t6)
	}
	t6.Add(&t3, &t6)
	for s := 0; s < 4; s++ {
		t6.Double(&t6)
	}
	t6.Add(&z, &t6)
	for s := 0; s < 7; s++ {
		t6.Double(&t6)
	}
	t5.Add(&t5, &t6)
	for s := 0; s < 8; s++ {
		t5.Double(&t5)
	}
	t4.Add(&t4, &t5)
	for s := 0; s < 5; s++ {
		t4.Double(&t4)
	}
	t3.Add(&t3, &t4)
	for s := 0; s < 6; s++ {
		t3.Double(&t3)
	}
	t2.Add(&t2, &t3)
	for s := 0; s < 7; s++ {
		t2.Double(&t2)
	}
	t1.Add(&t1, &t2)
	for s := 0; s < 6; s++ {
		t1.Double(&t1)
	}
	t0.Add(&t0, &t1)
	for s := 0; s < 6; s++ {
		t0.Double(&t0)
	}
	z.Add(&z, &t0)
	x.Double(&z)
}
