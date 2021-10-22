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

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
	endo     [2]fr.Element // in Montgomery form
	lambda   big.Int
	glvBasis ecc.Lattice
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BLS12-381's Fr
func GetEdwardsCurve() CurveParams {

	// copy to keep Order private
	var res CurveParams

	res.A.Set(&edwards.A)
	res.D.Set(&edwards.D)
	res.Cofactor.Set(&edwards.Cofactor)
	res.Order.Set(&edwards.Order)
	res.Base.Set(&edwards.Base)
	res.endo[0].Set(&edwards.endo[0])
	res.endo[1].Set(&edwards.endo[1])
	res.lambda.Set(&edwards.lambda)
	res.glvBasis = edwards.glvBasis

	return res
}

func init() {

	edwards.A.SetUint64(5).Neg(&edwards.A)
	edwards.D.SetString("45022363124591815672509500913686876175488063829319466900776701791074614335719") // -(138827208126141220649022263972958607803/171449701953573178309673572579671231137)
	edwards.Cofactor.SetUint64(4).FromMont()
	edwards.Order.SetString("13108968793781547619861935127046491459309155893440570251786403306729687672801", 10)

	edwards.Base.X.SetString("18886178867200960497001835917649091219057080094937609519140440539760939937304")
	edwards.Base.Y.SetString("19188667384257783945677642223292697773471335439753913231509108946878080696678")
	edwards.endo[0].SetString("37446463827641770816307242315180085052603635617490163568005256780843403514036")
	edwards.endo[1].SetString("49199877423542878313146170939139662862850515542392585932876811575731455068989")
	edwards.lambda.SetString("8913659658109529928382530854484400854125314752504019737736543920008458395397", 10)
	ecc.PrecomputeLattice(&edwards.Order, &edwards.lambda, &edwards.glvBasis)

}
