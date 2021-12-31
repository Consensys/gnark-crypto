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

	"github.com/consensys/gnark-crypto/ecc/bw6-756/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BW6-756's Fr
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

	edwards.A.SetUint64(143580)
	edwards.D.SetUint64(143576)
	edwards.Cofactor.SetUint64(8).FromMont()
	edwards.Order.SetString("75656025759413271466656060197725120092480961471365614219134998880569790930794516726065877484428941069706901665493", 10)

	edwards.Base.X.SetString("178620376715698421301710631119120785579284871526578026139185646772672252736182448135689014711987732666078420387915")
	edwards.Base.Y.SetString("279345325880910540799960837653138904956852780817349960193932651092957355032339063742900216468694143617372745972501")
}

// mulByA multiplies fr.Element by edwards.A
func mulByA(x *fr.Element) {
	x.Mul(x, &edwards.A)
}
