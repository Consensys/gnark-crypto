package twistededwards

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// CurveParams curve parameters: -x^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	D        fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BN254's Fr
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

	edwards.D.SetString("12181644023421730124874158521699555681764249180949974110617291017600649128846")
	edwards.Cofactor.SetUint64(8).FromMont()
	edwards.Order.SetString("2736030358979909402780800718157159386076813972158567259200215660948447373041", 10)

	edwards.Base.X.SetString("9671717474070082183213120605117400219616337014328744928644933853176787189663")
	edwards.Base.Y.SetString("16950150798460657717958625567821834550301663161624707787222815936182638968203")
}
