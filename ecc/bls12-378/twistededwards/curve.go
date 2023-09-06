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

	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element
	Cofactor fr.Element
	Order    big.Int
	Base     PointAffine
}

// GetEdwardsCurve returns the twisted Edwards curve on bls12-378/Fr
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
	curveParams.A.SetString("16249")
	curveParams.D.SetString("826857503717340716663906603396009292766308904506333520048618402505612607353")
	curveParams.Cofactor.SetString("8")
	curveParams.Order.SetString("1860429383364016612493789857641020908721690454530426945748883177201355593303", 10)

	curveParams.Base.X.SetString("6772953896463446981848394912418300623023000177913479948380771331313783560843")
	curveParams.Base.Y.SetString("9922290044608088599966879240752111513195706854076002240583420830067351093249")
}

// mulByA multiplies fr.Element by curveParams.A
func mulByA(x *fr.Element) {
	x.Mul(x, &curveParams.A)
}

// mulByD multiplies fr.Element by curveParams.D
func mulByD(x *fr.Element) {
	x.Mul(x, &curveParams.D)
}
