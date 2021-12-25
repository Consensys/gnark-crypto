/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package twistededwards

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BLS12-378's Fr
func GetEdwardsCurve() CurveParams {

	// copy to keep Order private
	var res CurveParams

	res.A.Set(&edwards.A)
	res.D.Set(&edwards.D)
	res.Cofactor.Set(&edwards.Cofactor)
	res.Order.Set(&edwards.Order)
	res.Base.Set(&edwards.Base)

	return res
}

func init() {

	edwards.A.SetString("1169928")
	edwards.D.SetString("1169924")
	edwards.Cofactor.SetUint64(8).FromMont()
	edwards.Order.SetString("1860429383364016612493789857641020908721690454530426945748883177201355593303", 10)

	edwards.Base.X.SetString("4274983589151226901853657690021194631121133716096168671136076068148698830183")
	edwards.Base.Y.SetString("9922290044608088599966879240752111513195706854076002240583420830067351093249")
}

// mulByA multiplies fr.Element by edwards.A
func mulByA(x *fr.Element) {
	x.Mul(x, &edwards.A)
}
