package twistededwards

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
)

// CurveParams curve parameters: -x^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	D        fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BLS24-315's Fr
func GetEdwardsCurve() CurveParams {

	// copy to keep Order private
	var res CurveParams

	res.D.Set(&edwards.D)
	res.Cofactor.Set(&edwards.Cofactor)
	res.Order.Set(&edwards.Order)
	res.Base.Set(&edwards.Base)

	return res
}

func init() {

	edwards.D.SetString("8771873785799030510227956919069912715983412030268481769609515223557738569779")
	edwards.Cofactor.SetUint64(8).FromMont()
	edwards.Order.SetString("1437753473921907580703509300571927811987591765799164617677716990775193563777", 10)

	edwards.Base.X.SetString("750878639751052675245442739791837325424717022593512121860796337974109802674")
	edwards.Base.Y.SetString("1210739767513185331118744674165833946943116652645479549122735386298364723201")
}
